package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const (
	ExpenseTypeTravel     = "travel"     // 差旅费
	ExpenseTypeEntertain = "entertain"   // 招待费
	ExpenseTypeOffice    = "office"     // 办公费
	ExpenseTypeTransport = "transport"  // 交通费
	ExpenseTypeOther     = "other"      // 其他
)

// BusReimbursement represents the bus_reimbursements table.
type BusReimbursement struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	ReimbursementNo string         `json:"reimbursement_no" db:"reimbursement_no"`
	EmployeeName    string         `json:"employee_name" db:"employee_name"`
	Department      *string        `json:"department,omitempty" db:"department"`
	ExpenseType     string         `json:"expense_type" db:"expense_type"`
	Amount          decimal.Decimal `json:"amount" db:"amount"`
	PostingDate     time.Time      `json:"posting_date" db:"posting_date"`
	Description     *string        `json:"description,omitempty" db:"description"`
	BankAccount     *string        `json:"bank_account,omitempty" db:"bank_account"`
	DocStatus       int16          `json:"docstatus" db:"docstatus"`
	VoucherID       *uuid.UUID     `json:"voucher_id,omitempty" db:"voucher_id"`
	VoucherNo       *string        `json:"voucher_no,omitempty" db:"voucher_no"`
	CreatedBy       *uuid.UUID     `json:"created_by,omitempty" db:"created_by"`
	TenantID        uuid.UUID      `json:"tenant_id" db:"tenant_id"`
	CompanyID       uuid.UUID      `json:"company_id" db:"company_id"`
	CreatedAt       time.Time      `json:"created_at" db:"created_at"`
}