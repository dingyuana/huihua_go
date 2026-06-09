package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/model"
)

// SettlementLogRepository manages the immutable settlement_logs table.
type SettlementLogRepository struct {
	pool *pgxpool.Pool
}

// NewSettlementLogRepository creates a new SettlementLogRepository.
func NewSettlementLogRepository(pool *pgxpool.Pool) *SettlementLogRepository {
	return &SettlementLogRepository{pool: pool}
}

// Create inserts a settlement log entry within a transaction.
func (r *SettlementLogRepository) Create(ctx context.Context, tx pgx.Tx, log *model.SettlementLog) error {
	log.ID = uuid.New()
	log.CreatedAt = time.Now()
	_, err := tx.Exec(ctx, `
		INSERT INTO settlement_logs (
			id, tenant_id, source_type, source_id,
			doc_type, doc_id, direction,
			amount, outstanding_before, outstanding_after,
			is_reversal, reversed_log_id, created_by, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		log.ID, log.TenantID, log.SourceType, log.SourceID,
		log.DocType, log.DocID, log.Direction,
		log.Amount.String(), log.OutstandingBefore.String(), log.OutstandingAfter.String(),
		log.IsReversal, log.ReversedLogID, log.CreatedBy, log.CreatedAt)
	return err
}

// ListByDoc retrieves all settlement logs for a given document, ordered by time.
func (r *SettlementLogRepository) ListByDoc(ctx context.Context, tenantID uuid.UUID, docType model.SettlementLogDocType, docID uuid.UUID) ([]*model.SettlementLog, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, source_type, source_id,
		       doc_type, doc_id, direction,
		       amount, outstanding_before, outstanding_after,
		       is_reversal, reversed_log_id, created_by, created_at
		FROM settlement_logs
		WHERE tenant_id = $1 AND doc_type = $2 AND doc_id = $3
		ORDER BY created_at ASC`,
		tenantID, docType, docID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*model.SettlementLog
	for rows.Next() {
		var l model.SettlementLog
		var amount, before, after string
		if err := rows.Scan(
			&l.ID, &l.TenantID, &l.SourceType, &l.SourceID,
			&l.DocType, &l.DocID, &l.Direction,
			&amount, &before, &after,
			&l.IsReversal, &l.ReversedLogID, &l.CreatedBy, &l.CreatedAt,
		); err != nil {
			return nil, err
		}
		l.Amount, _ = decimal.NewFromString(amount)
		l.OutstandingBefore, _ = decimal.NewFromString(before)
		l.OutstandingAfter, _ = decimal.NewFromString(after)
		logs = append(logs, &l)
	}
	return logs, rows.Err()
}

// LogWriteOff creates a settlement log entry for a write-off (allocation).
// direction: 'debit' for AR reduction, 'credit' for AP reduction.
func LogWriteOff(
	ctx context.Context, tx pgx.Tx,
	repo *SettlementLogRepository,
	tenantID, sourceID, docID uuid.UUID,
	sourceType model.SettlementLogSourceType,
	docType model.SettlementLogDocType,
	direction model.SettlementLogDirection,
	amount decimal.Decimal,
	outstandingBefore decimal.Decimal,
	outstandingAfter decimal.Decimal,
	createdBy *uuid.UUID,
) error {
	log := &model.SettlementLog{
		TenantID:          tenantID,
		SourceType:        sourceType,
		SourceID:          sourceID,
		DocType:           docType,
		DocID:             docID,
		Direction:         direction,
		Amount:            amount,
		OutstandingBefore: outstandingBefore,
		OutstandingAfter:  outstandingAfter,
		IsReversal:        false,
		CreatedBy:         createdBy,
	}
	return repo.Create(ctx, tx, log)
}

// LogReversal creates a settlement log entry for a reversal (rollback).
func LogReversal(
	ctx context.Context, tx pgx.Tx,
	repo *SettlementLogRepository,
	tenantID, sourceID, docID uuid.UUID,
	sourceType model.SettlementLogSourceType,
	docType model.SettlementLogDocType,
	direction model.SettlementLogDirection,
	amount decimal.Decimal,
	outstandingBefore decimal.Decimal,
	outstandingAfter decimal.Decimal,
	reversedLogID uuid.UUID,
	createdBy *uuid.UUID,
) error {
	log := &model.SettlementLog{
		TenantID:          tenantID,
		SourceType:        sourceType,
		SourceID:          sourceID,
		DocType:           docType,
		DocID:             docID,
		Direction:         direction,
		Amount:            amount,
		OutstandingBefore: outstandingBefore,
		OutstandingAfter:  outstandingAfter,
		IsReversal:        true,
		ReversedLogID:     &reversedLogID,
		CreatedBy:         createdBy,
	}
	return repo.Create(ctx, tx, log)
}
