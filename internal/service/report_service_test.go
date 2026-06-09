package service

import (
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"huihua/finance/internal/repository"
)

func newTestReportService(t *testing.T) *ReportService {
	t.Helper()
	if os.Getenv("SKIP_INTEGRATION") == "1" {
		t.Skip("SKIP_INTEGRATION=1")
	}
	pool, err := pgxpool.New(testCtx, "postgres://huihua:hfpwd@127.0.0.1:5432/huihua_finance?sslmode=disable")
	if err != nil {
		t.Fatalf("connect to test db: %v", err)
	}
	t.Cleanup(func() { pool.Close() })

	// Ensure period 202605 exists for report tests
	_, err = pool.Exec(testCtx, `
		INSERT INTO accounting_periods (tenant_id, period_no, period_name, start_date, end_date, status)
		VALUES ('00000000-0000-0000-0000-000000000001', 202605, '2026年5月', '2026-05-01', '2026-05-31', 'open')
		ON CONFLICT (tenant_id, period_no) DO NOTHING
	`)
	if err != nil {
		t.Fatalf("seed period 202605: %v", err)
	}

	return NewReportService(
		repository.NewGLEntryRepository(pool),
		repository.NewOpeningBalanceRepository(pool),
		repository.NewAccountRepository(pool),
		repository.NewPeriodRepository(pool),
	)
}

func TestReportService_TrialBalance(t *testing.T) {
	svc := newTestReportService(t)

	tb, err := svc.GetTrialBalance(testCtx, testTenantID, 202605)
	if err != nil {
		t.Fatalf("GetTrialBalance failed: %v", err)
	}

	if tb.PeriodNo != 202605 {
		t.Errorf("expected period_no=202605, got %d", tb.PeriodNo)
	}
	t.Logf("TrialBalance: %d entries, total_debit=%s, total_credit=%s, balanced=%v",
		len(tb.Entries), tb.TotalDebit.String(), tb.TotalCredit.String(), tb.IsBalanced)
}

func TestReportService_IncomeStatement(t *testing.T) {
	svc := newTestReportService(t)

	is, err := svc.GetIncomeStatement(testCtx, testTenantID, 202605)
	if err != nil {
		t.Fatalf("GetIncomeStatement failed: %v", err)
	}

	if is.PeriodNo != 202605 {
		t.Errorf("expected period_no=202605, got %d", is.PeriodNo)
	}
	t.Logf("IncomeStatement: income=%s, expense=%s, net=%s",
		is.TotalIncome.String(), is.TotalExpense.String(), is.NetProfit.String())
}

func TestReportService_BalanceSheet(t *testing.T) {
	svc := newTestReportService(t)

	bs, err := svc.GetBalanceSheet(testCtx, testTenantID, 202605)
	if err != nil {
		t.Fatalf("GetBalanceSheet failed: %v", err)
	}

	if bs.PeriodNo != 202605 {
		t.Errorf("expected period_no=202605, got %d", bs.PeriodNo)
	}
	t.Logf("BalanceSheet: assets=%s, liabilities=%s, equity=%s, balanced=%v",
		bs.TotalAssets.String(), bs.TotalLiabilities.String(), bs.TotalEquity.String(), bs.IsBalanced)
}

func TestReportService_CashFlow(t *testing.T) {
	svc := newTestReportService(t)

	cf, err := svc.GetCashFlowStatement(testCtx, testTenantID, 202605)
	if err != nil {
		t.Fatalf("GetCashFlowStatement failed: %v", err)
	}

	if cf.PeriodNo != 202605 {
		t.Errorf("expected period_no=202605, got %d", cf.PeriodNo)
	}
	if len(cf.Items) == 0 {
		t.Error("cash flow items should not be empty")
	}
	for _, item := range cf.Items {
		if item.Level == 0 {
			t.Logf("  section: %s = %s", item.Category, item.Current.String())
		}
	}
}
