package handler

import (
	"io"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
	"huihua/finance/internal/model"
	"huihua/finance/internal/service"
)

// InvoiceHandler handles invoice HTTP requests.
type InvoiceHandler struct {
	svc *service.InvoiceService
}

// NewInvoiceHandler creates a new InvoiceHandler.
func NewInvoiceHandler(svc *service.InvoiceService) *InvoiceHandler {
	return &InvoiceHandler{svc: svc}
}

// List returns a list of invoices.
// GET /api/v1/invoices
func (h *InvoiceHandler) List(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	filters := model.InvoiceFilter{}

	// Parse query parameters
	if customerID := c.Query("customer_id"); customerID != "" {
		id, err := uuid.Parse(customerID)
		if err == nil {
			filters.CustomerID = &id
		}
	}
	if status := c.Query("status"); status != "" {
		filters.Status = status
	}
	if fromDate := c.Query("from_date"); fromDate != "" {
		if t, err := time.Parse("2006-01-02", fromDate); err == nil {
			filters.FromDate = &t
		}
	}
	if toDate := c.Query("to_date"); toDate != "" {
		if t, err := time.Parse("2006-01-02", toDate); err == nil {
			filters.ToDate = &t
		}
	}

	invoices, err := h.svc.List(c.Context(), tenantID, filters)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	result := make([]map[string]interface{}, len(invoices))
	for i, inv := range invoices {
		lineItems, _ := h.svc.GetLineItems(c.Context(), tenantID, inv.ID)
		itemsResult := make([]map[string]interface{}, len(lineItems))
		for j, item := range lineItems {
			itemsResult[j] = map[string]interface{}{
				"id":          item.ID,
				"item_code":   item.ItemCode,
				"description": item.Description,
				"quantity":    item.Quantity.String(),
				"unit_price":  item.UnitPrice.String(),
				"tax_rate":    item.TaxRate.String(),
				"tax_amount":  item.TaxAmount.String(),
				"net_amount":  item.NetAmount.String(),
				"total_amount": item.TotalAmount.String(),
				"unit":        item.Unit,
			}
		}

		var remarkStr string
		if inv.Remark != nil {
			remarkStr = *inv.Remark
		}

		var sourceRedInvoiceNoStr string
		if inv.SourceRedInvoiceNo != nil {
			sourceRedInvoiceNoStr = *inv.SourceRedInvoiceNo
		}

		var invoiceCodeStr string
		if inv.InvoiceCode != nil {
			invoiceCodeStr = *inv.InvoiceCode
		}

		var invoiceCategoryStr string
		if inv.InvoiceCategory != nil {
			invoiceCategoryStr = *inv.InvoiceCategory
		}

		var taxIdStr string
		if inv.TaxID != nil {
			taxIdStr = *inv.TaxID
		}

		result[i] = map[string]interface{}{
			"id":                    inv.ID,
			"type":                  inv.InvoiceType,
			"invoice_no":            inv.InvoiceNo,
			"invoice_code":          invoiceCodeStr,
			"invoice_category":      invoiceCategoryStr,
			"customer_id":           inv.CustomerID,
			"customer_name":         inv.CustomerName,
			"tax_id":                taxIdStr,
			"posting_date":          inv.PostingDate.Format("2006-01-02"),
			"total_amount":          inv.TotalAmount.String(),
			"tax_amount":            inv.TaxAmount.String(),
			"net_amount":            inv.NetAmount.String(),
			"outstanding_amount":    inv.OutstandingAmount.String(),
			"status":                mapInvoiceStatus(inv.Status),
			"docstatus":             inv.DocStatus,
			"is_return":             inv.IsReturn,
			"source_red_invoice_no": sourceRedInvoiceNoStr,
			"remark":                remarkStr,
			"created_at":            inv.CreatedAt.Format("2006-01-02 15:04:05"),
			"line_items":            itemsResult,
		}
	}

	return c.JSON(fiber.Map{
		"list":  result,
		"total": len(result),
	})
}

func mapInvoiceStatus(status string) string {
	switch status {
	case "正常", "unpaid":
		return "draft"
	case "已确认", "verified":
		return "verified"
	case "部分核销", "partially_paid":
		return "partially_paid"
	case "已核销", "paid":
		return "paid"
	case "已红冲-全额", "已红冲":
		return "cancelled"
	default:
		return "draft"
	}
}

