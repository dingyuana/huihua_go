package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// GLEntry represents the gl_entries table (general ledger, written on journal entry submit).
type GLEntry struct {
	ID                 uuid.UUID       `json:"id" db:"id"`
	AccountID          uuid.UUID       `json:"account_id" db:"account_id"`
	PostingDate        time.Time       `json:"posting_date" db:"posting_date"`
	Debit              decimal.Decimal `json:"debit" db:"debit"`
	Credit             decimal.Decimal `json:"credit" db:"credit"`
	DebitCcy           decimal.Decimal `json:"debit_ccy" db:"debit_ccy"`
	CreditCcy          decimal.Decimal `json:"credit_ccy" db:"credit_ccy"`
	AccountCcy         *string         `json:"account_ccy,omitempty" db:"account_ccy"`
	VoucherType        *string         `json:"voucher_type,omitempty" db:"voucher_type"`
	VoucherID          *uuid.UUID      `json:"voucher_id,omitempty" db:"voucher_id"`
	AgainstVoucherType *string         `json:"against_voucher_type,omitempty" db:"against_voucher_type"`
	AgainstVoucherID   *uuid.UUID      `json:"against_voucher_id,omitempty" db:"against_voucher_id"`
	PartyType          *string         `json:"party_type,omitempty" db:"party_type"`
	PartyID            *uuid.UUID      `json:"party_id,omitempty" db:"party_id"`
	CostCenterID       *uuid.UUID      `json:"cost_center_id,omitempty" db:"cost_center_id"`
	ProjectID          *uuid.UUID      `json:"project_id,omitempty" db:"project_id"`
	CompanyID          uuid.UUID       `json:"company_id" db:"company_id"`
	TenantID           uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	IsCancelled        bool            `json:"is_cancelled" db:"is_cancelled"`
	CreatedAt          time.Time       `json:"created_at" db:"created_at"`
}
