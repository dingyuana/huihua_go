package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"huihua/finance/internal/service"
)

// PeriodHandler handles period operations.
type PeriodHandler struct {
	svc *service.PeriodService
}

// NewPeriodHandler creates a new PeriodHandler.
func NewPeriodHandler(svc *service.PeriodService) *PeriodHandler {
	return &PeriodHandler{svc: svc}
}

// ClosePeriodRequest is the request body for closing a period.
type ClosePeriodRequest struct {
	PeriodNo   int  `json:"period_no"`
	UserID     string `json:"user_id"`
	UserName   string `json:"user_name"`
	GenerateClosingEntries bool `json:"generate_closing_entries"`
}

// Close handles POST /api/v1/periods/:period_no/close
func (h *PeriodHandler) Close(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	var req ClosePeriodRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.PeriodNo <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "period_no is required"})
	}

	if req.UserID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "user_id is required"})
	}

	if err := h.svc.ClosePeriod(c.Context(), tenantID, &service.ClosePeriodRequest{
		PeriodNo:   req.PeriodNo,
		UserID:     req.UserID,
		UserName:   req.UserName,
		GenerateClosingEntries: req.GenerateClosingEntries,
	}); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "period closed successfully"})
}

// List handles GET /api/v1/periods
func (h *PeriodHandler) List(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	periods, err := h.svc.ListPeriods(c.Context(), tenantID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"periods": periods})
}

// GetCurrent handles GET /api/v1/periods/current
func (h *PeriodHandler) GetCurrent(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	period, err := h.svc.GetCurrentPeriod(c.Context(), tenantID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"period": period})
}

// Unclose handles POST /api/v1/periods/:period_no/unclose
func (h *PeriodHandler) Unclose(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	periodNoStr := c.Params("period_no")
	periodNo, err := strconv.Atoi(periodNoStr)
	if err != nil || periodNo <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid period_no"})
	}

	if err := h.svc.UnclosePeriod(c.Context(), tenantID, periodNo); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "period unclosed successfully"})
}

// PreCloseCheck handles GET /api/v1/periods/pre-close-check?year=2026&month=5
func (h *PeriodHandler) PreCloseCheck(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	yearStr := c.Query("year")
	monthStr := c.Query("month")

	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2000 || year > 2100 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid year"})
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid month"})
	}

	result, err := h.svc.PreCloseCheck(c.Context(), tenantID, year, month)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(result)
}

// VoucherGaps handles GET /api/v1/periods/voucher-gaps?year=2026&month=5
func (h *PeriodHandler) VoucherGaps(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	yearStr := c.Query("year")
	monthStr := c.Query("month")

	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2000 || year > 2100 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid year"})
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid month"})
	}

	gaps, err := h.svc.ScanVoucherGaps(c.Context(), tenantID, year, month)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if gaps == nil {
		gaps = []service.VoucherGap{}
	}

	// Count gaps by type
	missingCount := 0
	voidedCount := 0
	for _, g := range gaps {
		if g.GapType == "missing" {
			missingCount++
		} else if g.GapType == "voided" {
			voidedCount++
		}
	}

	return c.JSON(fiber.Map{
		"voucher_gaps":  gaps,
		"missing_count": missingCount,
		"voided_count":  voidedCount,
		"total_gaps":    len(gaps),
		"has_missing":   missingCount > 0,
	})
}

// CloseCheckSummary handles GET /api/v1/periods/close-check-summary?year=2026&month=5
func (h *PeriodHandler) CloseCheckSummary(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	yearStr := c.Query("year")
	monthStr := c.Query("month")

	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2000 || year > 2100 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid year"})
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid month"})
	}

	result, err := h.svc.GetCloseCheckSummary(c.Context(), tenantID, year, month)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(result)
}