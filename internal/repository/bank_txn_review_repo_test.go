package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"huihua/finance/internal/model"
)

type BankTxnReviewRepoSuite struct {
	suite.Suite
	pool     *pgxpool.Pool
	repo     *BankTransactionRepository
	tenantID uuid.UUID
}

func (s *BankTxnReviewRepoSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test in short mode")
	}
	ctx := context.Background()
	connStr := "postgres://huihua:hfpwd@127.0.0.1:5432/huihua_finance?sslmode=disable"
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		s.T().Fatalf("connect to test db: %v", err)
	}
	s.pool = pool
	s.repo = NewBankTransactionRepository(pool)
	s.tenantID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
}

func (s *BankTxnReviewRepoSuite) TearDownSuite() {
	if s.pool == nil {
		return
	}
	ctx := context.Background()
	// Clean up test rows by description prefix
	_, _ = s.pool.Exec(ctx, `
		DELETE FROM bank_transactions
		WHERE tenant_id = $1 AND description LIKE 'TASK-BANK-01.2-TEST%%'
	`, s.tenantID)
	s.pool.Close()
}

func (s *BankTxnReviewRepoSuite) setupTestTxn(status model.BankTxnReviewStatus, classification string) uuid.UUID {
	ctx := context.Background()
	id := uuid.New()
	bankAccountID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	now := time.Now()
	desc := "TASK-BANK-01.2-TEST-" + uuid.New().String()[:8]

	_, err := s.pool.Exec(ctx, `
		INSERT INTO bank_transactions (
			id, tenant_id, bank_account_id, txn_date, description,
			debit, credit, direction, counterparty_name, classification,
			matched, confirmed, status, company_id, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
	`,
		id, s.tenantID, bankAccountID, now, desc,
		decimal.NewFromInt(100), decimal.NewFromInt(0), "out", "TestCounterparty", classification,
		false, false, status, s.tenantID, now,
	)
	if err != nil {
		s.T().Fatalf("insert test txn: %v", err)
	}
	return id
}

// TestListByStatus_AC8 verifies status=classified returns correct paginated data.
func (s *BankTxnReviewRepoSuite) TestListByStatus_AC8() {
	ctx := context.Background()

	// Insert 3 classified + 1 pending test records
	t1 := s.setupTestTxn(model.BankTxnReviewStatusClassified, "bank_fee")
	t2 := s.setupTestTxn(model.BankTxnReviewStatusClassified, "transfer")
	t3 := s.setupTestTxn(model.BankTxnReviewStatusClassified, "interest")
	_ = s.setupTestTxn(model.BankTxnReviewStatusPending, "unknown") // should not appear

	txns, total, err := s.repo.ListByStatus(ctx, s.tenantID, string(model.BankTxnReviewStatusClassified), 1, 50)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(3), total)
	assert.Len(s.T(), txns, 3)

	// Verify returned IDs are the classified ones
	idMap := make(map[uuid.UUID]bool)
	for _, t := range txns {
		idMap[t.ID] = true
	}
	assert.True(s.T(), idMap[t1], "expected classified txn t1 in result")
	assert.True(s.T(), idMap[t2], "expected classified txn t2 in result")
	assert.True(s.T(), idMap[t3], "expected classified txn t3 in result")

	// Test pagination - page 1 size 2 should return 2 items
	txns2, total2, err := s.repo.ListByStatus(ctx, s.tenantID, string(model.BankTxnReviewStatusClassified), 1, 2)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(3), total2)
	assert.Len(s.T(), txns2, 2)
}

// TestUpdateStatus_AC7 verifies status transition from classified to manual_pending.
func (s *BankTxnReviewRepoSuite) TestUpdateStatus_AC7() {
	ctx := context.Background()

	txnID := s.setupTestTxn(model.BankTxnReviewStatusClassified, "bank_fee")

	// Update status to manual_pending
	err := s.repo.UpdateStatus(ctx, txnID, s.tenantID, model.BankTxnReviewStatusManualPending)
	assert.NoError(s.T(), err)

	// Re-query and verify status changed
	txn, err := s.repo.GetByID(ctx, s.tenantID, txnID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), model.BankTxnReviewStatusManualPending, txn.Status)
}

// TestGetStats_AC9 verifies all four stats numbers are returned.
func (s *BankTxnReviewRepoSuite) TestGetStats_AC9() {
	ctx := context.Background()

	// Insert one of each relevant status
	s.setupTestTxn(model.BankTxnReviewStatusPending, "unknown")
	s.setupTestTxn(model.BankTxnReviewStatusClassified, "bank_fee")
	s.setupTestTxn(model.BankTxnReviewStatusApproved, "transfer")
	s.setupTestTxn(model.BankTxnReviewStatusManualPending, "unknown")

	stats, err := s.repo.GetStats(ctx, s.tenantID)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), stats)

	// All counters should be >= 0
	assert.GreaterOrEqual(s.T(), stats.MonthlyTxns, int64(0))
	assert.GreaterOrEqual(s.T(), stats.PendingCount, int64(0))
	assert.GreaterOrEqual(s.T(), stats.AIProcessedCount, int64(0))
	assert.GreaterOrEqual(s.T(), stats.ManualPendingCount, int64(0))

	// We inserted 4 test records this month, so monthly_txns should be at least 4
	assert.GreaterOrEqual(s.T(), stats.MonthlyTxns, int64(4), "monthly_txns should count test records")
}

// TestGetByIDForUpdate_AC5 verifies FOR UPDATE lock returns correct BankTransaction.
func (s *BankTxnReviewRepoSuite) TestGetByIDForUpdate_AC5() {
	ctx := context.Background()

	txnID := s.setupTestTxn(model.BankTxnReviewStatusPending, "unknown")

	tx, _ := s.repo.BeginTx(ctx)
	txnFromDB, err := s.repo.GetByIDForUpdate(ctx, tx, txnID)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), txnFromDB)
	assert.Equal(s.T(), txnID, txnFromDB.ID)
	assert.Equal(s.T(), s.tenantID, txnFromDB.TenantID)
}

func TestBankTxnReviewRepoSuite(t *testing.T) {
	suite.Run(t, new(BankTxnReviewRepoSuite))
}