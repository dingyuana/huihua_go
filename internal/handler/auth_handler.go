package handler

import (
	"huihua/finance/internal/service"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authSvc *service.AuthService
}

func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	result, err := h.authSvc.Login(c.Context(), req.Username, req.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"token":      result.Token,
		"user_id":    result.UserID,
		"tenant_id":  result.TenantID,
		"role":       result.Role,
		"expires_at": result.ExpiresAt,
	})
}

// Logout invalidates the current session.
// As of the current JWT-only implementation, logout is a client-side concern
// (the client discards the token). The server returns 200 to acknowledge.
// A future Redis-backed token blacklist would replace this stub.
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "logged out (client should discard token)",
	})
}
