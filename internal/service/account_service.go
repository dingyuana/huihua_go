package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
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

// seedAccountRow is the in-memory shape used by both InitFromSeed and
// InitFromSeedWithTx. Materializing rows into a slice closes the cursor
// before any INSERT — required because pgx refuses concurrent use of the
// same pool while rows are still open (returns "conn busy").
type seedAccountRow struct {
	code, name, accountType, rootType string
	parentCode                        *string
	isGroup                           *bool
	lft, rgt                          *int
}

// fetchSeedAccounts reads all rows from standard_accounts_seed into memory.
func (s *AccountService) fetchSeedAccounts(ctx context.Context) ([]seedAccountRow, error) {
	rows, err := s.seedPool.Query(ctx, `
		SELECT code, name, account_type, root_type, parent_code, is_group, lft, rgt
		FROM standard_accounts_seed ORDER BY lft`)
	if err != nil {
		return nil, fmt.Errorf("fetch seed accounts: %w", err)
	}
	defer rows.Close()

	var out []seedAccountRow
	for rows.Next() {
		var sd seedAccountRow
		if err := rows.Scan(&sd.code, &sd.name, &sd.accountType, &sd.rootType, &sd.parentCode, &sd.isGroup, &sd.lft, &sd.rgt); err != nil {
			return nil, fmt.Errorf("scan seed account: %w", err)
		}
		out = append(out, sd)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate seed accounts: %w", err)
	}
	return out, nil
}

// fetchSeedAccountsInTx is the transaction-bound variant. Use this when the
// caller is already inside a transaction (e.g. SetupService.CreateCompany) so
// the SELECT participates in the same RLS context as the subsequent INSERTs.
func (s *AccountService) fetchSeedAccountsInTx(ctx context.Context, tx pgx.Tx) ([]seedAccountRow, error) {
	rows, err := tx.Query(ctx, `
		SELECT code, name, account_type, root_type, parent_code, is_group, lft, rgt
		FROM standard_accounts_seed ORDER BY lft`)
	if err != nil {
		return nil, fmt.Errorf("fetch seed accounts: %w", err)
	}
	defer rows.Close()

	var out []seedAccountRow
	for rows.Next() {
		var sd seedAccountRow
		if err := rows.Scan(&sd.code, &sd.name, &sd.accountType, &sd.rootType, &sd.parentCode, &sd.isGroup, &sd.lft, &sd.rgt); err != nil {
			return nil, fmt.Errorf("scan seed account: %w", err)
		}
		out = append(out, sd)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate seed accounts: %w", err)
	}
	return out, nil
}

func (s *AccountService) buildAccountFromSeed(sd seedAccountRow, tenantID, companyID uuid.UUID, parentCodeMap map[string]uuid.UUID) *model.Account {
	ig := false
	if sd.isGroup != nil {
		ig = *sd.isGroup
	}
	lft, rgt := 0, 0
	if sd.lft != nil {
		lft = *sd.lft
	}
	if sd.rgt != nil {
		rgt = *sd.rgt
	}
	acc := &model.Account{
		ID:             uuid.New(),
		Code:           sd.code,
		Name:           sd.name,
		AccountType:    utils.StrPtr(sd.accountType),
		RootType:       utils.StrPtr(sd.rootType),
		IsGroup:        ig,
		Lft:            lft,
		Rgt:            rgt,
		CompanyID:      companyID,
		TenantID:       tenantID,
		Currency:       "CNY",
		IsActive:       true,
		OpeningBalance: decimal.Zero,
	}
	if sd.parentCode != nil && *sd.parentCode != "" {
		if pid, ok := parentCodeMap[*sd.parentCode]; ok {
			acc.ParentID = &pid
		}
	}
	return acc
}

// InitFromSeed initializes accounts from the standard_accounts_seed table.
// Creates a copy of all seed accounts for the given tenant/company.
// Skips accounts that already exist (same tenant_id + code).
func (s *AccountService) InitFromSeed(ctx context.Context, tenantID, companyID uuid.UUID) error {
	seeds, err := s.fetchSeedAccounts(ctx)
	if err != nil {
		return err
	}

	parentCodeMap := make(map[string]uuid.UUID)
	for _, sd := range seeds {
		acc := s.buildAccountFromSeed(sd, tenantID, companyID, parentCodeMap)
		parentCodeMap[sd.code] = acc.ID

		if _, err := s.seedPool.Exec(ctx, `
			INSERT INTO accounts (id, code, name, account_type, root_type, parent_id, lft, rgt, is_group, company_id, tenant_id, currency, is_active, opening_balance)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
			ON CONFLICT (tenant_id, code) DO NOTHING`,
			acc.ID, acc.Code, acc.Name, acc.AccountType, acc.RootType, acc.ParentID,
			acc.Lft, acc.Rgt, acc.IsGroup, acc.CompanyID, acc.TenantID, acc.Currency,
			acc.IsActive, acc.OpeningBalance,
		); err != nil {
			return fmt.Errorf("insert account %s: %w", sd.code, err)
		}
	}
	return nil
}

// InitFromSeedWithTx initializes accounts from the standard_accounts_seed table within a transaction.
func (s *AccountService) InitFromSeedWithTx(ctx context.Context, tx pgx.Tx, tenantID, companyID uuid.UUID) error {
	// Materialize seed rows into a slice first so the rows cursor is closed
	// before we issue INSERTs. Required because pgx refuses concurrent use of
	// the same pool while rows are still open ("conn busy").
	seeds, err := s.fetchSeedAccountsInTx(ctx, tx)
	if err != nil {
		return err
	}

	parentCodeMap := make(map[string]uuid.UUID)
	for _, sd := range seeds {
		acc := s.buildAccountFromSeed(sd, tenantID, companyID, parentCodeMap)
		parentCodeMap[sd.code] = acc.ID

		if _, err := s.repo.CreateWithTx(ctx, tx, tenantID, acc); err != nil {
			return fmt.Errorf("create account %s: %w", sd.code, err)
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

// List returns a paginated slice of accounts for the tenant.
// `code` empty lists all; non-empty filters by exact code match.
func (s *AccountService) List(ctx context.Context, tenantID uuid.UUID, limit, offset int, code string) ([]model.Account, int, error) {
	return s.repo.ListPaginated(ctx, tenantID, limit, offset, code)
}
