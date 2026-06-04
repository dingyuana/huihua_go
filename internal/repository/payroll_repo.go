package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"huihua/finance/internal/model"
)

// PayrollRepository provides data access for payroll_records.
type PayrollRepository struct {
	pool *pgxpool.Pool
}

// NewPayrollRepository creates a new PayrollRepository.
func NewPayrollRepository(pool *pgxpool.Pool) *PayrollRepository {
	return &PayrollRepository{pool: pool}
}

// Create inserts a new payroll record.
func (r *PayrollRepository) Create(ctx context.Context, tenantID uuid.UUID, payroll *model.Payroll) (*model.Payroll, error) {
	query := `
		INSERT INTO payroll_records (id, tenant_id, company_id, payroll_no, employee_name, department_name, period_no, gross_salary, individual_tax, social_security, housing_fund, other_deductions, net_salary, payment_date, bank_account_no, status, docstatus, voucher_id, voucher_no, source, remark, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)
		RETURNING created_at, updated_at`

	if payroll.ID == uuid.Nil {
		payroll.ID = uuid.New()
	}

	err := r.pool.QueryRow(ctx, query,
		payroll.ID, tenantID, payroll.CompanyID, payroll.PayrollNo,
		payroll.EmployeeName, payroll.DepartmentName, payroll.PeriodNo,
		payroll.GrossSalary, payroll.IndividualTax, payroll.SocialSecurity,
		payroll.HousingFund, payroll.OtherDeductions, payroll.NetSalary,
		payroll.PaymentDate, payroll.BankAccountNo, payroll.Status,
		payroll.DocStatus, payroll.VoucherID, payroll.VoucherNo,
		payroll.Source, payroll.Remark, payroll.CreatedBy,
	).Scan(&payroll.CreatedAt, &payroll.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create payroll: %w", err)
	}
	payroll.TenantID = tenantID
	return payroll, nil
}

// GetByID retrieves a payroll record by ID.
func (r *PayrollRepository) GetByID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*model.Payroll, error) {
	query := `
		SELECT id, tenant_id, company_id, payroll_no, employee_name, department_name, period_no, gross_salary, individual_tax, social_security, housing_fund, other_deductions, net_salary, payment_date, bank_account_no, status, docstatus, voucher_id, voucher_no, source, remark, created_by, created_at, updated_at
		FROM payroll_records
		WHERE id = $1 AND tenant_id = $2`

	p := &model.Payroll{}
	err := r.pool.QueryRow(ctx, query, id, tenantID).Scan(
		&p.ID, &p.TenantID, &p.CompanyID, &p.PayrollNo,
		&p.EmployeeName, &p.DepartmentName, &p.PeriodNo,
		&p.GrossSalary, &p.IndividualTax, &p.SocialSecurity,
		&p.HousingFund, &p.OtherDeductions, &p.NetSalary,
		&p.PaymentDate, &p.BankAccountNo, &p.Status,
		&p.DocStatus, &p.VoucherID, &p.VoucherNo,
		&p.Source, &p.Remark, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get payroll by id: %w", err)
	}
	return p, nil
}

// ListByTenant retrieves all payroll records for a tenant.
func (r *PayrollRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]model.Payroll, error) {
	query := `
		SELECT id, tenant_id, company_id, payroll_no, employee_name, department_name, period_no, gross_salary, individual_tax, social_security, housing_fund, other_deductions, net_salary, payment_date, bank_account_no, status, docstatus, voucher_id, voucher_no, source, remark, created_by, created_at, updated_at
		FROM payroll_records
		WHERE tenant_id = $1
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list payroll: %w", err)
	}
	defer rows.Close()

	var list []model.Payroll
	for rows.Next() {
		var p model.Payroll
		if err := rows.Scan(
			&p.ID, &p.TenantID, &p.CompanyID, &p.PayrollNo,
			&p.EmployeeName, &p.DepartmentName, &p.PeriodNo,
			&p.GrossSalary, &p.IndividualTax, &p.SocialSecurity,
			&p.HousingFund, &p.OtherDeductions, &p.NetSalary,
			&p.PaymentDate, &p.BankAccountNo, &p.Status,
			&p.DocStatus, &p.VoucherID, &p.VoucherNo,
			&p.Source, &p.Remark, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan payroll: %w", err)
		}
		list = append(list, p)
	}
	return list, rows.Err()
}

// ListByPeriod retrieves payroll records by period number.
func (r *PayrollRepository) ListByPeriod(ctx context.Context, tenantID uuid.UUID, periodNo int) ([]model.Payroll, error) {
	query := `
		SELECT id, tenant_id, company_id, payroll_no, employee_name, department_name, period_no, gross_salary, individual_tax, social_security, housing_fund, other_deductions, net_salary, payment_date, bank_account_no, status, docstatus, voucher_id, voucher_no, source, remark, created_by, created_at, updated_at
		FROM payroll_records
		WHERE tenant_id = $1 AND period_no = $2
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, tenantID, periodNo)
	if err != nil {
		return nil, fmt.Errorf("list payroll by period: %w", err)
	}
	defer rows.Close()

	var list []model.Payroll
	for rows.Next() {
		var p model.Payroll
		if err := rows.Scan(
			&p.ID, &p.TenantID, &p.CompanyID, &p.PayrollNo,
			&p.EmployeeName, &p.DepartmentName, &p.PeriodNo,
			&p.GrossSalary, &p.IndividualTax, &p.SocialSecurity,
			&p.HousingFund, &p.OtherDeductions, &p.NetSalary,
			&p.PaymentDate, &p.BankAccountNo, &p.Status,
			&p.DocStatus, &p.VoucherID, &p.VoucherNo,
			&p.Source, &p.Remark, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan payroll: %w", err)
		}
		list = append(list, p)
	}
	return list, rows.Err()
}

// UpdateStatus updates the docstatus of a payroll record.
func (r *PayrollRepository) UpdateStatus(ctx context.Context, tenantID uuid.UUID, id uuid.UUID, docStatus int16) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE payroll_records SET docstatus = $1, updated_at = NOW() WHERE id = $2 AND tenant_id = $3`,
		docStatus, id, tenantID,
	)
	if err != nil {
		return fmt.Errorf("update payroll status: %w", err)
	}
	return nil
}

// SetVoucherID updates voucher_id and voucher_no fields.
func (r *PayrollRepository) SetVoucherID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID, voucherID uuid.UUID, voucherNo string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE payroll_records SET voucher_id = $1, voucher_no = $2, updated_at = NOW() WHERE id = $3 AND tenant_id = $4`,
		voucherID, voucherNo, id, tenantID,
	)
	if err != nil {
		return fmt.Errorf("set payroll voucher: %w", err)
	}
	return nil
}

// GetNextPayrollNo generates the next payroll number in format PY-YYYYMM-XXXX.
func (r *PayrollRepository) GetNextPayrollNo(ctx context.Context, tenantID uuid.UUID) (string, error) {
	period := time.Now().Format("PY-200601-")
	var seq int

	query := `
		SELECT COALESCE(MAX(SUBSTRING(payroll_no FROM length($1) + 1 FOR 4)::int), 0)
		FROM payroll_records
		WHERE tenant_id = $2 AND payroll_no LIKE $1 || '%'`

	if err := r.pool.QueryRow(ctx, query, period[:12], tenantID).Scan(&seq); err != nil {
		return "", fmt.Errorf("get next payroll no: %w", err)
	}
	seq++
	return fmt.Sprintf("%s%04d", period[:16], seq), nil
}