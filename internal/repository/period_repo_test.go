package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"huihua/finance/internal/model"
)

func TestPeriodRepo_FindByPeriodNo(t *testing.T) {
	pool := testRepoPool(t)
	repo := NewPeriodRepository(pool)
	ctx := context.Background()

	periods, err := repo.ListByTenant(ctx, testTenantID)
	if err != nil {
		t.Fatalf("ListByTenant failed: %v", err)
	}
	var target *model.AccountingPeriod
	for _, p := range periods {
		if p.PeriodNo == 202605 {
			target = &p
			break
		}
	}
	if target == nil {
		t.Fatal("period 202605 not found")
	}
	t.Logf("Period: no=%d status=%s start=%s end=%s", target.PeriodNo, target.Status, target.StartDate, target.EndDate)
}

func TestPeriodRepo_ListByTenant(t *testing.T) {
	pool := testRepoPool(t)
	repo := NewPeriodRepository(pool)

	periods, err := repo.ListByTenant(context.Background(), testTenantID)
	if err != nil {
		t.Fatalf("ListByTenant failed: %v", err)
	}
	if len(periods) == 0 {
		t.Error("expected at least one period")
	}
	t.Logf("Found %d periods", len(periods))
	for _, p := range periods[:3] {
		t.Logf("  %d status=%s", p.PeriodNo, p.Status)
	}
}

func TestPeriodRepo_UpdateStatus(t *testing.T) {
	pool := testRepoPool(t)
	repo := NewPeriodRepository(pool)
	ctx := context.Background()

	periodNo := 202605
	adminID := uuid.MustParse("00000000-0000-0000-0000-000000000101")

	periods, err := repo.ListByTenant(ctx, testTenantID)
	if err != nil {
		t.Fatalf("ListByTenant failed: %v", err)
	}
	var period *model.AccountingPeriod
	for i, p := range periods {
		if p.PeriodNo == periodNo {
			period = &periods[i]
			break
		}
	}
	if period == nil {
		t.Fatalf("period %d not found", periodNo)
	}
	origStatus := period.Status
	t.Logf("Original status: %s", origStatus)

	// Toggle: open -> closing -> back to open
	if origStatus == "open" {
		err = repo.UpdateStatus(ctx, testTenantID, periodNo, "closing", adminID)
		if err != nil {
			t.Fatalf("UpdateStatus to closing failed: %v", err)
		}
		err = repo.UpdateStatus(ctx, testTenantID, periodNo, "open", adminID)
		if err != nil {
			t.Fatalf("UpdateStatus back to open failed: %v", err)
		}
		t.Log("Status update cycle succeeded")
	} else {
		t.Logf("Skipping status update test for non-open period (status=%s)", origStatus)
	}
}
