package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"huihua/finance/internal/service"
)

// OpeningBalanceHandler handles opening balance HTTP requests.
type OpeningBalanceHandler struct {
	svc *service.OpeningBalanceService
}

// NewOpeningBalanceHandler creates a new OpeningBalanceHandler.
func NewOpeningBalanceHandler(svc *service.OpeningBalanceService) *OpeningBalanceHandler {
	return &OpeningBalanceHandler{svc: svc}
}

// Import handles Excel import of opening balances.
// POST /api/v1/opening-balances/import
func (h *OpeningBalanceHandler) Import(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	var req struct {
		CompanyID string `json:"company_id"`
		PeriodNo  int    `json:"period_no"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.CompanyID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "company_id is required",
		})
	}

	companyID, err := uuid.Parse(req.CompanyID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid company_id",
		})
	}

	if req.PeriodNo == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "period_no is required",
		})
	}

	// Get Excel file from form
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "file is required",
		})
	}

	f, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "cannot open file",
		})
	}
	defer f.Close()

	data := make([]byte, file.Size)
	if _, err := f.Read(data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "cannot read file",
		})
	}

	entries, err := h.svc.ImportFromExcel(c.Context(), tenantID, companyID, req.PeriodNo, data)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"imported": len(entries),
		"entries":  entries,
	})
}

// List returns opening balances for a period.
// GET /api/v1/opening-balances?period_no=xxx
func (h *OpeningBalanceHandler) List(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	periodNoStr := c.Query("period_no")
	if periodNoStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "period_no is required",
		})
	}

	var periodNo int
	if _, err := fmt.Sscanf(periodNoStr, "%d", &periodNo); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid period_no",
		})
	}

	entries, err := h.svc.GetByPeriod(c.Context(), tenantID, periodNo)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"period_no": periodNo,
		"entries":   entries,
		"total":     len(entries),
	})
}

// GetTrialBalance returns the trial balance for a period.
// GET /api/v1/opening-balances/trial-balance?period_no=xxx
func (h *OpeningBalanceHandler) GetTrialBalance(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	periodNoStr := c.Query("period_no")
	if periodNoStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "period_no is required",
		})
	}

	var periodNo int
	if _, err := fmt.Sscanf(periodNoStr, "%d", &periodNo); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid period_no",
		})
	}

	trialBalance, err := h.svc.GetTrialBalance(c.Context(), tenantID, periodNo)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(trialBalance)
}

// Validate validates opening balances for a period.
// POST /api/v1/opening-balances/validate
func (h *OpeningBalanceHandler) Validate(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	var req struct {
		PeriodNo int `json:"period_no"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.PeriodNo == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "period_no is required",
		})
	}

	result, err := h.svc.ValidateOpeningBalance(c.Context(), tenantID, req.PeriodNo)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(result)
}

// GetByAccount returns opening balance for a specific account.
// GET /api/v1/opening-balances/:account_id?period_no=xxx
func (h *OpeningBalanceHandler) GetByAccount(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	accountIDStr := c.Params("account_id")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid account_id",
		})
	}

	periodNoStr := c.Query("period_no")
	if periodNoStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "period_no is required",
		})
	}

	var periodNo int
	if _, err := fmt.Sscanf(periodNoStr, "%d", &periodNo); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid period_no",
		})
	}

	entry, err := h.svc.GetByAccount(c.Context(), tenantID, accountID, periodNo)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "opening balance not found",
		})
	}

	return c.JSON(entry)
}