package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// BankTransactionStatus represents the status of a bank transaction.
type BankTransactionStatus string

const (
	BankTxnStatusPending    BankTransactionStatus = "pending"
	BankTxnStatusMatched    BankTransactionStatus = "matched"
	BankTxnStatusReconciled BankTransactionStatus = "reconciled"
	BankTxnStatusExcluded   BankTransactionStatus = "excluded"
)

// BankTransactionImport represents a single row from Excel import.
type BankTransactionImport struct {
	RowNumber       int      `json:"row_number"`       // Excel row number (1-indexed, header=1)
	TransactionDate string   `json:"transaction_date"` // Date string from Excel
	VoucherNo       *string  `json:"voucher_no,omitempty"`
	Description     string   `json:"description"`
	Income          *float64 `json:"income,omitempty"`  // Positive if income
	Expense         *float64 `json:"expense,omitempty"` // Positive if expense
	Balance         *float64 `json:"balance,omitempty"`
	Counterparty    *string  `json:"counterparty,omitempty"`
}

// BankTxnFilter represents query filters for bank transaction listing.
type BankTxnFilter struct {
	StartDate      *time.Time             `json:"start_date,omitempty"`
	EndDate        *time.Time             `json:"end_date,omitempty"`
	BankAccountID  *uuid.UUID             `json:"bank_account_id,omitempty"`
	MinAmount      *decimal.Decimal       `json:"min_amount,omitempty"`
	MaxAmount      *decimal.Decimal       `json:"max_amount,omitempty"`
	Status         *BankTransactionStatus `json:"status,omitempty"`
	Classification *string                `json:"classification,omitempty"` // filter by business type
	Search         *string                `json:"search,omitempty"`         // description/counterparty search
	Page           int                    `json:"page,omitempty"`           // default 1
	PageSize       int                    `json:"page_size,omitempty"`      // default 50
}

// ImportResult represents the result of an Excel import operation.
type ImportResult struct {
	TotalRows     int               `json:"total_rows"`
	SuccessCount  int               `json:"success_count"`
	FailedCount   int               `json:"failed_count"`
	FailedRows    []int             `json:"failed_rows,omitempty"`    // row numbers that failed
	FailedReasons []FailedRowDetail `json:"failed_reasons,omitempty"` // per-row failure details
	ErrorMsg      string            `json:"error_message,omitempty"`
	ImportedTxns  []BankTransaction `json:"imported_txns,omitempty"` // 本次成功导入的流水（仅含当前批次，用于按需触发自动生成凭证）
}

// FailedRowDetail stores info about a single failed row during import.
type FailedRowDetail struct {
	Row    int    `json:"row"`
	Date   string `json:"date,omitempty"`
	Amount string `json:"amount,omitempty"`
	Desc   string `json:"desc,omitempty"`
	Reason string `json:"reason"`
}

