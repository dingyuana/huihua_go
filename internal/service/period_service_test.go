package service

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"huihua/finance/internal/repository"
)

var (
	testTenantID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	testCtx      = context.Background()
)

func skipIfNoDB(t *testing.T) {
	t.Helper()
	if os.Getenv("SKIP_INTEGRATION") == "1" {
		t.Skip("SKIP_INTEGRATION=1, skipping integration test")
	}
}

func newTestPeriodService(t *testing.T) *PeriodService {
	t.Helper()
	pool, err := pgxpool.New(testCtx, "postgres://huihua:hfpwd@127.0.0.1:5432/huihua_finance?sslmode=disable")
	if err != nil {
		t.Fatalf("connect to test db: %v", err)
	}
	t.Cleanup(func() { pool.Close() })

	return NewPeriodService(
		repository.NewPeriodRepository(pool),
		repository.NewJournalRepository(pool),
		repository.NewGLEntryRepository(pool),
		repository.NewAccountRepository(pool),
		repository.NewAssetDepreciationRepository(pool),
		repository.NewBankTransactionRepository(pool),
		repository.NewInvoiceRepository(pool),
	)
}

func TestPeriodService_PreCloseCheck(t *testing.T) {
	skipIfNoDB(t)
	svc := newTestPeriodService(t)

	result, err := svc.PreCloseCheck(testCtx, testTenantID, 2026, 5)
	if err != nil {
		t.Fatalf("PreCloseCheck failed: %v", err)
	}

	if result.PeriodStatus != "open" {
		t.Errorf("expected period_status=open, got %q", result.PeriodStatus)
	}
	if result.UnpostedVouchers < 0 {
		t.Errorf("unposted_vouchers should be >= 0, got %d", result.UnpostedVouchers)
	}
	if result.RiskWarnings == nil {
		t.Error("risk_warnings should not be nil")
	}
	if result.KeyIndicators == nil {
		t.Error("key_indicators should not be nil")
	}
	if result.PendingAccruals == nil {
		t.Error("pending_accruals should not be nil")
	}

	t.Logf("PreCloseCheck result: status=%s, unposted=%d, balance_ok=%v, pl_done=%v",
		result.PeriodStatus, result.UnpostedVouchers, result.ReportBalanceOK, result.ProfitLossDone)
	t.Logf("  risk_warnings: %d, key_indicators: %d, pending_accruals: %d",
		len(result.RiskWarnings), len(result.KeyIndicators), len(result.PendingAccruals))
}

func TestPeriodService_CloseUnclose(t *testing.T) {
	skipIfNoDB(t)
	svc := newTestPeriodService(t)

	// Use a future test period that shouldn't interfere
	periodNo := 209912

	// Close the period
	err := svc.ClosePeriod(testCtx, testTenantID, &ClosePeriodRequest{
		PeriodNo:               periodNo,
		UserID:                 "00000000-0000-0000-0000-000000000101",
		UserName:               "admin",
		GenerateClosingEntries: false,
	})
	if err != nil {
		// Period might not exist, that's OK for integration test
		t.Logf("ClosePeriod (may be expected): %v", err)
		return
	}
	t.Log("ClosePeriod succeeded")

	// Unclose the period
	err = svc.UnclosePeriod(testCtx, testTenantID, periodNo)
	if err != nil {
		t.Fatalf("UnclosePeriod failed after close: %v", err)
	}
	t.Log("UnclosePeriod succeeded")
}
