package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/model"
	"huihua/finance/internal/repository"
)

// ReconciliationService implements five-level matching between bank transactions and invoices.
type ReconciliationService struct {
	pool            *pgxpool.Pool
	bankTxnRepo     *repository.BankTransactionRepository
	paymentEntryRepo *repository.PaymentEntryRepository
	arInvoiceRepo   *repository.ArInvoiceRepository
	apInvoiceRepo   *repository.ApInvoiceRepository
	invoiceRepo     *repository.InvoiceRepository
	reconRepo       *repository.ReconciliationRepository
	journalRepo     *repository.JournalRepository
}

// UnmatchedSummary contains the summary of unmatched items grouped by counterparty.
type UnmatchedSummary struct {
	TotalUnmatchedAmount decimal.Decimal       `json:"total_unmatched_amount"`
	ByCounterparty       []CounterpartySummary `json:"by_counterparty"`
}

// CounterpartySummary aggregates unmatched amounts for a single counterparty.
type CounterpartySummary struct {
	CounterpartyID   uuid.UUID       `json:"counterparty_id"`
	CounterpartyName string          `json:"counterparty_name"`
	Amount           decimal.Decimal `json:"amount"`
	Count            int             `json:"count"`
}

// ToleranceConfig configures the tolerance for L5 amount matching.
type ToleranceConfig struct {
	Percent decimal.Decimal `json:"percent"` // percentage tolerance, default 10
	Enabled bool            `json:"enabled"`
}

// NewReconciliationService creates a new ReconciliationService.
func NewReconciliationService(
	pool *pgxpool.Pool,
	bankTxnRepo *repository.BankTransactionRepository,
	paymentEntryRepo *repository.PaymentEntryRepository,
	arInvoiceRepo *repository.ArInvoiceRepository,
	apInvoiceRepo *repository.ApInvoiceRepository,
	invoiceRepo *repository.InvoiceRepository,
	reconRepo *repository.ReconciliationRepository,
	journalRepo *repository.JournalRepository,
) *ReconciliationService {
	return &ReconciliationService{
		pool:            pool,
		bankTxnRepo:     bankTxnRepo,
		paymentEntryRepo: paymentEntryRepo,
		arInvoiceRepo:   arInvoiceRepo,
		apInvoiceRepo:   apInvoiceRepo,
		invoiceRepo:     invoiceRepo,
		reconRepo:       reconRepo,
		journalRepo:     journalRepo,
	}
}

