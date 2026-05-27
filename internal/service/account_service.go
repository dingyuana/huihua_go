package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"huihua/finance/internal/model"
	"huihua/finance/internal/repository"
	"huihua/finance/pkg/utils"
)

// AccountService handles account operations with automatic code generation.
type AccountService struct {
	repo     *repository.AccountRepository
	seedPool *pgxpool.Pool
}

// NewAccountService creates a new AccountService.
func NewAccountService(repo *repository.AccountRepository, seedPool *pgxpool.Pool) *AccountService {
	return &AccountService{repo: repo, seedPool: seedPool}
}

// InitFromSeed initializes accounts from the standard_accounts_seed table.
// Creates a copy of all seed accounts for the given tenant/company.
func (s *AccountService) InitFromSeed(ctx context.Context, tenantID, companyID uuid.UUID) error {
	rows, err := s.seedPool.Query(ctx, `
		SELECT code, name, account_type, root_type, parent_code, is_group, lft, rgt
		FROM standard_accounts_seed ORDER BY lft`)
	if err != nil {
		return fmt.Errorf("fetch seed accounts: %w", err)
	}
	defer rows.Close()

	// Build parent_code -> new_id mapping
	parentCodeMap := make(map[string]uuid.UUID)

	for rows.Next() {
		var code, name, accountType, rootType, parentCode string
		var isGroup bool
		var lft, rgt int
		if err := rows.Scan(&code, &name, &accountType, &rootType, &parentCode, &isGroup, &lft, &rgt); err != nil {
			return err
		}
		acc := &model.Account{
			ID:          uuid.New(),
			Code:        code,
			Name:        name,
			AccountType: utils.StrPtr(accountType),
			RootType:    utils.StrPtr(rootType),
			IsGroup:     isGroup,
			CompanyID:   companyID,
			TenantID:    tenantID,
			Currency:    "CNY",
			IsActive:    true,
		}
		if parentCode != "" {
			if pid, ok := parentCodeMap[parentCode]; ok {
				acc.ParentID = &pid
			}
		}
		parentCodeMap[code] = acc.ID

		if _, err := s.repo.Create(ctx, tenantID, acc); err != nil {
			return fmt.Errorf("create account %s: %w", code, err)
		}
	}
	return nil
}

