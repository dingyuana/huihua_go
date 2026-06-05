package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/model"
	"huihua/finance/internal/repository"
)

// VoucherService handles manual voucher CRUD operations.
type VoucherService struct {
	journalRepo        *repository.JournalRepository
	templateSvc        *VoucherTemplateService
	bankTxnRepo        *repository.BankTransactionRepository
	paymentRepo        *repository.PaymentEntryRepository
	accountRepo        *repository.AccountRepository
	classificationSvc  *ClassificationRuleService
	paymentStateMachine *PaymentStateMachine
	invoiceRepo        *repository.InvoiceRepository
}

// NewVoucherService creates a new VoucherService.
func NewVoucherService(
	journalRepo *repository.JournalRepository,
	templateSvc *VoucherTemplateService,
	bankTxnRepo *repository.BankTransactionRepository,
	paymentRepo *repository.PaymentEntryRepository,
	accountRepo *repository.AccountRepository,
	classificationSvc *ClassificationRuleService,
	paymentStateMachine *PaymentStateMachine,
	invoiceRepo *repository.InvoiceRepository,
) *VoucherService {
	return &VoucherService{
		journalRepo:        journalRepo,
		templateSvc:        templateSvc,
		bankTxnRepo:        bankTxnRepo,
		paymentRepo:        paymentRepo,
		accountRepo:        accountRepo,
		classificationSvc:  classificationSvc,
		paymentStateMachine: paymentStateMachine,
		invoiceRepo:        invoiceRepo,
	}
}

// CreateVoucherRequest is the request body for creating a voucher.
type CreateVoucherRequest struct {
	VoucherType *string              `json:"voucher_type"`
	PostingDate time.Time            `json:"posting_date"`
	CompanyID   uuid.UUID            `json:"company_id"`
	Remark      *string              `json:"remark,omitempty"`
	Lines       []VoucherLineRequest `json:"lines"`
}

// VoucherLineRequest represents a single line in a voucher creation request.
type VoucherLineRequest struct {
	AccountID    uuid.UUID  `json:"account_id"`
	Debit        string     `json:"debit"`  // decimal as string
	Credit       string     `json:"credit"` // decimal as string
	PartyType    *string    `json:"party_type,omitempty"`
	PartyID      *uuid.UUID `json:"party_id,omitempty"`
	CostCenterID *uuid.UUID `json:"cost_center_id,omitempty"`
	ProjectID    *uuid.UUID `json:"project_id,omitempty"`
	UserRemark   *string    `json:"user_remark,omitempty"`
}

