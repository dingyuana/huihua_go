package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"huihua/finance/internal/model"
	"huihua/finance/internal/repository"
)

// BankTxnReviewService handles the review workflow for bank transactions.
// Transactions are first classified by classifyType(), then routed:
//   - 第一类（A）: 直接制证（银行费用/税费/社保/利息/保险）
//   - 第二类（B）: 生成 PaymentEntry 后制证（收付款往来/内部转账）
//   - C类: status=manual_pending，待处理工作台
type BankTxnReviewService struct {
	repo           *repository.BankTransactionRepository
	voucherAutoSvc *VoucherAutoGenerateService
	paymentSvc     *PaymentEntryService
	pool           *pgxpool.Pool
}

// SubmitReviewResult is returned by SubmitReview.
type SubmitReviewResult struct {
	Results       []TxnResult
	ApprovedCount int
}

// TxnResult describes the outcome of processing a single transaction.
type TxnResult struct {
	TxnID      string     `json:"txn_id"`
	Outcome   string     `json:"outcome"` // "awaiting_approval" | "pending_review" | "skipped"
	VoucherID *uuid.UUID `json:"voucher_id,omitempty"`
	PaymentID *uuid.UUID `json:"payment_id,omitempty"`
	Reason    string     `json:"reason,omitempty"` // reason if skipped/pending
}

// DraftContent carries human-modified draft fields for a specific txn.
type DraftContent struct {
	SuggestedAction string // overrides classification (e.g. "B" forces payment flow)
}

// NewBankTxnReviewService creates a new BankTxnReviewService.
func NewBankTxnReviewService(
	pool *pgxpool.Pool,
	repo *repository.BankTransactionRepository,
	voucherAutoSvc *VoucherAutoGenerateService,
	paymentSvc *PaymentEntryService,
) *BankTxnReviewService {
	return &BankTxnReviewService{
		pool:           pool,
		repo:           repo,
		voucherAutoSvc: voucherAutoSvc,
		paymentSvc:     paymentSvc,
	}
}

