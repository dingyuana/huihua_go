package handler

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"huihua/finance/internal/service"
)

// PayrollHandler handles payroll HTTP requests.
type PayrollHandler struct {
	svc *service.PayrollService
}

// NewPayrollHandler creates a new PayrollHandler.
func NewPayrollHandler(svc *service.PayrollService) *PayrollHandler {
	return &PayrollHandler{svc: svc}
}

// PayrollCreateRequest is the request body for creating a payroll record.
type PayrollCreateRequest struct {
	EmployeeName    string  `json:"employee_name"`
	DepartmentName  string  `json:"department_name"`
	PeriodNo        int     `json:"period_no"`
	GrossSalary     string  `json:"gross_salary"`
	IndividualTax   string  `json:"individual_tax"`
	SocialSecurity  string  `json:"social_security"`
	HousingFund     string  `json:"housing_fund"`
	OtherDeductions string  `json:"other_deductions"`
	NetSalary       string  `json:"net_salary"`
	PaymentDate     string  `json:"payment_date"`
	BankAccountNo   string  `json:"bank_account_no"`
	Source          string  `json:"source"`
	Remark          *string `json:"remark,omitempty"`
	CompanyID       string  `json:"company_id"`
}

// Create handles POST /api/v1/payroll
func (h *PayrollHandler) Create(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	userID := c.Locals("user_id").(uuid.UUID)

	var req PayrollCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	payroll, err := h.svc.CreatePayroll(c.Context(), tenantID, userID, &service.CreatePayrollRequest{
		EmployeeName:    req.EmployeeName,
		DepartmentName:  req.DepartmentName,
		PeriodNo:        req.PeriodNo,
		GrossSalary:     req.GrossSalary,
		IndividualTax:   req.IndividualTax,
		SocialSecurity:  req.SocialSecurity,
		HousingFund:     req.HousingFund,
		OtherDeductions: req.OtherDeductions,
		NetSalary:       req.NetSalary,
		PaymentDate:     req.PaymentDate,
		BankAccountNo:   req.BankAccountNo,
		Source:          req.Source,
		Remark:          req.Remark,
		CompanyID:       req.CompanyID,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"payroll": payroll})
}

// List handles GET /api/v1/payroll
func (h *PayrollHandler) List(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	var periodNo *int
	if p := c.Query("period_no"); p != "" {
		if parsed := c.QueryInt("period_no"); parsed > 0 {
			periodNo = &parsed
		}
	}

	list, err := h.svc.ListPayroll(c.Context(), tenantID, periodNo)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"payrolls": list})
}

// GetByID handles GET /api/v1/payroll/:id
func (h *PayrollHandler) GetByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	tenantID := c.Locals("tenant_id").(uuid.UUID)

	payroll, err := h.svc.GetPayroll(c.Context(), tenantID, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(payroll)
}

// Submit handles POST /api/v1/payroll/:id/submit
func (h *PayrollHandler) Submit(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	tenantID := c.Locals("tenant_id").(uuid.UUID)
	userID := c.Locals("user_id").(uuid.UUID)

	if err := h.svc.Submit(c.Context(), tenantID, id, userID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "payroll submitted"})
}

// Approve handles POST /api/v1/payroll/:id/approve
func (h *PayrollHandler) Approve(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	tenantID := c.Locals("tenant_id").(uuid.UUID)
	userID := c.Locals("user_id").(uuid.UUID)

	voucher, err := h.svc.Approve(c.Context(), tenantID, id, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message":     "payroll approved",
		"voucher_id":  voucher.ID,
		"voucher_no":  voucher.VoucherNo,
	})
}

// GeneratePeriodVouchers handles POST /api/v1/payroll/generate-period-vouchers
func (h *PayrollHandler) GeneratePeriodVouchers(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	userID := c.Locals("user_id").(uuid.UUID)

	var req service.GeneratePeriodVouchersRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.PeriodNo <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "period_no is required"})
	}

	vouchers, err := h.svc.GeneratePeriodVouchers(c.Context(), tenantID, userID, &req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"vouchers": vouchers})
}

// CalculatePeriodSocial handles POST /api/v1/payroll/calculate-period-social
func (h *PayrollHandler) CalculatePeriodSocial(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	userID := c.Locals("user_id").(uuid.UUID)

	var req service.CalculatePeriodSocialRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.PeriodNo <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "period_no is required"})
	}

	result, err := h.svc.CalculatePeriodSocial(c.Context(), tenantID, userID, &req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"result": result})
}

// CalculatePeriodTax handles POST /api/v1/payroll/calculate-period-tax
func (h *PayrollHandler) CalculatePeriodTax(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	userID := c.Locals("user_id").(uuid.UUID)

	var req service.CalculatePeriodTaxRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.PeriodNo <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "period_no is required"})
	}

	result, err := h.svc.CalculatePeriodTax(c.Context(), tenantID, userID, &req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"result": result})
}

// GenerateVoucher handles POST /api/v1/payroll/:id/generate-voucher
func (h *PayrollHandler) GenerateVoucher(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	tenantID := c.Locals("tenant_id").(uuid.UUID)
	userID := c.Locals("user_id").(uuid.UUID)

	voucher, err := h.svc.GenerateVoucherFromPayroll(c.Context(), tenantID, id, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"voucher_id": voucher.ID,
		"voucher_no": voucher.VoucherNo,
	})
}

// ExportSalary handles GET /api/v1/payroll/export-salary
func (h *PayrollHandler) ExportSalary(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	userID := c.Locals("user_id").(uuid.UUID)
	periodNo, err := strconv.Atoi(c.Query("period_no"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "period_no is required"})
	}

	data, err := h.svc.ExportSalaryExcel(c.Context(), tenantID, userID, periodNo)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="salary_%d.xlsx"`, periodNo))
	return c.Send(data)
}

// ExportTax handles GET /api/v1/payroll/export-tax
func (h *PayrollHandler) ExportTax(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	userID := c.Locals("user_id").(uuid.UUID)
	periodNo, err := strconv.Atoi(c.Query("period_no"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "period_no is required"})
	}

	data, err := h.svc.ExportTaxExcel(c.Context(), tenantID, userID, periodNo)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="tax_%d.xlsx"`, periodNo))
	return c.Send(data)
}

// HandleListSocialConfigs handles GET /api/v1/payroll/social-config
func (h *PayrollHandler) HandleListSocialConfigs(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	configs, err := h.svc.ListSocialConfigs(c.Context(), tenantID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"configs": configs})
}

// HandleUpdateSocialConfig handles PUT /api/v1/payroll/social-config/:id
func (h *PayrollHandler) HandleUpdateSocialConfig(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	tenantID := c.Locals("tenant_id").(uuid.UUID)

	var req service.UpdateSocialConfigRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if err := h.svc.UpdateSocialConfig(c.Context(), tenantID, id, &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "social config updated"})
}