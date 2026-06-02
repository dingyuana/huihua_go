package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Budget represents the budgets table.
type Budget struct {
	ID                  uuid.UUID  `json:"id" db:"id"`
	CompanyID           uuid.UUID  `json:"company_id" db:"company_id"`
	TenantID            uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	BudgetAgainst       string     `json:"budget_against" db:"budget_against"`
	FiscalYear          string     `json:"fiscal_year" db:"fiscal_year"`
	MonthlyDistribution *string    `json:"monthly_distribution,omitempty" db:"monthly_distribution"`
	Status              string     `json:"status" db:"status"`
	CreatedBy           *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
}

// BudgetAccount represents the budget_accounts table (budget line items per account).
type BudgetAccount struct {
	ID           uuid.UUID       `json:"id" db:"id"`
	BudgetID     uuid.UUID       `json:"budget_id" db:"budget_id"`
	AccountID    uuid.UUID       `json:"account_id" db:"account_id"`
	AnnualBudget decimal.Decimal `json:"annual_budget" db:"annual_budget"`
	TenantID     uuid.UUID       `json:"tenant_id" db:"tenant_id"`
}

// BudgetDistribution represents the budget_distributions table (period-based budget allocation).
type BudgetDistribution struct {
	ID              uuid.UUID        `json:"id" db:"id"`
	BudgetAccountID uuid.UUID        `json:"budget_account_id" db:"budget_account_id"`
	StartDate       time.Time        `json:"start_date" db:"start_date"`
	EndDate         time.Time        `json:"end_date" db:"end_date"`
	Amount          decimal.Decimal  `json:"amount" db:"amount"`
	Percent         *decimal.Decimal `json:"percent,omitempty" db:"percent"`
	TenantID        uuid.UUID        `json:"tenant_id" db:"tenant_id"`
}

// BudgetControlConfig represents the budget_control_configs table (rules for budget enforcement).
type BudgetControlConfig struct {
	ID                    uuid.UUID  `json:"id" db:"id"`
	AccountID             *uuid.UUID `json:"account_id,omitempty" db:"account_id"`
	CostCenterID          *uuid.UUID `json:"cost_center_id,omitempty" db:"cost_center_id"`
	FiscalYear            *string    `json:"fiscal_year,omitempty" db:"fiscal_year"`
	ActionAnnual          string     `json:"action_annual" db:"action_annual"`
	ActionMonthly         string     `json:"action_monthly" db:"action_monthly"`
	ApplicableOnMR        bool       `json:"applicable_on_mr" db:"applicable_on_mr"`
	ApplicableOnPO        bool       `json:"applicable_on_po" db:"applicable_on_po"`
	ApplicableOnActual    bool       `json:"applicable_on_actual" db:"applicable_on_actual"`
	ExceptionApproverRole *string    `json:"exception_approver_role,omitempty" db:"exception_approver_role"`
	CompanyID             uuid.UUID  `json:"company_id" db:"company_id"`
	TenantID              uuid.UUID  `json:"tenant_id" db:"tenant_id"`
}
