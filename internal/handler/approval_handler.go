package handler

import (
	"encoding/json"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"huihua/finance/internal/service"
)

// ApprovalHandler handles approval workflow HTTP requests.
type ApprovalHandler struct {
	approvalSvc *service.ApprovalService
}

// NewApprovalHandler creates a new ApprovalHandler.
func NewApprovalHandler(approvalSvc *service.ApprovalService) *ApprovalHandler {
	return &ApprovalHandler{approvalSvc: approvalSvc}
}

// ApprovalSubmitRequest is the request body for submitting a voucher for approval.
type ApprovalSubmitRequest struct {
	UserID     string `json:"user_id"`
	VoucherID  string `json:"voucher_id"`
}

// SubmitForApproval handles POST /api/v1/approvals/submit
func (h *ApprovalHandler) SubmitForApproval(c *fiber.Ctx) error {
	var req ApprovalSubmitRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user_id"})
	}

	voucherID, err := uuid.Parse(req.VoucherID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid voucher_id"})
	}

	tenantID := c.Locals("tenant_id").(uuid.UUID)

	if err := h.approvalSvc.SubmitForApproval(c.Context(), tenantID, voucherID, userID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "voucher submitted for approval"})
}

// ApproveRequest is the request body for approving a voucher.
type ApproveRequest struct {
	UserID   string `json:"user_id"`
	Comment  string `json:"comment"`
}

// Approve handles POST /api/v1/approvals/:id/approve
func (h *ApprovalHandler) Approve(c *fiber.Ctx) error {
	idStr := c.Params("id")
	taskID, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid task id"})
	}

	var req ApproveRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user_id"})
	}

	tenantID := c.Locals("tenant_id").(uuid.UUID)

	if err := h.approvalSvc.Approve(c.Context(), tenantID, taskID, userID, req.Comment); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "voucher approved"})
}

// ApprovalRejectRequest is the request body for rejecting a voucher.
type ApprovalRejectRequest struct {
	UserID  string `json:"user_id"`
	Comment string `json:"comment"`
}

// Reject handles POST /api/v1/approvals/:id/reject
func (h *ApprovalHandler) Reject(c *fiber.Ctx) error {
	idStr := c.Params("id")
	taskID, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid task id"})
	}

	var req ApprovalRejectRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user_id"})
	}

	tenantID := c.Locals("tenant_id").(uuid.UUID)

	if err := h.approvalSvc.Reject(c.Context(), tenantID, taskID, userID, req.Comment); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "voucher rejected"})
}

// GetPendingTasks handles GET /api/v1/approvals/pending
func (h *ApprovalHandler) GetPendingTasks(c *fiber.Ctx) error {
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "user_id required"})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user_id"})
	}

	tenantID := c.Locals("tenant_id").(uuid.UUID)

	tasks, err := h.approvalSvc.GetPendingTasks(c.Context(), tenantID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"tasks": tasks,
		"count": len(tasks),
	})
}

// GetApprovalHistory handles GET /api/v1/approvals/history
func (h *ApprovalHandler) GetApprovalHistory(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	limitStr := c.Query("limit", "20")
	offsetStr := c.Query("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	history, err := h.approvalSvc.GetApprovalHistory(c.Context(), tenantID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"history": history,
		"count":   len(history),
		"limit":   limit,
		"offset":  offset,
	})
}

// GetVoucherApprovalStatus handles GET /api/v1/approvals/voucher/:id/status
func (h *ApprovalHandler) GetVoucherApprovalStatus(c *fiber.Ctx) error {
	idStr := c.Params("id")
	journalEntryID, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid voucher id"})
	}

	tenantID := c.Locals("tenant_id").(uuid.UUID)

	status, tasks, err := h.approvalSvc.GetJournalEntryApprovalStatus(c.Context(), tenantID, journalEntryID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"status": status,
		"tasks":  tasks,
	})
}

// ApproverInput represents an approver input for creating flows.
type ApproverInput struct {
	Level      int    `json:"level"`
	ApproverID string `json:"approver_id"`
	Role       string `json:"role"`
}

// CreateFlowRequest is the request for creating an approval flow.
type CreateFlowRequest struct {
	FlowName    string          `json:"flow_name"`
	Description *string         `json:"description"`
	Approvers   []ApproverInput `json:"approvers"`
	UserID      string          `json:"user_id"`
}

// CreateApprovalFlow handles POST /api/v1/approval-flows
func (h *ApprovalHandler) CreateApprovalFlow(c *fiber.Ctx) error {
	var req CreateFlowRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	tenantID := c.Locals("tenant_id").(uuid.UUID)

	var createdBy *uuid.UUID
	if req.UserID != "" {
		id, err := uuid.Parse(req.UserID)
		if err == nil {
			createdBy = &id
		}
	}

	// Convert approvers to service format
	approvers, err := h.convertApprovers(req.Approvers)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	flow, err := h.approvalSvc.CreateApprovalFlow(c.Context(), tenantID, req.FlowName, req.Description, approvers, createdBy)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(flow)
}

// convertApprovers converts handler ApproverInput to service ApproverInput
func (h *ApprovalHandler) convertApprovers(inputs []ApproverInput) ([]service.ApproverInput, error) {
	approvers := make([]service.ApproverInput, len(inputs))
	for i, a := range inputs {
		approvers[i] = service.ApproverInput{
			Level: a.Level,
			Role:  a.Role,
		}
		if a.ApproverID != "" {
			id, err := uuid.Parse(a.ApproverID)
			if err != nil {
				return nil, err
			}
			approvers[i].ApproverID = id
		}
	}
	return approvers, nil
}

// ListApprovalFlows handles GET /api/v1/approval-flows
func (h *ApprovalHandler) ListApprovalFlows(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	flows, err := h.approvalSvc.ListApprovalFlows(c.Context(), tenantID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"flows": flows,
		"count": len(flows),
	})
}

// MarshalJSON for ApproverInput - custom marshal for internal use
func (a ApproverInput) MarshalJSON() ([]byte, error) {
	type Alias ApproverInput
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(&a),
	})
}