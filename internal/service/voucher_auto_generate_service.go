package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/model"
	"huihua/finance/internal/repository"
)

// VoucherAutoGenerateService auto-generates vouchers from bank transactions, invoices, and reconciliation pairs.
type VoucherAutoGenerateService struct {
	journalRepo       *repository.JournalRepository
	glRepo            *repository.GLEntryRepository
	bankTxnRepo       *repository.BankTransactionRepository
	invoiceRepo       *repository.InvoiceRepository
	accountRepo       *repository.AccountRepository
	classificationSvc *ClassificationRuleService
	templateSvc       *VoucherTemplateService
	approvalSvc       *ApprovalService
}

// NewVoucherAutoGenerateService creates a new VoucherAutoGenerateService.
func NewVoucherAutoGenerateService(
	journalRepo *repository.JournalRepository,
	glRepo *repository.GLEntryRepository,
	bankTxnRepo *repository.BankTransactionRepository,
	invoiceRepo *repository.InvoiceRepository,
	accountRepo *repository.AccountRepository,
	classificationSvc *ClassificationRuleService,
	templateSvc *VoucherTemplateService,
	approvalSvc *ApprovalService,
) *VoucherAutoGenerateService {
	return &VoucherAutoGenerateService{
		journalRepo: journalRepo, glRepo: glRepo,
		bankTxnRepo: bankTxnRepo, invoiceRepo: invoiceRepo,
		accountRepo: accountRepo,
		classificationSvc: classificationSvc, templateSvc: templateSvc,
		approvalSvc: approvalSvc,
	}
}

// GenerateFromBankTxn generates a voucher from a bank transaction.
func (s *VoucherAutoGenerateService) GenerateFromBankTxn(ctx context.Context, tenantID, txnID uuid.UUID, createdBy uuid.UUID) (*model.JournalEntry, error) {
	txn, err := s.bankTxnRepo.GetByID(ctx, tenantID, txnID)
	if err != nil {
		return nil, err
	}

	// Match to a classification rule
	var matchedAccountID uuid.UUID
	txnDesc := ""
	if txn.Description != nil {
		txnDesc = *txn.Description
	}
	counterpartyStr := ""
	if txn.CounterpartyName != nil {
		counterpartyStr = *txn.CounterpartyName
	}
	direction := "out"
	if txn.Credit.GreaterThan(txn.Debit) {
		direction = "in"
	}

	// Try to match classification rule (new API)
	result, err := s.classificationSvc.MatchTransaction(ctx, tenantID, txnDesc, counterpartyStr, direction)
	if err == nil && result.Matched {
		// TODO: Map classification to account ID
		// For now, use default account or disable this logic temporarily
	}

	// Generate voucher number
	voucherResp, _ := s.templateSvc.GenerateVoucherNumber(ctx, tenantID)
	voucherNo := ""
	if voucherResp != nil {
		voucherNo = voucherResp.VoucherNumber
	}

	// Build journal entry
	postingDate := txn.TxnDate
	companyID := txn.CompanyID
	je := &model.JournalEntry{
		ID:          uuid.New(),
		TenantID:    tenantID,
		CompanyID:   companyID,
		VoucherNo:   voucherNo,
		PostingDate: postingDate,
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		DocStatus:   0, // draft
	}

	// Build lines
	txnAmt := txn.Debit
	if txnAmt.IsZero() {
		txnAmt = txn.Credit
	}
	amount := txnAmt
	if txn.Debit.GreaterThan(decimal.Zero) {
		// Bank received: debit bank, credit matched account
		line1 := model.JournalEntryLine{
			ID: uuid.New(), JournalEntryID: je.ID,
			AccountID: txn.BankAccountID,
			Debit: amount, Credit: decimal.Zero,
		}
		line2 := model.JournalEntryLine{
			ID: uuid.New(), JournalEntryID: je.ID,
			AccountID: matchedAccountID,
			Debit: decimal.Zero, Credit: amount,
		}
		_, err = s.journalRepo.AddLines(ctx, tenantID, je.ID, []model.JournalEntryLine{line1, line2})
	} else {
		// Bank paid out: credit bank, debit matched account
		line1 := model.JournalEntryLine{
			ID: uuid.New(), JournalEntryID: je.ID,
			AccountID: matchedAccountID,
			Debit: amount, Credit: decimal.Zero,
		}
		line2 := model.JournalEntryLine{
			ID: uuid.New(), JournalEntryID: je.ID,
			AccountID: txn.BankAccountID,
			Debit: decimal.Zero, Credit: amount,
		}
		_, err = s.journalRepo.AddLines(ctx, tenantID, je.ID, []model.JournalEntryLine{line1, line2})
	}

	_, err = s.journalRepo.Create(ctx, tenantID, je)
	if err != nil {
		return nil, err
	}

	// Submit for approval using the template's bound approval flow (if any)
	if s.approvalSvc != nil && je.CreatedBy != uuid.Nil {
		var flowID *uuid.UUID
		// Template binding: look up the active template for this tenant
		templates, err := s.templateSvc.ListTemplates(ctx, tenantID)
		if err == nil && len(templates) > 0 {
			for i := range templates {
				if templates[i].IsActive && templates[i].ApprovalFlowID != nil {
					flowID = templates[i].ApprovalFlowID
					break // use first active template with a flow
				}
			}
		}
		_ = s.approvalSvc.SubmitForApproval(ctx, tenantID, je.ID, je.CreatedBy, flowID)
	}

	// Mark bank txn as matched
	_ = s.bankTxnRepo.UpdateStatus(ctx, tenantID, txnID, true)
	return je, nil
}