// Reconcile runs the five-level matching strategy.
func (s *ReconciliationService) Reconcile(ctx context.Context, tenantID uuid.UUID, periodNo int, tolerance ToleranceConfig) (*model.ReconciliationResult, error) {
	bankTxns, err := s.bankTxnRepo.GetUnmatched(ctx, tenantID, uuid.Nil)
	if err != nil {
		return nil, err
	}
	invoices, err := s.invoiceRepo.ListByTenant(ctx, tenantID, model.InvoiceFilter{Status: "verified"})
	if err != nil {
		return nil, err
	}

	result := &model.ReconciliationResult{TotalScanned: len(bankTxns) + len(invoices)}
	matchedBank := make(map[string]bool)
	matchedInv := make(map[string]bool)

	// L1: bank_txn ID exact match (via reference_no field)
	for _, txn := range bankTxns {
		if matchedBank[txn.ID.String()] {
			continue
		}
		ref := ""
		if txn.ReferenceNo != nil {
			ref = *txn.ReferenceNo
		}
		for _, inv := range invoices {
			if matchedInv[inv.ID.String()] {
				continue
			}
			// Match if txn's reference contains invoice ID
			if ref != "" && strings.Contains(ref, inv.ID.String()) {
				amt := txn.Debit
				if amt.IsZero() {
					amt = txn.Credit
				}
				pair := s.makePair(tenantID, "bank_txn", txn.ID, "invoice", inv.ID, amt, "L1")
				if err := s.reconRepo.Create(ctx, &pair); err == nil {
					result.Pairs = append(result.Pairs, pair)
					matchedBank[txn.ID.String()] = true
					matchedInv[inv.ID.String()] = true
					_ = s.bankTxnRepo.UpdateMatched(ctx, tenantID, txn.ID, true)
					break
				}
			}
		}
	}

	// L2: invoice number match in txn description
	for _, txn := range bankTxns {
		if matchedBank[txn.ID.String()] {
			continue
		}
		desc := ""
		if txn.Description != nil {
			desc = *txn.Description
		}
		for _, inv := range invoices {
			if matchedInv[inv.ID.String()] {
				continue
			}
			if desc != "" && inv.InvoiceNo != "" && strings.Contains(desc, inv.InvoiceNo) {
				amt := txn.Debit
				if amt.IsZero() {
					amt = txn.Credit
				}
				pair := s.makePair(tenantID, "bank_txn", txn.ID, "invoice", inv.ID, amt, "L2")
				if err := s.reconRepo.Create(ctx, &pair); err == nil {
					result.Pairs = append(result.Pairs, pair)
					matchedBank[txn.ID.String()] = true
					matchedInv[inv.ID.String()] = true
					_ = s.bankTxnRepo.UpdateMatched(ctx, tenantID, txn.ID, true)
					break
				}
			}
		}
	}

	// L3: counterparty + amount + date (±3 days)
	for _, txn := range bankTxns {
		if matchedBank[txn.ID.String()] {
			continue
		}
		txnAmt := txn.Debit
		if txnAmt.IsZero() {
			txnAmt = txn.Credit
		}
		for _, inv := range invoices {
			if matchedInv[inv.ID.String()] {
				continue
			}
			txnParty := ""
			if txn.CounterpartyName != nil {
				txnParty = *txn.CounterpartyName
			}
			// Match by counterparty name presence and same amount/date
			if txnAmt.Equal(inv.TotalAmount) &&
				txnParty != "" &&
				withinDays(txn.TxnDate, inv.PostingDate, 3) {
				pair := s.makePair(tenantID, "bank_txn", txn.ID, "invoice", inv.ID, txnAmt, "L3")
				if err := s.reconRepo.Create(ctx, &pair); err == nil {
					result.Pairs = append(result.Pairs, pair)
					matchedBank[txn.ID.String()] = true
					matchedInv[inv.ID.String()] = true
					_ = s.bankTxnRepo.UpdateMatched(ctx, tenantID, txn.ID, true)
					break
				}
			}
		}
	}

	// L4: amount exact + no counterparty needed
	for _, txn := range bankTxns {
		if matchedBank[txn.ID.String()] {
			continue
		}
		txnAmt := txn.Debit
		if txnAmt.IsZero() {
			txnAmt = txn.Credit
		}
		for _, inv := range invoices {
			if matchedInv[inv.ID.String()] {
				continue
			}
			txnParty := ""
			if txn.CounterpartyName != nil {
				txnParty = *txn.CounterpartyName
			}
			hasCounterparty := txnParty != "" && inv.CustomerID != uuid.Nil
			if txnAmt.Equal(inv.TotalAmount) && !hasCounterparty {
				pair := s.makePair(tenantID, "bank_txn", txn.ID, "invoice", inv.ID, txnAmt, "L4")
				if err := s.reconRepo.Create(ctx, &pair); err == nil {
					result.Pairs = append(result.Pairs, pair)
					matchedBank[txn.ID.String()] = true
					matchedInv[inv.ID.String()] = true
					_ = s.bankTxnRepo.UpdateMatched(ctx, tenantID, txn.ID, true)
					break
				}
			}
		}
	}

	// L5: partial amount (≤tolerance% difference, split match)
	for _, txn := range bankTxns {
		if matchedBank[txn.ID.String()] {
			continue
		}
		txnAmt := txn.Debit
		if txnAmt.IsZero() {
			txnAmt = txn.Credit
		}
		for _, inv := range invoices {
			if matchedInv[inv.ID.String()] {
				continue
			}
			smaller := txnAmt
			larger := inv.TotalAmount
			if inv.TotalAmount.LessThan(txnAmt) {
				smaller, larger = inv.TotalAmount, txnAmt
			}
			diff := larger.Sub(smaller)
			threshold := decimal.NewFromInt(10)
			if tolerance.Enabled && tolerance.Percent.GreaterThan(decimal.Zero) {
				threshold = tolerance.Percent
			}
			thresholdAmt := larger.Mul(threshold).Div(decimal.NewFromInt(100))
			if diff.LessThanOrEqual(thresholdAmt) && smaller.GreaterThan(decimal.Zero) {
				pair := s.makePair(tenantID, "bank_txn", txn.ID, "invoice", inv.ID, smaller, "L5")
				if err := s.reconRepo.Create(ctx, &pair); err == nil {
					result.Pairs = append(result.Pairs, pair)
					matchedBank[txn.ID.String()] = true
					matchedInv[inv.ID.String()] = true
					_ = s.bankTxnRepo.UpdateMatched(ctx, tenantID, txn.ID, true)
					break
				}
			}
		}
	}

	result.Matched = len(result.Pairs)
	result.Unmatched = result.TotalScanned - result.Matched

	// Collect unmatched items
	for _, txn := range bankTxns {
		if !matchedBank[txn.ID.String()] {
			amt := txn.Debit
			if amt.IsZero() {
				amt = txn.Credit
			}
			desc := ""
			if txn.Description != nil {
				desc = *txn.Description
			}
			party := ""
			if txn.CounterpartyName != nil {
				party = *txn.CounterpartyName
			}
			result.UnmatchedItems = append(result.UnmatchedItems, model.UnmatchedItem{
				Type: "bank_txn", ID: txn.ID, Date: txn.TxnDate,
				Amount: amt, PartyName: party, Summary: desc,
			})
		}
	}
	for _, inv := range invoices {
		if !matchedInv[inv.ID.String()] {
			result.UnmatchedItems = append(result.UnmatchedItems, model.UnmatchedItem{
				Type: "invoice", ID: inv.ID, Date: inv.PostingDate,
				Amount: inv.TotalAmount,
			})
		}
	}
	return result, nil
}

func (s *ReconciliationService) makePair(tenantID uuid.UUID, st string, sid uuid.UUID, tt string, tid uuid.UUID, amt decimal.Decimal, level string) model.ReconciliationPair {
	now := time.Now()
	return model.ReconciliationPair{
		ID: uuid.New(), TenantID: tenantID,
		SourceType: st, SourceID: sid, TargetType: tt, TargetID: tid,
		Amount: amt, Status: "matched", MatchLevel: level, MatchedAt: &now, CreatedAt: now,
	}
}

// ConfirmPair confirms a matched pair.
// For bank_txn pairs: simple status change matched → confirmed.
// For payment_entry pairs: uses repo-level ConfirmPair (handles payment_allocations).
func (s *ReconciliationService) ConfirmPair(ctx context.Context, tenantID, pairID uuid.UUID) error {
	pair, err := s.reconRepo.GetByID(ctx, tenantID, pairID)
	if err != nil {
		return fmt.Errorf("pair not found: %w", err)
	}
	if pair.SourceType == "bank_txn" {
		if pair.Status != "matched" {
			return fmt.Errorf("bank_txn pair status must be matched, got %s", pair.Status)
		}
		return s.reconRepo.UpdateStatus(ctx, tenantID, pairID, "confirmed")
	}
	return s.reconRepo.ConfirmPair(ctx, tenantID, pairID)
}

type ManualMatchRequest struct {
	BankTransactionID uuid.UUID
	Allocations       []ManualAllocation
}

type ManualAllocation struct {
	InvoiceID uuid.UUID
	Amount    decimal.Decimal
}

