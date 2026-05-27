package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
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
			default_currency, chart_template, is_initialized, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		cs.ID, cs.TenantID, cs.CompanyName, cs.FiscalYearStartMonth, cs.EnableDate,
		cs.DefaultCurrency, cs.ChartTemplate, cs.IsInitialized, cs.CreatedAt, cs.UpdatedAt)
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
			default_currency, chart_template, is_initialized, created_at, updated_at
		FROM company_settings WHERE tenant_id = $1`, tenantID).
		Scan(&cs.ID, &cs.TenantID, &cs.CompanyName, &cs.FiscalYearStartMonth, &cs.EnableDate,
			&cs.DefaultCurrency, &cs.ChartTemplate, &cs.IsInitialized, &cs.CreatedAt, &cs.UpdatedAt)
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
			enable_date = $5, default_currency = $6, chart_template = $7,
			is_initialized = $8, updated_at = $9
		WHERE tenant_id = $1 AND id = $2`,
		tenantID, cs.ID, cs.CompanyName, cs.FiscalYearStartMonth, cs.EnableDate,
		cs.DefaultCurrency, cs.ChartTemplate, cs.IsInitialized, cs.UpdatedAt)
	return err
}

// SetInitialized marks the company as initialized.
func (r *CompanyRepository) SetInitialized(ctx context.Context, tenantID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE company_settings SET is_initialized = TRUE, updated_at = $2
		WHERE tenant_id = $1`, tenantID, time.Now())
	return err
}