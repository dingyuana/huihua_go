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
	if accounts == nil {
		accounts = []model.BankAccount{}
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
		BankAccountType   string `json:"bank_account_type"`
		IBAN              string `json:"iban"`
		SwiftCode         string `json:"swift_code"`
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
		BankAccountType:   strPtr(req.BankAccountType),
		IBAN:              strPtr(req.IBAN),
		SwiftCode:         strPtr(req.SwiftCode),
	}
	account, err := h.svc.Create(c.Context(), tenantID, bankAcc)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"data": account})
}

// GetByID returns a single bank account.
func (h *BankHandler) GetByID(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}
	account, err := h.svc.GetByID(c.Context(), tenantID, id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "not found"})
	}
	return c.JSON(fiber.Map{"data": account})
}

// Update updates a bank account.
func (h *BankHandler) Update(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}
	var req struct {
		BankName          string `json:"bank_name"`
		AccountNumber     string `json:"account_number"`
		ClearingAccountID string `json:"clearing_account_id"`
		Currency          string `json:"currency"`
		BankAccountType   string `json:"bank_account_type"`
		IBAN              string `json:"iban"`
		SwiftCode         string `json:"swift_code"`
		IsActive          *bool  `json:"is_active"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	bankAcc := &model.BankAccount{
		BankName:        req.BankName,
		AccountNumber:   req.AccountNumber,
		Currency:        req.Currency,
		BankAccountType: strPtr(req.BankAccountType),
		IBAN:            strPtr(req.IBAN),
		SwiftCode:       strPtr(req.SwiftCode),
		IsActive:        true,
	}
	if req.ClearingAccountID != "" {
		clearingID, _ := uuid.Parse(req.ClearingAccountID)
		bankAcc.ClearingAccountID = &clearingID
	}
	if req.IsActive != nil {
		bankAcc.IsActive = *req.IsActive
	}
	if err := h.svc.Update(c.Context(), tenantID, id, bankAcc); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "updated"})
}

// strPtr converts a string to a *string, returning nil for empty strings.
func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// Delete deletes a bank account.
func (h *BankHandler) Delete(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}
	if err := h.svc.Delete(c.Context(), tenantID, id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "deleted"})
}