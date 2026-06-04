package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/model"
	"huihua/finance/internal/repository"
)

// AssetDepreciationService handles asset depreciation operations.
type AssetDepreciationService struct {
	depreciationRepo *repository.AssetDepreciationRepository
	journalRepo      *repository.JournalRepository
}

// NewAssetDepreciationService creates a new AssetDepreciationService.
func NewAssetDepreciationService(
	depreciationRepo *repository.AssetDepreciationRepository,
	journalRepo *repository.JournalRepository,
) *AssetDepreciationService {
	return &AssetDepreciationService{
		depreciationRepo: depreciationRepo,
		journalRepo:      journalRepo,
	}
}

// CalculateStraightLine calculates monthly depreciation using straight-line method.
// Formula: (cost - salvage_value) / useful_life_months
func (s *AssetDepreciationService) CalculateStraightLine(
	depreciableAmount decimal.Decimal,
	usefulLifeMonths int,
) decimal.Decimal {
	if usefulLifeMonths <= 0 {
		return decimal.Zero
	}
	return depreciableAmount.Div(decimal.NewFromInt(int64(usefulLifeMonths)))
}

// CalculateDoubleDeclining calculates depreciation using double-declining balance method.
// For a specific period, returns the depreciation amount.
// purchaseDate is used to determine the period start.
// periodNo is the accounting period number (YYYYMM format).
func (s *AssetDepreciationService) CalculateDoubleDeclining(
	purchaseDate time.Time,
	originalValue decimal.Decimal,
	usefulLifeMonths int,
	periodNo int,
) decimal.Decimal {
	if usefulLifeMonths <= 0 {
		return decimal.Zero
	}

	// Calculate useful life in years
	usefulLifeYears := float64(usefulLifeMonths) / 12.0
	if usefulLifeYears <= 0 {
		return decimal.Zero
	}

	// Double declining rate: 2 / useful_life_years
	rate := decimal.NewFromFloat(2.0 / usefulLifeYears)

	// Monthly rate
	monthlyRate := rate.Div(decimal.NewFromInt(12))

	// Calculate accumulated depreciation up to this period
	// Period format: YYYYMM, so extract year and month
	periodYear := periodNo / 100
	periodMonth := periodNo % 100

	// Calculate how many months since purchase
	yearsSince := periodYear - purchaseDate.Year()
	monthsSince := yearsSince*12 + (periodMonth - int(purchaseDate.Month()))

	if monthsSince <= 0 {
		return originalValue.Mul(monthlyRate)
	}

	// Book value at start of this period = originalValue * (1 - rate)^monthsSince
	// But for simplicity, we use the remaining book value approach
	// Note: This is a simplified calculation; real DDB may need iteration
	bookValue := originalValue

	// Simple approach: use remaining book value at start of this period
	// In production, you'd track accumulated depreciation separately
	accumulated := decimal.Zero
	for i := 0; i < monthsSince; i++ {
		depr := bookValue.Mul(monthlyRate)
		if depr.GreaterThan(bookValue.Sub(decimal.NewFromFloat(0.01))) {
			depr = bookValue.Sub(decimal.NewFromFloat(0.01))
		}
		accumulated = accumulated.Add(depr)
		bookValue = bookValue.Sub(depr)
		if bookValue.LessThan(decimal.NewFromFloat(0.01)) {
			return decimal.Zero
		}
	}
	_ = accumulated // track if needed in future

	// Current period depreciation
	currentDepr := bookValue.Mul(monthlyRate)
	if currentDepr.GreaterThan(bookValue) {
		currentDepr = bookValue
	}

	return currentDepr
}

