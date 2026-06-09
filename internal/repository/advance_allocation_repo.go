package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"huihua/finance/internal/model"
)

type AdvanceAllocationRepository struct {
	pool *pgxpool.Pool
}

func NewAdvanceAllocationRepository(pool *pgxpool.Pool) *AdvanceAllocationRepository {
	return &AdvanceAllocationRepository{pool: pool}
}

// BeginTx starts a new transaction.
func (r *AdvanceAllocationRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pool.Begin(ctx)
}

// CreateTx inserts an advance allocation record within a transaction.
func (r *AdvanceAllocationRepository) CreateTx(ctx context.Context, tx pgx.Tx, a *model.AdvanceAllocation) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	if a.CreatedAt.IsZero() {
		a.CreatedAt = time.Now()
	}
	_, err := tx.Exec(ctx, `
		INSERT INTO advance_allocations (id, tenant_id, advance_id, advance_type, target_id,
			target_type, allocated_amount, allocation_date, voucher_id, voucher_no, remark,
			created_by, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		a.ID, a.TenantID, a.AdvanceID, a.AdvanceType, a.TargetID, a.TargetType,
		a.AllocatedAmount.String(), a.AllocationDate, a.VoucherID, a.VoucherNo, a.Remark,
		a.CreatedBy, a.CreatedAt)
	return err
}

// MarkAllocationReversed sets reversed_at on an advance allocation.
func (r *AdvanceAllocationRepository) MarkAllocationReversed(ctx context.Context, tx pgx.Tx, allocationID uuid.UUID) error {
	_, err := tx.Exec(ctx, `
		UPDATE advance_allocations SET reversed_at = NOW()
		WHERE id = $1`,
		allocationID)
	return err
}

func (r *AdvanceAllocationRepository) SetVoucher(ctx context.Context, id uuid.UUID, voucherNo string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE advance_allocations SET voucher_id = $2, voucher_no = $3
		WHERE id = $1`,
		id, uuid.Nil, voucherNo)
	return err
}

func (r *AdvanceAllocationRepository) Create(ctx context.Context, a *model.AdvanceAllocation) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	if a.CreatedAt.IsZero() {
		a.CreatedAt = time.Now()
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO advance_allocations (id, tenant_id, advance_id, advance_type, target_id,
			target_type, allocated_amount, allocation_date, voucher_id, voucher_no, remark,
			created_by, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		a.ID, a.TenantID, a.AdvanceID, a.AdvanceType, a.TargetID, a.TargetType,
		a.AllocatedAmount, a.AllocationDate, a.VoucherID, a.VoucherNo, a.Remark,
		a.CreatedBy, a.CreatedAt)
	return err
}

func (r *AdvanceAllocationRepository) ListByAdvance(ctx context.Context, tenantID, advanceID uuid.UUID) ([]*model.AdvanceAllocation, error) {
	return r.listByColumn(ctx, tenantID, "advance_id", advanceID)
}

func (r *AdvanceAllocationRepository) ListByTarget(ctx context.Context, tenantID, targetID uuid.UUID) ([]*model.AdvanceAllocation, error) {
	return r.listByColumn(ctx, tenantID, "target_id", targetID)
}

func (r *AdvanceAllocationRepository) listByColumn(ctx context.Context, tenantID uuid.UUID, column string, value uuid.UUID) ([]*model.AdvanceAllocation, error) {
	var query string
	var args []interface{}
	if column == "advance_id" {
		query = `SELECT id, tenant_id, advance_id, advance_type, target_id, target_type,
			allocated_amount, allocation_date, voucher_id, voucher_no, remark, created_by, created_at
			FROM advance_allocations WHERE tenant_id = $1 AND advance_id = $2
			ORDER BY allocation_date DESC, created_at DESC`
		args = []interface{}{tenantID, value}
	} else {
		query = `SELECT id, tenant_id, advance_id, advance_type, target_id, target_type,
			allocated_amount, allocation_date, voucher_id, voucher_no, remark, created_by, created_at
			FROM advance_allocations WHERE tenant_id = $1 AND target_id = $2
			ORDER BY allocation_date DESC, created_at DESC`
		args = []interface{}{tenantID, value}
	}
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*model.AdvanceAllocation
	for rows.Next() {
		var a model.AdvanceAllocation
		if err := rows.Scan(&a.ID, &a.TenantID, &a.AdvanceID, &a.AdvanceType, &a.TargetID, &a.TargetType,
			&a.AllocatedAmount, &a.AllocationDate, &a.VoucherID, &a.VoucherNo, &a.Remark,
			&a.CreatedBy, &a.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &a)
	}
	return list, rows.Err()
}
