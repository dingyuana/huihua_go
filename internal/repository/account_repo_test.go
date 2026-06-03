package repository

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"huihua/finance/internal/model"
)

var testTenantID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

func testRepoPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	if os.Getenv("SKIP_INTEGRATION") == "1" {
		t.Skip("SKIP_INTEGRATION=1")
	}
	pool, err := pgxpool.New(context.Background(), "postgres://huihua:hfpwd@127.0.0.1:5432/huihua_finance?sslmode=disable")
	if err != nil {
		t.Fatalf("connect to test db: %v", err)
	}
	t.Cleanup(func() { pool.Close() })
	return pool
}

func createTestAccount(t *testing.T, repo *AccountRepository, code, name string, lft, rgt int) *model.Account {
	t.Helper()
	acct := &model.Account{
		Code:      code + "-" + uuid.New().String()[:8],
		Name:      name,
		CompanyID: testTenantID,
		TenantID:  testTenantID,
		Currency:  "CNY",
		IsActive:  true,
		Lft:       lft,
		Rgt:       rgt,
		IsGroup:   false,
	}
	created, err := repo.Create(context.Background(), testTenantID, acct)
	if err != nil {
		t.Fatalf("Create(%s) failed: %v", code, err)
	}
	return created
}

func TestAccountRepo_CreateAndGetByCode(t *testing.T) {
	pool := testRepoPool(t)
	repo := NewAccountRepository(pool)

	created := createTestAccount(t, repo, "TST001", "Test Account 001", 1, 2)
	t.Logf("Created account id=%s code=%s", created.ID.String(), created.Code)
	t.Logf("Name=%s", created.Name)
}

func TestAccountRepo_ListByTenant(t *testing.T) {
	pool := testRepoPool(t)
	repo := NewAccountRepository(pool)

	createTestAccount(t, repo, "TSTLST", "Test List Acct", 10, 11)

	accounts, err := repo.ListByTenant(context.Background(), testTenantID)
	if err != nil {
		t.Fatalf("ListByTenant failed: %v", err)
	}
	if len(accounts) == 0 {
		t.Fatal("expected at least one account after create")
	}
	t.Logf("Found %d accounts", len(accounts))
	for _, a := range accounts {
		t.Logf("  %s %s", a.Code, a.Name)
	}
}

func TestAccountRepo_GetTree(t *testing.T) {
	pool := testRepoPool(t)
	repo := NewAccountRepository(pool)

	createTestAccount(t, repo, "TSTTRE", "Test Tree Root", 20, 21)

	tree, err := repo.GetTree(context.Background(), testTenantID)
	if err != nil {
		t.Fatalf("GetTree failed: %v", err)
	}
	if len(tree) == 0 {
		t.Error("expected non-empty tree after create")
	}
	t.Logf("Tree has %d nodes", len(tree))
}
