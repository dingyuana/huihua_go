package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"huihua/finance/internal/model"
)

// ReimbursementRepository provides data access for bus_reimbursements.
type ReimbursementRepository struct {
	pool *pgxpool.Pool
}

// NewReimbursementRepository creates a new ReimbursementRepository.
func NewReimbursementRepository(pool *pgxpool.Pool) *ReimbursementRepository {
	return &ReimbursementRepository{pool: pool}
}

// Create inserts a new reimbursement record.
func (r *ReimbursementRepository) Create(ctx context.Context, tenantID uuid.UUID, reimbursement *model.BusReimbursement) (*model.BusReimbursement, error) {
	query := `
		INSERT INTO bus_reimbursements (id, reimbursement_no, employee_name, department, expense_type, sub_expense_type, amount, posting_date, description, bank_account, docstatus, voucher_id, voucher_no, created_by, tenant_id, company_id, reject_reason, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		RETURNING created_at`

	if reimbursement.ID == uuid.Nil {
		reimbursement.ID = uuid.New()
	}

	err := r.pool.QueryRow(ctx, query,
		reimbursement.ID, reimbursement.ReimbursementNo, reimbursement.EmployeeName,
		reimbursement.Department, reimbursement.ExpenseType, reimbursement.SubExpenseType, reimbursement.Amount,
		reimbursement.PostingDate, reimbursement.Description, reimbursement.BankAccount,
		reimbursement.DocStatus, reimbursement.VoucherID, reimbursement.VoucherNo,
		reimbursement.CreatedBy, tenantID, reimbursement.CompanyID,
		reimbursement.RejectReason, reimbursement.UpdatedAt,
	).Scan(&reimbursement.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create reimbursement: %w", err)
	}
	reimbursement.TenantID = tenantID
	return reimbursement, nil
}

// GetByID retrieves a reimbursement by its ID.
func (r *ReimbursementRepository) GetByID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*model.BusReimbursement, error) {
	query := `
		SELECT id, reimbursement_no, employee_name, department, expense_type, sub_expense_type, amount, posting_date, description, bank_account, docstatus, voucher_id, voucher_no, created_by, tenant_id, company_id, reject_reason, updated_at, created_at
		FROM bus_reimbursements
		WHERE id = $1 AND tenant_id = $2`

	reim := &model.BusReimbursement{}
	err := r.pool.QueryRow(ctx, query, id, tenantID).Scan(
		&reim.ID, &reim.ReimbursementNo, &reim.EmployeeName, &reim.Department,
		&reim.ExpenseType, &reim.SubExpenseType, &reim.Amount, &reim.PostingDate, &reim.Description,
		&reim.BankAccount, &reim.DocStatus, &reim.VoucherID, &reim.VoucherNo,
		&reim.CreatedBy, &reim.TenantID, &reim.CompanyID,
		&reim.RejectReason, &reim.UpdatedAt, &reim.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get reimbursement by id: %w", err)
	}
	return reim, nil
}

// Update updates a reimbursement record.
func (r *ReimbursementRepository) Update(ctx context.Context, tenantID uuid.UUID, reimbursement *model.BusReimbursement) error {
	query := `
		UPDATE bus_reimbursements
		SET employee_name = $1, department = $2, expense_type = $3, amount = $4, posting_date = $5, description = $6, bank_account = $7
		WHERE id = $8 AND tenant_id = $9`

	_, err := r.pool.Exec(ctx, query,
		reimbursement.EmployeeName, reimbursement.Department, reimbursement.ExpenseType,
		reimbursement.Amount, reimbursement.PostingDate, reimbursement.Description,
		reimbursement.BankAccount, reimbursement.ID, tenantID,
	)
	if err != nil {
		return fmt.Errorf("update reimbursement: %w", err)
	}
	return nil
}

