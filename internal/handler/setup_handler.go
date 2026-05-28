package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"huihua/finance/internal/service"
)

// SetupHandler handles account setup wizard HTTP requests.
type SetupHandler struct {
	svc *service.SetupService
}

// NewSetupHandler creates a new SetupHandler.
func NewSetupHandler(svc *service.SetupService) *SetupHandler {
	return &SetupHandler{svc: svc}
}

// GetStatus returns the current setup status.
func (h *SetupHandler) GetStatus(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	status, err := h.svc.GetStatus(c.Context(), tenantID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": status})
}

// CreateCompany handles POST /account-setup/wizard.
func (h *SetupHandler) CreateCompany(c *fiber.Ctx) error {
	var req struct {
		CompanyName string `json:"company_name"`
		TenantID    string `json:"tenant_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	tenantID, _ := uuid.Parse(req.TenantID)
	settings, err := h.svc.CreateCompany(c.Context(), tenantID, service.CreateCompanyRequest{
		CompanyName: req.CompanyName,
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"data": settings})
}