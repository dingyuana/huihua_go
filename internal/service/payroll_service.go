package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/model"
	"huihua/finance/internal/repository"
)

// PayrollService handles payroll business logic.
type PayrollService struct {
	payrollRepo   *repository.PayrollRepository
	journalRepo   *repository.JournalRepository
	accountRepo   *repository.AccountRepository
	templateSvc   *VoucherTemplateService
}

// NewPayrollService creates a new PayrollService.
func NewPayrollService(
	payrollRepo *repository.PayrollRepository,
	journalRepo *repository.JournalRepository,
	accountRepo *repository.AccountRepository,
	templateSvc *VoucherTemplateService,
) *PayrollService {
	return &PayrollService{
		payrollRepo: payrollRepo,
		journalRepo: journalRepo,
		accountRepo: accountRepo,
		templateSvc: templateSvc,
	}
}

// CreatePayrollRequest is the request body for creating a payroll record.
type CreatePayrollRequest struct {
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

// CreatePayroll creates a new payroll record (docstatus=0 draft).
func (s *PayrollService) CreatePayroll(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, req *CreatePayrollRequest) (*model.Payroll, error) {
	payrollNo, err := s.payrollRepo.GetNextPayrollNo(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get next payroll no: %w", err)
	}

	grossSalary, err := decimal.NewFromString(req.GrossSalary)
	if err != nil {
		return nil, fmt.Errorf("invalid gross_salary: %w", err)
	}
	individualTax, err := decimal.NewFromString(req.IndividualTax)
	if err != nil {
		return nil, fmt.Errorf("invalid individual_tax: %w", err)
	}
	socialSecurity, err := decimal.NewFromString(req.SocialSecurity)
	if err != nil {
		return nil, fmt.Errorf("invalid social_security: %w", err)
	}
	housingFund, err := decimal.NewFromString(req.HousingFund)
	if err != nil {
		return nil, fmt.Errorf("invalid housing_fund: %w", err)
	}
	otherDeductions, err := decimal.NewFromString(req.OtherDeductions)
	if err != nil {
		return nil, fmt.Errorf("invalid other_deductions: %w", err)
	}
	netSalary, err := decimal.NewFromString(req.NetSalary)
	if err != nil {
		return nil, fmt.Errorf("invalid net_salary: %w", err)
	}

	paymentDate, err := time.Parse("2006-01-02", req.PaymentDate)
	if err != nil {
		return nil, fmt.Errorf("invalid payment_date: %w", err)
	}

	companyID, err := uuid.Parse(req.CompanyID)
	if err != nil {
		return nil, fmt.Errorf("invalid company_id: %w", err)
	}

	source := req.Source
	if source == "" {
		source = "manual"
	}

	payroll := &model.Payroll{
		ID:              uuid.New(),
		CompanyID:       companyID,
		PayrollNo:       payrollNo,
		EmployeeName:    req.EmployeeName,
		DepartmentName:  req.DepartmentName,
		PeriodNo:        req.PeriodNo,
		GrossSalary:     grossSalary,
		IndividualTax:   individualTax,
		SocialSecurity:  socialSecurity,
		HousingFund:     housingFund,
		OtherDeductions: otherDeductions,
		NetSalary:       netSalary,
		PaymentDate:     paymentDate,
		BankAccountNo:   req.BankAccountNo,
		Status:          "draft",
		DocStatus:       0,
		Source:         source,
		Remark:         req.Remark,
		CreatedBy:       &userID,
	}

	return s.payrollRepo.Create(ctx, tenantID, payroll)
}

// GetPayroll retrieves a payroll record by ID.
func (s *PayrollService) GetPayroll(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*model.Payroll, error) {
	return s.payrollRepo.GetByID(ctx, tenantID, id)
}

// ListPayroll lists payroll records with optional period filter.
func (s *PayrollService) ListPayroll(ctx context.Context, tenantID uuid.UUID, periodNo *int) ([]model.Payroll, error) {
	if periodNo != nil {
		return s.payrollRepo.ListByPeriod(ctx, tenantID, *periodNo)
	}
	return s.payrollRepo.ListByTenant(ctx, tenantID)
}

// Submit submits a payroll record: changes docstatus from 0 to 1.
func (s *PayrollService) Submit(ctx context.Context, tenantID uuid.UUID, id uuid.UUID, userID uuid.UUID) error {
	payroll, err := s.payrollRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("get payroll: %w", err)
	}
	if payroll.DocStatus != 0 {
		return errors.New("only draft payrolls can be submitted")
	}
	return s.payrollRepo.UpdateStatus(ctx, tenantID, id, 1)
}

