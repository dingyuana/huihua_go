package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"huihua/finance/pkg/database"
)

func Tenant(db *database.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tenantID, ok := c.Locals("tenant_id").(uuid.UUID)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing tenant context",
			})
		}

		if err := db.SetTenant(tenantID); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "tenant switch failed",
			})
		}

		return c.Next()
	}
}

func TenantFromContext(c *fiber.Ctx) uuid.UUID {
	return c.Locals("tenant_id").(uuid.UUID)
}
