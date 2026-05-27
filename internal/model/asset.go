package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// AssetCategory represents the asset_categories table.
type AssetCategory struct {
	ID                               uuid.UUID       `json:"id" db:"id"`
	Name                             string          `json:"name" db:"name"`
	DepreciationMethod               *string         `json:"depreciation_method,omitempty" db:"depreciation_method"`
	TotalNumberDepreciations         *int            `json:"total_number_depreciations,omitempty" db:"total_number_depreciations"`
	FrequencyOfDepreciation          *int            `json:"frequency_of_depreciation,omitempty" db:"frequency_of_depreciation"`
	Rate                             *decimal.Decimal `json:"rate,omitempty" db:"rate"`
	CompanyID                        uuid.UUID       `json:"company_id" db:"company_id"`
	TenantID                         uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	FixedAssetAccountID              *uuid.UUID      `json:"fixed_asset_account_id,omitempty" db:"fixed_asset_account_id"`
	AccumulatedDepreciationAccountID *uuid.UUID      `json:"accumulated_depreciation_account_id,omitempty" db:"accumulated_depreciation_account_id"`
	DepreciationExpenseAccountID     *uuid.UUID      `json:"depreciation_expense_account_id,omitempty" db:"depreciation_expense_account_id"`
	CWIPAccountID                    *uuid.UUID      `json:"cwip_account_id,omitempty" db:"cwip_account_id"`
	CreatedAt                        time.Time       `json:"created_at" db:"created_at"`
}

// Asset represents the assets table.
type Asset struct {
	ID                               uuid.UUID       `json:"id" db:"id"`
	AssetName                        string          `json:"asset_name" db:"asset_name"`
	AssetCategoryID                  *uuid.UUID      `json:"asset_category_id,omitempty" db:"asset_category_id"`
	ItemID                           *uuid.UUID      `json:"item_id,omitempty" db:"item_id"`
	PurchaseDate                     *time.Time      `json:"purchase_date,omitempty" db:"purchase_date"`
	GrossPurchaseAmount              decimal.Decimal `json:"gross_purchase_amount" db:"gross_purchase_amount"`
	AvailableForUseDate              *time.Time      `json:"available_for_use_date,omitempty" db:"available_for_use_date"`
	CalculateDepreciation            bool            `json:"calculate_depreciation" db:"calculate_depreciation"`
	DepreciationMethod               *string         `json:"depreciation_method,omitempty" db:"depreciation_method"`
	TotalNumberDepreciations         *int            `json:"total_number_depreciations,omitempty" db:"total_number_depreciations"`
	FrequencyOfDepreciation          *int            `json:"frequency_of_depreciation,omitempty" db:"frequency_of_depreciation"`
	ExpectedValueAfterUsefulLife     decimal.Decimal `json:"expected_value_after_useful_life" db:"expected_value_after_useful_life"`
	CurrentValue                     *decimal.Decimal `json:"current_value,omitempty" db:"current_value"`
	AccumulatedDepreciation          decimal.Decimal `json:"accumulated_depreciation" db:"accumulated_depreciation"`
	Status                           string          `json:"status" db:"status"`
	FixedAssetAccountID              *uuid.UUID      `json:"fixed_asset_account_id,omitempty" db:"fixed_asset_account_id"`
	DepreciationExpenseAccountID     *uuid.UUID      `json:"depreciation_expense_account_id,omitempty" db:"depreciation_expense_account_id"`
	AccumulatedDepreciationAccountID *uuid.UUID      `json:"accumulated_depreciation_account_id,omitempty" db:"accumulated_depreciation_account_id"`
	CompanyID                        uuid.UUID       `json:"company_id" db:"company_id"`
	TenantID                         uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	CustodianID                      *uuid.UUID      `json:"custodian_id,omitempty" db:"custodian_id"`
	Location                         *string         `json:"location,omitempty" db:"location"`
	CreatedBy                        *uuid.UUID      `json:"created_by,omitempty" db:"created_by"`
	CreatedAt                        time.Time       `json:"created_at" db:"created_at"`
}

// DepreciationSchedule represents the depreciation_schedules table.
type DepreciationSchedule struct {
	ID                 uuid.UUID       `json:"id" db:"id"`
	AssetID            uuid.UUID       `json:"asset_id" db:"asset_id"`
	ScheduleDate       time.Time       `json:"schedule_date" db:"schedule_date"`
	DepreciationAmount decimal.Decimal `json:"depreciation_amount" db:"depreciation_amount"`
	Posted             bool            `json:"posted" db:"posted"`
	JournalEntryID     *uuid.UUID      `json:"journal_entry_id,omitempty" db:"journal_entry_id"`
	TenantID           uuid.UUID       `json:"tenant_id" db:"tenant_id"`
}
