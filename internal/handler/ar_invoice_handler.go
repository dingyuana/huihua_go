package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"huihua/finance/internal/repository"
	"huihua/finance/internal/service"
)

type ArInvoiceHandler struct {
	repo      *repository.ArInvoiceRepository
	partyRepo *repository.PartyRepository
	svc       *service.ArInvoiceService
}

func NewArInvoiceHandler(
	repo *repository.ArInvoiceRepository,
	partyRepo *repository.PartyRepository,
	svc *service.ArInvoiceService,
) *ArInvoiceHandler {
	return &ArInvoiceHandler{repo: repo, partyRepo: partyRepo, svc: svc}
}

func (h *ArInvoiceHandler) partyNameMap(c *fiber.Ctx, tenantID uuid.UUID) (map[uuid.UUID]string, error) {
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

	nameMap, _ := h.partyNameMap(c, tenantID)

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
			"id":               ar.ID,
			"invoice_id":       ar.InvoiceID,
			"invoice_no":       ar.InvoiceNo,
			"customer_id":      ar.CustomerID,
			"customer_name":    nameMap[ar.CustomerID],
			"amount":           ar.Amount.String(),
			"paid_amount":      ar.PaidAmount.String(),
			"outstanding_amount": ar.OutstandingAmount.String(),
			"due_date":         dueDate,
			"status":           ar.Status,
			"source_type":      ar.SourceType,
			"remark":           derefStr(ar.Remark),
			"created_at":       ar.CreatedAt.Format("2006-01-02 15:04:05"),
			"confirmed_at":     confirmedAt,
			"approved_at":      approvedAt,
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
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "ar_invoice not found"})
	}

	nameMap, _ := h.partyNameMap(c, tenantID)
	remark, _ := h.repo.GetSourceInvoiceRemark(c.Context(), tenantID, ar.InvoiceID)

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
		"id":               ar.ID,
		"invoice_id":       ar.InvoiceID,
		"invoice_no":       ar.InvoiceNo,
		"customer_id":      ar.CustomerID,
		"customer_name":    nameMap[ar.CustomerID],
		"amount":           ar.Amount.String(),
		"paid_amount":      ar.PaidAmount.String(),
		"outstanding_amount": ar.OutstandingAmount.String(),
		"due_date":         dueDate,
		"status":           ar.Status,
		"source_type":      ar.SourceType,
		"remark":           remark,
		"created_at":       ar.CreatedAt,
		"confirmed_at":     confirmedAt,
		"approved_at":      approvedAt,
	})
}

// Confirm POST /api/v1/ar-invoices/:id/confirm
func (h *ArInvoiceHandler) Confirm(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	userID := c.Locals("user_id").(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	if err := h.svc.Confirm(c.Context(), tenantID, id, userID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "ar_invoice confirmed"})
}

// Delete DELETE /api/v1/ar-invoices/:id
func (h *ArInvoiceHandler) Delete(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	if err := h.svc.Delete(c.Context(), tenantID, id); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "ar_invoice deleted"})
}

// ListByCustomer GET /api/v1/ar-invoices/customer/:customerId
func (h *ArInvoiceHandler) ListByCustomer(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	customerID, err := uuid.Parse(c.Params("customerId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid customer_id"})
	}
	var status *string
	if v := c.Query("status"); v != "" {
		status = &v
	}
	ars, err := h.svc.ListByCustomer(c.Context(), tenantID, customerID, status)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	nameMap, _ := h.partyNameMap(c, tenantID)
	result := make([]map[string]interface{}, len(ars))
	for i, ar := range ars {
		result[i] = map[string]interface{}{
			"id":                 ar.ID,
			"invoice_no":         ar.InvoiceNo,
			"amount":             ar.Amount.String(),
			"paid_amount":        ar.PaidAmount.String(),
			"outstanding_amount": ar.OutstandingAmount.String(),
			"status":             ar.Status,
			"customer_name":      nameMap[ar.CustomerID],
		}
	}
	return c.JSON(fiber.Map{"list": result, "total": len(result)})
}

// ListUnpaidByCustomer GET /api/v1/ar-invoices/customer/:customerId/unpaid
func (h *ArInvoiceHandler) ListUnpaidByCustomer(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	customerID, err := uuid.Parse(c.Params("customerId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid customer_id"})
	}
	ars, err := h.svc.ListUnpaidByCustomer(c.Context(), tenantID, customerID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"list": ars, "total": len(ars)})
}

// CustomerSummary GET /api/v1/ar-invoices/customer/:customerId/summary
func (h *ArInvoiceHandler) CustomerSummary(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	customerID, err := uuid.Parse(c.Params("customerId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid customer_id"})
	}
	summary, err := h.svc.GetCustomerSummary(c.Context(), tenantID, customerID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(summary)
}
