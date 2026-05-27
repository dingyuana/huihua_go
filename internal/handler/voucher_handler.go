package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"huihua/finance/internal/service"
	"huihua/finance/internal/repository"
)

// VoucherHandler handles voucher state machine operations.
type VoucherHandler struct {
	stateMachine *service.VoucherStateMachine
	journalRepo  *repository.JournalRepository
}

// NewVoucherHandler creates a new VoucherHandler.
func NewVoucherHandler(stateMachine *service.VoucherStateMachine, journalRepo *repository.JournalRepository) *VoucherHandler {
	return &VoucherHandler{
		stateMachine: stateMachine,
		journalRepo:  journalRepo,
	}
}

// SubmitRequest is the request body for submitting a voucher.
type SubmitRequest struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
}

// Submit handles POST /api/v1/vouchers/:id/submit
func (h *VoucherHandler) Submit(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	var req SubmitRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user_id"})
	}

	tenantID := c.Locals("tenant_id").(uuid.UUID)

	if err := h.stateMachine.ExecuteTransition(c.Context(), tenantID, id, "submit", userID, req.UserName, ""); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "voucher submitted successfully"})
}

// Approve handles POST /api/v1/vouchers/:id/approve
func (h *VoucherHandler) Approve(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	var req SubmitRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user_id"})
	}

	tenantID := c.Locals("tenant_id").(uuid.UUID)

	if err := h.stateMachine.ExecuteTransition(c.Context(), tenantID, id, "approve", userID, req.UserName, ""); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "voucher approved successfully"})
}

// RejectRequest is the request body for rejecting a voucher.
type RejectRequest struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
	Reason   string `json:"reason"`
}

// Reject handles POST /api/v1/vouchers/:id/reject
func (h *VoucherHandler) Reject(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	var req RejectRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user_id"})
	}

	tenantID := c.Locals("tenant_id").(uuid.UUID)

	if err := h.stateMachine.ExecuteTransition(c.Context(), tenantID, id, "reject", userID, req.UserName, req.Reason); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "voucher rejected"})
}

// CancelRequest is the request body for cancelling a voucher.
type CancelRequest struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
	Reason   string `json:"reason"`
}

// Cancel handles POST /api/v1/vouchers/:id/cancel
func (h *VoucherHandler) Cancel(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	var req CancelRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user_id"})
	}

	tenantID := c.Locals("tenant_id").(uuid.UUID)

	if err := h.stateMachine.ExecuteTransition(c.Context(), tenantID, id, "cancel", userID, req.UserName, req.Reason); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "voucher cancelled"})
}

// ReverseRequest is the request body for reversing a voucher.
type ReverseRequest struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
}

// Reverse handles POST /api/v1/vouchers/:id/reverse
func (h *VoucherHandler) Reverse(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	var req ReverseRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user_id"})
	}

	tenantID := c.Locals("tenant_id").(uuid.UUID)

	reversal, err := h.stateMachine.ReverseVoucher(c.Context(), tenantID, id, userID, req.UserName)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message":             "voucher reversed successfully",
		"reversal_voucher_id": reversal.ID,
	})
}

// GetStatus handles GET /api/v1/vouchers/:id/status
func (h *VoucherHandler) GetStatus(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	tenantID := c.Locals("tenant_id").(uuid.UUID)

	status, err := h.journalRepo.GetStatus(c.Context(), tenantID, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"docstatus": status,
		"status":    service.DocStatusToVoucherStatus(status),
	})
}

// GetTransitions handles GET /api/v1/vouchers/:id/transitions
func (h *VoucherHandler) GetTransitions(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	tenantID := c.Locals("tenant_id").(uuid.UUID)

	transitions, err := h.journalRepo.GetTransitions(c.Context(), tenantID, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"transitions": transitions})
}