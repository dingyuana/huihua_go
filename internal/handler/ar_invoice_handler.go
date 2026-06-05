package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"huihua/finance/internal/repository"
)

type ArInvoiceHandler struct {
	repo *repository.ArInvoiceRepository
}

func NewArInvoiceHandler(repo *repository.ArInvoiceRepository) *ArInvoiceHandler {
	return &ArInvoiceHandler{repo: repo}
}

func (h *ArInvoiceHandler) List(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	var status *string
	if v := c.Query("status"); v != "" {
		status = &v
	}

	ars, err := h.repo.ListByTenant(c.Context(), tenantID, status)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	result := make([]map[string]interface{}, len(ars))
	for i, ar := range ars {
		var dueDate, confirmedAt, approvedAt string
		if ar.DueDate != nil {
			dueDate = ar.DueDate.Format("2006-01-02")
		}
		if ar.ConfirmedAt != nil {
			confirmedAt = ar.ConfirmedAt.Format("2006-01-02 15:04:05")
		}
		if ar.ApprovedAt != nil {
			approvedAt = ar.ApprovedAt.Format("2006-01-02 15:04:05")
		}
		result[i] = map[string]interface{}{
			"id":                ar.ID,
			"invoice_id":        ar.InvoiceID,
			"invoice_no":        ar.InvoiceNo,
			"customer_id":       ar.CustomerID,
			"amount":            ar.Amount.String(),
			"due_date":          dueDate,
			"status":            ar.Status,
			"source_type":       ar.SourceType,
			"created_at":        ar.CreatedAt.Format("2006-01-02 15:04:05"),
			"confirmed_at":      confirmedAt,
			"approved_at":       approvedAt,
		}
	}

	return c.JSON(fiber.Map{
		"list":  result,
		"total": len(result),
	})
}

func (h *ArInvoiceHandler) GetByID(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	ar, err := h.repo.GetByID(c.Context(), tenantID, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if ar == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "ar invoice not found"})
	}

	var dueDate, confirmedAt, approvedAt *time.Time
	if ar.DueDate != nil {
		dueDate = ar.DueDate
	}
	if ar.ConfirmedAt != nil {
		confirmedAt = ar.ConfirmedAt
	}
	if ar.ApprovedAt != nil {
		approvedAt = ar.ApprovedAt
	}

	return c.JSON(fiber.Map{
		"id":           ar.ID,
		"invoice_id":   ar.InvoiceID,
		"invoice_no":   ar.InvoiceNo,
		"customer_id":  ar.CustomerID,
		"amount":       ar.Amount.String(),
		"due_date":     dueDate,
		"status":       ar.Status,
		"source_type":  ar.SourceType,
		"created_at":   ar.CreatedAt,
		"confirmed_at": confirmedAt,
		"approved_at":  approvedAt,
	})
}
