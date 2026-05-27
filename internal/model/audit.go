package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AuditLog represents the audit_logs table (append-only).
type AuditLog struct {
	ID            uuid.UUID       `json:"id" db:"id"`
	Action        string          `json:"action" db:"action"`
	ObjectType    string          `json:"object_type" db:"object_type"`
	ObjectID      uuid.UUID       `json:"object_id" db:"object_id"`
	TenantID      uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	ActorID       uuid.UUID       `json:"actor_id" db:"actor_id"`
	ActorName     *string         `json:"actor_name,omitempty" db:"actor_name"`
	ChangedFields json.RawMessage `json:"changed_fields,omitempty" db:"changed_fields"`
	Metadata      json.RawMessage `json:"metadata,omitempty" db:"metadata"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
}
