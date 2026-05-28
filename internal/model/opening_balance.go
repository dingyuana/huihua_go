package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// OpeningBalanceEntry represents an opening balance record for an account at a given period.
type OpeningBalanceEntry struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	TenantID    uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	CompanyID   uuid.UUID       `json:"company_id" db:"company_id"`
	AccountID   uuid.UUID       `json:"account_id" db:"account_id"`
	DebitAmount decimal.Decimal `json:"debit_amount" db:"debit_amount"`
	CreditAmount decimal.Decimal `json:"credit_amount" db:"credit_amount"`
	PeriodNo    int             `json:"period_no" db:"period_no"`
	IsVerified  bool            `json:"is_verified" db:"is_verified"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
}

// OpeningBalanceImportRow represents a single row from the Excel import.
type OpeningBalanceImportRow struct {
	AccountCode   string          `json:"account_code"`
	AccountName   string          `json:"account_name"`
	DebitBalance  decimal.Decimal `json:"debit_balance"`
	CreditBalance decimal.Decimal `json:"credit_balance"`
}

// OpeningBalanceValidationResult holds the result of validating opening balances.
type OpeningBalanceValidationResult struct {
	Valid        bool                       `json:"valid"`
	TotalDebit   decimal.Decimal            `json:"total_debit"`
	TotalCredit  decimal.Decimal            `json:"total_credit"`
	BalanceDiff  decimal.Decimal            `json:"balance_diff"`
	Errors       []OpeningBalanceError     `json:"errors,omitempty"`
	Warnings     []OpeningBalanceWarning   `json:"warnings,omitempty"`
}

// OpeningBalanceError represents a validation error for a specific account.
type OpeningBalanceError struct {
	AccountCode string `json:"account_code"`
	AccountName string `json:"account_name"`
	Message     string `json:"message"`
}

// OpeningBalanceWarning represents a non-critical issue during validation.
type OpeningBalanceWarning struct {
	AccountCode string `json:"account_code"`
	AccountName string `json:"account_name"`
	Message     string `json:"message"`
}

// TrialBalanceEntry represents a single line in the trial balance.
type TrialBalanceEntry struct {
	AccountCode   string          `json:"account_code"`
	AccountName   string          `json:"account_name"`
	AccountType   string          `json:"account_type"`
	RootType      string          `json:"root_type"`
	DebitBalance  decimal.Decimal `json:"debit_balance"`
	CreditBalance decimal.Decimal `json:"credit_balance"`
}

// TrialBalance represents the full trial balance for a period.
type TrialBalance struct {
	PeriodNo       int                `json:"period_no"`
	PeriodName     string             `json:"period_name,omitempty"`
	StartDate      time.Time          `json:"start_date,omitempty"`
	EndDate        time.Time          `json:"end_date,omitempty"`
	Entries        []TrialBalanceEntry `json:"entries"`
	TotalDebit     decimal.Decimal    `json:"total_debit"`
	TotalCredit    decimal.Decimal    `json:"total_credit"`
	IsBalanced     bool               `json:"is_balanced"`
}

// IncomeStatementAccountEntry is a line item in the income statement detail.
type IncomeStatementAccountEntry struct {
	AccountCode   string          `json:"account_code"`
	AccountName   string          `json:"account_name"`
	PeriodDebit   decimal.Decimal `json:"period_debit"`
	PeriodCredit  decimal.Decimal `json:"period_credit"`
	NetAmount     decimal.Decimal `json:"net_amount"`
}

// IncomeStatement represents a P&L report for a period.
type IncomeStatement struct {
	PeriodNo        int                             `json:"period_no"`
	PeriodName      string                          `json:"period_name"`
	StartDate       time.Time                       `json:"start_date"`
	EndDate         time.Time                       `json:"end_date"`
	TotalIncome     decimal.Decimal                 `json:"total_income"`
	TotalExpense    decimal.Decimal                 `json:"total_expense"`
	NetProfit       decimal.Decimal                 `json:"net_profit"`
	IncomeDetails  []IncomeStatementAccountEntry   `json:"income_details,omitempty"`
	ExpenseDetails []IncomeStatementAccountEntry   `json:"expense_details,omitempty"`
}

// BalanceSheetAccountEntry is a line item in the balance sheet.
type BalanceSheetAccountEntry struct {
	AccountCode string          `json:"account_code"`
	AccountName string          `json:"account_name"`
	Balance     decimal.Decimal `json:"balance"`
}

// BalanceSheet represents the balance sheet as of a period end.
type BalanceSheet struct {
	PeriodNo         int                       `json:"period_no"`
	PeriodName       string                    `json:"period_name"`
	AsOfDate         time.Time                 `json:"as_of_date"`
	TotalAssets      decimal.Decimal           `json:"total_assets"`
	TotalLiabilities decimal.Decimal           `json:"total_liabilities"`
	TotalEquity      decimal.Decimal           `json:"total_equity"`
	DerivedEquity    decimal.Decimal           `json:"derived_equity,omitempty"`
	IsBalanced       bool                      `json:"is_balanced"`
	AssetEntries     []BalanceSheetAccountEntry `json:"asset_entries,omitempty"`
	LiabilityEntries []BalanceSheetAccountEntry `json:"liability_entries,omitempty"`
	EquityEntries    []BalanceSheetAccountEntry `json:"equity_entries,omitempty"`
}