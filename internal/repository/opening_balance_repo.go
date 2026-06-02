package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/model"
)

// OpeningBalanceRepository provides data access for opening_balances table.
type OpeningBalanceRepository struct {
	pool *pgxpool.Pool
}

// NewOpeningBalanceRepository creates a new OpeningBalanceRepository.
func NewOpeningBalanceRepository(pool *pgxpool.Pool) *OpeningBalanceRepository {
	return &OpeningBalanceRepository{pool: pool}
}

// UpsertBatch inserts or updates multiple opening balance entries.
func (r *OpeningBalanceRepository) UpsertBatch(ctx context.Context, tenantID, companyID uuid.UUID, periodNo int, entries []model.OpeningBalanceEntry) error {
	if len(entries) == 0 {
		return nil
	}

	query := `
		INSERT INTO opening_balances (id, tenant_id, company_id, account_id, debit_amount, credit_amount, period_no, is_verified)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (tenant_id, company_id, account_id, period_no) DO UPDATE SET
			debit_amount = EXCLUDED.debit_amount,
			credit_amount = EXCLUDED.credit_amount,
			is_verified = EXCLUDED.is_verified
	`

	batch := &pgx.Batch{}
	for _, entry := range entries {
		if entry.ID == uuid.Nil {
			entry.ID = uuid.New()
		}
		batch.Queue(query, entry.ID, tenantID, companyID, entry.AccountID, entry.DebitAmount, entry.CreditAmount, periodNo, entry.IsVerified)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < len(entries); i++ {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("upsert batch entry %d: %w", i, err)
		}
	}

	return nil
}

// GetByPeriod retrieves all opening balances for a given tenant and period.
func (r *OpeningBalanceRepository) GetByPeriod(ctx context.Context, tenantID uuid.UUID, periodNo int) ([]model.OpeningBalanceEntry, error) {
	query := `
		SELECT id, tenant_id, company_id, account_id, debit_amount, credit_amount, period_no, is_verified, created_at
		FROM opening_balances
		WHERE tenant_id = $1 AND period_no = $2
		ORDER BY account_id
	`

	rows, err := r.pool.Query(ctx, query, tenantID, periodNo)
	if err != nil {
		return nil, fmt.Errorf("get opening balances by period: %w", err)
	}
	defer rows.Close()

	var entries []model.OpeningBalanceEntry
	for rows.Next() {
		var e model.OpeningBalanceEntry
		if err := rows.Scan(
			&e.ID, &e.TenantID, &e.CompanyID, &e.AccountID,
			&e.DebitAmount, &e.CreditAmount, &e.PeriodNo, &e.IsVerified, &e.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan opening balance: %w", err)
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate opening balances: %w", err)
	}

	return entries, nil
}

// GetByAccount retrieves the opening balance for a specific account and period.
func (r *OpeningBalanceRepository) GetByAccount(ctx context.Context, tenantID, accountID uuid.UUID, periodNo int) (*model.OpeningBalanceEntry, error) {
	query := `
		SELECT id, tenant_id, company_id, account_id, debit_amount, credit_amount, period_no, is_verified, created_at
		FROM opening_balances
		WHERE tenant_id = $1 AND account_id = $2 AND period_no = $3
	`

	var e model.OpeningBalanceEntry
	err := r.pool.QueryRow(ctx, query, tenantID, accountID, periodNo).Scan(
		&e.ID, &e.TenantID, &e.CompanyID, &e.AccountID,
		&e.DebitAmount, &e.CreditAmount, &e.PeriodNo, &e.IsVerified, &e.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get opening balance by account: %w", err)
	}

	return &e, nil
}

// GetTrialBalance returns all accounts with their opening balances for a period.
func (r *OpeningBalanceRepository) GetTrialBalance(ctx context.Context, tenantID uuid.UUID, periodNo int) ([]model.OpeningBalanceEntry, error) {
	// Get all accounts with their opening balances for the period
	query := `
		SELECT ob.id, ob.tenant_id, ob.company_id, ob.account_id, 
		       COALESCE(ob.debit_amount, 0) as debit_amount, 
		       COALESCE(ob.credit_amount, 0) as credit_amount, 
		       ob.period_no, ob.is_verified, ob.created_at
		FROM opening_balances ob
		WHERE ob.tenant_id = $1 AND ob.period_no = $2
		ORDER BY ob.account_id
	`

	rows, err := r.pool.Query(ctx, query, tenantID, periodNo)
	if err != nil {
		return nil, fmt.Errorf("get trial balance: %w", err)
	}
	defer rows.Close()

	var entries []model.OpeningBalanceEntry
	for rows.Next() {
		var e model.OpeningBalanceEntry
		if err := rows.Scan(
			&e.ID, &e.TenantID, &e.CompanyID, &e.AccountID,
			&e.DebitAmount, &e.CreditAmount, &e.PeriodNo, &e.IsVerified, &e.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan trial balance entry: %w", err)
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate trial balance: %w", err)
	}

	return entries, nil
}

// MarkVerified marks opening balances as verified for a period.
func (r *OpeningBalanceRepository) MarkVerified(ctx context.Context, tenantID uuid.UUID, periodNo int, accountIDs []uuid.UUID) error {
	if len(accountIDs) == 0 {
		return nil
	}

	query := `
		UPDATE opening_balances 
		SET is_verified = TRUE 
		WHERE tenant_id = $1 AND period_no = $2 AND account_id = ANY($3)
	`

	_, err := r.pool.Exec(ctx, query, tenantID, periodNo, accountIDs)
	if err != nil {
		return fmt.Errorf("mark verified: %w", err)
	}

	return nil
}

// GetSumByPeriod returns total debit and credit for a period.
func (r *OpeningBalanceRepository) GetSumByPeriod(ctx context.Context, tenantID uuid.UUID, periodNo int) (decimal.Decimal, decimal.Decimal, error) {
	query := `
		SELECT COALESCE(SUM(debit_amount), 0), COALESCE(SUM(credit_amount), 0)
		FROM opening_balances
		WHERE tenant_id = $1 AND period_no = $2
	`

	var debit, credit decimal.Decimal
	err := r.pool.QueryRow(ctx, query, tenantID, periodNo).Scan(&debit, &credit)
	if err != nil {
		return decimal.Zero, decimal.Zero, fmt.Errorf("get sum by period: %w", err)
	}

	return debit, credit, nil
}
