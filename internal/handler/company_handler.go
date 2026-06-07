package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"huihua/finance/internal/middleware"
	"huihua/finance/internal/repository"
)

// CompanyHandler returns the current tenant's company settings.
type CompanyHandler struct {
	repo *repository.CompanyRepository
}

// NewCompanyHandler creates a new CompanyHandler.
func NewCompanyHandler(repo *repository.CompanyRepository) *CompanyHandler {
	return &CompanyHandler{repo: repo}
}

// GetCurrent returns the current tenant's company settings.
func (h *CompanyHandler) GetCurrent(c *fiber.Ctx) error {
	tenantID := middleware.TenantFromContext(c)
	cs, err := h.repo.GetByTenant(c.Context(), tenantID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "company settings not configured for this tenant",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(cs)
}
