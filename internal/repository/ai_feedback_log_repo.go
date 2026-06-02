package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/pgtype"
	"huihua/finance/internal/model"
)

// AIFeedbackLogRepository handles persistence for ai_feedback_logs.
type AIFeedbackLogRepository struct {
	pool *pgxpool.Pool
}

// NewAIFeedbackLogRepository creates a new AIFeedbackLogRepository.
func NewAIFeedbackLogRepository(pool *pgxpool.Pool) *AIFeedbackLogRepository {
	return &AIFeedbackLogRepository{pool: pool}
}

// Pool returns the underlying pgxpool.Pool.
func (r *AIFeedbackLogRepository) Pool() *pgxpool.Pool {
	return r.pool
}

// Create inserts a new AI feedback log entry.
func (r *AIFeedbackLogRepository) Create(ctx context.Context, log *model.AIFeedbackLog) error {
	return r.createImpl(ctx, nil, log)
}

// CreateTx inserts a new AI feedback log entry within an existing transaction.
func (r *AIFeedbackLogRepository) CreateTx(ctx context.Context, tx pgx.Tx, log *model.AIFeedbackLog) error {
	return r.createImpl(ctx, tx, log)
}

func (r *AIFeedbackLogRepository) createImpl(ctx context.Context, tx pgx.Tx, log *model.AIFeedbackLog) error {
	var modifiedFieldsJSON []byte
	if log.HumanModifiedFields != nil {
		var err error
		modifiedFieldsJSON, err = json.Marshal(log.HumanModifiedFields)
		if err != nil {
			return fmt.Errorf("marshal human_modified_fields: %w", err)
		}
	}

	query := `
		INSERT INTO ai_feedback_logs
			(id, tenant_id, bank_txn_id, ai_suggested_action, ai_confidence,
			 ai_business_scene, human_action, human_modified_fields, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	var err error
	if tx != nil {
		_, err = tx.Exec(ctx, query,
			log.ID, log.TenantID, log.BankTxnID,
			log.AISuggestedAction, log.AIConfidence,
			log.AIBusinessScene, log.HumanAction,
			modifiedFieldsJSON, log.CreatedBy, log.CreatedAt,
		)
	} else {
		_, err = r.pool.Exec(ctx, query,
			log.ID, log.TenantID, log.BankTxnID,
			log.AISuggestedAction, log.AIConfidence,
			log.AIBusinessScene, log.HumanAction,
			modifiedFieldsJSON, log.CreatedBy, log.CreatedAt,
		)
	}
	if err != nil {
		return fmt.Errorf("create ai_feedback_log: %w", err)
	}
	return nil
}

// FetchAIFields reads AI suggestion fields from the bank_transactions table.
// NULL AI fields result in nil pointers being written to the output args so
// that the AIFeedbackLog model correctly reflects "no AI suggestion".
func (r *AIFeedbackLogRepository) FetchAIFields(ctx context.Context, bankTxnID uuid.UUID, tenantID *uuid.UUID, aiSuggestedAction, aiBusinessScene *string, aiConfidence *int) error {
	var suggestedAction, businessScene          pgtype.Text
	var confidence                             pgtype.Int4
	err := r.pool.QueryRow(ctx, `
		SELECT tenant_id, ai_suggested_action, ai_confidence, ai_business_scene
		FROM bank_transactions WHERE id = $1`,
		bankTxnID).Scan(tenantID, &suggestedAction, &confidence, &businessScene)
	if err != nil {
		return fmt.Errorf("fetchAIFields: %w", err)
	}
	if suggestedAction.Valid {
		*aiSuggestedAction = suggestedAction.String
	}
	if businessScene.Valid {
		*aiBusinessScene = businessScene.String
	}
	if confidence.Valid {
		*aiConfidence = int(confidence.Int32)
	}
	return nil
}

// ListByTxnID returns all feedback log entries for a given bank transaction.
func (r *AIFeedbackLogRepository) ListByTxnID(ctx context.Context, bankTxnID uuid.UUID) ([]model.AIFeedbackLog, error) {
	query := `
		SELECT id, tenant_id, bank_txn_id, ai_suggested_action, ai_confidence,
			   ai_business_scene, human_action, human_modified_fields, created_by, created_at
		FROM ai_feedback_logs
		WHERE bank_txn_id = $1
		ORDER BY created_at ASC`
	rows, err := r.pool.Query(ctx, query, bankTxnID)
	if err != nil {
		return nil, fmt.Errorf("list by txn id: %w", err)
	}
	defer rows.Close()

	var logs []model.AIFeedbackLog
	for rows.Next() {
		var log model.AIFeedbackLog
		var modifiedFieldsJSON []byte
		err := rows.Scan(
			&log.ID, &log.TenantID, &log.BankTxnID,
			&log.AISuggestedAction, &log.AIConfidence,
			&log.AIBusinessScene, &log.HumanAction,
			&modifiedFieldsJSON, &log.CreatedBy, &log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan ai_feedback_log: %w", err)
		}
		if modifiedFieldsJSON != nil {
			if err := json.Unmarshal(modifiedFieldsJSON, &log.HumanModifiedFields); err != nil {
				return nil, fmt.Errorf("unmarshal human_modified_fields: %w", err)
			}
		}
		logs = append(logs, log)
	}
	return logs, rows.Err()
}