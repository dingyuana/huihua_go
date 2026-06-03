package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"huihua/finance/internal/model"
	"huihua/finance/internal/repository"
)

// AIFeedbackService records human actions on bank transactions for AI learning.
// It implements AC10: every human modification must be logged with the AI
// suggestion that was originally made and the fields that were changed.
type AIFeedbackService struct {
	repo *repository.AIFeedbackLogRepository
}

// NewAIFeedbackService creates a new AIFeedbackService.
func NewAIFeedbackService(repo *repository.AIFeedbackLogRepository) *AIFeedbackService {
	return &AIFeedbackService{repo: repo}
}

// Log records a human action on a bank transaction. If tx is nil the service
// uses its own DB connection; otherwise it participates in the caller's txn.
// The AI fields (suggested_action, confidence, business_scene) are read from
// the bank_transactions table at the time of logging so the log captures the
// AI state that existed when the human acted.
func (s *AIFeedbackService) Log(ctx context.Context, tx pgx.Tx, bankTxnID uuid.UUID, humanAction string) error {
	return s.LogWithModifiedFields(ctx, tx, bankTxnID, humanAction, nil)
}

// LogWithModifiedFields records a human action with the set of fields that were modified.
// The humanModifiedFields map may be nil if no fields were changed.
func (s *AIFeedbackService) LogWithModifiedFields(ctx context.Context, tx pgx.Tx, bankTxnID uuid.UUID, humanAction string, humanModifiedFields map[string]any) error {
	// Fetch AI fields and tenant_id from bank_transactions using pgtype
	// to handle NULL values gracefully.
	var tenantID uuid.UUID
	var suggestedAction, businessScene pgtype.Text
	var confidence pgtype.Int4

	var err error
	if tx != nil {
		err = tx.QueryRow(ctx, `
			SELECT tenant_id, ai_suggested_action, ai_confidence, ai_business_scene
			FROM bank_transactions WHERE id = $1`,
			bankTxnID).Scan(&tenantID, &suggestedAction, &confidence, &businessScene)
	} else {
		err = s.repo.Pool().QueryRow(ctx, `
			SELECT tenant_id, ai_suggested_action, ai_confidence, ai_business_scene
			FROM bank_transactions WHERE id = $1`,
			bankTxnID).Scan(&tenantID, &suggestedAction, &confidence, &businessScene)
	}
	if err != nil {
		return fmt.Errorf("fetch bank_txn AI fields: %w", err)
	}

	// Build the log entry, converting pgtype values to model pointer types.
	log := &model.AIFeedbackLog{
		ID:            uuid.New(),
		TenantID:      tenantID,
		BankTxnID:     bankTxnID,
		HumanAction:   humanAction,
		HumanModifiedFields: humanModifiedFields,
		CreatedAt:     time.Now(),
	}
	if suggestedAction.Valid {
		log.AISuggestedAction = &suggestedAction.String
	}
	if businessScene.Valid {
		log.AIBusinessScene = &businessScene.String
	}
	if confidence.Valid {
		c := int(confidence.Int32)
		log.AIConfidence = &c
	}

	if tx != nil {
		return s.repo.CreateTx(ctx, tx, log)
	}
	return s.repo.Create(ctx, log)
}