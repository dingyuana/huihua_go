package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ExpenseInvoice struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	TenantID        uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	CompanyID       uuid.UUID       `json:"company_id" db:"company_id"`
	InvoiceNo       string          `json:"invoice_no" db:"invoice_no"`
	InvoiceCode     *string         `json:"invoice_code,omitempty" db:"invoice_code"`
	InvoiceDate     time.Time       `json:"invoice_date" db:"invoice_date"`
	InvoiceKind     string          `json:"invoice_kind" db:"invoice_kind"`
	TaxAmount       decimal.Decimal `json:"tax_amount" db:"tax_amount"`
	TotalAmount     decimal.Decimal `json:"total_amount" db:"total_amount"`
	VendorID        *uuid.UUID      `json:"vendor_id,omitempty" db:"vendor_id"`
	VendorName      *string         `json:"vendor_name,omitempty" db:"vendor_name"`
	TaxID           *string         `json:"tax_id,omitempty" db:"tax_id"`
	VerifyStatus    string          `json:"verify_status" db:"verify_status"`
	VerifiedAt      *time.Time      `json:"verified_at,omitempty" db:"verified_at"`
	VerifyResult    *string         `json:"verify_result,omitempty" db:"verify_result"`
	DeductionStatus string          `json:"deduction_status" db:"deduction_status"`
	DeductedAt      *time.Time      `json:"deducted_at,omitempty" db:"deducted_at"`
	SourceFile      *string         `json:"source_file,omitempty" db:"source_file"`
	OcrData         *string         `json:"ocr_data,omitempty" db:"ocr_data"`
	Status          string          `json:"status" db:"status"`
	DocStatus       int16           `json:"docstatus" db:"docstatus"`
	Remark          *string         `json:"remark,omitempty" db:"remark"`
	CreatedBy       *uuid.UUID      `json:"created_by,omitempty" db:"created_by"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       *time.Time      `json:"updated_at,omitempty" db:"updated_at"`
}

// ExpenseInvoiceFilter for listing
type ExpenseInvoiceFilter struct {
	VendorID     *uuid.UUID
	VerifyStatus string
	FromDate     *time.Time
	ToDate       *time.Time
	Limit        int
	Offset       int
}

// ExpenseInvoiceCreateRequest is the request body for creating an expense invoice
type ExpenseInvoiceCreateRequest struct {
	CompanyID   uuid.UUID       `json:"company_id"`
	InvoiceNo   string          `json:"invoice_no"`
	InvoiceCode *string         `json:"invoice_code,omitempty"`
	InvoiceDate time.Time       `json:"invoice_date"`
	InvoiceKind string          `json:"invoice_kind"`
	TaxAmount   decimal.Decimal `json:"tax_amount"`
	TotalAmount decimal.Decimal `json:"total_amount"`
	VendorID    *uuid.UUID      `json:"vendor_id,omitempty"`
	VendorName  *string         `json:"vendor_name,omitempty"`
	TaxID       *string         `json:"tax_id,omitempty"`
	SourceFile  *string         `json:"source_file,omitempty"`
	OcrData     *string         `json:"ocr_data,omitempty"`
	Remark      *string         `json:"remark,omitempty"`
}

// ExpenseInvoiceVerifyResponse is the response for verify operation
type ExpenseInvoiceVerifyResponse struct {
	InvoiceID    uuid.UUID `json:"invoice_id"`
	VerifyStatus string    `json:"verify_status"`
	VerifyResult string    `json:"verify_result"`
	VerifiedAt   time.Time `json:"verified_at"`
}