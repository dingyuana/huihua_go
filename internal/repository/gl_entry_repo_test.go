package repository

import (
	"context"
	"testing"
	"time"
)

func TestGLEntryRepo_GetBalancesByPeriod(t *testing.T) {
	pool := testRepoPool(t)
	repo := NewGLEntryRepository(pool)
	ctx := context.Background()

	start := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 5, 31, 23, 59, 59, 0, time.UTC)

	balances, err := repo.GetBalancesByPeriod(ctx, testTenantID, start, end)
	if err != nil {
		t.Fatalf("GetBalancesByPeriod failed: %v", err)
	}
	t.Logf("Got %d account balances for 202605", len(balances))
	for _, b := range balances {
		t.Logf("  %s %s: debit=%s credit=%s", b.Code, b.Name, b.Debit.String(), b.Credit.String())
	}
}

func TestGLEntryRepo_GetIncomeExpenseSummary(t *testing.T) {
	pool := testRepoPool(t)
	repo := NewGLEntryRepository(pool)
	ctx := context.Background()

	start := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 5, 31, 23, 59, 59, 0, time.UTC)

	income, expense, err := repo.GetIncomeExpenseSummary(ctx, testTenantID, start, end)
	if err != nil {
		t.Fatalf("GetIncomeExpenseSummary failed: %v", err)
	}
	t.Logf("Income summary: income_credit=%s expense_debit=%s", income.String(), expense.String())
}

func TestGLEntryRepo_GetByAccountAndPeriod(t *testing.T) {
	pool := testRepoPool(t)
	repo := NewGLEntryRepository(pool)
	ctx := context.Background()

	start := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 5, 31, 23, 59, 59, 0, time.UTC)

	entries, err := repo.GetByTenantInRange(ctx, testTenantID, start, end)
	if err != nil {
		t.Fatalf("GetByTenantInRange failed: %v", err)
	}
	t.Logf("Found %d GL entries for 202605", len(entries))
	if len(entries) > 0 {
		e := entries[0]
		t.Logf("  Example: account_id=%s debit=%s credit=%s",
			e.AccountID.String(), e.Debit.String(), e.Credit.String())
	}
}
