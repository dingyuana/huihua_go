package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
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
	if ba.CurrentBalance.IsZero() {
		ba.CurrentBalance = ba.OpeningBalance
	}
	now := time.Now()
	ba.BalanceUpdatedAt = &now

	_, err := r.pool.Exec(ctx, `
		INSERT INTO bank_accounts (id, bank_name, account_number, clearing_account_id, company_id,
			tenant_id, currency, iban, swift_code, bank_account_type, is_active, is_cash, custodian, location,
			opening_balance, opening_date, current_balance, balance_updated_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)`,
		ba.ID, ba.BankName, ba.AccountNumber, ba.ClearingAccountID, ba.CompanyID,
			ba.TenantID, ba.Currency, ba.IBAN, ba.SwiftCode, ba.BankAccountType, ba.IsActive,
			ba.IsCash, ba.Custodian, ba.Location,
			ba.OpeningBalance, ba.OpeningDate, ba.CurrentBalance, ba.BalanceUpdatedAt, ba.CreatedAt)
	if err != nil {
		return nil, err
	}
	return ba, nil
}

// ListByTenant retrieves all bank accounts for a tenant.
func (r *BankRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]model.BankAccount, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, bank_name, account_number, clearing_account_id, company_id,
			tenant_id, currency, iban, swift_code, bank_account_type, is_active, is_cash, custodian, location,
			opening_balance, opening_date, current_balance, balance_updated_at, created_at
		FROM bank_accounts WHERE tenant_id = $1 ORDER BY is_cash DESC, bank_name`,
		tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []model.BankAccount
	for rows.Next() {
		var ba model.BankAccount
		if err := rows.Scan(&ba.ID, &ba.BankName, &ba.AccountNumber, &ba.ClearingAccountID, &ba.CompanyID,
			&ba.TenantID, &ba.Currency, &ba.IBAN, &ba.SwiftCode, &ba.BankAccountType, &ba.IsActive,
			&ba.IsCash, &ba.Custodian, &ba.Location,
			&ba.OpeningBalance, &ba.OpeningDate, &ba.CurrentBalance, &ba.BalanceUpdatedAt, &ba.CreatedAt); err != nil {
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
			tenant_id, currency, iban, swift_code, bank_account_type, is_active, is_cash, custodian, location,
			opening_balance, opening_date, current_balance, balance_updated_at, created_at
		FROM bank_accounts WHERE tenant_id = $1 AND id = $2`,
		tenantID, id).
		Scan(&ba.ID, &ba.BankName, &ba.AccountNumber, &ba.ClearingAccountID, &ba.CompanyID,
			&ba.TenantID, &ba.Currency, &ba.IBAN, &ba.SwiftCode, &ba.BankAccountType, &ba.IsActive,
			&ba.IsCash, &ba.Custodian, &ba.Location,
			&ba.OpeningBalance, &ba.OpeningDate, &ba.CurrentBalance, &ba.BalanceUpdatedAt, &ba.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &ba, nil
}

// Update updates a bank account.
func (r *BankRepository) Update(ctx context.Context, tenantID, id uuid.UUID, ba *model.BankAccount) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE bank_accounts SET
			bank_name = $3, account_number = $4, currency = $5,
			iban = $6, swift_code = $7, bank_account_type = $8,
			clearing_account_id = $9, is_active = $10,
			is_cash = $11, custodian = $12, location = $13,
			opening_balance = $14, opening_date = $15
		WHERE tenant_id = $1 AND id = $2`,
		tenantID, id,
		ba.BankName, ba.AccountNumber, ba.Currency,
		ba.IBAN, ba.SwiftCode, ba.BankAccountType,
		ba.ClearingAccountID, ba.IsActive,
		ba.IsCash, ba.Custodian, ba.Location,
		ba.OpeningBalance, ba.OpeningDate)
	return err
}

// Delete soft-deletes a bank account.
func (r *BankRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE bank_accounts SET is_active = FALSE
		WHERE tenant_id = $1 AND id = $2`, tenantID, id)
	return err
}

func (r *BankRepository) UpdateCurrentBalance(ctx context.Context, tenantID, id uuid.UUID, newBalance decimal.Decimal) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE bank_accounts SET current_balance = $3, balance_updated_at = NOW()
		WHERE tenant_id = $1 AND id = $2`,
		tenantID, id, newBalance)
	return err
}

func (r *BankRepository) CreateAdjustment(ctx context.Context, adj *model.BankBalanceAdjustment) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO bank_balance_adjustments (id, tenant_id, bank_account_id, adjustment_type, before_balance, after_balance, reason, operator_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())`,
		adj.ID, adj.TenantID, adj.BankAccountID, adj.AdjustmentType,
		adj.BeforeBalance, adj.AfterBalance, adj.Reason, adj.OperatorID)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `
		UPDATE bank_accounts SET current_balance = $2, balance_updated_at = NOW()
		WHERE tenant_id = $1 AND id = $3`,
		adj.TenantID, adj.AfterBalance, adj.BankAccountID)
	return err
}

func (r *BankRepository) ListAdjustmentsByAccount(ctx context.Context, tenantID, bankAccountID uuid.UUID) ([]model.BankBalanceAdjustment, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, bank_account_id, adjustment_type, before_balance, after_balance, delta, reason, operator_id, created_at
		FROM bank_balance_adjustments
		WHERE tenant_id = $1 AND bank_account_id = $2
		ORDER BY created_at DESC`,
		tenantID, bankAccountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.BankBalanceAdjustment
	for rows.Next() {
		var adj model.BankBalanceAdjustment
		if err := rows.Scan(&adj.ID, &adj.TenantID, &adj.BankAccountID, &adj.AdjustmentType,
			&adj.BeforeBalance, &adj.AfterBalance, &adj.Delta, &adj.Reason, &adj.OperatorID, &adj.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, adj)
	}
	return items, rows.Err()
}