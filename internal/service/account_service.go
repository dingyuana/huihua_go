package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
// Skips accounts that already exist (same tenant_id + code).
func (s *AccountService) InitFromSeed(ctx context.Context, tenantID, companyID uuid.UUID) error {
	rows, err := s.seedPool.Query(ctx, `
		SELECT code, name, account_type, root_type, parent_code, is_group, lft, rgt
		FROM standard_accounts_seed ORDER BY lft`)
	if err != nil {
		return fmt.Errorf("fetch seed accounts: %w", err)
	}
	defer rows.Close()

	parentCodeMap := make(map[string]uuid.UUID)

	for rows.Next() {
		var code, name, accountType, rootType string
		var parentCode *string
		var isGroup *bool
		var lft, rgt *int
		if err := rows.Scan(&code, &name, &accountType, &rootType, &parentCode, &isGroup, &lft, &rgt); err != nil {
			return fmt.Errorf("scan seed account %s: %w", code, err)
		}

		ig := false
		if isGroup != nil {
			ig = *isGroup
		}

		acc := &model.Account{
			ID:          uuid.New(),
			Code:        code,
			Name:        name,
			AccountType: utils.StrPtr(accountType),
			RootType:    utils.StrPtr(rootType),
			IsGroup:     ig,
			CompanyID:   companyID,
			TenantID:    tenantID,
			Currency:    "CNY",
			IsActive:    true,
		}
		if parentCode != nil && *parentCode != "" {
			if pid, ok := parentCodeMap[*parentCode]; ok {
				acc.ParentID = &pid
			}
		}
		parentCodeMap[code] = acc.ID

		_, insertErr := s.seedPool.Exec(ctx, `
			INSERT INTO accounts (id, code, name, account_type, root_type, parent_id, is_group, company_id, tenant_id, currency, is_active, created_at, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,NOW(),NOW())
			ON CONFLICT (tenant_id, code) DO NOTHING`,
			acc.ID, acc.Code, acc.Name, acc.AccountType, acc.RootType, acc.ParentID,
			acc.IsGroup, acc.CompanyID, acc.TenantID, acc.Currency, acc.IsActive)
		if insertErr != nil {
			return fmt.Errorf("insert account %s: %w", code, insertErr)
		}
	}
	return nil
}

// InitFromSeedWithTx initializes accounts from the standard_accounts_seed table within a transaction.
func (s *AccountService) InitFromSeedWithTx(ctx context.Context, tx pgx.Tx, tenantID, companyID uuid.UUID) error {
	rows, err := tx.Query(ctx, `
		SELECT code, name, account_type, root_type, parent_code, is_group, lft, rgt
		FROM standard_accounts_seed ORDER BY lft`)
	if err != nil {
		return fmt.Errorf("fetch seed accounts: %w", err)
	}
	defer rows.Close()

	parentCodeMap := make(map[string]uuid.UUID)

	for rows.Next() {
		var code, name, accountType, rootType string
		var parentCode *string
		var isGroup *bool
		var lft, rgt *int
		if err := rows.Scan(&code, &name, &accountType, &rootType, &parentCode, &isGroup, &lft, &rgt); err != nil {
			return fmt.Errorf("scan seed account %s: %w", code, err)
		}

		ig := false
		if isGroup != nil {
			ig = *isGroup
		}

		acc := &model.Account{
			ID:          uuid.New(),
			Code:        code,
			Name:        name,
			AccountType: utils.StrPtr(accountType),
			RootType:    utils.StrPtr(rootType),
			IsGroup:     ig,
			CompanyID:   companyID,
			TenantID:    tenantID,
			Currency:    "CNY",
			IsActive:    true,
		}
		if parentCode != nil && *parentCode != "" {
			if pid, ok := parentCodeMap[*parentCode]; ok {
				acc.ParentID = &pid
			}
		}
		parentCodeMap[code] = acc.ID

		if _, err := s.repo.CreateWithTx(ctx, tx, tenantID, acc); err != nil {
			return fmt.Errorf("create account %s: %w", code, err)
		}
	}
	return nil
}

// GetTree returns accounts as a nested tree wrapped in a single synthetic root
// so the frontend can treat the root's children as level-1 options.
func (s *AccountService) GetTree(ctx context.Context, tenantID uuid.UUID) ([]model.Account, error) {
	flat, err := s.repo.GetTree(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	byID := make(map[uuid.UUID]*model.Account, len(flat))
	for i := range flat {
		byID[flat[i].ID] = &flat[i]
	}
	var roots []*model.Account
	for i := range flat {
		a := &flat[i]
		if a.ParentID == nil {
			roots = append(roots, a)
			continue
		}
		parent, ok := byID[*a.ParentID]
		if !ok {
			continue
		}
		parent.Children = append(parent.Children, a)
	}
	synthetic := model.Account{
		Name:     "全部科目",
		IsGroup:  true,
		Children: roots,
	}
	return []model.Account{synthetic}, nil
}
