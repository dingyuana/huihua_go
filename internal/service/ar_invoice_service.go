package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/model"
	"huihua/finance/internal/repository"
)

// ArInvoiceService provides business logic for accounts receivable invoices.
type ArInvoiceService struct {
	repo      *repository.ArInvoiceRepository
	invoiceRepo *repository.InvoiceRepository
}

// NewArInvoiceService creates a new ArInvoiceService.
func NewArInvoiceService(
	repo *repository.ArInvoiceRepository,
	invoiceRepo *repository.InvoiceRepository,
) *ArInvoiceService {
	return &ArInvoiceService{
		repo:        repo,
		invoiceRepo: invoiceRepo,
	}
}

// GetByID retrieves an ArInvoice by ID.
func (s *ArInvoiceService) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*model.ArInvoice, error) {
	return s.repo.GetByID(ctx, tenantID, id)
}

// ListByTenant returns all ArInvoices for a tenant, optionally filtered by status.
func (s *ArInvoiceService) ListByTenant(ctx context.Context, tenantID uuid.UUID, status *string) ([]*model.ArInvoice, error) {
	return s.repo.ListByTenant(ctx, tenantID, status)
}

// ListByCustomer returns ArInvoices for a specific customer, optionally filtered by status.
func (s *ArInvoiceService) ListByCustomer(ctx context.Context, tenantID, customerID uuid.UUID, status *string) ([]*model.ArInvoice, error) {
	return s.repo.ListByCustomer(ctx, tenantID, customerID, status)
}

// ListUnpaidByCustomer returns ArInvoices with outstanding balance for a given customer.
func (s *ArInvoiceService) ListUnpaidByCustomer(ctx context.Context, tenantID, customerID uuid.UUID) ([]*model.ArInvoice, error) {
	return s.repo.ListUnpaidByCustomer(ctx, tenantID, customerID)
}

// Confirm marks a draft ArInvoice as confirmed.
func (s *ArInvoiceService) Confirm(ctx context.Context, tenantID, id, userID uuid.UUID) error {
	ar, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("load ar_invoice: %w", err)
	}
	if ar == nil {
		return errors.New("ar_invoice not found")
	}
	if ar.Status != string(model.ArInvoiceStatusDraft) && ar.Status != "confirmed" {
		return fmt.Errorf("cannot confirm ar_invoice with status %s", ar.Status)
	}
	return s.repo.Confirm(ctx, tenantID, id, userID)
}

// Delete removes an ArInvoice (only allowed for draft status).
func (s *ArInvoiceService) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	ar, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("load ar_invoice: %w", err)
	}
	if ar == nil {
		return errors.New("ar_invoice not found")
	}
	if ar.Status != string(model.ArInvoiceStatusDraft) {
		return errors.New("only draft ar_invoice can be deleted")
	}
	return s.repo.Delete(ctx, tenantID, id)
}

// RecordPayment applies a payment amount to an ArInvoice, updating paid/outstanding and status.
func (s *ArInvoiceService) RecordPayment(ctx context.Context, tenantID, id uuid.UUID, paymentAmount decimal.Decimal) error {
	ar, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("load ar_invoice: %w", err)
	}
	if ar == nil {
		return errors.New("ar_invoice not found")
	}
	if paymentAmount.LessThanOrEqual(decimal.Zero) {
		return errors.New("payment amount must be positive")
	}
	if paymentAmount.GreaterThan(ar.OutstandingAmount) {
		return fmt.Errorf("payment amount %s exceeds outstanding %s",
			paymentAmount.String(), ar.OutstandingAmount.String())
	}

	newOutstanding := ar.OutstandingAmount.Sub(paymentAmount)
	var newStatus string
	if newOutstanding.LessThanOrEqual(decimal.Zero) {
		newStatus = string(model.ArInvoiceStatusPaid)
	} else {
		newStatus = string(model.ArInvoiceStatusPartiallyPaid)
	}
	return s.repo.IncrementPaid(ctx, tenantID, id, paymentAmount, newStatus)
}

// CustomerSummary aggregates total, paid, and outstanding amounts for a customer.
type CustomerSummary struct {
	TotalAmount       decimal.Decimal `json:"total_amount"`
	PaidAmount        decimal.Decimal `json:"paid_amount"`
	OutstandingAmount decimal.Decimal `json:"outstanding_amount"`
	InvoiceCount      int             `json:"invoice_count"`
}

// GetCustomerSummary returns a summary of receivables for a customer.
func (s *ArInvoiceService) GetCustomerSummary(ctx context.Context, tenantID, customerID uuid.UUID) (*CustomerSummary, error) {
	list, err := s.repo.ListUnpaidByCustomer(ctx, tenantID, customerID)
	if err != nil {
		return nil, err
	}
	summary := &CustomerSummary{}
	for _, ar := range list {
		summary.TotalAmount = summary.TotalAmount.Add(ar.Amount)
		summary.PaidAmount = summary.PaidAmount.Add(ar.PaidAmount)
		summary.OutstandingAmount = summary.OutstandingAmount.Add(ar.OutstandingAmount)
		summary.InvoiceCount++
	}
	return summary, nil
}
