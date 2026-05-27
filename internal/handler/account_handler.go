package handler

import (
	"github.com/gofiber/fiber/v2"
	"huihua/finance/internal/service"
)

// AccountHandler handles account HTTP requests.
type AccountHandler struct {
	svc *service.AccountService
}

// NewAccountHandler creates a new AccountHandler.
func NewAccountHandler(svc *service.AccountService) *AccountHandler {
	return &AccountHandler{svc: svc}
}

// GetTree returns the full account tree.
func (h *AccountHandler) GetTree(c *fiber.Ctx) error {
	_ = c.Locals("tenant_id") // tenant_id from middleware
	return c.JSON(fiber.Map{"status": "ok"})
}

// InitFromSeed initializes accounts from the standard seed.
func (h *AccountHandler) InitFromSeed(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}