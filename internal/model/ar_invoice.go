package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ArInvoiceStatus string

const (
	ArInvoiceStatusDraft     ArInvoiceStatus = "draft"
	ArInvoiceStatusConfirmed ArInvoiceStatus = "confirmed"
	ArInvoiceStatusReversed  ArInvoiceStatus = "reversed"
)

type ArInvoice struct {
	ID          uuid.UUID       `db:"id"`
	TenantID    uuid.UUID       `db:"tenant_id"`
	CompanyID   uuid.UUID       `db:"company_id"`
	CustomerID  uuid.UUID       `db:"customer_id"`
	InvoiceID   uuid.UUID       `db:"invoice_id"`
	InvoiceNo   string          `db:"invoice_no"`
	Amount      decimal.Decimal `db:"amount"`
	DueDate     *time.Time      `db:"due_date"`
	Status      string          `db:"status"`
	SourceType  string          `db:"source_type"`
	Remark      *string         `json:"remark,omitempty" db:"remark"`
	CreatedBy   *uuid.UUID      `db:"created_by"`
	CreatedAt   time.Time       `db:"created_at"`
	ConfirmedAt *time.Time      `db:"confirmed_at"`
	ConfirmedBy *uuid.UUID      `db:"confirmed_by"`
	ApprovedBy  *uuid.UUID      `json:"approved_by,omitempty" db:"approved_by"`
	ApprovedAt  *time.Time      `json:"approved_at,omitempty" db:"approved_at"`
}