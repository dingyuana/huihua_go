package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"huihua/finance/internal/model"
)

type ExpenseInvoiceRepository struct {
	pool *pgxpool.Pool
}

func NewExpenseInvoiceRepository(pool *pgxpool.Pool) *ExpenseInvoiceRepository {
	return &ExpenseInvoiceRepository{pool: pool}
}

func (r *ExpenseInvoiceRepository) Create(ctx context.Context, tenantID uuid.UUID, inv *model.ExpenseInvoice) error {
	if inv.ID == uuid.Nil {
		inv.ID = uuid.New()
	}
	if inv.CreatedAt.IsZero() {
		inv.CreatedAt = time.Now()
	}
	if inv.VerifyStatus == "" {
		inv.VerifyStatus = "unverified"
	}
	if inv.DeductionStatus == "" {
		inv.DeductionStatus = "undeducted"
	}
	if inv.Status == "" {
		inv.Status = "pending"
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO expense_invoices (
			id, tenant_id, company_id, invoice_no, invoice_code, invoice_date,
			invoice_kind, tax_amount, total_amount, vendor_id, vendor_name, tax_id,
			verify_status, verified_at, verify_result, deduction_status, deducted_at,
			source_file, ocr_data, status, docstatus, remark, created_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25)`,
		inv.ID, tenantID, inv.CompanyID, inv.InvoiceNo, inv.InvoiceCode, inv.InvoiceDate,
		inv.InvoiceKind, inv.TaxAmount.String(), inv.TotalAmount.String(), inv.VendorID, inv.VendorName, inv.TaxID,
		inv.VerifyStatus, inv.VerifiedAt, inv.VerifyResult, inv.DeductionStatus, inv.DeductedAt,
		inv.SourceFile, inv.OcrData, inv.Status, inv.DocStatus, inv.Remark, inv.CreatedBy, inv.CreatedAt, inv.UpdatedAt)
	return err
}

func (r *ExpenseInvoiceRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*model.ExpenseInvoice, error) {
	var inv model.ExpenseInvoice
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, company_id, invoice_no, invoice_code, invoice_date,
			invoice_kind, tax_amount, total_amount, vendor_id, vendor_name, tax_id,
			verify_status, verified_at, verify_result, deduction_status, deducted_at,
			source_file, ocr_data, status, docstatus, remark, created_by, created_at, updated_at
		FROM expense_invoices WHERE tenant_id = $1 AND id = $2`,
		tenantID, id).
		Scan(
			&inv.ID, &inv.TenantID, &inv.CompanyID, &inv.InvoiceNo, &inv.InvoiceCode, &inv.InvoiceDate,
			&inv.InvoiceKind, &inv.TaxAmount, &inv.TotalAmount, &inv.VendorID, &inv.VendorName, &inv.TaxID,
			&inv.VerifyStatus, &inv.VerifiedAt, &inv.VerifyResult, &inv.DeductionStatus, &inv.DeductedAt,
			&inv.SourceFile, &inv.OcrData, &inv.Status, &inv.DocStatus, &inv.Remark, &inv.CreatedBy, &inv.CreatedAt, &inv.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &inv, nil
}

func (r *ExpenseInvoiceRepository) GetByInvoiceNo(ctx context.Context, tenantID uuid.UUID, invoiceNo string) (*model.ExpenseInvoice, error) {
	var inv model.ExpenseInvoice
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, company_id, invoice_no, invoice_code, invoice_date,
			invoice_kind, tax_amount, total_amount, vendor_id, vendor_name, tax_id,
			verify_status, verified_at, verify_result, deduction_status, deducted_at,
			source_file, ocr_data, status, docstatus, remark, created_by, created_at, updated_at
		FROM expense_invoices WHERE tenant_id = $1 AND invoice_no = $2`,
		tenantID, invoiceNo).
		Scan(
			&inv.ID, &inv.TenantID, &inv.CompanyID, &inv.InvoiceNo, &inv.InvoiceCode, &inv.InvoiceDate,
			&inv.InvoiceKind, &inv.TaxAmount, &inv.TotalAmount, &inv.VendorID, &inv.VendorName, &inv.TaxID,
			&inv.VerifyStatus, &inv.VerifiedAt, &inv.VerifyResult, &inv.DeductionStatus, &inv.DeductedAt,
			&inv.SourceFile, &inv.OcrData, &inv.Status, &inv.DocStatus, &inv.Remark, &inv.CreatedBy, &inv.CreatedAt, &inv.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &inv, nil
}

func (r *ExpenseInvoiceRepository) List(ctx context.Context, tenantID uuid.UUID, filters model.ExpenseInvoiceFilter) ([]*model.ExpenseInvoice, error) {
	query := `
		SELECT id, tenant_id, company_id, invoice_no, invoice_code, invoice_date,
			invoice_kind, tax_amount, total_amount, vendor_id, vendor_name, tax_id,
			verify_status, verified_at, verify_result, deduction_status, deducted_at,
			source_file, ocr_data, status, docstatus, remark, created_by, created_at, updated_at
		FROM expense_invoices WHERE tenant_id = $1`
	args := []interface{}{tenantID}
	argIdx := 2

	if filters.VendorID != nil {
		query += fmt.Sprintf(" AND vendor_id = $%d", argIdx)
		args = append(args, *filters.VendorID)
		argIdx++
	}
	if filters.VerifyStatus != "" {
		query += fmt.Sprintf(" AND verify_status = $%d", argIdx)
		args = append(args, filters.VerifyStatus)
		argIdx++
	}
	if filters.FromDate != nil {
		query += fmt.Sprintf(" AND invoice_date >= $%d", argIdx)
		args = append(args, *filters.FromDate)
		argIdx++
	}
	if filters.ToDate != nil {
		query += fmt.Sprintf(" AND invoice_date <= $%d", argIdx)
		args = append(args, *filters.ToDate)
		argIdx++
	}

	query += " ORDER BY created_at DESC"

	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, filters.Limit)
		argIdx++
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

	var list []*model.ExpenseInvoice
	for rows.Next() {
		var inv model.ExpenseInvoice
		if err := rows.Scan(
			&inv.ID, &inv.TenantID, &inv.CompanyID, &inv.InvoiceNo, &inv.InvoiceCode, &inv.InvoiceDate,
			&inv.InvoiceKind, &inv.TaxAmount, &inv.TotalAmount, &inv.VendorID, &inv.VendorName, &inv.TaxID,
			&inv.VerifyStatus, &inv.VerifiedAt, &inv.VerifyResult, &inv.DeductionStatus, &inv.DeductedAt,
			&inv.SourceFile, &inv.OcrData, &inv.Status, &inv.DocStatus, &inv.Remark, &inv.CreatedBy, &inv.CreatedAt, &inv.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &inv)
	}
	return list, rows.Err()
}

func (r *ExpenseInvoiceRepository) UpdateFields(ctx context.Context, tenantID, id uuid.UUID, fields map[string]interface{}) error {
	var setClauses []string
	var args []interface{}
	argIdx := 1

	for key, val := range fields {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", key, argIdx))
		args = append(args, val)
		argIdx++
	}

	if len(setClauses) == 0 {
		return nil
	}

	args = append(args, tenantID, id)
	query := fmt.Sprintf("UPDATE expense_invoices SET %s WHERE tenant_id = $%d AND id = $%d",
		strings.Join(setClauses, ", "), argIdx, argIdx+1)

	_, err := r.pool.Exec(ctx, query, args...)
	return err
}

func (r *ExpenseInvoiceRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM expense_invoices WHERE tenant_id = $1 AND id = $2`, tenantID, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("expense_invoice not found")
	}
	return nil
}

