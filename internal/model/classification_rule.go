package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ClassificationRule represents the classification_rules table for bank transaction auto-classification.
type ClassificationRule struct {
	ID            uuid.UUID       `json:"id" db:"id"`
	TenantID      uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	RuleName      string          `json:"rule_name" db:"rule_name"`
	Priority      int             `json:"priority" db:"priority"`
	Keywords      json.RawMessage `json:"keywords" db:"keywords"` // JSONB string array for OR matching
	AccountID     uuid.UUID       `json:"account_id" db:"account_id"`
	PartyType     *string         `json:"party_type,omitempty" db:"party_type"`
	DebitDirection *string        `json:"debit_direction,omitempty" db:"debit_direction"` // debit/credit/both
	IsActive      bool            `json:"is_active" db:"is_active"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at" db:"updated_at"`
}

// RuleMatchRequest is the request payload for testing rule matching.
type RuleMatchRequest struct {
	Keywords string          `json:"keywords"` //摘要文本
	Amount   decimal.Decimal `json:"amount"`
	Direction string         `json:"direction"` // debit/credit
}

// RuleMatchResult is the result of a rule match operation.
type RuleMatchResult struct {
	Matched    bool              `json:"matched"`
	RuleID     *uuid.UUID        `json:"rule_id,omitempty"`
	RuleName   *string           `json:"rule_name,omitempty"`
	AccountID  *uuid.UUID        `json:"account_id,omitempty"`
	AccountCode *string          `json:"account_code,omitempty"`
	AccountName *string          `json:"account_name,omitempty"`
	Priority   *int              `json:"priority,omitempty"`
	PartyType  *string           `json:"party_type,omitempty"`
}

// CreateRuleRequest is the request payload for creating a classification rule.
type CreateRuleRequest struct {
	RuleName       string   `json:"rule_name"`
	Keywords       []string `json:"keywords"` // array of keywords for OR matching
	AccountID      string   `json:"account_id"`
	PartyType      *string  `json:"party_type,omitempty"`
	DebitDirection *string  `json:"debit_direction,omitempty"`
	IsActive       *bool    `json:"is_active,omitempty"`
}

// UpdateRuleRequest is the request payload for updating a classification rule.
type UpdateRuleRequest struct {
	RuleName       *string  `json:"rule_name,omitempty"`
	Keywords       []string `json:"keywords,omitempty"`
	AccountID      *string  `json:"account_id,omitempty"`
	PartyType      *string  `json:"party_type,omitempty"`
	DebitDirection *string  `json:"debit_direction,omitempty"`
	IsActive       *bool    `json:"is_active,omitempty"`
}

// ReorderPriorityRequest is the request payload for reordering rule priorities.
type ReorderPriorityRequest struct {
	RuleIDs []string `json:"rule_ids"` // ordered list of rule IDs (priority ASC)
}