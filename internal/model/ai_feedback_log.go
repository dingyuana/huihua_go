package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AIFeedbackLog records each human action on a bank transaction so the AI
// system can learn from corrections (AC10).
type AIFeedbackLog struct {
	ID                  uuid.UUID     `json:"id" db:"id"`
	TenantID            uuid.UUID     `json:"tenant_id" db:"tenant_id"`
	BankTxnID           uuid.UUID     `json:"bank_txn_id" db:"bank_txn_id"`
	AISuggestedAction   *string       `json:"ai_suggested_action,omitempty" db:"ai_suggested_action"`
	AIConfidence        *int          `json:"ai_confidence,omitempty" db:"ai_confidence"`
	AIBusinessScene     *string       `json:"ai_business_scene,omitempty" db:"ai_business_scene"`
	HumanAction         string        `json:"human_action" db:"human_action"`
	HumanModifiedFields map[string]any `json:"human_modified_fields,omitempty" db:"human_modified_fields"`
	CreatedBy           *uuid.UUID    `json:"created_by,omitempty" db:"created_by"`
	CreatedAt           time.Time     `json:"created_at" db:"created_at"`
}

// HumanModifiedFieldsJSON returns the human_modified_fields as a JSON bytes
// for storage in a jsonb column.
func (l *AIFeedbackLog) HumanModifiedFieldsJSON() ([]byte, error) {
	if l.HumanModifiedFields == nil {
		return nil, nil
	}
	return json.Marshal(l.HumanModifiedFields)
}