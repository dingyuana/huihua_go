package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/model"
	"huihua/finance/internal/repository"
)

// ApInvoiceService provides business logic for accounts payable invoices.
type ApInvoiceService struct {
	repo *repository.ApInvoiceRepository
}

// NewApInvoiceService creates a new ApInvoiceService.
func NewApInvoiceService(repo *repository.ApInvoiceRepository) *ApInvoiceService {
	return &ApInvoiceService{repo: repo}
}

// GetByID retrieves an ApInvoice by ID.
func (s *ApInvoiceService) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*model.ApInvoice, error) {
	return s.repo.GetByID(ctx, tenantID, id)
}

// ListByTenant returns all ApInvoices for a tenant, optionally filtered by status.
func (s *ApInvoiceService) ListByTenant(ctx context.Context, tenantID uuid.UUID, status *string) ([]*model.ApInvoice, error) {
	return s.repo.ListByTenant(ctx, tenantID, status)
}

// ListBySupplier returns ApInvoices for a specific supplier, optionally filtered by status.
func (s *ApInvoiceService) ListBySupplier(ctx context.Context, tenantID, supplierID uuid.UUID, status *string) ([]*model.ApInvoice, error) {
	return s.repo.ListBySupplier(ctx, tenantID, supplierID, status)
}

// ListOutstanding returns ApInvoices with outstanding balance > 0 for the tenant.
func (s *ApInvoiceService) ListOutstanding(ctx context.Context, tenantID uuid.UUID) ([]*model.ApInvoice, error) {
	return s.repo.ListOutstanding(ctx, tenantID)
}

// Create creates a new ApInvoice in draft status.
func (s *ApInvoiceService) Create(ctx context.Context, ap *model.ApInvoice) error {
	if ap.Amount.LessThanOrEqual(decimal.Zero) {
		return errors.New("amount must be positive")
	}
	if ap.InvoiceID == uuid.Nil {
		return errors.New("invoice_id required")
	}
	if ap.SupplierID == uuid.Nil {
		return errors.New("supplier_id required")
	}
	if ap.Status == "" {
		ap.Status = string(model.ApInvoiceStatusDraft)
	}
	if ap.PaidAmount.IsZero() {
		ap.PaidAmount = decimal.Zero
	}
	ap.OutstandingAmount = ap.Amount.Sub(ap.PaidAmount)
	if ap.ID == uuid.Nil {
		ap.ID = uuid.New()
	}
	if ap.CreatedAt.IsZero() {
		ap.CreatedAt = time.Now()
	}
	return s.repo.Create(ctx, ap)
}

// Update updates editable fields on a draft ApInvoice.
func (s *ApInvoiceService) Update(ctx context.Context, tenantID, id uuid.UUID, ap *model.ApInvoice) error {
	existing, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("load ap_invoice: %w", err)
	}
	if existing == nil {
		return errors.New("ap_invoice not found")
	}
	if existing.Status != string(model.ApInvoiceStatusDraft) {
		return fmt.Errorf("only draft ap_invoice can be updated (current: %s)", existing.Status)
	}
	if ap.Amount.LessThanOrEqual(decimal.Zero) {
		return errors.New("amount must be positive")
	}
	ap.OutstandingAmount = ap.Amount.Sub(existing.PaidAmount)
	return s.repo.Update(ctx, tenantID, id, ap)
}

// Confirm marks a draft ApInvoice as confirmed.
func (s *ApInvoiceService) Confirm(ctx context.Context, tenantID, id, userID uuid.UUID) error {
	ap, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("load ap_invoice: %w", err)
	}
	if ap == nil {
		return errors.New("ap_invoice not found")
	}
	if ap.Status != string(model.ApInvoiceStatusDraft) {
		return fmt.Errorf("cannot confirm ap_invoice with status %s", ap.Status)
	}
	return s.repo.Confirm(ctx, tenantID, id, userID)
}

// Approve marks a confirmed ApInvoice as approved (ready for payment).
// In the current schema, the "approved" status is stored as 'confirmed' with
// approved_at/approved_by populated, since the lifecycle model
// draft -> confirmed -> partially_paid/paid does not require a distinct
// "approved" state. Adjust if a future schema adds an 'approved' status.
func (s *ApInvoiceService) Approve(ctx context.Context, tenantID, id, userID uuid.UUID) error {
	ap, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("load ap_invoice: %w", err)
	}
	if ap == nil {
		return errors.New("ap_invoice not found")
	}
	if ap.Status != string(model.ApInvoiceStatusDraft) && ap.Status != string(model.ApInvoiceStatusConfirmed) {
		return fmt.Errorf("cannot approve ap_invoice with status %s", ap.Status)
	}
	if ap.ConfirmedAt == nil {
		// Auto-confirm first when approve is called on a draft.
		if err := s.repo.Confirm(ctx, tenantID, id, userID); err != nil {
			return err
		}
	}
	return s.repo.Approve(ctx, tenantID, id, userID)
}

// Delete removes an ApInvoice (only allowed for draft status).
func (s *ApInvoiceService) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	ap, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("load ap_invoice: %w", err)
	}
	if ap == nil {
		return errors.New("ap_invoice not found")
	}
	if ap.Status != string(model.ApInvoiceStatusDraft) {
		return errors.New("only draft ap_invoice can be deleted")
	}
	return s.repo.Delete(ctx, tenantID, id)
}

// Allocate applies a payment amount to an ApInvoice, updating paid/outstanding and status.
// This is the inverse of RecordPayment on ArInvoiceService: it represents a payment
// made to the supplier that reduces the AP outstanding balance. The optional
// paymentEntryID links the allocation to an originating payment entry (if any).
func (s *ApInvoiceService) Allocate(ctx context.Context, tenantID, id uuid.UUID, paymentAmount decimal.Decimal) error {
	ap, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("load ap_invoice: %w", err)
	}
	if ap == nil {
		return errors.New("ap_invoice not found")
	}
	if paymentAmount.LessThanOrEqual(decimal.Zero) {
		return errors.New("payment amount must be positive")
	}
	if paymentAmount.GreaterThan(ap.OutstandingAmount) {
		return fmt.Errorf("payment amount %s exceeds outstanding %s",
			paymentAmount.String(), ap.OutstandingAmount.String())
	}

	newOutstanding := ap.OutstandingAmount.Sub(paymentAmount)
	var newStatus string
	if newOutstanding.LessThanOrEqual(decimal.Zero) {
		newStatus = string(model.ApInvoiceStatusPaid)
	} else {
		newStatus = string(model.ApInvoiceStatusPartiallyPaid)
	}
	return s.repo.IncrementPaid(ctx, tenantID, id, paymentAmount, newStatus)
}
