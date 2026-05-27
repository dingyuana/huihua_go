package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"huihua/finance/internal/model"
	"huihua/finance/internal/repository"
)

// SetupService handles the account setup wizard.
type SetupService struct {
	companyRepo  *repository.CompanyRepository
	periodRepo   *repository.PeriodRepository
	accountSvc   *AccountService
}

// NewSetupService creates a new SetupService.
func NewSetupService(companyRepo *repository.CompanyRepository, periodRepo *repository.PeriodRepository, accountSvc *AccountService) *SetupService {
	return &SetupService{companyRepo: companyRepo, periodRepo: periodRepo, accountSvc: accountSvc}
}

// CreateCompanyRequest is the request body for creating a new company/account setup.
type CreateCompanyRequest struct {
	CompanyName         string `json:"company_name"`
	FiscalYearStartMonth int    `json:"fiscal_year_start_month"`
	EnableDate          string `json:"enable_date"`
	DefaultCurrency     string `json:"default_currency"`
	ChartTemplate       string `json:"chart_template"`
}

// CreateCompany creates a new company with accounting periods and optionally initializes the chart of accounts.
func (s *SetupService) CreateCompany(ctx context.Context, tenantID uuid.UUID, req CreateCompanyRequest) (*model.CompanySettings, error) {
	if req.FiscalYearStartMonth < 1 || req.FiscalYearStartMonth > 12 {
		return nil, fmt.Errorf("fiscal_year_start_month must be between 1 and 12")
	}
	enableDate, err := time.Parse("2006-01-02", req.EnableDate)
	if err != nil {
		return nil, fmt.Errorf("enable_date must be in YYYY-MM-DD format")
	}
	if req.DefaultCurrency == "" {
		req.DefaultCurrency = "CNY"
	}
	if req.ChartTemplate == "" {
		req.ChartTemplate = "small_enterprise"
	}

	// Create company settings
	cs := &model.CompanySettings{
		CompanyName:          req.CompanyName,
		FiscalYearStartMonth: req.FiscalYearStartMonth,
		EnableDate:           enableDate,
		DefaultCurrency:      req.DefaultCurrency,
		ChartTemplate:        req.ChartTemplate,
		IsInitialized:        false,
	}
	created, err := s.companyRepo.Create(ctx, tenantID, cs)
	if err != nil {
		return nil, fmt.Errorf("create company: %w", err)
	}

	// Generate 12 accounting periods
	year := enableDate.Year()
	periods := make([]model.AccountingPeriod, 12)
	for i := 0; i < 12; i++ {
		month := (req.FiscalYearStartMonth - 1 + i) % 12
		adjustedYear := year
		if req.FiscalYearStartMonth > 1 && month < req.FiscalYearStartMonth-1 {
			adjustedYear++
		}
		start := time.Date(adjustedYear, time.Month(month+1), 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(adjustedYear, time.Month(month+2), 0, 23, 59, 59, 0, time.UTC)
		periods[i] = model.AccountingPeriod{
			PeriodNo:   adjustedYear*100 + month + 1,
			PeriodName: fmt.Sprintf("%d年%d月", adjustedYear, month+1),
			StartDate:  start,
			EndDate:    end,
			Status:     "open",
		}
	}
	if err := s.periodRepo.BatchCreate(ctx, tenantID, periods); err != nil {
		return nil, fmt.Errorf("create periods: %w", err)
	}

	// Initialize chart of accounts from seed
	if req.ChartTemplate == "small_enterprise" {
		if err := s.accountSvc.InitFromSeed(ctx, tenantID, created.ID); err != nil {
			return nil, fmt.Errorf("init chart of accounts: %w", err)
		}
	}

	// Mark as initialized
	if err := s.companyRepo.SetInitialized(ctx, tenantID); err != nil {
		return nil, fmt.Errorf("mark initialized: %w", err)
	}
	created.IsInitialized = true

	return created, nil
}

// GetStatus returns the current setup status for a tenant.
func (s *SetupService) GetStatus(ctx context.Context, tenantID uuid.UUID) (map[string]interface{}, error) {
	cs, err := s.companyRepo.GetByTenant(ctx, tenantID)
	if err != nil {
		return map[string]interface{}{"initialized": false}, nil
	}
	periods, err := s.periodRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"initialized":    cs.IsInitialized,
		"company":        cs,
		"periods_count":  len(periods),
		"periods":        periods,
	}, nil
}