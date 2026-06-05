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
		SELECT id, tenant_id, company_id, supplier_id, invoice_id, invoice_no,
			amount, due_date, status, source_type, created_by, created_at, confirmed_at, confirmed_by,
			approved_by, approved_at
		FROM ap_invoices WHERE tenant_id = $1`
	args := []interface{}{tenantID}

	if status != nil {
		query += " AND status = $2"
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
