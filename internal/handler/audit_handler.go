package handler

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"huihua/finance/internal/middleware"
	"huihua/finance/internal/model"
	"huihua/finance/internal/repository"
	"huihua/finance/internal/service"
)

// AuditHandler handles HTTP requests for audit log queries.
type AuditHandler struct {
	svc *service.AuditService
}

// NewAuditHandler creates a new AuditHandler.
func NewAuditHandler(svc *service.AuditService) *AuditHandler {
	return &AuditHandler{svc: svc}
}

// ListAuditLogs handles GET /api/v1/audit-logs
// Supports query parameters: object_type, object_id, actor_id, start_time, end_time, limit, offset
func (h *AuditHandler) ListAuditLogs(c *fiber.Ctx) error {
	tenantID := middleware.MustGetTenantID(c)

	filter := repository.AuditFilter{}

	if v := c.Query("object_type"); v != "" {
		filter.ObjectType = v
	}

	if v := c.Query("object_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid object_id format",
			})
		}
		filter.ObjectID = id
	}

	if v := c.Query("actor_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid actor_id format",
			})
		}
		filter.ActorID = id
	}

	if v := c.Query("start_time"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid start_time format (use RFC3339)",
			})
		}
		filter.StartTime = &t
	}

	if v := c.Query("end_time"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid end_time format (use RFC3339)",
			})
		}
		filter.EndTime = &t
	}

	if v := c.Query("limit"); v != "" {
		limit, err := strconv.Atoi(v)
		if err != nil || limit < 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid limit value",
			})
		}
		filter.Limit = limit
	}

	if v := c.Query("offset"); v != "" {
		offset, err := strconv.Atoi(v)
		if err != nil || offset < 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid offset value",
			})
		}
		filter.Offset = offset
	}

	logs, err := h.svc.ListByTenant(c.Context(), tenantID, filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list audit logs",
		})
	}

	if logs == nil {
		logs = make([]model.AuditLog, 0)
	}

	return c.JSON(fiber.Map{
		"data": logs,
	})
}

// GetAuditLogsByObject handles GET /api/v1/audit-logs/:object_type/:object_id
func (h *AuditHandler) GetAuditLogsByObject(c *fiber.Ctx) error {
	tenantID := middleware.MustGetTenantID(c)

	objectType := c.Params("object_type")
	if objectType == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "object_type is required",
		})
	}

	objectIDStr := c.Params("object_id")
	objectID, err := uuid.Parse(objectIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid object_id format",
		})
	}

	logs, err := h.svc.GetByObject(c.Context(), tenantID, objectType, objectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get audit logs",
		})
	}

	if logs == nil {
		logs = make([]model.AuditLog, 0)
	}

	return c.JSON(fiber.Map{
		"data": logs,
	})
}

// AuditWorkbenchHandler handles HTTP requests for audit workbench.
type AuditWorkbenchHandler struct {
	invoiceRepo   *repository.InvoiceRepository
	arInvoiceRepo *repository.ArInvoiceRepository
	journalRepo   *repository.JournalRepository
}

// NewAuditWorkbenchHandler creates a new AuditWorkbenchHandler.
func NewAuditWorkbenchHandler(
	invoiceRepo *repository.InvoiceRepository,
	arInvoiceRepo *repository.ArInvoiceRepository,
	journalRepo *repository.JournalRepository,
) *AuditWorkbenchHandler {
	return &AuditWorkbenchHandler{
		invoiceRepo:   invoiceRepo,
		arInvoiceRepo: arInvoiceRepo,
		journalRepo:   journalRepo,
	}
}

// AuditTasksResult is the response for GET /api/v1/audit/tasks
type AuditTasksResult struct {
	InvoiceDrafts []model.SalesInvoice `json:"invoice_drafts"`
	ArInvoices    []*model.ArInvoice   `json:"ar_invoices"`
	Vouchers      []model.JournalEntry `json:"vouchers"`
	Summary       AuditTaskSummary     `json:"summary"`
}

type AuditTaskSummary struct {
	InvoiceDraftCount   int `json:"invoice_draft_count"`
	ArInvoiceDraftCount int `json:"ar_invoice_draft_count"`
	VoucherDraftCount  int `json:"voucher_draft_count"`
	BlockedCount       int `json:"blocked_count"`
}

// GetAuditTasks handles GET /api/v1/audit/tasks
func (h *AuditWorkbenchHandler) GetAuditTasks(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	ctx := context.Background()

	statusDraft := "draft"

	// Invoice drafts
	invoiceDrafts, err := h.invoiceRepo.ListByTenant(ctx, tenantID, model.InvoiceFilter{Status: statusDraft})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// ArInvoice drafts
	arInvoices, err := h.arInvoiceRepo.ListByTenant(ctx, tenantID, &statusDraft)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Voucher drafts (docstatus=0) using ListVouchers
	docStatusZero := int16(0)
	vouchers, err := h.journalRepo.ListVouchers(ctx, tenantID, nil, nil, nil, &docStatusZero, nil, 50, 0)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// BlockedCount: vouchers where source_type='invoice' and the ArInvoice is still draft
	blockedCount := 0
	for _, v := range vouchers {
		if v.SourceType == "invoice" && v.SourceID != uuid.Nil {
			ar, err := h.arInvoiceRepo.GetByID(ctx, tenantID, v.SourceID)
			if err == nil && ar != nil && ar.Status == "draft" {
				blockedCount++
			}
		}
	}

	// Limit each list to 50 for performance
	limit := 50
	if len(invoiceDrafts) > limit {
		invoiceDrafts = invoiceDrafts[:limit]
	}
	if len(arInvoices) > limit {
		arInvoices = arInvoices[:limit]
	}
	if len(vouchers) > limit {
		vouchers = vouchers[:limit]
	}

	result := AuditTasksResult{
		InvoiceDrafts: invoiceDrafts,
		ArInvoices:    arInvoices,
		Vouchers:     vouchers,
		Summary: AuditTaskSummary{
			InvoiceDraftCount:   len(invoiceDrafts),
			ArInvoiceDraftCount: len(arInvoices),
			VoucherDraftCount:   len(vouchers),
			BlockedCount:       blockedCount,
		},
	}

	return c.JSON(result)
}