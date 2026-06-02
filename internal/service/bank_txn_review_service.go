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

// BankTxnReviewService handles the atomic review workflow for bank transactions.
// It coordinates SubmitReview (approve → generate voucher or payment) and
// RejectManual (send back to manual_pending) operations within a single DB txn.
type BankTxnReviewService struct {
	repo           *repository.BankTransactionRepository
	voucherAutoSvc *VoucherAutoGenerateService
	paymentSvc     *PaymentEntryService
	aiSvc          *BankTxnAIService
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
	Outcome   string     `json:"outcome"` // "voucher_generated" | "payment_created" | "skipped"
	VoucherID *uuid.UUID `json:"voucher_id,omitempty"`
	PaymentID *uuid.UUID `json:"payment_id,omitempty"`
	Reason    string     `json:"reason,omitempty"`
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
	aiSvc *BankTxnAIService,
) *BankTxnReviewService {
	return &BankTxnReviewService{
		pool:           pool,
		repo:           repo,
		voucherAutoSvc: voucherAutoSvc,
		paymentSvc:     paymentSvc,
		aiSvc:          aiSvc,
	}
}

// SubmitReview atomically reviews one or more classified bank transactions.
// For A-class transactions it generates a voucher (docstatus=1); for B-class
// transactions it generates a payment entry (confirmed). Both update the txn
// status and set matched=true — all inside a single DB transaction.
//
// AC5: the status change, matched flag, and generated document are committed
// together; AC6: any error triggers a full rollback with no dirty data.
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
		case "A":
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

		case "B":
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
				Reason: fmt.Sprintf("classification %q is C-class (unhandled)", classification),
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
// If the transaction has no AI analysis yet (ai_confidence=0), it calls the
// BankTxnAIService to analyse it via DeepSeek before returning the preview.
func (s *BankTxnReviewService) PreviewDraft(ctx context.Context, tenantID, txnID uuid.UUID) (*model.JournalEntry, error) {
	txn, err := s.repo.GetByID(ctx, tenantID, txnID)
	if err != nil {
		return nil, fmt.Errorf("get txn: %w", err)
	}

	// If status is still pending, the transaction has not been analysed yet.
	// Trigger AI analysis before proceeding.
	if txn.Status != nil && *txn.Status == string(model.BankTxnReviewStatusPending) {
		if s.aiSvc != nil && (txn.AIConfidence == nil || *txn.AIConfidence == 0) {
			aiResult, aiErr := s.aiSvc.AnalyzeBankTxn(ctx, *txn)
			if aiErr != nil {
				return nil, fmt.Errorf("ai analysis: %w", aiErr)
			}
			// Persist AI result into the transaction record.
			txn.AIConfidence = &aiResult.Confidence
			txn.AISuggestedAction = &aiResult.SuggestedAction
			txn.AIBusinessScene = &aiResult.BusinessScene

			newStatus := string(model.BankTxnReviewStatusClassified)
			txn.Status = &newStatus
			_ = s.repo.UpdateAIFields(ctx, txnID, aiResult.BusinessScene, aiResult.SuggestedAction, aiResult.Confidence)
		}
	}

	if txn.Status != nil && *txn.Status != string(model.BankTxnReviewStatusClassified) {
		return nil, fmt.Errorf("txn is not in classified status")
	}

	classification := ""
	if txn.Classification != nil {
		classification = *txn.Classification
	}
	typ := classifyType(classification)
	if typ != "A" && typ != "B" {
		return nil, nil // C-class — no preview available
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
// classification string. The A set are bank/tax/social/interest/insurance
// categories that map to vouchers; the B set are payment/receipt categories
// that map to payment entries; everything else is C.
func classifyType(classification string) string {
	switch classification {
	case
		"bank_fee", "interest_income", "tax_payment",
		"social_security", "insurance_fee", "internal_transfer":
		return "A"
	case
		"business_receipt", "business_payment",
		"pay", "receive", "expense":
		return "B"
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