package handler

import (
	"io"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
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

	return c.JSON(fiber.Map{
		"invoices": invoices,
		"total":    len(invoices),
	})
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
		InvoiceNo         string          `json:"invoice_no"`
		InvoiceType       string          `json:"invoice_type"`
		CustomerID        string          `json:"customer_id"`
		TaxID             string          `json:"tax_id,omitempty"`
		PostingDate       string          `json:"posting_date"`
		DueDate           string          `json:"due_date,omitempty"`
		TotalAmount       float64         `json:"total_amount"`
		TaxAmount         float64         `json:"tax_amount"`
		NetAmount         float64         `json:"net_amount"`
		OutstandingAmount float64         `json:"outstanding_amount"`
		CompanyID         string          `json:"company_id"`
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