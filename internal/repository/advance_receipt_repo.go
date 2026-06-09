package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/model"
)

type AdvanceReceiptRepository struct {
	pool *pgxpool.Pool
}

// CustomerAdvanceSummary aggregates advance receipt balances per customer.
type CustomerAdvanceSummary struct {
	CustomerID   uuid.UUID
	CustomerName string
	TotalAdvance decimal.Decimal
	Allocated    decimal.Decimal
	Outstanding  decimal.Decimal
	Count        int
}

func NewAdvanceReceiptRepository(pool *pgxpool.Pool) *AdvanceReceiptRepository {
	return &AdvanceReceiptRepository{pool: pool}
}

// BeginTx starts a new transaction.
func (r *AdvanceReceiptRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pool.Begin(ctx)
}

// LockForUpdate locks an advance_receipts row with SELECT FOR UPDATE.
func (r *AdvanceReceiptRepository) LockForUpdate(ctx context.Context, tx pgx.Tx, tenantID, id uuid.UUID) error {
	_, err := tx.Exec(ctx, `
		SELECT id FROM advance_receipts
		WHERE tenant_id = $1 AND id = $2
		FOR UPDATE`,
		tenantID, id)
	return err
}

// IncrementAllocatedTx atomically increases allocated_amount within a transaction.
func (r *AdvanceReceiptRepository) IncrementAllocatedTx(ctx context.Context, tx pgx.Tx, tenantID, id uuid.UUID, delta string) error {
	_, err := tx.Exec(ctx, `
		UPDATE advance_receipts
		SET allocated_amount = allocated_amount + $3,
			outstanding_amount = amount - (allocated_amount + $3)
		WHERE tenant_id = $1 AND id = $2`,
		tenantID, id, delta)
	return err
}

// UpdateOutstandingAmountTx updates the outstanding_amount directly within a transaction.
func (r *AdvanceReceiptRepository) UpdateOutstandingAmountTx(ctx context.Context, tx pgx.Tx, tenantID, id uuid.UUID, outstanding string) error {
	_, err := tx.Exec(ctx, `
		UPDATE advance_receipts
		SET outstanding_amount = $3,
			allocated_amount = amount - $3
		WHERE tenant_id = $1 AND id = $2`,
		tenantID, id, outstanding)
	if err != nil {
		return fmt.Errorf("update advance receipt outstanding (tx): %w", err)
	}
	return nil
}

func (r *AdvanceReceiptRepository) Create(ctx context.Context, a *model.AdvanceReceipt) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	if a.CreatedAt.IsZero() {
		a.CreatedAt = time.Now()
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO advance_receipts (id, tenant_id, company_id, customer_id, advance_no,
			amount, allocated_amount, outstanding_amount, received_date, due_date, status,
			source_type, bank_account_id, reference_no, remark, voucher_id, voucher_no,
			created_by, created_at, confirmed_by, confirmed_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)`,
		a.ID, a.TenantID, a.CompanyID, a.CustomerID, a.AdvanceNo,
		a.Amount, a.AllocatedAmount, a.OutstandingAmount, a.ReceivedDate, a.DueDate, a.Status,
		a.SourceType, a.BankAccountID, a.ReferenceNo, a.Remark, a.VoucherID, a.VoucherNo,
		a.CreatedBy, a.CreatedAt, a.ConfirmedBy, a.ConfirmedAt)
	return err
}

func (r *AdvanceReceiptRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*model.AdvanceReceipt, error) {
	var a model.AdvanceReceipt
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, company_id, customer_id, advance_no, amount, allocated_amount,
			outstanding_amount, received_date, due_date, status, source_type, bank_account_id,
			reference_no, remark, voucher_id, voucher_no, created_by, created_at,
			confirmed_by, confirmed_at, reversed_by, reversed_at
		FROM advance_receipts WHERE tenant_id = $1 AND id = $2`,
		tenantID, id).
		Scan(&a.ID, &a.TenantID, &a.CompanyID, &a.CustomerID, &a.AdvanceNo, &a.Amount,
			&a.AllocatedAmount, &a.OutstandingAmount, &a.ReceivedDate, &a.DueDate, &a.Status,
			&a.SourceType, &a.BankAccountID, &a.ReferenceNo, &a.Remark, &a.VoucherID,
			&a.VoucherNo, &a.CreatedBy, &a.CreatedAt, &a.ConfirmedBy, &a.ConfirmedAt,
			&a.ReversedBy, &a.ReversedAt)
	if err != nil {
		return nil, nil
	}
	return &a, nil
}

