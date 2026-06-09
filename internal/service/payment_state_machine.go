package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"huihua/finance/internal/event"
	"huihua/finance/internal/model"
	"huihua/finance/internal/repository"
)

// PaymentStateMachine handles payment entry status transitions.
type PaymentStateMachine struct {
	paymentRepo      *repository.PaymentEntryRepository
	invoiceRepo      *repository.InvoiceRepository
	auditRepo        *repository.AuditRepository
	settlementLogRepo *repository.SettlementLogRepository
	eventBus          event.Bus
}

// NewPaymentStateMachine creates a new PaymentStateMachine.
func NewPaymentStateMachine(
	paymentRepo *repository.PaymentEntryRepository,
	invoiceRepo *repository.InvoiceRepository,
	auditRepo *repository.AuditRepository,
) *PaymentStateMachine {
	return &PaymentStateMachine{
		paymentRepo: paymentRepo,
		invoiceRepo: invoiceRepo,
		auditRepo:   auditRepo,
	}
}

// InjectSettlementLogRepo injects the settlement log repository.
func (s *PaymentStateMachine) InjectSettlementLogRepo(repo *repository.SettlementLogRepository) {
	s.settlementLogRepo = repo
}

// InjectEventBus injects the event bus. If non-nil, audit log writes inside
// the state machine are also published as AuditLogEvent for any external
// subscribers (notification, analytics, etc.).
func (s *PaymentStateMachine) InjectEventBus(bus event.Bus) {
	s.eventBus = bus
}

// ValidateTransition checks if a state transition is legal.
func (s *PaymentStateMachine) ValidateTransition(currentStatus int16, action model.PaymentAction) error {
	switch currentStatus {
	case int16(model.PaymentStatusDraft):
		switch action {
		case model.PaymentActionSubmit, model.PaymentActionCancel:
			return nil
		default:
			return fmt.Errorf("invalid action %q for status %d", action, currentStatus)
		}
	case int16(model.PaymentStatusSubmitted):
		switch action {
		case model.PaymentActionApprove, model.PaymentActionReject:
			return nil
		default:
			return fmt.Errorf("invalid action %q for status %d", action, currentStatus)
		}
	case int16(model.PaymentStatusApproved):
		switch action {
		case model.PaymentActionGenerateVoucher, model.PaymentActionCancel:
			return nil
		default:
			return fmt.Errorf("invalid action %q for status %d", action, currentStatus)
		}
	case int16(model.PaymentStatusPosted):
		switch action {
		case model.PaymentActionReverse:
			return nil
		default:
			return fmt.Errorf("invalid action %q for status %d", action, currentStatus)
		}
	case int16(model.PaymentStatusCancelled):
		return errors.New("payment entry is already cancelled")
	default:
		return fmt.Errorf("unknown status: %d", currentStatus)
	}
}

