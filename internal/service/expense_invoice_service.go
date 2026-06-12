package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/model"
	"huihua/finance/internal/repository"
)

type ExpenseInvoiceService struct {
	repo          *repository.ExpenseInvoiceRepository
	apInvoiceRepo *repository.ApInvoiceRepository
}

func NewExpenseInvoiceService(repo *repository.ExpenseInvoiceRepository, apInvoiceRepo *repository.ApInvoiceRepository) *ExpenseInvoiceService {
	return &ExpenseInvoiceService{
		repo:          repo,
		apInvoiceRepo: apInvoiceRepo,
	}
}

// Create creates a new expense invoice
func (s *ExpenseInvoiceService) Create(ctx context.Context, tenantID uuid.UUID, req *model.ExpenseInvoiceCreateRequest) (*model.ExpenseInvoice, error) {
	// Check for duplicate invoice number
	existing, err := s.repo.GetByInvoiceNo(ctx, tenantID, req.InvoiceNo)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("invoice number already exists")
	}

	now := time.Now()
	inv := &model.ExpenseInvoice{
		ID:              uuid.New(),
		TenantID:        tenantID,
		CompanyID:       req.CompanyID,
		InvoiceNo:       req.InvoiceNo,
		InvoiceCode:     req.InvoiceCode,
		InvoiceDate:     req.InvoiceDate,
		InvoiceKind:     req.InvoiceKind,
		TaxAmount:       req.TaxAmount,
		TotalAmount:     req.TotalAmount,
		VendorID:        req.VendorID,
		VendorName:      req.VendorName,
		TaxID:           req.TaxID,
		VerifyStatus:    model.ExpenseVerifyStatusUnverified,
		DeductionStatus: model.ExpenseDeductionStatusUndeducted,
		SourceFile:      req.SourceFile,
		OcrData:         req.OcrData,
		Status:          "pending",
		CreatedAt:       now,
		Remark:          req.Remark,
	}

	if err := s.repo.Create(ctx, tenantID, inv); err != nil {
		return nil, err
	}
	return inv, nil
}

// List returns expense invoices for a tenant
func (s *ExpenseInvoiceService) List(ctx context.Context, tenantID uuid.UUID, filters model.ExpenseInvoiceFilter) ([]*model.ExpenseInvoice, error) {
	return s.repo.List(ctx, tenantID, filters)
}

// GetByID retrieves an expense invoice by ID
func (s *ExpenseInvoiceService) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*model.ExpenseInvoice, error) {
	return s.repo.GetByID(ctx, tenantID, id)
}

// Update updates an expense invoice
func (s *ExpenseInvoiceService) Update(ctx context.Context, tenantID, id uuid.UUID, fields map[string]interface{}) error {
	existing, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("expense invoice not found")
	}
	return s.repo.UpdateFields(ctx, tenantID, id, fields)
}

// Delete removes an expense invoice
func (s *ExpenseInvoiceService) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	return s.repo.Delete(ctx, tenantID, id)
}

// VerifyInvoice performs verification on a single expense invoice
// This is a mock implementation - actual API call will be added later
func (s *ExpenseInvoiceService) VerifyInvoice(ctx context.Context, tenantID, id uuid.UUID) (*model.ExpenseInvoiceVerifyResponse, error) {
	inv, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if inv == nil {
		return nil, errors.New("expense invoice not found")
	}

	// Mock verification result
	now := time.Now()
	verifyResult := "verified"
	inv.VerifyStatus = model.ExpenseVerifyStatusVerified
	inv.VerifiedAt = &now
	inv.VerifyResult = &verifyResult

	// Update the record
	if err := s.repo.UpdateFields(ctx, tenantID, id, map[string]interface{}{
		"verify_status": model.ExpenseVerifyStatusVerified,
		"verified_at":   now,
		"verify_result": verifyResult,
	}); err != nil {
		return nil, err
	}

	return &model.ExpenseInvoiceVerifyResponse{
		InvoiceID:    id,
		VerifyStatus: model.ExpenseVerifyStatusVerified,
		VerifyResult: verifyResult,
		VerifiedAt:   now,
	}, nil
}

// BatchVerify performs verification on multiple expense invoices
func (s *ExpenseInvoiceService) BatchVerify(ctx context.Context, tenantID uuid.UUID, ids []uuid.UUID) ([]*model.ExpenseInvoiceVerifyResponse, error) {
	results := make([]*model.ExpenseInvoiceVerifyResponse, 0, len(ids))
	for _, id := range ids {
		result, err := s.VerifyInvoice(ctx, tenantID, id)
		if err != nil {
			// Continue with other invoices even if one fails
			continue
		}
		results = append(results, result)
	}
	return results, nil
}

// Confirm confirms an expense invoice and auto-creates the corresponding ApInvoice (应付单).
// This links the expense invoice to the payables ledger for payment processing.
func (s *ExpenseInvoiceService) Confirm(ctx context.Context, tenantID, id, userID uuid.UUID) (*model.ExpenseInvoice, error) {
	inv, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if inv == nil {
		return nil, errors.New("expense invoice not found")
	}
	if inv.Status == "confirmed" {
		return nil, errors.New("expense invoice already confirmed")
	}

	now := time.Now()

	// Create ApInvoice linked to this expense invoice
	var supplierID uuid.UUID
	if inv.VendorID != nil {
		supplierID = *inv.VendorID
	}
	ap := &model.ApInvoice{
		ID:                uuid.New(),
		TenantID:          tenantID,
		CompanyID:         inv.CompanyID,
		SupplierID:        supplierID,
		InvoiceID:         id,
		InvoiceNo:         inv.InvoiceNo,
		Amount:            inv.TotalAmount,
		PaidAmount:        decimal.Zero,
		OutstandingAmount: inv.TotalAmount,
		Status:            string(model.ApInvoiceStatusConfirmed),
		SourceType:        "expense_invoice",
		CreatedBy:         &userID,
		CreatedAt:         now,
		ConfirmedAt:       &now,
		ConfirmedBy:       &userID,
	}
	if err := s.apInvoiceRepo.Create(ctx, ap); err != nil {
		return nil, err
	}

	// Update expense invoice status
	if err := s.repo.UpdateFields(ctx, tenantID, id, map[string]interface{}{
		"status":     "confirmed",
		"updated_at": now,
	}); err != nil {
		return nil, err
	}

	inv.Status = "confirmed"
	inv.UpdatedAt = &now
	return inv, nil
}

// DeductInvoice marks an expense invoice as deducted
// Precondition: verify_status must be "verified"
func (s *ExpenseInvoiceService) DeductInvoice(ctx context.Context, tenantID, id uuid.UUID) (*model.ExpenseInvoice, error) {
	inv, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if inv == nil {
		return nil, errors.New("expense invoice not found")
	}

	if inv.VerifyStatus != model.ExpenseVerifyStatusVerified {
		return nil, errors.New("invoice must be verified before deduction")
	}

	if inv.DeductionStatus == model.ExpenseDeductionStatusDeducted {
		return nil, errors.New("invoice already deducted")
	}

	now := time.Now()
	if err := s.repo.UpdateFields(ctx, tenantID, id, map[string]interface{}{
		"deduction_status": model.ExpenseDeductionStatusDeducted,
		"deducted_at":      now,
	}); err != nil {
		return nil, err
	}

	inv.DeductionStatus = model.ExpenseDeductionStatusDeducted
	inv.DeductedAt = &now
	return inv, nil
}