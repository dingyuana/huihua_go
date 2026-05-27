package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"huihua/finance/internal/model"
	"huihua/finance/internal/service"
)

// VoucherTemplateHandler handles voucher template HTTP requests.
type VoucherTemplateHandler struct {
	svc *service.VoucherTemplateService
}

// NewVoucherTemplateHandler creates a new VoucherTemplateHandler.
func NewVoucherTemplateHandler(svc *service.VoucherTemplateService) *VoucherTemplateHandler {
	return &VoucherTemplateHandler{svc: svc}
}

// getTenantID extracts tenant_id from context.
func (h *VoucherTemplateHandler) getTenantID(c *fiber.Ctx) uuid.UUID {
	tenantID, _ := c.Locals("tenant_id").(uuid.UUID)
	return tenantID
}

// List returns all voucher templates.
func (h *VoucherTemplateHandler) List(c *fiber.Ctx) error {
	tenantID := h.getTenantID(c)
	templates, err := h.svc.ListTemplates(c.Context(), tenantID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"data": templates,
	})
}

// Create creates a new voucher template.
func (h *VoucherTemplateHandler) Create(c *fiber.Ctx) error {
	tenantID := h.getTenantID(c)

	var req model.CreateTemplateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body: " + err.Error(),
		})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "name is required",
		})
	}

	template, err := h.svc.CreateTemplate(c.Context(), tenantID, &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": template,
	})
}

// GetByID returns a voucher template with its lines.
func (h *VoucherTemplateHandler) GetByID(c *fiber.Ctx) error {
	tenantID := h.getTenantID(c)
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid id",
		})
	}

	template, err := h.svc.GetTemplateByID(c.Context(), tenantID, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if template == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "template not found",
		})
	}
	return c.JSON(fiber.Map{
		"data": template,
	})
}

// Update updates a voucher template.
func (h *VoucherTemplateHandler) Update(c *fiber.Ctx) error {
	tenantID := h.getTenantID(c)
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid id",
		})
	}

	var req model.UpdateTemplateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body: " + err.Error(),
		})
	}

	if err := h.svc.UpdateTemplate(c.Context(), tenantID, id, &req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"status": "ok",
	})
}

// Delete performs a soft delete on a voucher template.
func (h *VoucherTemplateHandler) Delete(c *fiber.Ctx) error {
	tenantID := h.getTenantID(c)
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid id",
		})
	}

	if err := h.svc.DeleteTemplate(c.Context(), tenantID, id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"status": "deleted",
	})
}

// GetNumberingRule returns the numbering rule for the tenant.
func (h *VoucherTemplateHandler) GetNumberingRule(c *fiber.Ctx) error {
	tenantID := h.getTenantID(c)
	rule, err := h.svc.GetOrCreateNumberingRule(c.Context(), tenantID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"data": rule,
	})
}

// UpdateNumberingRule creates or updates the numbering rule.
func (h *VoucherTemplateHandler) UpdateNumberingRule(c *fiber.Ctx) error {
	tenantID := h.getTenantID(c)

	var req model.NumberingRuleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body: " + err.Error(),
		})
	}

	if req.Prefix == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "prefix is required",
		})
	}

	rule, err := h.svc.UpdateNumberingRule(c.Context(), tenantID, &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"data": rule,
	})
}

// GenerateNextNumber generates the next voucher number.
func (h *VoucherTemplateHandler) GenerateNextNumber(c *fiber.Ctx) error {
	tenantID := h.getTenantID(c)
	resp, err := h.svc.GenerateVoucherNumber(c.Context(), tenantID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"data": resp,
	})
}
