package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/model"
)

// ExchangeRateRepository handles exchange_rates table operations.
type ExchangeRateRepository struct {
	pool *pgxpool.Pool
}

// NewExchangeRateRepository creates a new ExchangeRateRepository.
func NewExchangeRateRepository(pool *pgxpool.Pool) *ExchangeRateRepository {
	return &ExchangeRateRepository{pool: pool}
}

// Create inserts a new exchange rate.
func (r *ExchangeRateRepository) Create(ctx context.Context, tenantID uuid.UUID, rate *model.ExchangeRate) error {
	query := `
		INSERT INTO exchange_rates (id, tenant_id, currency_code, target_currency, rate, effective_date, rate_type, source, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.pool.Exec(ctx, query,
		rate.ID, tenantID, rate.CurrencyCode, rate.TargetCurrency,
		rate.Rate, rate.EffectiveDate, rate.RateType, rate.Source, time.Now())
	return err
}

// GetRate retrieves the rate for a specific date (exact match).
func (r *ExchangeRateRepository) GetRate(ctx context.Context, tenantID uuid.UUID, currency, target string, date time.Time) (*model.ExchangeRate, error) {
	query := `
		SELECT id, tenant_id, currency_code, target_currency, rate, effective_date, rate_type, source, created_at
		FROM exchange_rates
		WHERE tenant_id = $1 AND currency_code = $2 AND target_currency = $3 AND effective_date = $4`
	var rate model.ExchangeRate
	err := r.pool.QueryRow(ctx, query, tenantID, currency, target, date).Scan(
		&rate.ID, &rate.TenantID, &rate.CurrencyCode, &rate.TargetCurrency,
		&rate.Rate, &rate.EffectiveDate, &rate.RateType, &rate.Source, &rate.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &rate, nil
}

// GetLatestRate retrieves the most recent rate before or on a date.
func (r *ExchangeRateRepository) GetLatestRate(ctx context.Context, tenantID uuid.UUID, currency, target string, date time.Time) (*model.ExchangeRate, error) {
	query := `
		SELECT id, tenant_id, currency_code, target_currency, rate, effective_date, rate_type, source, created_at
		FROM exchange_rates
		WHERE tenant_id = $1 AND currency_code = $2 AND target_currency = $3 AND effective_date <= $4
		ORDER BY effective_date DESC LIMIT 1`
	var rate model.ExchangeRate
	err := r.pool.QueryRow(ctx, query, tenantID, currency, target, date).Scan(
		&rate.ID, &rate.TenantID, &rate.CurrencyCode, &rate.TargetCurrency,
		&rate.Rate, &rate.EffectiveDate, &rate.RateType, &rate.Source, &rate.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &rate, nil
}

// ListByCurrency lists all rates for a given currency pair.
func (r *ExchangeRateRepository) ListByCurrency(ctx context.Context, tenantID uuid.UUID, currency, target string) ([]model.ExchangeRate, error) {
	query := `
		SELECT id, tenant_id, currency_code, target_currency, rate, effective_date, rate_type, source, created_at
		FROM exchange_rates
		WHERE tenant_id = $1 AND currency_code = $2 AND target_currency = $3
		ORDER BY effective_date DESC`
	rows, err := r.pool.Query(ctx, query, tenantID, currency, target)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var rates []model.ExchangeRate
	for rows.Next() {
		var rate model.ExchangeRate
		rows.Scan(&rate.ID, &rate.TenantID, &rate.CurrencyCode, &rate.TargetCurrency,
			&rate.Rate, &rate.EffectiveDate, &rate.RateType, &rate.Source, &rate.CreatedAt)
		rates = append(rates, rate)
	}
	return rates, nil
}

// UpdateRate updates an existing rate.
func (r *ExchangeRateRepository) UpdateRate(ctx context.Context, tenantID uuid.UUID, id uuid.UUID, rate decimal.Decimal) error {
	query := `UPDATE exchange_rates SET rate = $4 WHERE id = $1 AND tenant_id = $2`
	_, err := r.pool.Exec(ctx, query, id, tenantID, rate)
	return err
}

// DeleteRate deletes an exchange rate.
func (r *ExchangeRateRepository) DeleteRate(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error {
	query := `DELETE FROM exchange_rates WHERE id = $1 AND tenant_id = $2`
	_, err := r.pool.Exec(ctx, query, id, tenantID)
	return err
}

// ListAll lists all exchange rates for a tenant.
func (r *ExchangeRateRepository) ListAll(ctx context.Context, tenantID uuid.UUID) ([]model.ExchangeRate, error) {
	query := `
		SELECT id, tenant_id, currency_code, target_currency, rate, effective_date, rate_type, source, created_at
		FROM exchange_rates
		WHERE tenant_id = $1 ORDER BY effective_date DESC`
	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var rates []model.ExchangeRate
	for rows.Next() {
		var rate model.ExchangeRate
		rows.Scan(&rate.ID, &rate.TenantID, &rate.CurrencyCode, &rate.TargetCurrency,
			&rate.Rate, &rate.EffectiveDate, &rate.RateType, &rate.Source, &rate.CreatedAt)
		rates = append(rates, rate)
	}
	return rates, nil
}
