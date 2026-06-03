package repository

import (
	"context"
	"testing"
)

func TestJournalRepo_GetByPeriod(t *testing.T) {
	pool := testRepoPool(t)
	repo := NewJournalRepository(pool)
	ctx := context.Background()

	entries, err := repo.GetByPeriod(ctx, testTenantID, "202605")
	if err != nil {
		t.Fatalf("GetByPeriod failed: %v", err)
	}
	t.Logf("Found %d journal entries for 202605", len(entries))
	if len(entries) > 0 {
		e := entries[0]
		t.Logf("  Example: id=%s docstatus=%d voucher_no=%s",
			e.ID.String(), e.DocStatus, e.VoucherNo)
	}
}

func TestJournalRepo_CountUnpostedByPeriod(t *testing.T) {
	pool := testRepoPool(t)
	repo := NewJournalRepository(pool)
	ctx := context.Background()

	count, err := repo.CountUnpostedByPeriod(ctx, testTenantID, "202605")
	if err != nil {
		t.Fatalf("CountUnpostedByPeriod failed: %v", err)
	}
	if count < 0 {
		t.Errorf("expected count >= 0, got %d", count)
	}
	t.Logf("Unposted vouchers in 202605: %d", count)
}

func TestJournalRepo_CountClosingByPeriod(t *testing.T) {
	pool := testRepoPool(t)
	repo := NewJournalRepository(pool)
	ctx := context.Background()

	count, err := repo.CountClosingByPeriod(ctx, testTenantID, "202605")
	if err != nil {
		t.Fatalf("CountClosingByPeriod failed: %v", err)
	}
	if count < 0 {
		t.Errorf("expected count >= 0, got %d", count)
	}
	t.Logf("Closing vouchers in 202605: %d", count)
}
