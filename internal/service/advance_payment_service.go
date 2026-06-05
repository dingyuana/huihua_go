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

type AdvancePaymentService struct {
	repo       *repository.AdvancePaymentRepository
	partyRepo  *repository.PartyRepository
	voucherSvc *VoucherAutoGenerateService
}

func NewAdvancePaymentService(
	repo *repository.AdvancePaymentRepository,
	partyRepo *repository.PartyRepository,
	voucherSvc *VoucherAutoGenerateService,
) *AdvancePaymentService {
	return &AdvancePaymentService{repo: repo, partyRepo: partyRepo, voucherSvc: voucherSvc}
}

type CreateAdvancePaymentRequest struct {
	CompanyID     uuid.UUID       `json:"company_id"`
	SupplierID    uuid.UUID       `json:"supplier_id"`
	Amount        decimal.Decimal `json:"amount"`
	PaidDate      time.Time       `json:"paid_date"`
	DueDate       *time.Time      `json:"due_date,omitempty"`
	BankAccountID *uuid.UUID      `json:"bank_account_id,omitempty"`
	ReferenceNo   *string         `json:"reference_no,omitempty"`
	Remark        *string         `json:"remark,omitempty"`
}

func (s *AdvancePaymentService) Create(ctx context.Context, tenantID, userID uuid.UUID, req *CreateAdvancePaymentRequest) (*model.AdvancePayment, error) {
	if req.Amount.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("amount must be positive")
	}
	if req.SupplierID == uuid.Nil {
		return nil, errors.New("supplier_id required")
	}

	no, err := s.repo.GenerateAdvanceNo(ctx, tenantID, "ADV-P")
	if err != nil {
		return nil, err
	}

	a := &model.AdvancePayment{
		ID:               uuid.New(),
		TenantID:         tenantID,
		CompanyID:        req.CompanyID,
		SupplierID:       req.SupplierID,
		AdvanceNo:        no,
		Amount:           req.Amount,
		AllocatedAmount:  decimal.Zero,
		OutstandingAmount: req.Amount,
		PaidDate:         req.PaidDate,
		DueDate:          req.DueDate,
		Status:           string(model.AdvancePaymentStatusDraft),
		SourceType:       "supplier_prepayment",
		BankAccountID:    req.BankAccountID,
		ReferenceNo:      req.ReferenceNo,
		Remark:           req.Remark,
		CreatedBy:        &userID,
		CreatedAt:        time.Now(),
	}
	if err := s.repo.Create(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *AdvancePaymentService) Confirm(ctx context.Context, tenantID, userID, advanceID uuid.UUID) (*model.AdvancePayment, error) {
	a, err := s.repo.GetByID(ctx, tenantID, advanceID)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, errors.New("advance payment not found")
	}
	if a.Status != string(model.AdvancePaymentStatusDraft) {
		return nil, errors.New("only draft can be confirmed")
	}

	if s.voucherSvc != nil {
		voucherNo, err := s.voucherSvc.GenerateFromAdvancePayment(ctx, tenantID, advanceID, userID)
		if err != nil {
			return nil, err
		}
		if err := s.repo.SetVoucher(ctx, tenantID, advanceID, uuid.Nil, voucherNo); err != nil {
			return nil, err
		}
	}

	if err := s.repo.UpdateStatus(ctx, tenantID, advanceID, string(model.AdvancePaymentStatusConfirmed)); err != nil {
		return nil, err
	}
	if err := s.repo.MarkConfirmed(ctx, tenantID, advanceID, userID); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, tenantID, advanceID)
}

func (s *AdvancePaymentService) List(ctx context.Context, tenantID uuid.UUID, status *string) ([]*model.AdvancePayment, error) {
	return s.repo.ListByTenant(ctx, tenantID, status)
}

func (s *AdvancePaymentService) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*model.AdvancePayment, error) {
	return s.repo.GetByID(ctx, tenantID, id)
}

func (s *AdvancePaymentService) ListOutstanding(ctx context.Context, tenantID, supplierID uuid.UUID) ([]*model.AdvancePayment, error) {
	return s.repo.ListOutstanding(ctx, tenantID, supplierID)
}
