package handler

import (
	"github.com/gofiber/fiber/v2"
	"huihua/finance/internal/model"
	"huihua/finance/internal/service"
)

type ExpenseInvoiceOcrHandler struct {
	svc *service.OcrService
}

func NewExpenseInvoiceOcrHandler(svc *service.OcrService) *ExpenseInvoiceOcrHandler {
	return &ExpenseInvoiceOcrHandler{svc: svc}
}

// Recognize handles POST /expense-invoices/ocr
// Accepts either a file upload (multipart/form-data) or a JSON body with file_url
func (h *ExpenseInvoiceOcrHandler) Recognize(c *fiber.Ctx) error {
	var req model.OcrInvoiceRequest

	// Try to parse JSON body first
	if err := c.BodyParser(&req); err != nil {
		// If JSON parse fails, might be multipart form
		file, err := c.FormFile("file")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "no file uploaded or invalid request",
			})
		}
		// For file upload, we use the filename as file_url
		req.FileURL = file.Filename
	}

	if req.FileURL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "file_url is required",
		})
	}

	result, err := h.svc.RecognizeInvoice(c.Context(), req.FileURL)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(result)
}
