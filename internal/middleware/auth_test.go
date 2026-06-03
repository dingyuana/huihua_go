package middleware

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"huihua/finance/internal/config"
	"huihua/finance/pkg/jwt"
)

func TestAuth_MissingHeader(t *testing.T) {
	app := fiber.New()
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret",
			Expiry: "30m",
		},
	}

	app.Use(Auth(cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestAuth_InvalidFormat(t *testing.T) {
	app := fiber.New()
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret",
			Expiry: "30m",
		},
	}

	app.Use(Auth(cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "InvalidFormat token123")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestAuth_ValidToken(t *testing.T) {
	app := fiber.New()
	secret := "test-secret-key-256-bits-long!!"
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret: secret,
			Expiry: "30m",
		},
	}

	// Create a valid token
	userID := uuid.New()
	tenantID := uuid.New()
	token, err := jwt.GenerateToken(secret, userID, tenantID, "admin", 30*time.Minute)
	if err != nil {
		t.Fatalf("generate token failed: %v", err)
	}

	var capturedUserID, capturedTenantID interface{}
	app.Use(Auth(cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		capturedUserID = c.Locals("user_id")
		capturedTenantID = c.Locals("tenant_id")
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	if capturedUserID == nil {
		t.Error("user_id should be set in locals")
	}
	if capturedTenantID == nil {
		t.Error("tenant_id should be set in locals")
	}
}

func TestAuth_InvalidToken(t *testing.T) {
	app := fiber.New()
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret",
			Expiry: "30m",
		},
	}

	app.Use(Auth(cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}
