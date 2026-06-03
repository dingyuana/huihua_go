package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"huihua/finance/internal/model"
)

// AccountRepository provides data access for the accounts table.
type AccountRepository struct {
	pool *pgxpool.Pool
}

// NewAccountRepository creates a new AccountRepository.
func NewAccountRepository(pool *pgxpool.Pool) *AccountRepository {
	return &AccountRepository{pool: pool}
}

// Create inserts a new account and returns it.
func (r *AccountRepository) Create(ctx context.Context, tenantID uuid.UUID, a *model.Account) (*model.Account, error) {
	query := `
		INSERT INTO accounts (id, code, name, account_type, root_type, parent_id, lft, rgt, is_group, company_id, tenant_id, currency, is_active, opening_balance)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING created_at, updated_at`

	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}

	err := r.pool.QueryRow(ctx, query,
		a.ID, a.Code, a.Name, a.AccountType, a.RootType, a.ParentID,
		a.Lft, a.Rgt, a.IsGroup, a.CompanyID, tenantID, a.Currency,
		a.IsActive, a.OpeningBalance,
	).Scan(&a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create account: %w", err)
	}

	a.TenantID = tenantID
	return a, nil
}

// CreateWithTx inserts a new account within a transaction.
func (r *AccountRepository) CreateWithTx(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID, a *model.Account) (*model.Account, error) {
	query := `
		INSERT INTO accounts (id, code, name, account_type, root_type, parent_id, lft, rgt, is_group, company_id, tenant_id, currency, is_active, opening_balance)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING created_at, updated_at`

	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}

	err := tx.QueryRow(ctx, query,
		a.ID, a.Code, a.Name, a.AccountType, a.RootType, a.ParentID,
		a.Lft, a.Rgt, a.IsGroup, a.CompanyID, tenantID, a.Currency,
		a.IsActive, a.OpeningBalance,
	).Scan(&a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create account: %w", err)
	}

	a.TenantID = tenantID
	return a, nil
}

// GetByID retrieves an account by its ID within the given tenant.
func (r *AccountRepository) GetByID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*model.Account, error) {
	query := `
		SELECT id, code, name, account_type, root_type, parent_id, lft, rgt, is_group,
		       company_id, tenant_id, currency, is_active, opening_balance, created_at, updated_at
		FROM accounts
		WHERE id = $1 AND tenant_id = $2`

	a := &model.Account{}
	err := r.pool.QueryRow(ctx, query, id, tenantID).Scan(
		&a.ID, &a.Code, &a.Name, &a.AccountType, &a.RootType, &a.ParentID,
		&a.Lft, &a.Rgt, &a.IsGroup, &a.CompanyID, &a.TenantID,
		&a.Currency, &a.IsActive, &a.OpeningBalance, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get account by id: %w", err)
	}
	return a, nil
}

// GetByCode retrieves an account by its code within the given tenant.
func (r *AccountRepository) GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (*model.Account, error) {
	query := `
		SELECT id, code, name, account_type, root_type, parent_id, lft, rgt, is_group,
		       company_id, tenant_id, currency, is_active, opening_balance, created_at, updated_at
		FROM accounts
		WHERE code = $1 AND tenant_id = $2`

	a := &model.Account{}
	err := r.pool.QueryRow(ctx, query, code, tenantID).Scan(
		&a.ID, &a.Code, &a.Name, &a.AccountType, &a.RootType, &a.ParentID,
		&a.Lft, &a.Rgt, &a.IsGroup, &a.CompanyID, &a.TenantID,
		&a.Currency, &a.IsActive, &a.OpeningBalance, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get account by code: %w", err)
	}
	return a, nil
}

// ListByTenant retrieves all accounts for the given tenant, ordered by lft.
func (r *AccountRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]model.Account, error) {
	query := `
		SELECT id, code, name, account_type, root_type, parent_id, lft, rgt, is_group,
		       company_id, tenant_id, currency, is_active, opening_balance, created_at, updated_at
		FROM accounts
		WHERE tenant_id = $1
		ORDER BY lft`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list accounts by tenant: %w", err)
	}
	defer rows.Close()

	var accounts []model.Account
	for rows.Next() {
		var a model.Account
		if err := rows.Scan(
			&a.ID, &a.Code, &a.Name, &a.AccountType, &a.RootType, &a.ParentID,
			&a.Lft, &a.Rgt, &a.IsGroup, &a.CompanyID, &a.TenantID,
			&a.Currency, &a.IsActive, &a.OpeningBalance, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan account: %w", err)
		}
		accounts = append(accounts, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate accounts: %w", err)
	}
	return accounts, nil
}

// GetTree retrieves the account tree for the given tenant using nested set ordering.
// Returns all accounts ordered by lft to allow the caller to reconstruct the tree.
func (r *AccountRepository) GetTree(ctx context.Context, tenantID uuid.UUID) ([]model.Account, error) {
	query := `
		SELECT id, code, name, account_type, root_type, parent_id, lft, rgt, is_group,
		       company_id, tenant_id, currency, is_active, opening_balance, created_at, updated_at
		FROM accounts
		WHERE tenant_id = $1
		ORDER BY lft ASC`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get account tree: %w", err)
	}
	defer rows.Close()

	var accounts []model.Account
	for rows.Next() {
		var a model.Account
		if err := rows.Scan(
			&a.ID, &a.Code, &a.Name, &a.AccountType, &a.RootType, &a.ParentID,
			&a.Lft, &a.Rgt, &a.IsGroup, &a.CompanyID, &a.TenantID,
			&a.Currency, &a.IsActive, &a.OpeningBalance, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan account tree node: %w", err)
		}
		accounts = append(accounts, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate account tree: %w", err)
	}
	return accounts, nil
}

// ListByType retrieves accounts by account_type.
func (r *AccountRepository) ListByType(ctx context.Context, tenantID uuid.UUID, accountType string) ([]model.Account, error) {
	query := `
		SELECT id, code, name, account_type, root_type, parent_id, lft, rgt, is_group,
		       company_id, tenant_id, currency, is_active, opening_balance, created_at
		FROM accounts WHERE tenant_id = $1 AND account_type = $2
		ORDER BY code`

	rows, err := r.pool.Query(ctx, query, tenantID, accountType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []model.Account
	for rows.Next() {
		var a model.Account
		if err := rows.Scan(
			&a.ID, &a.Code, &a.Name, &a.AccountType, &a.RootType, &a.ParentID,
			&a.Lft, &a.Rgt, &a.IsGroup, &a.CompanyID, &a.TenantID,
			&a.Currency, &a.IsActive, &a.OpeningBalance, &a.CreatedAt,
		); err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, rows.Err()
}
