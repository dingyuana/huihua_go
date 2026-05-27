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

// InvoiceService handles invoice operations.
type InvoiceService struct {
	repo *repository.InvoiceRepository
}

// NewInvoiceService creates a new InvoiceService.
func NewInvoiceService(repo *repository.InvoiceRepository) *InvoiceService {
	return &InvoiceService{repo: repo}
}

// Create creates a new invoice.
func (s *InvoiceService) Create(ctx context.Context, tenantID uuid.UUID, req *model.SalesInvoice) (*model.SalesInvoice, error) {
	// Validate invoice data
	if err := s.ValidateInvoice(req); err != nil {
		return nil, err
	}

	// Check for duplicate invoice number
	exists, err := s.repo.ValidateDuplicateInvoiceNo(ctx, tenantID, req.InvoiceNo)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("invoice number already exists")
	}

	req.Status = string(model.InvoiceStatusDraft)
	return s.repo.Create(ctx, tenantID, req)
}

// List returns invoices for a tenant with optional filters.
func (s *InvoiceService) List(ctx context.Context, tenantID uuid.UUID, filters model.InvoiceFilter) ([]model.SalesInvoice, error) {
	return s.repo.ListByTenant(ctx, tenantID, filters)
}

// GetByID retrieves an invoice by ID.
func (s *InvoiceService) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*model.SalesInvoice, error) {
	return s.repo.GetByID(ctx, tenantID, id)
}

// UpdateStatus updates the status of an invoice.
func (s *InvoiceService) UpdateStatus(ctx context.Context, tenantID, id uuid.UUID, status string) error {
	// Validate status transition
	validStatuses := map[string]bool{
		"draft": true, "submitted": true, "verified": true, "invalid": true,
		"unpaid": true, "partially_paid": true, "paid": true,
	}
	if !validStatuses[status] {
		return errors.New("invalid status")
	}

	inv, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return err
	}
	if inv == nil {
		return errors.New("invoice not found")
	}

	return s.repo.UpdateStatus(ctx, tenantID, id, status)
}

// ValidateInvoice validates invoice data.
func (s *InvoiceService) ValidateInvoice(inv *model.SalesInvoice) error {
	if inv.InvoiceNo == "" {
		return errors.New("invoice_no is required")
	}
	if inv.CustomerID == uuid.Nil {
		return errors.New("customer_id is required")
	}
	if inv.CompanyID == uuid.Nil {
		return errors.New("company_id is required")
	}
	if inv.TotalAmount.LessThan(decimal.Zero) {
		return errors.New("total_amount cannot be negative")
	}
	if inv.NetAmount.LessThan(decimal.Zero) {
		return errors.New("net_amount cannot be negative")
	}
	if inv.TaxAmount.LessThan(decimal.Zero) {
		return errors.New("tax_amount cannot be negative")
	}
	return nil
}

// ValidateLineItems validates that line item amounts are consistent.
func (s *InvoiceService) ValidateLineItems(items []model.InvoiceLineItem) error {
	for _, item := range items {
		// Calculate expected net amount: quantity * unit_price
		expectedNet := item.Quantity.Mul(item.UnitPrice)

		// Verify net amount matches
		if !expectedNet.Equal(item.NetAmount) {
			return fmt.Errorf("line item net amount mismatch: expected %s, got %s",
				expectedNet.String(), item.NetAmount.String())
		}

		// Calculate expected tax amount: net_amount * tax_rate
		taxRate := item.TaxRate.Div(decimal.NewFromInt(100))
		expectedTax := item.NetAmount.Mul(taxRate)

		// Verify tax amount (allow small floating point differences)
		diff := expectedTax.Sub(item.TaxAmount).Abs()
		if diff.GreaterThan(decimal.NewFromFloat(0.01)) {
			return fmt.Errorf("line item tax amount mismatch: expected %s, got %s",
				expectedTax.String(), item.TaxAmount.String())
		}

		// Calculate total amount: net_amount + tax_amount
		expectedTotal := item.NetAmount.Add(item.TaxAmount)
		if !expectedTotal.Equal(item.TotalAmount) {
			return fmt.Errorf("line item total amount mismatch: expected %s, got %s",
				expectedTotal.String(), item.TotalAmount.String())
		}
	}
	return nil
}

