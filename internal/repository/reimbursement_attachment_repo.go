package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ReimbursementAttachment represents an attachment for a reimbursement.
type ReimbursementAttachment struct {
	ID uuid.UUID  `json:"id" db:"id"`
	ReimbursementID uuid.UUID  `json:"reimbursement_id" db:"reimbursement_id"`
	FileName         string    `json:"file_name" db:"file_name"`
	FilePath         string    `json:"file_path" db:"file_path"`
	FileSize         int64     `json:"file_size" db:"file_size"`
	MimeType         *string   `json:"mime_type,omitempty" db:"mime_type"`
	UploadedBy       *uuid.UUID `json:"uploaded_by,omitempty" db:"uploaded_by"`
	UploadedAt       time.Time `json:"uploaded_at" db:"uploaded_at"`
}

// ReimbursementAttachmentRepository provides data access for reimbursement_attachments.
type ReimbursementAttachmentRepository struct {
	pool *pgxpool.Pool
}

// NewReimbursementAttachmentRepository creates a new ReimbursementAttachmentRepository.
func NewReimbursementAttachmentRepository(pool *pgxpool.Pool) *ReimbursementAttachmentRepository {
	return &ReimbursementAttachmentRepository{pool: pool}
}

// Create inserts a new attachment record.
func (r *ReimbursementAttachmentRepository) Create(ctx context.Context, attachment *ReimbursementAttachment) error {
	if attachment.ID == uuid.Nil {
		attachment.ID = uuid.New()
	}
	if attachment.UploadedAt.IsZero() {
		attachment.UploadedAt = time.Now()
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO reimbursement_attachments (id, reimbursement_id, file_name, file_path, file_size, mime_type, uploaded_by, uploaded_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		attachment.ID, attachment.ReimbursementID, attachment.FileName, attachment.FilePath,
		attachment.FileSize, attachment.MimeType, attachment.UploadedBy, attachment.UploadedAt,
	)
	if err != nil {
		return fmt.Errorf("create reimbursement attachment: %w", err)
	}
	return nil
}

// ListByReimbursementID retrieves all attachments for a reimbursement.
func (r *ReimbursementAttachmentRepository) ListByReimbursementID(ctx context.Context, reimbursementID uuid.UUID) ([]ReimbursementAttachment, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, reimbursement_id, file_name, file_path, file_size, mime_type, uploaded_by, uploaded_at
		FROM reimbursement_attachments
		WHERE reimbursement_id = $1
		ORDER BY uploaded_at DESC`,
		reimbursementID)
	if err != nil {
		return nil, fmt.Errorf("list reimbursement attachments: %w", err)
	}
	defer rows.Close()

	var list []ReimbursementAttachment
	for rows.Next() {
		var a ReimbursementAttachment
		if err := rows.Scan(&a.ID, &a.ReimbursementID, &a.FileName, &a.FilePath, &a.FileSize, &a.MimeType, &a.UploadedBy, &a.UploadedAt); err != nil {
			return nil, fmt.Errorf("scan attachment: %w", err)
		}
		list = append(list, a)
	}
	return list, rows.Err()
}

// Delete deletes an attachment by ID.
func (r *ReimbursementAttachmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM reimbursement_attachments WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete reimbursement attachment: %w", err)
	}
	return nil
}

// GetByID retrieves an attachment by its ID.
func (r *ReimbursementAttachmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*ReimbursementAttachment, error) {
	var a ReimbursementAttachment
	err := r.pool.QueryRow(ctx, `
		SELECT id, reimbursement_id, file_name, file_path, file_size, mime_type, uploaded_by, uploaded_at
		FROM reimbursement_attachments WHERE id = $1`,
		id).Scan(&a.ID, &a.ReimbursementID, &a.FileName, &a.FilePath, &a.FileSize, &a.MimeType, &a.UploadedBy, &a.UploadedAt)
	if err != nil {
		return nil, fmt.Errorf("get attachment by id: %w", err)
	}
	return &a, nil
}