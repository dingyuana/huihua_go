package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"huihua/finance/internal/repository"
)

type ApInvoiceHandler struct {
	repo      *repository.ApInvoiceRepository
	partyRepo *repository.PartyRepository
}

func NewApInvoiceHandler(repo *repository.ApInvoiceRepository, partyRepo *repository.PartyRepository) *ApInvoiceHandler {
	return &ApInvoiceHandler{repo: repo, partyRepo: partyRepo}
}

func (h *ApInvoiceHandler) partyNameMap(c *fiber.Ctx, tenantID uuid.UUID) (map[uuid.UUID]string, error) {
	parties, err := h.partyRepo.List(c.Context(), tenantID)
	if err != nil {
		return nil, err
	}
	m := make(map[uuid.UUID]string, len(parties))
	for _, p := range parties {
		m[p.ID] = p.Name
	}
	return m, nil
}

func (h *ApInvoiceHandler) List(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	var status *string
	if v := c.Query("status"); v != "" {
		status = &v
	}

	aps, err := h.repo.ListByTenant(c.Context(), tenantID, status)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	nameMap, _ := h.partyNameMap(c, tenantID)

	result := make([]map[string]interface{}, len(aps))
	for i, ap := range aps {
		var dueDate, confirmedAt, approvedAt string
		if ap.DueDate != nil {
			dueDate = ap.DueDate.Format("2006-01-02")
		}
		if ap.ConfirmedAt != nil {
			confirmedAt = ap.ConfirmedAt.Format("2006-01-02 15:04:05")
		}
		if ap.ApprovedAt != nil {
			approvedAt = ap.ApprovedAt.Format("2006-01-02 15:04:05")
		}
		result[i] = map[string]interface{}{
			"id":               ap.ID,
			"invoice_id":       ap.InvoiceID,
			"invoice_no":       ap.InvoiceNo,
			"supplier_id":      ap.SupplierID,
			"supplier_name":    nameMap[ap.SupplierID],
			"amount":           ap.Amount.String(),
			"paid_amount":      ap.PaidAmount.String(),
			"outstanding_amount": ap.OutstandingAmount.String(),
			"due_date":         dueDate,
			"status":           ap.Status,
			"source_type":      ap.SourceType,
			"remark":           derefStr(ap.Remark),
			"created_at":       ap.CreatedAt.Format("2006-01-02 15:04:05"),
			"confirmed_at":     confirmedAt,
			"approved_at":      approvedAt,
		}
	}

	return c.JSON(fiber.Map{
		"list":  result,
		"total": len(result),
	})
}

func (h *ApInvoiceHandler) GetByID(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	ap, err := h.repo.GetByID(c.Context(), tenantID, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if ap == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "ap_invoice not found"})
	}

	nameMap, _ := h.partyNameMap(c, tenantID)
	remark, _ := h.repo.GetSourceInvoiceRemark(c.Context(), tenantID, ap.InvoiceID)

	var dueDate, confirmedAt, approvedAt *time.Time
	if ap.DueDate != nil {
		dueDate = ap.DueDate
	}
	if ap.ConfirmedAt != nil {
		confirmedAt = ap.ConfirmedAt
	}
	if ap.ApprovedAt != nil {
		approvedAt = ap.ApprovedAt
	}

	return c.JSON(fiber.Map{
		"id":               ap.ID,
		"invoice_id":       ap.InvoiceID,
		"invoice_no":       ap.InvoiceNo,
		"supplier_id":      ap.SupplierID,
		"supplier_name":    nameMap[ap.SupplierID],
		"amount":           ap.Amount.String(),
		"paid_amount":      ap.PaidAmount.String(),
		"outstanding_amount": ap.OutstandingAmount.String(),
		"due_date":         dueDate,
		"status":           ap.Status,
		"source_type":      ap.SourceType,
		"remark":           remark,
		"created_at":       ap.CreatedAt,
		"confirmed_at":     confirmedAt,
		"approved_at":      approvedAt,
	})
}