// Delete deletes a reimbursement record.
func (r *ReimbursementRepository) Delete(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM bus_reimbursements WHERE id = $1 AND tenant_id = $2`, id, tenantID)
	if err != nil {
		return fmt.Errorf("delete reimbursement: %w", err)
	}
	return nil
}

// ListByTenant retrieves reimbursements for a tenant with optional status filter.
func (r *ReimbursementRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, status *int16) ([]model.BusReimbursement, error) {
	query := `
		SELECT id, reimbursement_no, employee_name, department, expense_type, sub_expense_type, amount, posting_date, description, bank_account, docstatus, voucher_id, voucher_no, created_by, tenant_id, company_id, reject_reason, updated_at, created_at
		FROM bus_reimbursements
		WHERE tenant_id = $1`
	args := []interface{}{tenantID}

	if status != nil {
		query += " AND docstatus = $2"
		args = append(args, *status)
	}
	query += " ORDER BY created_at DESC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list reimbursements: %w", err)
	}
	defer rows.Close()

	var list []model.BusReimbursement
	for rows.Next() {
		var reim model.BusReimbursement
		if err := rows.Scan(
			&reim.ID, &reim.ReimbursementNo, &reim.EmployeeName, &reim.Department,
			&reim.ExpenseType, &reim.SubExpenseType, &reim.Amount, &reim.PostingDate, &reim.Description,
			&reim.BankAccount, &reim.DocStatus, &reim.VoucherID, &reim.VoucherNo,
			&reim.CreatedBy, &reim.TenantID, &reim.CompanyID,
			&reim.RejectReason, &reim.UpdatedAt, &reim.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan reimbursement: %w", err)
		}
		list = append(list, reim)
	}
	return list, rows.Err()
}

// GetNextReimbursementNo generates the next reimbursement number in format RQ-YYYYMMDD-XXXX.
func (r *ReimbursementRepository) GetNextReimbursementNo(ctx context.Context, tenantID uuid.UUID) (string, error) {
	prefix := time.Now().Format("RQ-20060102-")
	var seq int

	query := `
		SELECT COALESCE(MAX(SUBSTRING(reimbursement_no FROM length($1) + 1 FOR 4)::int), 0)
		FROM bus_reimbursements
		WHERE tenant_id = $2 AND reimbursement_no LIKE $1 || '%'`

	if err := r.pool.QueryRow(ctx, query, prefix[:11], tenantID).Scan(&seq); err != nil {
		return "", fmt.Errorf("get next reimbursement no: %w", err)
	}
	seq++

	return fmt.Sprintf("%s%04d", prefix[:18], seq), nil
}

// UpdateStatus updates the docstatus of a reimbursement.
func (r *ReimbursementRepository) UpdateStatus(ctx context.Context, tenantID uuid.UUID, id uuid.UUID, status int16) error {
	_, err := r.pool.Exec(ctx, `UPDATE bus_reimbursements SET docstatus = $1 WHERE id = $2 AND tenant_id = $3`, status, id, tenantID)
	if err != nil {
		return fmt.Errorf("update reimbursement status: %w", err)
	}
	return nil
}

// UpdateVoucher updates voucher_id and voucher_no fields.
func (r *ReimbursementRepository) UpdateVoucher(ctx context.Context, tenantID uuid.UUID, id uuid.UUID, voucherID uuid.UUID, voucherNo string) error {
	_, err := r.pool.Exec(ctx, `UPDATE bus_reimbursements SET voucher_id = $1, voucher_no = $2 WHERE id = $3 AND tenant_id = $4`, voucherID, voucherNo, id, tenantID)
	if err != nil {
		return fmt.Errorf("update reimbursement voucher: %w", err)
	}
	return nil
}

// UpdateFields updates specific fields of a reimbursement record.
func (r *ReimbursementRepository) UpdateFields(ctx context.Context, tenantID uuid.UUID, id uuid.UUID, fields map[string]interface{}) error {
	if len(fields) == 0 {
		return nil
	}

	setClauses := []string{}
	args := []interface{}{}
	argIdx := 1

	for col, val := range fields {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, argIdx))
		args = append(args, val)
		argIdx++
	}

	query := fmt.Sprintf("UPDATE bus_reimbursements SET %s WHERE id = $%d AND tenant_id = $%d",
		strings.Join(setClauses, ", "), argIdx, argIdx+1)
	args = append(args, id, tenantID)

	_, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update reimbursement fields: %w", err)
	}
	return nil
}