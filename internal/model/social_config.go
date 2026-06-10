package model

import (
	"time"

	"github.com/google/uuid"
)

// SocialConfig represents the wage_social_config table.
// 社保公积金比例配置表（系统级配置，不是按员工配置）
type SocialConfig struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	TenantID      uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	InsuranceType string     `json:"insurance_type" db:"insurance_type"`   // pension/medical/unemployment/injury/maternity/housing
	InsuranceName string     `json:"insurance_name" db:"insurance_name"`   // 显示名称：养老保险/医疗保险等
	CompanyRate   string     `json:"company_rate" db:"company_rate"`       // 公司比例，如 "0.16" = 16%
	PersonalRate  string     `json:"personal_rate" db:"personal_rate"`     // 个人比例，如 "0.08" = 8%
	IsActive      bool       `json:"is_active" db:"is_active"`
	Remark        *string    `json:"remark,omitempty" db:"remark"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}