func (r *ExpenseInvoiceRepository) FindUnverified(ctx context.Context, tenantID uuid.UUID) ([]*model.ExpenseInvoice, error) {
	query := `
		SELECT id, tenant_id, company_id, invoice_no, invoice_code, invoice_date,
			invoice_kind, tax_amount, total_amount, vendor_id, vendor_name, tax_id,
			verify_status, verified_at, verify_result, deduction_status, deducted_at,
			source_file, ocr_data, status, docstatus, remark, created_by, created_at, updated_at
		FROM expense_invoices WHERE tenant_id = $1 AND verify_status = 'unverified'
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*model.ExpenseInvoice
	for rows.Next() {
		var inv model.ExpenseInvoice
		if err := rows.Scan(
			&inv.ID, &inv.TenantID, &inv.CompanyID, &inv.InvoiceNo, &inv.InvoiceCode, &inv.InvoiceDate,
			&inv.InvoiceKind, &inv.TaxAmount, &inv.TotalAmount, &inv.VendorID, &inv.VendorName, &inv.TaxID,
			&inv.VerifyStatus, &inv.VerifiedAt, &inv.VerifyResult, &inv.DeductionStatus, &inv.DeductedAt,
			&inv.SourceFile, &inv.OcrData, &inv.Status, &inv.DocStatus, &inv.Remark, &inv.CreatedBy, &inv.CreatedAt, &inv.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &inv)
	}
	return list, rows.Err()
}

func (r *ExpenseInvoiceRepository) FindUndeducted(ctx context.Context, tenantID uuid.UUID) ([]*model.ExpenseInvoice, error) {
	query := `
		SELECT id, tenant_id, company_id, invoice_no, invoice_code, invoice_date,
			invoice_kind, tax_amount, total_amount, vendor_id, vendor_name, tax_id,
			verify_status, verified_at, verify_result, deduction_status, deducted_at,
			source_file, ocr_data, status, docstatus, remark, created_by, created_at, updated_at
		FROM expense_invoices WHERE tenant_id = $1 AND deduction_status = 'undeducted'
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*model.ExpenseInvoice
	for rows.Next() {
		var inv model.ExpenseInvoice
		if err := rows.Scan(
			&inv.ID, &inv.TenantID, &inv.CompanyID, &inv.InvoiceNo, &inv.InvoiceCode, &inv.InvoiceDate,
			&inv.InvoiceKind, &inv.TaxAmount, &inv.TotalAmount, &inv.VendorID, &inv.VendorName, &inv.TaxID,
			&inv.VerifyStatus, &inv.VerifiedAt, &inv.VerifyResult, &inv.DeductionStatus, &inv.DeductedAt,
			&inv.SourceFile, &inv.OcrData, &inv.Status, &inv.DocStatus, &inv.Remark, &inv.CreatedBy, &inv.CreatedAt, &inv.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &inv)
	}
	return list, rows.Err()
}