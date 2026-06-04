package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Payroll represents the payroll_records table.
type Payroll struct {
	ID              uuid.UUID      `json:"id" db:"id"`
	TenantID        uuid.UUID     `json:"tenant_id" db:"tenant_id"`
	CompanyID       uuid.UUID     `json:"company_id" db:"company_id"`
	PayrollNo       string        `json:"payroll_no" db:"payroll_no"`
	EmployeeName    string        `json:"employee_name" db:"employee_name"`
	DepartmentName  string        `json:"department_name" db:"department_name"`
	PeriodNo        int           `json:"period_no" db:"period_no"` // e.g. 202506
	GrossSalary     decimal.Decimal `json:"gross_salary" db:"gross_salary"`
	IndividualTax   decimal.Decimal `json:"individual_tax" db:"individual_tax"`
	SocialSecurity  decimal.Decimal `json:"social_security" db:"social_security"`
	HousingFund     decimal.Decimal `json:"housing_fund" db:"housing_fund"`
	OtherDeductions decimal.Decimal `json:"other_deductions" db:"other_deductions"`
	NetSalary       decimal.Decimal `json:"net_salary" db:"net_salary"`
	PaymentDate     time.Time      `json:"payment_date" db:"payment_date"`
	BankAccountNo   string        `json:"bank_account_no" db:"bank_account_no"`
	Status          string        `json:"status" db:"status"`
	DocStatus       int16         `json:"docstatus" db:"docstatus"`
	VoucherID       *uuid.UUID    `json:"voucher_id,omitempty" db:"voucher_id"`
	VoucherNo       *string       `json:"voucher_no,omitempty" db:"voucher_no"`
	Source          string        `json:"source" db:"source"` // manual/excel/import
	Remark          *string       `json:"remark,omitempty" db:"remark"`
	CreatedBy       *uuid.UUID    `json:"created_by,omitempty" db:"created_by"`
	CreatedAt       time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at" db:"updated_at"`
}