package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"huihua/finance/internal/service"
)

type CreditControlHandler struct {
	svc *service.CreditControlService
}

func NewCreditControlHandler(svc *service.CreditControlService) *CreditControlHandler {
	return &CreditControlHandler{svc: svc}
}

func (h *CreditControlHandler) GetStatus(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	partyID, err := uuid.Parse(c.Query("party_id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "party_id required"})
	}
	status, err := h.svc.GetStatus(c.Context(), tenantID, partyID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": status})
}

func (h *CreditControlHandler) ListOverLimit(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	list, err := h.svc.ListOverLimit(c.Context(), tenantID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"list": list, "total": len(list)})
}
