package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/service"
)

type WriteOffHandler struct {
	svc *service.WriteOffService
}

func NewWriteOffHandler(svc *service.WriteOffService) *WriteOffHandler {
	return &WriteOffHandler{svc: svc}
}

func (h *WriteOffHandler) AutoWriteOff(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	var req struct {
		DocumentType   string `json:"document_type"`
		StartDate      string `json:"start_date,omitempty"`
		EndDate        string `json:"end_date,omitempty"`
		CounterpartyID string `json:"counterparty_id,omitempty"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.DocumentType == "" {
		return c.Status(400).JSON(fiber.Map{"error": "document_type is required"})
	}

	var opts service.AutoWriteOffOptions
	opts.DocumentType = req.DocumentType

	if req.StartDate != "" {
		startDate, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid start_date format"})
		}
		opts.StartDate = startDate
	}

	if req.EndDate != "" {
		endDate, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid end_date format"})
		}
		opts.EndDate = endDate
	}

	if req.CounterpartyID != "" {
		cpID, err := uuid.Parse(req.CounterpartyID)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid counterparty_id"})
		}
		opts.CounterpartyID = cpID
	}

	result, err := h.svc.AutoWriteOff(c.Context(), tenantID, opts)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"data": result})
}

func (h *WriteOffHandler) ManualWriteOff(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	operatorID := c.Locals("user_id").(uuid.UUID)

	var req struct {
		ReceiptPaymentID      string `json:"receipt_payment_id"`
		ReceivablePayableID   string `json:"receivable_payable_id"`
		ReceivablePayableType string `json:"receivable_payable_type"`
		Amount                string `json:"amount"`
		Remark                string `json:"remark,omitempty"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.ReceiptPaymentID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "receipt_payment_id is required"})
	}
	if req.ReceivablePayableID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "receivable_payable_id is required"})
	}
	if req.ReceivablePayableType == "" {
		return c.Status(400).JSON(fiber.Map{"error": "receivable_payable_type is required"})
	}
	if req.Amount == "" {
		return c.Status(400).JSON(fiber.Map{"error": "amount is required"})
	}

	rpID, err := uuid.Parse(req.ReceiptPaymentID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid receipt_payment_id"})
	}

	rpTypeID, err := uuid.Parse(req.ReceivablePayableID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid receivable_payable_id"})
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid amount"})
	}

	record, err := h.svc.ManualWriteOff(c.Context(), tenantID, operatorID, service.ManualWriteOffRequest{
		ReceiptPaymentID:      rpID,
		ReceivablePayableID:   rpTypeID,
		ReceivablePayableType: req.ReceivablePayableType,
		Amount:               amount,
		Remark:               req.Remark,
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{"data": record})
}

func (h *WriteOffHandler) SubmitApproval(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	operatorID := c.Locals("user_id").(uuid.UUID)

	recordIDStr := c.Params("id")
	recordID, err := strconv.ParseInt(recordIDStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid record id"})
	}

	var req struct {
		Remark string `json:"remark,omitempty"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	if err := h.svc.SubmitApproval(c.Context(), tenantID, operatorID, service.SubmitApprovalRequest{
		RecordID: recordID,
		Remark:   req.Remark,
	}); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "submitted"})
}

func (h *WriteOffHandler) Approve(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	approverID := c.Locals("user_id").(uuid.UUID)

	recordIDStr := c.Params("id")
	recordID, err := strconv.ParseInt(recordIDStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid record id"})
	}

	if err := h.svc.Approve(c.Context(), tenantID, approverID, service.ApproveRequest{
		RecordID: recordID,
	}); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "approved"})
}

func (h *WriteOffHandler) Reject(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	approverID := c.Locals("user_id").(uuid.UUID)

	recordIDStr := c.Params("id")
	recordID, err := strconv.ParseInt(recordIDStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid record id"})
	}

	var req struct {
		RejectReason string `json:"reject_reason"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.RejectReason == "" {
		return c.Status(400).JSON(fiber.Map{"error": "reject_reason is required"})
	}

	if err := h.svc.Reject(c.Context(), tenantID, approverID, service.RejectRequest{
		RecordID:     recordID,
		RejectReason: req.RejectReason,
	}); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "rejected"})
}

