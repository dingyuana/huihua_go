package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"huihua/finance/internal/service"
)

type AgingHandler struct {
	svc *service.AgingService
}

func NewAgingHandler(svc *service.AgingService) *AgingHandler {
	return &AgingHandler{svc: svc}
}

func (h *AgingHandler) AR(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	asOf, _ := time.Parse("2006-01-02", c.Query("as_of"))
	if asOf.IsZero() {
		asOf = time.Now()
	}
	buckets, err := h.svc.ComputeAR(c.Context(), tenantID, asOf)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	summary := h.svc.Summarize(buckets, asOf)
	return c.JSON(fiber.Map{"buckets": buckets, "summary": summary})
}

func (h *AgingHandler) AP(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	asOf, _ := time.Parse("2006-01-02", c.Query("as_of"))
	if asOf.IsZero() {
		asOf = time.Now()
	}
	buckets, err := h.svc.ComputeAP(c.Context(), tenantID, asOf)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	summary := h.svc.Summarize(buckets, asOf)
	return c.JSON(fiber.Map{"buckets": buckets, "summary": summary})
}
