package model

import (
	"time"

	"github.com/google/uuid"
)

// VoucherType represents a accounting voucher type (凭证类型).
type VoucherType struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	TenantID    uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	Code        string     `json:"code" db:"code"`
	Name        string     `json:"name" db:"name"`
	Description string     `json:"description,omitempty" db:"description"`
	SortOrder   int        `json:"sort_order" db:"sort_order"`
	IsActive    bool       `json:"is_active" db:"is_active"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// CreateVoucherTypeRequest is the request body for creating a voucher type.
type CreateVoucherTypeRequest struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	SortOrder   int    `json:"sort_order"`
}

// UpdateVoucherTypeRequest is the request body for updating a voucher type.
type UpdateVoucherTypeRequest struct {
	Code        *string `json:"code,omitempty"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	SortOrder   *int    `json:"sort_order,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

// SysConfig represents a system configuration key-value pair.
type SysConfig struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	TenantID    uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	ConfigKey   string     `json:"config_key" db:"config_key"`
	ConfigValue string     `json:"config_value" db:"config_value"`
	Description string     `json:"description,omitempty" db:"description"`
	Group       string     `json:"group,omitempty" db:"group_name"`
	IsSystem    bool       `json:"is_system" db:"is_system"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// CreateSysConfigRequest is the request body for creating a system config.
type CreateSysConfigRequest struct {
	ConfigKey   string `json:"config_key"`
	ConfigValue string `json:"config_value"`
	Description string `json:"description,omitempty"`
	Group       string `json:"group,omitempty"`
	IsSystem    bool   `json:"is_system,omitempty"`
}

// UpdateSysConfigRequest is the request body for updating a system config.
type UpdateSysConfigRequest struct {
	ConfigValue *string `json:"config_value,omitempty"`
	Description *string `json:"description,omitempty"`
	Group       *string `json:"group,omitempty"`
	IsSystem    *bool   `json:"is_system,omitempty"`
}

// BatchGetConfigsRequest is the request body for batch getting configs.
type BatchGetConfigsRequest struct {
	Keys []string `json:"keys"`
}