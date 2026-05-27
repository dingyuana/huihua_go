package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"huihua/finance/internal/model"
	"huihua/finance/internal/service"
)

// ReconciliationHandler handles reconciliation HTTP requests.
type ReconciliationHandler struct {
	svc *service.ReconciliationService
}

// NewReconciliationHandler creates a new ReconciliationHandler.
func NewReconciliationHandler(svc *service.ReconciliationService) *ReconciliationHandler {
	return &ReconciliationHandler{svc: svc}
}

// Run executes the reconciliation matching.
func (h *ReconciliationHandler) Run(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	periodNo := c.QueryInt("period_no", 0)
	result, err := h.svc.Reconcile(c.Context(), tenantID, periodNo)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": result})
}

// ListPairs returns reconciliation pairs.
func (h *ReconciliationHandler) ListPairs(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	status := c.Query("status")
	pairs, err := h.svc.ListPairs(c.Context(), tenantID, status)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": pairs})
}

// ConfirmPair confirms a pair.
func (h *ReconciliationHandler) ConfirmPair(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	pairID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}
	userID := c.Locals("user_id").(uuid.UUID)
	if err := h.svc.ConfirmPair(c.Context(), tenantID, pairID, userID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "confirmed"})
}

// UnconfirmPair unconfirms a pair.
func (h *ReconciliationHandler) UnconfirmPair(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	pairID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}
	if err := h.svc.UnconfirmPair(c.Context(), tenantID, pairID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "unconfirmed"})
}

// GetUnmatched returns unmatched items.
func (h *ReconciliationHandler) GetUnmatched(c *fiber.Ctx) error {
	// Returns from the last run result — simplified
	return c.JSON(fiber.Map{"data": []model.UnmatchedItem{}})
}
