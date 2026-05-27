package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/model"
	"huihua/finance/internal/service"
)

// ExchangeRateHandler handles exchange rate HTTP requests.
type ExchangeRateHandler struct {
	svc *service.ExchangeRateService
}

// NewExchangeRateHandler creates a new ExchangeRateHandler.
func NewExchangeRateHandler(svc *service.ExchangeRateService) *ExchangeRateHandler {
	return &ExchangeRateHandler{svc: svc}
}

// List returns all exchange rates.
func (h *ExchangeRateHandler) List(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	rates, err := h.svc.ListAll(c.Context(), tenantID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": rates})
}

// Create adds a new exchange rate.
func (h *ExchangeRateHandler) Create(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	var req model.ExchangeRateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	rate, err := h.svc.AddExchangeRate(c.Context(), tenantID, req)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"data": rate})
}

// GetByID returns a single exchange rate.
func (h *ExchangeRateHandler) GetByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	rates, err := h.svc.ListAll(c.Context(), tenantID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	for _, r := range rates {
		if r.ID == id {
			return c.JSON(fiber.Map{"data": r})
		}
	}
	return c.Status(404).JSON(fiber.Map{"error": "not found"})
}

// Delete removes an exchange rate.
func (h *ExchangeRateHandler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	if err := h.svc.Delete(c.Context(), tenantID, id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "deleted"})
}

// Convert performs currency conversion.
func (h *ExchangeRateHandler) Convert(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	from := c.Query("from")
	to := c.Query("to")
	amountStr := c.Query("amount")
	dateStr := c.Query("date")
	if from == "" || to == "" || amountStr == "" {
		return c.Status(400).JSON(fiber.Map{"error": "from, to, and amount are required"})
	}
	amount, err := decimal.NewFromString(amountStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid amount"})
	}
	date := time.Now()
	if dateStr != "" {
		date, _ = time.Parse("2006-01-02", dateStr)
	}
	result, err := h.svc.ConvertAmount(c.Context(), tenantID, amount, from, to, date)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": result})
}
