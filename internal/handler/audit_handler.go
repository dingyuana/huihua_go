package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"huihua/finance/internal/middleware"
	"huihua/finance/internal/model"
	"huihua/finance/internal/repository"
	"huihua/finance/internal/service"
)

// AuditHandler handles HTTP requests for audit log queries.
type AuditHandler struct {
	svc *service.AuditService
}

// NewAuditHandler creates a new AuditHandler.
func NewAuditHandler(svc *service.AuditService) *AuditHandler {
	return &AuditHandler{svc: svc}
}

// ListAuditLogs handles GET /api/v1/audit-logs
// Supports query parameters: object_type, object_id, actor_id, start_time, end_time, limit, offset
func (h *AuditHandler) ListAuditLogs(c *fiber.Ctx) error {
	tenantID := middleware.MustGetTenantID(c)

	filter := repository.AuditFilter{}

	if v := c.Query("object_type"); v != "" {
		filter.ObjectType = v
	}

	if v := c.Query("object_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid object_id format",
			})
		}
		filter.ObjectID = id
	}

	if v := c.Query("actor_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid actor_id format",
			})
		}
		filter.ActorID = id
	}

	if v := c.Query("start_time"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid start_time format (use RFC3339)",
			})
		}
		filter.StartTime = &t
	}

	if v := c.Query("end_time"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid end_time format (use RFC3339)",
			})
		}
		filter.EndTime = &t
	}

	if v := c.Query("limit"); v != "" {
		limit, err := strconv.Atoi(v)
		if err != nil || limit < 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid limit value",
			})
		}
		filter.Limit = limit
	}

	if v := c.Query("offset"); v != "" {
		offset, err := strconv.Atoi(v)
		if err != nil || offset < 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid offset value",
			})
		}
		filter.Offset = offset
	}

	logs, err := h.svc.ListByTenant(c.Context(), tenantID, filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list audit logs",
		})
	}

	if logs == nil {
		logs = make([]model.AuditLog, 0)
	}

	return c.JSON(fiber.Map{
		"data": logs,
	})
}

// GetAuditLogsByObject handles GET /api/v1/audit-logs/:object_type/:object_id
func (h *AuditHandler) GetAuditLogsByObject(c *fiber.Ctx) error {
	tenantID := middleware.MustGetTenantID(c)

	objectType := c.Params("object_type")
	if objectType == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "object_type is required",
		})
	}

	objectIDStr := c.Params("object_id")
	objectID, err := uuid.Parse(objectIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid object_id format",
		})
	}

	logs, err := h.svc.GetByObject(c.Context(), tenantID, objectType, objectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get audit logs",
		})
	}

	if logs == nil {
		logs = make([]model.AuditLog, 0)
	}

	return c.JSON(fiber.Map{
		"data": logs,
	})
}
