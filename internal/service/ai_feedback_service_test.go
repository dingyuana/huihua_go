package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/model"
	"huihua/finance/internal/repository"
)

// --------------------------------------------------------------------------
// TestLog_AC10: Log() successfully records an AI feedback log entry
// --------------------------------------------------------------------------

func TestLog_AC10(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	ctx := context.Background()
	pool := testPool(t)
	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	// 1. Create a bank_txn in classified status
	txnID := seedBankTxn(ctx, pool, t, &bankTxnSeed{
		tenantID:       tenantID,
		status:         string(model.BankTxnReviewStatusClassified),
		classification: ptrString("bank_fee"),
		debit:          decimal.NewFromFloat(50.00),
		credit:         decimal.Zero,
		direction:      ptrString("out"),
	})

	// 2. Create the service with a real repo backed by the pool
	feedbackRepo := repository.NewAIFeedbackLogRepository(pool)
	feedbackSvc := NewAIFeedbackService(feedbackRepo)

	// 3. Call Log() to record the human action
	humanAction := "submit_review"
	err := feedbackSvc.Log(ctx, nil, txnID, humanAction)
	if err != nil {
		t.Fatalf("Log failed: %v", err)
	}

	// 4. Query ai_feedback_logs and verify the record exists
	logs, err := feedbackRepo.ListByTxnID(ctx, txnID)
	if err != nil {
		t.Fatalf("ListByTxnID failed: %v", err)
	}
	if len(logs) == 0 {
		t.Fatal("expected at least one ai_feedback_log entry, got 0")
	}
	found := false
	for _, l := range logs {
		if l.BankTxnID == txnID && l.HumanAction == humanAction {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected log entry with human_action=%q for txn %s", humanAction, txnID)
	}
}

// --------------------------------------------------------------------------
// TestLogRecordsModifiedFields: Log() records human_modified_fields
// when a human changes the account/classification
// --------------------------------------------------------------------------

func TestLogRecordsModifiedFields(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	ctx := context.Background()
	pool := testPool(t)
	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	// 1. Create a bank_txn with AI-suggested fields set
	txnID := seedBankTxnWithAIFields(ctx, pool, t, &bankTxnWithAISeed{
		tenantID:           tenantID,
		status:             string(model.BankTxnReviewStatusClassified),
		classification:     ptrString("bank_fee"),
		aiSuggestedAction:  ptrString("auto_voucher"),
		aiConfidence:       intPtr(85),
		aiBusinessScene:    ptrString("银行手续费"),
		debit:              decimal.NewFromFloat(50.00),
		credit:             decimal.Zero,
		direction:          ptrString("out"),
	})

	feedbackRepo := repository.NewAIFeedbackLogRepository(pool)
	feedbackSvc := NewAIFeedbackService(feedbackRepo)

	// 2. Simulate human modifying the account from "财务费用" to "管理费用"
	modifiedFields := map[string]any{
		"account_from": "财务费用",
		"account_to":   "管理费用",
		"classification_from": "bank_fee",
		"classification_to": "general_expense",
	}

	// 3. Call Log with the modified fields
	humanAction := "modify_account"
	err := feedbackSvc.LogWithModifiedFields(ctx, nil, txnID, humanAction, modifiedFields)
	if err != nil {
		t.Fatalf("LogWithModifiedFields failed: %v", err)
	}

	// 4. Verify human_modified_fields is persisted
	logs, err := feedbackRepo.ListByTxnID(ctx, txnID)
	if err != nil {
		t.Fatalf("ListByTxnID failed: %v", err)
	}
	if len(logs) == 0 {
		t.Fatal("expected at least one ai_feedback_log entry, got 0")
	}

	found := false
	var lastLog *model.AIFeedbackLog
	for _, l := range logs {
		if l.BankTxnID == txnID && l.HumanAction == humanAction {
			found = true
			lastLog = &l
			break
		}
	}
	if !found {
		t.Fatalf("expected log entry with human_action=%q, not found", humanAction)
	}

	if lastLog.HumanModifiedFields == nil {
		t.Fatal("expected human_modified_fields to be set, got nil")
	}

	// Verify JSON round-trip
	data, err := json.Marshal(lastLog.HumanModifiedFields)
	if err != nil {
		t.Fatalf("json marshal human_modified_fields: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json unmarshal human_modified_fields: %v", err)
	}
	if decoded["account_from"] != "财务费用" {
		t.Errorf("expected account_from=财务费用, got %v", decoded["account_from"])
	}
	if decoded["account_to"] != "管理费用" {
		t.Errorf("expected account_to=管理费用, got %v", decoded["account_to"])
	}

	// Also verify AI fields are preserved
	if lastLog.AISuggestedAction == nil || *lastLog.AISuggestedAction != "auto_voucher" {
		t.Errorf("expected ai_suggested_action=auto_voucher, got %v", lastLog.AISuggestedAction)
	}
	if lastLog.AIConfidence == nil || *lastLog.AIConfidence != 85 {
		t.Errorf("expected ai_confidence=85, got %v", lastLog.AIConfidence)
	}
}

// --------------------------------------------------------------------------
// Test helper types
// --------------------------------------------------------------------------'

type bankTxnWithAISeed struct {
	tenantID           uuid.UUID
	status             string
	classification     *string
	aiSuggestedAction  *string
	aiConfidence       *int
	aiBusinessScene    *string
	debit              decimal.Decimal
	credit             decimal.Decimal
	direction          *string
}

func seedBankTxnWithAIFields(ctx context.Context, pool *pgxpool.Pool, t *testing.T, s *bankTxnWithAISeed) uuid.UUID {
	t.Helper()
	id := uuid.New()
	now := time.Now()
	bankAccountID := uuid.MustParse("7cd2c7fd-68c5-419f-be60-579c0f1a7b4b")
	companyID := uuid.MustParse("3586c914-5eb4-426a-9d20-f297a10b147c")
	desc := "ai test txn " + id.String()

	// Ensure bank_account row exists before referencing it
	_, err := pool.Exec(ctx, `
		INSERT INTO bank_accounts (id, bank_name, account_number, company_id, tenant_id)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO NOTHING
	`, bankAccountID, "测试银行", "TEST-0003", s.tenantID, s.tenantID)
	if err != nil {
		t.Fatalf("seed bank_account: %v", err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO bank_transactions
			(id, tenant_id, bank_account_id, company_id, txn_date, description,
			 debit, credit, direction, classification, status,
			 ai_suggested_action, ai_confidence, ai_business_scene,
			 matched, confirmed, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)`,
		id, s.tenantID, bankAccountID, companyID, now, desc,
		s.debit, s.credit, s.direction, s.classification, s.status,
		s.aiSuggestedAction, s.aiConfidence, s.aiBusinessScene,
		false, false, now, now)
	if err != nil {
		t.Fatalf("seed bank_txn with ai fields: %v", err)
	}
	return id
}

func intPtr(i int) *int { return &i }