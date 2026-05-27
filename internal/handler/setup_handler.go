package handler

import (
	"github.com/gofiber/fiber/v2"
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
	return c.JSON(fiber.Map{"status": "ok"})
}

// CreateCompany handles POST /account-setup/wizard.
func (h *SetupHandler) CreateCompany(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}