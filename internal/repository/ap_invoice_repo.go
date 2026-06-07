package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/model"
)

type ApInvoiceRepository struct {
	pool *pgxpool.Pool
}

func NewApInvoiceRepository(pool *pgxpool.Pool) *ApInvoiceRepository {
	return &ApInvoiceRepository{pool: pool}
}

func (r *ApInvoiceRepository) Create(ctx context.Context, ap *model.ApInvoice) error {
	if ap.ID == uuid.Nil {
		ap.ID = uuid.New()
	}
	if ap.CreatedAt.IsZero() {
		ap.CreatedAt = time.Now()
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO ap_invoices (id, tenant_id, company_id, supplier_id, invoice_id, invoice_no,
			amount, paid_amount, outstanding_amount, due_date, status, source_type, created_by, created_at, confirmed_at, confirmed_by,
			approved_by, approved_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)`,
		ap.ID, ap.TenantID, ap.CompanyID, ap.SupplierID, ap.InvoiceID, ap.InvoiceNo,
		ap.Amount, ap.PaidAmount, ap.OutstandingAmount, ap.DueDate, ap.Status, ap.SourceType, ap.CreatedBy, ap.CreatedAt,
		ap.ConfirmedAt, ap.ConfirmedBy, ap.ApprovedBy, ap.ApprovedAt)
	return err
}

func (r *ApInvoiceRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*model.ApInvoice, error) {
	var ap model.ApInvoice
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, company_id, supplier_id, invoice_id, invoice_no,
			amount, due_date, status, source_type, created_by, created_at, confirmed_at, confirmed_by,
			approved_by, approved_at
		FROM ap_invoices WHERE tenant_id = $1 AND id = $2`,
		tenantID, id).
		Scan(&ap.ID, &ap.TenantID, &ap.CompanyID, &ap.SupplierID, &ap.InvoiceID, &ap.InvoiceNo,
			&ap.Amount, &ap.DueDate, &ap.Status, &ap.SourceType, &ap.CreatedBy, &ap.CreatedAt,
			&ap.ConfirmedAt, &ap.ConfirmedBy, &ap.ApprovedBy, &ap.ApprovedAt)
	if err != nil {
		return nil, nil
	}
	return &ap, nil
}

func (r *ApInvoiceRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, status *string) ([]*model.ApInvoice, error) {
	query := `
		SELECT a.id, a.tenant_id, a.company_id, a.supplier_id, a.invoice_id, a.invoice_no,
			a.amount, a.due_date, a.status, a.source_type, a.created_by, a.created_at,
			a.confirmed_at, a.confirmed_by, a.approved_by, a.approved_at,
			COALESCE(s.remark, '') AS remark
		FROM ap_invoices a
		LEFT JOIN sales_invoices s ON s.id = a.invoice_id
		WHERE a.tenant_id = $1`
	args := []interface{}{tenantID}

	if status != nil {
		query += " AND a.status = $2"
		args = append(args, *status)
	}

	query += " ORDER BY a.created_at DESC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var aps []*model.ApInvoice
	for rows.Next() {
		var ap model.ApInvoice
		var remark string
		if err := rows.Scan(&ap.ID, &ap.TenantID, &ap.CompanyID, &ap.SupplierID, &ap.InvoiceID, &ap.InvoiceNo,
			&ap.Amount, &ap.DueDate, &ap.Status, &ap.SourceType, &ap.CreatedBy, &ap.CreatedAt,
			&ap.ConfirmedAt, &ap.ConfirmedBy, &ap.ApprovedBy, &ap.ApprovedAt, &remark); err != nil {
			return nil, err
		}
		if remark != "" {
			ap.Remark = &remark
		}
		aps = append(aps, &ap)
	}
	return aps, rows.Err()
}

func (r *ApInvoiceRepository) GetSourceInvoiceRemark(ctx context.Context, tenantID, invoiceID uuid.UUID) (string, error) {
	var remark *string
	err := r.pool.QueryRow(ctx, `
		SELECT remark FROM sales_invoices WHERE tenant_id = $1 AND id = $2`,
		tenantID, invoiceID).Scan(&remark)
	if err != nil {
		return "", err
	}
	if remark == nil {
		return "", nil
	}
	return *remark, nil
}

