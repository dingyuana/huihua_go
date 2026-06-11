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

// Update updates an account's editable fields (code, name, account_type, root_type, parent_id, is_group, company_id, currency, is_active, opening_balance).
// Does NOT modify tree-structure fields (lft, rgt, level, path, tenant_id).
func (r *AccountRepository) Update(ctx context.Context, a *model.Account) error {
	query := `
		UPDATE accounts
		SET code=$1, name=$2, account_type=$3, root_type=$4, parent_id=$5,
		    is_group=$6, company_id=$7, currency=$8, is_active=$9,
		    opening_balance=$10, updated_at=NOW()
		WHERE id=$11`
	_, err := r.pool.Exec(ctx, query,
		a.Code, a.Name, a.AccountType, a.RootType, a.ParentID,
		a.IsGroup, a.CompanyID, a.Currency, a.IsActive,
		a.OpeningBalance, a.ID,
	)
	if err != nil {
		return fmt.Errorf("update account: %w", err)
	}
	return nil
}

// Delete removes an account by ID (tenant-scoped).
func (r *AccountRepository) Delete(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM accounts WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	if err != nil {
		return fmt.Errorf("delete account: %w", err)
	}
	return nil
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

// ListPaginated returns a paginated slice of accounts for a tenant, optionally
// filtered by exact code match. Returns the page slice and the total count
// (before pagination) for client-side paging.
func (r *AccountRepository) ListPaginated(ctx context.Context, tenantID uuid.UUID, limit, offset int, code string) ([]model.Account, int, error) {
	var total int
	countQuery := `SELECT COUNT(*) FROM accounts WHERE tenant_id = $1`
	countArgs := []interface{}{tenantID}
	if code != "" {
		countQuery += ` AND code = $2`
		countArgs = append(countArgs, code)
	}
	if err := r.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count accounts: %w", err)
	}

	query := `
		SELECT id, code, name, account_type, root_type, parent_id, lft, rgt, is_group,
		       company_id, tenant_id, currency, is_active, opening_balance, created_at, updated_at
		FROM accounts
		WHERE tenant_id = $1`
	args := []interface{}{tenantID}
	if code != "" {
		query += ` AND code = $2`
		args = append(args, code)
	}
	query += ` ORDER BY lft LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list accounts paginated: %w", err)
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
			return nil, 0, fmt.Errorf("scan account: %w", err)
		}
		accounts = append(accounts, a)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate accounts: %w", err)
	}
	return accounts, total, nil
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

// HasChildren checks whether an account has child nodes (direct children).
func (r *AccountRepository) HasChildren(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (bool, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM accounts WHERE parent_id = $1 AND tenant_id = $2`, id, tenantID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("has children: %w", err)
	}
	return count > 0, nil
}

// GetMaxRgt returns the maximum rgt value across all accounts for the tenant.
// Returns 0 if no accounts exist.
func (r *AccountRepository) GetMaxRgt(ctx context.Context, tenantID uuid.UUID) (int, error) {
	var maxRgt int
	err := r.pool.QueryRow(ctx, `SELECT COALESCE(MAX(rgt), 0) FROM accounts WHERE tenant_id = $1`, tenantID).Scan(&maxRgt)
	if err != nil {
		return 0, fmt.Errorf("get max rgt: %w", err)
	}
	return maxRgt, nil
}

// GetMaxSiblingRgt returns the maximum rgt value among direct children of the given parent.
// Returns parent.lft if no children exist (so the new child starts at parent.lft + 1).
func (r *AccountRepository) GetMaxSiblingRgt(ctx context.Context, tenantID uuid.UUID, parentID uuid.UUID) (int, error) {
	var maxRgt int
	err := r.pool.QueryRow(ctx, `SELECT COALESCE(MAX(rgt), 0) FROM accounts WHERE parent_id = $1 AND tenant_id = $2`, parentID, tenantID).Scan(&maxRgt)
	if err != nil {
		return 0, fmt.Errorf("get max sibling rgt: %w", err)
	}
	return maxRgt, nil
}

// ListByParent retrieves direct children of a given parent account.
func (r *AccountRepository) ListByParent(ctx context.Context, tenantID uuid.UUID, parentID uuid.UUID) ([]model.Account, error) {
	query := `
		SELECT id, code, name, account_type, root_type, parent_id, lft, rgt, is_group,
		       company_id, tenant_id, currency, is_active, opening_balance, created_at, updated_at
		FROM accounts
		WHERE parent_id = $1 AND tenant_id = $2
		ORDER BY code`

	rows, err := r.pool.Query(ctx, query, parentID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list accounts by parent: %w", err)
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
	return accounts, rows.Err()
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
