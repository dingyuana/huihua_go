package model

// OcrInvoiceRequest is the request body for OCR invoice recognition
type OcrInvoiceRequest struct {
	FileURL string `json:"file_url"` // Image URL or base64 encoded image
}

// OcrInvoiceResponse is the response from OCR invoice recognition
type OcrInvoiceResponse struct {
	InvoiceNo   string  `json:"invoice_no"`
	InvoiceCode string  `json:"invoice_code,omitempty"`
	InvoiceDate string  `json:"invoice_date"` // YYYY-MM-DD format
	TotalAmount float64 `json:"total_amount"`
	TaxAmount   float64 `json:"tax_amount"`
	VendorName  string  `json:"vendor_name"`
	InvoiceKind string  `json:"invoice_kind"` // paper_special/paper_normal/electronic_special/electronic_normal
	RawData     string  `json:"raw_data,omitempty"` // Raw OCR result JSON
}