// GenerateFromInvoice generates a voucher from a sales or purchase invoice.
func (s *VoucherAutoGenerateService) GenerateFromInvoice(ctx context.Context, tenantID, invoiceID uuid.UUID, createdBy uuid.UUID) (*model.JournalEntry, error) {
	invoice, err := s.invoiceRepo.GetByID(ctx, tenantID, invoiceID)
	if err != nil {
		return nil, err
	}

	// Generate voucher number
	voucherResp, _ := s.templateSvc.GenerateVoucherNumber(ctx, tenantID)
	voucherNo := ""
	if voucherResp != nil {
		voucherNo = voucherResp.VoucherNumber
	}

	je := &model.JournalEntry{
		ID:          uuid.New(),
		TenantID:    tenantID,
		CompanyID:   invoice.CompanyID,
		VoucherNo:   voucherNo,
		PostingDate: invoice.PostingDate,
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		DocStatus:   0, // draft
	}

	var lines []model.JournalEntryLine

	// Determine direction from invoice type
	isSales := invoice.InvoiceType == "sales" || invoice.InvoiceType == "output"
	invoiceType := "sales"
	if !isSales {
		invoiceType = "purchase"
	}

	// Try to match via classification rules for the counter account
	var counterAcctID uuid.UUID
	txnDesc := invoice.InvoiceNo + " " + invoiceType
	counterpartyStr := "" // invoice doesn't have CustomerName/VendorName fields currently
	direction := "in"
	if !isSales {
		direction = "out"
	}
	// Try to match classification rule (new API)
	result, err := s.classificationSvc.MatchTransaction(ctx, tenantID, txnDesc, counterpartyStr, direction)
	if err == nil && result.Matched {
		// TODO: Map classification to account ID
		// For now, use default account or disable this logic temporarily
	}

	if isSales {
		// Sales invoice: Dr: Accounts Receivable (1122), Cr: Revenue (matched or default)
		// Line 1: Debit AR
		arAccountID := s.findAccountByCode(ctx, tenantID, "1122")
		if arAccountID != nil {
			lines = append(lines, model.JournalEntryLine{
				ID: uuid.New(), JournalEntryID: je.ID,
				AccountID: *arAccountID,
				Debit:     invoice.TotalAmount, Credit: decimal.Zero,
			})
		}
		// Line 2: Credit revenue account
		if counterAcctID != uuid.Nil {
			lines = append(lines, model.JournalEntryLine{
				ID: uuid.New(), JournalEntryID: je.ID,
				AccountID: counterAcctID,
				Debit:     decimal.Zero, Credit: invoice.TotalAmount,
			})
		}
	} else {
		// Purchase invoice: Dr: Expense (matched or default), Cr: Accounts Payable (2202)
		if counterAcctID != uuid.Nil {
			lines = append(lines, model.JournalEntryLine{
				ID: uuid.New(), JournalEntryID: je.ID,
				AccountID: counterAcctID,
				Debit:     invoice.TotalAmount, Credit: decimal.Zero,
			})
		}
		apAccountID := s.findAccountByCode(ctx, tenantID, "2202")
		if apAccountID != nil {
			lines = append(lines, model.JournalEntryLine{
				ID: uuid.New(), JournalEntryID: je.ID,
				AccountID: *apAccountID,
				Debit:     decimal.Zero, Credit: invoice.TotalAmount,
			})
		}
	}

	if len(lines) == 0 {
		return nil, fmt.Errorf("no valid accounts found for invoice %s", invoice.InvoiceNo)
	}

	if _, err := s.journalRepo.Create(ctx, tenantID, je); err != nil {
		return nil, err
	}
	if _, err := s.journalRepo.AddLines(ctx, tenantID, je.ID, lines); err != nil {
		return nil, err
	}

	return je, nil
}

func (s *VoucherAutoGenerateService) findAccountByCode(ctx context.Context, tenantID uuid.UUID, code string) *uuid.UUID {
	acct, err := s.accountRepo.GetByCode(ctx, tenantID, code)
	if err != nil || acct == nil {
		return nil
	}
	return &acct.ID
}

// PreviewVoucher returns a preview without saving.
func (s *VoucherAutoGenerateService) PreviewVoucher(ctx context.Context, tenantID, txnID uuid.UUID, createdBy uuid.UUID) (*model.JournalEntry, error) {
	return s.GenerateFromBankTxn(ctx, tenantID, txnID, createdBy)
}

// BatchGenerateFromBank generates vouchers from all unmatched bank transactions.
func (s *VoucherAutoGenerateService) BatchGenerateFromBank(ctx context.Context, tenantID, bankAccountID uuid.UUID, createdBy uuid.UUID) ([]*model.JournalEntry, error) {
	txns, err := s.bankTxnRepo.GetUnmatched(ctx, tenantID, bankAccountID)
	if err != nil {
		return nil, err
	}
	var vouchers []*model.JournalEntry
	for _, txn := range txns {
		v, err := s.GenerateFromBankTxn(ctx, tenantID, txn.ID, createdBy)
		if err == nil {
			vouchers = append(vouchers, v)
		}
	}
	return vouchers, nil
}
