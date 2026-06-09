package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/service"
)

// ReconciliationHandler handles reconciliation HTTP requests.
type ReconciliationHandler struct {
	svc *service.ReconciliationService
}

// NewReconciliationHandler creates a new ReconciliationHandler.
func NewReconciliationHandler(svc *service.ReconciliationService) *ReconciliationHandler {
	return &ReconciliationHandler{svc: svc}
}

// Run executes the reconciliation matching.
func (h *ReconciliationHandler) Run(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	periodNo := c.QueryInt("period_no", 0)
	result, err := h.svc.Reconcile(c.Context(), tenantID, periodNo)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": result})
}

// ListPairs returns reconciliation pairs.
func (h *ReconciliationHandler) ListPairs(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	status := c.Query("status")
	pairs, err := h.svc.ListPairs(c.Context(), tenantID, status)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": pairs})
}

// ConfirmPair confirms a pair.
func (h *ReconciliationHandler) ConfirmPair(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	pairID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}
	if err := h.svc.ConfirmPair(c.Context(), tenantID, pairID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "confirmed"})
}

// UnconfirmPair unconfirms a pair.
func (h *ReconciliationHandler) UnconfirmPair(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	pairID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}
	if err := h.svc.UnconfirmPair(c.Context(), tenantID, pairID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "unconfirmed"})
}

// GetUnmatched returns unmatched items.
func (h *ReconciliationHandler) GetUnmatched(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	items, err := h.svc.GetUnmatched(c.Context(), tenantID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": items})
}

// PreCheck handles POST /reconciliation/precheck.
func (h *ReconciliationHandler) PreCheck(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	var req struct {
		PaymentID string `json:"payment_id"`
		InvoiceID string `json:"invoice_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.PaymentID == "" || req.InvoiceID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "payment_id and invoice_id are required"})
	}
	paymentID, err := uuid.Parse(req.PaymentID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid payment_id"})
	}
	invoiceID, err := uuid.Parse(req.InvoiceID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid invoice_id"})
	}

	checks, err := h.svc.PreCheck(c.Context(), tenantID, paymentID, invoiceID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"data": fiber.Map{"checks": checks}})
}

// ManualMatch handles POST /reconciliation/manual.
func (h *ReconciliationHandler) ManualMatch(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	userID := c.Locals("user_id").(uuid.UUID)

	var req struct {
		BankTransactionID string `json:"bank_transaction_id"`
		Allocations       []struct {
			InvoiceID string `json:"invoice_id"`
			Amount    string `json:"amount"`
		} `json:"allocations"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.BankTransactionID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "bank_transaction_id is required"})
	}
	bankTxnID, err := uuid.Parse(req.BankTransactionID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid bank_transaction_id"})
	}

	allocations := make([]service.ManualAllocation, 0, len(req.Allocations))
	for _, a := range req.Allocations {
		invID, err := uuid.Parse(a.InvoiceID)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid invoice_id: " + a.InvoiceID})
		}
		amt, err := decimal.NewFromString(a.Amount)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid amount: " + a.Amount})
		}
		allocations = append(allocations, service.ManualAllocation{InvoiceID: invID, Amount: amt})
	}

	pairs, err := h.svc.ManualMatch(c.Context(), tenantID, userID, &service.ManualMatchRequest{
		BankTransactionID: bankTxnID,
		Allocations:       allocations,
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"data": pairs, "count": len(pairs)})
}