// SubmitReview 审核人对银行流水进行审批操作。
// 本函数生成的是草稿状态的业务单据（凭证草稿 docstatus=0 / 付款单草稿 docstatus=0），
// 所有 status 标签（voucher_generated/payment_created）表示"草稿已生成、等待审核"，
// 不是"已完成"。真正的审核（approve/reject）由人执行。
func (s *BankTxnReviewService) SubmitReview(
	ctx context.Context,
	tenantID uuid.UUID,
	txnIDs []string,
	humanModifiedDrafts map[string]*DraftContent,
) (*SubmitReviewResult, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	results := &SubmitReviewResult{Results: make([]TxnResult, 0, len(txnIDs))}

	for _, txnIDStr := range txnIDs {
		txnID, parseErr := uuid.Parse(txnIDStr)
		if parseErr != nil {
			results.Results = append(results.Results, TxnResult{
				TxnID: txnIDStr, Outcome: "skipped", Reason: "invalid uuid",
			})
			continue
		}

		txn, err := s.repo.GetByIDForUpdate(ctx, tx, txnID)
		if err != nil {
			results.Results = append(results.Results, TxnResult{
				TxnID: txnIDStr, Outcome: "skipped", Reason: fmt.Sprintf("get txn: %v", err),
			})
			continue
		}

		// Must be in classified status to be submitted
		if txn.Status == nil || *txn.Status != string(model.BankTxnReviewStatusClassified) {
			results.Results = append(results.Results, TxnResult{
				TxnID: txnIDStr, Outcome: "skipped",
				Reason: fmt.Sprintf("status is %v, expected classified", ptrStrVal(txn.Status)),
			})
			continue
		}

		// Determine effective classification (human override wins)
		classification := ""
		if txn.Classification != nil {
			classification = *txn.Classification
		}
		if draft := humanModifiedDrafts[txnIDStr]; draft != nil && draft.SuggestedAction != "" {
			classification = draft.SuggestedAction
		}

		switch classifyType(classification) {
		// 第一类直接制证：银行费用/税费/社保/利息/保险
		case "第一类":
			// 凭证草稿已生成，等人审核
			voucher, err := s.voucherAutoSvc.GenerateFromBankTxn(ctx, tenantID, txnID, uuid.Nil)
			if err != nil {
				return nil, fmt.Errorf("generate voucher for %s: %w", txnIDStr, err)
			}
			txn.Matched = true
			txn.MatchedGLEntryID = &voucher.ID
			newStatus := string(model.BankTxnReviewStatusVoucherGenerated)
			txn.Status = &newStatus
			if err := s.updateTxnFields(ctx, tx, txn); err != nil {
				return nil, fmt.Errorf("update txn %s: %w", txnIDStr, err)
			}
			results.Results = append(results.Results, TxnResult{
				TxnID: txnIDStr, Outcome: "voucher_generated", VoucherID: &voucher.ID,
			})

		// 第二类需中转（生成 PaymentEntry）
		case "第二类":
			// 付款单草稿已生成，等人审核
			paymentType := "pay"
			if txn.Direction != nil && *txn.Direction == "in" {
				paymentType = "receive"
			}
			req := &CreatePaymentFromBankTxnRequest{
				BankTransactionID: txnID,
				PaymentType:       paymentType,
				PartyType:         "",
				PartyID:           uuid.Nil,
				CounterpartyName:  txn.CounterpartyName,
				PostingDate:       txn.TxnDate,
				ReferenceNo:       ptrStrVal(txn.ReferenceNo),
			}
			payment, err := s.paymentSvc.CreateFromBankTransaction(ctx, tenantID, uuid.Nil, req, txn, txn.CompanyID)
			if err != nil {
				return nil, fmt.Errorf("generate payment for %s: %w", txnIDStr, err)
			}
			txn.Matched = true
			txn.MatchedPaymentEntryID = &payment.ID
			newStatus := string(model.BankTxnReviewStatusPaymentCreated)
			txn.Status = &newStatus
			if err := s.updateTxnFields(ctx, tx, txn); err != nil {
				return nil, fmt.Errorf("update txn %s: %w", txnIDStr, err)
			}
			results.Results = append(results.Results, TxnResult{
				TxnID: txnIDStr, Outcome: "payment_created", PaymentID: &payment.ID,
			})

		default:
			// C-class or unknown — skip
			results.Results = append(results.Results, TxnResult{
				TxnID: txnIDStr, Outcome: "skipped",
				Reason: fmt.Sprintf("classification %q needs manual handling", classification),
			})
			continue
		}

		results.ApprovedCount++
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return results, nil
}

// RejectManual moves one or more transactions back to manual_pending status (AC7).
func (s *BankTxnReviewService) RejectManual(ctx context.Context, txnIDs []string) error {
	for _, txnIDStr := range txnIDs {
		txnID, err := uuid.Parse(txnIDStr)
		if err != nil {
			continue
		}
		if err := s.repo.UpdateStatus(ctx, txnID, uuid.Nil, model.BankTxnReviewStatusManualPending); err != nil {
			return fmt.Errorf("reject txn %s: %w", txnIDStr, err)
		}
	}
	return nil
}

// PreviewDraft generates a draft voucher (docstatus=0) for a single classified
// transaction without changing its status. Returns nil for C-class txns (AC3).
func (s *BankTxnReviewService) PreviewDraft(ctx context.Context, tenantID, txnID uuid.UUID) (*model.JournalEntry, error) {
	txn, err := s.repo.GetByID(ctx, tenantID, txnID)
	if err != nil {
		return nil, fmt.Errorf("get txn: %w", err)
	}
	if txn.Status != nil && *txn.Status != string(model.BankTxnReviewStatusClassified) {
		return nil, fmt.Errorf("txn is not in classified status")
	}

	classification := ""
	if txn.Classification != nil {
		classification = *txn.Classification
	}
	typ := classifyType(classification)
	if typ != "第一类" && typ != "第二类" {
		return nil, nil // C类 — 无预览
	}

	voucher, err := s.voucherAutoSvc.GenerateFromBankTxn(ctx, tenantID, txnID, uuid.Nil)
	if err != nil {
		return nil, fmt.Errorf("preview draft: %w", err)
	}
	return voucher, nil
}

// ---------------------------------------------------------------------------
// Internal helpers

func (s *BankTxnReviewService) updateTxnFields(ctx context.Context, tx pgx.Tx, txn *model.BankTransaction) error {
	query := `
		UPDATE bank_transactions
		SET status = $2, matched = $3, matched_gl_entry_id = $4,
			matched_payment_entry_id = $5, updated_at = NOW()
		WHERE id = $1`
	_, err := tx.Exec(ctx, query,
		txn.ID, ptrStrVal(txn.Status), txn.Matched,
		txn.MatchedGLEntryID, txn.MatchedPaymentEntryID,
	)
	return err
}

// classifyType returns the high-level type ("A", "B", or "C") for a given
// classification string.
//
// 第一类（可直接制证）：信息完整、无歧义，直接生成记账凭证
//   - bank_fee, interest_income, tax_payment, social_security,
//     insurance_fee → A类 → 直接 GenerateFromBankTxn
//
// 第二类（需中转）：
//   - internal_transfer → B类 → 生成 PaymentEntry（内部转账单）→ 直接制证
//   - business_receipt, business_payment, pay, receive, expense → B类
//     → 生成 PaymentEntry → 核销发票（如有）→ 生成凭证
//
// C类：无法自动分类，status=manual_pending，待人工处理工作台
func classifyType(classification string) string {
	switch classification {
	case
		"bank_fee", "interest_income", "tax_payment",
		"social_security", "insurance_fee":
		return "第一类"
	case
		"business_receipt", "business_payment",
		"pay", "receive", "expense", "internal_transfer":
		return "第二类"
	default:
		return "C"
	}
}

func ptrStrVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// ProcessManual 处理 C 类流水（manual_pending），由人选择处理方式。
// action="第一类": 调用 GenerateFromBankTxn 生成凭证草稿
// action="第二类": 调用 CreateFromBankTransaction 生成付款单草稿（payment_type 必填）
// 人是唯一审核主体，系统不猜测分类。
func (s *BankTxnReviewService) ProcessManual(
	ctx context.Context,
	tenantID uuid.UUID,
	txnID string,
	action string,
	paymentType string,
	userID uuid.UUID,
) (*TxnResult, error) {
	id, err := uuid.Parse(txnID)
	if err != nil {
		return &TxnResult{TxnID: txnID, Outcome: "skipped", Reason: "invalid uuid"}, nil
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	txn, err := s.repo.GetByIDForUpdate(ctx, tx, id)
	if err != nil {
		return &TxnResult{TxnID: txnID, Outcome: "skipped", Reason: fmt.Sprintf("get txn: %v", err)}, nil
	}

	if txn.Status == nil || *txn.Status != string(model.BankTxnReviewStatusManualPending) {
		return &TxnResult{TxnID: txnID, Outcome: "skipped",
			Reason: fmt.Sprintf("status is %v, expected manual_pending", ptrStrVal(txn.Status))}, nil
	}

	switch action {
	case "第一类":
		voucher, err := s.voucherAutoSvc.GenerateFromBankTxn(ctx, tenantID, id, uuid.Nil)
		if err != nil {
			return nil, fmt.Errorf("generate voucher for %s: %w", txnID, err)
		}
		txn.Matched = true
		txn.MatchedGLEntryID = &voucher.ID
		newStatus := string(model.BankTxnReviewStatusVoucherGenerated)
		txn.Status = &newStatus
		if err := s.updateTxnFields(ctx, tx, txn); err != nil {
			return nil, fmt.Errorf("update txn %s: %w", txnID, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return nil, fmt.Errorf("commit: %w", err)
		}
		return &TxnResult{TxnID: txnID, Outcome: "voucher_generated", VoucherID: &voucher.ID}, nil

	case "第二类":
		if paymentType == "" {
			return &TxnResult{TxnID: txnID, Outcome: "skipped", Reason: "payment_type required for action 第二类"}, nil
		}
		req := &CreatePaymentFromBankTxnRequest{
			BankTransactionID: id,
			PaymentType:      paymentType,
			PartyType:        "",
			PartyID:          uuid.Nil,
			CounterpartyName: txn.CounterpartyName,
			PostingDate:      txn.TxnDate,
			ReferenceNo:      ptrStrVal(txn.ReferenceNo),
		}
		payment, err := s.paymentSvc.CreateFromBankTransaction(ctx, tenantID, uuid.Nil, req, txn, txn.CompanyID)
		if err != nil {
			return nil, fmt.Errorf("generate payment for %s: %w", txnID, err)
		}
		txn.Matched = true
		txn.MatchedPaymentEntryID = &payment.ID
		newStatus := string(model.BankTxnReviewStatusPaymentCreated)
		txn.Status = &newStatus
		if err := s.updateTxnFields(ctx, tx, txn); err != nil {
			return nil, fmt.Errorf("update txn %s: %w", txnID, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return nil, fmt.Errorf("commit: %w", err)
		}
		return &TxnResult{TxnID: txnID, Outcome: "payment_created", PaymentID: &payment.ID}, nil

	default:
		return &TxnResult{TxnID: txnID, Outcome: "skipped",
			Reason: fmt.Sprintf("unknown action %q", action)}, nil
	}
}