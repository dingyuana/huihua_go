package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"huihua/finance/internal/model"
)

type ReimbursementInvoiceLinkRepository struct {
	pool *pgxpool.Pool
}

func NewReimbursementInvoiceLinkRepository(pool *pgxpool.Pool) *ReimbursementInvoiceLinkRepository {
	return &ReimbursementInvoiceLinkRepository{pool: pool}
}

func (r *ReimbursementInvoiceLinkRepository) Create(ctx context.Context, link *model.ReimbursementInvoiceLink) error {
	link.ID = uuid.New()
	link.LinkedAt = time.Now()
	_, err := r.pool.Exec(ctx, `
		INSERT INTO reimbursement_invoice_links (id, reimbursement_id, invoice_id, invoice_type, linked_amount, linked_by, linked_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, link.ID, link.ReimbursementID, link.InvoiceID, link.InvoiceType, link.LinkedAmount, link.LinkedBy, link.LinkedAt)
	return err
}

func (r *ReimbursementInvoiceLinkRepository) ListByReimbursementID(ctx context.Context, reimbID uuid.UUID) ([]model.ReimbursementInvoiceLink, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, reimbursement_id, invoice_id, invoice_type, linked_amount, linked_by, linked_at
		FROM reimbursement_invoice_links WHERE reimbursement_id = $1
	`, reimbID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var links []model.ReimbursementInvoiceLink
	for rows.Next() {
		var l model.ReimbursementInvoiceLink
		if err := rows.Scan(&l.ID, &l.ReimbursementID, &l.InvoiceID, &l.InvoiceType, &l.LinkedAmount, &l.LinkedBy, &l.LinkedAt); err != nil {
			return nil, err
		}
		links = append(links, l)
	}
	return links, nil
}

func (r *ReimbursementInvoiceLinkRepository) Delete(ctx context.Context, reimbID, invoiceID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM reimbursement_invoice_links WHERE reimbursement_id=$1 AND invoice_id=$2`, reimbID, invoiceID)
	return err
}

func (r *ReimbursementInvoiceLinkRepository) GetByReimbursementAndInvoice(ctx context.Context, reimbID, invoiceID uuid.UUID) (*model.ReimbursementInvoiceLink, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, reimbursement_id, invoice_id, invoice_type, linked_amount, linked_by, linked_at
		FROM reimbursement_invoice_links WHERE reimbursement_id=$1 AND invoice_id=$2
	`, reimbID, invoiceID)
	var l model.ReimbursementInvoiceLink
	err := row.Scan(&l.ID, &l.ReimbursementID, &l.InvoiceID, &l.InvoiceType, &l.LinkedAmount, &l.LinkedBy, &l.LinkedAt)
	if err != nil {
		return nil, err
	}
	return &l, nil
}