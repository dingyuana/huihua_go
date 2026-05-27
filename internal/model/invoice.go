package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// SalesInvoice represents the sales_invoices table.
type SalesInvoice struct {
	ID                uuid.UUID       `json:"id" db:"id"`
	InvoiceNo         string          `json:"invoice_no" db:"invoice_no"`
	InvoiceType       string          `json:"invoice_type" db:"invoice_type"`
	CustomerID        uuid.UUID       `json:"customer_id" db:"customer_id"`
	TaxID             *string         `json:"tax_id,omitempty" db:"tax_id"`
	CompanyID         uuid.UUID       `json:"company_id" db:"company_id"`
	TenantID          uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	PostingDate       time.Time       `json:"posting_date" db:"posting_date"`
	DueDate           *time.Time      `json:"due_date,omitempty" db:"due_date"`
	TotalAmount       decimal.Decimal `json:"total_amount" db:"total_amount"`
	TaxAmount         decimal.Decimal `json:"tax_amount" db:"tax_amount"`
	NetAmount         decimal.Decimal `json:"net_amount" db:"net_amount"`
	OutstandingAmount decimal.Decimal `json:"outstanding_amount" db:"outstanding_amount"`
	Status            string          `json:"status" db:"status"`
	TaxTemplateID     *uuid.UUID      `json:"tax_template_id,omitempty" db:"tax_template_id"`
	ReturnAgainst     *uuid.UUID      `json:"return_against,omitempty" db:"return_against"`
	IsReturn          bool            `json:"is_return" db:"is_return"`
	DocStatus         int16           `json:"docstatus" db:"docstatus"`
	CreatedBy         *uuid.UUID      `json:"created_by,omitempty" db:"created_by"`
	CreatedAt         time.Time       `json:"created_at" db:"created_at"`
}
