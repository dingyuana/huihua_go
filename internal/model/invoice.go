package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// InvoiceStatus represents the status of an invoice.
type InvoiceStatus string

const (
	InvoiceStatusDraft     InvoiceStatus = "draft"
	InvoiceStatusSubmitted InvoiceStatus = "submitted"
	InvoiceStatusVerified  InvoiceStatus = "verified"
	InvoiceStatusInvalid   InvoiceStatus = "invalid"
	InvoiceStatusUnpaid    InvoiceStatus = "unpaid"
)

// SalesInvoice represents the sales_invoices table.
type SalesInvoice struct {
	ID                  uuid.UUID       `json:"id" db:"id"`
	InvoiceNo           string          `json:"invoice_no" db:"invoice_no"`
	InvoiceType         string          `json:"invoice_type" db:"invoice_type"`
	CustomerID          uuid.UUID       `json:"customer_id" db:"customer_id"`
	TaxID               *string         `json:"tax_id,omitempty" db:"tax_id"`
	CompanyID           uuid.UUID       `json:"company_id" db:"company_id"`
	TenantID            uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	PostingDate         time.Time       `json:"posting_date" db:"posting_date"`
	DueDate             *time.Time      `json:"due_date,omitempty" db:"due_date"`
	TotalAmount         decimal.Decimal `json:"total_amount" db:"total_amount"`
	TaxAmount           decimal.Decimal `json:"tax_amount" db:"tax_amount"`
	NetAmount           decimal.Decimal `json:"net_amount" db:"net_amount"`
	OutstandingAmount   decimal.Decimal `json:"outstanding_amount" db:"outstanding_amount"`
	Status              string          `json:"status" db:"status"`
	TaxTemplateID       *uuid.UUID      `json:"tax_template_id,omitempty" db:"tax_template_id"`
	ReturnAgainst       *uuid.UUID      `json:"return_against,omitempty" db:"return_against"`
	IsReturn            bool            `json:"is_return" db:"is_return"`
	InvoiceCategory     *string         `json:"invoice_category,omitempty" db:"invoice_category"`
	Remark              *string         `json:"remark,omitempty" db:"remark"`
	SourceRedInvoiceNo  *string         `json:"source_red_invoice_no,omitempty" db:"source_red_invoice_no"`
	DocStatus           int16           `json:"docstatus" db:"docstatus"`
	CreatedBy           *uuid.UUID      `json:"created_by,omitempty" db:"created_by"`
	CreatedAt           time.Time       `json:"created_at" db:"created_at"`
}