type BankBalanceAdjustment struct {
	ID             uuid.UUID       `json:"id" db:"id"`
	TenantID       uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	BankAccountID  uuid.UUID       `json:"bank_account_id" db:"bank_account_id"`
	AdjustmentType string          `json:"adjustment_type" db:"adjustment_type"`
	BeforeBalance  decimal.Decimal `json:"before_balance" db:"before_balance"`
	AfterBalance   decimal.Decimal `json:"after_balance" db:"after_balance"`
	Delta          decimal.Decimal `json:"delta" db:"delta"`
	Reason         *string         `json:"reason,omitempty" db:"reason"`
	OperatorID     *uuid.UUID      `json:"operator_id,omitempty" db:"operator_id"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
}

// MatchResult represents the result of marking transactions as matched.
type MatchResult struct {
	MatchedIDs []uuid.UUID `json:"matched_ids"`
	Count      int         `json:"count"`
}

// BankAccount represents the bank_accounts table.
type BankAccount struct {
	ID                uuid.UUID       `json:"id" db:"id"`
	BankName          string          `json:"bank_name" db:"bank_name"`
	AccountNumber     string          `json:"account_number" db:"account_number"`
	ClearingAccountID *uuid.UUID      `json:"clearing_account_id,omitempty" db:"clearing_account_id"`
	CompanyID         uuid.UUID       `json:"company_id" db:"company_id"`
	TenantID          uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	Currency          string          `json:"currency" db:"currency"`
	IBAN              *string         `json:"iban,omitempty" db:"iban"`
	SwiftCode         *string         `json:"swift_code,omitempty" db:"swift_code"`
	BankAccountType   *string         `json:"bank_account_type,omitempty" db:"bank_account_type"`
	IsActive          bool            `json:"is_active" db:"is_active"`
	IsCash            bool            `json:"is_cash" db:"is_cash"`
	Custodian         *string         `json:"custodian,omitempty" db:"custodian"`
	Location          *string         `json:"location,omitempty" db:"location"`
	OpeningBalance    decimal.Decimal `json:"opening_balance" db:"opening_balance"`
	OpeningDate       *time.Time      `json:"opening_date,omitempty" db:"opening_date"`
	CurrentBalance    decimal.Decimal `json:"current_balance" db:"current_balance"`
	BalanceUpdatedAt  *time.Time      `json:"balance_updated_at,omitempty" db:"balance_updated_at"`
	CreatedAt         time.Time       `json:"created_at" db:"created_at"`
}

// BankTransaction represents the bank_transactions table (imported bank statements).
type BankTransaction struct {
	ID                    uuid.UUID       `json:"id" db:"id"`
	BankAccountID         uuid.UUID       `json:"bank_account_id" db:"bank_account_id"`
	TxnDate               time.Time       `json:"txn_date" db:"txn_date"`
	Description           *string         `json:"description,omitempty" db:"description"`
	Debit                 decimal.Decimal `json:"debit" db:"debit"`
	Credit                decimal.Decimal `json:"credit" db:"credit"`
	Direction             *string         `json:"direction,omitempty" db:"direction"`
	ReferenceNo           *string         `json:"reference_no,omitempty" db:"reference_no"`
	CounterpartyName      *string         `json:"counterparty_name,omitempty" db:"counterparty_name"`
	Classification        *string         `json:"classification,omitempty" db:"classification"`
	Matched               bool            `json:"matched" db:"matched"`
	MatchedPaymentEntryID *uuid.UUID      `json:"matched_payment_entry_id,omitempty" db:"matched_payment_entry_id"`
	MatchedGLEntryID      *uuid.UUID      `json:"matched_gl_entry_id,omitempty" db:"matched_gl_entry_id"`
	ImportedFrom          *string         `json:"imported_from,omitempty" db:"imported_from"`
	RawData               json.RawMessage `json:"raw_data,omitempty" db:"raw_data"`
	CompanyID             uuid.UUID       `json:"company_id" db:"company_id"`
	TenantID              uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	CreatedAt             time.Time       `json:"created_at" db:"created_at"`
}

// BankReconciliationDetail represents the bank_reconciliation_details table.
type BankReconciliationDetail struct {
	ID                  uuid.UUID        `json:"id" db:"id"`
	BankTransactionID   uuid.UUID        `json:"bank_transaction_id" db:"bank_transaction_id"`
	PaymentEntryID      *uuid.UUID       `json:"payment_entry_id,omitempty" db:"payment_entry_id"`
	GLEntryID           *uuid.UUID       `json:"gl_entry_id,omitempty" db:"gl_entry_id"`
	MatchScore          *decimal.Decimal `json:"match_score,omitempty" db:"match_score"`
	DifferenceAccountID *uuid.UUID       `json:"difference_account_id,omitempty" db:"difference_account_id"`
	ReconciledAt        *time.Time       `json:"reconciled_at,omitempty" db:"reconciled_at"`
	ReconciledBy        *uuid.UUID       `json:"reconciled_by,omitempty" db:"reconciled_by"`
	TenantID            uuid.UUID        `json:"tenant_id" db:"tenant_id"`
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

// ReconciliationRecord represents the reconciliation_records table.
type ReconciliationRecord struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	TenantID        uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	BankAccountID   uuid.UUID       `json:"bank_account_id" db:"bank_account_id"`
	PeriodNo        int             `json:"period_no" db:"period_no"`
	BankBalance     decimal.Decimal `json:"bank_balance" db:"bank_balance"`
	BookBalance     decimal.Decimal `json:"book_balance" db:"book_balance"`
	AdjustedBalance decimal.Decimal `json:"adjusted_balance" db:"adjusted_balance"`
	BankOnlyTotal   decimal.Decimal `json:"bank_only_total" db:"bank_only_total"`
	BookOnlyTotal   decimal.Decimal `json:"book_only_total" db:"book_only_total"`
	Status          string          `json:"status" db:"status"`
	ReconciledBy    *uuid.UUID      `json:"reconciled_by,omitempty" db:"reconciled_by"`
	ReconciledAt    *time.Time      `json:"reconciled_at,omitempty" db:"reconciled_at"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
}

// UnreconciledItem represents the unreconciled_items table.
type UnreconciledItem struct {
	ID                     uuid.UUID       `json:"id" db:"id"`
	ReconciliationRecordID uuid.UUID       `json:"reconciliation_record_id" db:"reconciliation_record_id"`
	ItemType               string          `json:"item_type" db:"item_type"`     // bank_only/book_only
	SourceType             string          `json:"source_type" db:"source_type"` // bank_transaction/gl_entry
	SourceID               uuid.UUID       `json:"source_id" db:"source_id"`
	TxnDate                time.Time       `json:"txn_date" db:"txn_date"`
	Description            *string         `json:"description,omitempty" db:"description"`
	Debit                  decimal.Decimal `json:"debit" db:"debit"`
	Credit                 decimal.Decimal `json:"credit" db:"credit"`
	Amount                 decimal.Decimal `json:"amount" db:"amount"`
	TenantID               uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	CreatedAt              time.Time       `json:"created_at" db:"created_at"`
}

// ReconciliationReport represents a bank reconciliation report.
type ReconciliationReport struct {
	ID              uuid.UUID          `json:"id" db:"id"`
	BankAccountID   uuid.UUID          `json:"bank_account_id" db:"bank_account_id"`
	PeriodNo        int                `json:"period_no" db:"period_no"`
	BankBalance     decimal.Decimal    `json:"bank_balance" db:"bank_balance"`
	BookBalance     decimal.Decimal    `json:"book_balance" db:"book_balance"`
	AdjustedBalance decimal.Decimal    `json:"adjusted_balance" db:"adjusted_balance"`
	BankOnlyItems   []UnreconciledItem `json:"bank_only_items,omitempty"`
	BookOnlyItems   []UnreconciledItem `json:"book_only_items,omitempty"`
	Status          string             `json:"status" db:"status"`
	ReconciledBy    *uuid.UUID         `json:"reconciled_by,omitempty" db:"reconciled_by"`
	ReconciledAt    *time.Time         `json:"reconciled_at,omitempty" db:"reconciled_at"`
}
