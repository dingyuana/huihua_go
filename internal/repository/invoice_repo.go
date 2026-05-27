package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"huihua/finance/internal/model"
)

// InvoiceRepository handles sales_invoices table operations.
type InvoiceRepository struct {
	pool *pgxpool.Pool
}

// NewInvoiceRepository creates a new InvoiceRepository.
func NewInvoiceRepository(pool *pgxpool.Pool) *InvoiceRepository {
	return &InvoiceRepository{pool: pool}
}

// Create inserts a new invoice.
func (r *InvoiceRepository) Create(ctx context.Context, tenantID uuid.UUID, inv *model.SalesInvoice) (*model.SalesInvoice, error) {
	inv.ID = uuid.New()
	inv.TenantID = tenantID
	inv.CreatedAt = time.Now()

	_, err := r.pool.Exec(ctx, `
		INSERT INTO sales_invoices (id, invoice_no, invoice_type, customer_id, tax_id, company_id,
			tenant_id, posting_date, due_date, total_amount, tax_amount, net_amount, outstanding_amount,
			status, tax_template_id, return_against, is_return, docstatus, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)`,
		inv.ID, inv.InvoiceNo, inv.InvoiceType, inv.CustomerID, inv.TaxID, inv.CompanyID,
		inv.TenantID, inv.PostingDate, inv.DueDate, inv.TotalAmount, inv.TaxAmount, inv.NetAmount,
		inv.OutstandingAmount, inv.Status, inv.TaxTemplateID, inv.ReturnAgainst, inv.IsReturn,
		inv.DocStatus, inv.CreatedBy, inv.CreatedAt)
	if err != nil {
		return nil, err
	}
	return inv, nil
}

// ListByTenant retrieves invoices with filters.
func (r *InvoiceRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, filters model.InvoiceFilter) ([]model.SalesInvoice, error) {
	query := `
		SELECT id, invoice_no, invoice_type, customer_id, tax_id, company_id,
			tenant_id, posting_date, due_date, total_amount, tax_amount, net_amount, outstanding_amount,
			status, tax_template_id, return_against, is_return, docstatus, created_by, created_at
		FROM sales_invoices WHERE tenant_id = $1`
	args := []interface{}{tenantID}
	argIdx := 2

	if filters.CustomerID != nil {
		query += fmt.Sprintf(" AND customer_id = $%d", argIdx)
		args = append(args, *filters.CustomerID)
		argIdx++
	}
	if filters.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, filters.Status)
		argIdx++
	}
	if filters.FromDate != nil {
		query += fmt.Sprintf(" AND posting_date >= $%d", argIdx)
		args = append(args, *filters.FromDate)
		argIdx++
	}
	if filters.ToDate != nil {
		query += fmt.Sprintf(" AND posting_date <= $%d", argIdx)
		args = append(args, *filters.ToDate)
		argIdx++
	}

	query += " ORDER BY posting_date DESC"

	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, filters.Limit)
		argIdx++
	} else {
		query += " LIMIT 100"
	}
	if filters.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIdx)
		args = append(args, filters.Offset)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invoices []model.SalesInvoice
	for rows.Next() {
		var inv model.SalesInvoice
		if err := rows.Scan(&inv.ID, &inv.InvoiceNo, &inv.InvoiceType, &inv.CustomerID, &inv.TaxID,
			&inv.CompanyID, &inv.TenantID, &inv.PostingDate, &inv.DueDate, &inv.TotalAmount,
			&inv.TaxAmount, &inv.NetAmount, &inv.OutstandingAmount, &inv.Status, &inv.TaxTemplateID,
			&inv.ReturnAgainst, &inv.IsReturn, &inv.DocStatus, &inv.CreatedBy, &inv.CreatedAt); err != nil {
			return nil, err
		}
		invoices = append(invoices, inv)
	}
	return invoices, rows.Err()
}

