package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
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