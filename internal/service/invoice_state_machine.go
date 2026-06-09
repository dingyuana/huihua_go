package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/model"
	"huihua/finance/internal/repository"
)

// InvoiceStateMachine handles invoice status transitions.
type InvoiceStateMachine struct {
	invoiceRepo      *repository.InvoiceRepository
	paymentRepo      *repository.PaymentEntryRepository
	auditRepo        *repository.AuditRepository
	settlementLogRepo *repository.SettlementLogRepository
}

// NewInvoiceStateMachine creates a new InvoiceStateMachine.
func NewInvoiceStateMachine(
	invoiceRepo *repository.InvoiceRepository,
	paymentRepo *repository.PaymentEntryRepository,
	auditRepo *repository.AuditRepository,
) *InvoiceStateMachine {
	return &InvoiceStateMachine{
		invoiceRepo: invoiceRepo,
		paymentRepo: paymentRepo,
		auditRepo:   auditRepo,
	}
}

// InjectSettlementLogRepo injects the settlement log repository.
func (s *InvoiceStateMachine) InjectSettlementLogRepo(repo *repository.SettlementLogRepository) {
	s.settlementLogRepo = repo
}

// InvoiceAction represents an action that triggers a status transition.
type InvoiceAction string

const (
	InvoiceActionSubmit      InvoiceAction = "submit"
	InvoiceActionVerify      InvoiceAction = "verify"
	InvoiceActionReject      InvoiceAction = "reject"
	InvoiceActionCancel      InvoiceAction = "cancel"
	InvoiceActionAllocate    InvoiceAction = "allocate"
	InvoiceActionReverse     InvoiceAction = "reverse"
)

// ValidateTransition checks if a state transition is legal.
func (s *InvoiceStateMachine) ValidateTransition(currentStatus string, action InvoiceAction) error {
	switch currentStatus {
	case string(model.InvoiceStatusDraft):
		switch action {
		case InvoiceActionSubmit, InvoiceActionCancel:
			return nil
		default:
			return fmt.Errorf("invalid action %q for status %s", action, currentStatus)
		}
	case string(model.InvoiceStatusSubmitted):
		switch action {
		case InvoiceActionVerify, InvoiceActionReject:
			return nil
		default:
			return fmt.Errorf("invalid action %q for status %s", action, currentStatus)
		}
	case string(model.InvoiceStatusVerified):
		switch action {
		case InvoiceActionAllocate, InvoiceActionCancel:
			return nil
		default:
			return fmt.Errorf("invalid action %q for status %s", action, currentStatus)
		}
	case string(model.InvoiceStatusPartiallyPaid):
		switch action {
		case InvoiceActionAllocate, InvoiceActionReverse:
			return nil
		default:
			return fmt.Errorf("invalid action %q for status %s", action, currentStatus)
		}
	case string(model.InvoiceStatusPaid):
		switch action {
		case InvoiceActionReverse:
			return nil
		default:
			return fmt.Errorf("invalid action %q for status %s", action, currentStatus)
		}
	case string(model.InvoiceStatusInvalid):
		return errors.New("invoice is already invalid")
	default:
		return fmt.Errorf("unknown status: %s", currentStatus)
	}
}

// ExecuteTransition performs a state transition on an invoice.
func (s *InvoiceStateMachine) ExecuteTransition(
	ctx context.Context,
	tenantID, invoiceID, userID uuid.UUID,
	action InvoiceAction,
	userName string,
	reason string,
) error {
	invoice, err := s.invoiceRepo.GetByID(ctx, tenantID, invoiceID)
	if err != nil {
		return fmt.Errorf("get invoice: %w", err)
	}
	if invoice == nil {
		return errors.New("invoice not found")
	}

	if err := s.ValidateTransition(invoice.Status, action); err != nil {
		return fmt.Errorf("invalid transition: %w", err)
	}

	var newStatus string
	switch action {
	case InvoiceActionSubmit:
		newStatus = string(model.InvoiceStatusSubmitted)
	case InvoiceActionVerify:
		newStatus = string(model.InvoiceStatusVerified)
	case InvoiceActionReject:
		newStatus = string(model.InvoiceStatusDraft)
	case InvoiceActionCancel:
		newStatus = string(model.InvoiceStatusInvalid)
	case InvoiceActionAllocate:
		if invoice.OutstandingAmount.IsPositive() {
			newStatus = string(model.InvoiceStatusPartiallyPaid)
		} else {
			newStatus = string(model.InvoiceStatusPaid)
		}
	case InvoiceActionReverse:
		newStatus = string(model.InvoiceStatusUnpaid)
	}

	if err := s.invoiceRepo.UpdateStatus(ctx, tenantID, invoiceID, newStatus); err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	if s.auditRepo != nil {
		changedFields, _ := json.Marshal(map[string][]interface{}{
			"status": {invoice.Status, newStatus},
		})
		auditLog := &model.AuditLog{
			ID:            uuid.New(),
			Action:        "invoice_status_change",
			ObjectType:    "sales_invoice",
			ObjectID:      invoiceID,
			TenantID:      tenantID,
			ActorID:       userID,
			ActorName:     &userName,
			ChangedFields: changedFields,
		}
		if reason != "" {
			metadata, _ := json.Marshal(map[string]string{"reason": reason})
			auditLog.Metadata = metadata
		}
		_ = s.auditRepo.Create(ctx, tenantID, auditLog)
	}

	return nil
}

