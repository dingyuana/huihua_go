package model

import (
	"time"
)

// ImportBatch represents a pending import batch (in-memory or DB-backed)
type ImportBatch struct {
	BatchID   string       `json:"batch_id"`
	TotalRows int          `json:"total_rows"`
	ValidRows int          `json:"valid_rows"`
	ErrorRows int          `json:"error_rows"`
	Status    string       `json:"status"` // pending/confirmed
	Rows      []ImportRow `json:"rows"`
	CreatedAt time.Time   `json:"created_at"`
}

type ImportRow struct {
	RowIndex    int     `json:"row_index"`
	InvoiceNo   string  `json:"invoice_no"`
	InvoiceCode string  `json:"invoice_code,omitempty"`
	InvoiceDate string  `json:"invoice_date"`
	TotalAmount float64 `json:"total_amount"`
	TaxAmount   float64 `json:"tax_amount"`
	VendorName  string  `json:"vendor_name,omitempty"`
	Status      string  `json:"status"` // valid/error/duplicate
	ErrorMsg    string  `json:"error_msg,omitempty"`
}

// ImportUploadResponse is returned after uploading an Excel file
type ImportUploadResponse struct {
	BatchID    string       `json:"batch_id"`
	TotalRows  int          `json:"total_rows"`
	ValidRows  int          `json:"valid_rows"`
	ErrorRows  int          `json:"error_rows"`
	ValidCount int          `json:"valid_count"`
	ErrorCount int          `json:"error_count"`
	Details    []ImportRow `json:"details"`
}

// ImportConfirmRequest is the request body for confirming an import
type ImportConfirmRequest struct {
	BatchID     string   `json:"batch_id"`
	SelectedIDs []string `json:"selected_ids"` // which rows to import
}

// ImportConfirmResponse is the response for confirming an import
type ImportConfirmResponse struct {
	Imported int `json:"imported"`
	Failed   int `json:"failed"`
}
