package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"huihua/finance/internal/model"
	"huihua/finance/internal/service"
)

type ExpenseInvoiceImportHandler struct {
	svc *service.ExpenseInvoiceImportService
}

func NewExpenseInvoiceImportHandler(svc *service.ExpenseInvoiceImportService) *ExpenseInvoiceImportHandler {
	return &ExpenseInvoiceImportHandler{svc: svc}
}

// Upload handles POST /expense-invoices/import/upload
func (h *ExpenseInvoiceImportHandler) Upload(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "no file uploaded",
		})
	}

	f, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to open file",
		})
	}
	defer f.Close()

	result, err := h.svc.Upload(c.Context(), tenantID, f, file.Filename)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(result)
}

// Preview handles GET /expense-invoices/import/:batch_id/preview
func (h *ExpenseInvoiceImportHandler) Preview(c *fiber.Ctx) error {
	batchID := c.Params("batch_id")
	if batchID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "batch_id is required",
		})
	}

	batch, err := h.svc.Preview(c.Context(), batchID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(batch)
}

// Confirm handles POST /expense-invoices/import/:batch_id/confirm
func (h *ExpenseInvoiceImportHandler) Confirm(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	batchID := c.Params("batch_id")
	if batchID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "batch_id is required",
		})
	}

	var req model.ImportConfirmRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}
	req.BatchID = batchID

	result, err := h.svc.Confirm(c.Context(), tenantID, &req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(result)
}
