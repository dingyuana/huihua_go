package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ApInvoiceStatus string

const (
	ApInvoiceStatusDraft     ApInvoiceStatus = "draft"
	ApInvoiceStatusConfirmed ApInvoiceStatus = "confirmed"
	ApInvoiceStatusReversed  ApInvoiceStatus = "reversed"
)

type ApInvoice struct {
	ID          uuid.UUID       `db:"id"`
	TenantID    uuid.UUID       `db:"tenant_id"`
	CompanyID   uuid.UUID       `db:"company_id"`
	SupplierID  uuid.UUID       `db:"supplier_id"`
	InvoiceID   uuid.UUID       `db:"invoice_id"`
	InvoiceNo   string          `db:"invoice_no"`
	Amount      decimal.Decimal `db:"amount"`
	DueDate     *time.Time      `db:"due_date"`
	Status      string          `db:"status"`
	SourceType  string          `db:"source_type"`
	CreatedBy   *uuid.UUID      `db:"created_by"`
	CreatedAt   time.Time       `db:"created_at"`
	ConfirmedAt *time.Time      `db:"confirmed_at"`
	ConfirmedBy *uuid.UUID      `db:"confirmed_by"`
	ApprovedBy  *uuid.UUID      `json:"approved_by,omitempty" db:"approved_by"`
	ApprovedAt  *time.Time      `json:"approved_at,omitempty" db:"approved_at"`
}
