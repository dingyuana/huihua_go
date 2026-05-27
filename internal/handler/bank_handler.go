package handler

import (
	"github.com/gofiber/fiber/v2"
	"huihua/finance/internal/service"
)

// BankHandler handles bank account HTTP requests.
type BankHandler struct {
	svc *service.BankService
}

// NewBankHandler creates a new BankHandler.
func NewBankHandler(svc *service.BankService) *BankHandler {
	return &BankHandler{svc: svc}
}

// List returns all bank accounts.
func (h *BankHandler) List(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}

// Create handles POST /bank-accounts.
func (h *BankHandler) Create(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}