func (r *ApInvoiceRepository) ListBySupplier(ctx context.Context, tenantID, supplierID uuid.UUID, status *string) ([]*model.ApInvoice, error) {
	query := `
		SELECT id, tenant_id, company_id, supplier_id, invoice_id, invoice_no,
			amount, due_date, status, source_type, created_by, created_at, confirmed_at, confirmed_by,
			approved_by, approved_at
		FROM ap_invoices WHERE tenant_id = $1 AND supplier_id = $2`
	args := []interface{}{tenantID, supplierID}

	if status != nil {
		query += " AND status = $3"
		args = append(args, *status)
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var aps []*model.ApInvoice
	for rows.Next() {
		var ap model.ApInvoice
		if err := rows.Scan(&ap.ID, &ap.TenantID, &ap.CompanyID, &ap.SupplierID, &ap.InvoiceID, &ap.InvoiceNo,
			&ap.Amount, &ap.DueDate, &ap.Status, &ap.SourceType, &ap.CreatedBy, &ap.CreatedAt,
			&ap.ConfirmedAt, &ap.ConfirmedBy, &ap.ApprovedBy, &ap.ApprovedAt); err != nil {
			return nil, err
		}
		aps = append(aps, &ap)
	}
	return aps, rows.Err()
}

func (r *ApInvoiceRepository) ListByInvoiceID(ctx context.Context, tenantID, invoiceID uuid.UUID) (*model.ApInvoice, error) {
	var ap model.ApInvoice
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, company_id, supplier_id, invoice_id, invoice_no,
			amount, due_date, status, source_type, created_by, created_at, confirmed_at, confirmed_by,
			approved_by, approved_at
		FROM ap_invoices WHERE tenant_id = $1 AND invoice_id = $2`,
		tenantID, invoiceID).
		Scan(&ap.ID, &ap.TenantID, &ap.CompanyID, &ap.SupplierID, &ap.InvoiceID, &ap.InvoiceNo,
			&ap.Amount, &ap.DueDate, &ap.Status, &ap.SourceType, &ap.CreatedBy, &ap.CreatedAt,
			&ap.ConfirmedAt, &ap.ConfirmedBy, &ap.ApprovedBy, &ap.ApprovedAt)
	if err != nil {
		return nil, nil
	}
	return &ap, nil
}

// Confirm confirms an AP invoice, setting confirmed_at, confirmed_by, and status = 'confirmed'.
func (r *ApInvoiceRepository) Confirm(ctx context.Context, tenantID, id, confirmedBy uuid.UUID) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx, `
		UPDATE ap_invoices SET confirmed_at = $3, confirmed_by = $4, status = $5
		WHERE tenant_id = $1 AND id = $2`,
		tenantID, id, now, confirmedBy, model.ApInvoiceStatusConfirmed)
	return err
}

// Approve approves an AP invoice (after confirmation), setting approved_at, approved_by, and status = 'approved'.
func (r *ApInvoiceRepository) Approve(ctx context.Context, tenantID, id, approvedBy uuid.UUID) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx, `
		UPDATE ap_invoices SET approved_at = $3, approved_by = $4, status = $5
		WHERE tenant_id = $1 AND id = $2`,
		tenantID, id, now, approvedBy, model.ApInvoiceStatusConfirmed)
	return err
}

// Update updates editable fields on a draft ApInvoice (supplier_id, invoice_no, amount, due_date).
func (r *ApInvoiceRepository) Update(ctx context.Context, tenantID, id uuid.UUID, ap *model.ApInvoice) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE ap_invoices
		SET supplier_id = $3, invoice_no = $4, amount = $5, due_date = $6
		WHERE tenant_id = $1 AND id = $2`,
		tenantID, id, ap.SupplierID, ap.InvoiceNo, ap.Amount, ap.DueDate)
	return err
}

