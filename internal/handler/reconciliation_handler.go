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
	tolerance := service.ToleranceConfig{
		Percent: decimal.NewFromInt(10),
		Enabled: true,
	}
	if t := c.Query("tolerance_percent"); t != "" {
		if p, err := decimal.NewFromString(t); err == nil && p.GreaterThan(decimal.Zero) {
			tolerance.Percent = p
		}
	}
	if t := c.Query("tolerance_enabled"); t == "false" {
		tolerance.Enabled = false
	}
	result, err := h.svc.Reconcile(c.Context(), tenantID, periodNo, tolerance)
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

// GetUnmatchedSummary returns unmatched amounts grouped by counterparty.
func (h *ReconciliationHandler) GetUnmatchedSummary(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	summary, err := h.svc.GetUnmatchedSummary(c.Context(), tenantID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": summary})
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

// ExecutePairs handles POST /reconciliation/execute.
func (h *ReconciliationHandler) ExecutePairs(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	var req struct {
		PairIDs []string `json:"pair_ids"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}
	if len(req.PairIDs) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "pair_ids is required"})
	}

	pairIDs := make([]uuid.UUID, 0, len(req.PairIDs))
	for _, s := range req.PairIDs {
		id, err := uuid.Parse(s)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid pair_id: " + s})
		}
		pairIDs = append(pairIDs, id)
	}

	result, err := h.svc.ExecutePairs(c.Context(), tenantID, pairIDs)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error(), "partial": result})
	}
	return c.JSON(fiber.Map{"data": result})
}

// ApprovePairs handles POST /reconciliation/approve.
func (h *ReconciliationHandler) ApprovePairs(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	var req struct {
		PairIDs []string `json:"pair_ids"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}
	if len(req.PairIDs) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "pair_ids is required"})
	}

	pairIDs := make([]uuid.UUID, 0, len(req.PairIDs))
	for _, s := range req.PairIDs {
		id, err := uuid.Parse(s)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid pair_id: " + s})
		}
		pairIDs = append(pairIDs, id)
	}

	result, err := h.svc.ApprovePairs(c.Context(), tenantID, pairIDs)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error(), "partial": result})
	}
	return c.JSON(fiber.Map{"data": result})
}

// RejectPairs handles POST /reconciliation/reject.
func (h *ReconciliationHandler) RejectPairs(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	var req struct {
		PairIDs []string `json:"pair_ids"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}
	if len(req.PairIDs) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "pair_ids is required"})
	}

	pairIDs := make([]uuid.UUID, 0, len(req.PairIDs))
	for _, s := range req.PairIDs {
		id, err := uuid.Parse(s)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid pair_id: " + s})
		}
		pairIDs = append(pairIDs, id)
	}

	result, err := h.svc.RejectPairs(c.Context(), tenantID, pairIDs)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error(), "partial": result})
	}
	return c.JSON(fiber.Map{"data": result})
}

// ReversePair handles POST /reconciliation/pairs/:id/reverse.
func (h *ReconciliationHandler) ReversePair(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	pairID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}
	if err := h.svc.ReversePair(c.Context(), tenantID, pairID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "reversed"})
}

// ForcePass handles POST /reconciliation/precheck/force-pass.
func (h *ReconciliationHandler) ForcePass(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	var req struct {
		PaymentID string `json:"payment_id"`
		InvoiceID string `json:"invoice_id"`
		Reason    string `json:"reason,omitempty"`
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

	pair, err := h.svc.ForcePass(c.Context(), tenantID, paymentID, invoiceID, req.Reason)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"data": pair})
}
