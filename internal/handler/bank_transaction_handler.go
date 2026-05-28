package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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