package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
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
			amount, due_date, status, source_type, created_by, created_at, confirmed_at, confirmed_by,
			approved_by, approved_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`,
		ap.ID, ap.TenantID, ap.CompanyID, ap.SupplierID, ap.InvoiceID, ap.InvoiceNo,
		ap.Amount, ap.DueDate, ap.Status, ap.SourceType, ap.CreatedBy, ap.CreatedAt,
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

func (r *ApInvoiceRepository) IncrementPaid(ctx context.Context, tenantID, id uuid.UUID, delta float64, newStatus string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE ap_invoices
		SET paid_amount = paid_amount + $3,
			outstanding_amount = amount - (paid_amount + $3),
			status = $4,
			last_allocation_at = NOW()
		WHERE tenant_id = $1 AND id = $2`,
		tenantID, id, delta, newStatus)
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
