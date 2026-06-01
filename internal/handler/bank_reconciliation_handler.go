package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/service"
)

// BankReconciliationHandler handles bank reconciliation HTTP requests.
type BankReconciliationHandler struct {
	svc *service.BankReconciliationService
}

// NewBankReconciliationHandler creates a new BankReconciliationHandler.
func NewBankReconciliationHandler(svc *service.BankReconciliationService) *BankReconciliationHandler {
	return &BankReconciliationHandler{svc: svc}
}

// ReconcileRequest represents the request body for reconcile endpoint.
type ReconcileRequest struct {
	BankAccountID string `json:"bank_account_id"`
	PeriodNo      int    `json:"period_no"`
}

// Reconcile performs bank reconciliation for a bank account in a period.
func (h *BankReconciliationHandler) Reconcile(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	var req ReconcileRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	bankAccountID, err := uuid.Parse(req.BankAccountID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid bank_account_id",
		})
	}

	result, err := h.svc.ReconcileBankAccount(c.Context(), tenantID, bankAccountID, req.PeriodNo)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Save the result
	err = h.svc.SaveReconciliationResult(c.Context(), tenantID, bankAccountID, req.PeriodNo, result)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to save reconciliation result: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"bank_balance":     result.BankBalance,
		"book_balance":     result.BookBalance,
		"adjusted_balance": result.AdjustedBalance,
		"bank_only_items":  result.BankOnlyItems,
		"book_only_items":  result.BookOnlyItems,
		"matched_count":    result.MatchedCount,
		"total_matched":    result.TotalMatched,
	})
}

// GetReport returns the reconciliation report for a bank account in a period.
func (h *BankReconciliationHandler) GetReport(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	bankAccountIDStr := c.Query("bank_account_id")
	periodNoStr := c.Query("period_no")

	if bankAccountIDStr == "" || periodNoStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "bank_account_id and period_no are required",
		})
	}

	bankAccountID, err := uuid.Parse(bankAccountIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid bank_account_id",
		})
	}

	var periodNo int
	var pErr error
	if periodNo, pErr = parsePeriodNo(periodNoStr); pErr != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid period_no",
		})
	}

	report, err := h.svc.GetReconciliationReport(c.Context(), tenantID, bankAccountID, periodNo)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(report)
}

// MarkDoneRequest represents the request body for mark-done endpoint.
type MarkDoneRequest struct {
	BankAccountID string `json:"bank_account_id"`
	PeriodNo      int    `json:"period_no"`
}

// MarkDone marks a reconciliation as completed.
func (h *BankReconciliationHandler) MarkDone(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	userID := c.Locals("user_id").(uuid.UUID)

	var req MarkDoneRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	bankAccountID, err := uuid.Parse(req.BankAccountID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid bank_account_id",
		})
	}

	err = h.svc.MarkAsReconciled(c.Context(), tenantID, bankAccountID, req.PeriodNo, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "reconciliation marked as done",
	})
}

// GetStatus returns the reconciliation status for a bank account in a period.
func (h *BankReconciliationHandler) GetStatus(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	bankAccountIDStr := c.Query("bank_account_id")
	periodNoStr := c.Query("period_no")

	if bankAccountIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "bank_account_id is required",
		})
	}

	bankAccountID, err := uuid.Parse(bankAccountIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid bank_account_id",
		})
	}

	var periodNo int
	if periodNoStr != "" {
		periodNo, err = parsePeriodNo(periodNoStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid period_no",
			})
		}
	}

	status, err := h.svc.GetReconciliationStatus(c.Context(), tenantID, bankAccountID, periodNo)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"bank_account_id": bankAccountID,
		"period_no":       periodNo,
		"status":          status,
	})
}

// parsePeriodNo parses a period string (YYYYMM) to int.
func parsePeriodNo(s string) (int, error) {
	var result int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + int(c-'0')
		} else {
			return 0, fiber.ErrBadRequest
		}
	}
	return result, nil
}

func (h *BankReconciliationHandler) BalanceCheck(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	result, err := h.svc.BalanceCheck(c.Context(), tenantID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": result})
}

type DiffReportItem struct {
	SourceType  string `json:"source_type"`
	TxnDate     string `json:"txn_date"`
	Description string `json:"description"`
	Amount      string `json:"amount"`
	Direction   string `json:"direction"`
	Reference   string `json:"reference,omitempty"`
	Reason      string `json:"reason"`
}

type DiffReport struct {
	BankAccountID    uuid.UUID        `json:"bank_account_id"`
	BankName         string           `json:"bank_name"`
	PeriodNo         int              `json:"period_no"`
	BankBalance      string           `json:"bank_balance"`
	BookBalance      string           `json:"book_balance"`
	Difference       string           `json:"difference"`
	BankOnlyItems    []DiffReportItem `json:"bank_only_items"`
	BookOnlyItems    []DiffReportItem `json:"book_only_items"`
	BankOnlyTotal    string           `json:"bank_only_total"`
	BookOnlyTotal    string           `json:"book_only_total"`
	BankOnlyCount    int              `json:"bank_only_count"`
	BookOnlyCount    int              `json:"book_only_count"`
	AdjustedReconciled bool           `json:"adjusted_reconciled"`
	GeneratedAt      string           `json:"generated_at"`
}

func (h *BankReconciliationHandler) GetDiffReport(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	bankAccountIDStr := c.Query("bank_account_id")
	periodNoStr := c.Query("period_no")
	if bankAccountIDStr == "" || periodNoStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "bank_account_id and period_no are required"})
	}
	bankAccountID, err := uuid.Parse(bankAccountIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid bank_account_id"})
	}
	periodNo, err := parsePeriodNo(periodNoStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid period_no"})
	}

	result, err := h.svc.ReconcileBankAccount(c.Context(), tenantID, bankAccountID, periodNo)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	bankAcct, err := h.svc.GetBankAccountForReport(c.Context(), tenantID, bankAccountID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	bankOnlyItems := toDiffItems(result.BankOnlyItems)
	bookOnlyItems := toDiffItems(result.BookOnlyItems)
	diff := result.BankBalance.Sub(result.BookBalance)
	adjusted := diff.IsZero()

	report := DiffReport{
		BankAccountID:      bankAccountID,
		BankName:           bankAcct,
		PeriodNo:           periodNo,
		BankBalance:        result.BankBalance.String(),
		BookBalance:        result.BookBalance.String(),
		Difference:         diff.String(),
		BankOnlyItems:      bankOnlyItems,
		BookOnlyItems:      bookOnlyItems,
		BankOnlyTotal:      sumAmounts(result.BankOnlyItems).String(),
		BookOnlyTotal:      sumAmounts(result.BookOnlyItems).String(),
		BankOnlyCount:      len(bankOnlyItems),
		BookOnlyCount:      len(bookOnlyItems),
		AdjustedReconciled: adjusted,
		GeneratedAt:        time.Now().Format("2006-01-02 15:04:05"),
	}
	return c.JSON(fiber.Map{"data": report})
}

func toDiffItems(items []service.UnreconciledItem) []DiffReportItem {
	out := make([]DiffReportItem, 0, len(items))
	for _, it := range items {
		out = append(out, DiffReportItem{
			SourceType:  it.SourceType,
			TxnDate:     it.TxnDate.Format("2006-01-02"),
			Description: it.Description,
			Amount:      it.Amount.String(),
			Direction:   it.Direction,
		})
	}
	return out
}

func sumAmounts(items []service.UnreconciledItem) decimal.Decimal {
	total := decimal.Zero
	for _, it := range items {
		total = total.Add(it.Amount)
	}
	return total
}