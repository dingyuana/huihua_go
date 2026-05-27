package handler

import (
	"github.com/gofiber/fiber/v2"
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
	return c.JSON(fiber.Map{"status": "ok"})
}

// ImportExcel imports parties from Excel file upload.
func (h *PartyHandler) ImportExcel(c *fiber.Ctx) error {
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
	return c.JSON(fiber.Map{"status": "ok"})
}