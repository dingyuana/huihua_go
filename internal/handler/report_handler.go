package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"huihua/finance/internal/service"
)

// ReportHandler handles financial report requests.
type ReportHandler struct {
	svc *service.ReportService
}

// NewReportHandler creates a new ReportHandler.
func NewReportHandler(svc *service.ReportService) *ReportHandler {
	return &ReportHandler{svc: svc}
}

// GetTrialBalance handles GET /api/v1/reports/trial-balance?period_no=X
func (h *ReportHandler) GetTrialBalance(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	periodNo := c.QueryInt("period_no", 0)
	if periodNo <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "period_no is required"})
	}

	report, err := h.svc.GetTrialBalance(c.Context(), tenantID, periodNo)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(report)
}

// GetIncomeStatement handles GET /api/v1/reports/income-statement?period_no=X
func (h *ReportHandler) GetIncomeStatement(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	periodNo := c.QueryInt("period_no", 0)
	if periodNo <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "period_no is required"})
	}

	report, err := h.svc.GetIncomeStatement(c.Context(), tenantID, periodNo)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(report)
}

// GetCashFlowStatement handles GET /api/v1/reports/cash-flow?period_no=X
func (h *ReportHandler) GetCashFlowStatement(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	periodNo := c.QueryInt("period_no", 0)
	if periodNo <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "period_no is required"})
	}

	report, err := h.svc.GetCashFlowStatement(c.Context(), tenantID, periodNo)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(report)
}

// GetBalanceSheet handles GET /api/v1/reports/balance-sheet?period_no=X
func (h *ReportHandler) GetBalanceSheet(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	periodNo := c.QueryInt("period_no", 0)
	if periodNo <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "period_no is required"})
	}

	report, err := h.svc.GetBalanceSheet(c.Context(), tenantID, periodNo)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(report)
}