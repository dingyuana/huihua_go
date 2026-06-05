package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/service"
)

type AdvancePaymentHandler struct {
	svc *service.AdvancePaymentService
}

func NewAdvancePaymentHandler(svc *service.AdvancePaymentService) *AdvancePaymentHandler {
	return &AdvancePaymentHandler{svc: svc}
}

type CreateAdvancePaymentHTTP struct {
	CompanyID     string  `json:"company_id"`
	SupplierID    string  `json:"supplier_id"`
	Amount        string  `json:"amount"`
	PaidDate      string  `json:"paid_date"`
	DueDate       *string `json:"due_date,omitempty"`
	BankAccountID *string `json:"bank_account_id,omitempty"`
	ReferenceNo   *string `json:"reference_no,omitempty"`
	Remark        *string `json:"remark,omitempty"`
}

func (h *AdvancePaymentHandler) List(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	var status *string
	if v := c.Query("status"); v != "" {
		status = &v
	}
	list, err := h.svc.List(c.Context(), tenantID, status)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	out := make([]map[string]interface{}, len(list))
	for i, a := range list {
		out[i] = map[string]interface{}{
			"id": a.ID, "advance_no": a.AdvanceNo, "supplier_id": a.SupplierID,
			"amount": a.Amount.String(), "allocated_amount": a.AllocatedAmount.String(),
			"outstanding_amount": a.OutstandingAmount.String(),
			"paid_date": a.PaidDate.Format("2006-01-02"),
			"due_date": formatDatePtr(a.DueDate),
			"status": a.Status, "source_type": a.SourceType,
			"bank_account_id": a.BankAccountID, "reference_no": derefStr(a.ReferenceNo),
			"remark": derefStr(a.Remark), "voucher_no": derefStr(a.VoucherNo),
			"created_at": a.CreatedAt.Format("2006-01-02 15:04:05"),
			"confirmed_at": formatDateTimePtr(a.ConfirmedAt),
		}
	}
	return c.JSON(fiber.Map{"list": out, "total": len(out)})
}

func (h *AdvancePaymentHandler) GetByID(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}
	a, err := h.svc.GetByID(c.Context(), tenantID, id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if a == nil {
		return c.Status(404).JSON(fiber.Map{"error": "not found"})
	}
	return c.JSON(fiber.Map{
		"id": a.ID, "advance_no": a.AdvanceNo, "supplier_id": a.SupplierID,
		"amount": a.Amount.String(), "allocated_amount": a.AllocatedAmount.String(),
		"outstanding_amount": a.OutstandingAmount.String(),
		"paid_date": a.PaidDate.Format("2006-01-02"),
		"due_date": formatDatePtr(a.DueDate),
		"status": a.Status, "source_type": a.SourceType,
		"bank_account_id": a.BankAccountID, "reference_no": derefStr(a.ReferenceNo),
		"remark": derefStr(a.Remark), "voucher_no": derefStr(a.VoucherNo),
		"created_at": a.CreatedAt.Format("2006-01-02 15:04:05"),
		"confirmed_at": formatDateTimePtr(a.ConfirmedAt),
	})
}

func (h *AdvancePaymentHandler) Create(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	userID := c.Locals("user_id").(uuid.UUID)

	var req CreateAdvancePaymentHTTP
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body: " + err.Error()})
	}
	companyID, _ := uuid.Parse(req.CompanyID)
	supplierID, _ := uuid.Parse(req.SupplierID)
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid amount"})
	}
	paidDate, _ := time.Parse("2006-01-02", req.PaidDate)
	var dueDate *time.Time
	if req.DueDate != nil && *req.DueDate != "" {
		t, _ := time.Parse("2006-01-02", *req.DueDate)
		dueDate = &t
	}
	var bankAccountID *uuid.UUID
	if req.BankAccountID != nil && *req.BankAccountID != "" {
		id, _ := uuid.Parse(*req.BankAccountID)
		bankAccountID = &id
	}
	svcReq := &service.CreateAdvancePaymentRequest{
		CompanyID: companyID, SupplierID: supplierID, Amount: amount,
		PaidDate: paidDate, DueDate: dueDate,
		BankAccountID: bankAccountID, ReferenceNo: req.ReferenceNo, Remark: req.Remark,
	}
	a, err := h.svc.Create(c.Context(), tenantID, userID, svcReq)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"data": a})
}

func (h *AdvancePaymentHandler) Confirm(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	userID := c.Locals("user_id").(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}
	a, err := h.svc.Confirm(c.Context(), tenantID, userID, id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": a, "message": "advance payment confirmed"})
}

func (h *AdvancePaymentHandler) ListOutstanding(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	supplierID, err := uuid.Parse(c.Query("supplier_id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "supplier_id required"})
	}
	list, err := h.svc.ListOutstanding(c.Context(), tenantID, supplierID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	out := make([]map[string]interface{}, len(list))
	for i, a := range list {
		out[i] = map[string]interface{}{
			"id": a.ID, "advance_no": a.AdvanceNo, "amount": a.Amount.String(),
			"outstanding_amount": a.OutstandingAmount.String(),
			"paid_date": a.PaidDate.Format("2006-01-02"),
			"status": a.Status,
		}
	}
	return c.JSON(fiber.Map{"list": out, "total": len(out)})
}
