package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"huihua/finance/internal/service"
)

// VoucherAutoGenerateHandler handles voucher auto-generation from bank transactions.
type VoucherAutoGenerateHandler struct {
	svc *service.VoucherAutoGenerateService
}

// NewVoucherAutoGenerateHandler creates a new VoucherAutoGenerateHandler.
func NewVoucherAutoGenerateHandler(svc *service.VoucherAutoGenerateService) *VoucherAutoGenerateHandler {
	return &VoucherAutoGenerateHandler{svc: svc}
}

// GenerateFromBankTxn generates a single voucher from a bank transaction.
// POST /api/v1/bank-transactions/:id/generate-voucher
func (h *VoucherAutoGenerateHandler) GenerateFromBankTxn(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	txnIDStr := c.Params("id")
	txnID, err := uuid.Parse(txnIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid bank transaction id"})
	}

	// created_by from auth context (use system user if not set)
	createdBy := c.Locals("user_id")
	var userID uuid.UUID
	if createdBy != nil {
		userID = createdBy.(uuid.UUID)
	}

	je, err := h.svc.GenerateFromBankTxn(c.Context(), tenantID, txnID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": je})
}

// GenerateFromInvoice generates a voucher from an invoice.
// POST /api/v1/invoices/:id/generate-voucher
func (h *VoucherAutoGenerateHandler) GenerateFromInvoice(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	invoiceIDStr := c.Params("id")
	invoiceID, err := uuid.Parse(invoiceIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid invoice id"})
	}

	createdBy := c.Locals("user_id")
	var userID uuid.UUID
	if createdBy != nil {
		userID = createdBy.(uuid.UUID)
	}

	je, err := h.svc.GenerateFromInvoice(c.Context(), tenantID, invoiceID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": je})
}

// BatchGenerate generates vouchers from all unmatched bank transactions for a bank account.
// POST /api/v1/bank-transactions/batch-generate?bank_account_id=xxx
func (h *VoucherAutoGenerateHandler) BatchGenerate(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	bankAccountIDStr := c.Query("bank_account_id")
	if bankAccountIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "bank_account_id is required"})
	}
	bankAccountID, err := uuid.Parse(bankAccountIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid bank_account_id"})
	}

	createdBy := c.Locals("user_id")
	var userID uuid.UUID
	if createdBy != nil {
		userID = createdBy.(uuid.UUID)
	}

	entries, err := h.svc.BatchGenerateFromBank(c.Context(), tenantID, bankAccountID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"data":      entries,
		"total":     len(entries),
		"generated": len(entries),
	})
}
