package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"huihua/finance/internal/config"
	"huihua/finance/internal/service"
)

// ClearDataHandler exposes the two destructive clear-data endpoints.
// Gated by app.mode != "production" so the buttons are inert in prod.
type ClearDataHandler struct {
	svc *service.ClearDataService
	cfg *config.Config
}

func NewClearDataHandler(svc *service.ClearDataService, cfg *config.Config) *ClearDataHandler {
	return &ClearDataHandler{svc: svc, cfg: cfg}
}

// devOnly refuses the call when running in production.
func (h *ClearDataHandler) devOnly(c *fiber.Ctx) error {
	if h.cfg.App.Mode == "production" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "clear-data endpoints are disabled in production mode",
		})
	}
	return nil
}

// ClearBusinessData handles POST /api/v1/setup/clear-business-data.
func (h *ClearDataHandler) ClearBusinessData(c *fiber.Ctx) error {
	if err := h.devOnly(c); err != nil {
		return err
	}
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	result, err := h.svc.ClearBusinessData(c.Context(), tenantID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": fiber.Map{
		"scope":  "business",
		"result": result,
		"total":  sumValues(result),
	}})
}

// ClearBasicInfo handles POST /api/v1/setup/clear-basic-info.
func (h *ClearDataHandler) ClearBasicInfo(c *fiber.Ctx) error {
	if err := h.devOnly(c); err != nil {
		return err
	}
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	result, err := h.svc.ClearBasicInfo(c.Context(), tenantID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": fiber.Map{
		"scope":  "basic",
		"result": result,
		"total":  sumValues(result),
	}})
}

func sumValues(m service.ClearResult) int64 {
	var total int64
	for _, v := range m {
		total += v
	}
	return total
}
