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