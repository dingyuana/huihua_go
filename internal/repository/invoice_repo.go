package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
		INSERT INTO sales_invoices (id, invoice_no, invoice_code, invoice_type, customer_id, tax_id, company_id,
			tenant_id, posting_date, due_date, total_amount, tax_amount, net_amount, outstanding_amount,
			status, tax_template_id, return_against, is_return, is_reversed, invoice_category, remark, source_red_invoice_no,
			docstatus, created_by, invoice_kind, electronic_url, red_letter_info_id, red_letter_reason,
			original_invoice_id, is_part_red, red_amount, tax_authority_code, confirm_status, confirm_date, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35)`,
		inv.ID, inv.InvoiceNo, inv.InvoiceCode, inv.InvoiceType, inv.CustomerID, inv.TaxID, inv.CompanyID,
		inv.TenantID, inv.PostingDate, inv.DueDate, inv.TotalAmount, inv.TaxAmount, inv.NetAmount,
		inv.OutstandingAmount, inv.Status, inv.TaxTemplateID, inv.ReturnAgainst,		inv.IsReturn, inv.IsReversed,
		inv.InvoiceCategory, inv.Remark, inv.SourceRedInvoiceNo,
		inv.DocStatus, inv.CreatedBy, inv.InvoiceKind, inv.ElectronicURL, inv.RedLetterInfoID, inv.RedLetterReason,
		inv.OriginalInvoiceID, inv.IsPartRed, inv.RedAmount, inv.TaxAuthorityCode, inv.ConfirmStatus, inv.ConfirmDate, inv.CreatedAt)
	if err != nil {
		return nil, err
	}
	return inv, nil
}

// ListByTenant retrieves invoices with filters.
func (r *InvoiceRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, filters model.InvoiceFilter) ([]model.SalesInvoice, error) {
	query := `
		SELECT s.id, s.invoice_no, s.invoice_code, s.invoice_type, s.customer_id, s.tax_id, s.company_id,
			s.tenant_id, s.posting_date, s.due_date, s.total_amount, s.tax_amount, s.net_amount, s.outstanding_amount,
			s.status, s.tax_template_id, s.return_against, s.is_return, s.is_reversed, s.invoice_category, s.remark, s.source_red_invoice_no,
			s.docstatus, s.created_by, s.invoice_kind, s.electronic_url, s.red_letter_info_id, s.red_letter_reason,
			s.original_invoice_id,
			COALESCE(s.is_part_red, FALSE) AS is_part_red,
			COALESCE(s.red_amount, 0) AS red_amount,
			s.tax_authority_code, COALESCE(s.confirm_status,'unconfirmed') AS confirm_status, COALESCE(s.confirm_date, CURRENT_DATE) AS confirm_date, s.created_at, COALESCE(p.name, '') AS customer_name
		FROM sales_invoices s
		LEFT JOIN parties p ON p.id = s.customer_id
		WHERE s.tenant_id = $1`
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
	if filters.Type != "" {
		query += fmt.Sprintf(" AND invoice_type = $%d", argIdx)
		args = append(args, filters.Type)
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
		if err := rows.Scan(&inv.ID, &inv.InvoiceNo, &inv.InvoiceCode, &inv.InvoiceType, &inv.CustomerID, &inv.TaxID,
			&inv.CompanyID, &inv.TenantID, &inv.PostingDate, &inv.DueDate, &inv.TotalAmount,
			&inv.TaxAmount, &inv.NetAmount, &inv.OutstandingAmount, &inv.Status, &inv.TaxTemplateID,
			&inv.ReturnAgainst, &inv.IsReturn, &inv.IsReversed, &inv.InvoiceCategory, &inv.Remark, &inv.SourceRedInvoiceNo,
			&inv.DocStatus, &inv.CreatedBy, &inv.InvoiceKind, &inv.ElectronicURL, &inv.RedLetterInfoID, &inv.RedLetterReason,
			&inv.OriginalInvoiceID, &inv.IsPartRed, &inv.RedAmount, &inv.TaxAuthorityCode, &inv.ConfirmStatus, &inv.ConfirmDate,
			&inv.CreatedAt, &inv.CustomerName); err != nil {
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
		SELECT id, invoice_no, invoice_code, invoice_type, customer_id, tax_id, company_id,
			tenant_id, posting_date, due_date, total_amount, tax_amount, net_amount, outstanding_amount,
			status, tax_template_id, return_against, is_return, is_reversed, invoice_category, remark, source_red_invoice_no,
			docstatus, created_by, invoice_kind, electronic_url, red_letter_info_id, red_letter_reason,
			original_invoice_id,
			COALESCE(is_part_red, FALSE) AS is_part_red,
			COALESCE(red_amount, 0) AS red_amount,
			tax_authority_code, COALESCE(confirm_status, 'unconfirmed') AS confirm_status, COALESCE(confirm_date, created_at) AS confirm_date, created_at
		FROM sales_invoices WHERE tenant_id = $1 AND id = $2`,
		tenantID, id).
		Scan(&inv.ID, &inv.InvoiceNo, &inv.InvoiceCode, &inv.InvoiceType, &inv.CustomerID, &inv.TaxID,
			&inv.CompanyID, &inv.TenantID, &inv.PostingDate, &inv.DueDate, &inv.TotalAmount,
			&inv.TaxAmount, &inv.NetAmount, &inv.OutstandingAmount, &inv.Status, &inv.TaxTemplateID,
			&inv.ReturnAgainst, &inv.IsReturn, &inv.IsReversed, &inv.InvoiceCategory, &inv.Remark, &inv.SourceRedInvoiceNo,
			&inv.DocStatus, &inv.CreatedBy, &inv.InvoiceKind, &inv.ElectronicURL, &inv.RedLetterInfoID, &inv.RedLetterReason,
			&inv.OriginalInvoiceID, &inv.IsPartRed, &inv.RedAmount, &inv.TaxAuthorityCode, &inv.ConfirmStatus, &inv.ConfirmDate,
			&inv.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

// GetByInvoiceNo retrieves an invoice by its number within the tenant.
// Used to resolve "red invoice" -> original blue invoice linkage on import.
func (r *InvoiceRepository) GetByInvoiceNo(ctx context.Context, tenantID uuid.UUID, invoiceNo string) (*model.SalesInvoice, error) {
	var inv model.SalesInvoice
	err := r.pool.QueryRow(ctx, `
		SELECT id, invoice_no, invoice_code, invoice_type, customer_id, tax_id, company_id,
			tenant_id, posting_date, due_date, total_amount, tax_amount, net_amount, outstanding_amount,
			status, tax_template_id, return_against, is_return, is_reversed, invoice_category, remark, source_red_invoice_no,
			docstatus, created_by, invoice_kind, electronic_url, red_letter_info_id, red_letter_reason,
			original_invoice_id,
			COALESCE(is_part_red, FALSE) AS is_part_red,
			COALESCE(red_amount, 0) AS red_amount,
			tax_authority_code, COALESCE(confirm_status, 'unconfirmed') AS confirm_status, COALESCE(confirm_date, created_at) AS confirm_date, created_at
		FROM sales_invoices WHERE tenant_id = $1 AND invoice_no = $2`,
		tenantID, invoiceNo).
		Scan(&inv.ID, &inv.InvoiceNo, &inv.InvoiceCode, &inv.InvoiceType, &inv.CustomerID, &inv.TaxID,
			&inv.CompanyID, &inv.TenantID, &inv.PostingDate, &inv.DueDate, &inv.TotalAmount,
			&inv.TaxAmount, &inv.NetAmount, &inv.OutstandingAmount, &inv.Status, &inv.TaxTemplateID,
			&inv.ReturnAgainst, &inv.IsReturn, &inv.IsReversed, &inv.InvoiceCategory, &inv.Remark, &inv.SourceRedInvoiceNo,
			&inv.DocStatus, &inv.CreatedBy, &inv.InvoiceKind, &inv.ElectronicURL, &inv.RedLetterInfoID, &inv.RedLetterReason,
			&inv.OriginalInvoiceID, &inv.IsPartRed, &inv.RedAmount, &inv.TaxAuthorityCode, &inv.ConfirmStatus, &inv.ConfirmDate,
			&inv.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

// GetRedInvoices returns all red-letter (is_return=true) invoices for a tenant.
func (r *InvoiceRepository) GetRedInvoices(ctx context.Context, tenantID uuid.UUID) ([]*model.SalesInvoice, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, invoice_no, invoice_code, invoice_type, customer_id, tax_id, company_id,
			tenant_id, posting_date, due_date, total_amount, tax_amount, net_amount, outstanding_amount,
			status, tax_template_id, return_against, is_return, is_reversed, invoice_category, remark, source_red_invoice_no,
			docstatus, created_by, invoice_kind, electronic_url, red_letter_info_id, red_letter_reason,
			original_invoice_id,
			COALESCE(is_part_red, FALSE) AS is_part_red,
			COALESCE(red_amount, 0) AS red_amount,
			tax_authority_code, COALESCE(confirm_status, 'unconfirmed') AS confirm_status, COALESCE(confirm_date, created_at) AS confirm_date, created_at
		FROM sales_invoices WHERE tenant_id = $1 AND is_return = true
		ORDER BY posting_date DESC`,
		tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*model.SalesInvoice
	for rows.Next() {
		var inv model.SalesInvoice
		if err := rows.Scan(&inv.ID, &inv.InvoiceNo, &inv.InvoiceCode, &inv.InvoiceType, &inv.CustomerID, &inv.TaxID,
			&inv.CompanyID, &inv.TenantID, &inv.PostingDate, &inv.DueDate, &inv.TotalAmount,
			&inv.TaxAmount, &inv.NetAmount, &inv.OutstandingAmount, &inv.Status, &inv.TaxTemplateID,
			&inv.ReturnAgainst, &inv.IsReturn, &inv.IsReversed, &inv.InvoiceCategory, &inv.Remark, &inv.SourceRedInvoiceNo,
			&inv.DocStatus, &inv.CreatedBy, &inv.InvoiceKind, &inv.ElectronicURL, &inv.RedLetterInfoID, &inv.RedLetterReason,
			&inv.OriginalInvoiceID, &inv.IsPartRed, &inv.RedAmount, &inv.TaxAuthorityCode, &inv.ConfirmStatus, &inv.ConfirmDate,
			&inv.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, &inv)
	}
	return result, rows.Err()
}

// UpdateStatusTx updates the invoice status within a transaction.
func (r *InvoiceRepository) UpdateStatusTx(ctx context.Context, tx pgx.Tx, tenantID, id uuid.UUID, status string) error {
	_, err := tx.Exec(ctx, `
		UPDATE sales_invoices SET status = $3
		WHERE tenant_id = $1 AND id = $2`,
		tenantID, id, status)
	return err
}

// UpdateOutstandingAmountTx updates outstanding amount within a transaction.
func (r *InvoiceRepository) UpdateOutstandingAmountTx(ctx context.Context, tx pgx.Tx, tenantID, id uuid.UUID, amount string) error {
	_, err := tx.Exec(ctx, `
		UPDATE sales_invoices SET outstanding_amount = $3
		WHERE tenant_id = $1 AND id = $2`,
		tenantID, id, amount)
	return err
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

func (r *InvoiceRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `
		DELETE FROM sales_invoices WHERE tenant_id = $1 AND id = $2`,
		tenantID, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("invoice not found")
	}
	return nil
}

// UpdateFields partially updates an invoice. Only non-nil pointers are written.
func (r *InvoiceRepository) UpdateFields(ctx context.Context, tenantID, id uuid.UUID, fields map[string]interface{}) error {
	if len(fields) == 0 {
		return nil
	}
	setClauses := make([]string, 0, len(fields))
	args := []interface{}{tenantID, id}
	idx := 3
	for col, val := range fields {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, idx))
		args = append(args, val)
		idx++
	}
	query := fmt.Sprintf("UPDATE sales_invoices SET %s WHERE tenant_id = $1 AND id = $2", strings.Join(setClauses, ", "))
	result, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("invoice not found")
	}
	return nil
}

// ImportBatch inserts multiple invoices in a transaction, skipping rows whose
// invoice_no already exists. Returns the rows actually inserted (i.e. without
// the duplicates the DB silently dropped).
func (r *InvoiceRepository) ImportBatch(ctx context.Context, tenantID uuid.UUID, invoices []model.SalesInvoice) ([]model.SalesInvoice, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	inserted := make([]model.SalesInvoice, 0, len(invoices))
	for _, inv := range invoices {
		inv.ID = uuid.New()
		inv.TenantID = tenantID
		inv.CreatedAt = time.Now()

		ctid, err := tx.Exec(ctx, `
			INSERT INTO sales_invoices (id, invoice_no, invoice_code, invoice_type, customer_id, tax_id, company_id,
				tenant_id, posting_date, due_date, total_amount, tax_amount, net_amount, outstanding_amount,
				status, tax_template_id, return_against, is_return, is_reversed, invoice_category, remark, source_red_invoice_no,
				docstatus, created_by, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24)
			ON CONFLICT (invoice_no) DO NOTHING`,
			inv.ID, inv.InvoiceNo, inv.InvoiceCode, inv.InvoiceType, inv.CustomerID, inv.TaxID, inv.CompanyID,
			inv.TenantID, inv.PostingDate, inv.DueDate, inv.TotalAmount, inv.TaxAmount, inv.NetAmount, inv.OutstandingAmount, inv.Status, inv.TaxTemplateID, inv.ReturnAgainst, inv.IsReturn, inv.IsReversed,
			inv.InvoiceCategory, inv.Remark, inv.SourceRedInvoiceNo,
			inv.DocStatus, inv.CreatedBy, inv.CreatedAt)
		if err != nil {
			return nil, err
		}
		if ctid.RowsAffected() == 0 {
			continue
		}
		inserted = append(inserted, inv)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return inserted, nil
}

// ValidateDuplicateBatch checks which invoice_nos from the list already exist in the DB.
// Returns the list of invoice_nos that conflict (already exist).
func (r *InvoiceRepository) ValidateDuplicateBatch(ctx context.Context, tenantID uuid.UUID, invoiceNos []string) ([]string, error) {
	if len(invoiceNos) == 0 {
		return nil, nil
	}
	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT invoice_no FROM sales_invoices
		WHERE tenant_id = $1 AND invoice_no = ANY($2)`,
		tenantID, invoiceNos)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conflicts []string
	for rows.Next() {
		var invNo string
		if err := rows.Scan(&invNo); err != nil {
			return nil, err
		}
		conflicts = append(conflicts, invNo)
	}
	return conflicts, rows.Err()
}

// ImportBatchWithItems inserts invoice headers and their line items in a single transaction.
// Unlike ImportBatch, this does NOT use ON CONFLICT DO NOTHING — duplicates should be
// pre-validated via ValidateDuplicateBatch before calling this.
// items[i] corresponds to invoices[i]; both slices must have the same length.
func (r *InvoiceRepository) ImportBatchWithItems(ctx context.Context, tenantID uuid.UUID, invoices []model.SalesInvoice, items [][]model.InvoiceLineItem) ([]model.SalesInvoice, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	inserted := make([]model.SalesInvoice, 0, len(invoices))
	for i, inv := range invoices {
		inv.ID = uuid.New()
		inv.TenantID = tenantID
		inv.CreatedAt = time.Now()

		_, err := tx.Exec(ctx, `
			INSERT INTO sales_invoices (id, invoice_no, invoice_code, invoice_type, customer_id, tax_id, company_id,
				tenant_id, posting_date, due_date, total_amount, tax_amount, net_amount, outstanding_amount,
				status, tax_template_id, return_against, is_return, is_reversed, invoice_category, remark, source_red_invoice_no,
				docstatus, created_by, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25)`,
			inv.ID, inv.InvoiceNo, inv.InvoiceCode, inv.InvoiceType, inv.CustomerID, inv.TaxID, inv.CompanyID,
			inv.TenantID, inv.PostingDate, inv.DueDate, inv.TotalAmount, inv.TaxAmount, inv.NetAmount, inv.OutstandingAmount, inv.Status, inv.TaxTemplateID, inv.ReturnAgainst, inv.IsReturn, inv.IsReversed,
			inv.InvoiceCategory, inv.Remark, inv.SourceRedInvoiceNo,
			inv.DocStatus, inv.CreatedBy, inv.CreatedAt)
		if err != nil {
			return nil, err
		}

		if i < len(items) {
			for _, item := range items[i] {
				item.ID = uuid.New()
				item.InvoiceID = inv.ID
				item.CreatedAt = time.Now()
				if _, err := tx.Exec(ctx, `
					INSERT INTO invoice_line_items (id, invoice_id, item_code, description, quantity, unit_price, tax_rate, tax_amount, net_amount, total_amount, unit, created_at)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
					item.ID, item.InvoiceID, item.ItemCode, item.Description, item.Quantity, item.UnitPrice,
					item.TaxRate, item.TaxAmount, item.NetAmount, item.TotalAmount, item.Unit, item.CreatedAt); err != nil {
					return nil, err
				}
			}
		}

		inserted = append(inserted, inv)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return inserted, nil
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
		SELECT id, invoice_no, invoice_code, invoice_type, customer_id, tax_id, company_id,
			tenant_id, posting_date, due_date, total_amount, tax_amount, net_amount, outstanding_amount,
			status, tax_template_id, return_against, is_return, is_reversed, invoice_category, remark, source_red_invoice_no,
			docstatus, created_by, invoice_kind, electronic_url, red_letter_info_id, red_letter_reason,
			original_invoice_id,
			COALESCE(is_part_red, FALSE) AS is_part_red,
			COALESCE(red_amount, 0) AS red_amount,
			tax_authority_code, COALESCE(confirm_status, 'unconfirmed') AS confirm_status, COALESCE(confirm_date, created_at) AS confirm_date, created_at
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
		if err := rows.Scan(&inv.ID, &inv.InvoiceNo, &inv.InvoiceCode, &inv.InvoiceType, &inv.CustomerID, &inv.TaxID,
			&inv.CompanyID, &inv.TenantID, &inv.PostingDate, &inv.DueDate, &inv.TotalAmount,
			&inv.TaxAmount, &inv.NetAmount, &inv.OutstandingAmount, &inv.Status, &inv.TaxTemplateID,
			&inv.ReturnAgainst, &inv.IsReturn, &inv.IsReversed, &inv.InvoiceCategory, &inv.Remark, &inv.SourceRedInvoiceNo,
			&inv.DocStatus, &inv.CreatedBy, &inv.InvoiceKind, &inv.ElectronicURL, &inv.RedLetterInfoID, &inv.RedLetterReason,
			&inv.OriginalInvoiceID, &inv.IsPartRed, &inv.RedAmount, &inv.TaxAuthorityCode, &inv.ConfirmStatus, &inv.ConfirmDate,
			&inv.CreatedAt); err != nil {
			return nil, err
		}
		invoices = append(invoices, inv)
	}
	return invoices, rows.Err()
}

// CreateAllocation creates a payment allocation record linking a payment entry to an invoice.
func (r *InvoiceRepository) CreateAllocation(ctx context.Context, alloc *model.PaymentAllocation) error {
	alloc.ID = uuid.New()
	alloc.CreatedAt = time.Now()
	_, err := r.pool.Exec(ctx, `
		INSERT INTO payment_allocations (id, payment_entry_id, invoice_id, invoice_type, allocated_amount, tenant_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		alloc.ID, alloc.PaymentEntryID, alloc.InvoiceID, alloc.InvoiceType,
		alloc.AllocatedAmount.String(), alloc.TenantID, alloc.CreatedAt)
	return err
}

// GetAllocationsByPaymentEntry retrieves all allocations for a given payment entry.
func (r *InvoiceRepository) GetAllocationsByPaymentEntry(ctx context.Context, tenantID, paymentEntryID uuid.UUID) ([]model.PaymentAllocation, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, payment_entry_id, invoice_id, invoice_type, allocated_amount, tenant_id, created_at
		FROM payment_allocations WHERE tenant_id = $1 AND payment_entry_id = $2`,
		tenantID, paymentEntryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allocs []model.PaymentAllocation
	for rows.Next() {
		var a model.PaymentAllocation
		if err := rows.Scan(&a.ID, &a.PaymentEntryID, &a.InvoiceID, &a.InvoiceType,
			&a.AllocatedAmount, &a.TenantID, &a.CreatedAt); err != nil {
			return nil, err
		}
		allocs = append(allocs, a)
	}
	return allocs, rows.Err()
}

// GetDefaultCompanyID returns the first company_id for the tenant.
// Used by file-based import to assign sales invoices to the current company.
func (r *InvoiceRepository) GetDefaultCompanyID(ctx context.Context, tenantID uuid.UUID) (uuid.UUID, error) {
	var companyID uuid.UUID
	err := r.pool.QueryRow(ctx, `
		SELECT id FROM company_settings
		WHERE tenant_id = $1
		ORDER BY created_at ASC
		LIMIT 1`,
		tenantID).Scan(&companyID)
	if err != nil {
		return uuid.Nil, err
	}
	return companyID, nil
}

// GetSummary returns aggregate stats for a tenant's invoices.
func (r *InvoiceRepository) GetSummary(ctx context.Context, tenantID uuid.UUID, filters model.InvoiceFilter) (*model.InvoiceSummary, error) {
	whereClauses := []string{"tenant_id = $1"}
	args := []interface{}{tenantID}
	argIdx := 2

	if filters.Type != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("invoice_type = $%d", argIdx))
		args = append(args, filters.Type)
		argIdx++
	}
	if filters.FromDate != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("posting_date >= $%d", argIdx))
		args = append(args, *filters.FromDate)
		argIdx++
	}
	if filters.ToDate != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("posting_date <= $%d", argIdx))
		args = append(args, *filters.ToDate)
		argIdx++
	}

	query := fmt.Sprintf(`
		SELECT
			COALESCE(SUM(CASE WHEN status = 'draft' THEN 1 ELSE 0 END), 0) AS draft_count,
			COALESCE(SUM(CASE WHEN status = 'submitted' THEN 1 ELSE 0 END), 0) AS submitted_count,
			COALESCE(SUM(CASE WHEN status = 'verified' THEN 1 ELSE 0 END), 0) AS verified_count,
			COALESCE(SUM(CASE WHEN status = 'partially_paid' THEN 1 ELSE 0 END), 0) AS partially_paid_count,
			COALESCE(SUM(CASE WHEN status = 'paid' THEN 1 ELSE 0 END), 0) AS paid_count,
			COALESCE(SUM(CASE WHEN status = 'reversed' THEN 1 ELSE 0 END), 0) AS reversed_count,
			COALESCE(SUM(total_amount), 0) AS total_amount,
			COALESCE(SUM(tax_amount), 0) AS tax_amount,
			COALESCE(SUM(net_amount), 0) AS net_amount,
			COALESCE(SUM(outstanding_amount), 0) AS outstanding_amount
		FROM sales_invoices
		WHERE %s`, strings.Join(whereClauses, " AND "))

	var s model.InvoiceSummary
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&s.DraftCount, &s.SubmittedCount, &s.VerifiedCount,
		&s.PartiallyPaidCount, &s.PaidCount, &s.ReversedCount,
		&s.TotalAmount, &s.TaxAmount, &s.NetAmount, &s.OutstandingAmount,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}
