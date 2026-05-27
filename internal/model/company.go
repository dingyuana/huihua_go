package model

import (
	"time"

	"github.com/google/uuid"
)

// CompanySettings represents the company/accounting setup configuration.
type CompanySettings struct {
	ID                   uuid.UUID `json:"id" db:"id"`
	TenantID             uuid.UUID `json:"tenant_id" db:"tenant_id"`
	CompanyName          string    `json:"company_name" db:"company_name"`
	FiscalYearStartMonth int       `json:"fiscal_year_start_month" db:"fiscal_year_start_month"`
	EnableDate           time.Time `json:"enable_date" db:"enable_date"`
	DefaultCurrency     string    `json:"default_currency" db:"default_currency"`
	ChartTemplate        string    `json:"chart_template" db:"chart_template"`
	IsInitialized        bool      `json:"is_initialized" db:"is_initialized"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time `json:"updated_at" db:"updated_at"`
}