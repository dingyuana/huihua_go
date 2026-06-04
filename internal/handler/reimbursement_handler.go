package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"huihua/finance/internal/service"
)

// ReimbursementHandler handles reimbursement HTTP requests.
type ReimbursementHandler struct {
	svc *service.ReimbursementService
}

// NewReimbursementHandler creates a new ReimbursementHandler.
func NewReimbursementHandler(svc *service.ReimbursementService) *ReimbursementHandler {
	return &ReimbursementHandler{svc: svc}
}

// ReimbursementCreateRequest is the request body for creating a reimbursement.
type ReimbursementCreateRequest struct {
	EmployeeName string  `json:"employee_name"`
	Department   *string `json:"department,omitempty"`
	ExpenseType  string  `json:"expense_type"`
	Amount       string  `json:"amount"`
	PostingDate  string  `json:"posting_date"`
	Description *string `json:"description,omitempty"`
	BankAccount *string `json:"bank_account,omitempty"`
	CompanyID   string  `json:"company_id"`
}

// Create handles POST /api/v1/reimbursements
func (h *ReimbursementHandler) Create(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	userID := c.Locals("user_id").(uuid.UUID)

	var req ReimbursementCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	reimb, err := h.svc.CreateReimbursement(c.Context(), tenantID, userID, &service.CreateReimbursementRequest{
		EmployeeName: req.EmployeeName,
		Department:   req.Department,
		ExpenseType:  req.ExpenseType,
		Amount:       req.Amount,
		PostingDate:  req.PostingDate,
		Description: req.Description,
		BankAccount: req.BankAccount,
		CompanyID:   req.CompanyID,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"reimbursement": reimb})
}

// List handles GET /api/v1/reimbursements
func (h *ReimbursementHandler) List(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	var status *int16
	if s := c.Query("status"); s != "" {
		if parsed := c.QueryInt("status"); parsed > 0 {
			p := int16(parsed)
			status = &p
		}
	}

	list, err := h.svc.ListReimbursements(c.Context(), tenantID, status)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"reimbursements": list})
}

// GetByID handles GET /api/v1/reimbursements/:id
func (h *ReimbursementHandler) GetByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	tenantID := c.Locals("tenant_id").(uuid.UUID)

	reimb, err := h.svc.GetReimbursement(c.Context(), tenantID, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(reimb)
}

// Update handles PUT /api/v1/reimbursements/:id
func (h *ReimbursementHandler) Update(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	tenantID := c.Locals("tenant_id").(uuid.UUID)

	var req ReimbursementCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if err := h.svc.UpdateReimbursement(c.Context(), tenantID, id, &service.CreateReimbursementRequest{
		EmployeeName: req.EmployeeName,
		Department:   req.Department,
		ExpenseType:  req.ExpenseType,
		Amount:       req.Amount,
		PostingDate:  req.PostingDate,
		Description: req.Description,
		BankAccount: req.BankAccount,
		CompanyID:   req.CompanyID,
	}); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "reimbursement updated"})
}

// Delete handles DELETE /api/v1/reimbursements/:id
func (h *ReimbursementHandler) Delete(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	tenantID := c.Locals("tenant_id").(uuid.UUID)

	if err := h.svc.DeleteReimbursement(c.Context(), tenantID, id); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "reimbursement deleted"})
}

// Submit handles POST /api/v1/reimbursements/:id/submit
func (h *ReimbursementHandler) Submit(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	tenantID := c.Locals("tenant_id").(uuid.UUID)
	userID := c.Locals("user_id").(uuid.UUID)

	if err := h.svc.Submit(c.Context(), tenantID, id, userID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "reimbursement submitted"})
}

// Approve handles POST /api/v1/reimbursements/:id/approve
func (h *ReimbursementHandler) Approve(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	tenantID := c.Locals("tenant_id").(uuid.UUID)
	userID := c.Locals("user_id").(uuid.UUID)

	voucher, err := h.svc.Approve(c.Context(), tenantID, id, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message":         "reimbursement approved",
		"voucher_id":      voucher.ID,
		"voucher_no":      voucher.VoucherNo,
	})
}

// GenerateVoucher handles POST /api/v1/reimbursements/:id/generate-voucher
func (h *ReimbursementHandler) GenerateVoucher(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	tenantID := c.Locals("tenant_id").(uuid.UUID)
	userID := c.Locals("user_id").(uuid.UUID)

	voucher, err := h.svc.GenerateVoucher(c.Context(), tenantID, id, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"voucher_id": voucher.ID,
		"voucher_no": voucher.VoucherNo,
	})
}