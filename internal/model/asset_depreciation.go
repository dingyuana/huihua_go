package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// DepreciationMethod represents the depreciation calculation method.
type DepreciationMethod string

const (
	DepreciationMethodStraightLine      DepreciationMethod = "straight_line"
	DepreciationMethodDoubleDeclining   DepreciationMethod = "double_declining"
	DepreciationMethodUnitsOfProduction DepreciationMethod = "units_of_production"
)

// AssetDepreciation represents the depreciation_schedules table.
// It stores the planned depreciation for each period for a specific asset.
type AssetDepreciation struct {
	ID                 uuid.UUID       `json:"id" db:"id"`
	AssetID            uuid.UUID       `json:"asset_id" db:"asset_id"`
	ScheduleDate       time.Time       `json:"schedule_date" db:"schedule_date"`
	DepreciationAmount decimal.Decimal `json:"depreciation_amount" db:"depreciation_amount"`
	Posted             bool            `json:"posted" db:"posted"`
	JournalEntryID     *uuid.UUID      `json:"journal_entry_id,omitempty" db:"journal_entry_id"`
	TenantID           uuid.UUID       `json:"tenant_id" db:"tenant_id"`
}

// DepreciationRun represents a depreciation run batch.
// It tracks when monthly depreciation was executed.
type DepreciationRun struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	PeriodNo    int             `json:"period_no" db:"period_no"`
	RunDate     time.Time       `json:"run_date" db:"run_date"`
	TenantID    uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	CompanyID   uuid.UUID       `json:"company_id" db:"company_id"`
	VoucherNo   string          `json:"voucher_no" db:"voucher_no"`
	VoucherType *string         `json:"voucher_type,omitempty" db:"voucher_type"`
	TotalAmount decimal.Decimal `json:"total_amount" db:"total_amount"`
	AssetCount  int             `json:"asset_count" db:"asset_count"`
	Status      string          `json:"status" db:"status"`
	CreatedBy   *uuid.UUID      `json:"created_by,omitempty" db:"created_by"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
}

// CreateScheduleRequest is the request body for creating a depreciation schedule.
type CreateScheduleRequest struct {
	Method       DepreciationMethod `json:"method"`
	UsefulLife   int                `json:"useful_life"` // in months
	SalvageValue decimal.Decimal    `json:"salvage_value"`
	PurchaseDate time.Time          `json:"purchase_date"`
}

// RunDepreciationRequest is the request body for running monthly depreciation.
type RunDepreciationRequest struct {
	PeriodNo int `json:"period_no"`
}