// PreCheck validates whether a bank transaction and invoice can be reconciled.
// Returns a list of check items with status: passed / warning / blocked.
func (s *ReconciliationService) PreCheck(ctx context.Context, tenantID, paymentID, invoiceID uuid.UUID) ([]model.PreCheckItem, error) {
	checks := make([]model.PreCheckItem, 0, 5)

	// 1. Check bank transaction (payment) exists and is valid
	bankTxn, err := s.bankTxnRepo.GetByID(ctx, tenantID, paymentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			bankTxn = nil
		} else {
			return nil, fmt.Errorf("get bank transaction: %w", err)
		}
	}
	if bankTxn == nil {
		checks = append(checks, model.PreCheckItem{
			ID: "payment_exists", Name: "收款单有效",
			Status: "blocked", Message: "收款单不存在",
		})
	} else {
		if bankTxn.MatchedGLEntryID != nil {
			checks = append(checks, model.PreCheckItem{
				ID: "payment_exists", Name: "收款单有效",
				Status: "blocked", Message: "该收款单已生成凭证，不可核销",
			})
		} else if bankTxn.MatchedPaymentEntryID != nil {
			checks = append(checks, model.PreCheckItem{
				ID: "payment_exists", Name: "收款单有效",
				Status: "warning", Message: "该收款单已有对应收付款单，核销后需关联处理",
			})
		} else if bankTxn.Credit.IsZero() && bankTxn.Debit.IsZero() {
			checks = append(checks, model.PreCheckItem{
				ID: "payment_exists", Name: "收款单有效",
				Status: "blocked", Message: "收款单金额为零，无法核销",
			})
		} else {
			amt := bankTxn.Credit
			if amt.IsZero() {
				amt = bankTxn.Debit
			}
			checks = append(checks, model.PreCheckItem{
				ID: "payment_exists", Name: "收款单有效",
				Status: "passed", Message: "收款单可核销，金额: " + amt.StringFixed(2),
			})
		}
	}

	// 2. Check invoice exists and is valid
	inv, err := s.invoiceRepo.GetByID(ctx, tenantID, invoiceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			inv = nil
		} else {
			return nil, fmt.Errorf("get invoice: %w", err)
		}
	}
	if inv == nil {
		checks = append(checks, model.PreCheckItem{
			ID: "invoice_exists", Name: "发票有效",
			Status: "blocked", Message: "发票不存在",
		})
	} else {
		if inv.IsReversed {
			checks = append(checks, model.PreCheckItem{
				ID: "invoice_exists", Name: "发票有效",
				Status: "blocked", Message: "发票已被红冲，不可核销",
			})
		} else if inv.OutstandingAmount.IsZero() || inv.OutstandingAmount.IsNegative() {
			checks = append(checks, model.PreCheckItem{
				ID: "invoice_exists", Name: "发票有效",
				Status: "blocked", Message: "发票已完全核销或金额异常",
			})
		} else if inv.Status == "reversed" {
			checks = append(checks, model.PreCheckItem{
				ID: "invoice_exists", Name: "发票有效",
				Status: "blocked", Message: "发票状态为已红冲",
			})
		} else {
			checks = append(checks, model.PreCheckItem{
				ID: "invoice_exists", Name: "发票有效",
				Status: "passed", Message: "发票可核销，未核销金额: " + inv.OutstandingAmount.StringFixed(2),
			})
		}
	}

	// 3. Customer/Party match check (only if both exist)
	if bankTxn != nil && inv != nil {
		bankCounterparty := ""
		if bankTxn.CounterpartyName != nil {
			bankCounterparty = *bankTxn.CounterpartyName
		}
		if bankCounterparty != "" && inv.CustomerName != "" &&
			bankCounterparty != inv.CustomerName {
			checks = append(checks, model.PreCheckItem{
				ID: "customer_match", Name: "客商一致性",
				Status: "warning", Message: "收款单对方(" + bankCounterparty + ") 与发票客户(" + inv.CustomerName + ") 不一致，请确认",
			})
		} else {
			checks = append(checks, model.PreCheckItem{
				ID: "customer_match", Name: "客商一致性",
				Status: "passed", Message: "收款单与发票客商信息匹配",
			})
		}
	}

	// 4. Amount check (payment amount >= invoice outstanding)
	if bankTxn != nil && inv != nil && !inv.OutstandingAmount.IsZero() && inv.OutstandingAmount.IsPositive() {
		paymentAmt := bankTxn.Credit
		if paymentAmt.IsZero() {
			paymentAmt = bankTxn.Debit
		}
		if paymentAmt.LessThan(inv.OutstandingAmount) {
			checks = append(checks, model.PreCheckItem{
				ID: "amount_check", Name: "金额匹配",
				Status: "warning",
				Message: "收款金额(" + paymentAmt.StringFixed(2) + ") 小于发票未核销金额(" + inv.OutstandingAmount.StringFixed(2) + ")，可能只能部分核销",
			})
		} else {
			checks = append(checks, model.PreCheckItem{
				ID: "amount_check", Name: "金额匹配",
				Status: "passed", Message: "收款金额充足，可全额核销",
			})
		}
	}

	// 5. Period check — warn if invoice is in a closed period
	if inv != nil {
		periodNo := inv.PostingDate.Year()*100 + int(inv.PostingDate.Month())
		// Quick check: just warn if period seems old
		now := time.Now()
		currentPeriod := now.Year()*100 + int(now.Month())
		if periodNo < currentPeriod-100 {
			checks = append(checks, model.PreCheckItem{
				ID: "period_check", Name: "期间检查",
				Status: "warning", Message: "发票所属期间(" + fmt.Sprintf("%d", periodNo) + ")较早，请确认期间是否已关闭",
			})
		} else {
			checks = append(checks, model.PreCheckItem{
				ID: "period_check", Name: "期间检查",
				Status: "passed", Message: "发票所属期间正常",
			})
		}
	}

	return checks, nil
}

