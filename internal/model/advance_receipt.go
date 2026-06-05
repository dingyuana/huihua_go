package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type AdvanceReceiptStatus string

const (
	AdvanceReceiptStatusDraft             AdvanceReceiptStatus = "draft"
	AdvanceReceiptStatusConfirmed         AdvanceReceiptStatus = "confirmed"
	AdvanceReceiptStatusPartiallyAllocated AdvanceReceiptStatus = "partially_allocated"
	AdvanceReceiptStatusFullyAllocated    AdvanceReceiptStatus = "fully_allocated"
	AdvanceReceiptStatusReversed          AdvanceReceiptStatus = "reversed"
)

type AdvanceReceipt struct {
	ID               uuid.UUID       `db:"id"`
	TenantID         uuid.UUID       `db:"tenant_id"`
	CompanyID        uuid.UUID       `db:"company_id"`
	CustomerID       uuid.UUID       `db:"customer_id"`
	AdvanceNo        string          `db:"advance_no"`
	Amount           decimal.Decimal `db:"amount"`
	AllocatedAmount  decimal.Decimal `db:"allocated_amount"`
	OutstandingAmount decimal.Decimal `db:"outstanding_amount"`
	ReceivedDate     time.Time       `db:"received_date"`
	DueDate          *time.Time      `db:"due_date"`
	Status           string          `db:"status"`
	SourceType       string          `db:"source_type"`
	BankAccountID    *uuid.UUID      `db:"bank_account_id"`
	ReferenceNo      *string         `db:"reference_no"`
	Remark           *string         `db:"remark"`
	VoucherID        *uuid.UUID      `db:"voucher_id"`
	VoucherNo        *string         `db:"voucher_no"`
	CreatedBy        *uuid.UUID      `db:"created_by"`
	CreatedAt        time.Time       `db:"created_at"`
	ConfirmedBy      *uuid.UUID      `db:"confirmed_by"`
	ConfirmedAt      *time.Time      `db:"confirmed_at"`
	ReversedBy       *uuid.UUID      `db:"reversed_by"`
	ReversedAt       *time.Time      `db:"reversed_at"`
}
