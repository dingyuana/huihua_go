package handler

import (
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"huihua/finance/internal/model"
	"huihua/finance/internal/service"
)

// BankTransactionHandler handles bank transaction HTTP requests.
type BankTransactionHandler struct {
	svc *service.BankTransactionService
}

// NewBankTransactionHandler creates a new BankTransactionHandler.
func NewBankTransactionHandler(svc *service.BankTransactionService) *BankTransactionHandler {
	return &BankTransactionHandler{svc: svc}
}

// List returns bank transactions with filters.
// GET /api/v1/bank-transactions?bank_account_id=xxx&start_date=2024-01-01&end_date=2024-12-31&status=pending&page=1&page_size=50
func (h *BankTransactionHandler) List(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	// Parse query params
	bankAccountIDStr := c.Query("bank_account_id")
	bankAccountID, err := uuid.Parse(bankAccountIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid bank_account_id"})
	}

	filter := model.BankTxnFilter{
		Page:     c.QueryInt("page", 1),
		PageSize: c.QueryInt("page_size", 50),
	}

	if startDate := c.Query("start_date"); startDate != "" {
		if t, err := parseDate(startDate); err == nil {
			filter.StartDate = &t
		}
	}
	if endDate := c.Query("end_date"); endDate != "" {
		if t, err := parseDate(endDate); err == nil {
			filter.EndDate = &t
		}
	}
	if status := c.Query("status"); status != "" {
		s := model.BankTransactionStatus(status)
		filter.Status = &s
	}
	if search := c.Query("search"); search != "" {
		filter.Search = &search
	}

	txns, total, err := h.svc.ListTransactions(c.Context(), tenantID, bankAccountID, filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"data":      txns,
		"total":     total,
		"page":      filter.Page,
		"page_size": filter.PageSize,
	})
}

// PreviewExcel parses Excel file and returns column names and sample data.
// POST /api/v1/bank-transactions/preview
func (h *BankTransactionHandler) PreviewExcel(c *fiber.Ctx) error {
	log.Println("=== PreviewExcel called ===")

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

	// Get first sheet
	sheetName := excelFile.GetSheetName(0)
	log.Println("Sheet name:", sheetName)

	rows, err := excelFile.GetRows(sheetName)
	if err != nil {
		log.Println("GetRows error:", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "read excel error: " + err.Error()})
	}

	log.Printf("Total rows in sheet: %d", len(rows))

	// Debug: print first 20 rows to find real header
	log.Println("=== First 20 rows of Excel ===")
	for i := 0; i < len(rows) && i < 20; i++ {
		log.Printf("Row %2d: %v", i, rows[i])
	}
	log.Println("===============================")

	if len(rows) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "empty excel file"})
	}

	// Find real header row (skip empty rows and title rows)
	headerRowIndex := 0
	for i, row := range rows {
		nonEmpty := 0
		for _, cell := range row {
			if strings.TrimSpace(cell) != "" {
				nonEmpty++
			}
		}
		// If row has 5+ non-empty cells, likely it's the real header
		if nonEmpty >= 5 {
			headerRowIndex = i
			log.Printf("Found potential header at row %d: %v", i, row)
			break
		}
	}

	// Extract column names from header row
	columns := make([]string, len(rows[headerRowIndex]))
	for i, col := range rows[headerRowIndex] {
		columns[i] = strings.TrimSpace(col)
	}
	log.Printf("Using columns found at row %d: %v", headerRowIndex, columns)

	// Extract sample data (from data rows after header)
	sampleData := [][]string{}
	for i := headerRowIndex + 1; i < len(rows) && i <= headerRowIndex+5; i++ {
		if len(rows[i]) > 0 {
			rowData := make([]string, len(rows[i]))
			for j, cell := range rows[i] {
				rowData[j] = strings.TrimSpace(cell)
			}
			sampleData = append(sampleData, rowData)
		}
	}

	log.Printf("Sample data rows: %d", len(sampleData))
	log.Println("=== PreviewExcel completed ===")

	return c.JSON(fiber.Map{
		"columns":         columns,
		"sample":          sampleData,
		"total_rows":      len(rows) - headerRowIndex - 1, // exclude header and skipped rows
		"header_row":      headerRowIndex,
	})
}

// Import imports bank transactions from Excel.
// POST /api/v1/bank-transactions/import
func (h *BankTransactionHandler) Import(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	bankAccountIDStr := c.Query("bank_account_id")
	bankAccountID, err := uuid.Parse(bankAccountIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid bank_account_id"})
	}

	data := c.Body()
	if len(data) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "empty body"})
	}

	result, err := h.svc.ImportFromExcel(c.Context(), tenantID, bankAccountID, data)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(result)
}

// Classify re-classifies a single bank transaction.
// POST /api/v1/bank-transactions/:id/classify
func (h *BankTransactionHandler) Classify(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	result, err := h.svc.ClassifyTransaction(c.Context(), tenantID, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(result)
}

// MarkMatched marks transactions as matched.
// POST /api/v1/bank-transactions/:id/mark-matched
func (h *BankTransactionHandler) MarkMatched(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	var req struct {
		JournalEntryID string `json:"journal_entry_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}

	journalEntryID, err := uuid.Parse(req.JournalEntryID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid journal_entry_id"})
	}

	err = h.svc.MarkAsMatched(c.Context(), tenantID, []uuid.UUID{id}, journalEntryID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "matched"})
}

// GetUnmatched returns all unmatched transactions.
// GET /api/v1/bank-transactions/unmatched?bank_account_id=xxx
func (h *BankTransactionHandler) GetUnmatched(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	bankAccountIDStr := c.Query("bank_account_id")
	bankAccountID, err := uuid.Parse(bankAccountIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid bank_account_id"})
	}

	txns, err := h.svc.GetUnmatched(c.Context(), tenantID, bankAccountID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"data": txns})
}

// GetByID returns a single bank transaction.
// GET /api/v1/bank-transactions/:id
func (h *BankTransactionHandler) GetByID(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	txn, err := h.svc.GetTransaction(c.Context(), tenantID, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"data": txn})
}

// Delete deletes a bank transaction.
// DELETE /api/v1/bank-transactions/:id
func (h *BankTransactionHandler) Delete(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	err = h.svc.DeleteTransaction(c.Context(), tenantID, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "deleted"})
}

// parseDate parses a date string in various formats.
func parseDate(s string) (time.Time, error) {
	formats := []string{
		"2006-01-02",
		"2006/01/02",
		"01/02/2006",
		"2006-01-02 15:04:05",
		"2006/01/02 15:04:05",
		"02/01/2006",
		"2006年01月02日",
	}
	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fiber.ErrBadRequest
}
