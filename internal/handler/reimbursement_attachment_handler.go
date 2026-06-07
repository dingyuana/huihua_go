package handler

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"huihua/finance/internal/service"
)

// ReimbursementAttachmentHandler handles attachment HTTP requests.
type ReimbursementAttachmentHandler struct {
	svc *service.ReimbursementAttachmentService
}

// NewReimbursementAttachmentHandler creates a new ReimbursementAttachmentHandler.
func NewReimbursementAttachmentHandler(svc *service.ReimbursementAttachmentService) *ReimbursementAttachmentHandler {
	return &ReimbursementAttachmentHandler{svc: svc}
}

// UploadAttachment handles POST /api/v1/reimbursements/:id/attachments
func (h *ReimbursementAttachmentHandler) UploadAttachment(c *fiber.Ctx) error {
	idStr := c.Params("id")
	reimbID, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid reimbursement id"})
	}

	tenantID := c.Locals("tenant_id").(uuid.UUID)
	userID := c.Locals("user_id").(uuid.UUID)

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "file required: " + err.Error()})
	}

	attachment, err := h.svc.UploadAttachment(c.Context(), tenantID, reimbID, userID, file)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"attachment": attachment})
}

// ListAttachments handles GET /api/v1/reimbursements/:id/attachments
func (h *ReimbursementAttachmentHandler) ListAttachments(c *fiber.Ctx) error {
	idStr := c.Params("id")
	reimbID, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid reimbursement id"})
	}

	tenantID := c.Locals("tenant_id").(uuid.UUID)

	list, err := h.svc.ListAttachments(c.Context(), tenantID, reimbID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"attachments": list})
}

// DeleteAttachment handles DELETE /api/v1/reimbursements/:id/attachments/:file_id
func (h *ReimbursementAttachmentHandler) DeleteAttachment(c *fiber.Ctx) error {
	reimbIDStr := c.Params("id")
	reimbID, err := uuid.Parse(reimbIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid reimbursement id"})
	}

	fileIDStr := c.Params("file_id")
	fileID, err := uuid.Parse(fileIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid file id"})
	}

	tenantID := c.Locals("tenant_id").(uuid.UUID)

	if err := h.svc.DeleteAttachment(c.Context(), tenantID, reimbID, fileID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "attachment deleted"})
}

// DownloadAttachment handles GET /api/v1/reimbursements/:id/attachments/:file_id/download
func (h *ReimbursementAttachmentHandler) DownloadAttachment(c *fiber.Ctx) error {
	reimbIDStr := c.Params("id")
	reimbID, err := uuid.Parse(reimbIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid reimbursement id"})
	}

	fileIDStr := c.Params("file_id")
	fileID, err := uuid.Parse(fileIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid file id"})
	}

	tenantID := c.Locals("tenant_id").(uuid.UUID)

	att, err := h.svc.GetAttachmentFile(c.Context(), tenantID, reimbID, fileID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	if _, err := os.Stat(att.FilePath); os.IsNotExist(err) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "file not found on disk"})
	}

	c.Set("Content-Disposition", "attachment; filename=\""+att.FileName+"\"")
	if att.MimeType != nil {
		c.Set("Content-Type", *att.MimeType)
	}
	return c.SendFile(att.FilePath)
}