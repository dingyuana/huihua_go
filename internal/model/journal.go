package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// JournalEntry represents the journal_entries table (voucher master).
type JournalEntry struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	VoucherNo    string     `json:"voucher_no" db:"voucher_no"`
	VoucherType  *string    `json:"voucher_type,omitempty" db:"voucher_type"`
	PostingDate  time.Time  `json:"posting_date" db:"posting_date"`
	CompanyID    uuid.UUID  `json:"company_id" db:"company_id"`
	TenantID     uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	Remark       *string    `json:"remark,omitempty" db:"remark"`
	DocStatus    int16      `json:"docstatus" db:"docstatus"`
	ReversedID   *uuid.UUID `json:"reversed_id,omitempty" db:"reversed_id"`
	ReversalID   *uuid.UUID `json:"reversal_id,omitempty" db:"reversal_id"`
	SubmittedBy  *uuid.UUID `json:"submitted_by,omitempty" db:"submitted_by"`
	SubmittedAt  *time.Time `json:"submitted_at,omitempty" db:"submitted_at"`
	CreatedBy    uuid.UUID  `json:"created_by" db:"created_by"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// JournalEntryLine represents the journal_entry_lines table.
type JournalEntryLine struct {
	ID             uuid.UUID       `json:"id" db:"id"`
	JournalEntryID uuid.UUID       `json:"journal_entry_id" db:"journal_entry_id"`
	AccountID      uuid.UUID       `json:"account_id" db:"account_id"`
	Debit          decimal.Decimal `json:"debit" db:"debit"`
	Credit         decimal.Decimal `json:"credit" db:"credit"`
	DebitCcy       decimal.Decimal `json:"debit_ccy" db:"debit_ccy"`
	CreditCcy      decimal.Decimal `json:"credit_ccy" db:"credit_ccy"`
	AccountCcy     *string         `json:"account_ccy,omitempty" db:"account_ccy"`
	ExchangeRate   decimal.Decimal `json:"exchange_rate" db:"exchange_rate"`
	PartyType      *string         `json:"party_type,omitempty" db:"party_type"`
	PartyID        *uuid.UUID      `json:"party_id,omitempty" db:"party_id"`
	CostCenterID   *uuid.UUID      `json:"cost_center_id,omitempty" db:"cost_center_id"`
	ProjectID      *uuid.UUID      `json:"project_id,omitempty" db:"project_id"`
	UserRemark     *string         `json:"user_remark,omitempty" db:"user_remark"`
	Reconciled     bool            `json:"reconciled" db:"reconciled"`
	TenantID       uuid.UUID       `json:"tenant_id" db:"tenant_id"`
}
