package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Account represents the accounts table (chart of accounts, tree structure with nested set).
type Account struct {
	ID             uuid.UUID       `json:"id" db:"id"`
	Code           string          `json:"code" db:"code"`
	Name           string          `json:"name" db:"name"`
	AccountType    *string         `json:"account_type,omitempty" db:"account_type"`
	RootType       *string         `json:"root_type,omitempty" db:"root_type"`
	ParentID       *uuid.UUID      `json:"parent_id,omitempty" db:"parent_id"`
	Lft            int             `json:"lft" db:"lft"`
	Rgt            int             `json:"rgt" db:"rgt"`
	Level          int             `json:"level" db:"level"`
	Path           string          `json:"path,omitempty" db:"path"`
	IsGroup        bool            `json:"is_group" db:"is_group"`
	CompanyID      uuid.UUID       `json:"company_id" db:"company_id"`
	TenantID       uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	Currency       string          `json:"currency" db:"currency"`
	IsActive       bool            `json:"is_active" db:"is_active"`
	OpeningBalance decimal.Decimal `json:"opening_balance" db:"opening_balance"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at" db:"updated_at"`
	Children       []*Account      `json:"children,omitempty"`
}
