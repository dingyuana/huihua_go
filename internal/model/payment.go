package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// PaymentEntry represents the payment_entries table.
type PaymentEntry struct {
	ID               uuid.UUID        `json:"id" db:"id"`
	PaymentNo        string           `json:"payment_no" db:"payment_no"`
	PaymentType      string           `json:"payment_type" db:"payment_type"`
	PartyType        string           `json:"party_type" db:"party_type"`
	PartyID          uuid.UUID        `json:"party_id" db:"party_id"`
	CounterpartyName *string          `json:"counterparty_name,omitempty" db:"counterparty_name"`
	PaidFromID       *uuid.UUID       `json:"paid_from_id,omitempty" db:"paid_from_id"`
	PaidToID         *uuid.UUID       `json:"paid_to_id,omitempty" db:"paid_to_id"`
	PaidAmount       decimal.Decimal  `json:"paid_amount" db:"paid_amount"`
	ReceivedAmount   *decimal.Decimal `json:"received_amount,omitempty" db:"received_amount"`
	ReferenceNo      *string          `json:"reference_no,omitempty" db:"reference_no"`
	ReferenceDate    *time.Time       `json:"reference_date,omitempty" db:"reference_date"`
	PostingDate      time.Time        `json:"posting_date" db:"posting_date"`
	CompanyID        uuid.UUID        `json:"company_id" db:"company_id"`
	TenantID         uuid.UUID        `json:"tenant_id" db:"tenant_id"`
	BankAccountID    *uuid.UUID       `json:"bank_account_id,omitempty" db:"bank_account_id"`
	DocStatus        int16            `json:"docstatus" db:"docstatus"`
	VoucherID        *uuid.UUID       `json:"voucher_id,omitempty" db:"voucher_id"`
	VoucherNo        *string          `json:"voucher_no,omitempty" db:"voucher_no"`
	Description      *string          `json:"description,omitempty" db:"description"`
	PaymentMethod    *string          `json:"payment_method,omitempty" db:"payment_method"`
	CreatedBy        *uuid.UUID       `json:"created_by,omitempty" db:"created_by"`
	CreatedAt        time.Time        `json:"created_at" db:"created_at"`
}

// PaymentAllocation represents the payment_allocations table (linking payments to invoices).
type PaymentAllocation struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	PaymentEntryID  uuid.UUID       `json:"payment_entry_id" db:"payment_entry_id"`
	InvoiceID       uuid.UUID       `json:"invoice_id" db:"invoice_id"`
	InvoiceType     *string         `json:"invoice_type,omitempty" db:"invoice_type"`
	AllocatedAmount decimal.Decimal `json:"allocated_amount" db:"allocated_amount"`
	TenantID        uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
}