// Approve approves a payroll record: changes docstatus from 1 to 2 and generates a voucher.
// This is a first-class source document — approve directly generates the journal voucher.
func (s *PayrollService) Approve(ctx context.Context, tenantID uuid.UUID, id uuid.UUID, userID uuid.UUID) (*model.JournalEntry, error) {
	payroll, err := s.payrollRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("get payroll: %w", err)
	}
	if payroll.DocStatus != 1 {
		return nil, errors.New("only submitted payrolls can be approved")
	}

	voucher, err := s.GenerateVoucherFromPayroll(ctx, tenantID, id, userID)
	if err != nil {
		return nil, fmt.Errorf("generate voucher: %w", err)
	}

	if err := s.payrollRepo.UpdateStatus(ctx, tenantID, id, 2); err != nil {
		return nil, fmt.Errorf("update status: %w", err)
	}
	if err := s.payrollRepo.SetVoucherID(ctx, tenantID, id, voucher.ID, voucher.VoucherNo); err != nil {
		return nil, fmt.Errorf("update voucher ref: %w", err)
	}

	return voucher, nil
}

// GeneratePeriodVouchersRequest request body
type GeneratePeriodVouchersRequest struct {
	PeriodNo int `json:"period_no"`
}

// GeneratePeriodVouchers generates 3 accrual vouchers for a payroll period.
//  1. 工资计提: Dr 5601(管理费用-工资) / Cr 2211(应付职工薪酬-工资) = sum(gross_salary)
//  2. 社保计提: Dr 5601(管理费用-社保) / Cr 2211(应付职工薪酬-社保) = sum(social_security + housing_fund)
//  3. 个税计提: Dr 2211(应付职工薪酬-个税) / Cr 2221(应交税费-个人所得税) = sum(individual_tax)
//
// All vouchers created with docstatus=1 (submitted for review).
func (s *PayrollService) GeneratePeriodVouchers(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, req *GeneratePeriodVouchersRequest) ([]*model.JournalEntry, error) {
	// 1. List all payroll records for the period
	records, err := s.payrollRepo.ListByPeriod(ctx, tenantID, req.PeriodNo)
	if err != nil {
		return nil, fmt.Errorf("list payroll by period: %w", err)
	}
	if len(records) == 0 {
		return nil, errors.New("no payroll records found for the period")
	}

	// 2. Aggregate totals
	var totalGross, totalSocial, totalTax decimal.Decimal
	for _, r := range records {
		totalGross = totalGross.Add(r.GrossSalary)
		totalSocial = totalSocial.Add(r.SocialSecurity).Add(r.HousingFund)
		totalTax = totalTax.Add(r.IndividualTax)
	}

	// 3. Lookup accounts
	account5601, err := s.accountRepo.GetByCode(ctx, tenantID, "5601")
	if err != nil {
		return nil, fmt.Errorf("lookup account 5601: %w", err)
	}
	account2211, err := s.accountRepo.GetByCode(ctx, tenantID, "2211")
	if err != nil {
		return nil, fmt.Errorf("lookup account 2211: %w", err)
	}
	account2221, err := s.accountRepo.GetByCode(ctx, tenantID, "2221")
	if err != nil {
		return nil, fmt.Errorf("lookup account 2221: %w", err)
	}

	// 4. Build period description
	periodDesc := fmt.Sprintf("%d年%d月薪资计提", req.PeriodNo/100, req.PeriodNo%100)

	// Use the first record's company and posting date
	companyID := records[0].CompanyID
	postingDate := records[0].PaymentDate

	// Helper to create a voucher
	createVoucher := func(desc string, lines []model.JournalEntryLine) (*model.JournalEntry, error) {
		voucherResp, err := s.templateSvc.GenerateVoucherNumber(ctx, tenantID)
		voucherNo := ""
		if err == nil && voucherResp != nil {
			voucherNo = voucherResp.VoucherNumber
		}
		if voucherNo == "" {
			voucherNo = fmt.Sprintf("PY-%s-%d", time.Now().Format("20060102"), time.Now().UnixNano()%1000000)
		}

		remark := periodDesc + " — " + desc
		je := &model.JournalEntry{
			ID:            uuid.New(),
			VoucherNo:     voucherNo,
			VoucherType:   ptr("Payroll"),
			PostingDate:   postingDate,
			CompanyID:     companyID,
			TenantID:      tenantID,
			DocStatus:     1,
			CreatedBy:     userID,
			SourceDocType: ptr("payroll"),
			Remark:        &remark,
		}

		je, err = s.journalRepo.Create(ctx, tenantID, je)
		if err != nil {
			return nil, fmt.Errorf("create journal entry: %w", err)
		}

		_, err = s.journalRepo.AddLines(ctx, tenantID, je.ID, lines)
		if err != nil {
			return nil, fmt.Errorf("add journal entry lines: %w", err)
		}

		return je, nil
	}

	// Voucher 1 - 工资计提: Dr 5601 / Cr 2211 = totalGross
	je1, err := createVoucher("工资计提", []model.JournalEntryLine{
		{
			ID:          uuid.New(),
			AccountID:   account5601.ID,
			Debit:       totalGross,
			Credit:      decimal.Zero,
			AccountCode: account5601.Code,
			AccountName: account5601.Name,
			TenantID:    tenantID,
		},
		{
			ID:          uuid.New(),
			AccountID:   account2211.ID,
			Debit:       decimal.Zero,
			Credit:      totalGross,
			AccountCode: account2211.Code,
			AccountName: account2211.Name,
			TenantID:    tenantID,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("工资计提 voucher: %w", err)
	}

	// Voucher 2 - 社保计提: Dr 5601 / Cr 2211 = totalSocial
	je2, err := createVoucher("社保计提", []model.JournalEntryLine{
		{
			ID:          uuid.New(),
			AccountID:   account5601.ID,
			Debit:       totalSocial,
			Credit:      decimal.Zero,
			AccountCode: account5601.Code,
			AccountName: account5601.Name,
			TenantID:    tenantID,
		},
		{
			ID:          uuid.New(),
			AccountID:   account2211.ID,
			Debit:       decimal.Zero,
			Credit:      totalSocial,
			AccountCode: account2211.Code,
			AccountName: account2211.Name,
			TenantID:    tenantID,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("社保计提 voucher: %w", err)
	}

	// Voucher 3 - 个税计提: Dr 2211 / Cr 2221 = totalTax
	je3, err := createVoucher("个税计提", []model.JournalEntryLine{
		{
			ID:          uuid.New(),
			AccountID:   account2211.ID,
			Debit:       totalTax,
			Credit:      decimal.Zero,
			AccountCode: account2211.Code,
			AccountName: account2211.Name,
			TenantID:    tenantID,
		},
		{
			ID:          uuid.New(),
			AccountID:   account2221.ID,
			Debit:       decimal.Zero,
			Credit:      totalTax,
			AccountCode: account2221.Code,
			AccountName: account2221.Name,
			TenantID:    tenantID,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("个税计提 voucher: %w", err)
	}

	return []*model.JournalEntry{je1, je2, je3}, nil
}

// GenerateVoucherFromPayroll generates a journal voucher from a payroll record.
// Voucher entries:
//   - Debit: 6602 应付职工薪酬 — 工资        (gross_salary)
//   - Credit: 2221 应交税费 — 应交个人所得税 (individual_tax)
//   - Credit: 2241 其他应付款 — 社保          (social_security)
//   - Credit: 2241 其他应付款 — 公积金        (housing_fund)
//   - Credit: 1002 银行存款                (net_salary)
func (s *PayrollService) GenerateVoucherFromPayroll(ctx context.Context, tenantID uuid.UUID, id uuid.UUID, userID uuid.UUID) (*model.JournalEntry, error) {
	payroll, err := s.payrollRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("get payroll: %w", err)
	}

	// Lookup accounts
	account6602, err := s.accountRepo.GetByCode(ctx, tenantID, "6602")
	if err != nil {
		return nil, fmt.Errorf("lookup account 6602: %w", err)
	}
	account2221, err := s.accountRepo.GetByCode(ctx, tenantID, "2221")
	if err != nil {
		return nil, fmt.Errorf("lookup account 2221: %w", err)
	}
	account2241, err := s.accountRepo.GetByCode(ctx, tenantID, "2241")
	if err != nil {
		return nil, fmt.Errorf("lookup account 2241: %w", err)
	}
	account1002, err := s.accountRepo.GetByCode(ctx, tenantID, "1002")
	if err != nil {
		return nil, fmt.Errorf("lookup account 1002: %w", err)
	}

	// Generate voucher number
	voucherResp, err := s.templateSvc.GenerateVoucherNumber(ctx, tenantID)
	voucherNo := ""
	if err == nil && voucherResp != nil {
		voucherNo = voucherResp.VoucherNumber
	}
	if voucherNo == "" {
		voucherNo = fmt.Sprintf("PY-%s-%d", time.Now().Format("20060102"), time.Now().UnixNano()%1000000)
	}

	// Build period description for voucher remark
	periodDesc := fmt.Sprintf("%d年%d月工资", payroll.PeriodNo/100, payroll.PeriodNo%100)
	remark := fmt.Sprintf("%s — %s", periodDesc, payroll.EmployeeName)
	if payroll.Remark != nil && *payroll.Remark != "" {
		remark += "; " + *payroll.Remark
	}

	// Build journal entry
	je := &model.JournalEntry{
		ID:            uuid.New(),
		VoucherNo:     voucherNo,
		VoucherType:   ptr("Payroll"),
		PostingDate:   payroll.PaymentDate,
		CompanyID:     payroll.CompanyID,
		TenantID:      tenantID,
		DocStatus:     1,
		CreatedBy:     userID,
		SourceDocType: ptr("payroll"),
		SourceDocID:   &id,
		SourceDocNo:   &payroll.PayrollNo,
		Remark:        &remark,
	}

	je, err = s.journalRepo.Create(ctx, tenantID, je)
	if err != nil {
		return nil, fmt.Errorf("create journal entry: %w", err)
	}

	// Build lines
	lines := []model.JournalEntryLine{
		{
			ID:             uuid.New(),
			JournalEntryID: je.ID,
			AccountID:      account6602.ID,
			Debit:          payroll.GrossSalary,
			Credit:         decimal.Zero,
			AccountCode:    account6602.Code,
			AccountName:    account6602.Name,
			TenantID:       tenantID,
		},
		{
			ID:             uuid.New(),
			JournalEntryID: je.ID,
			AccountID:      account2221.ID,
			Debit:          decimal.Zero,
			Credit:         payroll.IndividualTax,
			AccountCode:    account2221.Code,
			AccountName:    account2221.Name,
			TenantID:       tenantID,
		},
		{
			ID:             uuid.New(),
			JournalEntryID: je.ID,
			AccountID:      account2241.ID,
			Debit:          decimal.Zero,
			Credit:         payroll.SocialSecurity,
			AccountCode:    account2241.Code,
			AccountName:    account2241.Name,
			TenantID:       tenantID,
		},
		{
			ID:             uuid.New(),
			JournalEntryID: je.ID,
			AccountID:      account2241.ID,
			Debit:          decimal.Zero,
			Credit:         payroll.HousingFund,
			AccountCode:    account2241.Code,
			AccountName:    account2241.Name,
			TenantID:       tenantID,
		},
		{
			ID:             uuid.New(),
			JournalEntryID: je.ID,
			AccountID:      account1002.ID,
			Debit:          decimal.Zero,
			Credit:         payroll.NetSalary,
			AccountCode:    account1002.Code,
			AccountName:    account1002.Name,
			TenantID:       tenantID,
		},
	}

	_, err = s.journalRepo.AddLines(ctx, tenantID, je.ID, lines)
	if err != nil {
		return nil, fmt.Errorf("add journal entry lines: %w", err)
	}

	return je, nil
}