// GetByID retrieves an invoice by ID.
// GET /api/v1/invoices/:id
func (h *InvoiceHandler) GetByID(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid invoice id",
		})
	}

	inv, err := h.svc.GetByID(c.Context(), tenantID, id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "invoice not found",
		})
	}

	return c.JSON(inv)
}

// Create handles manual invoice creation.
// POST /api/v1/invoices
func (h *InvoiceHandler) Create(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	var req struct {
		InvoiceNo           string  `json:"invoice_no"`
		InvoiceType         string  `json:"invoice_type"`
		CustomerID          string  `json:"customer_id"`
		TaxID               string  `json:"tax_id,omitempty"`
		PostingDate         string  `json:"posting_date"`
		DueDate             string  `json:"due_date,omitempty"`
		TotalAmount         float64 `json:"total_amount"`
		TaxAmount           float64 `json:"tax_amount"`
		NetAmount           float64 `json:"net_amount"`
		OutstandingAmount   float64 `json:"outstanding_amount"`
		CompanyID           string  `json:"company_id"`
		IsReturn            bool    `json:"is_return,omitempty"`
		SourceRedInvoiceNo  string  `json:"source_red_invoice_no,omitempty"`
		Remark              string  `json:"remark,omitempty"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Parse required UUIDs
	customerID, err := uuid.Parse(req.CustomerID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid customer_id",
		})
	}

	companyID, err := uuid.Parse(req.CompanyID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid company_id",
		})
	}

	// Parse posting date
	postingDate, err := time.Parse("2006-01-02", req.PostingDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid posting_date format, use YYYY-MM-DD",
		})
	}

	// Build invoice
	inv := &model.SalesInvoice{
		InvoiceNo:         req.InvoiceNo,
		InvoiceType:       req.InvoiceType,
		CustomerID:        customerID,
		CompanyID:         companyID,
		PostingDate:       postingDate,
		TotalAmount:       decimal.NewFromFloat(req.TotalAmount),
		TaxAmount:         decimal.NewFromFloat(req.TaxAmount),
		NetAmount:         decimal.NewFromFloat(req.NetAmount),
		OutstandingAmount: decimal.NewFromFloat(req.OutstandingAmount),
		Status:            "draft",
	}

	// Parse optional due date
	if req.DueDate != "" {
		t, err := time.Parse("2006-01-02", req.DueDate)
		if err == nil {
			inv.DueDate = &t
		}
	}

	// Parse optional tax_id
	if req.TaxID != "" {
		inv.TaxID = &req.TaxID
	}

	// Apply red-letter invoice fields
	inv.IsReturn = req.IsReturn
	if req.SourceRedInvoiceNo != "" {
		s := req.SourceRedInvoiceNo
		inv.SourceRedInvoiceNo = &s
	}
	if req.Remark != "" {
		r := req.Remark
		inv.Remark = &r
	}

	// Resolve red→blue invoice linkage by invoice_no
	if inv.IsReturn && inv.SourceRedInvoiceNo != nil {
		if srcInv, lerr := h.svc.GetByInvoiceNo(c.Context(), tenantID, *inv.SourceRedInvoiceNo); lerr == nil && srcInv != nil {
			id := srcInv.ID
			inv.ReturnAgainst = &id
		}
	}

	result, err := h.svc.Create(c.Context(), tenantID, inv)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(result)
}

// ImportFromExcel handles batch import from Excel.
// POST /api/v1/invoices/import
func (h *InvoiceHandler) ImportFromExcel(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	var req model.InvoiceImportRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	invoices, err := h.svc.ImportFromExcel(c.Context(), tenantID, &req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"imported": len(invoices),
		"invoices": invoices,
	})
}

// PreviewExcel parses Excel file and returns column names and sample data for invoice import.
// POST /api/v1/invoices/preview-excel
func (h *InvoiceHandler) PreviewExcel(c *fiber.Ctx) error {
	log.Println("=== Invoice PreviewExcel called ===")

	file, err := c.FormFile("file")
	if err != nil {
		log.Println("FormFile error:", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "file required: " + err.Error()})
	}
	log.Printf("File received: name=%s, size=%d", file.Filename, file.Size)

	f, err := file.Open()
	if err != nil {
		log.Println("Open file error:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "open file error: " + err.Error()})
	}
	defer f.Close()

	excelFile, err := excelize.OpenReader(f)
	if err != nil {
		log.Println("OpenReader error:", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid excel file: " + err.Error()})
	}
	defer excelFile.Close()

	sheetName := excelFile.GetSheetName(0)
	log.Println("Sheet name:", sheetName)

	rows, err := excelFile.GetRows(sheetName)
	if err != nil {
		log.Println("GetRows error:", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "read excel error: " + err.Error()})
	}

	log.Printf("Total rows in sheet: %d", len(rows))

	if len(rows) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "empty excel file"})
	}

	headerRowIndex := 0
	for i, row := range rows {
		nonEmpty := 0
		for _, cell := range row {
			if strings.TrimSpace(cell) != "" {
				nonEmpty++
			}
		}
		if nonEmpty >= 5 {
			headerRowIndex = i
			log.Printf("Found potential header at row %d: %v", i, row)
			break
		}
	}

	columns := make([]string, len(rows[headerRowIndex]))
	for i, col := range rows[headerRowIndex] {
		columns[i] = strings.TrimSpace(col)
	}
	log.Printf("Using columns found at row %d: %v", headerRowIndex, columns)

	validRows := 0
	for i := headerRowIndex + 1; i < len(rows); i++ {
		hasContent := false
		for _, cell := range rows[i] {
			if strings.TrimSpace(cell) != "" {
				hasContent = true
				break
			}
		}
		if hasContent {
			validRows++
		}
	}

	sampleData := [][]string{}
	for i := headerRowIndex + 1; i < len(rows) && i <= headerRowIndex+10; i++ {
		if len(rows[i]) > 0 {
			rowData := make([]string, len(rows[i]))
			for j, cell := range rows[i] {
				rowData[j] = strings.TrimSpace(cell)
			}
			sampleData = append(sampleData, rowData)
		}
	}

	log.Printf("Valid data rows: %d", validRows)
	log.Println("=== Invoice PreviewExcel completed ===")

	return c.JSON(fiber.Map{
		"columns":    columns,
		"sample":     sampleData,
		"total_rows": validRows,
		"header_row": headerRowIndex,
	})
}

// BatchImportPreview handles batch import preview with AI deduplication check.
// POST /api/v1/invoices/sales/import/preview
func (h *InvoiceHandler) BatchImportPreview(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "file required: " + err.Error()})
	}

	f, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "open file error: " + err.Error()})
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "read file error: " + err.Error()})
	}

	result, err := h.svc.BatchImportPreview(c.Context(), tenantID, data)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(result)
}

// BatchImportConfirm confirms and imports the batch invoices.
// POST /api/v1/invoices/sales/import/confirm
func (h *InvoiceHandler) BatchImportConfirm(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	var req model.InvoiceBatchConfirmRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body: " + err.Error()})
	}

	result, err := h.svc.BatchImportConfirm(c.Context(), tenantID, &req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(result)
}

// ConfirmSalesInvoice confirms a sales invoice and generates accounts receivable.
// POST /api/v1/invoices/sales/confirm
func (h *InvoiceHandler) ConfirmSalesInvoice(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	userID := c.Locals("user_id").(uuid.UUID)

	var req model.InvoiceConfirmRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body: " + err.Error()})
	}

	invoiceID, err := uuid.Parse(req.InvoiceID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid invoice_id"})
	}

	err = h.svc.ConfirmSalesInvoice(c.Context(), tenantID, invoiceID, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "invoice confirmed and accounts receivable generated"})
}

// ConfirmPurchaseInvoice confirms a purchase invoice and generates accounts payable.
// POST /api/v1/invoices/purchase/confirm
func (h *InvoiceHandler) ConfirmPurchaseInvoice(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	userID := c.Locals("user_id").(uuid.UUID)

	var req model.InvoiceConfirmRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body: " + err.Error()})
	}

	invoiceID, err := uuid.Parse(req.InvoiceID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid invoice_id"})
	}

	err = h.svc.ConfirmPurchaseInvoice(c.Context(), tenantID, invoiceID, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "invoice confirmed and accounts payable generated"})
}

// BatchConfirm handles POST /api/v1/invoices/batch-confirm
func (h *InvoiceHandler) BatchConfirm(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	userID := c.Locals("user_id").(uuid.UUID)

	var req struct {
		InvoiceIDs []string `json:"invoice_ids"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	success := 0
	failed := 0
	var failedList []fiber.Map

	for _, idStr := range req.InvoiceIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			failed++
			failedList = append(failedList, fiber.Map{"id": idStr, "reason": "invalid uuid"})
			continue
		}
		if err := h.svc.ConfirmSalesInvoice(c.Context(), tenantID, id, userID); err != nil {
			failed++
			failedList = append(failedList, fiber.Map{"id": idStr, "reason": err.Error()})
			continue
		}
		success++
	}

	return c.JSON(fiber.Map{
		"success_count": success,
		"failed_count":  failed,
		"failed_list":   failedList,
	})
}

