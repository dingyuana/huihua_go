package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"huihua/finance/internal/middleware"
	"huihua/finance/internal/model"
	"huihua/finance/internal/service"
)

// ClassificationRuleHandler handles HTTP requests for classification rules.
type ClassificationRuleHandler struct {
	svc *service.ClassificationRuleService
}

// NewClassificationRuleHandler creates a new ClassificationRuleHandler.
func NewClassificationRuleHandler(svc *service.ClassificationRuleService) *ClassificationRuleHandler {
	return &ClassificationRuleHandler{svc: svc}
}

// List returns all classification rules for the tenant.
func (h *ClassificationRuleHandler) List(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	rules, err := h.svc.ListRules(c.Context(), tenantID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(rules)
}

// Create handles POST /api/v1/classification-rules.
func (h *ClassificationRuleHandler) Create(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var req model.CreateRuleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	rule, err := h.svc.CreateRule(c.Context(), tenantID, &req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(rule)
}

// Update handles PUT /api/v1/classification-rules/:id.
func (h *ClassificationRuleHandler) Update(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid rule id"})
	}

	var req model.UpdateRuleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if err := h.svc.UpdateRule(c.Context(), tenantID, id, &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "updated"})
}

// Delete handles DELETE /api/v1/classification-rules/:id.
func (h *ClassificationRuleHandler) Delete(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid rule id"})
	}

	if err := h.svc.DeleteRule(c.Context(), tenantID, id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "deleted"})
}

// Reorder handles POST /api/v1/classification-rules/reorder.
func (h *ClassificationRuleHandler) Reorder(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var req model.ReorderPriorityRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	ruleIDs := make([]uuid.UUID, 0, len(req.RuleIDs))
	for _, idStr := range req.RuleIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid rule id: " + idStr})
		}
		ruleIDs = append(ruleIDs, id)
	}

	if err := h.svc.ReorderPriority(c.Context(), tenantID, ruleIDs); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "reordered"})
}

// Match handles POST /api/v1/classification-rules/match.
func (h *ClassificationRuleHandler) Match(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var req model.RuleMatchRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	result, err := h.svc.MatchTransaction(c.Context(), tenantID, req.Keywords, req.Amount, req.Direction)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(result)
}