package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"huihua/finance/internal/config"
	"huihua/finance/pkg/jwt"
)

func Auth(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing authorization header",
			})
		}

		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid authorization format",
			})
		}
		tokenStr := authHeader[7:]

		claims, err := jwt.ParseToken(cfg.JWT.Secret, tokenStr)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid or expired token",
			})
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("tenant_id", claims.TenantID)
		c.Locals("role", claims.Role)

		return c.Next()
	}
}

func GetUserID(c *fiber.Ctx) (uuid.UUID, bool) {
	id, ok := c.Locals("user_id").(uuid.UUID)
	return id, ok
}

func GetTenantID(c *fiber.Ctx) (uuid.UUID, bool) {
	id, ok := c.Locals("tenant_id").(uuid.UUID)
	return id, ok
}

// MustGetTenantID returns tenantID or panics if not set (use only in protected routes).
func MustGetTenantID(c *fiber.Ctx) uuid.UUID {
	id, ok := c.Locals("tenant_id").(uuid.UUID)
	if !ok {
		panic("tenant_id not set in context")
	}
	return id
}

// MustGetUserID returns userID or panics if not set (use only in protected routes).
func MustGetUserID(c *fiber.Ctx) uuid.UUID {
	id, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		panic("user_id not set in context")
	}
	return id
}
