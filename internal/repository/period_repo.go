package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"huihua/finance/internal/model"
)

var errNoRows = pgx.ErrNoRows

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

// BatchCreateWithTx inserts multiple periods within a transaction.
func (r *PeriodRepository) BatchCreateWithTx(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID, periods []model.AccountingPeriod) error {
	for i, p := range periods {
		p.ID = uuid.New()
		p.TenantID = tenantID
		p.CreatedAt = time.Now()
		if p.Status == "" {
			p.Status = "open"
		}
		if _, err := tx.Exec(ctx, `INSERT INTO accounting_periods (id, tenant_id, period_no, period_name, start_date, end_date, status, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT DO NOTHING`,
			p.ID, p.TenantID, p.PeriodNo, p.PeriodName, p.StartDate, p.EndDate, p.Status, p.CreatedAt); err != nil {
			return err
		}
		periods[i] = p
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

// GetByID retrieves a single period by its ID and tenant.
func (r *PeriodRepository) GetByID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*model.AccountingPeriod, error) {
	query := `
		SELECT id, tenant_id, period_no, period_name, start_date, end_date, status, closed_by, closed_at, created_at
		FROM accounting_periods
		WHERE tenant_id = $1 AND id = $2`
	p := &model.AccountingPeriod{}
	err := r.pool.QueryRow(ctx, query, tenantID, id).Scan(
		&p.ID, &p.TenantID, &p.PeriodNo, &p.PeriodName, &p.StartDate, &p.EndDate,
		&p.Status, &p.ClosedBy, &p.ClosedAt, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// UpdateMeta updates period_name, start_date, and/or end_date for a period.
// Only non-nil fields are applied.
func (r *PeriodRepository) UpdateMeta(ctx context.Context, tenantID uuid.UUID, id uuid.UUID, periodName *string, startDate, endDate *time.Time) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE accounting_periods
		SET period_name = COALESCE($3, period_name),
		    start_date = COALESCE($4, start_date),
		    end_date   = COALESCE($5, end_date)
		WHERE tenant_id = $1 AND id = $2`,
		tenantID, id, periodName, startDate, endDate)
	return err
}

// Delete removes an accounting period by ID.
func (r *PeriodRepository) Delete(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM accounting_periods WHERE tenant_id = $1 AND id = $2`, tenantID, id)
	return err
}

// CheckOverlap returns true if another period for the same tenant overlaps the given date range.
// excludeID, when non-nil, excludes a specific period from the check (for updates).
func (r *PeriodRepository) CheckOverlap(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time, excludeID *uuid.UUID) (bool, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM accounting_periods
		WHERE tenant_id = $1
		  AND start_date <= $3
		  AND end_date   >= $2`
	args := []interface{}{tenantID, startDate, endDate}
	if excludeID != nil {
		query += ` AND id <> $4`
		args = append(args, *excludeID)
	}
	err := r.pool.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetCurrentOpen returns the currently open period for a tenant.
func (r *PeriodRepository) GetCurrentOpen(ctx context.Context, tenantID uuid.UUID) (*model.AccountingPeriod, error) {
	query := `
		SELECT id, tenant_id, period_no, period_name, start_date, end_date, status, closed_by, closed_at, created_at
		FROM accounting_periods
		WHERE tenant_id = $1 AND status = 'open'
		ORDER BY start_date DESC LIMIT 1`
	p := &model.AccountingPeriod{}
	err := r.pool.QueryRow(ctx, query, tenantID).Scan(
		&p.ID, &p.TenantID, &p.PeriodNo, &p.PeriodName, &p.StartDate, &p.EndDate,
		&p.Status, &p.ClosedBy, &p.ClosedAt, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}