func (s *ReconciliationService) ManualMatch(ctx context.Context, tenantID, userID uuid.UUID, req *ManualMatchRequest) ([]model.ReconciliationPair, error) {
	if req.BankTransactionID == uuid.Nil {
		return nil, errors.New("bank_transaction_id is required")
	}
	if len(req.Allocations) == 0 {
		return nil, errors.New("at least one allocation is required")
	}

	txn, err := s.bankTxnRepo.GetByID(ctx, tenantID, req.BankTransactionID)
	if err != nil {
		return nil, fmt.Errorf("bank transaction not found: %w", err)
	}
	if txn.Matched {
		return nil, errors.New("bank transaction is already matched")
	}

	pairs := make([]model.ReconciliationPair, 0, len(req.Allocations))
	for _, a := range req.Allocations {
		if a.InvoiceID == uuid.Nil {
			continue
		}
		if a.Amount.LessThanOrEqual(decimal.Zero) {
			continue
		}
		pair := s.makePair(tenantID, "bank_txn", req.BankTransactionID, "invoice", a.InvoiceID, a.Amount, "manual")
		now := time.Now()
		pair.MatchedAt = &now
		pair.ConfirmedAt = &now
		pair.Status = "confirmed"
		if err := s.reconRepo.Create(ctx, &pair); err != nil {
			return pairs, fmt.Errorf("create pair: %w", err)
		}
		pairs = append(pairs, pair)
	}

	if err := s.bankTxnRepo.UpdateMatched(ctx, tenantID, req.BankTransactionID, true); err != nil {
		return pairs, fmt.Errorf("mark bank txn matched: %w", err)
	}

	return pairs, nil
}

// UnconfirmPair cancels confirmation.
func (s *ReconciliationService) UnconfirmPair(ctx context.Context, tenantID, pairID uuid.UUID) error {
	return s.reconRepo.UpdateStatus(ctx, tenantID, pairID, "matched")
}

// ExecutePairResult contains the result of executing reconciliation pairs.
type ExecutePairResult struct {
	ExecutedCount int          `json:"executed_count"`
	FailedIDs     []uuid.UUID  `json:"failed_ids,omitempty"`
	Errors        []string     `json:"errors,omitempty"`
}

