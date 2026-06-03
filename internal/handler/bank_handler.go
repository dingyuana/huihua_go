package handler

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
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
		IsCash            bool   `json:"is_cash"`
		Custodian         string `json:"custodian"`
		Location          string `json:"location"`
		OpeningBalance    string `json:"opening_balance"`
		OpeningDate       string `json:"opening_date"`
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
		IsCash:            req.IsCash,
		Custodian:         strPtr(req.Custodian),
		Location:          strPtr(req.Location),
	}
	if req.OpeningBalance != "" {
		amt, err := decimal.NewFromString(req.OpeningBalance)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid opening_balance"})
		}
		bankAcc.OpeningBalance = amt
		bankAcc.CurrentBalance = amt
	}
	if req.OpeningDate != "" {
		t, err := time.Parse("2006-01-02", req.OpeningDate)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid opening_date format, expected YYYY-MM-DD"})
		}
		bankAcc.OpeningDate = &t
	}
	if req.IsCash && req.AccountNumber == "" {
		bankAcc.AccountNumber = fmt.Sprintf("CASH-%s", uuid.New().String()[:8])
	}
	if req.IsCash && req.BankAccountType == "" {
		bankAcc.BankAccountType = strPtr("cash")
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
		IsCash            *bool  `json:"is_cash"`
		Custodian         string `json:"custodian"`
		Location          string `json:"location"`
		OpeningBalance    string `json:"opening_balance"`
		OpeningDate       string `json:"opening_date"`
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
		Custodian:       strPtr(req.Custodian),
		Location:        strPtr(req.Location),
	}
	if req.ClearingAccountID != "" {
		clearingID, _ := uuid.Parse(req.ClearingAccountID)
		bankAcc.ClearingAccountID = &clearingID
	}
	if req.IsActive != nil {
		bankAcc.IsActive = *req.IsActive
	}
	if req.IsCash != nil {
		bankAcc.IsCash = *req.IsCash
	}
	if req.OpeningBalance != "" {
		amt, err := decimal.NewFromString(req.OpeningBalance)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid opening_balance"})
		}
		bankAcc.OpeningBalance = amt
	}
	if req.OpeningDate != "" {
		t, err := time.Parse("2006-01-02", req.OpeningDate)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid opening_date format"})
		}
		bankAcc.OpeningDate = &t
	}
	if err := h.svc.Update(c.Context(), tenantID, id, bankAcc); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "updated"})
}

// AdjustBalance records a manual opening or current-balance adjustment with audit trail.
func (h *BankHandler) AdjustBalance(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	userIDRaw := c.Locals("user_id")
	userID, _ := userIDRaw.(uuid.UUID)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}
	var req struct {
		AdjustmentType string `json:"adjustment_type"`
		NewBalance     string `json:"new_balance"`
		Reason         string `json:"reason"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	if req.AdjustmentType == "" {
		req.AdjustmentType = "manual_adjust"
	}
	if req.AdjustmentType != "opening" && req.AdjustmentType != "manual_adjust" {
		return c.Status(400).JSON(fiber.Map{"error": "invalid adjustment_type (allowed: opening, manual_adjust)"})
	}
	newBal, err := decimal.NewFromString(req.NewBalance)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid new_balance"})
	}
	adj, err := h.svc.AdjustBalance(c.Context(), tenantID, id, req.AdjustmentType, newBal, req.Reason, userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"data": adj})
}

// ListBalanceAdjustments returns the audit trail of balance adjustments.
func (h *BankHandler) ListBalanceAdjustments(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}
	items, err := h.svc.ListBalanceAdjustments(c.Context(), tenantID, id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": items})
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