func (h *WriteOffHandler) UpdateDraft(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	recordIDStr := c.Params("id")
	recordID, err := strconv.ParseInt(recordIDStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid record id"})
	}

	var req struct {
		ReceiptPaymentID      string `json:"receipt_payment_id,omitempty"`
		ReceivablePayableID   string `json:"receivable_payable_id"`
		ReceivablePayableType string `json:"receivable_payable_type"`
		Amount                string `json:"amount"`
		Remark                string `json:"remark,omitempty"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.ReceivablePayableID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "receivable_payable_id is required"})
	}
	if req.ReceivablePayableType == "" {
		return c.Status(400).JSON(fiber.Map{"error": "receivable_payable_type is required"})
	}
	if req.Amount == "" {
		return c.Status(400).JSON(fiber.Map{"error": "amount is required"})
	}

	rpTypeID, err := uuid.Parse(req.ReceivablePayableID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid receivable_payable_id"})
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid amount"})
	}

	updateReq := service.ManualWriteOffRequest{
		ReceivablePayableID:   rpTypeID,
		ReceivablePayableType: req.ReceivablePayableType,
		Amount:               amount,
		Remark:               req.Remark,
	}

	if req.ReceiptPaymentID != "" {
		rpID, err := uuid.Parse(req.ReceiptPaymentID)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid receipt_payment_id"})
		}
		updateReq.ReceiptPaymentID = rpID
	}

	if err := h.svc.UpdateDraft(c.Context(), tenantID, recordID, updateReq); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "updated"})
}

func (h *WriteOffHandler) DeleteDraft(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	recordIDStr := c.Params("id")
	recordID, err := strconv.ParseInt(recordIDStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid record id"})
	}

	if err := h.svc.DeleteDraft(c.Context(), tenantID, recordID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "deleted"})
}

func (h *WriteOffHandler) ListRecords(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	params := make(map[string]interface{})

	if status := c.Query("status"); status != "" {
		params["status"] = status
	}
	if documentType := c.Query("document_type"); documentType != "" {
		params["document_type"] = documentType
	}
	if startDate := c.Query("start_date"); startDate != "" {
		params["start_date"] = startDate
	}
	if endDate := c.Query("end_date"); endDate != "" {
		params["end_date"] = endDate
	}

	records, err := h.svc.GetWriteOffRecords(c.Context(), tenantID, params)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"data": records, "count": len(records)})
}

func (h *WriteOffHandler) GetRecord(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	recordIDStr := c.Params("id")
	recordID, err := strconv.ParseInt(recordIDStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid record id"})
	}

	records, err := h.svc.GetWriteOffRecords(c.Context(), tenantID, map[string]interface{}{"id": recordID})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	if len(records) == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "record not found"})
	}

	return c.JSON(fiber.Map{"data": records[0]})
}

func (h *WriteOffHandler) ReverseWriteOff(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	operatorID := c.Locals("user_id").(uuid.UUID)

	recordIDStr := c.Params("id")
	recordID, err := strconv.ParseInt(recordIDStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid record id"})
	}

	if err := h.svc.ReverseWriteOff(c.Context(), tenantID, operatorID, recordID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "reversed"})
}

func (h *WriteOffHandler) GetUnmatchedSummary(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	summary, err := h.svc.GetUnmatchedSummary(c.Context(), tenantID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"data": summary})
}

func (h *WriteOffHandler) GetUnmatchedItems(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	counterpartyID := c.Query("counterparty_id")
	if counterpartyID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "counterparty_id is required"})
	}

	cpID, err := uuid.Parse(counterpartyID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid counterparty_id"})
	}

	items, err := h.svc.GetUnmatchedItems(c.Context(), tenantID, cpID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"data": items, "count": len(items)})
}