func (r *AdvanceReceiptRepository) GetByNo(ctx context.Context, tenantID uuid.UUID, advanceNo string) (*model.AdvanceReceipt, error) {
	var a model.AdvanceReceipt
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, company_id, customer_id, advance_no, amount, allocated_amount,
			outstanding_amount, received_date, due_date, status, source_type, bank_account_id,
			reference_no, remark, voucher_id, voucher_no, created_by, created_at,
			confirmed_by, confirmed_at, reversed_by, reversed_at
		FROM advance_receipts WHERE tenant_id = $1 AND advance_no = $2`,
		tenantID, advanceNo).
		Scan(&a.ID, &a.TenantID, &a.CompanyID, &a.CustomerID, &a.AdvanceNo, &a.Amount,
			&a.AllocatedAmount, &a.OutstandingAmount, &a.ReceivedDate, &a.DueDate, &a.Status,
			&a.SourceType, &a.BankAccountID, &a.ReferenceNo, &a.Remark, &a.VoucherID,
			&a.VoucherNo, &a.CreatedBy, &a.CreatedAt, &a.ConfirmedBy, &a.ConfirmedAt,
			&a.ReversedBy, &a.ReversedAt)
	if err != nil {
		return nil, nil
	}
	return &a, nil
}

func (r *AdvanceReceiptRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, status *string) ([]*model.AdvanceReceipt, error) {
	query := `
		SELECT id, tenant_id, company_id, customer_id, advance_no, amount, allocated_amount,
			outstanding_amount, received_date, due_date, status, source_type, bank_account_id,
			reference_no, remark, voucher_id, voucher_no, created_by, created_at,
			confirmed_by, confirmed_at, reversed_by, reversed_at
		FROM advance_receipts WHERE tenant_id = $1`
	args := []interface{}{tenantID}
	if status != nil {
		query += " AND status = $2"
		args = append(args, *status)
	}
	query += " ORDER BY received_date DESC, created_at DESC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*model.AdvanceReceipt
	for rows.Next() {
		var a model.AdvanceReceipt
		if err := rows.Scan(&a.ID, &a.TenantID, &a.CompanyID, &a.CustomerID, &a.AdvanceNo, &a.Amount,
			&a.AllocatedAmount, &a.OutstandingAmount, &a.ReceivedDate, &a.DueDate, &a.Status,
			&a.SourceType, &a.BankAccountID, &a.ReferenceNo, &a.Remark, &a.VoucherID,
			&a.VoucherNo, &a.CreatedBy, &a.CreatedAt, &a.ConfirmedBy, &a.ConfirmedAt,
			&a.ReversedBy, &a.ReversedAt); err != nil {
			return nil, err
		}
		list = append(list, &a)
	}
	return list, rows.Err()
}

func (r *AdvanceReceiptRepository) ListOutstanding(ctx context.Context, tenantID, customerID uuid.UUID) ([]*model.AdvanceReceipt, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, company_id, customer_id, advance_no, amount, allocated_amount,
			outstanding_amount, received_date, due_date, status, source_type, bank_account_id,
			reference_no, remark, voucher_id, voucher_no, created_by, created_at,
			confirmed_by, confirmed_at, reversed_by, reversed_at
		FROM advance_receipts
		WHERE tenant_id = $1 AND customer_id = $2 AND outstanding_amount > 0
			AND status IN ('confirmed','partially_allocated')
		ORDER BY received_date ASC`,
		tenantID, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*model.AdvanceReceipt
	for rows.Next() {
		var a model.AdvanceReceipt
		if err := rows.Scan(&a.ID, &a.TenantID, &a.CompanyID, &a.CustomerID, &a.AdvanceNo, &a.Amount,
			&a.AllocatedAmount, &a.OutstandingAmount, &a.ReceivedDate, &a.DueDate, &a.Status,
			&a.SourceType, &a.BankAccountID, &a.ReferenceNo, &a.Remark, &a.VoucherID,
			&a.VoucherNo, &a.CreatedBy, &a.CreatedAt, &a.ConfirmedBy, &a.ConfirmedAt,
			&a.ReversedBy, &a.ReversedAt); err != nil {
			return nil, err
		}
		list = append(list, &a)
	}
	return list, rows.Err()
}

