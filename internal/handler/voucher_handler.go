package handler

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"huihua/finance/internal/repository"
	"huihua/finance/internal/service"
)

// VoucherHandler handles voucher state machine operations and manual CRUD.
type VoucherHandler struct {
	stateMachine *service.VoucherStateMachine
	journalRepo  *repository.JournalRepository
	voucherSvc   *service.VoucherService
}

// NewVoucherHandler creates a new VoucherHandler.
func NewVoucherHandler(stateMachine *service.VoucherStateMachine, journalRepo *repository.JournalRepository, voucherSvc *service.VoucherService) *VoucherHandler {
	return &VoucherHandler{
		stateMachine: stateMachine,
		journalRepo:  journalRepo,
		voucherSvc:   voucherSvc,
	}
}

// parseDateRange parses start_date and end_date query params.
func parseDateRange(c *fiber.Ctx) (startDate, endDate *time.Time) {
	if sd := c.Query("start_date"); sd != "" {
		if t, err := time.Parse("2006-01-02", sd); err == nil {
			startDate = &t
		}
	}
	if ed := c.Query("end_date"); ed != "" {
		if t, err := time.Parse("2006-01-02", ed); err == nil {
			endDate = &t
		}
	}
	return
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

	tenantID := c.Locals("tenant_id").(uuid.UUID)
	userID := c.Locals("user_id").(uuid.UUID)

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

	tenantID := c.Locals("tenant_id").(uuid.UUID)
	userID := c.Locals("user_id").(uuid.UUID)

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

	tenantID := c.Locals("tenant_id").(uuid.UUID)
	userID := c.Locals("user_id").(uuid.UUID)

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

	tenantID := c.Locals("tenant_id").(uuid.UUID)
	userID := c.Locals("user_id").(uuid.UUID)

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

	tenantID := c.Locals("tenant_id").(uuid.UUID)
	userID := c.Locals("user_id").(uuid.UUID)

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

// ---- Manual Voucher CRUD methods ----

// CreateRequest is the request body for creating a voucher.
type CreateRequest struct {
	VoucherType *string       `json:"voucher_type"`
	PostingDate string        `json:"posting_date"`
	CompanyID   string        `json:"company_id"`
	Remark      *string       `json:"remark,omitempty"`
	Lines       []LineRequest `json:"lines"`
}

// LineRequest is a single line in a voucher create/update request.
type LineRequest struct {
	AccountID    string  `json:"account_id"`
	Debit        string  `json:"debit"`
	Credit       string  `json:"credit"`
	PartyType    *string `json:"party_type,omitempty"`
	PartyID      *string `json:"party_id,omitempty"`
	CostCenterID *string `json:"cost_center_id,omitempty"`
	ProjectID    *string `json:"project_id,omitempty"`
	UserRemark   *string `json:"user_remark,omitempty"`
}

// List handles GET /api/v1/vouchers
func (h *VoucherHandler) List(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	startDate, endDate := parseDateRange(c)
	limit := c.QueryInt("limit", 50)
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := c.QueryInt("offset", 0)

	vouchers, err := h.voucherSvc.ListVouchers(c.Context(), tenantID, &service.ListVouchersRequest{
		StartDate: startDate,
		EndDate:   endDate,
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"vouchers": vouchers})
}

// GetByID handles GET /api/v1/vouchers/:id
func (h *VoucherHandler) GetByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	tenantID := c.Locals("tenant_id").(uuid.UUID)

	voucher, err := h.voucherSvc.GetVoucher(c.Context(), tenantID, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "voucher not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(voucher)
}

// Create handles POST /api/v1/vouchers
func (h *VoucherHandler) Create(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	userID := c.Locals("user_id").(uuid.UUID)

	var req CreateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	companyID, err := uuid.Parse(req.CompanyID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid company_id"})
	}

	postingDate, err := time.Parse("2006-01-02", req.PostingDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid posting_date, use YYYY-MM-DD"})
	}

	svcReq := &service.CreateVoucherRequest{
		VoucherType: req.VoucherType,
		PostingDate: postingDate,
		CompanyID:   companyID,
		Remark:      req.Remark,
		Lines:       make([]service.VoucherLineRequest, 0, len(req.Lines)),
	}

	for _, l := range req.Lines {
		accountID, err := uuid.Parse(l.AccountID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid line account_id"})
		}
		line := service.VoucherLineRequest{
			AccountID:  accountID,
			Debit:      l.Debit,
			Credit:     l.Credit,
			PartyType:  l.PartyType,
			UserRemark: l.UserRemark,
		}
		if l.PartyID != nil {
			pid, err := uuid.Parse(*l.PartyID)
			if err == nil {
				line.PartyID = &pid
			}
		}
		if l.CostCenterID != nil {
			ccid, err := uuid.Parse(*l.CostCenterID)
			if err == nil {
				line.CostCenterID = &ccid
			}
		}
		if l.ProjectID != nil {
			projID, err := uuid.Parse(*l.ProjectID)
			if err == nil {
				line.ProjectID = &projID
			}
		}
		svcReq.Lines = append(svcReq.Lines, line)
	}

	voucher, err := h.voucherSvc.CreateVoucher(c.Context(), tenantID, userID, svcReq)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"voucher": voucher})
}

// UpdateRequest is the request body for updating a voucher.
type UpdateRequest struct {
	VoucherType *string       `json:"voucher_type,omitempty"`
	PostingDate *string       `json:"posting_date,omitempty"`
	Remark      *string       `json:"remark,omitempty"`
	Lines       []LineRequest `json:"lines"`
}

// Update handles PUT /api/v1/vouchers/:id
func (h *VoucherHandler) Update(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	tenantID := c.Locals("tenant_id").(uuid.UUID)
	userID := c.Locals("user_id").(uuid.UUID)

	var req UpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	svcReq := &service.UpdateVoucherRequest{
		VoucherType: req.VoucherType,
		Remark:      req.Remark,
		Lines:       make([]service.VoucherLineRequest, 0, len(req.Lines)),
	}

	if req.PostingDate != nil {
		t, err := time.Parse("2006-01-02", *req.PostingDate)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid posting_date, use YYYY-MM-DD"})
		}
		svcReq.PostingDate = &t
	}

	for _, l := range req.Lines {
		accountID, err := uuid.Parse(l.AccountID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid line account_id"})
		}
		line := service.VoucherLineRequest{
			AccountID:  accountID,
			Debit:      l.Debit,
			Credit:     l.Credit,
			PartyType:  l.PartyType,
			UserRemark: l.UserRemark,
		}
		if l.PartyID != nil {
			pid, err := uuid.Parse(*l.PartyID)
			if err == nil {
				line.PartyID = &pid
			}
		}
		if l.CostCenterID != nil {
			ccid, err := uuid.Parse(*l.CostCenterID)
			if err == nil {
				line.CostCenterID = &ccid
			}
		}
		if l.ProjectID != nil {
			projID, err := uuid.Parse(*l.ProjectID)
			if err == nil {
				line.ProjectID = &projID
			}
		}
		svcReq.Lines = append(svcReq.Lines, line)
	}

	if err := h.voucherSvc.UpdateVoucher(c.Context(), tenantID, id, userID, svcReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "voucher updated"})
}

// Delete handles DELETE /api/v1/vouchers/:id
func (h *VoucherHandler) Delete(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	tenantID := c.Locals("tenant_id").(uuid.UUID)
	userID := c.Locals("user_id").(uuid.UUID)

	if err := h.voucherSvc.DeleteVoucher(c.Context(), tenantID, id, userID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "voucher deleted"})
}