// GetByID retrieves an invoice by ID.
func (r *InvoiceRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*model.SalesInvoice, error) {
	var inv model.SalesInvoice
	err := r.pool.QueryRow(ctx, `
		SELECT id, invoice_no, invoice_type, customer_id, tax_id, company_id,
			tenant_id, posting_date, due_date, total_amount, tax_amount, net_amount, outstanding_amount,
			status, tax_template_id, return_against, is_return, docstatus, created_by, created_at
		FROM sales_invoices WHERE tenant_id = $1 AND id = $2`,
		tenantID, id).
		Scan(&inv.ID, &inv.InvoiceNo, &inv.InvoiceType, &inv.CustomerID, &inv.TaxID,
			&inv.CompanyID, &inv.TenantID, &inv.PostingDate, &inv.DueDate, &inv.TotalAmount,
			&inv.TaxAmount, &inv.NetAmount, &inv.OutstandingAmount, &inv.Status, &inv.TaxTemplateID,
			&inv.ReturnAgainst, &inv.IsReturn, &inv.DocStatus, &inv.CreatedBy, &inv.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

// UpdateStatus updates the status of an invoice.
func (r *InvoiceRepository) UpdateStatus(ctx context.Context, tenantID, id uuid.UUID, status string) error {
	result, err := r.pool.Exec(ctx, `
		UPDATE sales_invoices SET status = $3
		WHERE tenant_id = $1 AND id = $2`,
		tenantID, id, status)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("invoice not found")
	}
	return nil
}

// ImportBatch inserts multiple invoices in a transaction.
func (r *InvoiceRepository) ImportBatch(ctx context.Context, tenantID uuid.UUID, invoices []model.SalesInvoice) ([]model.SalesInvoice, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	for _, inv := range invoices {
		inv.ID = uuid.New()
		inv.TenantID = tenantID
		inv.CreatedAt = time.Now()

		_, err := tx.Exec(ctx, `
			INSERT INTO sales_invoices (id, invoice_no, invoice_type, customer_id, tax_id, company_id,
				tenant_id, posting_date, due_date, total_amount, tax_amount, net_amount, outstanding_amount,
				status, tax_template_id, return_against, is_return, docstatus, created_by, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)`,
			inv.ID, inv.InvoiceNo, inv.InvoiceType, inv.CustomerID, inv.TaxID, inv.CompanyID,
			inv.TenantID, inv.PostingDate, inv.DueDate, inv.TotalAmount, inv.TaxAmount, inv.NetAmount,
			inv.OutstandingAmount, inv.Status, inv.TaxTemplateID, inv.ReturnAgainst, inv.IsReturn,
			inv.DocStatus, inv.CreatedBy, inv.CreatedAt)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return invoices, nil
}

// GetLineItems retrieves line items for an invoice.
func (r *InvoiceRepository) GetLineItems(ctx context.Context, tenantID, invoiceID uuid.UUID) ([]model.InvoiceLineItem, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, invoice_id, item_code, description, quantity, unit_price, tax_rate, tax_amount, net_amount, total_amount, unit, created_at
		FROM invoice_line_items WHERE invoice_id = $1`,
		invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.InvoiceLineItem
	for rows.Next() {
		var item model.InvoiceLineItem
		if err := rows.Scan(&item.ID, &item.InvoiceID, &item.ItemCode, &item.Description,
			&item.Quantity, &item.UnitPrice, &item.TaxRate, &item.TaxAmount, &item.NetAmount,
			&item.TotalAmount, &item.Unit, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// CreateLineItem inserts a line item for an invoice.
func (r *InvoiceRepository) CreateLineItem(ctx context.Context, item *model.InvoiceLineItem) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO invoice_line_items (id, invoice_id, item_code, description, quantity, unit_price, tax_rate, tax_amount, net_amount, total_amount, unit, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		item.ID, item.InvoiceID, item.ItemCode, item.Description, item.Quantity, item.UnitPrice,
		item.TaxRate, item.TaxAmount, item.NetAmount, item.TotalAmount, item.Unit, item.CreatedAt)
	return err
}

// MatchToBankTxn creates a payment allocation linking invoice to bank transaction.
func (r *InvoiceRepository) MatchToBankTxn(ctx context.Context, tenantID, invoiceID, bankTxnID uuid.UUID, amount string) error {
	// Parse amount - in real implementation, this would be decimal
	_, err := r.pool.Exec(ctx, `
		INSERT INTO payment_allocations (id, payment_entry_id, invoice_id, invoice_type, allocated_amount, tenant_id, created_at)
		VALUES ($1, $2, $3, 'sale', $4, $5, $6)`,
		uuid.New(), bankTxnID, invoiceID, amount, tenantID, time.Now())
	return err
}

// GetByInvoiceNo retrieves an invoice by its invoice number.
func (r *InvoiceRepository) GetByInvoiceNo(ctx context.Context, tenantID uuid.UUID, invoiceNo string) (*model.SalesInvoice, error) {
	var inv model.SalesInvoice
	err := r.pool.QueryRow(ctx, `
		SELECT id, invoice_no, invoice_type, customer_id, tax_id, company_id,
			tenant_id, posting_date, due_date, total_amount, tax_amount, net_amount, outstanding_amount,
			status, tax_template_id, return_against, is_return, docstatus, created_by, created_at
		FROM sales_invoices WHERE tenant_id = $1 AND invoice_no = $2`,
		tenantID, invoiceNo).
		Scan(&inv.ID, &inv.InvoiceNo, &inv.InvoiceType, &inv.CustomerID, &inv.TaxID,
			&inv.CompanyID, &inv.TenantID, &inv.PostingDate, &inv.DueDate, &inv.TotalAmount,
			&inv.TaxAmount, &inv.NetAmount, &inv.OutstandingAmount, &inv.Status, &inv.TaxTemplateID,
			&inv.ReturnAgainst, &inv.IsReturn, &inv.DocStatus, &inv.CreatedBy, &inv.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

// ValidateDuplicateInvoiceNo checks if invoice_no already exists for this tenant.
func (r *InvoiceRepository) ValidateDuplicateInvoiceNo(ctx context.Context, tenantID uuid.UUID, invoiceNo string) (bool, error) {
	var count int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM sales_invoices WHERE tenant_id = $1 AND invoice_no = $2`,
		tenantID, invoiceNo).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// UpdateOutstandingAmount updates the outstanding amount for an invoice.
func (r *InvoiceRepository) UpdateOutstandingAmount(ctx context.Context, tenantID, id uuid.UUID, outstandingAmount string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE sales_invoices SET outstanding_amount = $3
		WHERE tenant_id = $1 AND id = $2`,
		tenantID, id, outstandingAmount)
	return err
}

// ListInvoicesForMatching retrieves invoices that can be matched to a bank transaction.
func (r *InvoiceRepository) ListInvoicesForMatching(ctx context.Context, tenantID uuid.UUID, customerID *uuid.UUID) ([]model.SalesInvoice, error) {
	query := `
		SELECT id, invoice_no, invoice_type, customer_id, tax_id, company_id,
			tenant_id, posting_date, due_date, total_amount, tax_amount, net_amount, outstanding_amount,
			status, tax_template_id, return_against, is_return, docstatus, created_by, created_at
		FROM sales_invoices WHERE tenant_id = $1 AND outstanding_amount > 0`
	args := []interface{}{tenantID}
	argIdx := 2

	if customerID != nil {
		query += fmt.Sprintf(" AND customer_id = $%d", argIdx)
		args = append(args, *customerID)
	}

	query += " ORDER BY posting_date ASC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invoices []model.SalesInvoice
	for rows.Next() {
		var inv model.SalesInvoice
		if err := rows.Scan(&inv.ID, &inv.InvoiceNo, &inv.InvoiceType, &inv.CustomerID, &inv.TaxID,
			&inv.CompanyID, &inv.TenantID, &inv.PostingDate, &inv.DueDate, &inv.TotalAmount,
			&inv.TaxAmount, &inv.NetAmount, &inv.OutstandingAmount, &inv.Status, &inv.TaxTemplateID,
			&inv.ReturnAgainst, &inv.IsReturn, &inv.DocStatus, &inv.CreatedBy, &inv.CreatedAt); err != nil {
			return nil, err
		}
		invoices = append(invoices, inv)
	}
	return invoices, rows.Err()
}