// ImportExcelFile handles file-upload-based invoice import.
// POST /api/v1/invoices/import-excel
func (h *InvoiceHandler) ImportExcelFile(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "file required: " + err.Error()})
	}

	f, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "open file error: " + err.Error()})
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "read file error: " + err.Error()})
	}

	tenantID := c.Locals("tenant_id").(uuid.UUID)
	result, err := h.svc.ImportFromExcelFile(c.Context(), tenantID, data)
	if err != nil {
		log.Printf("Invoice import error: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(result)
}

func (h *InvoiceHandler) Delete(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid invoice id"})
	}
	if err := h.svc.Delete(c.Context(), tenantID, id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "deleted"})
}

// Parse handles OCR parsing of invoices.
// POST /api/v1/invoices/parse
func (h *InvoiceHandler) Parse(c *fiber.Ctx) error {
	var req model.InvoiceParseRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error":   "OCR not implemented, use manual import",
		"message": "Invoice parsing via OCR is not yet available. Please use manual entry or Excel import.",
	})
}

// Update updates mutable fields of an invoice (amounts, dates, etc).
// PUT /api/v1/invoices/:id
func (h *InvoiceHandler) Update(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid invoice id"})
	}

	var body map[string]interface{}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	allowed := map[string]bool{
		"total_amount": true, "tax_amount": true, "net_amount": true,
		"outstanding_amount": true, "posting_date": true, "due_date": true,
		"status": true, "remark": true, "tax_id": true,
	}
	fields := make(map[string]interface{}, len(body))
	for k, v := range body {
		if allowed[k] {
			fields[k] = v
		}
	}
	if len(fields) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "no updatable fields provided"})
	}

	if err := h.svc.Update(c.Context(), tenantID, id, fields); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "updated"})
}

