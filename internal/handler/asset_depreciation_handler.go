package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"huihua/finance/internal/model"
	"huihua/finance/internal/service"
)

// AssetDepreciationHandler handles asset depreciation HTTP requests.
type AssetDepreciationHandler struct {
	svc *service.AssetDepreciationService
}

// NewAssetDepreciationHandler creates a new AssetDepreciationHandler.
func NewAssetDepreciationHandler(svc *service.AssetDepreciationService) *AssetDepreciationHandler {
	return &AssetDepreciationHandler{svc: svc}
}

// CreateSchedule handles POST /api/v1/assets/:id/depreciation/schedule
// It generates a depreciation schedule for the specified asset.
func (h *AssetDepreciationHandler) CreateSchedule(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	assetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid asset_id",
		})
	}

	var req model.CreateScheduleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body: " + err.Error(),
		})
	}

	// Validate method
	switch req.Method {
	case model.DepreciationMethodStraightLine,
		model.DepreciationMethodDoubleDeclining,
		model.DepreciationMethodUnitsOfProduction:
		// valid
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid depreciation_method: must be straight_line, double_declining, or units_of_production",
		})
	}

	// Validate useful_life
	if req.UsefulLife <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "useful_life must be positive",
		})
	}

	schedules, err := h.svc.CreateSchedule(c.Context(), tenantID, assetID, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":   "depreciation schedule created",
		"schedules": schedules,
	})
}

// RunDepreciation handles POST /api/v1/depreciation/run
// It executes monthly depreciation for a specified period.
func (h *AssetDepreciationHandler) RunDepreciation(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	var req model.RunDepreciationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body: " + err.Error(),
		})
	}

	if req.PeriodNo <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "period_no is required and must be positive",
		})
	}

	run, err := h.svc.GenerateMonthlyDepreciation(c.Context(), tenantID, req.PeriodNo)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "depreciation run completed",
		"run":     run,
	})
}

// GenerateDepreciation POST /api/v1/depreciation/generate?period_no=202506
// 生成指定期间的折旧凭证（草稿状态）。
// 人审核后在凭证列表点击"核准"过账。
func (h *AssetDepreciationHandler) GenerateDepreciation(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	periodNo := c.QueryInt("period_no", 0)
	if periodNo == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "period_no is required"})
	}
	run, err := h.svc.GenerateMonthlyDepreciation(c.Context(), tenantID, periodNo)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": run})
}

// GenerateAmortization POST /api/v1/depreciation/generate-amortization?period_no=202506
// 生成指定期间的无形资产摊销凭证（草稿状态）。
func (h *AssetDepreciationHandler) GenerateAmortization(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	periodNo := c.QueryInt("period_no", 0)
	if periodNo == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "period_no is required"})
	}
	run, err := h.svc.GenerateMonthlyAmortization(c.Context(), tenantID, periodNo)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": run})
}

// ListDepreciationRuns handles GET /api/v1/depreciation/run
// It retrieves depreciation run history, optionally filtered by period.
func (h *AssetDepreciationHandler) ListDepreciationRuns(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	periodNo := c.QueryInt("period_no", 0)

	runs, err := h.svc.GetDepreciationRuns(c.Context(), tenantID, periodNo)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"runs": runs,
	})
}

// GetSchedule handles GET /api/v1/assets/:id/depreciation/schedule
// It retrieves the depreciation schedule for a specific asset.
func (h *AssetDepreciationHandler) GetSchedule(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	assetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid asset_id",
		})
	}

	schedules, err := h.svc.GetAssetDepreciationSchedule(c.Context(), tenantID, assetID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"schedules": schedules,
	})
}
