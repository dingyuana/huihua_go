package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// BankAccount represents the bank_accounts table.
type BankAccount struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	BankName          string     `json:"bank_name" db:"bank_name"`
	AccountNumber     string     `json:"account_number" db:"account_number"`
	ClearingAccountID *uuid.UUID `json:"clearing_account_id,omitempty" db:"clearing_account_id"`
	CompanyID         uuid.UUID  `json:"company_id" db:"company_id"`
	TenantID          uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	Currency          string     `json:"currency" db:"currency"`
	IBAN              *string    `json:"iban,omitempty" db:"iban"`
	SwiftCode         *string    `json:"swift_code,omitempty" db:"swift_code"`
	BankAccountType   *string    `json:"bank_account_type,omitempty" db:"bank_account_type"`
	IsActive          bool       `json:"is_active" db:"is_active"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
}

// BankTransaction represents the bank_transactions table (imported bank statements).
type BankTransaction struct {
	ID                     uuid.UUID       `json:"id" db:"id"`
	BankAccountID          uuid.UUID       `json:"bank_account_id" db:"bank_account_id"`
	TxnDate                time.Time       `json:"txn_date" db:"txn_date"`
	Description            *string         `json:"description,omitempty" db:"description"`
	Debit                  decimal.Decimal `json:"debit" db:"debit"`
	Credit                 decimal.Decimal `json:"credit" db:"credit"`
	Direction              *string         `json:"direction,omitempty" db:"direction"`
	ReferenceNo            *string         `json:"reference_no,omitempty" db:"reference_no"`
	CounterpartyName       *string         `json:"counterparty_name,omitempty" db:"counterparty_name"`
	Matched                bool            `json:"matched" db:"matched"`
	MatchedPaymentEntryID  *uuid.UUID      `json:"matched_payment_entry_id,omitempty" db:"matched_payment_entry_id"`
	MatchedGLEntryID       *uuid.UUID      `json:"matched_gl_entry_id,omitempty" db:"matched_gl_entry_id"`
	ImportedFrom           *string         `json:"imported_from,omitempty" db:"imported_from"`
	RawData                json.RawMessage `json:"raw_data,omitempty" db:"raw_data"`
	CompanyID              uuid.UUID       `json:"company_id" db:"company_id"`
	TenantID               uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	CreatedAt              time.Time       `json:"created_at" db:"created_at"`
}

// BankReconciliationDetail represents the bank_reconciliation_details table.
type BankReconciliationDetail struct {
	ID                uuid.UUID       `json:"id" db:"id"`
	BankTransactionID uuid.UUID       `json:"bank_transaction_id" db:"bank_transaction_id"`
	PaymentEntryID    *uuid.UUID      `json:"payment_entry_id,omitempty" db:"payment_entry_id"`
	GLEntryID         *uuid.UUID      `json:"gl_entry_id,omitempty" db:"gl_entry_id"`
	MatchScore        *decimal.Decimal `json:"match_score,omitempty" db:"match_score"`
	DifferenceAccountID *uuid.UUID    `json:"difference_account_id,omitempty" db:"difference_account_id"`
	ReconciledAt      *time.Time      `json:"reconciled_at,omitempty" db:"reconciled_at"`
	ReconciledBy      *uuid.UUID      `json:"reconciled_by,omitempty" db:"reconciled_by"`
	TenantID          uuid.UUID       `json:"tenant_id" db:"tenant_id"`
}

// BankReconciliationStatement represents the bank_reconciliation_statements table.
type BankReconciliationStatement struct {
	ID                   uuid.UUID       `json:"id" db:"id"`
	BankAccountID        uuid.UUID       `json:"bank_account_id" db:"bank_account_id"`
	StatementDate        time.Time       `json:"statement_date" db:"statement_date"`
	BankStatementBalance decimal.Decimal `json:"bank_statement_balance" db:"bank_statement_balance"`
	GLBalance            decimal.Decimal `json:"gl_balance" db:"gl_balance"`
	Difference           decimal.Decimal `json:"difference" db:"difference"`
	BankOnlyTotal        decimal.Decimal `json:"bank_only_total" db:"bank_only_total"`
	GLOnlyTotal          decimal.Decimal `json:"gl_only_total" db:"gl_only_total"`
	Status               string          `json:"status" db:"status"`
	Locked               bool            `json:"locked" db:"locked"`
	LockedBy             *uuid.UUID      `json:"locked_by,omitempty" db:"locked_by"`
	TenantID             uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	CreatedBy            *uuid.UUID      `json:"created_by,omitempty" db:"created_by"`
	CreatedAt            time.Time       `json:"created_at" db:"created_at"`
}
