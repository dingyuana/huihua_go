package service

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"huihua/finance/internal/repository"
)

// ReimbursementAttachmentService handles attachment business logic.
type ReimbursementAttachmentService struct {
	attachmentRepo *repository.ReimbursementAttachmentRepository
	reimbRepo     *repository.ReimbursementRepository
}

// NewReimbursementAttachmentService creates a new ReimbursementAttachmentService.
func NewReimbursementAttachmentService(
	attachmentRepo *repository.ReimbursementAttachmentRepository,
	reimbRepo *repository.ReimbursementRepository,
) *ReimbursementAttachmentService {
	return &ReimbursementAttachmentService{
		attachmentRepo: attachmentRepo,
		reimbRepo:     reimbRepo,
	}
}

// UploadAttachment saves an uploaded file and creates an attachment record.
func (s *ReimbursementAttachmentService) UploadAttachment(ctx context.Context, tenantID, reimbID, userID uuid.UUID, file *multipart.FileHeader) (*repository.ReimbursementAttachment, error) {
	// Verify reimbursement exists and belongs to tenant
	reimb, err := s.reimbRepo.GetByID(ctx, tenantID, reimbID)
	if err != nil {
		return nil, fmt.Errorf("reimbursement not found: %w", err)
	}
	if reimb.DocStatus != 0 {
		return nil, fmt.Errorf("attachments can only be uploaded to draft reimbursements")
	}

	// Ensure upload directory exists
	uploadDir := filepath.Join("uploads", "attachments", reimb.ReimbursementNo)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("create upload dir: %w", err)
	}

	// Open source file
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("open uploaded file: %w", err)
	}
	defer src.Close()

	// Save file
	dst := filepath.Join(uploadDir, file.Filename)
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, fmt.Errorf("open file for writing: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, src); err != nil {
		return nil, fmt.Errorf("save file: %w", err)
	}

	// Create attachment record
	attachment := &repository.ReimbursementAttachment{
		ID:              uuid.New(),
		ReimbursementID: reimbID,
		FileName:        file.Filename,
		FilePath:        dst,
		FileSize:        file.Size,
		MimeType:        strPtr(file.Header.Get("Content-Type")),
		UploadedBy:      &userID,
	}

	if err := s.attachmentRepo.Create(ctx, attachment); err != nil {
		// Clean up file on DB failure
		os.Remove(dst)
		return nil, fmt.Errorf("create attachment record: %w", err)
	}

	return attachment, nil
}

// ListAttachments retrieves all attachments for a reimbursement.
func (s *ReimbursementAttachmentService) ListAttachments(ctx context.Context, tenantID, reimbID uuid.UUID) ([]repository.ReimbursementAttachment, error) {
	// Verify reimbursement exists and belongs to tenant
	if _, err := s.reimbRepo.GetByID(ctx, tenantID, reimbID); err != nil {
		return nil, fmt.Errorf("reimbursement not found: %w", err)
	}
	return s.attachmentRepo.ListByReimbursementID(ctx, reimbID)
}

// DeleteAttachment deletes an attachment by ID.
func (s *ReimbursementAttachmentService) DeleteAttachment(ctx context.Context, tenantID, reimbID, attachmentID uuid.UUID) error {
	// Verify reimbursement belongs to tenant
	if _, err := s.reimbRepo.GetByID(ctx, tenantID, reimbID); err != nil {
		return fmt.Errorf("reimbursement not found: %w", err)
	}

	// Get attachment to find file path
	att, err := s.attachmentRepo.GetByID(ctx, attachmentID)
	if err != nil {
		return fmt.Errorf("attachment not found: %w", err)
	}
	if att.ReimbursementID != reimbID {
		return fmt.Errorf("attachment does not belong to this reimbursement")
	}

	// Delete file
	os.Remove(att.FilePath)

	// Delete DB record
	return s.attachmentRepo.Delete(ctx, attachmentID)
}

// GetAttachmentFile retrieves attachment file info for download.
func (s *ReimbursementAttachmentService) GetAttachmentFile(ctx context.Context, tenantID, reimbID, attachmentID uuid.UUID) (*repository.ReimbursementAttachment, error) {
	if _, err := s.reimbRepo.GetByID(ctx, tenantID, reimbID); err != nil {
		return nil, fmt.Errorf("reimbursement not found: %w", err)
	}

	att, err := s.attachmentRepo.GetByID(ctx, attachmentID)
	if err != nil {
		return nil, fmt.Errorf("attachment not found: %w", err)
	}
	if att.ReimbursementID != reimbID {
		return nil, fmt.Errorf("attachment does not belong to this reimbursement")
	}
	return att, nil
}

func strPtr(s string) *string {
	return &s
}