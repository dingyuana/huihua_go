package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/service"
)

type AdvanceAllocationHandler struct {
	svc *service.AdvanceAllocationService
}

func NewAdvanceAllocationHandler(svc *service.AdvanceAllocationService) *AdvanceAllocationHandler {
	return &AdvanceAllocationHandler{svc: svc}
}

type AllocateHTTP struct {
	AdvanceID      string  `json:"advance_id"`
	AdvanceType    string  `json:"advance_type"`
	TargetID       string  `json:"target_id"`
	TargetType     string  `json:"target_type"`
	AllocatedAmount string `json:"allocated_amount"`
	AllocationDate string  `json:"allocation_date"`
	Remark         *string `json:"remark,omitempty"`
}

func (h *AdvanceAllocationHandler) Allocate(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	userID := c.Locals("user_id").(uuid.UUID)

	var req AllocateHTTP
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body: " + err.Error()})
	}
	advanceID, err := uuid.Parse(req.AdvanceID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid advance_id"})
	}
	targetID, err := uuid.Parse(req.TargetID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid target_id"})
	}
	amount, err := decimal.NewFromString(req.AllocatedAmount)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid allocated_amount"})
	}
	allocDate, _ := time.Parse("2006-01-02", req.AllocationDate)
	if allocDate.IsZero() {
		allocDate = time.Now()
	}
	svcReq := &service.AllocateRequest{
		AdvanceID: advanceID, AdvanceType: req.AdvanceType,
		TargetID: targetID, TargetType: req.TargetType,
		AllocatedAmount: amount, AllocationDate: allocDate,
		Remark: req.Remark,
	}
	alloc, err := h.svc.Allocate(c.Context(), tenantID, userID, svcReq)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"data": alloc, "message": "allocation successful"})
}

func (h *AdvanceAllocationHandler) AutoMatch(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	userID := c.Locals("user_id").(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}
	allocs, err := h.svc.AutoMatch(c.Context(), tenantID, userID, id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": allocs, "count": len(allocs), "message": "auto-match complete"})
}

func (h *AdvanceAllocationHandler) ListByAdvance(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	advanceID, err := uuid.Parse(c.Query("advance_id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "advance_id required"})
	}
	list, err := h.svc.ListByAdvance(c.Context(), tenantID, advanceID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"list": list, "total": len(list)})
}

func (h *AdvanceAllocationHandler) ListByTarget(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	targetID, err := uuid.Parse(c.Query("target_id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "target_id required"})
	}
	list, err := h.svc.ListByTarget(c.Context(), tenantID, targetID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"list": list, "total": len(list)})
}