// ExecuteTransition performs a state transition on a payment entry.
func (s *PaymentStateMachine) ExecuteTransition(
	ctx context.Context,
	tenantID, paymentID, userID uuid.UUID,
	action model.PaymentAction,
	userName string,
	reason string,
) error {
	payment, err := s.paymentRepo.GetByID(ctx, tenantID, paymentID)
	if err != nil {
		return fmt.Errorf("get payment entry: %w", err)
	}
	if payment == nil {
		return errors.New("payment entry not found")
	}

	if err := s.ValidateTransition(payment.DocStatus, action); err != nil {
		return fmt.Errorf("invalid transition: %w", err)
	}

	var newStatus int16
	switch action {
	case model.PaymentActionSubmit:
		newStatus = int16(model.PaymentStatusSubmitted)
	case model.PaymentActionApprove:
		newStatus = int16(model.PaymentStatusApproved)
	case model.PaymentActionReject:
		newStatus = int16(model.PaymentStatusDraft)
	case model.PaymentActionCancel:
		newStatus = int16(model.PaymentStatusCancelled)
	case model.PaymentActionGenerateVoucher:
		newStatus = int16(model.PaymentStatusPosted)
	case model.PaymentActionReverse:
		newStatus = int16(model.PaymentStatusCancelled)
	}

	if err := s.paymentRepo.UpdateStatus(ctx, tenantID, paymentID, newStatus); err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	if s.auditRepo != nil {
		changedFields, _ := json.Marshal(map[string][]interface{}{
			"docstatus": {payment.DocStatus, newStatus},
		})
		auditLog := &model.AuditLog{
			ID:            uuid.New(),
			Action:        "payment_status_change",
			ObjectType:    "payment_entry",
			ObjectID:      paymentID,
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

		// Also publish as an event for external subscribers (notification,
		// analytics, etc.). The audit row was already written by the call above.
		if s.eventBus != nil {
			_ = s.eventBus.Publish(ctx, event.AuditLogEvent{
				OccurredAt: time.Now(),
				TenantID:   tenantID,
				Action:     "payment_status_change",
				ObjectType: "payment_entry",
				ObjectID:   paymentID,
				ActorID:    userID,
				ActorName:  &userName,
				OldValues:  map[string]interface{}{"docstatus": payment.DocStatus},
				NewValues:  map[string]interface{}{"docstatus": newStatus},
			})
		}
	}

	return nil
}

// RollbackOnVoucherDelete handles the reverse linkage when a voucher is deleted.
// This rolls back the payment entry status and restores invoice allocations.
func (s *PaymentStateMachine) RollbackOnVoucherDelete(
	ctx context.Context,
	tenantID, paymentID, voucherID, userID uuid.UUID,
	voucherDocStatus int16,
) error {
	payment, err := s.paymentRepo.GetByID(ctx, tenantID, paymentID)
	if err != nil {
		return fmt.Errorf("get payment entry: %w", err)
	}
	if payment == nil {
		return errors.New("payment entry not found")
	}

	var targetStatus int16
	switch voucherDocStatus {
	case 3: // voucher was posted -> payment should go back to approved
		targetStatus = int16(model.PaymentStatusApproved)
	case 2: // voucher was verified -> payment should go back to submitted
		targetStatus = int16(model.PaymentStatusSubmitted)
	case 1: // voucher was submitted -> payment should go back to draft
		targetStatus = int16(model.PaymentStatusDraft)
	default:
		targetStatus = int16(model.PaymentStatusDraft)
	}

	tx, err := s.paymentRepo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.paymentRepo.UpdateStatusAndClearVoucherTx(ctx, tx, tenantID, paymentID, targetStatus); err != nil {
		return fmt.Errorf("update payment status: %w", err)
	}

	if err := s.rollbackInvoiceAllocationsTx(ctx, tx, tenantID, paymentID); err != nil {
		return fmt.Errorf("rollback invoice allocations: %w", err)
	}

	if s.auditRepo != nil {
		changedFields, _ := json.Marshal(map[string][]interface{}{
			"docstatus":   {payment.DocStatus, targetStatus},
			"voucher_id":  {voucherID, nil},
			"voucher_no":  {payment.VoucherNo, nil},
		})
		metadata, _ := json.Marshal(map[string]string{
			"reason":        "voucher deleted",
			"voucher_id":    voucherID.String(),
			"voucher_status": fmt.Sprintf("%d", voucherDocStatus),
		})
		auditLog := &model.AuditLog{
			ID:            uuid.New(),
			Action:        "payment_rollback_on_voucher_delete",
			ObjectType:    "payment_entry",
			ObjectID:      paymentID,
			TenantID:      tenantID,
			ActorID:       userID,
			ChangedFields: changedFields,
			Metadata:      metadata,
		}
		if err := s.auditRepo.CreateTx(ctx, tx, tenantID, auditLog); err != nil {
			return fmt.Errorf("record audit log: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (s *PaymentStateMachine) rollbackInvoiceAllocationsTx(
	ctx context.Context,
	tx pgx.Tx,
	tenantID, paymentID uuid.UUID,
) error {
	allocations, err := s.paymentRepo.GetAllocations(ctx, tenantID, paymentID)
	if err != nil {
		return fmt.Errorf("get allocations: %w", err)
	}

	for _, allocation := range allocations {
		// Lock the invoice row first to prevent concurrent settlement from racing.
		if err := s.invoiceRepo.LockInvoiceForUpdate(ctx, tx, tenantID, allocation.InvoiceID); err != nil {
			return fmt.Errorf("lock invoice: %w", err)
		}

		// Re-read inside the transaction (and after the lock) to get fresh outstanding.
		invoice, err := s.invoiceRepo.GetByIDTx(ctx, tx, tenantID, allocation.InvoiceID)
		if err != nil || invoice == nil {
			continue
		}

		newOutstanding := invoice.OutstandingAmount.Add(allocation.AllocatedAmount)
		if err := s.invoiceRepo.UpdateOutstandingAmountTx(ctx, tx, tenantID, allocation.InvoiceID, newOutstanding.String()); err != nil {
			return fmt.Errorf("update invoice outstanding: %w", err)
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
			if err := s.invoiceRepo.UpdateStatusTx(ctx, tx, tenantID, allocation.InvoiceID, newStatus); err != nil {
				return fmt.Errorf("update invoice status: %w", err)
			}
		}

		// Mark allocation as reversed (soft delete via reversed_at column).
		// This preserves the audit trail; the allocation row stays in the table.
		if err := s.invoiceRepo.MarkAllocationReversed(ctx, tx, allocation.ID); err != nil {
			return fmt.Errorf("mark allocation reversed: %w", err)
		}

		// Write immutable reversal settlement log capturing before/after balances.
		if s.settlementLogRepo != nil {
			if err := repository.LogWriteOff(
				ctx, tx, s.settlementLogRepo,
				tenantID, allocation.ID, allocation.InvoiceID,
				model.SettlementLogSourcePaymentAllocation,
				model.SettlementLogDocSalesInvoice,
				model.SettlementLogDirectionDebit,
				allocation.AllocatedAmount,
				invoice.OutstandingAmount,
				newOutstanding,
				nil,
			); err != nil {
				return fmt.Errorf("write reversal settlement log: %w", err)
			}
		}
	}

	return nil
}

// PaymentStatusToString converts a PaymentStatus to string.
func PaymentStatusToString(status model.PaymentStatus) string {
	switch status {
	case model.PaymentStatusDraft:
		return "draft"
	case model.PaymentStatusSubmitted:
		return "submitted"
	case model.PaymentStatusApproved:
		return "approved"
	case model.PaymentStatusPosted:
		return "posted"
	case model.PaymentStatusCancelled:
		return "cancelled"
	default:
		return "draft"
	}
}

// StringToPaymentStatus converts a string to PaymentStatus.
func StringToPaymentStatus(status string) model.PaymentStatus {
	switch status {
	case "draft":
		return model.PaymentStatusDraft
	case "submitted":
		return model.PaymentStatusSubmitted
	case "approved":
		return model.PaymentStatusApproved
	case "posted":
		return model.PaymentStatusPosted
	case "cancelled":
		return model.PaymentStatusCancelled
	default:
		return model.PaymentStatusDraft
	}
}