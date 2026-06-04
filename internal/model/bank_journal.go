package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// BankJournalEntry represents bank_journal_entries table.
type BankJournalEntry struct {
	ID            uuid.UUID       `json:"id" db:"id"`
	BankAccountID uuid.UUID      `json:"bank_account_id" db:"bank_account_id"`
	TxnDate       time.Time      `json:"txn_date" db:"txn_date"`
	Description   *string        `json:"description,omitempty" db:"description"`
	Debit         decimal.Decimal `json:"debit" db:"debit"`
	Credit        decimal.Decimal `json:"credit" db:"credit"`
	VoucherID     *uuid.UUID     `json:"voucher_id,omitempty" db:"voucher_id"`
	VoucherNo     *string        `json:"voucher_no,omitempty" db:"voucher_no"`
	TenantID      uuid.UUID      `json:"tenant_id" db:"tenant_id"`
	CreatedAt     time.Time      `json:"created_at" db:"created_at"`
}