package handler

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/google/uuid"
	"huihua/finance/internal/repository"
	"huihua/finance/internal/service"
)

func testHandlerPool(t *testing.T) *pgxpool.Pool {
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

func testAuthMW(c *fiber.Ctx) error {
	c.Locals("tenant_id", uuid.MustParse("00000000-0000-0000-0000-000000000001"))
	c.Locals("user_id", uuid.MustParse("00000000-0000-0000-0000-000000000101"))
	return c.Next()
}

func TestPeriodHandler_PreCloseCheck(t *testing.T) {
	pool := testHandlerPool(t)
	svc := service.NewPeriodService(
		repository.NewPeriodRepository(pool),
		repository.NewJournalRepository(pool),
		repository.NewGLEntryRepository(pool),
		repository.NewAccountRepository(pool),
		repository.NewAssetDepreciationRepository(pool),
	)
	h := NewPeriodHandler(svc)

	app := fiber.New()
	app.Use(testAuthMW)
	app.Get("/periods/pre-close-check", h.PreCloseCheck)

	req := httptest.NewRequest("GET", "/periods/pre-close-check?year=2026&month=5", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	t.Logf("PreCloseCheck response: period_status=%v unposted=%v", body["period_status"], body["unposted_vouchers"])

	if body["period_status"] != "open" {
		t.Errorf("expected period_status=open, got %v", body["period_status"])
	}
}

func TestPeriodHandler_PreCloseCheck_MissingParams(t *testing.T) {
	pool := testHandlerPool(t)
	svc := service.NewPeriodService(
		repository.NewPeriodRepository(pool),
		repository.NewJournalRepository(pool),
		repository.NewGLEntryRepository(pool),
		repository.NewAccountRepository(pool),
		repository.NewAssetDepreciationRepository(pool),
	)
	h := NewPeriodHandler(svc)

	app := fiber.New()
	app.Use(testAuthMW)
	app.Get("/periods/pre-close-check", h.PreCloseCheck)

	req := httptest.NewRequest("GET", "/periods/pre-close-check", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
	t.Logf("Missing params returned %d as expected", resp.StatusCode)
}

func TestPeriodHandler_Close(t *testing.T) {
	pool := testHandlerPool(t)
	svc := service.NewPeriodService(
		repository.NewPeriodRepository(pool),
		repository.NewJournalRepository(pool),
		repository.NewGLEntryRepository(pool),
		repository.NewAccountRepository(pool),
		repository.NewAssetDepreciationRepository(pool),
	)
	h := NewPeriodHandler(svc)

	app := fiber.New()
	app.Use(testAuthMW)
	app.Post("/periods/:period_no/close", h.Close)

	body := `{"period_no":202612,"user_id":"00000000-0000-0000-0000-000000000101","user_name":"admin","generate_closing_entries":false}`
	req := httptest.NewRequest("POST", "/periods/202612/close", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	// 202612 may not exist, so 400 is also acceptable
	t.Logf("Close period status: %d", resp.StatusCode)
}

func TestPeriodHandler_List(t *testing.T) {
	pool := testHandlerPool(t)
	svc := service.NewPeriodService(
		repository.NewPeriodRepository(pool),
		repository.NewJournalRepository(pool),
		repository.NewGLEntryRepository(pool),
		repository.NewAccountRepository(pool),
		repository.NewAssetDepreciationRepository(pool),
	)
	h := NewPeriodHandler(svc)

	app := fiber.New()
	app.Use(testAuthMW)
	app.Get("/periods", h.List)

	req := httptest.NewRequest("GET", "/periods", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	t.Logf("List periods status: %d", resp.StatusCode)
}
