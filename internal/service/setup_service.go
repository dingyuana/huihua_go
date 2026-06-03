package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"huihua/finance/internal/model"
	"huihua/finance/internal/repository"
)

// SetupService handles the account setup wizard.
type SetupService struct {
	companyRepo *repository.CompanyRepository
	periodRepo  *repository.PeriodRepository
	accountSvc  *AccountService
	pool        *pgxpool.Pool
}

// NewSetupService creates a new SetupService.
func NewSetupService(companyRepo *repository.CompanyRepository, periodRepo *repository.PeriodRepository, accountSvc *AccountService, pool *pgxpool.Pool) *SetupService {
	return &SetupService{companyRepo: companyRepo, periodRepo: periodRepo, accountSvc: accountSvc, pool: pool}
}

// CreateCompanyRequest is the request body for creating a new company/account setup.
type CreateCompanyRequest struct {
	CompanyName          string `json:"company_name"`
	FiscalYearStartMonth int    `json:"fiscal_year_start_month"`
	EnableDate           string `json:"enable_date"`
	DefaultCurrency      string `json:"default_currency"`
	ChartTemplate        string `json:"chart_template"`
}

// CreateCompany creates a new company with accounting periods and optionally initializes the chart of accounts.
// All operations run in a single transaction with SET app.current_tenant to satisfy RLS policies.
func (s *SetupService) CreateCompany(ctx context.Context, tenantID uuid.UUID, req CreateCompanyRequest) (*model.CompanySettings, error) {
	// Check if already initialized
	existing, err := s.companyRepo.GetByTenant(ctx, tenantID)
	if err == nil && existing.IsInitialized {
		// Already initialized: update company info without re-creating periods/accounts
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
		if req.ChartTemplate == "small_business" {
			req.ChartTemplate = "small_enterprise"
		}

		existing.CompanyName = req.CompanyName
		existing.FiscalYearStartMonth = req.FiscalYearStartMonth
		existing.EnableDate = enableDate
		existing.DefaultCurrency = req.DefaultCurrency
		existing.ChartOfAccountsTemplate = req.ChartTemplate

		if err := s.companyRepo.Update(ctx, tenantID, existing); err != nil {
			return nil, fmt.Errorf("update company: %w", err)
		}
		return existing, nil
	}
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
	// Normalize chart_template alias
	if req.ChartTemplate == "small_business" {
		req.ChartTemplate = "small_enterprise"
	}

	// Acquire a dedicated connection and run everything in a single transaction
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Set tenant context for RLS on this connection
	if _, err := tx.Exec(ctx, fmt.Sprintf("SET app.current_tenant = '%s'", tenantID)); err != nil {
		return nil, fmt.Errorf("set tenant context: %w", err)
	}

	// Create company settings
	cs := &model.CompanySettings{
		CompanyName:             req.CompanyName,
		FiscalYearStartMonth:    req.FiscalYearStartMonth,
		EnableDate:              enableDate,
		DefaultCurrency:         req.DefaultCurrency,
		ChartOfAccountsTemplate: req.ChartTemplate,
		IsInitialized:           false,
	}
	created, err := s.companyRepo.CreateWithTx(ctx, tx, tenantID, cs)
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
	if err := s.periodRepo.BatchCreateWithTx(ctx, tx, tenantID, periods); err != nil {
		return nil, fmt.Errorf("create periods: %w", err)
	}

	// Initialize chart of accounts from seed
	if err := s.accountSvc.InitFromSeedWithTx(ctx, tx, tenantID, created.ID); err != nil {
		return nil, fmt.Errorf("init chart of accounts: %w", err)
	}

	// Mark as initialized
	if err := s.companyRepo.SetInitializedWithTx(ctx, tx, tenantID); err != nil {
		return nil, fmt.Errorf("mark initialized: %w", err)
	}
	created.IsInitialized = true

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

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
		"initialized":   cs.IsInitialized,
		"company":       cs,
		"periods_count": len(periods),
		"periods":       periods,
	}, nil
}
