package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"huihua/finance/internal/model"
)

// BankRepository handles bank_accounts table operations.
type BankRepository struct {
	pool *pgxpool.Pool
}

// NewBankRepository creates a new BankRepository.
func NewBankRepository(pool *pgxpool.Pool) *BankRepository {
	return &BankRepository{pool: pool}
}

// Create inserts a new bank account.
func (r *BankRepository) Create(ctx context.Context, tenantID uuid.UUID, ba *model.BankAccount) (*model.BankAccount, error) {
	ba.ID = uuid.New()
	ba.TenantID = tenantID
	ba.CreatedAt = time.Now()

	_, err := r.pool.Exec(ctx, `
		INSERT INTO bank_accounts (id, bank_name, account_number, clearing_account_id, company_id,
			tenant_id, currency, iban, swift_code, bank_account_type, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		ba.ID, ba.BankName, ba.AccountNumber, ba.ClearingAccountID, ba.CompanyID,
			ba.TenantID, ba.Currency, ba.IBAN, ba.SwiftCode, ba.BankAccountType, ba.IsActive, ba.CreatedAt)
	if err != nil {
		return nil, err
	}
	return ba, nil
}

// ListByTenant retrieves all bank accounts for a tenant.
func (r *BankRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]model.BankAccount, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, bank_name, account_number, clearing_account_id, company_id,
			tenant_id, currency, iban, swift_code, bank_account_type, is_active, created_at
		FROM bank_accounts WHERE tenant_id = $1 ORDER BY bank_name`,
		tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []model.BankAccount
	for rows.Next() {
		var ba model.BankAccount
		if err := rows.Scan(&ba.ID, &ba.BankName, &ba.AccountNumber, &ba.ClearingAccountID, &ba.CompanyID,
			&ba.TenantID, &ba.Currency, &ba.IBAN, &ba.SwiftCode, &ba.BankAccountType, &ba.IsActive, &ba.CreatedAt); err != nil {
			return nil, err
		}
		accounts = append(accounts, ba)
	}
	return accounts, rows.Err()
}

// GetByID retrieves a bank account by ID.
func (r *BankRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*model.BankAccount, error) {
	var ba model.BankAccount
	err := r.pool.QueryRow(ctx, `
		SELECT id, bank_name, account_number, clearing_account_id, company_id,
			tenant_id, currency, iban, swift_code, bank_account_type, is_active, created_at
		FROM bank_accounts WHERE tenant_id = $1 AND id = $2`,
		tenantID, id).
		Scan(&ba.ID, &ba.BankName, &ba.AccountNumber, &ba.ClearingAccountID, &ba.CompanyID,
			&ba.TenantID, &ba.Currency, &ba.IBAN, &ba.SwiftCode, &ba.BankAccountType, &ba.IsActive, &ba.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &ba, nil
}

// Update updates a bank account.
func (r *BankRepository) Update(ctx context.Context, tenantID, id uuid.UUID, ba *model.BankAccount) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE bank_accounts SET bank_name = $3, account_number = $4, swift_code = $5, is_active = $6
		WHERE tenant_id = $1 AND id = $2`,
		tenantID, id, ba.BankName, ba.AccountNumber, ba.SwiftCode, ba.IsActive)
	return err
}

// Delete soft-deletes a bank account.
func (r *BankRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE bank_accounts SET is_active = FALSE
		WHERE tenant_id = $1 AND id = $2`, tenantID, id)
	return err
}