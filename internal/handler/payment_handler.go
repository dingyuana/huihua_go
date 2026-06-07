package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/model"
	"huihua/finance/internal/repository"
	"huihua/finance/internal/service"
)

type PaymentEntryHandler struct {
	svc         *service.PaymentEntryService
	bankTxnRepo *repository.BankTransactionRepository
	invoiceSvc  *service.InvoiceService
}

func NewPaymentEntryHandler(
	svc *service.PaymentEntryService,
	bankTxnRepo *repository.BankTransactionRepository,
	invoiceSvc *service.InvoiceService,
) *PaymentEntryHandler {
	return &PaymentEntryHandler{
		svc:         svc,
		bankTxnRepo: bankTxnRepo,
		invoiceSvc:  invoiceSvc,
	}
}

type CreateFromBankTxnRequest struct {
	BankTransactionID string          `json:"bank_transaction_id"`
	PaymentType       string          `json:"payment_type"`
	PartyType         string          `json:"party_type"`
	PartyID           string          `json:"party_id"`
	PostingDate       string          `json:"posting_date"`
	// UnallocatedAmount 可选，新建时未核销金额；缺省时 service 端 fallback 到 paid_amount
	UnallocatedAmount decimal.Decimal `json:"unallocated_amount,omitempty"`
	// Currency 可选，缺省 "CNY"；V1.1 仅落库
	Currency *string `json:"currency,omitempty"`
	// ExchangeRate 可选，缺省 1.0；V1.1 仅落库
	ExchangeRate decimal.Decimal `json:"exchange_rate,omitempty"`
}

func (h *PaymentEntryHandler) CreateFromBankTransaction(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	userID := c.Locals("user_id").(uuid.UUID)

	var req CreateFromBankTxnRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	bankTxnID, err := uuid.Parse(req.BankTransactionID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid bank_transaction_id",
		})
	}

	partyID, err := uuid.Parse(req.PartyID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid party_id",
		})
	}

	postingDate, err := time.Parse("2006-01-02", req.PostingDate)
	if err != nil {
		postingDate = time.Now()
	}

	bankTxn, err := h.bankTxnRepo.GetByIDSimple(c.Context(), bankTxnID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "bank transaction not found",
		})
	}

	createReq := &service.CreatePaymentFromBankTxnRequest{
		BankTransactionID: bankTxnID,
		PaymentType:       req.PaymentType,
		PartyType:         req.PartyType,
		PartyID:           partyID,
		PostingDate:       postingDate,
		ReferenceNo:       "",
		UnallocatedAmount: req.UnallocatedAmount,
		Currency:          req.Currency,
		ExchangeRate:      req.ExchangeRate,
	}

	if bankTxn.ReferenceNo != nil {
		createReq.ReferenceNo = *bankTxn.ReferenceNo
	}
	createReq.CounterpartyName = bankTxn.CounterpartyName

	entry, err := h.svc.CreateFromBankTransaction(c.Context(), tenantID, userID, createReq, bankTxn, bankTxn.CompanyID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"payment_entry": entry,
	})
}

// Allocate handles POST /api/v1/payment-entries/:id/allocate
func (h *PaymentEntryHandler) Allocate(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	paymentEntryID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payment entry id"})
	}

	var req struct {
		Allocations []struct {
			InvoiceID       string          `json:"invoice_id"`
			AllocatedAmount decimal.Decimal `json:"allocated_amount"`
			DiscountAmount  decimal.Decimal `json:"discount_amount"`
		} `json:"allocations"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	var allocs []service.AllocationRequest
	for _, a := range req.Allocations {
		invoiceID, err := uuid.Parse(a.InvoiceID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid invoice_id: " + a.InvoiceID,
			})
		}
		allocs = append(allocs, service.AllocationRequest{
			InvoiceID:       invoiceID,
			AllocatedAmount: a.AllocatedAmount,
			DiscountAmount:  a.DiscountAmount,
		})
	}

	result, err := h.invoiceSvc.AllocateToPaymentEntry(c.Context(), tenantID, paymentEntryID, allocs)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"data": result})
}

func (h *PaymentEntryHandler) List(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	bankAccountIDStr := c.Query("bank_account_id")
	if bankAccountIDStr != "" {
		bankAccountID, err := uuid.Parse(bankAccountIDStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid bank_account_id",
			})
		}

		entries, err := h.svc.ListByBankAccount(c.Context(), tenantID, bankAccountID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.JSON(fiber.Map{
			"data": entries,
		})
	}

	entries, err := h.svc.ListPaymentEntries(c.Context(), tenantID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"data": entries,
	})
}

func (h *PaymentEntryHandler) GetByID(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid id",
		})
	}

	entry, err := h.svc.GetPaymentEntry(c.Context(), tenantID, id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "payment entry not found",
		})
	}

	return c.JSON(fiber.Map{
		"data": entry,
	})
}

func (h *PaymentEntryHandler) Update(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid id",
		})
	}

	var pe model.PaymentEntry
	if err := c.BodyParser(&pe); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}
	pe.ID = id

	if err := h.svc.UpdatePaymentEntry(c.Context(), tenantID, &pe); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "payment entry updated",
	})
}

func (h *PaymentEntryHandler) Delete(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid id",
		})
	}

	if err := h.svc.DeletePaymentEntry(c.Context(), tenantID, id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "payment entry deleted",
	})
}

// ApprovePaymentEntry handles POST /api/v1/payment-entries/:id/approve
func (h *PaymentEntryHandler) ApprovePaymentEntry(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	userID := c.Locals("user_id").(uuid.UUID)

	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid id",
		})
	}

	pair, err := h.svc.Approve(c.Context(), tenantID, id, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": pair,
	})
}
