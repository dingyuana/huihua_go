package model

import (
	"time"

	"github.com/google/uuid"
)

// VoucherTemplate represents a voucher template (凭证模板).
type VoucherTemplate struct {
	ID             uuid.UUID            `json:"id" db:"id"`
	TenantID       uuid.UUID            `json:"tenant_id" db:"tenant_id"`
	Name           string               `json:"name" db:"name"`
	Description    string               `json:"description,omitempty" db:"description"`
	NumberPrefix   string               `json:"number_prefix" db:"number_prefix"`
	IsActive       bool                 `json:"is_active" db:"is_active"`
	Classification *string              `json:"classification,omitempty" db:"classification"` // bank_fee / business_receipt / ...
	ApprovalFlowID *uuid.UUID           `json:"approval_flow_id,omitempty" db:"approval_flow_id"` // bound approval flow
	CreatedAt      time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at" db:"updated_at"`
	Lines          []VoucherTemplateLine `json:"lines,omitempty"`
}

// VoucherTemplateLine represents a line item in a voucher template (模板行).
type VoucherTemplateLine struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	TemplateID       uuid.UUID  `json:"template_id" db:"template_id"`
	AccountID        uuid.UUID  `json:"account_id" db:"account_id"`
	AccountCode      string     `json:"account_code,omitempty" db:"account_code"`
	AccountName      string     `json:"account_name,omitempty" db:"account_name"`
	DrAmountTemplate string     `json:"dr_amount_template,omitempty" db:"dr_amount_template"` // e.g., "{{amount}}"
	CrAmountTemplate string     `json:"cr_amount_template,omitempty" db:"cr_amount_template"`
	SummaryTemplate  string     `json:"summary_template,omitempty" db:"summary_template"` // e.g., "支付{{party}}"
	DimensionType    string     `json:"dimension_type,omitempty" db:"dimension_type"`     // e.g., "department", "project"
	DimensionValue   string     `json:"dimension_value,omitempty" db:"dimension_value"`   // e.g., "{{department_id}}"
	LineOrder        int        `json:"line_order" db:"line_order"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
}

// VoucherNumberingRule represents the voucher numbering rule (编号规则).
type VoucherNumberingRule struct {
	ID          uuid.UUID `json:"id" db:"id"`
	TenantID    uuid.UUID `json:"tenant_id" db:"tenant_id"`
	Prefix      string    `json:"prefix" db:"prefix"`           // e.g., "PZ"
	NextNumber  int       `json:"next_number" db:"next_number"` // next sequence number
	DateFormat  string    `json:"date_format" db:"date_format"` // e.g., "20060102"
	ResetRule   string    `json:"reset_rule" db:"reset_rule"`   // yearly/monthly/daily
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// CreateTemplateRequest represents the request to create a voucher template.
type CreateTemplateRequest struct {
	Name           string                       `json:"name" validate:"required"`
	Description    string                       `json:"description"`
	NumberPrefix   string                       `json:"number_prefix"`
	IsActive       bool                         `json:"is_active"`
	Classification *string                      `json:"classification,omitempty"`
	ApprovalFlowID *uuid.UUID                   `json:"approval_flow_id,omitempty"`
	Lines          []CreateTemplateLineRequest `json:"lines"`
}

// CreateTemplateLineRequest represents a line item in create template request.
type CreateTemplateLineRequest struct {
	AccountID        uuid.UUID `json:"account_id" validate:"required"`
	DrAmountTemplate string    `json:"dr_amount_template"`
	CrAmountTemplate string    `json:"cr_amount_template"`
	SummaryTemplate  string    `json:"summary_template"`
	DimensionType    string    `json:"dimension_type"`
	DimensionValue   string    `json:"dimension_value"`
	LineOrder        int       `json:"line_order"`
}

// UpdateTemplateRequest represents the request to update a voucher template.
type UpdateTemplateRequest struct {
	Name           string                       `json:"name"`
	Description    string                       `json:"description"`
	NumberPrefix   string                       `json:"number_prefix"`
	IsActive       *bool                        `json:"is_active"`
	Classification *string                      `json:"classification,omitempty"`
	ApprovalFlowID *uuid.UUID                   `json:"approval_flow_id,omitempty"`
	Lines          []CreateTemplateLineRequest `json:"lines"`
}

// NumberingRuleRequest represents the request to create/update numbering rule.
type NumberingRuleRequest struct {
	Prefix     string `json:"prefix" validate:"required"`
	DateFormat string `json:"date_format"`
	ResetRule  string `json:"reset_rule"` // yearly/monthly/daily
}

// VoucherNumberResponse represents the response for generated voucher number.
type VoucherNumberResponse struct {
	VoucherNumber string `json:"voucher_number"`
	Sequence      int    `json:"sequence"`
}
