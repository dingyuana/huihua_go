package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"huihua/finance/internal/model"
)

// ArInvoiceRepository handles ar_invoices table operations.
type ArInvoiceRepository struct {
	pool *pgxpool.Pool
}

// NewArInvoiceRepository creates a new ArInvoiceRepository.
func NewArInvoiceRepository(pool *pgxpool.Pool) *ArInvoiceRepository {
	return &ArInvoiceRepository{pool: pool}
}

// Create inserts a new AR invoice.
func (r *ArInvoiceRepository) Create(ctx context.Context, ar *model.ArInvoice) error {
	if ar.ID == uuid.Nil {
		ar.ID = uuid.New()
	}
	if ar.CreatedAt.IsZero() {
		ar.CreatedAt = time.Now()
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO ar_invoices (id, tenant_id, company_id, customer_id, invoice_id, invoice_no,
			amount, due_date, status, source_type, created_by, created_at, confirmed_at, confirmed_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
		ar.ID, ar.TenantID, ar.CompanyID, ar.CustomerID, ar.InvoiceID, ar.InvoiceNo,
		ar.Amount, ar.DueDate, ar.Status, ar.SourceType, ar.CreatedBy, ar.CreatedAt,
		ar.ConfirmedAt, ar.ConfirmedBy)
	return err
}

// GetByID retrieves an AR invoice by ID within a tenant.
func (r *ArInvoiceRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*model.ArInvoice, error) {
	var ar model.ArInvoice
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, company_id, customer_id, invoice_id, invoice_no,
			amount, due_date, status, source_type, created_by, created_at, confirmed_at, confirmed_by
		FROM ar_invoices WHERE tenant_id = $1 AND id = $2`,
		tenantID, id).
		Scan(&ar.ID, &ar.TenantID, &ar.CompanyID, &ar.CustomerID, &ar.InvoiceID, &ar.InvoiceNo,
			&ar.Amount, &ar.DueDate, &ar.Status, &ar.SourceType, &ar.CreatedBy, &ar.CreatedAt,
			&ar.ConfirmedAt, &ar.ConfirmedBy)
	if err != nil {
		return nil, err
	}
	return &ar, nil
}

// ListByTenant retrieves AR invoices by tenant, optionally filtering by status.
func (r *ArInvoiceRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, status *string) ([]*model.ArInvoice, error) {
	query := `
		SELECT id, tenant_id, company_id, customer_id, invoice_id, invoice_no,
			amount, due_date, status, source_type, created_by, created_at, confirmed_at, confirmed_by
		FROM ar_invoices WHERE tenant_id = $1`
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

	var ars []*model.ArInvoice
	for rows.Next() {
		var ar model.ArInvoice
		if err := rows.Scan(&ar.ID, &ar.TenantID, &ar.CompanyID, &ar.CustomerID, &ar.InvoiceID, &ar.InvoiceNo,
			&ar.Amount, &ar.DueDate, &ar.Status, &ar.SourceType, &ar.CreatedBy, &ar.CreatedAt,
			&ar.ConfirmedAt, &ar.ConfirmedBy); err != nil {
			return nil, err
		}
		ars = append(ars, &ar)
	}
	return ars, rows.Err()
}

// ListByCustomer retrieves AR invoices by customer within a tenant.
func (r *ArInvoiceRepository) ListByCustomer(ctx context.Context, tenantID, customerID uuid.UUID, status *string) ([]*model.ArInvoice, error) {
	query := `
		SELECT id, tenant_id, company_id, customer_id, invoice_id, invoice_no,
			amount, due_date, status, source_type, created_by, created_at, confirmed_at, confirmed_by
		FROM ar_invoices WHERE tenant_id = $1 AND customer_id = $2`
	args := []interface{}{tenantID, customerID}

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

	var ars []*model.ArInvoice
	for rows.Next() {
		var ar model.ArInvoice
		if err := rows.Scan(&ar.ID, &ar.TenantID, &ar.CompanyID, &ar.CustomerID, &ar.InvoiceID, &ar.InvoiceNo,
			&ar.Amount, &ar.DueDate, &ar.Status, &ar.SourceType, &ar.CreatedBy, &ar.CreatedAt,
			&ar.ConfirmedAt, &ar.ConfirmedBy); err != nil {
			return nil, err
		}
		ars = append(ars, &ar)
	}
	return ars, rows.Err()
}

// ListByInvoiceID retrieves an AR invoice by invoice ID within a tenant.
// Returns nil, nil if not found.
func (r *ArInvoiceRepository) ListByInvoiceID(ctx context.Context, tenantID, invoiceID uuid.UUID) (*model.ArInvoice, error) {
	var ar model.ArInvoice
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, company_id, customer_id, invoice_id, invoice_no,
			amount, due_date, status, source_type, created_by, created_at, confirmed_at, confirmed_by
		FROM ar_invoices WHERE tenant_id = $1 AND invoice_id = $2`,
		tenantID, invoiceID).
		Scan(&ar.ID, &ar.TenantID, &ar.CompanyID, &ar.CustomerID, &ar.InvoiceID, &ar.InvoiceNo,
			&ar.Amount, &ar.DueDate, &ar.Status, &ar.SourceType, &ar.CreatedBy, &ar.CreatedAt,
			&ar.ConfirmedAt, &ar.ConfirmedBy)
	if err != nil {
		return nil, nil
	}
	return &ar, nil
}

// UpdateStatus updates the status of an AR invoice.
func (r *ArInvoiceRepository) UpdateStatus(ctx context.Context, tenantID, id uuid.UUID, status string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE ar_invoices SET status = $3
		WHERE tenant_id = $1 AND id = $2`,
		tenantID, id, status)
	return err
}

// Confirm confirms an AR invoice, setting confirmed_at, confirmed_by, and status = 'confirmed'.
func (r *ArInvoiceRepository) Confirm(ctx context.Context, tenantID, id, confirmedBy uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE ar_invoices SET confirmed_at = $3, confirmed_by = $4, status = $5
		WHERE tenant_id = $1 AND id = $2`,
		tenantID, id, time.Now(), confirmedBy, model.ArInvoiceStatusConfirmed)
	return err
}

// BatchCreate inserts multiple AR invoices in a loop.
func (r *ArInvoiceRepository) BatchCreate(ctx context.Context, ars []*model.ArInvoice) error {
	for _, ar := range ars {
		if ar.ID == uuid.Nil {
			ar.ID = uuid.New()
		}
		if ar.CreatedAt.IsZero() {
			ar.CreatedAt = time.Now()
		}
		_, err := r.pool.Exec(ctx, `
			INSERT INTO ar_invoices (id, tenant_id, company_id, customer_id, invoice_id, invoice_no,
				amount, due_date, status, source_type, created_by, created_at, confirmed_at, confirmed_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
			ar.ID, ar.TenantID, ar.CompanyID, ar.CustomerID, ar.InvoiceID, ar.InvoiceNo,
			ar.Amount, ar.DueDate, ar.Status, ar.SourceType, ar.CreatedBy, ar.CreatedAt,
			ar.ConfirmedAt, ar.ConfirmedBy)
		if err != nil {
			return err
		}
	}
	return nil
}