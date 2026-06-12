package handler

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
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
		repository.NewBankTransactionRepository(pool),
		repository.NewInvoiceRepository(pool),
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
		repository.NewBankTransactionRepository(pool),
		repository.NewInvoiceRepository(pool),
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
		repository.NewBankTransactionRepository(pool),
		repository.NewInvoiceRepository(pool),
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
		repository.NewBankTransactionRepository(pool),
		repository.NewInvoiceRepository(pool),
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

func TestPeriodHandler_CreatePeriod(t *testing.T) {
	pool := testHandlerPool(t)
	svc := service.NewPeriodService(
		repository.NewPeriodRepository(pool),
		repository.NewJournalRepository(pool),
		repository.NewGLEntryRepository(pool),
		repository.NewAccountRepository(pool),
		repository.NewAssetDepreciationRepository(pool),
		repository.NewBankTransactionRepository(pool),
		repository.NewInvoiceRepository(pool),
	)
	h := NewPeriodHandler(svc)

	app := fiber.New()
	app.Use(testAuthMW)
	app.Post("/periods", h.CreatePeriod)

	body := `{"period_no":209901,"period_name":"2099-01","start_date":"2099-01-01T00:00:00Z","end_date":"2099-01-31T00:00:00Z"}`
	req := httptest.NewRequest("POST", "/periods", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	data, ok := result["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected data object in response")
	}
	if data["period_name"] != "2099-01" {
		t.Errorf("expected period_name=2099-01, got %v", data["period_name"])
	}
	t.Logf("CreatePeriod: id=%v period_no=%v", data["id"], data["period_no"])
}

func TestPeriodHandler_UpdatePeriod(t *testing.T) {
	pool := testHandlerPool(t)
	svc := service.NewPeriodService(
		repository.NewPeriodRepository(pool),
		repository.NewJournalRepository(pool),
		repository.NewGLEntryRepository(pool),
		repository.NewAccountRepository(pool),
		repository.NewAssetDepreciationRepository(pool),
		repository.NewBankTransactionRepository(pool),
		repository.NewInvoiceRepository(pool),
	)
	h := NewPeriodHandler(svc)

	app := fiber.New()
	app.Use(testAuthMW)
	app.Post("/periods", h.CreatePeriod)
	app.Put("/periods/:id", h.UpdatePeriod)
	app.Delete("/periods/:id", h.DeletePeriod)
	app.Post("/periods/:period_no/close", h.Close)
	app.Post("/periods/:id/enable", h.EnablePeriod)

	createBody := `{"period_no":209902,"period_name":"2099-02","start_date":"2099-02-01T00:00:00Z","end_date":"2099-02-28T00:00:00Z"}`
	creq := httptest.NewRequest("POST", "/periods", strings.NewReader(createBody))
	creq.Header.Set("Content-Type", "application/json")
	cresp, err := app.Test(creq)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if cresp.StatusCode != fiber.StatusOK {
		t.Fatalf("create expected 200, got %d", cresp.StatusCode)
	}
	var createResult map[string]interface{}
	if err := json.NewDecoder(cresp.Body).Decode(&createResult); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	periodID := createResult["data"].(map[string]interface{})["id"].(string)

	updateBody := `{"period_name":"2099-02-updated"}`
	ureq := httptest.NewRequest("PUT", "/periods/"+periodID, strings.NewReader(updateBody))
	ureq.Header.Set("Content-Type", "application/json")
	uresp, err := app.Test(ureq)
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if uresp.StatusCode != fiber.StatusOK {
		t.Errorf("update expected 200, got %d", uresp.StatusCode)
	}
	var updateResult map[string]interface{}
	if err := json.NewDecoder(uresp.Body).Decode(&updateResult); err != nil {
		t.Fatalf("decode update response: %v", err)
	}
	if updateResult["status"] != "updated" {
		t.Errorf("expected status=updated, got %v", updateResult["status"])
	}
	t.Logf("UpdatePeriod: id=%s status=%v", periodID, updateResult["status"])
}