// CreateVoucher creates a new manual voucher with lines (draft status).
func (s *VoucherService) CreateVoucher(ctx context.Context, tenantID, createdBy uuid.UUID, req *CreateVoucherRequest) (*model.JournalEntry, error) {
	if len(req.Lines) == 0 {
		return nil, errors.New("at least one line is required")
	}

	// Generate voucher number
	voucherResp, err := s.templateSvc.GenerateVoucherNumber(ctx, tenantID)
	voucherNo := ""
	if err == nil && voucherResp != nil {
		voucherNo = voucherResp.VoucherNumber
	}
	if voucherNo == "" {
		voucherNo = fmt.Sprintf("MANUAL-%s-%d", time.Now().Format("20060102"), time.Now().UnixNano()%1000000)
	}

	je := &model.JournalEntry{
		ID:          uuid.New(),
		VoucherNo:   voucherNo,
		VoucherType: req.VoucherType,
		PostingDate: req.PostingDate,
		CompanyID:   req.CompanyID,
		TenantID:    tenantID,
		Remark:      req.Remark,
		DocStatus:   0, // draft
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Build lines first (validate amounts before transaction)
	lines := make([]model.JournalEntryLine, 0, len(req.Lines))
	for _, lr := range req.Lines {
		debit, err := decimal.NewFromString(lr.Debit)
		if err != nil {
			return nil, fmt.Errorf("invalid debit amount %q: %w", lr.Debit, err)
		}
		credit, err := decimal.NewFromString(lr.Credit)
		if err != nil {
			return nil, fmt.Errorf("invalid credit amount %q: %w", lr.Credit, err)
		}
		line := model.JournalEntryLine{
			ID:             uuid.New(),
			JournalEntryID: je.ID,
			AccountID:      lr.AccountID,
			Debit:          debit,
			Credit:         credit,
			DebitCcy:       debit,
			CreditCcy:      credit,
			ExchangeRate:   decimal.NewFromInt(1),
			PartyType:      lr.PartyType,
			PartyID:        lr.PartyID,
			CostCenterID:   lr.CostCenterID,
			ProjectID:      lr.ProjectID,
			UserRemark:     lr.UserRemark,
			Reconciled:     false,
			TenantID:       tenantID,
		}
		lines = append(lines, line)
	}

	// Use transaction to ensure atomicity of header + lines creation
	tx, err := s.journalRepo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	created, err := s.journalRepo.CreateTx(ctx, tx, tenantID, je)
	if err != nil {
		return nil, fmt.Errorf("create journal entry: %w", err)
	}

	if _, err = s.journalRepo.AddLinesTx(ctx, tx, tenantID, created.ID, lines); err != nil {
		return nil, fmt.Errorf("add lines: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return created, nil
}

// ListVouchersRequest holds filter params for listing vouchers.
type ListVouchersRequest struct {
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	VoucherType *string    `json:"voucher_type,omitempty"`
	DocStatus   *int16     `json:"doc_status,omitempty"`
	AccountID   *uuid.UUID `json:"account_id,omitempty"` // filter by line account
	Limit       int        `json:"limit,omitempty"`
	Offset      int        `json:"offset,omitempty"`
}

// VoucherDetail is a journal entry with its lines.
type VoucherDetail struct {
	JournalEntry            model.JournalEntry             `json:"journal_entry"`
	JournalEntryLines       []model.JournalEntryLine       `json:"journal_entry_lines"`
	VoucherStateTransitions []model.VoucherStateTransition `json:"voucher_state_transitions,omitempty"`
}

// ListVouchers lists vouchers with optional filters.
func (s *VoucherService) ListVouchers(ctx context.Context, tenantID uuid.UUID, req *ListVouchersRequest) ([]model.JournalEntry, error) {
	limit := 50
	if req.Limit > 0 && req.Limit <= 200 {
		limit = req.Limit
	}
	return s.journalRepo.ListVouchers(ctx, tenantID, req.StartDate, req.EndDate, req.VoucherType, req.DocStatus, req.AccountID, limit, req.Offset)
}

// GetVoucher retrieves a voucher with its lines.
func (s *VoucherService) GetVoucher(ctx context.Context, tenantID, voucherID uuid.UUID) (*VoucherDetail, error) {
	je, err := s.journalRepo.GetByID(ctx, tenantID, voucherID)
	if err != nil {
		return nil, fmt.Errorf("get voucher: %w", err)
	}

	lines, err := s.journalRepo.GetLines(ctx, tenantID, voucherID)
	if err != nil {
		return nil, fmt.Errorf("get lines: %w", err)
	}

	transitions, err := s.journalRepo.GetTransitions(ctx, tenantID, voucherID)
	if err != nil {
		return nil, fmt.Errorf("get transitions: %w", err)
	}

	return &VoucherDetail{
		JournalEntry:            *je,
		JournalEntryLines:       lines,
		VoucherStateTransitions: transitions,
	}, nil
}

// UpdateVoucherRequest is the request body for updating a voucher.
type UpdateVoucherRequest struct {
	VoucherType *string              `json:"voucher_type,omitempty"`
	PostingDate *time.Time           `json:"posting_date,omitempty"`
	Remark      *string              `json:"remark,omitempty"`
	Lines       []VoucherLineRequest `json:"lines"`
}

// UpdateVoucher updates a draft voucher and replaces its lines.
func (s *VoucherService) UpdateVoucher(ctx context.Context, tenantID, voucherID, updatedBy uuid.UUID, req *UpdateVoucherRequest) error {
	je, err := s.journalRepo.GetByID(ctx, tenantID, voucherID)
	if err != nil {
		return fmt.Errorf("get voucher: %w", err)
	}

	if je.DocStatus != 0 {
		return errors.New("only draft vouchers can be updated")
	}

	// Update header fields
	if req.VoucherType != nil {
		je.VoucherType = req.VoucherType
	}
	if req.PostingDate != nil {
		je.PostingDate = *req.PostingDate
	}
	if req.Remark != nil {
		je.Remark = req.Remark
	}
	je.UpdatedAt = time.Now()

	// Use transaction for header + lines update
	tx, err := s.journalRepo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.journalRepo.UpdateTx(ctx, tx, tenantID, je); err != nil {
		return fmt.Errorf("update journal entry: %w", err)
	}

	if len(req.Lines) > 0 {
		// Delete existing lines
		if err := s.journalRepo.DeleteLinesTx(ctx, tx, tenantID, voucherID); err != nil {
			return fmt.Errorf("delete existing lines: %w", err)
		}

		// Insert new lines
		lines := make([]model.JournalEntryLine, 0, len(req.Lines))
		for _, lr := range req.Lines {
			debit, err := decimal.NewFromString(lr.Debit)
			if err != nil {
				return fmt.Errorf("invalid debit amount %q: %w", lr.Debit, err)
			}
			credit, err := decimal.NewFromString(lr.Credit)
			if err != nil {
				return fmt.Errorf("invalid credit amount %q: %w", lr.Credit, err)
			}
			line := model.JournalEntryLine{
				ID:             uuid.New(),
				JournalEntryID: voucherID,
				AccountID:      lr.AccountID,
				Debit:          debit,
				Credit:         credit,
				DebitCcy:       debit,
				CreditCcy:      credit,
				ExchangeRate:   decimal.NewFromInt(1),
				PartyType:      lr.PartyType,
				PartyID:        lr.PartyID,
				CostCenterID:   lr.CostCenterID,
				ProjectID:      lr.ProjectID,
				UserRemark:     lr.UserRemark,
				Reconciled:     false,
				TenantID:       tenantID,
			}
			lines = append(lines, line)
		}

		if _, err := s.journalRepo.AddLinesTx(ctx, tx, tenantID, voucherID, lines); err != nil {
			return fmt.Errorf("add lines: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// SuggestAccountsRequest is the body of POST /vouchers/suggest-accounts.
type SuggestAccountsRequest struct {
	Remark       string `json:"remark"`
	Counterparty string `json:"counterparty"`
	Direction    string `json:"direction"` // "in" | "out" | ""
	Amount       string `json:"amount"`
}

// SuggestAccountsResponse is the response.
type SuggestAccountsResponse struct {
	DebitAccount  *SuggestedAccount `json:"debit_account"`
	CreditAccount *SuggestedAccount `json:"credit_account"`
	Matched       bool             `json:"matched"`
	MatchedRule   string           `json:"matched_rule,omitempty"`
}

type SuggestedAccount struct {
	Code string `json:"code"`
	Name string `json:"name"`
	ID   string `json:"id"`
}

// SuggestAccounts recommends debit/credit accounts for a manual voucher
// based on the user's input summary, counterparty, direction and amount.
func (s *VoucherService) SuggestAccounts(ctx context.Context, tenantID uuid.UUID, req *SuggestAccountsRequest) (*SuggestAccountsResponse, error) {
	resp := &SuggestAccountsResponse{}

	direction := req.Direction
	if direction == "" {
		direction = "in"
	}

	result, _ := s.classificationSvc.MatchTransaction(ctx, tenantID, req.Remark, req.Counterparty, direction)
	if result != nil && result.Matched && result.RuleID != nil {
		rule, ruleErr := s.classificationSvc.GetRuleByID(ctx, tenantID, *result.RuleID)
		if ruleErr == nil && rule != nil {
			if rule.DebitAccountID != nil && rule.CreditAccountID != nil {
				if dAcc, _ := s.accountRepo.GetByID(ctx, tenantID, *rule.DebitAccountID); dAcc != nil {
					resp.DebitAccount = &SuggestedAccount{Code: dAcc.Code, Name: dAcc.Name, ID: dAcc.ID.String()}
				}
				if cAcc, _ := s.accountRepo.GetByID(ctx, tenantID, *rule.CreditAccountID); cAcc != nil {
					resp.CreditAccount = &SuggestedAccount{Code: cAcc.Code, Name: cAcc.Name, ID: cAcc.ID.String()}
				}
				if resp.DebitAccount != nil && resp.CreditAccount != nil {
					resp.Matched = true
					if result.RuleName != nil {
						resp.MatchedRule = *result.RuleName
					}
					return resp, nil
				}
			}
		}
	}

	debitCode, creditCode := s.heuristicByKeywords(req.Remark, direction)
	if dAcc, _ := s.accountRepo.GetByCode(ctx, tenantID, debitCode); dAcc != nil {
		resp.DebitAccount = &SuggestedAccount{Code: dAcc.Code, Name: dAcc.Name, ID: dAcc.ID.String()}
	}
	if cAcc, _ := s.accountRepo.GetByCode(ctx, tenantID, creditCode); cAcc != nil {
		resp.CreditAccount = &SuggestedAccount{Code: cAcc.Code, Name: cAcc.Name, ID: cAcc.ID.String()}
	}
	return resp, nil
}

// heuristicByKeywords provides a best-effort fallback when no classification
// rule matches. Returns (debit_account_code, credit_account_code) for the
// given direction. Codes reference parent account codes when sub-accounts
// are not seeded in the chart of accounts.
func (s *VoucherService) heuristicByKeywords(remark, direction string) (string, string) {
	lower := strings.ToLower(remark)
	switch {
	case strings.Contains(lower, "工资") || strings.Contains(lower, "薪资") || strings.Contains(lower, "薪酬"):
		return "2211", "1002"
	case strings.Contains(lower, "差旅"),
		strings.Contains(lower, "办公费") || strings.Contains(lower, "办公"),
		strings.Contains(lower, "招待") || strings.Contains(lower, "餐饮"),
		strings.Contains(lower, "运输") || strings.Contains(lower, "物流"),
		strings.Contains(lower, "服务费"),
		strings.Contains(lower, "费用"):
		return "5601", "1002"
	case strings.Contains(lower, "货款") || strings.Contains(lower, "应收") || strings.Contains(lower, "收款") || direction == "in":
		return "1002", "1122"
	case strings.Contains(lower, "货款") && direction == "out" || strings.Contains(lower, "付款"):
		return "2202", "1002"
	default:
		if direction == "in" {
			return "1002", "1122"
		}
		return "5601", "1002"
	}
}

// DeleteVoucher deletes a voucher and reverts its source document
// status (payment entry DocStatus, voucher link, bank transaction matched
// flag) so the user can regenerate the voucher from scratch.
// Supports deleting vouchers in any status with proper reverse linkage.
func (s *VoucherService) DeleteVoucher(ctx context.Context, tenantID, voucherID, updatedBy uuid.UUID) error {
	je, err := s.journalRepo.GetByID(ctx, tenantID, voucherID)
	if err != nil {
		return fmt.Errorf("get voucher: %w", err)
	}

	linkedTxns, _ := s.bankTxnRepo.FindByMatchedGLEntryID(ctx, tenantID, voucherID)

	tx, err := s.journalRepo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if je.SourceDocType != nil && *je.SourceDocType == "payment_entry" && je.SourceDocID != nil {
		if s.paymentStateMachine != nil {
			if err := s.paymentStateMachine.RollbackOnVoucherDelete(
				ctx, tenantID, *je.SourceDocID, voucherID, updatedBy, je.DocStatus); err != nil {
				return fmt.Errorf("rollback payment entry %s: %w", *je.SourceDocID, err)
			}
		} else {
			pe, perr := s.paymentRepo.GetByID(ctx, tenantID, *je.SourceDocID)
			if perr == nil && pe != nil {
				var targetStatus int16
				switch je.DocStatus {
				case 3:
					targetStatus = int16(model.PaymentStatusApproved)
				case 2:
					targetStatus = int16(model.PaymentStatusSubmitted)
				default:
					targetStatus = int16(model.PaymentStatusDraft)
				}
				pe.DocStatus = targetStatus
				pe.VoucherID = nil
				pe.VoucherNo = nil
				if uerr := s.paymentRepo.Update(ctx, tenantID, pe); uerr != nil {
					return fmt.Errorf("revert payment entry %s: %w", pe.ID, uerr)
				}
			}
		}
	}

	for _, txn := range linkedTxns {
		if err := s.bankTxnRepo.UnlinkVoucher(ctx, tenantID, txn.ID); err != nil {
			return fmt.Errorf("unlink bank txn %s: %w", txn.ID, err)
		}
		if txn.MatchedPaymentEntryID != nil {
			pe, perr := s.paymentRepo.GetByID(ctx, tenantID, *txn.MatchedPaymentEntryID)
			if perr == nil && pe != nil {
				var targetStatus int16
				switch je.DocStatus {
				case 3:
					targetStatus = int16(model.PaymentStatusApproved)
				case 2:
					targetStatus = int16(model.PaymentStatusSubmitted)
				default:
					targetStatus = int16(model.PaymentStatusDraft)
				}
				pe.DocStatus = targetStatus
				pe.VoucherID = nil
				pe.VoucherNo = nil
				if uerr := s.paymentRepo.Update(ctx, tenantID, pe); uerr != nil {
					return fmt.Errorf("revert payment entry %s: %w", pe.ID, uerr)
				}
			}
		}
	}

	if err := s.journalRepo.DeleteLinesTx(ctx, tx, tenantID, voucherID); err != nil {
		return fmt.Errorf("delete lines: %w", err)
	}

	if err := s.journalRepo.DeleteVoucherTx(ctx, tx, tenantID, voucherID); err != nil {
		return fmt.Errorf("delete voucher: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	// Revert invoice docstatus so it can be re-generated
	if je.SourceType == "invoice" && je.SourceInvoiceID != uuid.Nil {
		if err := s.invoiceRepo.UpdateFields(ctx, tenantID, je.SourceInvoiceID, map[string]interface{}{
			"docstatus": 0,
		}); err != nil {
			return fmt.Errorf("revert invoice docstatus on voucher delete: %w", err)
		}
	}

	return nil
}

// RevertSourceOnVoucherReject reverts the source document's docstatus when
// a voucher is rejected/returned, allowing the source to re-generate a voucher.
// Note: Only invoice source docstatus is reverted on reject. Payment entry status
// is independent of voucher status and is NOT affected by reject.
func (s *VoucherService) RevertSourceOnVoucherReject(ctx context.Context, tenantID, voucherID uuid.UUID) error {
	je, err := s.journalRepo.GetByID(ctx, tenantID, voucherID)
	if err != nil {
		return fmt.Errorf("get voucher for revert: %w", err)
	}

	if je.SourceType == "invoice" && je.SourceInvoiceID != uuid.Nil {
		if err := s.invoiceRepo.UpdateFields(ctx, tenantID, je.SourceInvoiceID, map[string]interface{}{
			"docstatus": 0,
		}); err != nil {
			return fmt.Errorf("revert invoice docstatus on voucher reject: %w", err)
		}
	}

	return nil
}
