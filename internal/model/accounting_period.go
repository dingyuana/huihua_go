package model

import (
	"time"

	"github.com/google/uuid"
)

// AccountingPeriod represents a fiscal accounting period.
type AccountingPeriod struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	TenantID   uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	PeriodNo   int        `json:"period_no" db:"period_no"`
	PeriodName string     `json:"period_name" db:"period_name"`
	StartDate  time.Time  `json:"start_date" db:"start_date"`
	EndDate    time.Time  `json:"end_date" db:"end_date"`
	Status    string     `json:"status" db:"status"`
	ClosedBy   *uuid.UUID `json:"closed_by,omitempty" db:"closed_by"`
	ClosedAt   *time.Time `json:"closed_at,omitempty" db:"closed_at"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}