// GenerateMonthlyDepreciation generates journal entries for depreciation in a period.
// It processes all unposted depreciation schedules for the period.
// Generated voucher is in draft status (docstatus=0) and needs human review before posting via VoucherStateMachine.
func (s *AssetDepreciationService) GenerateMonthlyDepreciation(
	ctx context.Context,
	tenantID uuid.UUID,
	periodNo int,
) (*model.DepreciationRun, error) {
	// Get all unposted schedules for the period
	schedules, err := s.depreciationRepo.GetUnpostedSchedulesByPeriod(ctx, tenantID, periodNo)
	if err != nil {
		return nil, fmt.Errorf("get unposted schedules: %w", err)
	}

	if len(schedules) == 0 {
		return nil, fmt.Errorf("no unposted depreciation schedules found for period %d", periodNo)
	}

	// Group schedules by asset to get company info
	assetIDs := make(map[uuid.UUID]bool)
	for _, sch := range schedules {
		assetIDs[sch.AssetID] = true
	}

	// Get company ID from first asset
	var companyID uuid.UUID
	firstAssetID := uuid.Nil
	for aid := range assetIDs {
		firstAssetID = aid
		break
	}
	if firstAssetID != uuid.Nil {
		asset, err := s.depreciationRepo.GetAssetByID(ctx, tenantID, firstAssetID)
		if err != nil {
			return nil, fmt.Errorf("get asset company: %w", err)
		}
		companyID = asset.CompanyID
	}

	// Generate voucher number
	voucherNo := fmt.Sprintf("DEP-%d-%d", periodNo, time.Now().UnixNano()%1000000)

	// Create journal entry for the entire depreciation run
	je := &model.JournalEntry{
		ID:          uuid.New(),
		VoucherNo:   voucherNo,
		VoucherType: func() *string { s := "Depreciation"; return &s }(),
		PostingDate: time.Now(),
		CompanyID:   companyID,
		DocStatus:   0, // 草稿，等人审核
		CreatedBy:   uuid.Nil, // system
	}

	je, err = s.journalRepo.Create(ctx, tenantID, je)
	if err != nil {
		return nil, fmt.Errorf("create journal entry: %w", err)
	}

	// Track total amounts for summary
	totalAmount := decimal.Zero
	processedAssets := make(map[uuid.UUID]bool)
	assetCount := 0

	// Process each schedule
	for _, schedule := range schedules {
		asset, err := s.depreciationRepo.GetAssetByID(ctx, tenantID, schedule.AssetID)
		if err != nil {
			continue // skip invalid assets
		}

		// Ensure accounts are set
		if asset.DepreciationExpenseAccountID == nil || asset.AccumulatedDepreciationAccountID == nil {
			continue // skip assets without depreciation accounts
		}

		// Create journal entry lines
		// Debit: Depreciation Expense Account
		// Credit: Accumulated Depreciation Account
		lines := []model.JournalEntryLine{
			{
				ID:             uuid.New(),
				JournalEntryID: je.ID,
				AccountID:      *asset.DepreciationExpenseAccountID,
				Debit:          schedule.DepreciationAmount,
				Credit:         decimal.Zero,
				DebitCcy:       schedule.DepreciationAmount,
				CreditCcy:      decimal.Zero,
				ExchangeRate:   decimal.NewFromInt(1),
				Reconciled:     false,
			},
			{
				ID:             uuid.New(),
				JournalEntryID: je.ID,
				AccountID:      *asset.AccumulatedDepreciationAccountID,
				Debit:          decimal.Zero,
				Credit:         schedule.DepreciationAmount,
				DebitCcy:       decimal.Zero,
				CreditCcy:      schedule.DepreciationAmount,
				ExchangeRate:   decimal.NewFromInt(1),
				Reconciled:     false,
			},
		}

		_, err = s.journalRepo.AddLines(ctx, tenantID, je.ID, lines)
		if err != nil {
			continue // skip on line creation error
		}

		// Mark schedule as posted
		if err := s.depreciationRepo.MarkAsPosted(ctx, schedule.ID, je.ID); err != nil {
			continue
		}

		totalAmount = totalAmount.Add(schedule.DepreciationAmount)
		if !processedAssets[schedule.AssetID] {
			processedAssets[schedule.AssetID] = true
			assetCount++
		}
	}

	// Create depreciation run record
	run := &model.DepreciationRun{
		ID:          uuid.New(),
		PeriodNo:    periodNo,
		RunDate:     time.Now(),
		TenantID:    tenantID,
		CompanyID:   companyID,
		VoucherNo:   voucherNo,
		VoucherType: je.VoucherType,
		TotalAmount: totalAmount,
		AssetCount:  assetCount,
		Status:      "completed",
	}

	if err := s.depreciationRepo.CreateDepreciationRun(ctx, run); err != nil {
		return nil, fmt.Errorf("create depreciation run: %w", err)
	}

	return run, nil
}

// GetAssetDepreciationSchedule retrieves the depreciation schedule for an asset.
func (s *AssetDepreciationService) GetAssetDepreciationSchedule(
	ctx context.Context,
	tenantID, assetID uuid.UUID,
) ([]model.AssetDepreciation, error) {
	return s.depreciationRepo.GetScheduleByAsset(ctx, tenantID, assetID)
}

// CreateSchedule creates a new depreciation schedule for an asset.
func (s *AssetDepreciationService) CreateSchedule(
	ctx context.Context,
	tenantID, assetID uuid.UUID,
	req model.CreateScheduleRequest,
) ([]model.AssetDepreciation, error) {
	// Check if any posted schedules exist (can't regenerate if depreciation has started)
	exists, err := s.depreciationRepo.CheckScheduleExists(ctx, tenantID, assetID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("cannot regenerate schedule: depreciation has already been posted for this asset")
	}

	// Delete existing unposted schedules
	if err := s.depreciationRepo.DeleteSchedulesByAsset(ctx, tenantID, assetID); err != nil {
		return nil, err
	}

	// Create new schedule
	return s.depreciationRepo.CreateSchedule(
		ctx, tenantID, assetID,
		req.Method, req.UsefulLife, req.SalvageValue, req.PurchaseDate,
	)
}

