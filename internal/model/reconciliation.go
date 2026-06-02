package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ReconciliationPair represents a matched pair for reconciliation.
type ReconciliationPair struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	TenantID    uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	SourceType  string          `json:"source_type" db:"source_type"` // bank_txn / invoice
	SourceID    uuid.UUID       `json:"source_id" db:"source_id"`
	TargetType  string          `json:"target_type" db:"target_type"` // invoice / payment
	TargetID    uuid.UUID       `json:"target_id" db:"target_id"`
	Amount      decimal.Decimal `json:"amount" db:"amount"`
	Status      string          `json:"status" db:"status"`           // pending / matched / confirmed
	MatchLevel  string          `json:"match_level" db:"match_level"` // L1/L2/L3/L4/L5
	MatchedAt   *time.Time      `json:"matched_at" db:"matched_at"`
	ConfirmedAt *time.Time      `json:"confirmed_at" db:"confirmed_at"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
}

// ReconciliationResult is the result of a reconciliation run.
type ReconciliationResult struct {
	TotalScanned   int                  `json:"total_scanned"`
	Matched        int                  `json:"matched"`
	Unmatched      int                  `json:"unmatched"`
	Pairs          []ReconciliationPair `json:"pairs"`
	UnmatchedItems []UnmatchedItem      `json:"unmatched_items"`
}

// UnmatchedItem is an item that couldn't be matched.
type UnmatchedItem struct {
	Type      string          `json:"type"`
	ID        uuid.UUID       `json:"id"`
	Date      time.Time       `json:"date"`
	Amount    decimal.Decimal `json:"amount"`
	PartyName string          `json:"party_name"`
	Summary   string          `json:"summary"`
}

// MatchSuggestion represents a suggested match.
type MatchSuggestion struct {
	SourceType string          `json:"source_type"`
	SourceID   uuid.UUID       `json:"source_id"`
	TargetType string          `json:"target_type"`
	TargetID   uuid.UUID       `json:"target_id"`
	Amount     decimal.Decimal `json:"amount"`
	MatchLevel string          `json:"match_level"`
	Confidence float64         `json:"confidence"`
}
