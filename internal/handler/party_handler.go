package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"huihua/finance/internal/model"
	"huihua/finance/internal/service"
)

// PartyHandler handles party HTTP requests.
type PartyHandler struct {
	svc *service.PartyService
}

// NewPartyHandler creates a new PartyHandler.
func NewPartyHandler(svc *service.PartyService) *PartyHandler {
	return &PartyHandler{svc: svc}
}

// List returns parties filtered by type.
func (h *PartyHandler) List(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	partyType := c.Query("type", "")
	parties, err := h.svc.List(c.Context(), tenantID, partyType)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": parties})
}

// ImportExcel imports parties from Excel file upload.
func (h *PartyHandler) ImportExcel(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "file required"})
	}
	f, err := file.Open()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer f.Close()
	data := make([]byte, file.Size)
	f.Read(data)
	result, err := h.svc.ImportFromExcel(c.Context(), tenantID, data)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": result})
}

// GetByID returns a single party.
func (h *PartyHandler) GetByID(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}
	party, err := h.svc.GetByID(c.Context(), tenantID, id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "not found"})
	}
	return c.JSON(fiber.Map{"data": party})
}

// Create creates a new party.
func (h *PartyHandler) Create(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	var req struct {
		Name        string `json:"name"`
		PartyType   string `json:"party_type"`
		TaxNumber   string `json:"tax_number,omitempty"`
		ContactName string `json:"contact_name,omitempty"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	taxNum := req.TaxNumber
	contact := req.ContactName
	party := &model.Party{
		Name:        req.Name,
		PartyType:   req.PartyType,
		TaxNumber:   &taxNum,
		ContactName: &contact,
	}
	result, err := h.svc.CreateParty(c.Context(), tenantID, party)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"data": result})
}

// Update updates a party.
func (h *PartyHandler) Update(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}
	var req struct {
		Name        string `json:"name"`
		PartyType   string `json:"party_type"`
		TaxNumber   string `json:"tax_number,omitempty"`
		ContactName string `json:"contact_name,omitempty"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	taxNum := req.TaxNumber
	contact := req.ContactName
	party := &model.Party{
		Name:        req.Name,
		PartyType:   req.PartyType,
		TaxNumber:   &taxNum,
		ContactName: &contact,
	}
	if err := h.svc.UpdateParty(c.Context(), tenantID, id, party); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "updated"})
}

// Delete deletes a party.
func (h *PartyHandler) Delete(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}
	if err := h.svc.DeleteParty(c.Context(), tenantID, id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "deleted"})
}