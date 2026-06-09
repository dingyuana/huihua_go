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
