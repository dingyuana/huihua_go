package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/service"
)

type AdvanceReceiptHandler struct {
	svc *service.AdvanceReceiptService
}

func NewAdvanceReceiptHandler(svc *service.AdvanceReceiptService) *AdvanceReceiptHandler {
	return &AdvanceReceiptHandler{svc: svc}
}

type CreateAdvanceReceiptHTTP struct {
	CompanyID     string  `json:"company_id"`
	CustomerID    string  `json:"customer_id"`
	Amount        string  `json:"amount"`
	ReceivedDate  string  `json:"received_date"`
	DueDate       *string `json:"due_date,omitempty"`
	BankAccountID *string `json:"bank_account_id,omitempty"`
	ReferenceNo   *string `json:"reference_no,omitempty"`
	Remark        *string `json:"remark,omitempty"`
}

func (h *AdvanceReceiptHandler) List(c *fiber.Ctx) error {
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
			"id": a.ID, "advance_no": a.AdvanceNo, "customer_id": a.CustomerID,
			"amount": a.Amount.String(), "allocated_amount": a.AllocatedAmount.String(),
			"outstanding_amount": a.OutstandingAmount.String(),
			"received_date": a.ReceivedDate.Format("2006-01-02"),
			"due_date": formatDatePtr(a.DueDate),
			"status": a.Status, "source_type": a.SourceType,
			"bank_account_id": a.BankAccountID, "reference_no": a.ReferenceNo,
			"remark": derefStr(a.Remark), "voucher_no": derefStr(a.VoucherNo),
			"created_at": a.CreatedAt.Format("2006-01-02 15:04:05"),
			"confirmed_at": formatDateTimePtr(a.ConfirmedAt),
		}
	}
	return c.JSON(fiber.Map{"list": out, "total": len(out)})
}

func (h *AdvanceReceiptHandler) GetByID(c *fiber.Ctx) error {
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
		"id": a.ID, "advance_no": a.AdvanceNo, "customer_id": a.CustomerID,
		"amount": a.Amount.String(), "allocated_amount": a.AllocatedAmount.String(),
		"outstanding_amount": a.OutstandingAmount.String(),
		"received_date": a.ReceivedDate.Format("2006-01-02"),
		"due_date": formatDatePtr(a.DueDate),
		"status": a.Status, "source_type": a.SourceType,
		"bank_account_id": a.BankAccountID, "reference_no": derefStr(a.ReferenceNo),
		"remark": derefStr(a.Remark), "voucher_no": derefStr(a.VoucherNo),
		"created_at": a.CreatedAt.Format("2006-01-02 15:04:05"),
		"confirmed_at": formatDateTimePtr(a.ConfirmedAt),
	})
}

func (h *AdvanceReceiptHandler) Create(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	userID := c.Locals("user_id").(uuid.UUID)

	var req CreateAdvanceReceiptHTTP
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body: " + err.Error()})
	}
	companyID, _ := uuid.Parse(req.CompanyID)
	customerID, _ := uuid.Parse(req.CustomerID)
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid amount"})
	}
	receivedDate, _ := time.Parse("2006-01-02", req.ReceivedDate)
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
	svcReq := &service.CreateAdvanceReceiptRequest{
		CompanyID: companyID, CustomerID: customerID, Amount: amount,
		ReceivedDate: receivedDate, DueDate: dueDate,
		BankAccountID: bankAccountID, ReferenceNo: req.ReferenceNo, Remark: req.Remark,
	}
	a, err := h.svc.Create(c.Context(), tenantID, userID, svcReq)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"data": a})
}

func (h *AdvanceReceiptHandler) Confirm(c *fiber.Ctx) error {
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
	return c.JSON(fiber.Map{"data": a, "message": "advance receipt confirmed"})
}

func (h *AdvanceReceiptHandler) ListOutstanding(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	customerID, err := uuid.Parse(c.Query("customer_id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "customer_id required"})
	}
	list, err := h.svc.ListOutstanding(c.Context(), tenantID, customerID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	out := make([]map[string]interface{}, len(list))
	for i, a := range list {
		out[i] = map[string]interface{}{
			"id": a.ID, "advance_no": a.AdvanceNo, "amount": a.Amount.String(),
			"outstanding_amount": a.OutstandingAmount.String(),
			"received_date": a.ReceivedDate.Format("2006-01-02"),
			"status": a.Status,
		}
	}
	return c.JSON(fiber.Map{"list": out, "total": len(out)})
}

func formatDatePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}

func formatDateTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}
