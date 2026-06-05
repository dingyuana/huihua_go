package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/model"
)

// ReconciliationRepository handles reconciliation_pairs table.
type ReconciliationRepository struct {
	pool *pgxpool.Pool
}

// NewReconciliationRepository creates a new ReconciliationRepository.
func NewReconciliationRepository(pool *pgxpool.Pool) *ReconciliationRepository {
	return &ReconciliationRepository{pool: pool}
}

// Create inserts a new reconciliation pair.
func (r *ReconciliationRepository) Create(ctx context.Context, pair *model.ReconciliationPair) error {
	query := `
		INSERT INTO reconciliation_pairs (id, tenant_id, source_type, source_id, target_type, target_id, amount, status, match_level, matched_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	_, err := r.pool.Exec(ctx, query, pair.ID, pair.TenantID, pair.SourceType, pair.SourceID,
		pair.TargetType, pair.TargetID, pair.Amount, pair.Status, pair.MatchLevel, pair.MatchedAt, pair.CreatedAt)
	return err
}

// CreateBatch inserts multiple reconciliation pairs in a transaction.
func (r *ReconciliationRepository) CreateBatch(ctx context.Context, pairs []model.ReconciliationPair) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	for _, p := range pairs {
		if err := r.Create(ctx, &p); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// ListByTenant returns all pairs for a tenant.
func (r *ReconciliationRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, status string) ([]model.ReconciliationPair, error) {
	query := `
		SELECT id, tenant_id, source_type, source_id, target_type, target_id, amount, status, match_level, matched_at, confirmed_at, created_at
		FROM reconciliation_pairs WHERE tenant_id = $1`
	args := []interface{}{tenantID}
	if status != "" {
		query += " AND status = $2"
		args = append(args, status)
	}
	query += " ORDER BY created_at DESC"
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var pairs []model.ReconciliationPair
	for rows.Next() {
		var p model.ReconciliationPair
		rows.Scan(&p.ID, &p.TenantID, &p.SourceType, &p.SourceID, &p.TargetType, &p.TargetID,
			&p.Amount, &p.Status, &p.MatchLevel, &p.MatchedAt, &p.ConfirmedAt, &p.CreatedAt)
		pairs = append(pairs, p)
	}
	return pairs, nil
}

// UpdateStatus updates the status of a reconciliation pair.
func (r *ReconciliationRepository) UpdateStatus(ctx context.Context, tenantID, pairID uuid.UUID, status string) error {
	query := `UPDATE reconciliation_pairs SET status = $3, confirmed_at = $4 WHERE id = $1 AND tenant_id = $2`
	_, err := r.pool.Exec(ctx, query, pairID, tenantID, status, time.Now())
	return err
}

// GetByID returns a single pair.
func (r *ReconciliationRepository) GetByID(ctx context.Context, tenantID, pairID uuid.UUID) (*model.ReconciliationPair, error) {
	query := `
		SELECT id, tenant_id, source_type, source_id, target_type, target_id, amount, status, match_level, matched_at, confirmed_at, created_at
		FROM reconciliation_pairs WHERE id = $1 AND tenant_id = $2`
	var p model.ReconciliationPair
	err := r.pool.QueryRow(ctx, query, pairID, tenantID).Scan(
		&p.ID, &p.TenantID, &p.SourceType, &p.SourceID, &p.TargetType, &p.TargetID,
		&p.Amount, &p.Status, &p.MatchLevel, &p.MatchedAt, &p.ConfirmedAt, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// ConfirmPair confirms a pending reconciliation pair and deducts invoice outstanding amount.
func (r *ReconciliationRepository) ConfirmPair(ctx context.Context, tenantID, pairID uuid.UUID) error {
	// 1. Load pair and verify status="pending"
	pair, err := r.GetByID(ctx, tenantID, pairID)
	if err != nil {
		return fmt.Errorf("load pair: %w", err)
	}
	if pair.Status != "pending" {
		return fmt.Errorf("pair status must be pending, got %s", pair.Status)
	}

	// 2. Update pair status to "confirmed"
	now := time.Now()
	_, err = r.pool.Exec(ctx, `
		UPDATE reconciliation_pairs SET status = 'confirmed', confirmed_at = $3 WHERE id = $1 AND tenant_id = $2`,
		pairID, tenantID, now)
	if err != nil {
		return fmt.Errorf("update pair status: %w", err)
	}

	// 3. Load payment_allocation to get allocated_amount and invoice_id
	var allocAmt decimal.Decimal
	var invoiceID uuid.UUID
	err = r.pool.QueryRow(ctx, `
		SELECT allocated_amount, invoice_id FROM payment_allocations
		WHERE payment_entry_id = $1 AND invoice_id = $2 AND tenant_id = $3`,
		pair.SourceID, pair.TargetID, tenantID,
	).Scan(&allocAmt, &invoiceID)
	if err != nil {
		return fmt.Errorf("load payment allocation: %w", err)
	}

	// 4. Update outstanding_amount on the correct invoice table based on target_type
	switch pair.TargetType {
	case "ar_invoice":
		// Update ar_invoices: paid_amount += allocAmt, outstanding_amount = amount - paid_amount
		_, err = r.pool.Exec(ctx, `
			UPDATE ar_invoices
			SET paid_amount = paid_amount + $1,
				outstanding_amount = amount - (paid_amount + $1)
			WHERE id = $2 AND tenant_id = $3`,
			allocAmt, invoiceID, tenantID)
	case "invoice", "sales_invoice":
		// Update sales_invoices: outstanding_amount -= allocAmt
		_, err = r.pool.Exec(ctx, `
			UPDATE invoices SET outstanding_amount = outstanding_amount - $1 WHERE id = $2 AND tenant_id = $3`,
			allocAmt, invoiceID, tenantID)
	}
	if err != nil {
		return fmt.Errorf("update invoice outstanding: %w", err)
	}

	// 5. Update payment_allocation confirmed_at
	_, err = r.pool.Exec(ctx, `
		UPDATE payment_allocations SET confirmed_at = $3
		WHERE payment_entry_id = $1 AND invoice_id = $2 AND tenant_id = $3`,
		pair.SourceID, pair.TargetID, tenantID, now)
	if err != nil {
		return fmt.Errorf("update payment allocation: %w", err)
	}

	return nil
}