// RollbackAllocation rolls back invoice allocation when a payment entry is cancelled.
// Runs inside a DB transaction with SELECT FOR UPDATE on the invoice row.
// Writes an immutable reversal settlement log capturing the before/after outstanding balances.
func (s *InvoiceStateMachine) RollbackAllocation(
	ctx context.Context,
	tenantID, invoiceID, userID uuid.UUID,
	allocatedAmount interface{},
) error {
	var amount decimal.Decimal
	switch v := allocatedAmount.(type) {
	case decimal.Decimal:
		amount = v
	case float64:
		amount = decimal.NewFromFloat(v)
	case int:
		amount = decimal.NewFromInt(int64(v))
	default:
		return fmt.Errorf("invalid amount type: %T", allocatedAmount)
	}

	tx, err := s.invoiceRepo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.invoiceRepo.LockInvoiceForUpdate(ctx, tx, tenantID, invoiceID); err != nil {
		return fmt.Errorf("lock invoice: %w", err)
	}

	invoice, err := s.invoiceRepo.GetByIDTx(ctx, tx, tenantID, invoiceID)
	if err != nil {
		return fmt.Errorf("get invoice (tx): %w", err)
	}
	if invoice == nil {
		return errors.New("invoice not found")
	}

	newOutstanding := invoice.OutstandingAmount.Add(amount)

	if err := s.invoiceRepo.UpdateOutstandingAmountTx(ctx, tx, tenantID, invoiceID, newOutstanding.String()); err != nil {
		return fmt.Errorf("update outstanding amount: %w", err)
	}

	var newStatus string
	if newOutstanding.IsPositive() {
		if invoice.Status == string(model.InvoiceStatusPaid) {
			newStatus = string(model.InvoiceStatusPartiallyPaid)
		} else {
			newStatus = invoice.Status
		}
	} else {
		newStatus = string(model.InvoiceStatusUnpaid)
	}

	if newStatus != invoice.Status {
		if err := s.invoiceRepo.UpdateStatusTx(ctx, tx, tenantID, invoiceID, newStatus); err != nil {
			return fmt.Errorf("update status: %w", err)
		}
	}

	if s.settlementLogRepo != nil {
		// Use a manual_reversal source since RollbackAllocation is called from
		// payment-cancellation flows that don't always pass the original payment_allocation ID.
		// The synthetic source_id is the audit user_id so the log is traceable.
		if err := repository.LogWriteOff(
			ctx, tx, s.settlementLogRepo,
			tenantID, uuid.New(), invoiceID,
			model.SettlementLogSourceManualReversal,
			model.SettlementLogDocSalesInvoice,
			model.SettlementLogDirectionDebit,
			amount,
			invoice.OutstandingAmount,
			newOutstanding,
			&userID,
		); err != nil {
			return fmt.Errorf("write reversal settlement log: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	if s.auditRepo != nil {
		changedFields, _ := json.Marshal(map[string][]interface{}{
			"outstanding_amount": {invoice.OutstandingAmount.String(), newOutstanding.String()},
			"status":             {invoice.Status, newStatus},
		})
		metadata, _ := json.Marshal(map[string]string{
			"reason":           "payment cancelled",
			"allocated_amount": amount.String(),
		})
		auditLog := &model.AuditLog{
			ID:            uuid.New(),
			Action:        "invoice_allocation_rollback",
			ObjectType:    "sales_invoice",
			ObjectID:      invoiceID,
			TenantID:      tenantID,
			ActorID:       userID,
			ChangedFields: changedFields,
			Metadata:      metadata,
		}
		_ = s.auditRepo.Create(ctx, tenantID, auditLog)
	}

	return nil
}