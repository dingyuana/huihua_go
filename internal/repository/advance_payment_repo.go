package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/model"
)

type AdvancePaymentRepository struct {
	pool *pgxpool.Pool
}

// SupplierAdvanceSummary aggregates advance payment balances per supplier.
type SupplierAdvanceSummary struct {
	SupplierID   uuid.UUID
	SupplierName string
	TotalAdvance decimal.Decimal
	Allocated    decimal.Decimal
	Outstanding  decimal.Decimal
	Count        int
}

func NewAdvancePaymentRepository(pool *pgxpool.Pool) *AdvancePaymentRepository {
	return &AdvancePaymentRepository{pool: pool}
}

func (r *AdvancePaymentRepository) Create(ctx context.Context, a *model.AdvancePayment) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	if a.CreatedAt.IsZero() {
		a.CreatedAt = time.Now()
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO advance_payments (id, tenant_id, company_id, supplier_id, advance_no,
			amount, allocated_amount, outstanding_amount, paid_date, due_date, status,
			source_type, bank_account_id, reference_no, remark, voucher_id, voucher_no,
			created_by, created_at, confirmed_by, confirmed_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)`,
		a.ID, a.TenantID, a.CompanyID, a.SupplierID, a.AdvanceNo,
		a.Amount, a.AllocatedAmount, a.OutstandingAmount, a.PaidDate, a.DueDate, a.Status,
		a.SourceType, a.BankAccountID, a.ReferenceNo, a.Remark, a.VoucherID, a.VoucherNo,
		a.CreatedBy, a.CreatedAt, a.ConfirmedBy, a.ConfirmedAt)
	return err
}

func (r *AdvancePaymentRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*model.AdvancePayment, error) {
	var a model.AdvancePayment
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, company_id, supplier_id, advance_no, amount, allocated_amount,
			outstanding_amount, paid_date, due_date, status, source_type, bank_account_id,
			reference_no, remark, voucher_id, voucher_no, created_by, created_at,
			confirmed_by, confirmed_at, reversed_by, reversed_at
		FROM advance_payments WHERE tenant_id = $1 AND id = $2`,
		tenantID, id).
		Scan(&a.ID, &a.TenantID, &a.CompanyID, &a.SupplierID, &a.AdvanceNo, &a.Amount,
			&a.AllocatedAmount, &a.OutstandingAmount, &a.PaidDate, &a.DueDate, &a.Status,
			&a.SourceType, &a.BankAccountID, &a.ReferenceNo, &a.Remark, &a.VoucherID,
			&a.VoucherNo, &a.CreatedBy, &a.CreatedAt, &a.ConfirmedBy, &a.ConfirmedAt,
			&a.ReversedBy, &a.ReversedAt)
	if err != nil {
		return nil, nil
	}
	return &a, nil
}

func (r *AdvancePaymentRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, status *string) ([]*model.AdvancePayment, error) {
	query := `
		SELECT id, tenant_id, company_id, supplier_id, advance_no, amount, allocated_amount,
			outstanding_amount, paid_date, due_date, status, source_type, bank_account_id,
			reference_no, remark, voucher_id, voucher_no, created_by, created_at,
			confirmed_by, confirmed_at, reversed_by, reversed_at
		FROM advance_payments WHERE tenant_id = $1`
	args := []interface{}{tenantID}
	if status != nil {
		query += " AND status = $2"
		args = append(args, *status)
	}
	query += " ORDER BY paid_date DESC, created_at DESC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*model.AdvancePayment
	for rows.Next() {
		var a model.AdvancePayment
		if err := rows.Scan(&a.ID, &a.TenantID, &a.CompanyID, &a.SupplierID, &a.AdvanceNo, &a.Amount,
			&a.AllocatedAmount, &a.OutstandingAmount, &a.PaidDate, &a.DueDate, &a.Status,
			&a.SourceType, &a.BankAccountID, &a.ReferenceNo, &a.Remark, &a.VoucherID,
			&a.VoucherNo, &a.CreatedBy, &a.CreatedAt, &a.ConfirmedBy, &a.ConfirmedAt,
			&a.ReversedBy, &a.ReversedAt); err != nil {
			return nil, err
		}
		list = append(list, &a)
	}
	return list, rows.Err()
}

