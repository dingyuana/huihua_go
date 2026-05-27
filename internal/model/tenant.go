package model

import (
	"time"

	"github.com/google/uuid"
)

// Tenant represents the tenants table.
type Tenant struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Domain    *string   `json:"domain,omitempty" db:"domain"`
	Plan      string    `json:"plan" db:"plan"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
