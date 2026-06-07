package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ReimbursementInvoiceLink 关联报销单与进项发票
type ReimbursementInvoiceLink struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	ReimbursementID uuid.UUID       `json:"reimbursement_id" db:"reimbursement_id"`
	InvoiceID       uuid.UUID       `json:"invoice_id" db:"invoice_id"`
	InvoiceType     string          `json:"invoice_type" db:"invoice_type"` // expense_invoice
	LinkedAmount    decimal.Decimal `json:"linked_amount" db:"linked_amount"`
	LinkedBy        *uuid.UUID      `json:"linked_by,omitempty" db:"linked_by"`
	LinkedAt        time.Time       `json:"linked_at" db:"linked_at"`
}

// ReimbursementInvoiceLinkRequest 请求结构
type ReimbursementInvoiceLinkRequest struct {
	LinkedAmount decimal.Decimal `json:"linked_amount"`
}