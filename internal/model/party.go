package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Party represents a business partner (customer/supplier/both).
type Party struct {
	ID           uuid.UUID       `json:"id" db:"id"`
	TenantID     uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	PartyType    string          `json:"party_type" db:"party_type"`
	Name         string          `json:"name" db:"name"`
	TaxNumber    *string         `json:"tax_number,omitempty" db:"tax_number"`
	BankName     *string         `json:"bank_name,omitempty" db:"bank_name"`
	BankAccount  *string         `json:"bank_account,omitempty" db:"bank_account"`
	ContactName  *string         `json:"contact_name,omitempty" db:"contact_name"`
	ContactPhone *string         `json:"contact_phone,omitempty" db:"contact_phone"`
	CreditLimit  decimal.Decimal `json:"credit_limit" db:"credit_limit"`
	PaymentDays  int             `json:"payment_days" db:"payment_days"`
	IsActive     bool            `json:"is_active" db:"is_active"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at" db:"updated_at"`
}