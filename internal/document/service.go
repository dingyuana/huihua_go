package document

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/model"
)

// ErrInvalidDelta is returned when an outstanding-amount adjustment would
// underflow the balance (e.g. subtracting more than is still outstanding).
var ErrInvalidDelta = errors.New("invalid outstanding delta")

// Service provides cross-cutting operations against any ProcessDocument.
//
// Today the operations are simple: lock the row, adjust the outstanding
// balance, optionally transition status. Centralizing them here means new
// settlement flows (cross-document netting, batch reconciliation) can be
// added without re-implementing the lock + write boilerplate for each kind.
type Service struct{}

// NewService returns a stateless DocumentService.
func NewService() *Service { return &Service{} }

// Settle subtracts delta from the document's outstanding amount, recomputes
// the status, and writes both inside the caller's transaction.
//
// Returns the new outstanding amount on success. Returns ErrInvalidDelta
// if the new outstanding would be negative. The caller is responsible for
// also writing a settlement_log row inside the same transaction.
func (s *Service) Settle(
	ctx context.Context,
	tx pgx.Tx,
	doc Adapter,
	delta decimal.Decimal,
) (decimal.Decimal, error) {
	if !delta.IsPositive() {
		return decimal.Zero, fmt.Errorf("%w: delta must be positive", ErrInvalidDelta)
	}
	locker := doc.AsLocker()
	if err := locker.LockForUpdate(ctx, tx); err != nil {
		return decimal.Zero, fmt.Errorf("lock %s: %w", doc.Kind(), err)
	}
	current := doc.GetOutstandingAmount()
	if current.LessThan(delta) {
		return decimal.Zero, fmt.Errorf("%w: outstanding %s < delta %s",
			ErrInvalidDelta, current.String(), delta.String())
	}
	newOutstanding := current.Sub(delta)
	if err := locker.UpdateOutstanding(ctx, tx, newOutstanding); err != nil {
		return decimal.Zero, fmt.Errorf("update %s outstanding: %w", doc.Kind(), err)
	}
	newStatus := computeStatusOnSettle(doc.Kind(), newOutstanding)
	if newStatus != doc.GetStatus() {
		if err := locker.UpdateStatus(ctx, tx, newStatus); err != nil {
			return decimal.Zero, fmt.Errorf("update %s status: %w", doc.Kind(), err)
		}
	}
	return newOutstanding, nil
}

// Reverse adds delta back to the document's outstanding amount (used during
// settlement rollback). Returns the new outstanding amount.
func (s *Service) Reverse(
	ctx context.Context,
	tx pgx.Tx,
	doc Adapter,
	delta decimal.Decimal,
) (decimal.Decimal, error) {
	if !delta.IsPositive() {
		return decimal.Zero, fmt.Errorf("%w: delta must be positive", ErrInvalidDelta)
	}
	locker := doc.AsLocker()
	if err := locker.LockForUpdate(ctx, tx); err != nil {
		return decimal.Zero, fmt.Errorf("lock %s: %w", doc.Kind(), err)
	}
	current := doc.GetOutstandingAmount()
	newOutstanding := current.Add(delta)
	if err := locker.UpdateOutstanding(ctx, tx, newOutstanding); err != nil {
		return decimal.Zero, fmt.Errorf("update %s outstanding: %w", doc.Kind(), err)
	}
	newStatus := computeStatusOnReverse(doc.Kind(), doc.GetTotalAmount(), newOutstanding, doc.GetStatus())
	if newStatus != doc.GetStatus() {
		if err := locker.UpdateStatus(ctx, tx, newStatus); err != nil {
			return decimal.Zero, fmt.Errorf("update %s status: %w", doc.Kind(), err)
		}
	}
	return newOutstanding, nil
}

// computeStatusOnSettle returns the new status when an outstanding amount
// has been reduced (settlement). If new outstanding is 0 -> paid; otherwise
// partially_paid. The original status is preserved if neither matches.
func computeStatusOnSettle(kind Kind, newOutstanding decimal.Decimal) string {
	if newOutstanding.IsZero() || newOutstanding.IsNegative() {
		switch kind {
		case KindSalesInvoice:
			return string(model.InvoiceStatusPaid)
		case KindArInvoice:
			return string(model.ArInvoiceStatusPaid)
		case KindApInvoice:
			return string(model.ApInvoiceStatusPaid)
		}
	}
	switch kind {
	case KindSalesInvoice:
		return string(model.InvoiceStatusPartiallyPaid)
	case KindArInvoice:
		return string(model.ArInvoiceStatusPartiallyPaid)
	case KindApInvoice:
		return string(model.ApInvoiceStatusPartiallyPaid)
	}
	return ""
}

// computeStatusOnReverse returns the new status when outstanding is restored.
// If new outstanding has reached the total, the document is fully unpaid again.
// If new outstanding is positive but less than total, it stays partially paid
// (or transitions from paid back to partially_paid).
func computeStatusOnReverse(kind Kind, total, newOutstanding decimal.Decimal, currentStatus string) string {
	if newOutstanding.GreaterThanOrEqual(total) {
		switch kind {
		case KindSalesInvoice:
			return string(model.InvoiceStatusUnpaid)
		case KindArInvoice:
			return "unpaid"
		case KindApInvoice:
			return "unpaid"
		}
	}
	if newOutstanding.IsPositive() {
		if currentStatus == string(model.InvoiceStatusPaid) ||
			currentStatus == string(model.ArInvoiceStatusPaid) ||
			currentStatus == string(model.ApInvoiceStatusPaid) {
			switch kind {
			case KindSalesInvoice:
				return string(model.InvoiceStatusPartiallyPaid)
			case KindArInvoice:
				return string(model.ArInvoiceStatusPartiallyPaid)
			case KindApInvoice:
				return string(model.ApInvoiceStatusPartiallyPaid)
			}
		}
		return currentStatus
	}
	return currentStatus
}
