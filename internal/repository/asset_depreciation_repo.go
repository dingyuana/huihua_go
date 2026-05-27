package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/model"
)

// AssetDepreciationRepository provides data access for depreciation operations.
type AssetDepreciationRepository struct {
	pool *pgxpool.Pool
}

// NewAssetDepreciationRepository creates a new AssetDepreciationRepository.
func NewAssetDepreciationRepository(pool *pgxpool.Pool) *AssetDepreciationRepository {
	return &AssetDepreciationRepository{pool: pool}
}

// CreateSchedule generates a depreciation schedule for an asset.
// It calculates monthly depreciation based on the method and creates schedule entries.
func (r *AssetDepreciationRepository) CreateSchedule(
	ctx context.Context,
	tenantID, assetID uuid.UUID,
	method model.DepreciationMethod,
	usefulLifeMonths int,
	salvageValue decimal.Decimal,
	purchaseDate time.Time,
) ([]model.AssetDepreciation, error) {
	// Get the asset's gross purchase amount
	var grossAmount decimal.Decimal
	err := r.pool.QueryRow(ctx, `
		SELECT gross_purchase_amount FROM assets WHERE id = $1 AND tenant_id = $2
	`, assetID, tenantID).Scan(&grossAmount)
	if err != nil {
		return nil, fmt.Errorf("get asset purchase amount: %w", err)
	}

	// Calculate depreciable amount
	depreciableAmount := grossAmount.Sub(salvageValue)

	// Generate schedule entries based on method
	var schedules []model.AssetDepreciation

	switch method {
	case model.DepreciationMethodStraightLine:
		monthlyAmount := depreciableAmount.Div(decimal.NewFromInt(int64(usefulLifeMonths)))
		for i := 0; i < usefulLifeMonths; i++ {
			scheduleDate := purchaseDate.AddDate(0, i+1, 0)
			schedules = append(schedules, model.AssetDepreciation{
				ID:                 uuid.New(),
				AssetID:            assetID,
				ScheduleDate:       scheduleDate,
				DepreciationAmount: monthlyAmount,
				Posted:             false,
				TenantID:           tenantID,
			})
		}

	case model.DepreciationMethodDoubleDeclining:
		// Double declining balance: 2 / useful_life_years * book_value
		// First convert usefulLifeMonths to years for the rate calculation
		usefulLifeYears := float64(usefulLifeMonths) / 12.0
		rate := decimal.NewFromFloat(2.0 / usefulLifeYears)

		bookValue := grossAmount
		for i := 0; i < usefulLifeMonths; i++ {
			// Calculate depreciation for this period
			depr := bookValue.Mul(rate).Div(decimal.NewFromInt(12))
			// Cap at remaining book value minus salvage
			remaining := bookValue.Sub(salvageValue)
			if depr.GreaterThan(remaining) {
				depr = remaining
			}
			if depr.LessThan(decimal.Zero) {
				depr = decimal.Zero
			}

			scheduleDate := purchaseDate.AddDate(0, i+1, 0)
			schedules = append(schedules, model.AssetDepreciation{
				ID:                 uuid.New(),
				AssetID:            assetID,
				ScheduleDate:       scheduleDate,
				DepreciationAmount: depr,
				Posted:             false,
				TenantID:           tenantID,
			})
			bookValue = bookValue.Sub(depr)
		}

	default:
		return nil, fmt.Errorf("unsupported depreciation method: %s", method)
	}

	// Insert schedules
	for _, s := range schedules {
		_, err := r.pool.Exec(ctx, `
			INSERT INTO depreciation_schedules (id, asset_id, schedule_date, depreciation_amount, posted, tenant_id)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, s.ID, s.AssetID, s.ScheduleDate, s.DepreciationAmount, s.Posted, s.TenantID)
		if err != nil {
			return nil, fmt.Errorf("insert depreciation schedule: %w", err)
		}
	}

	return schedules, nil
}

// GetScheduleByAsset retrieves all depreciation schedules for a specific asset.
func (r *AssetDepreciationRepository) GetScheduleByAsset(ctx context.Context, tenantID, assetID uuid.UUID) ([]model.AssetDepreciation, error) {
	query := `
		SELECT id, asset_id, schedule_date, depreciation_amount, posted, journal_entry_id, tenant_id
		FROM depreciation_schedules
		WHERE asset_id = $1 AND tenant_id = $2
		ORDER BY schedule_date ASC
	`

	rows, err := r.pool.Query(ctx, query, assetID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("query depreciation schedules: %w", err)
	}
	defer rows.Close()

	var schedules []model.AssetDepreciation
	for rows.Next() {
		var s model.AssetDepreciation
		if err := rows.Scan(&s.ID, &s.AssetID, &s.ScheduleDate, &s.DepreciationAmount, &s.Posted, &s.JournalEntryID, &s.TenantID); err != nil {
			return nil, fmt.Errorf("scan depreciation schedule: %w", err)
		}
		schedules = append(schedules, s)
	}
	return schedules, rows.Err()
}

// GetDepreciationByPeriod retrieves all depreciation schedules for a specific period.
func (r *AssetDepreciationRepository) GetDepreciationByPeriod(ctx context.Context, tenantID uuid.UUID, periodNo int) ([]model.AssetDepreciation, error) {
	// Get the period date range
	var startDate, endDate time.Time
	err := r.pool.QueryRow(ctx, `
		SELECT start_date, end_date FROM accounting_periods WHERE period_no = $1 AND tenant_id = $2
	`, periodNo, tenantID).Scan(&startDate, &endDate)
	if err != nil {
		return nil, fmt.Errorf("get period: %w", err)
	}

	query := `
		SELECT id, asset_id, schedule_date, depreciation_amount, posted, journal_entry_id, tenant_id
		FROM depreciation_schedules
		WHERE tenant_id = $1 AND schedule_date >= $2 AND schedule_date <= $3
		ORDER BY asset_id, schedule_date ASC
	`

	rows, err := r.pool.Query(ctx, query, tenantID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("query depreciation by period: %w", err)
	}
	defer rows.Close()

	var schedules []model.AssetDepreciation
	for rows.Next() {
		var s model.AssetDepreciation
		if err := rows.Scan(&s.ID, &s.AssetID, &s.ScheduleDate, &s.DepreciationAmount, &s.Posted, &s.JournalEntryID, &s.TenantID); err != nil {
			return nil, fmt.Errorf("scan depreciation schedule: %w", err)
		}
		schedules = append(schedules, s)
	}
	return schedules, rows.Err()
}

// GetUnpostedSchedulesByPeriod retrieves unposted depreciation schedules for a period.
func (r *AssetDepreciationRepository) GetUnpostedSchedulesByPeriod(ctx context.Context, tenantID uuid.UUID, periodNo int) ([]model.AssetDepreciation, error) {
	// Get the period date range
	var startDate, endDate time.Time
	err := r.pool.QueryRow(ctx, `
		SELECT start_date, end_date FROM accounting_periods WHERE period_no = $1 AND tenant_id = $2
	`, periodNo, tenantID).Scan(&startDate, &endDate)
	if err != nil {
		return nil, fmt.Errorf("get period: %w", err)
	}

	query := `
		SELECT ds.id, ds.asset_id, ds.schedule_date, ds.depreciation_amount, ds.posted, ds.journal_entry_id, ds.tenant_id
		FROM depreciation_schedules ds
		WHERE ds.tenant_id = $1 
		  AND ds.schedule_date >= $2 AND ds.schedule_date <= $3
		  AND ds.posted = false
		ORDER BY ds.asset_id, ds.schedule_date ASC
	`

	rows, err := r.pool.Query(ctx, query, tenantID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("query unposted depreciation: %w", err)
	}
	defer rows.Close()

	var schedules []model.AssetDepreciation
	for rows.Next() {
		var s model.AssetDepreciation
		if err := rows.Scan(&s.ID, &s.AssetID, &s.ScheduleDate, &s.DepreciationAmount, &s.Posted, &s.JournalEntryID, &s.TenantID); err != nil {
			return nil, fmt.Errorf("scan depreciation schedule: %w", err)
		}
		schedules = append(schedules, s)
	}
	return schedules, rows.Err()
}

// MarkAsPosted marks a depreciation schedule as posted with the journal entry ID.
func (r *AssetDepreciationRepository) MarkAsPosted(ctx context.Context, id, journalEntryID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE depreciation_schedules SET posted = true, journal_entry_id = $1 WHERE id = $2
	`, journalEntryID, id)
	if err != nil {
		return fmt.Errorf("mark schedule as posted: %w", err)
	}
	return nil
}

// GetAssetByID retrieves asset info needed for depreciation posting.
func (r *AssetDepreciationRepository) GetAssetByID(ctx context.Context, tenantID, assetID uuid.UUID) (*model.Asset, error) {
	query := `
		SELECT id, asset_name, fixed_asset_account_id, depreciation_expense_account_id, 
		       accumulated_depreciation_account_id, company_id, tenant_id
		FROM assets
		WHERE id = $1 AND tenant_id = $2
	`

	var asset model.Asset
	err := r.pool.QueryRow(ctx, query, assetID, tenantID).Scan(
		&asset.ID, &asset.AssetName, &asset.FixedAssetAccountID,
		&asset.DepreciationExpenseAccountID, &asset.AccumulatedDepreciationAccountID,
		&asset.CompanyID, &asset.TenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("get asset: %w", err)
	}
	return &asset, nil
}

// CreateDepreciationRun creates a depreciation run record.
func (r *AssetDepreciationRepository) CreateDepreciationRun(ctx context.Context, run *model.DepreciationRun) error {
	query := `
		INSERT INTO depreciation_runs (id, period_no, run_date, tenant_id, company_id, voucher_no, voucher_type, total_amount, asset_count, status, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING created_at
	`

	if run.ID == uuid.Nil {
		run.ID = uuid.New()
	}

	err := r.pool.QueryRow(ctx, query,
		run.ID, run.PeriodNo, run.RunDate, run.TenantID, run.CompanyID,
		run.VoucherNo, run.VoucherType, run.TotalAmount, run.AssetCount, run.Status, run.CreatedBy,
	).Scan(&run.CreatedAt)
	if err != nil {
		return fmt.Errorf("create depreciation run: %w", err)
	}
	return nil
}

// GetDepreciationRunsByPeriod retrieves depreciation runs for a period.
func (r *AssetDepreciationRepository) GetDepreciationRunsByPeriod(ctx context.Context, tenantID uuid.UUID, periodNo int) ([]model.DepreciationRun, error) {
	query := `
		SELECT id, period_no, run_date, tenant_id, company_id, voucher_no, voucher_type, total_amount, asset_count, status, created_by, created_at
		FROM depreciation_runs
		WHERE tenant_id = $1 AND period_no = $2
		ORDER BY run_date DESC
	`

	rows, err := r.pool.Query(ctx, query, tenantID, periodNo)
	if err != nil {
		return nil, fmt.Errorf("query depreciation runs: %w", err)
	}
	defer rows.Close()

	var runs []model.DepreciationRun
	for rows.Next() {
		var run model.DepreciationRun
		if err := rows.Scan(
			&run.ID, &run.PeriodNo, &run.RunDate, &run.TenantID, &run.CompanyID,
			&run.VoucherNo, &run.VoucherType, &run.TotalAmount, &run.AssetCount, &run.Status,
			&run.CreatedBy, &run.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan depreciation run: %w", err)
		}
		runs = append(runs, run)
	}
	return runs, rows.Err()
}

// ListDepreciationRuns lists all depreciation runs for a tenant.
func (r *AssetDepreciationRepository) ListDepreciationRuns(ctx context.Context, tenantID uuid.UUID) ([]model.DepreciationRun, error) {
	query := `
		SELECT id, period_no, run_date, tenant_id, company_id, voucher_no, voucher_type, total_amount, asset_count, status, created_by, created_at
		FROM depreciation_runs
		WHERE tenant_id = $1
		ORDER BY run_date DESC
	`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("query depreciation runs: %w", err)
	}
	defer rows.Close()

	var runs []model.DepreciationRun
	for rows.Next() {
		var run model.DepreciationRun
		if err := rows.Scan(
			&run.ID, &run.PeriodNo, &run.RunDate, &run.TenantID, &run.CompanyID,
			&run.VoucherNo, &run.VoucherType, &run.TotalAmount, &run.AssetCount, &run.Status,
			&run.CreatedBy, &run.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan depreciation run: %w", err)
		}
		runs = append(runs, run)
	}
	return runs, rows.Err()
}

// DeleteSchedulesByAsset deletes all depreciation schedules for an asset (for regeneration).
func (r *AssetDepreciationRepository) DeleteSchedulesByAsset(ctx context.Context, tenantID, assetID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		DELETE FROM depreciation_schedules WHERE asset_id = $1 AND tenant_id = $2 AND posted = false
	`, assetID, tenantID)
	if err != nil {
		return fmt.Errorf("delete schedules: %w", err)
	}
	return nil
}

// CheckScheduleExists checks if any posted schedule exists for an asset.
func (r *AssetDepreciationRepository) CheckScheduleExists(ctx context.Context, tenantID, assetID uuid.UUID) (bool, error) {
	var count int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM depreciation_schedules WHERE asset_id = $1 AND tenant_id = $2 AND posted = true
	`, assetID, tenantID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check schedule exists: %w", err)
	}
	return count > 0, nil
}