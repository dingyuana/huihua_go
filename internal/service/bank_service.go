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

// BankService handles bank account operations.
type BankService struct {
	repo        *repository.BankRepository
	accountRepo *repository.AccountRepository
}

// NewBankService creates a new BankService.
func NewBankService(repo *repository.BankRepository, accountRepo *repository.AccountRepository) *BankService {
	return &BankService{repo: repo, accountRepo: accountRepo}
}

// Create creates a new bank account after validating clearing_account_id.
func (s *BankService) Create(ctx context.Context, tenantID uuid.UUID, req *model.BankAccount) (*model.BankAccount, error) {
	// Validate clearing_account_id is an asset type account with debit balance direction
	clearingAcc, err := s.accountRepo.GetByID(ctx, tenantID, *req.ClearingAccountID)
	if err != nil {
		return nil, errors.New("clearing_account_id not found")
	}
	if clearingAcc.AccountType == nil || *clearingAcc.AccountType != "asset" {
		return nil, errors.New("clearing_account must be of type 'asset'")
	}
	if clearingAcc.RootType == nil || *clearingAcc.RootType != "debit" {
		return nil, errors.New("clearing_account must have root_type 'debit'")
	}
	return s.repo.Create(ctx, tenantID, req)
}

// List returns all bank accounts for a tenant.
func (s *BankService) List(ctx context.Context, tenantID uuid.UUID) ([]model.BankAccount, error) {
	return s.repo.ListByTenant(ctx, tenantID)
}

// GetByID returns a single bank account.
func (s *BankService) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*model.BankAccount, error) {
	return s.repo.GetByID(ctx, tenantID, id)
}

// Update updates bank account (only non-protected fields).
func (s *BankService) Update(ctx context.Context, tenantID, id uuid.UUID, req *model.BankAccount) error {
	return s.repo.Update(ctx, tenantID, id, req)
}

// Delete deletes a bank account.
func (s *BankService) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	return s.repo.Delete(ctx, tenantID, id)
}

func (s *BankService) AdjustBalance(ctx context.Context, tenantID, bankAccountID uuid.UUID, adjustmentType string, newBalance decimal.Decimal, reason string, operatorID uuid.UUID) (*model.BankBalanceAdjustment, error) {
	acct, err := s.repo.GetByID(ctx, tenantID, bankAccountID)
	if err != nil {
		return nil, errors.New("bank account not found")
	}

	now := time.Now()
	adj := &model.BankBalanceAdjustment{
		ID:             uuid.New(),
		TenantID:       tenantID,
		BankAccountID:  bankAccountID,
		AdjustmentType: adjustmentType,
		BeforeBalance:  acct.CurrentBalance,
		AfterBalance:   newBalance,
		Delta:          newBalance.Sub(acct.CurrentBalance),
		CreatedAt:      now,
	}
	if reason != "" {
		adj.Reason = &reason
	}
	if operatorID != uuid.Nil {
		adj.OperatorID = &operatorID
	}

	if err := s.repo.CreateAdjustment(ctx, adj); err != nil {
		return nil, err
	}
	return adj, nil
}

func (s *BankService) ListBalanceAdjustments(ctx context.Context, tenantID, bankAccountID uuid.UUID) ([]model.BankBalanceAdjustment, error) {
	return s.repo.ListAdjustmentsByAccount(ctx, tenantID, bankAccountID)
}