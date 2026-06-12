package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"huihua/finance/internal/model"
	"huihua/finance/internal/service"
)

type ExpenseInvoiceHandler struct {
	svc *service.ExpenseInvoiceService
}

func NewExpenseInvoiceHandler(svc *service.ExpenseInvoiceService) *ExpenseInvoiceHandler {
	return &ExpenseInvoiceHandler{svc: svc}
}

// Create handles POST /expense-invoices
func (h *ExpenseInvoiceHandler) Create(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	var req model.ExpenseInvoiceCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	inv, err := h.svc.Create(c.Context(), tenantID, &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(inv)
}

// List handles GET /expense-invoices
func (h *ExpenseInvoiceHandler) List(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	filters := model.ExpenseInvoiceFilter{}

	if vendorID := c.Query("vendor_id"); vendorID != "" {
		if id, err := uuid.Parse(vendorID); err == nil {
			filters.VendorID = &id
		}
	}
	if verifyStatus := c.Query("verify_status"); verifyStatus != "" {
		filters.VerifyStatus = verifyStatus
	}
	if fromDate := c.Query("from_date"); fromDate != "" {
		if t, err := time.Parse("2006-01-02", fromDate); err == nil {
			filters.FromDate = &t
		}
	}
	if toDate := c.Query("to_date"); toDate != "" {
		if t, err := time.Parse("2006-01-02", toDate); err == nil {
			filters.ToDate = &t
		}
	}
	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			filters.Limit = l
		}
	}
	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			filters.Offset = o
		}
	}

	list, err := h.svc.List(c.Context(), tenantID, filters)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": list,
	})
}

// GetByID handles GET /expense-invoices/:id
func (h *ExpenseInvoiceHandler) GetByID(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid id",
		})
	}

	inv, err := h.svc.GetByID(c.Context(), tenantID, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if inv == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "expense invoice not found",
		})
	}

	return c.JSON(inv)
}

// Update handles PUT /expense-invoices/:id
func (h *ExpenseInvoiceHandler) Update(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid id",
		})
	}

	var fields map[string]interface{}
	if err := c.BodyParser(&fields); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if err := h.svc.Update(c.Context(), tenantID, id, fields); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{"message": "updated"})
}

// Delete handles DELETE /expense-invoices/:id
func (h *ExpenseInvoiceHandler) Delete(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid id",
		})
	}

	if err := h.svc.Delete(c.Context(), tenantID, id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{"message": "deleted"})
}

// Verify handles POST /expense-invoices/:id/verify
func (h *ExpenseInvoiceHandler) Verify(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid id",
		})
	}

	result, err := h.svc.VerifyInvoice(c.Context(), tenantID, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(result)
}

// BatchVerify handles POST /expense-invoices/verify/batch
func (h *ExpenseInvoiceHandler) BatchVerify(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	var req struct {
		IDs []uuid.UUID `json:"ids"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	results, err := h.svc.BatchVerify(c.Context(), tenantID, req.IDs)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": results,
	})
}

// Confirm handles POST /expense-invoices/:id/confirm
// Confirms the expense invoice and auto-creates the corresponding ApInvoice.
func (h *ExpenseInvoiceHandler) Confirm(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	userID := c.Locals("user_id").(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid id",
		})
	}

	inv, err := h.svc.Confirm(c.Context(), tenantID, id, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(inv)
}

// Deduct handles POST /expense-invoices/:id/deduct
func (h *ExpenseInvoiceHandler) Deduct(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid id",
		})
	}

	inv, err := h.svc.DeductInvoice(c.Context(), tenantID, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(inv)
}