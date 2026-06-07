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

type AdvanceReceiptService struct {
	repo        *repository.AdvanceReceiptRepository
	partyRepo   *repository.PartyRepository
	voucherSvc  *VoucherAutoGenerateService
}

func NewAdvanceReceiptService(
	repo *repository.AdvanceReceiptRepository,
	partyRepo *repository.PartyRepository,
	voucherSvc *VoucherAutoGenerateService,
) *AdvanceReceiptService {
	return &AdvanceReceiptService{repo: repo, partyRepo: partyRepo, voucherSvc: voucherSvc}
}

type CreateAdvanceReceiptRequest struct {
	CompanyID     uuid.UUID       `json:"company_id"`
	CustomerID    uuid.UUID       `json:"customer_id"`
	Amount        decimal.Decimal `json:"amount"`
	ReceivedDate  time.Time       `json:"received_date"`
	DueDate       *time.Time      `json:"due_date,omitempty"`
	BankAccountID *uuid.UUID      `json:"bank_account_id,omitempty"`
	ReferenceNo   *string         `json:"reference_no,omitempty"`
	Remark        *string         `json:"remark,omitempty"`
}

func (s *AdvanceReceiptService) Create(ctx context.Context, tenantID, userID uuid.UUID, req *CreateAdvanceReceiptRequest) (*model.AdvanceReceipt, error) {
	if req.Amount.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("amount must be positive")
	}
	if req.CustomerID == uuid.Nil {
		return nil, errors.New("customer_id required")
	}

	no, err := s.repo.GenerateAdvanceNo(ctx, tenantID, "ADV-R")
	if err != nil {
		return nil, err
	}

	a := &model.AdvanceReceipt{
		ID:               uuid.New(),
		TenantID:         tenantID,
		CompanyID:        req.CompanyID,
		CustomerID:       req.CustomerID,
		AdvanceNo:        no,
		Amount:           req.Amount,
		AllocatedAmount:  decimal.Zero,
		OutstandingAmount: req.Amount,
		ReceivedDate:     req.ReceivedDate,
		DueDate:          req.DueDate,
		Status:           string(model.AdvanceReceiptStatusDraft),
		SourceType:       "customer_prepayment",
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

func (s *AdvanceReceiptService) Confirm(ctx context.Context, tenantID, userID, advanceID uuid.UUID) (*model.AdvanceReceipt, error) {
	a, err := s.repo.GetByID(ctx, tenantID, advanceID)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, errors.New("advance receipt not found")
	}
	if a.Status != string(model.AdvanceReceiptStatusDraft) {
		return nil, errors.New("only draft can be confirmed")
	}

	if s.voucherSvc != nil {
		voucherNo, err := s.voucherSvc.GenerateFromAdvanceReceipt(ctx, tenantID, advanceID, userID)
		if err != nil {
			return nil, err
		}
		if err := s.repo.SetVoucher(ctx, tenantID, advanceID, uuid.Nil, voucherNo); err != nil {
			return nil, err
		}
	}

	if err := s.repo.UpdateStatus(ctx, tenantID, advanceID, string(model.AdvanceReceiptStatusConfirmed)); err != nil {
		return nil, err
	}
	if err := s.repo.MarkConfirmed(ctx, tenantID, advanceID, userID); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, tenantID, advanceID)
}

func (s *AdvanceReceiptService) List(ctx context.Context, tenantID uuid.UUID, status *string) ([]*model.AdvanceReceipt, error) {
	return s.repo.ListByTenant(ctx, tenantID, status)
}

func (s *AdvanceReceiptService) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*model.AdvanceReceipt, error) {
	return s.repo.GetByID(ctx, tenantID, id)
}

func (s *AdvanceReceiptService) ListOutstanding(ctx context.Context, tenantID, customerID uuid.UUID) ([]*model.AdvanceReceipt, error) {
	return s.repo.ListOutstanding(ctx, tenantID, customerID)
}

// GetCustomerSummary returns per-customer aggregated advance receipt balances.
// companyID == uuid.Nil means "all companies in tenant".
func (s *AdvanceReceiptService) GetCustomerSummary(ctx context.Context, tenantID, companyID uuid.UUID) ([]repository.CustomerAdvanceSummary, error) {
	return s.repo.GetCustomerSummary(ctx, tenantID, companyID)
}
