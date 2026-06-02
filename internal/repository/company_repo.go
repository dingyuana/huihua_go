package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"huihua/finance/internal/model"
)

// CompanyRepository handles company_settings table operations.
type CompanyRepository struct {
	pool *pgxpool.Pool
}

// NewCompanyRepository creates a new CompanyRepository.
func NewCompanyRepository(pool *pgxpool.Pool) *CompanyRepository {
	return &CompanyRepository{pool: pool}
}

// Create inserts a new company settings record.
func (r *CompanyRepository) Create(ctx context.Context, tenantID uuid.UUID, cs *model.CompanySettings) (*model.CompanySettings, error) {
	cs.ID = uuid.New()
	cs.TenantID = tenantID
	cs.CreatedAt = time.Now()
	cs.UpdatedAt = time.Now()

	_, err := r.pool.Exec(ctx, `
		INSERT INTO company_settings (id, tenant_id, company_name, fiscal_year_start_month, enable_date,
			default_currency, chart_of_accounts_template, is_initialized, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		cs.ID, cs.TenantID, cs.CompanyName, cs.FiscalYearStartMonth, cs.EnableDate,
		cs.DefaultCurrency, cs.ChartOfAccountsTemplate, cs.IsInitialized, cs.CreatedAt, cs.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return cs, nil
}

// CreateWithTx inserts or updates a company settings record within a transaction.
func (r *CompanyRepository) CreateWithTx(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID, cs *model.CompanySettings) (*model.CompanySettings, error) {
	if cs.ID == uuid.Nil {
		cs.ID = uuid.New()
	}
	cs.TenantID = tenantID
	cs.CreatedAt = time.Now()
	cs.UpdatedAt = time.Now()

	// Use a subquery trick: only DO UPDATE when the existing row is NOT initialized.
	_, err := tx.Exec(ctx, `
		INSERT INTO company_settings (id, tenant_id, company_name, fiscal_year_start_month, enable_date,
			default_currency, chart_of_accounts_template, is_initialized, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (tenant_id) DO UPDATE SET
			company_name = EXCLUDED.company_name,
			fiscal_year_start_month = EXCLUDED.fiscal_year_start_month,
			enable_date = EXCLUDED.enable_date,
			default_currency = EXCLUDED.default_currency,
			chart_of_accounts_template = EXCLUDED.chart_of_accounts_template,
			updated_at = EXCLUDED.updated_at
		WHERE company_settings.id != EXCLUDED.id`,
		cs.ID, cs.TenantID, cs.CompanyName, cs.FiscalYearStartMonth, cs.EnableDate,
		cs.DefaultCurrency, cs.ChartOfAccountsTemplate, cs.IsInitialized, cs.CreatedAt, cs.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return cs, nil
}

// GetByTenant retrieves company settings for a tenant.
func (r *CompanyRepository) GetByTenant(ctx context.Context, tenantID uuid.UUID) (*model.CompanySettings, error) {
	var cs model.CompanySettings
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, company_name, fiscal_year_start_month, enable_date,
			default_currency, chart_of_accounts_template, is_initialized, created_at, updated_at
		FROM company_settings WHERE tenant_id = $1`, tenantID).
		Scan(&cs.ID, &cs.TenantID, &cs.CompanyName, &cs.FiscalYearStartMonth, &cs.EnableDate,
			&cs.DefaultCurrency, &cs.ChartOfAccountsTemplate, &cs.IsInitialized, &cs.CreatedAt, &cs.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &cs, nil
}

// Update updates company settings.
func (r *CompanyRepository) Update(ctx context.Context, tenantID uuid.UUID, cs *model.CompanySettings) error {
	cs.UpdatedAt = time.Now()
	_, err := r.pool.Exec(ctx, `
		UPDATE company_settings SET company_name = $3, fiscal_year_start_month = $4,
			enable_date = $5, default_currency = $6, chart_of_accounts_template = $7,
			is_initialized = $8, updated_at = $9
		WHERE tenant_id = $1 AND id = $2`,
		tenantID, cs.ID, cs.CompanyName, cs.FiscalYearStartMonth, cs.EnableDate,
		cs.DefaultCurrency, cs.ChartOfAccountsTemplate, cs.IsInitialized, cs.UpdatedAt)
	return err
}

// SetInitialized marks the company as initialized.
func (r *CompanyRepository) SetInitialized(ctx context.Context, tenantID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE company_settings SET is_initialized = TRUE, updated_at = $2
		WHERE tenant_id = $1`, tenantID, time.Now())
	return err
}

// SetInitializedWithTx marks the company as initialized within a transaction.
func (r *CompanyRepository) SetInitializedWithTx(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID) error {
	_, err := tx.Exec(ctx, `
		UPDATE company_settings SET is_initialized = TRUE, updated_at = $2
		WHERE tenant_id = $1`, tenantID, time.Now())
	return err
}
