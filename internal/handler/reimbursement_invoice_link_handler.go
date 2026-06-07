package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/model"
	"huihua/finance/internal/service"
)

// ReimbursementInvoiceLinkHandler handles invoice-link HTTP requests.
type ReimbursementInvoiceLinkHandler struct {
	svc *service.ReimbursementInvoiceLinkService
}

// NewReimbursementInvoiceLinkHandler creates a new ReimbursementInvoiceLinkHandler.
func NewReimbursementInvoiceLinkHandler(svc *service.ReimbursementInvoiceLinkService) *ReimbursementInvoiceLinkHandler {
	return &ReimbursementInvoiceLinkHandler{svc: svc}
}

// ListAvailable handles GET /api/v1/reimbursements/:id/invoices
func (h *ReimbursementInvoiceLinkHandler) ListAvailable(c *fiber.Ctx) error {
	idStr := c.Params("id")
	reimbID, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid reimbursement id"})
	}

	tenantID := c.Locals("tenant_id").(uuid.UUID)

	list, err := h.svc.ListAvailableInvoices(c.Context(), reimbID, tenantID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"invoices": list})
}

// Link handles POST /api/v1/reimbursements/:id/invoices/:invoice_id
func (h *ReimbursementInvoiceLinkHandler) Link(c *fiber.Ctx) error {
	reimbIDStr := c.Params("id")
	reimbID, err := uuid.Parse(reimbIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid reimbursement id"})
	}

	invoiceIDStr := c.Params("invoice_id")
	invoiceID, err := uuid.Parse(invoiceIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid invoice id"})
	}

	tenantID := c.Locals("tenant_id").(uuid.UUID)
	userID := c.Locals("user_id").(uuid.UUID)

	var req model.ReimbursementInvoiceLinkRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	link, err := h.svc.LinkInvoice(c.Context(), tenantID, reimbID, invoiceID, userID, req.LinkedAmount)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"link": link})
}

// Unlink handles DELETE /api/v1/reimbursements/:id/invoices/:invoice_id
func (h *ReimbursementInvoiceLinkHandler) Unlink(c *fiber.Ctx) error {
	reimbIDStr := c.Params("id")
	reimbID, err := uuid.Parse(reimbIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid reimbursement id"})
	}

	invoiceIDStr := c.Params("invoice_id")
	invoiceID, err := uuid.Parse(invoiceIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid invoice id"})
	}

	if err := h.svc.UnlinkInvoice(c.Context(), reimbID, invoiceID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "invoice unlinked"})
}

// ListLinked handles GET /api/v1/reimbursements/:id/invoices/linked
func (h *ReimbursementInvoiceLinkHandler) ListLinked(c *fiber.Ctx) error {
	idStr := c.Params("id")
	reimbID, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid reimbursement id"})
	}

	list, err := h.svc.GetLinkedInvoices(c.Context(), reimbID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"links": list})
}

// Helper to convert decimal string to decimal.Decimal
func parseDecimal(s string) (decimal.Decimal, error) {
	return decimal.NewFromString(s)
}