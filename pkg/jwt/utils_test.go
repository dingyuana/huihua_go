package jwt

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestGenerateAndParseToken(t *testing.T) {
	secret := "test-secret-key-256-bits-long!!"
	userID := uuid.New()
	tenantID := uuid.New()
	role := "admin"

	token, err := GenerateToken(secret, userID, tenantID, role, 30*time.Minute)
	if err != nil {
		t.Fatalf("generate token failed: %v", err)
	}

	if token == "" {
		t.Fatal("token should not be empty")
	}

	// Parse
	claims, err := ParseToken(secret, token)
	if err != nil {
		t.Fatalf("parse token failed: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("user_id mismatch: got %v, want %v", claims.UserID, userID)
	}
	if claims.TenantID != tenantID {
		t.Errorf("tenant_id mismatch: got %v, want %v", claims.TenantID, tenantID)
	}
	if claims.Role != role {
		t.Errorf("role mismatch: got %v, want %v", claims.Role, role)
	}
}

func TestParseToken_WrongSecret(t *testing.T) {
	secret := "test-secret"
	token, _ := GenerateToken(secret, uuid.New(), uuid.New(), "admin", 30*time.Minute)

	_, err := ParseToken("wrong-secret", token)
	if err == nil {
		t.Fatal("expected error for wrong secret")
	}
}

func TestParseToken_ExpiredToken(t *testing.T) {
	secret := "test-secret"
	userID := uuid.New()
	tenantID := uuid.New()

	// Generate token that expires immediately (negative expiry)
	token, err := GenerateToken(secret, userID, tenantID, "admin", -1*time.Hour)
	if err != nil {
		t.Fatalf("generate token failed: %v", err)
	}

	_, err = ParseToken(secret, token)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}