package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"huihua/finance/internal/model"
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
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	tree, err := h.svc.GetTree(c.Context(), tenantID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if tree == nil {
		tree = []model.Account{}
	}
	return c.JSON(fiber.Map{"data": tree})
}

// List returns accounts for the current tenant with optional pagination.
// Query params: limit (default 100, max 1000), offset (default 0), parent_id (optional filter).
// Account code is searchable via `?code=1001` exact match.
func (h *AccountHandler) List(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	limit := c.QueryInt("limit", 100)
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	offset := c.QueryInt("offset", 0)
	if offset < 0 {
		offset = 0
	}
	code := c.Query("code")
	accounts, total, err := h.svc.List(c.Context(), tenantID, limit, offset, code)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{
		"data":  accounts,
		"total": total,
		"page":  offset/limit + 1,
		"page_size": limit,
	})
}

// InitFromSeed initializes accounts from the standard seed.
func (h *AccountHandler) InitFromSeed(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	var req struct {
		CompanyID string `json:"company_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	companyID, err := uuid.Parse(req.CompanyID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid company_id"})
	}
	if err := h.svc.InitFromSeed(c.Context(), tenantID, companyID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "seed initialized"})
}

// Create handles POST /accounts.
// Body: { code, name, account_type, root_type, parent_id, is_group, company_id, currency }
// The service auto-computes lft/rgt/level/path and enforces code uniqueness.
func (h *AccountHandler) Create(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	var req struct {
		Code        string  `json:"code"`
		Name        string  `json:"name"`
		AccountType *string `json:"account_type"`
		RootType    *string `json:"root_type"`
		ParentID    *string `json:"parent_id"`
		IsGroup     bool    `json:"is_group"`
		CompanyID   string  `json:"company_id"`
		Currency    string  `json:"currency"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	if req.Code == "" || req.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "code and name are required"})
	}
	if req.CompanyID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "company_id is required"})
	}
	companyID, err := uuid.Parse(req.CompanyID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid company_id"})
	}

	var parentID *uuid.UUID
	if req.ParentID != nil && *req.ParentID != "" {
		pid, err := uuid.Parse(*req.ParentID)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid parent_id"})
		}
		parentID = &pid
	}

	currency := req.Currency
	if currency == "" {
		currency = "CNY"
	}

	account := &model.Account{
		Code:        req.Code,
		Name:        req.Name,
		AccountType: req.AccountType,
		RootType:    req.RootType,
		ParentID:    parentID,
		IsGroup:     req.IsGroup,
		CompanyID:   companyID,
		Currency:    currency,
	}

	created, err := h.svc.Create(c.Context(), tenantID, account)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": created})
}

// GetByID handles GET /accounts/:id.
func (h *AccountHandler) GetByID(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}
	account, err := h.svc.GetByID(c.Context(), tenantID, id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": account})
}

// Update handles PUT /accounts/:id.
// Only editable fields are accepted: name, account_type, root_type, is_group, currency, is_active.
// Code, parent_id, and tree-structure fields are preserved by the service.
func (h *AccountHandler) Update(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}

	var req struct {
		Name        *string `json:"name"`
		AccountType *string `json:"account_type"`
		RootType    *string `json:"root_type"`
		IsGroup     *bool   `json:"is_group"`
		Currency    *string `json:"currency"`
		IsActive    *bool   `json:"is_active"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Load existing to preserve un-supplied fields
	existing, err := h.svc.GetByID(c.Context(), tenantID, id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}
	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.AccountType != nil {
		existing.AccountType = req.AccountType
	}
	if req.RootType != nil {
		existing.RootType = req.RootType
	}
	if req.IsGroup != nil {
		existing.IsGroup = *req.IsGroup
	}
	if req.Currency != nil {
		existing.Currency = *req.Currency
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	if err := h.svc.Update(c.Context(), tenantID, existing); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": existing})
}

// Delete handles DELETE /accounts/:id.
// The service rejects deletion when child accounts exist.
func (h *AccountHandler) Delete(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}
	if err := h.svc.Delete(c.Context(), tenantID, id); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "deleted"})
}

// AutoCode handles POST /accounts/auto-code.
// Body: { parent_id }. Returns { suggested_code } based on existing siblings.
func (h *AccountHandler) AutoCode(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	var req struct {
		ParentID string `json:"parent_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	parentID, err := uuid.Parse(req.ParentID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid parent_id"})
	}
	suggested, err := h.svc.AutoCode(c.Context(), tenantID, parentID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"suggested_code": suggested})
}

// LedgerOnly handles GET /accounts/ledger-only.
// Returns accounts where is_group = false AND is_active = true (postable accounts).
func (h *AccountHandler) LedgerOnly(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	accounts, err := h.svc.ListLedgerOnly(c.Context(), tenantID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if accounts == nil {
		accounts = []model.Account{}
	}
	return c.JSON(fiber.Map{"data": accounts})
}