// ListOutstanding returns AP invoices with outstanding balance > 0 within a tenant.
func (r *ApInvoiceRepository) ListOutstanding(ctx context.Context, tenantID uuid.UUID) ([]*model.ApInvoice, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, company_id, supplier_id, invoice_id, invoice_no, amount,
			paid_amount, outstanding_amount, due_date, status, source_type, created_by, created_at,
			confirmed_at, confirmed_by, approved_by, approved_at
		FROM ap_invoices
		WHERE tenant_id = $1 AND outstanding_amount > 0
			AND status IN ('confirmed','partially_paid')
		ORDER BY due_date ASC NULLS LAST, created_at ASC`,
		tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*model.ApInvoice
	for rows.Next() {
		var ap model.ApInvoice
		if err := rows.Scan(&ap.ID, &ap.TenantID, &ap.CompanyID, &ap.SupplierID, &ap.InvoiceID, &ap.InvoiceNo,
			&ap.Amount, &ap.PaidAmount, &ap.OutstandingAmount, &ap.DueDate, &ap.Status, &ap.SourceType,
			&ap.CreatedBy, &ap.CreatedAt, &ap.ConfirmedAt, &ap.ConfirmedBy, &ap.ApprovedBy, &ap.ApprovedAt); err != nil {
			return nil, err
		}
		list = append(list, &ap)
	}
	return list, rows.Err()
}

// IncrementPaid increments paid_amount and recalculates outstanding_amount.
func (r *ApInvoiceRepository) IncrementPaid(ctx context.Context, tenantID, id uuid.UUID, delta decimal.Decimal, newStatus string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE ap_invoices
		SET paid_amount = paid_amount + $3,
			outstanding_amount = amount - (paid_amount + $3),
			status = $4,
			last_allocation_at = NOW()
		WHERE tenant_id = $1 AND id = $2`,
		tenantID, id, delta.String(), newStatus)
	return err
}

// UpdateStatus updates the status of an ApInvoice.
func (r *ApInvoiceRepository) UpdateStatus(ctx context.Context, tenantID, id uuid.UUID, status string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE ap_invoices SET status = $3
		WHERE tenant_id = $1 AND id = $2`,
		tenantID, id, status)
	return err
}

// Delete removes an ApInvoice by ID within a tenant.
func (r *ApInvoiceRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `
		DELETE FROM ap_invoices WHERE tenant_id = $1 AND id = $2`,
		tenantID, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("ap_invoice not found")
	}
	return nil
}

// SetVoucherID links a voucher to an ApInvoice.
func (r *ApInvoiceRepository) SetVoucherID(ctx context.Context, tenantID, id, voucherID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE ap_invoices SET voucher_id = $3
		WHERE tenant_id = $1 AND id = $2`,
		tenantID, id, voucherID)
	return err
}

// UpdateOutstandingAmount directly updates the outstanding amount on an ApInvoice.
func (r *ApInvoiceRepository) UpdateOutstandingAmount(ctx context.Context, tenantID, id uuid.UUID, outstanding decimal.Decimal) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE ap_invoices
		SET outstanding_amount = $3,
			last_allocation_at = NOW(),
			status = CASE
				WHEN $3 <= 0 THEN 'paid'
				WHEN $3 < amount THEN 'partially_paid'
				ELSE status
			END
		WHERE tenant_id = $1 AND id = $2`,
		tenantID, id, outstanding.String())
	return err
}

func (r *ApInvoiceRepository) ListUnpaidBySupplier(ctx context.Context, tenantID, supplierID uuid.UUID) ([]*model.ApInvoice, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, company_id, supplier_id, invoice_id, invoice_no, amount,
			paid_amount, outstanding_amount, due_date, status, source_type, created_by, created_at,
			confirmed_at, confirmed_by, approved_by, approved_at
		FROM ap_invoices
		WHERE tenant_id = $1 AND supplier_id = $2 AND outstanding_amount > 0
			AND status IN ('confirmed','partially_paid')
		ORDER BY due_date ASC NULLS LAST, created_at ASC`,
		tenantID, supplierID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*model.ApInvoice
	for rows.Next() {
		var ap model.ApInvoice
		if err := rows.Scan(&ap.ID, &ap.TenantID, &ap.CompanyID, &ap.SupplierID, &ap.InvoiceID, &ap.InvoiceNo,
			&ap.Amount, &ap.PaidAmount, &ap.OutstandingAmount, &ap.DueDate, &ap.Status, &ap.SourceType,
			&ap.CreatedBy, &ap.CreatedAt, &ap.ConfirmedAt, &ap.ConfirmedBy, &ap.ApprovedBy, &ap.ApprovedAt); err != nil {
			return nil, err
		}
		list = append(list, &ap)
	}
	return list, rows.Err()
}