// IncrementAllocated atomically increases allocated_amount and updates status.
func (r *AdvanceReceiptRepository) IncrementAllocated(ctx context.Context, tenantID, id uuid.UUID, delta float64) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE advance_receipts
		SET allocated_amount = allocated_amount + $3,
			outstanding_amount = amount - (allocated_amount + $3)
		WHERE tenant_id = $1 AND id = $2`,
		tenantID, id, delta)
	return err
}

func (r *AdvanceReceiptRepository) UpdateStatus(ctx context.Context, tenantID, id uuid.UUID, status string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE advance_receipts SET status = $3
		WHERE tenant_id = $1 AND id = $2`,
		tenantID, id, status)
	return err
}

func (r *AdvanceReceiptRepository) MarkConfirmed(ctx context.Context, tenantID, id, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE advance_receipts SET confirmed_by = $3, confirmed_at = NOW()
		WHERE tenant_id = $1 AND id = $2`,
		tenantID, id, userID)
	return err
}

func (r *AdvanceReceiptRepository) SetVoucher(ctx context.Context, tenantID, id, voucherID uuid.UUID, voucherNo string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE advance_receipts SET voucher_id = $3, voucher_no = $4
		WHERE tenant_id = $1 AND id = $2`,
		tenantID, id, voucherID, voucherNo)
	return err
}

func (r *AdvanceReceiptRepository) GenerateAdvanceNo(ctx context.Context, tenantID uuid.UUID, prefix string) (string, error) {
	var count int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM advance_receipts WHERE tenant_id = $1`,
		tenantID).Scan(&count)
	if err != nil {
		return "", err
	}
	year := time.Now().Format("2006")
	return prefix + "-" + year + "-" + padSeq(count+1, 4), nil
}

// GetCustomerSummary returns per-customer aggregated advance receipt balances.
// Only confirmed / partially-allocated / allocated records are included
// (draft and reversed are excluded). If companyID is uuid.Nil, all companies
// under the tenant are included.
func (r *AdvanceReceiptRepository) GetCustomerSummary(
	ctx context.Context, tenantID, companyID uuid.UUID,
) ([]CustomerAdvanceSummary, error) {
	query := `
		SELECT  a.customer_id,
		        COALESCE(p.name, '')            AS customer_name,
		        COALESCE(SUM(a.amount), 0)       AS total_advance,
		        COALESCE(SUM(a.allocated_amount), 0) AS allocated,
		        COALESCE(SUM(a.outstanding_amount), 0) AS outstanding,
		        COUNT(*)                         AS cnt
		FROM advance_receipts a
		LEFT JOIN parties p
		       ON p.tenant_id = a.tenant_id AND p.id = a.customer_id
		WHERE a.tenant_id = $1
		  AND a.status NOT IN ('draft', 'reversed')`
	args := []interface{}{tenantID}
	if companyID != uuid.Nil {
		query += " AND a.company_id = $2"
		args = append(args, companyID)
	}
	query += `
		GROUP BY a.customer_id, p.name
		ORDER BY outstanding DESC, customer_name ASC`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []CustomerAdvanceSummary
	for rows.Next() {
		var s CustomerAdvanceSummary
		if err := rows.Scan(
			&s.CustomerID, &s.CustomerName,
			&s.TotalAdvance, &s.Allocated, &s.Outstanding, &s.Count,
		); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func padSeq(n int, width int) string {
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	for len(s) < width {
		s = "0" + s
	}
	return s
}
