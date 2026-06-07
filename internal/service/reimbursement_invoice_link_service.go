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

// ReimbursementInvoiceLinkService handles invoice-link business logic.
type ReimbursementInvoiceLinkService struct {
	linkRepo  *repository.ReimbursementInvoiceLinkRepository
	reimbRepo *repository.ReimbursementRepository
}

// NewReimbursementInvoiceLinkService creates a new ReimbursementInvoiceLinkService.
func NewReimbursementInvoiceLinkService(
	linkRepo *repository.ReimbursementInvoiceLinkRepository,
	reimbRepo *repository.ReimbursementRepository,
) *ReimbursementInvoiceLinkService {
	return &ReimbursementInvoiceLinkService{
		linkRepo:  linkRepo,
		reimbRepo: reimbRepo,
	}
}

// ListAvailableInvoices lists expense invoices that can be linked to a reimbursement.
// An invoice is available if it is verified and not already linked to this reimbursement.
func (s *ReimbursementInvoiceLinkService) ListAvailableInvoices(ctx context.Context, reimbID, tenantID uuid.UUID) ([]model.ReimbursementInvoiceLink, error) {
	// Verify reimbursement exists
	if _, err := s.reimbRepo.GetByID(ctx, tenantID, reimbID); err != nil {
		return nil, fmt.Errorf("reimbursement not found: %w", err)
	}

	// Return empty list for now - expense_invoices table will be created in TASK-2.7
	return []model.ReimbursementInvoiceLink{}, nil
}

// LinkInvoice links an expense invoice to a reimbursement.
// The linked_amount cannot exceed the invoice's available balance.
func (s *ReimbursementInvoiceLinkService) LinkInvoice(ctx context.Context, tenantID, reimbID, invoiceID, userID uuid.UUID, linkedAmount decimal.Decimal) (*model.ReimbursementInvoiceLink, error) {
	// Verify reimbursement exists and belongs to tenant
	reimb, err := s.reimbRepo.GetByID(ctx, tenantID, reimbID)
	if err != nil {
		return nil, fmt.Errorf("reimbursement not found: %w", err)
	}
	_ = reimb // unused in this check

	// Check if already linked
	existing, err := s.linkRepo.GetByReimbursementAndInvoice(ctx, reimbID, invoiceID)
	if err == nil && existing != nil {
		return nil, errors.New("invoice already linked to this reimbursement")
	}

	// TODO: When expense_invoices table exists, check available balance
	// For now, allow linking as long as amount > 0
	if linkedAmount.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("linked_amount must be positive")
	}

	link := &model.ReimbursementInvoiceLink{
		ReimbursementID: reimbID,
		InvoiceID:       invoiceID,
		InvoiceType:     "expense_invoice",
		LinkedAmount:    linkedAmount,
		LinkedBy:        &userID,
	}

	if err := s.linkRepo.Create(ctx, link); err != nil {
		return nil, fmt.Errorf("create link: %w", err)
	}

	return link, nil
}

// UnlinkInvoice removes the link between a reimbursement and an invoice.
func (s *ReimbursementInvoiceLinkService) UnlinkInvoice(ctx context.Context, reimbID, invoiceID uuid.UUID) error {
	return s.linkRepo.Delete(ctx, reimbID, invoiceID)
}

// GetLinkedInvoices returns all invoice links for a reimbursement.
func (s *ReimbursementInvoiceLinkService) GetLinkedInvoices(ctx context.Context, reimbID uuid.UUID) ([]model.ReimbursementInvoiceLink, error) {
	return s.linkRepo.ListByReimbursementID(ctx, reimbID)
}