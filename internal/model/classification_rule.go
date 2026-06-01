package model

import (
	"time"

	"github.com/google/uuid"
)

// ClassificationRule represents the classification_rules table.
type ClassificationRule struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	TenantID        uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	Name            string     `json:"name" db:"name"`
	RuleType        string     `json:"rule_type" db:"rule_type"` // keyword, keyword_regex, counterparty_match
	Pattern         string     `json:"pattern" db:"pattern"`
	MatchField      string     `json:"match_field" db:"match_field"` // description, counterparty
	Direction       string     `json:"direction" db:"direction"` // in, out, '' (both)
	Classification  string     `json:"classification" db:"classification"` // business_receipt, business_payment, bank_fee, interest_income, internal_transfer, tax_payment
	Priority        int        `json:"priority" db:"priority"`
	IsActive        bool       `json:"is_active" db:"is_active"`
	DebitAccountID  *uuid.UUID `json:"debit_account_id,omitempty" db:"debit_account_id"`
	CreditAccountID *uuid.UUID `json:"credit_account_id,omitempty" db:"credit_account_id"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// CreateRuleRequest is the request payload for creating a classification rule.
type CreateRuleRequest struct {
	Name            string     `json:"name"`
	RuleType        string     `json:"rule_type"`
	Pattern         string     `json:"pattern"`
	MatchField      string     `json:"match_field"`
	Direction       string     `json:"direction"`
	Classification  string     `json:"classification"`
	Priority        int        `json:"priority"`
	IsActive        bool       `json:"is_active"`
	DebitAccountID  *uuid.UUID `json:"debit_account_id,omitempty"`
	CreditAccountID *uuid.UUID `json:"credit_account_id,omitempty"`
}

// UpdateRuleRequest is the request payload for updating a classification rule.
type UpdateRuleRequest struct {
	Name            *string    `json:"name,omitempty"`
	RuleType        *string    `json:"rule_type,omitempty"`
	Pattern         *string    `json:"pattern,omitempty"`
	MatchField      *string    `json:"match_field,omitempty"`
	Direction       *string    `json:"direction,omitempty"`
	Classification  *string    `json:"classification,omitempty"`
	Priority        *int       `json:"priority,omitempty"`
	IsActive        *bool      `json:"is_active,omitempty"`
	DebitAccountID  *uuid.UUID `json:"debit_account_id,omitempty"`
	CreditAccountID *uuid.UUID `json:"credit_account_id,omitempty"`
}

// ReorderPriorityRequest is the request payload for reordering rule priorities.
type ReorderPriorityRequest struct {
	RuleIDs []string `json:"rule_ids"` // ordered list of rule IDs (priority ASC)
}

// RuleMatchRequest is the request payload for testing rule matching.
type RuleMatchRequest struct {
	Description    string `json:"description"`
	Counterparty   string `json:"counterparty"`
	Amount         string `json:"amount"`
	Direction      string `json:"direction"` // in, out
}

// RuleMatchResult is the result of a rule match operation.
type RuleMatchResult struct {
	Matched        bool   `json:"matched"`
	RuleID         *uuid.UUID `json:"rule_id,omitempty"`
	RuleName       *string `json:"rule_name,omitempty"`
	Classification *string `json:"classification,omitempty"`
}
