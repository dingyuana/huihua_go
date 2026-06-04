package handler

import (
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"huihua/finance/internal/model"
	"huihua/finance/internal/service"
)

func isTaxRelated(desc string) bool {
	keywords := []string{"税款", "税金", "缴税", "扣税", "增值税", "所得税", "附加税", "社保", "公积金", "实时缴税"}
	lower := strings.ToLower(desc)
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

// BankTransactionHandler handles bank transaction HTTP requests.
type BankTransactionHandler struct {
	svc     *service.BankTransactionService
	autoSvc *service.VoucherAutoGenerateService
}

// NewBankTransactionHandler creates a new BankTransactionHandler.
func NewBankTransactionHandler(svc *service.BankTransactionService) *BankTransactionHandler {
	return &BankTransactionHandler{svc: svc}
}

// InjectAutoGenSvc sets the auto-generate service after construction
// (circular init dependency: autoGenSvc depends on many repos created after this handler).
func (h *BankTransactionHandler) InjectAutoGenSvc(autoSvc *service.VoucherAutoGenerateService) {
	h.autoSvc = autoSvc
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
	if classification := c.Query("classification"); classification != "" {
		filter.Classification = &classification
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

	// Count only rows with at least 1 non-blank cell (skip trailing blank rows)
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

	log.Printf("Valid data rows: %d", validRows)
	log.Println("=== PreviewExcel completed ===")

	return c.JSON(fiber.Map{
		"columns":    columns,
		"sample":     sampleData,
		"total_rows": validRows,
		"header_row": headerRowIndex,
	})
}

// Import imports bank transactions from Excel.
// POST /api/v1/bank-transactions/import?bank_account_id=xxx
func (h *BankTransactionHandler) Import(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	bankAccountIDStr := c.Query("bank_account_id")
	bankAccountID, err := uuid.Parse(bankAccountIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid bank_account_id"})
	}

	// Accept file upload as multipart/form-data (same as PreviewExcel)
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

	autoCreatePartyFlag := strings.ToLower(c.Query("auto_create_party"))
	autoCreateParty := !(autoCreatePartyFlag == "false" || autoCreatePartyFlag == "0")

	result, err := h.svc.ImportFromExcel(c.Context(), tenantID, bankAccountID, data, autoCreateParty)
	if err != nil {
		log.Printf("Import error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	log.Printf("Import result: total=%d success=%d failed=%d failedRows=%v",
		result.TotalRows, result.SuccessCount, result.FailedCount, result.FailedRows)
	for _, fr := range result.FailedReasons {
		log.Printf("  Row %d: %s", fr.Row, fr.Reason)
	}

	return c.JSON(fiber.Map{
		"total_rows":           result.TotalRows,
		"success_count":        result.SuccessCount,
		"failed_count":         result.FailedCount,
		"failed_rows":          result.FailedRows,
		"failed_reasons":       result.FailedReasons,
		"auto_created_parties": result.AutoCreatedParties,
		"hint":                 "导入成功。所有流水进入待确认状态，请在出纳工作台进行审核确认后生成凭证或单据",
	})
}

type batchConfirmReq struct {
	IDs []uuid.UUID `json:"ids"`
}

// BatchConfirm manually confirms bank transactions and triggers document/voucher generation.
// POST /api/v1/bank-transactions/batch-confirm
func (h *BankTransactionHandler) BatchConfirm(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	userID := c.Locals("user_id").(uuid.UUID)

	var req batchConfirmReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if len(req.IDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ids required"})
	}

	vouchersCreated := 0
	documentsNeeded := 0
	voucherNos := []string{}
	documentIDs := []string{}
	failed := []string{}

	for _, id := range req.IDs {
		txn, err := h.svc.GetTransaction(c.Context(), tenantID, id)
		if err != nil {
			failed = append(failed, fmt.Sprintf("%s: not found", id))
			continue
		}
		if txn.Matched || txn.Confirmed {
			failed = append(failed, fmt.Sprintf("%s: already processed", id))
			continue
		}
		if txn.Classification == nil {
			failed = append(failed, fmt.Sprintf("%s: no classification", id))
			continue
		}
		cls := *txn.Classification

		// Mark as confirmed first — this is the manual confirmation step
		_ = h.svc.MarkAsConfirmed(c.Context(), tenantID, id)

		switch cls {
		case "bank_fee", "interest_income", "tax_payment", "social_security", "insurance_fee":
			je, err := h.autoSvc.GenerateFromBankTxn(c.Context(), tenantID, id, userID)
			if err != nil {
				failed = append(failed, fmt.Sprintf("%s: %v", id, err))
				continue
			}
			_ = h.svc.MarkAsMatched(c.Context(), tenantID, []uuid.UUID{id}, je.ID)
			vouchersCreated++
			voucherNos = append(voucherNos, je.VoucherNo)
		case "business_receipt", "business_payment":
			documentsNeeded++
			documentIDs = append(documentIDs, id.String())
		default:
			documentsNeeded++
			documentIDs = append(documentIDs, id.String())
		}
	}

	return c.JSON(fiber.Map{
		"confirmed":        len(req.IDs) - len(failed),
		"failed":           len(failed),
		"failed_details":   failed,
		"vouchers_created": vouchersCreated,
		"voucher_nos":      voucherNos,
		"documents_needed": documentsNeeded,
		"document_ids":     documentIDs,
	})
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

// ClassifyAll re-classifies all pending transactions for the current bank account.
// POST /api/v1/bank-transactions/classify-all
func (h *BankTransactionHandler) ClassifyAll(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	bankAccountIDStr := c.Query("bank_account_id")
	if bankAccountIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "bank_account_id query parameter required"})
	}
	bankAccountID, err := uuid.Parse(bankAccountIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid bank_account_id"})
	}

	count, err := h.svc.ClassifyAllPending(c.Context(), tenantID, bankAccountID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"classified_count": count})
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

// ClearAll deletes ALL transactional data for the current tenant.
// POST /api/v1/admin/clear-transactional-data?confirm=true
// Preserves master data (accounts, parties, classification rules, bank accounts, etc.).
// Requires confirm=true query parameter to prevent accidental invocation.
func (h *BankTransactionHandler) ClearAll(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	if c.Query("confirm") != "true" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "requires confirm=true query parameter to proceed",
		})
	}

	result, err := h.svc.ClearTransactionalData(c.Context(), tenantID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message":   "transactional data cleared",
		"tenant_id": tenantID.String(),
		"deleted":   result,
	})
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
		"2006年07月02日",
	}
	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fiber.ErrBadRequest
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}