func (r *AdvancePaymentRepository) ListOutstanding(ctx context.Context, tenantID, supplierID uuid.UUID) ([]*model.AdvancePayment, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, company_id, supplier_id, advance_no, amount, allocated_amount,
			outstanding_amount, paid_date, due_date, status, source_type, bank_account_id,
			reference_no, remark, voucher_id, voucher_no, created_by, created_at,
			confirmed_by, confirmed_at, reversed_by, reversed_at
		FROM advance_payments
		WHERE tenant_id = $1 AND supplier_id = $2 AND outstanding_amount > 0
			AND status IN ('confirmed','partially_allocated')
		ORDER BY paid_date ASC`,
		tenantID, supplierID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*model.AdvancePayment
	for rows.Next() {
		var a model.AdvancePayment
		if err := rows.Scan(&a.ID, &a.TenantID, &a.CompanyID, &a.SupplierID, &a.AdvanceNo, &a.Amount,
			&a.AllocatedAmount, &a.OutstandingAmount, &a.PaidDate, &a.DueDate, &a.Status,
			&a.SourceType, &a.BankAccountID, &a.ReferenceNo, &a.Remark, &a.VoucherID,
			&a.VoucherNo, &a.CreatedBy, &a.CreatedAt, &a.ConfirmedBy, &a.ConfirmedAt,
			&a.ReversedBy, &a.ReversedAt); err != nil {
			return nil, err
		}
		list = append(list, &a)
	}
	return list, rows.Err()
}

func (r *AdvancePaymentRepository) IncrementAllocated(ctx context.Context, tenantID, id uuid.UUID, delta float64) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE advance_payments
		SET allocated_amount = allocated_amount + $3,
			outstanding_amount = amount - (allocated_amount + $3)
		WHERE tenant_id = $1 AND id = $2`,
		tenantID, id, delta)
	return err
}

func (r *AdvancePaymentRepository) UpdateStatus(ctx context.Context, tenantID, id uuid.UUID, status string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE advance_payments SET status = $3
		WHERE tenant_id = $1 AND id = $2`,
		tenantID, id, status)
	return err
}

func (r *AdvancePaymentRepository) MarkConfirmed(ctx context.Context, tenantID, id, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE advance_payments SET confirmed_by = $3, confirmed_at = NOW()
		WHERE tenant_id = $1 AND id = $2`,
		tenantID, id, userID)
	return err
}

func (r *AdvancePaymentRepository) SetVoucher(ctx context.Context, tenantID, id, voucherID uuid.UUID, voucherNo string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE advance_payments SET voucher_id = $3, voucher_no = $4
		WHERE tenant_id = $1 AND id = $2`,
		tenantID, id, voucherID, voucherNo)
	return err
}

func (r *AdvancePaymentRepository) GenerateAdvanceNo(ctx context.Context, tenantID uuid.UUID, prefix string) (string, error) {
	var count int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM advance_payments WHERE tenant_id = $1`,
		tenantID).Scan(&count)
	if err != nil {
		return "", err
	}
	year := time.Now().Format("2006")
	return prefix + "-" + year + "-" + padSeq(count+1, 4), nil
}

// GetSupplierSummary returns per-supplier aggregated advance payment balances.
// Only confirmed / partially-allocated / allocated records are included
// (draft and reversed are excluded). If companyID is uuid.Nil, all companies
// under the tenant are included.
func (r *AdvancePaymentRepository) GetSupplierSummary(
	ctx context.Context, tenantID, companyID uuid.UUID,
) ([]SupplierAdvanceSummary, error) {
	query := `
		SELECT  a.supplier_id,
		        COALESCE(p.name, '')            AS supplier_name,
		        COALESCE(SUM(a.amount), 0)       AS total_advance,
		        COALESCE(SUM(a.allocated_amount), 0) AS allocated,
		        COALESCE(SUM(a.outstanding_amount), 0) AS outstanding,
		        COUNT(*)                         AS cnt
		FROM advance_payments a
		LEFT JOIN parties p
		       ON p.tenant_id = a.tenant_id AND p.id = a.supplier_id
		WHERE a.tenant_id = $1
		  AND a.status NOT IN ('draft', 'reversed')`
	args := []interface{}{tenantID}
	if companyID != uuid.Nil {
		query += " AND a.company_id = $2"
		args = append(args, companyID)
	}
	query += `
		GROUP BY a.supplier_id, p.name
		ORDER BY outstanding DESC, supplier_name ASC`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []SupplierAdvanceSummary
	for rows.Next() {
		var s SupplierAdvanceSummary
		if err := rows.Scan(
			&s.SupplierID, &s.SupplierName,
			&s.TotalAdvance, &s.Allocated, &s.Outstanding, &s.Count,
		); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