// InvoiceLineItem represents a line item on an invoice.
type InvoiceLineItem struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	InvoiceID   uuid.UUID       `json:"invoice_id" db:"invoice_id"`
	ItemCode    string          `json:"item_code" db:"item_code"`
	Description string          `json:"description" db:"description"`
	Quantity    decimal.Decimal `json:"quantity" db:"quantity"`
	UnitPrice   decimal.Decimal `json:"unit_price" db:"unit_price"`
	TaxRate     decimal.Decimal `json:"tax_rate" db:"tax_rate"`
	TaxAmount   decimal.Decimal `json:"tax_amount" db:"tax_amount"`
	NetAmount   decimal.Decimal `json:"net_amount" db:"net_amount"`
	TotalAmount decimal.Decimal `json:"total_amount" db:"total_amount"`
	Unit        string          `json:"unit" db:"unit"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
}

// InvoiceImportRequest represents a batch import request for invoices.
type InvoiceImportRequest struct {
	Invoices []InvoiceImportItem `json:"invoices"`
}

// InvoiceImportItem represents a single invoice in an import batch.
type InvoiceImportItem struct {
	InvoiceNo                string          `json:"invoice_no"`
	InvoiceType              string          `json:"invoice_type"` // sale/purchase/credit_note
	CustomerID               string          `json:"customer_id"`
	TaxID                    string          `json:"tax_id,omitempty"`
	PostingDate              string          `json:"posting_date"` // Yyyy-MM-DD
	DueDate                  string          `json:"due_date,omitempty"`
	TotalAmount              float64         `json:"total_amount"`
	TaxAmount                float64         `json:"tax_amount"`
	NetAmount                float64         `json:"net_amount"`
	Status                   string          `json:"status"`
	IsReturn                 bool            `json:"is_return,omitempty"`
	SourceRedInvoiceNo       string          `json:"source_red_invoice_no,omitempty"`
	Remark                   string          `json:"remark,omitempty"`
	LineItems                []LineItemInput `json:"line_items,omitempty"`
}

// LineItemInput represents a line item in import data.
type LineItemInput struct {
	ItemCode    string  `json:"item_code"`
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	TaxRate     float64 `json:"tax_rate"`
	Unit        string  `json:"unit"`
}

// InvoiceParseRequest represents an OCR parse request.
type InvoiceParseRequest struct {
	FileURL string `json:"file_url"`
}

// InvoiceMatchBankRequest represents a request to match invoice to bank transaction.
type InvoiceMatchBankRequest struct {
	BankTxnID string `json:"bank_txn_id"`
}

// InvoiceFileImportResult represents the result of a file-based invoice import.
type InvoiceFileImportResult struct {
	TotalRows  int               `json:"total_rows"`
	Imported   int               `json:"imported"`
	Failed     int               `json:"failed"`
	FailedRows []FailedRowDetail `json:"failed_rows,omitempty"`
}

// InvoiceFilter represents query filters for listing invoices.
type InvoiceFilter struct {
	CustomerID *uuid.UUID `json:"customer_id,omitempty"`
	Status     string     `json:"status,omitempty"`
	FromDate   *time.Time `json:"from_date,omitempty"`
	ToDate     *time.Time `json:"to_date,omitempty"`
	Limit      int        `json:"limit,omitempty"`
	Offset     int        `json:"offset,omitempty"`
}

// InvoiceBatchPreviewResult represents the result of batch import preview.
type InvoiceBatchPreviewResult struct {
	BatchID     string                 `json:"batch_id"`
	TotalRows   int                    `json:"total_rows"`
	ValidRows   int                    `json:"valid_rows"`
	ErrorRows   int                    `json:"error_rows"`
	DuplicateRows int                  `json:"duplicate_rows"`
	Details     []InvoicePreviewDetail `json:"details"`
}

// InvoicePreviewDetail represents a single invoice row in preview.
type InvoicePreviewDetail struct {
	RowIndex       int     `json:"row_index"`
	InvoiceNo      string  `json:"invoice_no"`
	InvoiceType    string  `json:"invoice_type"`
	CustomerName   string  `json:"customer_name"`
	PostingDate    string  `json:"posting_date"`
	TotalAmount    float64 `json:"total_amount"`
	NetAmount      float64 `json:"net_amount"`
	TaxAmount      float64 `json:"tax_amount"`
	Status         string  `json:"status"`
	ValidationErr  string  `json:"validation_err,omitempty"`
	IsDuplicate    bool    `json:"is_duplicate,omitempty"`
	DuplicateInfo  string  `json:"duplicate_info,omitempty"`
}

// InvoiceBatchConfirmRequest represents a request to confirm batch import.
type InvoiceBatchConfirmRequest struct {
	BatchID      string   `json:"batch_id"`
	SelectedIDs  []string `json:"selected_ids"`
	CorrectedData []map[string]interface{} `json:"corrected_data,omitempty"`
}

// InvoiceBatchConfirmResult represents the result of batch import confirmation.
type InvoiceBatchConfirmResult struct {
	Imported   int               `json:"imported"`
	Skipped    int               `json:"skipped"`
	Errors     int               `json:"errors"`
	FailedRows []FailedRowDetail `json:"failed_rows,omitempty"`
}

// InvoiceConfirmRequest represents a request to confirm a sales invoice.
type InvoiceConfirmRequest struct {
	InvoiceID string `json:"invoice_id"`
}