// GetDepreciationRuns retrieves depreciation runs, optionally filtered by period.
func (s *AssetDepreciationService) GetDepreciationRuns(
	ctx context.Context,
	tenantID uuid.UUID,
	periodNo int,
) ([]model.DepreciationRun, error) {
	if periodNo > 0 {
		return s.depreciationRepo.GetDepreciationRunsByPeriod(ctx, tenantID, periodNo)
	}
	return s.depreciationRepo.ListDepreciationRuns(ctx, tenantID)
}

// GenerateMonthlyAmortization generates journal entries for intangible asset amortization.
// 借：6603 无形资产摊销费，贷：1702 累计摊销
// 生成的凭证为草稿状态（docstatus=0），人审核后过账。
func (s *AssetDepreciationService) GenerateMonthlyAmortization(
	ctx context.Context,
	tenantID uuid.UUID,
	periodNo int,
) (*model.DepreciationRun, error) {
	// 复用 GetUnpostedSchedulesByPeriod 获取无形资产的摊销计划
	// 如果 repository 不区分 asset_type，用现有方法然后过滤
	schedules, err := s.depreciationRepo.GetUnpostedSchedulesByPeriod(ctx, tenantID, periodNo)
	if err != nil {
		return nil, fmt.Errorf("get unposted schedules: %w", err)
	}

	if len(schedules) == 0 {
		return nil, fmt.Errorf("no unposted amortization schedules found for period %d", periodNo)
	}

	// Get company ID from first asset
	var companyID uuid.UUID
	for _, sch := range schedules {
		asset, err := s.depreciationRepo.GetAssetByID(ctx, tenantID, sch.AssetID)
		if err != nil {
			continue
		}
		companyID = asset.CompanyID
		break
	}

	// Generate voucher number
	voucherNo := fmt.Sprintf("AMORT-%d-%d", periodNo, time.Now().UnixNano()%1000000)

	// Create journal entry for amortization
	je := &model.JournalEntry{
		ID:          uuid.New(),
		VoucherNo:   voucherNo,
		VoucherType: func() *string { s := "Amortization"; return &s }(),
		PostingDate: time.Now(),
		CompanyID:   companyID,
		DocStatus:   0, // 草稿，等人审核
		CreatedBy:   uuid.Nil,
	}

	je, err = s.journalRepo.Create(ctx, tenantID, je)
	if err != nil {
		return nil, fmt.Errorf("create journal entry: %w", err)
	}

	totalAmount := decimal.Zero
	processedAssets := make(map[uuid.UUID]bool)
	assetCount := 0

	for _, schedule := range schedules {
		asset, err := s.depreciationRepo.GetAssetByID(ctx, tenantID, schedule.AssetID)
		if err != nil {
			continue
		}

		if asset.DepreciationExpenseAccountID == nil || asset.AccumulatedDepreciationAccountID == nil {
			continue
		}

		lines := []model.JournalEntryLine{
			{
				ID:             uuid.New(),
				JournalEntryID: je.ID,
				AccountID:      *asset.DepreciationExpenseAccountID,
				Debit:          schedule.DepreciationAmount,
				Credit:         decimal.Zero,
				DebitCcy:       schedule.DepreciationAmount,
				CreditCcy:      decimal.Zero,
				ExchangeRate:   decimal.NewFromInt(1),
				Reconciled:     false,
			},
			{
				ID:             uuid.New(),
				JournalEntryID: je.ID,
				AccountID:      *asset.AccumulatedDepreciationAccountID,
				Debit:          decimal.Zero,
				Credit:         schedule.DepreciationAmount,
				DebitCcy:       decimal.Zero,
				CreditCcy:      schedule.DepreciationAmount,
				ExchangeRate:   decimal.NewFromInt(1),
				Reconciled:     false,
			},
		}

		_, err = s.journalRepo.AddLines(ctx, tenantID, je.ID, lines)
		if err != nil {
			continue
		}

		if err := s.depreciationRepo.MarkAsPosted(ctx, schedule.ID, je.ID); err != nil {
			continue
		}

		totalAmount = totalAmount.Add(schedule.DepreciationAmount)
		if !processedAssets[schedule.AssetID] {
			processedAssets[schedule.AssetID] = true
			assetCount++
		}
	}

	run := &model.DepreciationRun{
		ID:          uuid.New(),
		PeriodNo:    periodNo,
		RunDate:     time.Now(),
		TenantID:    tenantID,
		CompanyID:   companyID,
		VoucherNo:   voucherNo,
		VoucherType: je.VoucherType,
		TotalAmount: totalAmount,
		AssetCount:  assetCount,
		Status:      "completed",
	}

	if err := s.depreciationRepo.CreateDepreciationRun(ctx, run); err != nil {
		return nil, fmt.Errorf("create depreciation run: %w", err)
	}

	return run, nil
}