// ImportFromExcel imports invoices from Excel data.
func (s *InvoiceService) ImportFromExcel(ctx context.Context, tenantID uuid.UUID, req *model.InvoiceImportRequest) ([]model.SalesInvoice, error) {
	if req == nil || len(req.Invoices) == 0 {
		return nil, errors.New("no invoices to import")
	}

	var invoices []model.SalesInvoice
	for _, item := range req.Invoices {
		// Parse posting date
		postingDate, err := time.Parse("2006-01-02", item.PostingDate)
		if err != nil {
			return nil, fmt.Errorf("invalid posting_date format for invoice %s: %v", item.InvoiceNo, err)
		}

		// Parse optional due date
		var dueDate *time.Time
		if item.DueDate != "" {
			t, err := time.Parse("2006-01-02", item.DueDate)
			if err != nil {
				return nil, fmt.Errorf("invalid due_date format for invoice %s: %v", item.InvoiceNo, err)
			}
			dueDate = &t
		}

		// Parse customer_id
		customerID, err := uuid.Parse(item.CustomerID)
		if err != nil {
			return nil, fmt.Errorf("invalid customer_id for invoice %s: %v", item.InvoiceNo, err)
		}

		// Create invoice model
		inv := &model.SalesInvoice{
			InvoiceNo:         item.InvoiceNo,
			InvoiceType:       item.InvoiceType,
			CustomerID:        customerID,
			PostingDate:       postingDate,
			DueDate:           dueDate,
			TotalAmount:       decimal.NewFromFloat(item.TotalAmount),
			TaxAmount:         decimal.NewFromFloat(item.TaxAmount),
			NetAmount:         decimal.NewFromFloat(item.NetAmount),
			OutstandingAmount: decimal.NewFromFloat(item.TotalAmount),
			Status:            item.Status,
		}

		// Validate
		if err := s.ValidateInvoice(inv); err != nil {
			return nil, fmt.Errorf("validation failed for invoice %s: %v", item.InvoiceNo, err)
		}

		// Check for duplicate
		exists, err := s.repo.ValidateDuplicateInvoiceNo(ctx, tenantID, item.InvoiceNo)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, fmt.Errorf("invoice number %s already exists", item.InvoiceNo)
		}

		invoices = append(invoices, *inv)
	}

	// Batch import
	return s.repo.ImportBatch(ctx, tenantID, invoices)
}

// ParseInvoicePDF is a placeholder for PDF OCR parsing.
func (s *InvoiceService) ParseInvoicePDF(ctx context.Context, tenantID uuid.UUID, fileURL string) (*model.SalesInvoice, error) {
	return nil, errors.New("OCR not implemented, use manual import")
}

// ParseInvoiceImage is a placeholder for image OCR parsing.
func (s *InvoiceService) ParseInvoiceImage(ctx context.Context, tenantID uuid.UUID, fileURL string) (*model.SalesInvoice, error) {
	return nil, errors.New("OCR not implemented, use manual import")
}

// MatchToBankTxn matches an invoice to a bank transaction.
func (s *InvoiceService) MatchToBankTxn(ctx context.Context, tenantID, invoiceID, bankTxnID uuid.UUID, amount decimal.Decimal) error {
	// Get the invoice to verify it exists
	inv, err := s.repo.GetByID(ctx, tenantID, invoiceID)
	if err != nil {
		return fmt.Errorf("invoice not found: %v", err)
	}
	if inv == nil {
		return errors.New("invoice not found")
	}

	// Check that amount doesn't exceed outstanding
	if amount.GreaterThan(inv.OutstandingAmount) {
		return errors.New("allocation amount exceeds outstanding invoice amount")
	}

	// Create the payment allocation
	return s.repo.MatchToBankTxn(ctx, tenantID, invoiceID, bankTxnID, amount.String())
}

// GetLineItems retrieves line items for an invoice.
func (s *InvoiceService) GetLineItems(ctx context.Context, tenantID, invoiceID uuid.UUID) ([]model.InvoiceLineItem, error) {
	return s.repo.GetLineItems(ctx, tenantID, invoiceID)
}

// ListInvoicesForMatching returns invoices eligible for bank matching.
func (s *InvoiceService) ListInvoicesForMatching(ctx context.Context, tenantID uuid.UUID, customerID *uuid.UUID) ([]model.SalesInvoice, error) {
	return s.repo.ListInvoicesForMatching(ctx, tenantID, customerID)
}