// UpdateStatus updates the status of an invoice.
// PUT /api/v1/invoices/:id/status
func (h *InvoiceHandler) UpdateStatus(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid invoice id",
		})
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.Status == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "status is required",
		})
	}

	if err := h.svc.UpdateStatus(c.Context(), tenantID, id, req.Status); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "status updated",
	})
}

// ListUnmatched returns invoices eligible for matching (outstanding_amount > 0).
// GET /api/v1/invoices/unmatched?party_id=xxx
func (h *InvoiceHandler) ListUnmatched(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	var partyID *uuid.UUID
	if pid := c.Query("party_id"); pid != "" {
		parsed, err := uuid.Parse(pid)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid party_id"})
		}
		partyID = &parsed
	}

	invoices, err := h.svc.ListUnmatchedInvoices(c.Context(), tenantID, partyID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"data": invoices})
}

// MatchToBank matches an invoice to a bank transaction.
// POST /api/v1/invoices/:id/match
func (h *InvoiceHandler) MatchToBank(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	invoiceID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid invoice id",
		})
	}

	var req model.InvoiceMatchBankRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	bankTxnID, err := uuid.Parse(req.BankTxnID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid bank_txn_id",
		})
	}

	// Amount is optional - if not provided, use full outstanding amount
	var amount decimal.Decimal

	if err := h.svc.MatchToBankTxn(c.Context(), tenantID, invoiceID, bankTxnID, amount); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "matched successfully",
	})
}