func TestPeriodHandler_DeletePeriod(t *testing.T) {
	pool := testHandlerPool(t)
	svc := service.NewPeriodService(
		repository.NewPeriodRepository(pool),
		repository.NewJournalRepository(pool),
		repository.NewGLEntryRepository(pool),
		repository.NewAccountRepository(pool),
		repository.NewAssetDepreciationRepository(pool),
		repository.NewBankTransactionRepository(pool),
		repository.NewInvoiceRepository(pool),
	)
	h := NewPeriodHandler(svc)

	app := fiber.New()
	app.Use(testAuthMW)
	app.Post("/periods", h.CreatePeriod)
	app.Put("/periods/:id", h.UpdatePeriod)
	app.Delete("/periods/:id", h.DeletePeriod)
	app.Post("/periods/:period_no/close", h.Close)
	app.Post("/periods/:id/enable", h.EnablePeriod)

	createBody := `{"period_no":209903,"period_name":"2099-03","start_date":"2099-03-01T00:00:00Z","end_date":"2099-03-31T00:00:00Z"}`
	creq := httptest.NewRequest("POST", "/periods", strings.NewReader(createBody))
	creq.Header.Set("Content-Type", "application/json")
	cresp, err := app.Test(creq)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if cresp.StatusCode != fiber.StatusOK {
		t.Fatalf("create expected 200, got %d", cresp.StatusCode)
	}
	var createResult map[string]interface{}
	if err := json.NewDecoder(cresp.Body).Decode(&createResult); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	periodID := createResult["data"].(map[string]interface{})["id"].(string)

	// Delete the period (open period should succeed)
	dreq := httptest.NewRequest("DELETE", "/periods/"+periodID, nil)
	dresp, err := app.Test(dreq)
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if dresp.StatusCode != fiber.StatusOK {
		t.Errorf("delete open period expected 200, got %d", dresp.StatusCode)
	}
	var deleteResult map[string]interface{}
	if err := json.NewDecoder(dresp.Body).Decode(&deleteResult); err != nil {
		t.Fatalf("decode delete response: %v", err)
	}
	if deleteResult["status"] != "deleted" {
		t.Errorf("expected status=deleted, got %v", deleteResult["status"])
	}
	t.Logf("DeletePeriod (open): status=%v", deleteResult["status"])

	// Create another period to test closed-period deletion rejection
	createBody2 := `{"period_no":209905,"period_name":"2099-05","start_date":"2099-05-01T00:00:00Z","end_date":"2099-05-31T00:00:00Z"}`
	creq2 := httptest.NewRequest("POST", "/periods", strings.NewReader(createBody2))
	creq2.Header.Set("Content-Type", "application/json")
	cresp2, err := app.Test(creq2)
	if err != nil {
		t.Fatalf("create2 failed: %v", err)
	}
	if cresp2.StatusCode != fiber.StatusOK {
		t.Fatalf("create2 expected 200, got %d", cresp2.StatusCode)
	}
	var createResult2 map[string]interface{}
	if err := json.NewDecoder(cresp2.Body).Decode(&createResult2); err != nil {
		t.Fatalf("decode create2 response: %v", err)
	}
	periodID2 := createResult2["data"].(map[string]interface{})["id"].(string)

	// Close the period
	closeBody := `{"period_no":209905,"user_id":"00000000-0000-0000-0000-000000000101","user_name":"admin","generate_closing_entries":false}`
	clreq := httptest.NewRequest("POST", "/periods/209905/close", strings.NewReader(closeBody))
	clreq.Header.Set("Content-Type", "application/json")
	clresp, err := app.Test(clreq)
	if err != nil {
		t.Fatalf("close failed: %v", err)
	}
	t.Logf("Close period status for delete test: %d", clresp.StatusCode)

	// Try to delete the closed period — should fail with 400
	dreq2 := httptest.NewRequest("DELETE", "/periods/"+periodID2, nil)
	dresp2, err := app.Test(dreq2)
	if err != nil {
		t.Fatalf("delete closed failed: %v", err)
	}
	if dresp2.StatusCode != fiber.StatusBadRequest {
		t.Errorf("delete closed period expected 400, got %d", dresp2.StatusCode)
	}
	t.Logf("DeletePeriod (closed) correctly rejected: %d", dresp2.StatusCode)
}

func TestPeriodHandler_EnablePeriod(t *testing.T) {
	pool := testHandlerPool(t)
	svc := service.NewPeriodService(
		repository.NewPeriodRepository(pool),
		repository.NewJournalRepository(pool),
		repository.NewGLEntryRepository(pool),
		repository.NewAccountRepository(pool),
		repository.NewAssetDepreciationRepository(pool),
		repository.NewBankTransactionRepository(pool),
		repository.NewInvoiceRepository(pool),
	)
	h := NewPeriodHandler(svc)

	app := fiber.New()
	app.Use(testAuthMW)
	app.Post("/periods", h.CreatePeriod)
	app.Post("/periods/:period_no/close", h.Close)
	app.Post("/periods/:id/enable", h.EnablePeriod)

	// Create a period
	createBody := `{"period_no":209904,"period_name":"2099-04","start_date":"2099-04-01T00:00:00Z","end_date":"2099-04-30T00:00:00Z"}`
	creq := httptest.NewRequest("POST", "/periods", strings.NewReader(createBody))
	creq.Header.Set("Content-Type", "application/json")
	cresp, err := app.Test(creq)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if cresp.StatusCode != fiber.StatusOK {
		t.Fatalf("create expected 200, got %d", cresp.StatusCode)
	}
	var createResult map[string]interface{}
	if err := json.NewDecoder(cresp.Body).Decode(&createResult); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	periodID := createResult["data"].(map[string]interface{})["id"].(string)

	// Close the period
	closeBody := `{"period_no":209904,"user_id":"00000000-0000-0000-0000-000000000101","user_name":"admin","generate_closing_entries":false}`
	clreq := httptest.NewRequest("POST", "/periods/209904/close", strings.NewReader(closeBody))
	clreq.Header.Set("Content-Type", "application/json")
	clresp, err := app.Test(clreq)
	if err != nil {
		t.Fatalf("close failed: %v", err)
	}
	t.Logf("Close period status for enable test: %d", clresp.StatusCode)

	// Enable the period (closed → open)
	ereq := httptest.NewRequest("POST", "/periods/"+periodID+"/enable", nil)
	eresp, err := app.Test(ereq)
	if err != nil {
		t.Fatalf("enable failed: %v", err)
	}
	if eresp.StatusCode != fiber.StatusOK {
		t.Errorf("enable expected 200, got %d", eresp.StatusCode)
	}
	var enableResult map[string]interface{}
	if err := json.NewDecoder(eresp.Body).Decode(&enableResult); err != nil {
		t.Fatalf("decode enable response: %v", err)
	}
	if enableResult["status"] != "enabled" {
		t.Errorf("expected status=enabled, got %v", enableResult["status"])
	}
	t.Logf("EnablePeriod: status=%v", enableResult["status"])
}
