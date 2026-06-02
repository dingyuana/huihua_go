package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/model"
	"huihua/finance/internal/repository"
)

// ReconciliationService implements five-level matching between bank transactions and invoices.
type ReconciliationService struct {
	pool        *pgxpool.Pool
	bankTxnRepo *repository.BankTransactionRepository
	invoiceRepo *repository.InvoiceRepository
	reconRepo   *repository.ReconciliationRepository
	journalRepo *repository.JournalRepository
}

// NewReconciliationService creates a new ReconciliationService.
func NewReconciliationService(
	pool *pgxpool.Pool,
	bankTxnRepo *repository.BankTransactionRepository,
	invoiceRepo *repository.InvoiceRepository,
	reconRepo *repository.ReconciliationRepository,
	journalRepo *repository.JournalRepository,
) *ReconciliationService {
	return &ReconciliationService{
		pool:        pool,
		bankTxnRepo: bankTxnRepo,
		invoiceRepo: invoiceRepo,
		reconRepo:   reconRepo,
		journalRepo: journalRepo,
	}
}

// Reconcile runs the five-level matching strategy.
func (s *ReconciliationService) Reconcile(ctx context.Context, tenantID uuid.UUID, periodNo int) (*model.ReconciliationResult, error) {
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
					break
				}
			}
		}
	}

	// L5: partial amount (≤10% difference, split match)
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
			threshold := larger.Div(decimal.NewFromInt(10))
			if diff.LessThanOrEqual(threshold) && smaller.GreaterThan(decimal.Zero) {
				pair := s.makePair(tenantID, "bank_txn", txn.ID, "invoice", inv.ID, smaller, "L5")
				if err := s.reconRepo.Create(ctx, &pair); err == nil {
					result.Pairs = append(result.Pairs, pair)
					matchedBank[txn.ID.String()] = true
					matchedInv[inv.ID.String()] = true
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
func (s *ReconciliationService) ConfirmPair(ctx context.Context, tenantID, pairID, userID uuid.UUID) error {
	return s.reconRepo.UpdateStatus(ctx, tenantID, pairID, "confirmed")
}

type ManualMatchRequest struct {
	BankTransactionID uuid.UUID
	Allocations       []ManualAllocation
}

type ManualAllocation struct {
	InvoiceID uuid.UUID
	Amount    decimal.Decimal
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

func withinDays(a, b time.Time, days int) bool {
	diff := a.Sub(b)
	if diff < 0 {
		diff = -diff
	}
	return diff.Hours() < float64(days)*24
}