// ExecutePairs executes confirmed reconciliation pairs in a single transaction.
// For each pair: INSERT payment_allocations, UPDATE ar_invoices paid/outstanding/status,
// UPDATE bank_transactions matched=true, UPDATE reconciliation_pairs status=executed.
func (s *ReconciliationService) ExecutePairs(ctx context.Context, tenantID uuid.UUID, pairIDs []uuid.UUID) (*ExecutePairResult, error) {
	if len(pairIDs) == 0 {
		return &ExecutePairResult{}, nil
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	result := &ExecutePairResult{}

	for _, pairID := range pairIDs {
		// 1. Verify pair exists and is confirmed
		pair, err := s.reconRepo.GetByID(ctx, tenantID, pairID)
		if err != nil {
			result.FailedIDs = append(result.FailedIDs, pairID)
			result.Errors = append(result.Errors, fmt.Sprintf("pair %s not found: %v", pairID, err))
			continue
		}
		if pair.Status != "confirmed" {
			result.FailedIDs = append(result.FailedIDs, pairID)
			result.Errors = append(result.Errors, fmt.Sprintf("pair %s status is %s, expected confirmed", pairID, pair.Status))
			continue
		}

		now := time.Now()
		allocID := uuid.New()

		// 2. Lock invoice FOR UPDATE within the transaction (ar_invoice or ap_invoice)
		if pair.TargetType == "ap_invoice" {
			if err := s.apInvoiceRepo.LockForUpdate(ctx, tx, tenantID, pair.TargetID); err != nil {
				result.FailedIDs = append(result.FailedIDs, pairID)
				result.Errors = append(result.Errors, fmt.Sprintf("lock ap_invoice %s: %v", pair.TargetID, err))
				continue
			}
		} else {
			if err := s.arInvoiceRepo.LockForUpdate(ctx, tx, tenantID, pair.TargetID); err != nil {
				result.FailedIDs = append(result.FailedIDs, pairID)
				result.Errors = append(result.Errors, fmt.Sprintf("lock invoice %s: %v", pair.TargetID, err))
				continue
			}
		}

		// 3. INSERT payment_allocations (only for payment_entry sources; bank_txn pairs
		//    use the reconciliation_pair itself as the allocation record)
		if pair.SourceType != "bank_txn" {
			invoiceType := "ar_invoice"
			if pair.TargetType == "ap_invoice" {
				invoiceType = "ap_invoice"
			}
			_, err = tx.Exec(ctx, `
				INSERT INTO payment_allocations (id, payment_entry_id, invoice_id, invoice_type, allocated_amount, tenant_id, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7)`,
				allocID, pair.SourceID, pair.TargetID, invoiceType, pair.Amount.String(), tenantID, now,
			)
			if err != nil {
				result.FailedIDs = append(result.FailedIDs, pairID)
				result.Errors = append(result.Errors, fmt.Sprintf("insert allocation for %s: %v", pairID, err))
				continue
			}
		}

		// 4. UPDATE invoice paid_amount, outstanding_amount, and status
		if pair.TargetType == "ap_invoice" {
			// For ApInvoice: use UpdateOutstandingAmountTx which handles status changes
			var currentOutstanding decimal.Decimal
			err = tx.QueryRow(ctx, `
				SELECT outstanding_amount FROM ap_invoices
				WHERE tenant_id = $1 AND id = $2 FOR UPDATE`,
				tenantID, pair.TargetID,
			).Scan(&currentOutstanding)
			if err != nil {
				result.FailedIDs = append(result.FailedIDs, pairID)
				result.Errors = append(result.Errors, fmt.Sprintf("select ap_invoice outstanding %s: %v", pair.TargetID, err))
				continue
			}
			newOutstanding := currentOutstanding.Sub(pair.Amount)
			if newOutstanding.LessThan(decimal.Zero) {
				newOutstanding = decimal.Zero
			}
			if err := s.apInvoiceRepo.UpdateOutstandingAmountTx(ctx, tx, tenantID, pair.TargetID, newOutstanding); err != nil {
				result.FailedIDs = append(result.FailedIDs, pairID)
				result.Errors = append(result.Errors, fmt.Sprintf("update ap_invoice %s: %v", pair.TargetID, err))
				continue
			}
		} else {
			_, err = tx.Exec(ctx, `
				UPDATE ar_invoices
				SET paid_amount = paid_amount + $3,
				    outstanding_amount = amount - (paid_amount + $3),
				    last_allocation_at = NOW(),
				    status = CASE
				        WHEN amount - (paid_amount + $3) <= 0 THEN 'paid'
				        WHEN amount - (paid_amount + $3) < amount THEN 'partially_paid'
				        ELSE status
				    END
				WHERE tenant_id = $1 AND id = $2`,
				tenantID, pair.TargetID, pair.Amount.String(),
			)
			if err != nil {
				result.FailedIDs = append(result.FailedIDs, pairID)
				result.Errors = append(result.Errors, fmt.Sprintf("update invoice %s: %v", pair.TargetID, err))
				continue
			}
		}

		// 5. UPDATE bank_transactions matched = true (for bank_txn source pairs)
		if pair.SourceType == "bank_txn" {
			_, err = tx.Exec(ctx, `
				UPDATE bank_transactions SET matched = $3, updated_at = NOW()
				WHERE tenant_id = $1 AND id = $2`,
				tenantID, pair.SourceID, true,
			)
			if err != nil {
				result.FailedIDs = append(result.FailedIDs, pairID)
				result.Errors = append(result.Errors, fmt.Sprintf("update bank txn %s: %v", pair.SourceID, err))
				continue
			}
		}

		// 6. UPDATE reconciliation_pairs status = 'executed'
		if err := s.reconRepo.UpdateStatus(ctx, tenantID, pairID, "executed"); err != nil {
			result.FailedIDs = append(result.FailedIDs, pairID)
			result.Errors = append(result.Errors, fmt.Sprintf("update pair status %s: %v", pairID, err))
			continue
		}

		result.ExecutedCount++
	}

	if len(result.FailedIDs) > 0 {
		tx.Rollback(ctx)
		return result, fmt.Errorf("%d of %d pairs failed: %s",
			len(result.FailedIDs), len(pairIDs), strings.Join(result.Errors, "; "))
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return result, nil
}

// ReversePair reverses an executed reconciliation pair within the same period.
// Reverts: payment_allocations (reversed_at), ar_invoices (paid_amount/outstanding/status),
// bank_transactions (matched=false), reconciliation_pairs (status=reversed).
func (s *ReconciliationService) ReversePair(ctx context.Context, tenantID, pairID uuid.UUID) error {
	// 1. Get pair — must be in executed status
	pair, err := s.reconRepo.GetByID(ctx, tenantID, pairID)
	if err != nil {
		return fmt.Errorf("pair not found: %w", err)
	}
	if pair.Status != "executed" {
		return fmt.Errorf("pair %s status is %s, expected executed to reverse", pairID, pair.Status)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// 2. Check no subsequent allocations on this invoice after this pair's execution
	// (pair.MatchedAt indicates when the pair was originally executed)
	var subsequentCount int
	err = tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM payment_allocations
		WHERE invoice_id = $1
		  AND tenant_id = $2
		  AND created_at > $3
		  AND reversed_at IS NULL`,
		pair.TargetID, tenantID, pair.MatchedAt,
	).Scan(&subsequentCount)
	if err != nil {
		return fmt.Errorf("check subsequent allocations: %w", err)
	}
	if subsequentCount > 0 {
		return fmt.Errorf("cannot reverse: invoice %s has %d subsequent allocation(s) after this pair", pair.TargetID, subsequentCount)
	}

	// 3. Lock invoice FOR UPDATE (ar_invoice or ap_invoice)
	if pair.TargetType == "ap_invoice" {
		if err := s.apInvoiceRepo.LockForUpdate(ctx, tx, tenantID, pair.TargetID); err != nil {
			return fmt.Errorf("lock ap_invoice %s: %w", pair.TargetID, err)
		}
	} else {
		if err := s.arInvoiceRepo.LockForUpdate(ctx, tx, tenantID, pair.TargetID); err != nil {
			return fmt.Errorf("lock ar_invoice %s: %w", pair.TargetID, err)
		}
	}

	// 4. Mark payment_allocations as reversed
	invoiceType := "ar_invoice"
	if pair.TargetType == "ap_invoice" {
		invoiceType = "ap_invoice"
	}
	_, err = tx.Exec(ctx, `
		UPDATE payment_allocations
		SET reversed_at = NOW()
		WHERE payment_entry_id = $1
		  AND invoice_id = $2
		  AND invoice_type = $3
		  AND reversed_at IS NULL`,
		pair.SourceID, pair.TargetID, invoiceType,
	)
	if err != nil {
		return fmt.Errorf("reverse payment allocations: %w", err)
	}

	// 5. Revert invoice paid_amount, outstanding_amount, status
	if pair.TargetType == "ap_invoice" {
		_, err = tx.Exec(ctx, `
			UPDATE ap_invoices
			SET paid_amount = GREATEST(paid_amount - $3, 0),
			    outstanding_amount = outstanding_amount + $3,
			    last_allocation_at = NOW(),
			    status = CASE
			        WHEN outstanding_amount + $3 >= amount THEN 'confirmed'
			        WHEN outstanding_amount + $3 > 0 THEN 'partially_paid'
			        ELSE 'confirmed'
			    END
			WHERE tenant_id = $1 AND id = $2`,
			tenantID, pair.TargetID, pair.Amount.String(),
		)
	} else {
		_, err = tx.Exec(ctx, `
			UPDATE ar_invoices
			SET paid_amount = GREATEST(paid_amount - $3, 0),
			    outstanding_amount = outstanding_amount + $3,
			    last_allocation_at = NOW(),
			    status = CASE
			            WHEN outstanding_amount + $3 >= amount THEN 'confirmed'
			            WHEN outstanding_amount + $3 > 0 THEN 'partially_paid'
			            ELSE 'confirmed'
			        END
			WHERE tenant_id = $1 AND id = $2`,
			tenantID, pair.TargetID, pair.Amount.String(),
		)
	}
	if err != nil {
		return fmt.Errorf("revert invoice %s: %w", pair.TargetID, err)
	}

	// 6. Revert bank_transactions matched = false (for bank_txn source)
	if pair.SourceType == "bank_txn" {
		_, err = tx.Exec(ctx, `
			UPDATE bank_transactions SET matched = $3, updated_at = NOW()
			WHERE tenant_id = $1 AND id = $2`,
			tenantID, pair.SourceID, false,
		)
		if err != nil {
			return fmt.Errorf("revert bank txn %s: %w", pair.SourceID, err)
		}
	}

	// 7. Mark pair as reversed
	if err := s.reconRepo.UpdateStatus(ctx, tenantID, pairID, "reversed"); err != nil {
		return fmt.Errorf("update pair status: %w", err)
	}

	return tx.Commit(ctx)
}

// ForcePass creates a confirmed reconciliation pair, bypassing precheck warnings.
// Calls PreCheck first — if any check is "blocked", force-pass is denied.
// If only "warning" checks exist, creates a confirmed pair with the override reason.
func (s *ReconciliationService) ForcePass(ctx context.Context, tenantID, paymentID, invoiceID uuid.UUID, reason string) (*model.ReconciliationPair, error) {
	// 1. Run PreCheck
	checks, err := s.PreCheck(ctx, tenantID, paymentID, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("precheck failed: %w", err)
	}

	// 2. Check for blocked items — blocked = cannot force-pass
	for _, c := range checks {
		if c.Status == "blocked" {
			return nil, fmt.Errorf("cannot force-pass: check '%s' is blocked: %s", c.Name, c.Message)
		}
	}

	// 3. Determine allocation amount from the matching payment/bank_txn
	bankTxn, err := s.bankTxnRepo.GetByID(ctx, tenantID, paymentID)
	if err != nil {
		return nil, fmt.Errorf("get payment: %w", err)
	}
	amt := bankTxn.Credit
	if amt.IsZero() {
		amt = bankTxn.Debit
	}

	// 4. Create a confirmed pair
	now := time.Now()
	pair := s.makePair(tenantID, "bank_txn", paymentID, "invoice", invoiceID, amt, "manual")
	pair.MatchedAt = &now
	pair.ConfirmedAt = &now
	pair.Status = "confirmed"
	if err := s.reconRepo.Create(ctx, &pair); err != nil {
		return nil, fmt.Errorf("create pair: %w", err)
	}

	return &pair, nil
}

// ListPairs returns all pairs.
func (s *ReconciliationService) ListPairs(ctx context.Context, tenantID uuid.UUID, status string) ([]model.ReconciliationPair, error) {
	return s.reconRepo.ListByTenant(ctx, tenantID, status)
}

// GetUnmatched returns unmatched bank transactions and journal entries.
func (s *ReconciliationService) GetUnmatched(ctx context.Context, tenantID uuid.UUID) ([]model.UnmatchedItem, error) {
	// Get unmatched bank transactions
	bankTxns, err := s.bankTxnRepo.ListUnmatched(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	// Get unmatched journal entries (status = submitted, no bank_txn_id)
	journalEntries, err := s.journalRepo.ListUnmatched(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	items := make([]model.UnmatchedItem, 0, len(bankTxns)+len(journalEntries))
	for _, txn := range bankTxns {
		counterparty := ""
		if txn.CounterpartyName != nil {
			counterparty = *txn.CounterpartyName
		}
		desc := ""
		if txn.Description != nil {
			desc = *txn.Description
		}
		items = append(items, model.UnmatchedItem{
			Type:      "bank_transaction",
			ID:        txn.ID,
			Amount:    txn.Debit.Add(txn.Credit),
			Date:      txn.TxnDate,
			PartyName: counterparty,
			Summary:   desc,
		})
	}
	for _, je := range journalEntries {
		remark := ""
		if je.Remark != nil {
			remark = *je.Remark
		}
		items = append(items, model.UnmatchedItem{
			Type:      "journal_entry",
			ID:        je.ID,
			Amount:    decimal.Zero, // amount requires line aggregation
			Date:      je.PostingDate,
			PartyName: "",
			Summary:   remark,
		})
	}
	return items, nil
}

// GetUnmatchedSummary returns unmatched amounts grouped by counterparty.
func (s *ReconciliationService) GetUnmatchedSummary(ctx context.Context, tenantID uuid.UUID) (*UnmatchedSummary, error) {
	// 1. Read unmatched bank transactions
	bankTxns, err := s.bankTxnRepo.ListUnmatched(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list unmatched bank txns: %w", err)
	}

	// 2. Read AR invoices with outstanding > 0
	arRows, err := s.pool.Query(ctx, `
		SELECT id, customer_id, amount, paid_amount, outstanding_amount
		FROM ar_invoices
		WHERE tenant_id = $1 AND outstanding_amount > 0
		ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("query ar invoices: %w", err)
	}
	defer arRows.Close()

	type arInv struct {
		ID                uuid.UUID
		CustomerID        uuid.UUID
		Amount           decimal.Decimal
		PaidAmount       decimal.Decimal
		OutstandingAmount decimal.Decimal
	}

	var arInvoices []arInv
	for arRows.Next() {
		var inv arInv
		if err := arRows.Scan(&inv.ID, &inv.CustomerID, &inv.Amount, &inv.PaidAmount, &inv.OutstandingAmount); err != nil {
			return nil, err
		}
		arInvoices = append(arInvoices, inv)
	}
	if err := arRows.Err(); err != nil {
		return nil, err
	}

	// 3. Read AP invoices with outstanding > 0
	apInvoices, err := s.apInvoiceRepo.ListOutstanding(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list outstanding ap invoices: %w", err)
	}

	// 4. Aggregate by counterparty
	type cpMapKey struct {
		id   string
		name string
	}

	summaryMap := make(map[cpMapKey]*CounterpartySummary)
	total := decimal.Zero

	// Process bank transactions — keyed by counterparty_name
	for _, txn := range bankTxns {
		name := ""
		if txn.CounterpartyName != nil {
			name = *txn.CounterpartyName
		}
		amt := txn.Debit
		if amt.IsZero() {
			amt = txn.Credit
		}
		key := cpMapKey{id: name, name: name}
		if existing, ok := summaryMap[key]; ok {
			existing.Amount = existing.Amount.Add(amt)
			existing.Count++
		} else {
			summaryMap[key] = &CounterpartySummary{
				CounterpartyID:   uuid.Nil,
				CounterpartyName: name,
				Amount:           amt,
				Count:            1,
			}
		}
		total = total.Add(amt)
	}

	// Process AR invoices — keyed by customer_id
	for _, inv := range arInvoices {
		if inv.OutstandingAmount.GreaterThan(decimal.Zero) {
			idStr := inv.CustomerID.String()
			key := cpMapKey{id: "ar:" + idStr, name: idStr}
			if existing, ok := summaryMap[key]; ok {
				existing.Amount = existing.Amount.Add(inv.OutstandingAmount)
				existing.Count++
			} else {
				summaryMap[key] = &CounterpartySummary{
					CounterpartyID:   inv.CustomerID,
					CounterpartyName: "",
					Amount:           inv.OutstandingAmount,
					Count:            1,
				}
			}
			total = total.Add(inv.OutstandingAmount)
		}
	}

	// Process AP invoices — keyed by supplier_id
	for _, inv := range apInvoices {
		if inv.OutstandingAmount.GreaterThan(decimal.Zero) {
			idStr := inv.SupplierID.String()
			key := cpMapKey{id: "ap:" + idStr, name: idStr}
			if existing, ok := summaryMap[key]; ok {
				existing.Amount = existing.Amount.Add(inv.OutstandingAmount)
				existing.Count++
			} else {
				summaryMap[key] = &CounterpartySummary{
					CounterpartyID:   inv.SupplierID,
					CounterpartyName: "",
					Amount:           inv.OutstandingAmount,
					Count:            1,
				}
			}
			total = total.Add(inv.OutstandingAmount)
		}
	}

	// Convert map to slice
	byCP := make([]CounterpartySummary, 0, len(summaryMap))
	for _, v := range summaryMap {
		byCP = append(byCP, *v)
	}

	// Sort by amount descending
	sort.Slice(byCP, func(i, j int) bool {
		return byCP[i].Amount.GreaterThan(byCP[j].Amount)
	})

	return &UnmatchedSummary{
		TotalUnmatchedAmount: total,
		ByCounterparty:       byCP,
	}, nil
}

func withinDays(a, b time.Time, days int) bool {
	diff := a.Sub(b)
	if diff < 0 {
		diff = -diff
	}
	return diff.Hours() < float64(days)*24
}

// ReconcilePaymentEntry reconciles a single PaymentEntry with outstanding invoices.
// Returns a pending reconciliation pair for human confirmation.
// Returns nil,nil if no match found (not an error — e.g. pure payment with no invoice).
func (s *ReconciliationService) ReconcilePaymentEntry(
	ctx context.Context,
	tenantID, paymentEntryID uuid.UUID,
) (*model.ReconciliationPair, error) {
	// 1. Load PaymentEntry
	var pe model.PaymentEntry
	err := s.pool.QueryRow(ctx, `
		SELECT id, payment_no, payment_type, party_type, party_id, counterparty_name,
			paid_from_id, paid_to_id, paid_amount, received_amount,
			reference_no, reference_date, posting_date,
			company_id, tenant_id, bank_account_id, docstatus, voucher_id, voucher_no, description, payment_method, created_by, created_at
		FROM payment_entries
		WHERE id = $1 AND tenant_id = $2`, paymentEntryID, tenantID,
	).Scan(
		&pe.ID, &pe.PaymentNo, &pe.PaymentType, &pe.PartyType, &pe.PartyID, &pe.CounterpartyName,
		&pe.PaidFromID, &pe.PaidToID, &pe.PaidAmount, &pe.ReceivedAmount,
		&pe.ReferenceNo, &pe.ReferenceDate, &pe.PostingDate,
		&pe.CompanyID, &pe.TenantID, &pe.BankAccountID, &pe.DocStatus, &pe.VoucherID, &pe.VoucherNo, &pe.Description, &pe.PaymentMethod, &pe.CreatedBy, &pe.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("load payment entry: %w", err)
	}

	// Branch on payment type: "payment" (付款单) → ApInvoice, "receipt" (收款单) → ArInvoice
	if pe.PaymentType == "payment" {
		return s.reconcilePaymentWithApInvoices(ctx, tenantID, paymentEntryID, &pe)
	}
	return s.reconcilePaymentWithArInvoices(ctx, tenantID, paymentEntryID, &pe)
}

// reconcilePaymentWithArInvoices matches a receipt payment entry against outstanding ArInvoices.
func (s *ReconciliationService) reconcilePaymentWithArInvoices(
	ctx context.Context,
	tenantID, paymentEntryID uuid.UUID,
	pe *model.PaymentEntry,
) (*model.ReconciliationPair, error) {
	// 2. Query outstanding ArInvoices (outstanding_amount > 0)
	var outstandingInvoices []*model.ArInvoice
	if s.arInvoiceRepo != nil {
		// status nil = all; filter in memory for outstanding > 0
		invoices, err := s.arInvoiceRepo.ListByTenant(ctx, tenantID, nil)
		if err == nil {
			for _, inv := range invoices {
				if inv.OutstandingAmount.GreaterThan(decimal.Zero) {
					outstandingInvoices = append(outstandingInvoices, inv)
				}
			}
		}
	}
	if len(outstandingInvoices) == 0 {
		return nil, nil // no outstanding invoices
	}

	// Collect payment amount
	paymentAmt := pe.PaidAmount
	if paymentAmt.IsZero() && pe.ReceivedAmount != nil {
		paymentAmt = *pe.ReceivedAmount
	}

	// Collect payment description/reference for matching
	peDesc := ""
	if pe.Description != nil {
		peDesc = *pe.Description
	}
	peRef := ""
	if pe.ReferenceNo != nil {
		peRef = *pe.ReferenceNo
	}
	peParty := ""
	if pe.CounterpartyName != nil {
		peParty = *pe.CounterpartyName
	}

	var matchedInv *model.ArInvoice

	// L1: payment entry description/reference contains ArInvoice ID
	for _, inv := range outstandingInvoices {
		if strings.Contains(peRef, inv.ID.String()) || strings.Contains(peDesc, inv.ID.String()) {
			matchedInv = inv
			break
		}
	}

	// L2: invoice_no appears in PaymentEntry description
	if matchedInv == nil {
		for _, inv := range outstandingInvoices {
			if peDesc != "" && inv.InvoiceNo != "" && strings.Contains(peDesc, inv.InvoiceNo) {
				matchedInv = inv
				break
			}
		}
	}

	// L3: counterparty + amount + date (±3 days)
	if matchedInv == nil {
		for _, inv := range outstandingInvoices {
			if paymentAmt.Equal(inv.Amount) &&
				peParty != "" &&
				inv.DueDate != nil &&
				withinDays(pe.PostingDate, *inv.DueDate, 3) {
				matchedInv = inv
				break
			}
		}
	}

	if matchedInv == nil {
		return nil, nil // no match found — not an error
	}

	// 3. Determine allocation amount (min of payment and outstanding)
	allocAmt := paymentAmt
	if allocAmt.GreaterThan(matchedInv.OutstandingAmount) {
		allocAmt = matchedInv.OutstandingAmount
	}

	// 4. Create payment_allocations record (invoice_type = 'ar_invoice')
	allocID := uuid.New()
	now := time.Now()
	_, err := s.pool.Exec(ctx, `
		INSERT INTO payment_allocations (id, payment_entry_id, invoice_id, invoice_type, allocated_amount, tenant_id, created_at)
		VALUES ($1, $2, $3, 'ar_invoice', $4, $5, $6)`,
		allocID, paymentEntryID, matchedInv.ID, allocAmt, tenantID, now,
	)
	if err != nil {
		return nil, fmt.Errorf("create payment allocation: %w", err)
	}

	// 5. Save ReconciliationPair to database
	pair := model.ReconciliationPair{
		ID:         uuid.New(),
		TenantID:   tenantID,
		SourceType: "payment_entry",
		SourceID:   paymentEntryID,
		TargetType: "ar_invoice",
		TargetID:   matchedInv.ID,
		Amount:     allocAmt,
		Status:     "pending",
		MatchLevel: "auto",
		MatchedAt:  &now,
		CreatedAt:  now,
	}
	if err := s.reconRepo.Create(ctx, &pair); err != nil {
		return nil, fmt.Errorf("create reconciliation pair: %w", err)
	}

	return &pair, nil
}

// reconcilePaymentWithApInvoices matches a payment entry (付款单) against outstanding ApInvoices.
func (s *ReconciliationService) reconcilePaymentWithApInvoices(
	ctx context.Context,
	tenantID, paymentEntryID uuid.UUID,
	pe *model.PaymentEntry,
) (*model.ReconciliationPair, error) {
	// 2. Query unpaid ApInvoices for the supplier (pe.PartyID)
	var outstandingInvoices []*model.ApInvoice
	if s.apInvoiceRepo != nil {
		invoices, err := s.apInvoiceRepo.ListUnpaidBySupplier(ctx, tenantID, pe.PartyID)
		if err == nil {
			outstandingInvoices = invoices
		}
	}
	if len(outstandingInvoices) == 0 {
		return nil, nil // no outstanding AP invoices
	}

	// Collect payment amount
	paymentAmt := pe.PaidAmount
	if paymentAmt.IsZero() && pe.ReceivedAmount != nil {
		paymentAmt = *pe.ReceivedAmount
	}

	// Collect payment description/reference for matching
	peDesc := ""
	if pe.Description != nil {
		peDesc = *pe.Description
	}
	peRef := ""
	if pe.ReferenceNo != nil {
		peRef = *pe.ReferenceNo
	}
	peParty := ""
	if pe.CounterpartyName != nil {
		peParty = *pe.CounterpartyName
	}

	var matchedInv *model.ApInvoice

	// L1: payment entry description/reference contains ApInvoice ID
	for _, inv := range outstandingInvoices {
		if strings.Contains(peRef, inv.ID.String()) || strings.Contains(peDesc, inv.ID.String()) {
			matchedInv = inv
			break
		}
	}

	// L2: invoice_no appears in PaymentEntry description
	if matchedInv == nil {
		for _, inv := range outstandingInvoices {
			if peDesc != "" && inv.InvoiceNo != "" && strings.Contains(peDesc, inv.InvoiceNo) {
				matchedInv = inv
				break
			}
		}
	}

	// L3: counterparty + amount + date (±3 days)
	if matchedInv == nil {
		for _, inv := range outstandingInvoices {
			if paymentAmt.Equal(inv.Amount) &&
				peParty != "" &&
				inv.DueDate != nil &&
				withinDays(pe.PostingDate, *inv.DueDate, 3) {
				matchedInv = inv
				break
			}
		}
	}

	if matchedInv == nil {
		return nil, nil // no match found — not an error
	}

	// 3. Determine allocation amount (min of payment and outstanding)
	allocAmt := paymentAmt
	if allocAmt.GreaterThan(matchedInv.OutstandingAmount) {
		allocAmt = matchedInv.OutstandingAmount
	}

	// 4. Create payment_allocations record (invoice_type = 'ap_invoice')
	allocID := uuid.New()
	now := time.Now()
	_, err := s.pool.Exec(ctx, `
		INSERT INTO payment_allocations (id, payment_entry_id, invoice_id, invoice_type, allocated_amount, tenant_id, created_at)
		VALUES ($1, $2, $3, 'ap_invoice', $4, $5, $6)`,
		allocID, paymentEntryID, matchedInv.ID, allocAmt, tenantID, now,
	)
	if err != nil {
		return nil, fmt.Errorf("create payment allocation: %w", err)
	}

	// 5. Save ReconciliationPair to database
	pair := model.ReconciliationPair{
		ID:         uuid.New(),
		TenantID:   tenantID,
		SourceType: "payment_entry",
		SourceID:   paymentEntryID,
		TargetType: "ap_invoice",
		TargetID:   matchedInv.ID,
		Amount:     allocAmt,
		Status:     "pending",
		MatchLevel: "auto",
		MatchedAt:  &now,
		CreatedAt:  now,
	}
	if err := s.reconRepo.Create(ctx, &pair); err != nil {
		return nil, fmt.Errorf("create reconciliation pair: %w", err)
	}

	return &pair, nil
}
