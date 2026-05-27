package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"huihua/finance/internal/model"
)

// PeriodRepository handles accounting_periods table operations.
type PeriodRepository struct {
	pool *pgxpool.Pool
}

// NewPeriodRepository creates a new PeriodRepository.
func NewPeriodRepository(pool *pgxpool.Pool) *PeriodRepository {
	return &PeriodRepository{pool: pool}
}

// Create inserts an accounting period.
func (r *PeriodRepository) Create(ctx context.Context, tenantID uuid.UUID, p *model.AccountingPeriod) (*model.AccountingPeriod, error) {
	p.ID = uuid.New()
	p.TenantID = tenantID
	p.CreatedAt = time.Now()
	if p.Status == "" {
		p.Status = "open"
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO accounting_periods (id, tenant_id, period_no, period_name, start_date, end_date, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		p.ID, p.TenantID, p.PeriodNo, p.PeriodName, p.StartDate, p.EndDate, p.Status, p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// BatchCreate inserts multiple periods in one call.
func (r *PeriodRepository) BatchCreate(ctx context.Context, tenantID uuid.UUID, periods []model.AccountingPeriod) error {
	batch := &pgx.Batch{}
	for _, p := range periods {
		p.ID = uuid.New()
		p.TenantID = tenantID
		p.CreatedAt = time.Now()
		if p.Status == "" {
			p.Status = "open"
		}
		batch.Queue(`INSERT INTO accounting_periods (id, tenant_id, period_no, period_name, start_date, end_date, status, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT DO NOTHING`,
			p.ID, p.TenantID, p.PeriodNo, p.PeriodName, p.StartDate, p.EndDate, p.Status, p.CreatedAt)
	}
	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()
	for i := 0; i < len(periods); i++ {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}

// ListByTenant retrieves all periods for a tenant.
func (r *PeriodRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]model.AccountingPeriod, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, period_no, period_name, start_date, end_date, status, closed_by, closed_at, created_at
		FROM accounting_periods WHERE tenant_id = $1 ORDER BY period_no`,
		tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var periods []model.AccountingPeriod
	for rows.Next() {
		var p model.AccountingPeriod
		if err := rows.Scan(&p.ID, &p.TenantID, &p.PeriodNo, &p.PeriodName, &p.StartDate, &p.EndDate,
			&p.Status, &p.ClosedBy, &p.ClosedAt, &p.CreatedAt); err != nil {
			return nil, err
		}
		periods = append(periods, p)
	}
	return periods, rows.Err()
}

// UpdateStatus updates the status of a period.
func (r *PeriodRepository) UpdateStatus(ctx context.Context, tenantID uuid.UUID, periodNo int, status string, closedBy uuid.UUID) error {
	var closedAt *time.Time
	var cb *uuid.UUID
	if status == "closed" {
		now := time.Now()
		closedAt = &now
		cb = &closedBy
	}
	_, err := r.pool.Exec(ctx, `
		UPDATE accounting_periods SET status = $3, closed_by = $4, closed_at = $5
		WHERE tenant_id = $1 AND period_no = $2`,
		tenantID, periodNo, status, cb, closedAt)
	return err
}