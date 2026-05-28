package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"huihua/finance/internal/model"
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
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	accounts, err := h.svc.List(c.Context(), tenantID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": accounts})
}

// Create handles POST /bank-accounts.
func (h *BankHandler) Create(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	var req struct {
		BankName          string `json:"bank_name"`
		AccountNumber     string `json:"account_number"`
		ClearingAccountID string `json:"clearing_account_id"`
		Currency          string `json:"currency"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	clearingID, _ := uuid.Parse(req.ClearingAccountID)
	bankAcc := &model.BankAccount{
		BankName:          req.BankName,
		AccountNumber:     req.AccountNumber,
		ClearingAccountID: &clearingID,
		Currency:          req.Currency,
	}
	account, err := h.svc.Create(c.Context(), tenantID, bankAcc)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"data": account})
}