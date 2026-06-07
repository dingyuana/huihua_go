package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/model"
	"huihua/finance/internal/repository"
	"huihua/finance/internal/service"
)

type ApInvoiceHandler struct {
	repo      *repository.ApInvoiceRepository
	partyRepo *repository.PartyRepository
	svc       *service.ApInvoiceService
}

func NewApInvoiceHandler(
	repo *repository.ApInvoiceRepository,
	partyRepo *repository.PartyRepository,
	svc *service.ApInvoiceService,
) *ApInvoiceHandler {
	return &ApInvoiceHandler{repo: repo, partyRepo: partyRepo, svc: svc}
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
			"id":                 ap.ID,
			"invoice_id":         ap.InvoiceID,
			"invoice_no":         ap.InvoiceNo,
			"supplier_id":        ap.SupplierID,
			"supplier_name":      nameMap[ap.SupplierID],
			"amount":             ap.Amount.String(),
			"paid_amount":        ap.PaidAmount.String(),
			"outstanding_amount": ap.OutstandingAmount.String(),
			"due_date":           dueDate,
			"status":             ap.Status,
			"source_type":        ap.SourceType,
			"remark":             derefStr(ap.Remark),
			"created_at":         ap.CreatedAt.Format("2006-01-02 15:04:05"),
			"confirmed_at":       confirmedAt,
			"approved_at":        approvedAt,
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
		"id":                 ap.ID,
		"invoice_id":         ap.InvoiceID,
		"invoice_no":         ap.InvoiceNo,
		"supplier_id":        ap.SupplierID,
		"supplier_name":      nameMap[ap.SupplierID],
		"amount":             ap.Amount.String(),
		"paid_amount":        ap.PaidAmount.String(),
		"outstanding_amount": ap.OutstandingAmount.String(),
		"due_date":           dueDate,
		"status":             ap.Status,
		"source_type":        ap.SourceType,
		"remark":             remark,
		"created_at":         ap.CreatedAt,
		"confirmed_at":       confirmedAt,
		"approved_at":        approvedAt,
	})
}

// Create POST /api/v1/ap-invoices
func (h *ApInvoiceHandler) Create(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	userID := c.Locals("user_id").(uuid.UUID)

	var req model.ApInvoiceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	entity := &req.ApInvoice
	entity.TenantID = tenantID
	if entity.CreatedBy == nil {
		entity.CreatedBy = &userID
	}
	if entity.PaidAmount.IsZero() {
		entity.PaidAmount = decimal.Zero
	}
	if due, ok := parseDueDate(req.DueDateStr); ok {
		entity.DueDate = due
	}

	if err := h.svc.Create(c.Context(), entity); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": entity})
}

// Confirm POST /api/v1/ap-invoices/:id/confirm
func (h *ApInvoiceHandler) Confirm(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	userID := c.Locals("user_id").(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	if err := h.svc.Confirm(c.Context(), tenantID, id, userID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "ap_invoice confirmed"})
}

// Approve POST /api/v1/ap-invoices/:id/approve
func (h *ApInvoiceHandler) Approve(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	userID := c.Locals("user_id").(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	if err := h.svc.Approve(c.Context(), tenantID, id, userID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "ap_invoice approved"})
}

// Allocate POST /api/v1/ap-invoices/:id/allocate
// Applies a payment amount to the ap_invoice, reducing outstanding balance.
// Body: { "payment_amount": 100.50, "payment_entry_id": "uuid" (optional) }
func (h *ApInvoiceHandler) Allocate(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	userID := c.Locals("user_id").(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	var req struct {
		PaymentAmount  float64 `json:"payment_amount"`
		PaymentEntryID string  `json:"payment_entry_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.PaymentAmount <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "payment_amount must be positive"})
	}

	if err := h.svc.Allocate(c.Context(), tenantID, id, userID, decimal.NewFromFloat(req.PaymentAmount)); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Reload the updated ap_invoice to return fresh state.
	ap, err := h.repo.GetByID(c.Context(), tenantID, id)
	if err != nil || ap == nil {
		return c.JSON(fiber.Map{"message": "ap_invoice allocated"})
	}
	return c.JSON(fiber.Map{
		"message": "ap_invoice allocated",
		"data": fiber.Map{
			"id":                 ap.ID,
			"paid_amount":        ap.PaidAmount.String(),
			"outstanding_amount": ap.OutstandingAmount.String(),
			"status":             ap.Status,
		},
	})
}

// Update PUT /api/v1/ap-invoices/:id
func (h *ApInvoiceHandler) Update(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	var req model.ApInvoiceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	entity := &req.ApInvoice
	if due, ok := parseDueDate(req.DueDateStr); ok {
		entity.DueDate = due
	}

	if err := h.svc.Update(c.Context(), tenantID, id, entity); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "ap_invoice updated"})
}

// Delete DELETE /api/v1/ap-invoices/:id
func (h *ApInvoiceHandler) Delete(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	if err := h.svc.Delete(c.Context(), tenantID, id); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "ap_invoice deleted"})
}

// ListBySupplier GET /api/v1/ap-invoices/by-supplier/:supplier_id
func (h *ApInvoiceHandler) ListBySupplier(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	supplierID, err := uuid.Parse(c.Params("supplier_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid supplier_id"})
	}

	var status *string
	if v := c.Query("status"); v != "" {
		status = &v
	}

	aps, err := h.svc.ListBySupplier(c.Context(), tenantID, supplierID, status)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	nameMap, _ := h.partyNameMap(c, tenantID)
	result := make([]map[string]interface{}, len(aps))
	for i, ap := range aps {
		result[i] = map[string]interface{}{
			"id":                 ap.ID,
			"invoice_no":         ap.InvoiceNo,
			"amount":             ap.Amount.String(),
			"paid_amount":        ap.PaidAmount.String(),
			"outstanding_amount": ap.OutstandingAmount.String(),
			"status":             ap.Status,
			"supplier_name":      nameMap[ap.SupplierID],
		}
	}
	return c.JSON(fiber.Map{"list": result, "total": len(result)})
}

// ListOutstanding GET /api/v1/ap-invoices/outstanding
func (h *ApInvoiceHandler) ListOutstanding(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	aps, err := h.svc.ListOutstanding(c.Context(), tenantID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"list": aps, "total": len(aps)})
}

// parseDueDate parses a YYYY-MM-DD string into *time.Time. Returns (nil, false) if empty or invalid.
func parseDueDate(s *string) (*time.Time, bool) {
	if s == nil || *s == "" {
		return nil, false
	}
	t, err := time.Parse("2006-01-02", *s)
	if err != nil {
		return nil, false
	}
	return &t, true
}
