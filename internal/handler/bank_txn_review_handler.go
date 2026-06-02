package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"huihua/finance/internal/repository"
	"huihua/finance/internal/service"
)

// BankTxnReviewHandler handles bank transaction review workflow HTTP requests.
type BankTxnReviewHandler struct {
	svc  *service.BankTxnReviewService
	repo *repository.BankTransactionRepository
}

// NewBankTxnReviewHandler creates a new BankTxnReviewHandler.
func NewBankTxnReviewHandler(
	svc *service.BankTxnReviewService,
	repo *repository.BankTransactionRepository,
) *BankTxnReviewHandler {
	return &BankTxnReviewHandler{svc: svc, repo: repo}
}

// RegisterRoutes registers the review workflow routes on the given Fiber app
// under the /api/v1/bank-transactions group.
func (h *BankTxnReviewHandler) RegisterRoutes(r *fiber.App, auth fiber.Handler) {
	g := r.Group("/api/v1/bank-transactions", auth)
	g.Get("/review-list", h.ReviewList)
	g.Get("/review-stats", h.ReviewStats)
	g.Post("/preview-draft/:id", h.PreviewDraft)
	g.Post("/submit-review", h.SubmitReview)
	g.Post("/reject-manual", h.RejectManual)
}

// ReviewList GET /api/v1/bank-transactions/review-list?status=&page=&page_size=
// Implements AC8: returns paginated bank transactions filtered by review status.
func (h *BankTxnReviewHandler) ReviewList(c *fiber.Ctx) error {
	tenantID, ok := c.Locals("tenant_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "tenant_id not found"})
	}

	status := c.Query("status", "")
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("page_size", 50)

	txns, total, err := h.repo.ListByStatus(c.Context(), tenantID, status, page, pageSize)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "failed to list transactions: " + err.Error()})
	}

	return c.JSON(fiber.Map{
		"data":      txns,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ReviewStats GET /api/v1/bank-transactions/review-stats
// Implements AC9: returns monthly_txns, pending_count, ai_processed_count, manual_pending_count.
func (h *BankTxnReviewHandler) ReviewStats(c *fiber.Ctx) error {
	tenantID, ok := c.Locals("tenant_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "tenant_id not found"})
	}

	stats, err := h.repo.GetStats(c.Context(), tenantID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get stats"})
	}

	return c.JSON(stats)
}

// PreviewDraft POST /api/v1/bank-transactions/preview-draft/:id
// Returns a draft voucher (docstatus=0) for the given classified transaction.
func (h *BankTxnReviewHandler) PreviewDraft(c *fiber.Ctx) error {
	tenantID, ok := c.Locals("tenant_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "tenant_id not found"})
	}

	txnIDStr := c.Params("id")
	txnID, err := uuid.Parse(txnIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid transaction id"})
	}

	voucher, err := h.svc.PreviewDraft(c.Context(), tenantID, txnID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if voucher == nil {
		return c.JSON(fiber.Map{"data": nil, "message": "no preview available for this transaction"})
	}

	return c.JSON(fiber.Map{"data": voucher})
}

// SubmitReviewRequest is the request body for POST /submit-review.
type SubmitReviewRequest struct {
	TxnIDs             []string                         `json:"txn_ids"`
	HumanModifiedDrafts map[string]*service.DraftContent `json:"human_modified_drafts"`
}

// SubmitReview POST /api/v1/bank-transactions/submit-review
// Implements AC5/AC6: atomically approves classified transactions, generating
// vouchers (A-class) or payment entries (B-class). Returns 4xx if any txn fails.
func (h *BankTxnReviewHandler) SubmitReview(c *fiber.Ctx) error {
	tenantID, ok := c.Locals("tenant_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "tenant_id not found"})
	}

	var req SubmitReviewRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if len(req.TxnIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "txn_ids is required"})
	}

	result, err := h.svc.SubmitReview(c.Context(), tenantID, req.TxnIDs, req.HumanModifiedDrafts)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":          err.Error(),
			"failed_txn_ids": req.TxnIDs,
		})
	}

	return c.JSON(fiber.Map{"data": result})
}

// RejectManualRequest is the request body for POST /reject-manual.
type RejectManualRequest struct {
	TxnIDs []string `json:"txn_ids"`
}

// RejectManual POST /api/v1/bank-transactions/reject-manual
// Implements AC7: moves one or more transactions back to manual_pending status.
func (h *BankTxnReviewHandler) RejectManual(c *fiber.Ctx) error {
	var req RejectManualRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if err := h.svc.RejectManual(c.Context(), req.TxnIDs); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to reject transactions"})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"rejected_count": len(req.TxnIDs),
		},
	})
}