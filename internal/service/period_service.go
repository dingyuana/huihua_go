package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/model"
	"huihua/finance/internal/repository"
)

// PeriodService handles accounting period operations including closing.
type PeriodService struct {
	periodRepo          *repository.PeriodRepository
	journalRepo         *repository.JournalRepository
	glEntryRepo         *repository.GLEntryRepository
	accountRepo         *repository.AccountRepository
	depreciationRepo    *repository.AssetDepreciationRepository
	bankTransactionRepo *repository.BankTransactionRepository
	invoiceRepo         *repository.InvoiceRepository
}

// NewPeriodService creates a new PeriodService.
func NewPeriodService(periodRepo *repository.PeriodRepository, journalRepo *repository.JournalRepository, glEntryRepo *repository.GLEntryRepository, accountRepo *repository.AccountRepository, depreciationRepo *repository.AssetDepreciationRepository, bankTransactionRepo *repository.BankTransactionRepository, invoiceRepo *repository.InvoiceRepository) *PeriodService {
	return &PeriodService{
		periodRepo:          periodRepo,
		journalRepo:         journalRepo,
		glEntryRepo:         glEntryRepo,
		accountRepo:         accountRepo,
		depreciationRepo:    depreciationRepo,
		bankTransactionRepo: bankTransactionRepo,
		invoiceRepo:         invoiceRepo,
	}
}

// ClosePeriodRequest is the request to close an accounting period.
type ClosePeriodRequest struct {
	PeriodNo               int    `json:"period_no"`
	UserID                 string `json:"user_id"`
	UserName               string `json:"user_name"`
	GenerateClosingEntries bool   `json:"generate_closing_entries"`
}

// ClosePeriod closes an accounting period and optionally generates closing entries.
func (s *PeriodService) ClosePeriod(ctx context.Context, tenantID uuid.UUID, req *ClosePeriodRequest) error {
	// Get period
	periods, err := s.periodRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("list periods: %w", err)
	}

	var period *model.AccountingPeriod
	for _, p := range periods {
		if p.PeriodNo == req.PeriodNo {
			period = &p
			break
		}
	}
	if period == nil {
		return errors.New("period not found")
	}

	if period.Status == "closed" {
		return errors.New("period is already closed")
	}

	// Check for unapproved vouchers in the period
	vouchers, err := s.journalRepo.GetPostedByPeriod(ctx, tenantID, fmt.Sprintf("%d", req.PeriodNo))
	if err != nil {
		return fmt.Errorf("get posted vouchers: %w", err)
	}

	// Check all posted vouchers are verified (docstatus=2) or already reversed/cancelled
	for _, v := range vouchers {
		if v.DocStatus == 1 {
			return fmt.Errorf("voucher %s is still pending approval, cannot close period", v.VoucherNo)
		}
	}

	if req.GenerateClosingEntries {
		// Generate closing entries: close income/expense to retained earnings
		if err := s.generateClosingEntries(ctx, tenantID, *period, req.UserID, req.UserName); err != nil {
			return fmt.Errorf("generate closing entries: %w", err)
		}
	}

	// Update period status to closed
	userID, _ := uuid.Parse(req.UserID)
	if err := s.periodRepo.UpdateStatus(ctx, tenantID, req.PeriodNo, "closed", userID); err != nil {
		return fmt.Errorf("update period status: %w", err)
	}

	return nil
}

// UnclosePeriod re-opens a closed accounting period.
func (s *PeriodService) UnclosePeriod(ctx context.Context, tenantID uuid.UUID, periodNo int) error {
	periods, err := s.periodRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("list periods: %w", err)
	}

	var found bool
	for _, p := range periods {
		if p.PeriodNo == periodNo {
			found = true
			if p.Status != "closed" {
				return errors.New("period is not closed, cannot unclose")
			}
			break
		}
	}
	if !found {
		return errors.New("period not found")
	}

	if err := s.periodRepo.UpdateStatus(ctx, tenantID, periodNo, "open", uuid.Nil); err != nil {
		return fmt.Errorf("update period status: %w", err)
	}
	return nil
}

// generateClosingEntries creates closing entries for income and expense accounts.
// It debits all income accounts and credits all expense accounts to a retained earnings account.
func (s *PeriodService) generateClosingEntries(ctx context.Context, tenantID uuid.UUID, period model.AccountingPeriod, userID, userName string) error {
	// Find income and expense accounts
	incomeAccts, err := s.accountRepo.ListByType(ctx, tenantID, "income")
	if err != nil {
		return fmt.Errorf("list income accounts: %w", err)
	}
	expenseAccts, err := s.accountRepo.ListByType(ctx, tenantID, "expense")
	if err != nil {
		return fmt.Errorf("list expense accounts: %w", err)
	}

	// Get GL entries for the period
	startDate := period.StartDate
	endDate := period.EndDate

	// Sum income
	var totalIncome decimal.Decimal
	var incomeLines []model.JournalEntryLine
	for _, acct := range incomeAccts {
		entries, err := s.glEntryRepo.GetByAccountAndPeriod(ctx, tenantID, acct.ID, startDate, endDate)
		if err != nil {
			return fmt.Errorf("get income gl entries for %s: %w", acct.Code, err)
		}
		for _, e := range entries {
			if !e.IsCancelled {
				totalIncome = totalIncome.Add(e.Credit.Sub(e.Debit))
				incomeLines = append(incomeLines, model.JournalEntryLine{
					AccountID: acct.ID,
					Credit:    e.Credit.Sub(e.Debit),
					Debit:     decimal.Zero,
				})
			}
		}
	}

	// Sum expenses
	var totalExpense decimal.Decimal
	var expenseLines []model.JournalEntryLine
	for _, acct := range expenseAccts {
		entries, err := s.glEntryRepo.GetByAccountAndPeriod(ctx, tenantID, acct.ID, startDate, endDate)
		if err != nil {
			return fmt.Errorf("get expense gl entries for %s: %w", acct.Code, err)
		}
		for _, e := range entries {
			if !e.IsCancelled {
				totalExpense = totalExpense.Add(e.Debit.Sub(e.Credit))
				expenseLines = append(expenseLines, model.JournalEntryLine{
					AccountID: acct.ID,
					Debit:     e.Debit.Sub(e.Credit),
					Credit:    decimal.Zero,
				})
			}
		}
	}

	// Find retained earnings account
	retainedEarningsAcct, err := s.accountRepo.GetByCode(ctx, tenantID, "3101") // 3101 is typical retained earnings code
	if err != nil {
		return errors.New("retained earnings account (3101) not found, please set up account 3101")
	}

	// Create closing voucher
	voucherNo := fmt.Sprintf("CLOSE-%d-%s", period.PeriodNo, time.Now().Format("20060102"))
	netIncome := totalIncome.Sub(totalExpense)
	// netIncome positive = profit (credit retained earnings), negative = loss (debit retained earnings)

	je := &model.JournalEntry{
		ID:          uuid.New(),
		VoucherNo:   voucherNo,
		VoucherType: stringPtr("closing"),
		PostingDate: period.EndDate,
		CompanyID:   uuid.Nil, // will be set from context or left nil
		DocStatus:   1,        // posted immediately
		CreatedBy:   uuid.MustParse(userID),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	created, err := s.journalRepo.Create(ctx, tenantID, je)
	if err != nil {
		return fmt.Errorf("create closing journal entry: %w", err)
	}

	// Build lines: close income accounts (debit income, credit retained earnings)
	closeLines := make([]model.JournalEntryLine, 0)
	for _, line := range incomeLines {
		if line.Credit.GreaterThan(decimal.Zero) {
			closeLines = append(closeLines, model.JournalEntryLine{
				ID:             uuid.New(),
				JournalEntryID: created.ID,
				AccountID:      line.AccountID,
				Debit:          line.Credit,
				Credit:         decimal.Zero,
				DebitCcy:       line.Credit,
				CreditCcy:      decimal.Zero,
				ExchangeRate:   decimal.NewFromInt(1),
				TenantID:       tenantID,
			})
		}
	}

	// Close expense accounts (credit expense, debit retained earnings)
	for _, line := range expenseLines {
		if line.Debit.GreaterThan(decimal.Zero) {
			closeLines = append(closeLines, model.JournalEntryLine{
				ID:             uuid.New(),
				JournalEntryID: created.ID,
				AccountID:      line.AccountID,
				Debit:          decimal.Zero,
				Credit:         line.Debit,
				DebitCcy:       decimal.Zero,
				CreditCcy:      line.Debit,
				ExchangeRate:   decimal.NewFromInt(1),
				TenantID:       tenantID,
			})
		}
	}

	// Retained earnings line
	if netIncome.GreaterThan(decimal.Zero) {
		// Profit: credit retained earnings
		closeLines = append(closeLines, model.JournalEntryLine{
			ID:             uuid.New(),
			JournalEntryID: created.ID,
			AccountID:      retainedEarningsAcct.ID,
			Debit:          decimal.Zero,
			Credit:         netIncome,
			DebitCcy:       decimal.Zero,
			CreditCcy:      netIncome,
			ExchangeRate:   decimal.NewFromInt(1),
			TenantID:       tenantID,
		})
	} else if netIncome.LessThan(decimal.Zero) {
		// Loss: debit retained earnings
		closeLines = append(closeLines, model.JournalEntryLine{
			ID:             uuid.New(),
			JournalEntryID: created.ID,
			AccountID:      retainedEarningsAcct.ID,
			Debit:          netIncome.Abs(),
			Credit:         decimal.Zero,
			DebitCcy:       netIncome.Abs(),
			CreditCcy:      decimal.Zero,
			ExchangeRate:   decimal.NewFromInt(1),
			TenantID:       tenantID,
		})
	}

	_, err = s.journalRepo.AddLines(ctx, tenantID, created.ID, closeLines)
	if err != nil {
		return fmt.Errorf("add closing lines: %w", err)
	}

	// Write GL entries for the closing voucher
	voucherType := "closing"
	if err := s.glEntryRepo.WriteGLEntries(ctx, tenantID, created.ID, closeLines, period.EndDate, &voucherType, created.CompanyID); err != nil {
		return fmt.Errorf("write closing GL entries: %w", err)
	}

	return nil
}

// ClosingPreview represents a preview of closing entries.
type ClosingPreview struct {
	TotalIncome  decimal.Decimal `json:"total_income"`
	TotalExpense decimal.Decimal `json:"total_expense"`
	NetIncome    decimal.Decimal `json:"net_income"`
	VoucherNo    string          `json:"voucher_no,omitempty"`
	EntryCount   int             `json:"entry_count,omitempty"`
}

// PreviewClosing calculates income/expense totals for the closing preview without writing entries.
func (s *PeriodService) PreviewClosing(ctx context.Context, tenantID uuid.UUID, periodNo int) (*ClosingPreview, error) {
	// Get period to determine date range
	periods, err := s.periodRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list periods: %w", err)
	}
	var period *model.AccountingPeriod
	for _, p := range periods {
		if p.PeriodNo == periodNo {
			period = &p
			break
		}
	}
	if period == nil {
		return nil, errors.New("period not found")
	}

	startDate := period.StartDate
	endDate := period.EndDate

	// Find income and expense accounts
	incomeAccts, err := s.accountRepo.ListByType(ctx, tenantID, "income")
	if err != nil {
		return nil, fmt.Errorf("list income accounts: %w", err)
	}
	expenseAccts, err := s.accountRepo.ListByType(ctx, tenantID, "expense")
	if err != nil {
		return nil, fmt.Errorf("list expense accounts: %w", err)
	}

	// Sum income
	var totalIncome decimal.Decimal
	incomeEntryCount := 0
	for _, acct := range incomeAccts {
		entries, err := s.glEntryRepo.GetByAccountAndPeriod(ctx, tenantID, acct.ID, startDate, endDate)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsCancelled {
				totalIncome = totalIncome.Add(e.Credit.Sub(e.Debit))
				if !e.Credit.Sub(e.Debit).IsZero() {
					incomeEntryCount++
				}
			}
		}
	}

	// Sum expenses
	var totalExpense decimal.Decimal
	expenseEntryCount := 0
	for _, acct := range expenseAccts {
		entries, err := s.glEntryRepo.GetByAccountAndPeriod(ctx, tenantID, acct.ID, startDate, endDate)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsCancelled {
				totalExpense = totalExpense.Add(e.Debit.Sub(e.Credit))
				if !e.Debit.Sub(e.Credit).IsZero() {
					expenseEntryCount++
				}
			}
		}
	}

	voucherNo := fmt.Sprintf("CLOSE-%d", periodNo)
	return &ClosingPreview{
		TotalIncome:  totalIncome,
		TotalExpense: totalExpense,
		NetIncome:    totalIncome.Sub(totalExpense),
		VoucherNo:    voucherNo,
		EntryCount:   incomeEntryCount + expenseEntryCount + 1, // +1 for retained earnings
	}, nil
}

// ExecuteClosing generates closing entries for the period without closing the period.
// Returns the closing journal entry voucher info.
func (s *PeriodService) ExecuteClosing(ctx context.Context, tenantID uuid.UUID, periodNo int, userID, userName string) (*ClosingPreview, error) {
	// Get period
	periods, err := s.periodRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list periods: %w", err)
	}
	var period *model.AccountingPeriod
	for _, p := range periods {
		if p.PeriodNo == periodNo {
			period = &p
			break
		}
	}
	if period == nil {
		return nil, errors.New("period not found")
	}

	// Generate closing entries using the existing private method
	if err := s.generateClosingEntries(ctx, tenantID, *period, userID, userName); err != nil {
		return nil, fmt.Errorf("generate closing entries: %w", err)
	}

	// Build preview for response
	incomeAccts, _ := s.accountRepo.ListByType(ctx, tenantID, "income")
	expenseAccts, _ := s.accountRepo.ListByType(ctx, tenantID, "expense")

	startDate, endDate := period.StartDate, period.EndDate

	var totalIncome, totalExpense decimal.Decimal
	for _, acct := range incomeAccts {
		entries, _ := s.glEntryRepo.GetByAccountAndPeriod(ctx, tenantID, acct.ID, startDate, endDate)
		for _, e := range entries {
			if !e.IsCancelled {
				totalIncome = totalIncome.Add(e.Credit.Sub(e.Debit))
			}
		}
	}
	for _, acct := range expenseAccts {
		entries, _ := s.glEntryRepo.GetByAccountAndPeriod(ctx, tenantID, acct.ID, startDate, endDate)
		for _, e := range entries {
			if !e.IsCancelled {
				totalExpense = totalExpense.Add(e.Debit.Sub(e.Credit))
			}
		}
	}

	voucherNo := fmt.Sprintf("CLOSE-%d", periodNo)
	return &ClosingPreview{
		TotalIncome:  totalIncome,
		TotalExpense: totalExpense,
		NetIncome:    totalIncome.Sub(totalExpense),
		VoucherNo:    voucherNo,
	}, nil
}

// ListPeriods returns all accounting periods.
func (s *PeriodService) ListPeriods(ctx context.Context, tenantID uuid.UUID) ([]model.AccountingPeriod, error) {
	return s.periodRepo.ListByTenant(ctx, tenantID)
}

// GetCurrentPeriod returns the current open period.
func (s *PeriodService) GetCurrentPeriod(ctx context.Context, tenantID uuid.UUID) (*model.AccountingPeriod, error) {
	periods, err := s.periodRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	for _, p := range periods {
		if p.Status == "open" {
			return &p, nil
		}
	}
	return nil, errors.New("no open period found")
}

// VoucherGap represents a gap in voucher numbering.
type VoucherGap struct {
	ExpectedNo    int     `json:"expected_no"`
	IsFilled      bool    `json:"is_filled"`
	GapType       string  `json:"gap_type"`
	FillVoucherID *string `json:"fill_voucher_id,omitempty"`
	Message       string  `json:"message"`
}

// ScanVoucherGaps detects gaps in voucher numbering for a given period.
func (s *PeriodService) ScanVoucherGaps(ctx context.Context, tenantID uuid.UUID, year, month int) ([]VoucherGap, error) {
	periodStr := fmt.Sprintf("%04d-%02d", year, month)
	vouchers, err := s.journalRepo.GetByPeriod(ctx, tenantID, periodStr)
	if err != nil {
		return nil, fmt.Errorf("get vouchers by period: %w", err)
	}

	if len(vouchers) == 0 {
		return []VoucherGap{}, nil
	}

	// Group vouchers by prefix (字头), e.g. "记-35" -> prefix="记", num=35
	type voucherNum struct {
		no     int
		id     string
		isVoid bool
	}
	groups := make(map[string][]voucherNum)

	for _, v := range vouchers {
		word, num := parseVoucherNo(v.VoucherNo)
		if num == 0 {
			continue // skip unparseable numbers
		}
		isVoid := v.DocStatus == 2 // cancelled/voided
		groups[word] = append(groups[word], voucherNum{no: num, id: v.ID.String(), isVoid: isVoid})
	}

	var gaps []VoucherGap

	for prefix, nums := range groups {
		// Sort by number ascending
		for i := 0; i < len(nums); i++ {
			for j := i + 1; j < len(nums); j++ {
				if nums[j].no < nums[i].no {
					nums[i], nums[j] = nums[j], nums[i]
				}
			}
		}

		// Check from 1 to max, find gaps
		if len(nums) == 0 {
			continue
		}
		maxNo := nums[len(nums)-1].no

		existing := make(map[int]voucherNum)
		for _, n := range nums {
			existing[n.no] = n
		}

		for expected := 1; expected <= maxNo; expected++ {
			if vn, ok := existing[expected]; ok {
				if vn.isVoid {
					msg := fmt.Sprintf("第 %d 号凭证已作废", expected)
					if prefix != "" {
						msg = fmt.Sprintf("%s-%d 号凭证已作废", prefix, expected)
					}
					idStr := vn.id
					gaps = append(gaps, VoucherGap{
						ExpectedNo:    expected,
						IsFilled:      true,
						GapType:       "voided",
						FillVoucherID: &idStr,
						Message:       msg,
					})
				}
				continue
			}
			msg := fmt.Sprintf("第 %d 号凭证缺失", expected)
			if prefix != "" {
				msg = fmt.Sprintf("%s-%d 号凭证缺失", prefix, expected)
			}
			gaps = append(gaps, VoucherGap{
				ExpectedNo: expected,
				IsFilled:   false,
				GapType:    "missing",
				Message:    msg,
			})
		}
	}

	return gaps, nil
}

// parseVoucherNo extracts the prefix and numeric part from a voucher number.
// Examples: "记-35" -> ("记", 35), "35" -> ("", 35), "转-12" -> ("转", 12)
func parseVoucherNo(voucherNo string) (string, int) {
	voucherNo = strings.TrimSpace(voucherNo)
	if voucherNo == "" {
		return "", 0
	}

	// Try to split by common delimiters: -, ／, ／, space
	delimiters := []string{"-", "／", " ", "_"}
	for _, d := range delimiters {
		if parts := strings.Split(voucherNo, d); len(parts) == 2 {
			word := strings.TrimSpace(parts[0])
			numStr := strings.TrimSpace(parts[1])
			if num, err := strconv.Atoi(numStr); err == nil && num > 0 {
				return word, num
			}
		}
	}

	// Try pure number
	if num, err := strconv.Atoi(voucherNo); err == nil && num > 0 {
		return "", num
	}

	return "", 0
}

// ─── CloseCheckSummary types ───

// BaseCheckItem represents a single basic check item.
type BaseCheckItem struct {
	ID      string      `json:"id"`
	Name    string      `json:"name"`
	Status  string      `json:"status"` // passed/blocked/warning/pending
	Message string      `json:"message"`
	Action  *ActionInfo `json:"action,omitempty"`
}

// ActionInfo describes an optional action for a check item.
type ActionInfo struct {
	Label string `json:"label"`
	Route string `json:"route,omitempty"`
}

// CloseCheckSummary aggregates all pre-close check data for the HealthCheck UI.
type CloseCheckSummary struct {
	BaseChecks      []BaseCheckItem  `json:"base_checks"`
	RiskWarnings    []RiskWarning    `json:"risk_warnings"`
	KeyIndicators   []KeyIndicator   `json:"key_indicators"`
	PendingAccruals []PendingAccrual `json:"pending_accruals"`
	ProfitLossDone  bool             `json:"profit_loss_done"`
	PeriodStatus    string           `json:"period_status"`
}

// GetCloseCheckSummary returns a complete summary including all 7 base checks.
func (s *PeriodService) GetCloseCheckSummary(ctx context.Context, tenantID uuid.UUID, year, month int) (*CloseCheckSummary, error) {
	// Build the period string and date range
	periodStr := fmt.Sprintf("%04d-%02d", year, month)
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1)

	result := &CloseCheckSummary{
		BaseChecks:      make([]BaseCheckItem, 0, 7),
		RiskWarnings:    make([]RiskWarning, 0),
		KeyIndicators:   make([]KeyIndicator, 0),
		PendingAccruals: make([]PendingAccrual, 0),
	}

	// ── Period status ──
	periods, _ := s.periodRepo.ListByTenant(ctx, tenantID)
	for _, p := range periods {
		if p.PeriodName == periodStr || fmt.Sprintf("%d", p.PeriodNo) == periodStr {
			result.PeriodStatus = p.Status
			break
		}
	}
	if result.PeriodStatus == "" {
		result.PeriodStatus = "open"
	}

	// ── b1: All vouchers posted ──
	unposted, _ := s.journalRepo.CountUnpostedByPeriod(ctx, tenantID, periodStr)
	if unposted > 0 {
		result.BaseChecks = append(result.BaseChecks, BaseCheckItem{
			ID:      "b1",
			Name:    "全部凭证已记账",
			Status:  "blocked",
			Message: fmt.Sprintf("%d 张未记账", unposted),
			Action:  &ActionInfo{Label: "去审核 →", Route: "/vouchers/review"},
		})
	} else {
		result.BaseChecks = append(result.BaseChecks, BaseCheckItem{
			ID:      "b1",
			Name:    "全部凭证已记账",
			Status:  "passed",
			Message: "已全部记账",
		})
	}

	// ── b2: Balance sheet balance ──
	allBalances, _ := s.glEntryRepo.GetBalancesByPeriod(ctx, tenantID, startDate, endDate)
	var totalDebit, totalCredit decimal.Decimal
	for _, b := range allBalances {
		totalDebit = totalDebit.Add(b.Debit)
		totalCredit = totalCredit.Add(b.Credit)
	}
	if totalDebit.Equal(totalCredit) {
		result.BaseChecks = append(result.BaseChecks, BaseCheckItem{
			ID:      "b2",
			Name:    "资产负债表平衡",
			Status:  "passed",
			Message: "差额: 0.00",
		})
	} else {
		diff, _ := totalDebit.Sub(totalCredit).Float64()
		result.BaseChecks = append(result.BaseChecks, BaseCheckItem{
			ID:      "b2",
			Name:    "资产负债表平衡",
			Status:  "blocked",
			Message: fmt.Sprintf("试算不平衡，差额 %.2f", diff),
			Action:  &ActionInfo{Label: "查看试算表", Route: "/period/reports"},
		})
	}

	// ── b3: Profit/loss closing done ──
	closingCount, _ := s.journalRepo.CountClosingByPeriod(ctx, tenantID, periodStr)
	result.ProfitLossDone = closingCount > 0
	if result.ProfitLossDone {
		result.BaseChecks = append(result.BaseChecks, BaseCheckItem{
			ID:      "b3",
			Name:    "损益已结转",
			Status:  "passed",
			Message: "已结转",
		})
	} else {
		result.BaseChecks = append(result.BaseChecks, BaseCheckItem{
			ID:      "b3",
			Name:    "损益已结转",
			Status:  "blocked",
			Message: "损益类科目尚未结转",
			Action:  &ActionInfo{Label: "结转损益"},
		})
	}

	// ── b4: Depreciation accrued ──
	periodNo := year*100 + month + month // intentional to use periodNo as computed
	_ = periodNo
	unpostedDepr, _ := s.depreciationRepo.GetUnpostedSchedulesByPeriod(ctx, tenantID, year*100+month)
	if len(unpostedDepr) > 0 {
		result.BaseChecks = append(result.BaseChecks, BaseCheckItem{
			ID:      "b4",
			Name:    "折旧已计提",
			Status:  "blocked",
			Message: fmt.Sprintf("存在 %d 项使用中资产未计提本月折旧", len(unpostedDepr)),
			Action:  &ActionInfo{Label: "处理折旧 →", Route: "/period/depreciation"},
		})
		result.PendingAccruals = append(result.PendingAccruals, PendingAccrual{
			Type:    "depreciation",
			Item:    "本月固定资产折旧",
			Missing: true,
			Details: fmt.Sprintf("存在 %d 项使用中资产未计提本月折旧", len(unpostedDepr)),
		})
	} else {
		result.BaseChecks = append(result.BaseChecks, BaseCheckItem{
			ID:      "b4",
			Name:    "折旧已计提",
			Status:  "passed",
			Message: "已计提",
		})
		result.PendingAccruals = append(result.PendingAccruals, PendingAccrual{
			Type:    "depreciation",
			Item:    "本月固定资产折旧",
			Missing: false,
		})
	}

	// ── b5: Voucher numbering continuity ──
	gaps, _ := s.ScanVoucherGaps(ctx, tenantID, year, month)
	hasMissing := false
	for _, g := range gaps {
		if g.GapType == "missing" {
			hasMissing = true
			break
		}
	}
	if hasMissing {
		missingCount := 0
		for _, g := range gaps {
			if g.GapType == "missing" {
				missingCount++
			}
		}
		result.BaseChecks = append(result.BaseChecks, BaseCheckItem{
			ID:      "b5",
			Name:    "凭证编号连续性",
			Status:  "blocked",
			Message: fmt.Sprintf("存在 %d 个断号", missingCount),
			Action:  &ActionInfo{Label: "查看断号", Route: "/period/voucher-gaps"},
		})
	} else {
		result.BaseChecks = append(result.BaseChecks, BaseCheckItem{
			ID:      "b5",
			Name:    "凭证编号连续性",
			Status:  "passed",
			Message: "编号连续",
		})
	}

	// ── b6: Bank statement vs GL consistency ──
	// Get all bank accounts that have transactions in the period
	bankAcctIDs, _ := s.bankTransactionRepo.GetAllBankAccountIDs(ctx, tenantID)
	bankInconsistent := false
	var bankMsg string
	for _, bankAcctID := range bankAcctIDs {
		bankBalance, err := s.bankTransactionRepo.GetBankStatementBalance(ctx, tenantID, bankAcctID, startDate, endDate)
		if err != nil {
			continue
		}
		glBalance, err := s.glEntryRepo.GetBankGLBalance(ctx, tenantID, startDate, endDate)
		if err != nil {
			continue
		}
		// Compare: check if bank GL balance (across all bank accounts) matches total bank statement balance
		// We aggregate across all bank accounts
		_ = bankAcctID
		_ = bankBalance
		_ = glBalance
		// For simplicity, compare aggregated bank statement total vs GL total for bank accounts
		if !bankBalance.Equal(glBalance) && !bankBalance.IsZero() && !glBalance.IsZero() {
			bankInconsistent = true
			diff, _ := bankBalance.Sub(glBalance).Abs().Float64()
			bankMsg = fmt.Sprintf("银行日记账与GL不一致，差额 %.2f", diff)
		}
	}
	if bankInconsistent {
		result.BaseChecks = append(result.BaseChecks, BaseCheckItem{
			ID:      "b6",
			Name:    "银行日记账一致性",
			Status:  "warning",
			Message: bankMsg,
			Action:  &ActionInfo{Label: "查看 →", Route: "/reconciliation-bank/balance"},
		})
	} else {
		result.BaseChecks = append(result.BaseChecks, BaseCheckItem{
			ID:      "b6",
			Name:    "银行日记账一致性",
			Status:  "passed",
			Message: "全部一致",
		})
	}

	// ── b7: AR/AP reconciliation completion ──
	// Count unmatched bank transactions older than 30 days (as proxy for AR/AP overdue)
	oldUnmatched, err := s.bankTransactionRepo.GetUnreconciledOldCount(ctx, tenantID, 30)
	if err != nil {
		oldUnmatched = 0
	}
	// Also check invoices with outstanding amounts > 30 days overdue
	var overdueInvoices int
	cutoffDate := time.Now().AddDate(0, 0, -30)
	invoices, err := s.invoiceRepo.ListByTenant(ctx, tenantID, model.InvoiceFilter{})
	if err == nil {
		for _, inv := range invoices {
			if inv.OutstandingAmount.IsPositive() && inv.DueDate != nil && inv.DueDate.Before(cutoffDate) {
				overdueInvoices++
			}
		}
	}

	totalUnreconciled := oldUnmatched + overdueInvoices
	if totalUnreconciled > 0 {
		result.BaseChecks = append(result.BaseChecks, BaseCheckItem{
			ID:      "b7",
			Name:    "往来核销完成度",
			Status:  "warning",
			Message: fmt.Sprintf("%d 笔超 30 天未核销", totalUnreconciled),
			Action:  &ActionInfo{Label: "查看 →", Route: "/reconciliation/manual"},
		})
	} else {
		result.BaseChecks = append(result.BaseChecks, BaseCheckItem{
			ID:      "b7",
			Name:    "往来核销完成度",
			Status:  "passed",
			Message: "无超期未核销项",
		})
	}

	// ── Risk warnings (reclassification) ──
	if allBalances != nil {
		for _, b := range allBalances {
			if b.Code == "1122" && b.Credit.GreaterThan(b.Debit) && !b.Debit.IsZero() {
				netBalance, _ := b.Debit.Sub(b.Credit).Float64()
				result.RiskWarnings = append(result.RiskWarnings, RiskWarning{
					Type: "reclassification", Severity: "warning",
					SubjectCode: b.Code, SubjectName: b.Name,
					Balance: netBalance,
					Message: fmt.Sprintf("%s 为贷方余额 %.2f 元，建议重分类至预收账款", b.Name, -netBalance),
				})
			}
			if b.Code == "2202" && b.Debit.GreaterThan(b.Credit) && !b.Credit.IsZero() {
				netBalance, _ := b.Debit.Sub(b.Credit).Float64()
				result.RiskWarnings = append(result.RiskWarnings, RiskWarning{
					Type: "reclassification", Severity: "warning",
					SubjectCode: b.Code, SubjectName: b.Name,
					Balance: netBalance,
					Message: fmt.Sprintf("%s 为借方余额 %.2f 元，建议重分类至预付账款", b.Name, netBalance),
				})
			}
		}
	}

	// ── Key indicators ──
	prevStart := startDate.AddDate(0, -1, 0)
	prevEnd := endDate.AddDate(0, -1, 0)
	result.KeyIndicators = computeKeyIndicators(ctx, s.glEntryRepo, tenantID, startDate, endDate, prevStart, prevEnd)

	return result, nil
}

// ─── PreCloseCheck types ───

type RiskWarning struct {
	Type        string  `json:"type"`
	Severity    string  `json:"severity"`
	SubjectCode string  `json:"subject_code"`
	SubjectName string  `json:"subject_name"`
	Balance     float64 `json:"balance"`
	Message     string  `json:"message"`
}

type KeyIndicator struct {
	Name         string   `json:"name"`
	CurrentValue *float64 `json:"current_value"`
	LastValue    *float64 `json:"last_value"`
	Unit         string   `json:"unit"`
	Alert        bool     `json:"alert"`
	Message      string   `json:"message"`
}

type PendingAccrual struct {
	Type    string `json:"type"`
	Item    string `json:"item"`
	Missing bool   `json:"missing"`
	Details string `json:"details,omitempty"`
}

type PreCloseCheckResult struct {
	PeriodStatus     string           `json:"period_status"`
	UnpostedVouchers int              `json:"unposted_vouchers"`
	ReportBalanceOK  bool             `json:"report_balance_ok"`
	ProfitLossDone   bool             `json:"profit_loss_done"`
	RiskWarnings     []RiskWarning    `json:"risk_warnings"`
	KeyIndicators    []KeyIndicator   `json:"key_indicators"`
	PendingAccruals  []PendingAccrual `json:"pending_accruals"`
}

// PreCloseCheck runs all pre-close checks for a given period.
func (s *PeriodService) PreCloseCheck(ctx context.Context, tenantID uuid.UUID, year, month int) (*PreCloseCheckResult, error) {
	periodStr := fmt.Sprintf("%04d-%02d", year, month)
	result := &PreCloseCheckResult{
		RiskWarnings:    make([]RiskWarning, 0),
		KeyIndicators:   make([]KeyIndicator, 0),
		PendingAccruals: make([]PendingAccrual, 0),
	}

	// ── 1. Period status ──
	periods, err := s.periodRepo.ListByTenant(ctx, tenantID)
	if err == nil {
		for _, p := range periods {
			if p.PeriodName == periodStr || fmt.Sprintf("%d", p.PeriodNo) == periodStr {
				result.PeriodStatus = p.Status
				break
			}
		}
	}
	if result.PeriodStatus == "" {
		result.PeriodStatus = "open"
	}

	// ── 2. Unposted vouchers ──
	unposted, _ := s.journalRepo.CountUnpostedByPeriod(ctx, tenantID, periodStr)
	result.UnpostedVouchers = unposted

	// ── 3. Report balance (trial balance) ──
	// Find period date range
	var startDate, endDate time.Time
	for _, p := range periods {
		if p.PeriodName == periodStr || fmt.Sprintf("%d", p.PeriodNo) == periodStr {
			startDate = p.StartDate
			endDate = p.EndDate
			break
		}
	}
	if startDate.IsZero() {
		// Default to month boundaries
		startDate = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		endDate = startDate.AddDate(0, 1, -1)
	}

	result.ReportBalanceOK = true
	_, _, err = s.glEntryRepo.GetIncomeExpenseSummary(ctx, tenantID, startDate, endDate)
	if err == nil {
		// Balance is OK if Income = Expense (both sides match). We also check all-accounts.
		allBalances, balErr := s.glEntryRepo.GetBalancesByPeriod(ctx, tenantID, startDate, endDate)
		if balErr == nil {
			var totalDebit, totalCredit decimal.Decimal
			for _, b := range allBalances {
				totalDebit = totalDebit.Add(b.Debit)
				totalCredit = totalCredit.Add(b.Credit)
			}
			result.ReportBalanceOK = totalDebit.Equal(totalCredit)
		}
	}

	// ── 4. Profit/loss closing done ──
	closingCount, _ := s.journalRepo.CountClosingByPeriod(ctx, tenantID, periodStr)
	result.ProfitLossDone = closingCount > 0

	// ── 5. Risk warnings — reclassification detection ──
	allBalances, err := s.glEntryRepo.GetBalancesByPeriod(ctx, tenantID, startDate, endDate)
	if err == nil {
		for _, b := range allBalances {
			// AR (code 1122) with negative balance → credit balance, suggest reclass to prepayments
			if b.Code == "1122" && b.Credit.GreaterThan(b.Debit) && !b.Debit.IsZero() {
				netBalance, _ := b.Debit.Sub(b.Credit).Float64()
				result.RiskWarnings = append(result.RiskWarnings, RiskWarning{
					Type: "reclassification", Severity: "warning",
					SubjectCode: b.Code, SubjectName: b.Name,
					Balance: netBalance,
					Message: fmt.Sprintf("%s 为贷方余额 %.2f 元，建议重分类至预收账款", b.Name, -netBalance),
				})
			}
			// AP (code 2202) with negative balance → debit balance, suggest reclass to prepayments
			if b.Code == "2202" && b.Debit.GreaterThan(b.Credit) && !b.Credit.IsZero() {
				netBalance, _ := b.Debit.Sub(b.Credit).Float64()
				result.RiskWarnings = append(result.RiskWarnings, RiskWarning{
					Type: "reclassification", Severity: "warning",
					SubjectCode: b.Code, SubjectName: b.Name,
					Balance: netBalance,
					Message: fmt.Sprintf("%s 为借方余额 %.2f 元，建议重分类至预付账款", b.Name, netBalance),
				})
			}
		}
	}

	// ── 6. Key indicators (month-over-month) ──
	prevStart := startDate.AddDate(0, -1, 0)
	prevEnd := endDate.AddDate(0, -1, 0)
	result.KeyIndicators = computeKeyIndicators(ctx, s.glEntryRepo, tenantID, startDate, endDate, prevStart, prevEnd)

	// ── 7. Pending accruals ──
	// Depreciation
	periodNo := year*100 + month
	unpostedDepr, err := s.depreciationRepo.GetUnpostedSchedulesByPeriod(ctx, tenantID, periodNo)
	if err == nil {
		missingCount := len(unpostedDepr)
		detail := ""
		if missingCount > 0 {
			detail = fmt.Sprintf("存在 %d 项使用中资产未计提本月折旧", missingCount)
		}
		result.PendingAccruals = append(result.PendingAccruals, PendingAccrual{
			Type: "depreciation", Item: "本月固定资产折旧",
			Missing: missingCount > 0, Details: detail,
		})
	} else {
		result.PendingAccruals = append(result.PendingAccruals, PendingAccrual{
			Type: "depreciation", Item: "本月固定资产折旧", Missing: false,
		})
	}

	return result, nil
}

// computeKeyIndicators calculates month-over-month financial ratios.
func computeKeyIndicators(ctx context.Context, glEntryRepo *repository.GLEntryRepository, tenantID uuid.UUID, curStart, curEnd, prevStart, prevEnd time.Time) []KeyIndicator {
	indicators := make([]KeyIndicator, 0)

	// Current period income/expense
	curIncome, curExpense, err := glEntryRepo.GetIncomeExpenseSummary(ctx, tenantID, curStart, curEnd)
	if err != nil {
		return indicators
	}
	// Previous period income/expense
	prevIncome, prevExpense, err := glEntryRepo.GetIncomeExpenseSummary(ctx, tenantID, prevStart, prevEnd)
	if err != nil {
		return indicators
	}

	// Gross margin
	curRevenue := curIncome
	prevRevenue := prevIncome
	curGrossProfit := curRevenue.Sub(curExpense)
	prevGrossProfit := prevRevenue.Sub(prevExpense)

	var curMargin, prevMargin float64
	var marginAlert bool
	var marginMsg string
	if !curRevenue.IsZero() {
		curMargin, _ = curGrossProfit.Div(curRevenue).Mul(decimal.NewFromInt(100)).Float64()
	}
	if !prevRevenue.IsZero() {
		prevMargin, _ = prevGrossProfit.Div(prevRevenue).Mul(decimal.NewFromInt(100)).Float64()
	}
	marginAlert = curMargin < prevMargin && prevMargin-curMargin > 5
	if marginAlert {
		marginMsg = fmt.Sprintf("毛利率从 %.1f%% 下降至 %.1f%%，降幅超过 5 个百分点", prevMargin, curMargin)
	} else {
		marginMsg = "毛利率基本持平"
	}
	cur := curMargin
	prev := prevMargin
	indicators = append(indicators, KeyIndicator{
		Name: "毛利率", CurrentValue: &cur, LastValue: &prev,
		Unit: "%", Alert: marginAlert, Message: marginMsg,
	})

	// Expense ratio (expense / revenue)
	var curRatio, prevRatio float64
	var ratioAlert bool
	var ratioMsg string
	if !curRevenue.IsZero() {
		curRatio, _ = curExpense.Div(curRevenue).Mul(decimal.NewFromInt(100)).Float64()
	}
	if !prevRevenue.IsZero() {
		prevRatio, _ = prevExpense.Div(prevRevenue).Mul(decimal.NewFromInt(100)).Float64()
	}
	ratioAlert = curRatio > prevRatio && curRatio-prevRatio > 5
	if ratioAlert {
		ratioMsg = fmt.Sprintf("费用率从 %.1f%% 上升至 %.1f%%，增幅超过 5 个百分点", prevRatio, curRatio)
	} else {
		ratioMsg = "费用率基本持平"
	}
	cr := curRatio
	pr := prevRatio
	indicators = append(indicators, KeyIndicator{
		Name: "期间费用率", CurrentValue: &cr, LastValue: &pr,
		Unit: "%", Alert: ratioAlert, Message: ratioMsg,
	})

	// Operating profit margin
	var curProfit, prevProfit float64
	var profitAlert bool
	var profitMsg string
	curOpProfit := curGrossProfit
	prevOpProfit := prevGrossProfit
	if !curRevenue.IsZero() {
		curProfit, _ = curOpProfit.Div(curRevenue).Mul(decimal.NewFromInt(100)).Float64()
	}
	if !prevRevenue.IsZero() {
		prevProfit, _ = prevOpProfit.Div(prevRevenue).Mul(decimal.NewFromInt(100)).Float64()
	}
	profitAlert = curProfit < prevProfit && prevProfit-curProfit > 3
	if profitAlert {
		profitMsg = fmt.Sprintf("营业利润率从 %.1f%% 下降至 %.1f%%", prevProfit, curProfit)
	} else {
		profitMsg = "营业利润率基本持平"
	}
	cp := curProfit
	pp := prevProfit
	indicators = append(indicators, KeyIndicator{
		Name: "营业利润率", CurrentValue: &cp, LastValue: &pp,
		Unit: "%", Alert: profitAlert, Message: profitMsg,
	})

	return indicators
}

func stringPtr(s string) *string { return &s }

// ─── CreatePeriodRequest / UpdatePeriodRequest ───

// CreatePeriodRequest is the request body for creating a new accounting period.
type CreatePeriodRequest struct {
	PeriodNo   int       `json:"period_no"`
	PeriodName string    `json:"period_name"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
}

// UpdatePeriodRequest is the request body for updating a period's metadata.
type UpdatePeriodRequest struct {
	PeriodName *string    `json:"period_name"`
	StartDate  *time.Time `json:"start_date"`
	EndDate    *time.Time `json:"end_date"`
}

// CreatePeriod creates a new accounting period after validating overlap and period_no uniqueness.
func (s *PeriodService) CreatePeriod(ctx context.Context, tenantID uuid.UUID, req CreatePeriodRequest) (*model.AccountingPeriod, error) {
	if req.PeriodNo <= 0 {
		return nil, errors.New("period_no is required")
	}
	if req.PeriodName == "" {
		return nil, errors.New("period_name is required")
	}
	if req.StartDate.IsZero() || req.EndDate.IsZero() {
		return nil, errors.New("start_date and end_date are required")
	}
	if req.EndDate.Before(req.StartDate) {
		return nil, errors.New("end_date must be after start_date")
	}

	// Check overlap with existing periods
	overlap, err := s.periodRepo.CheckOverlap(ctx, tenantID, req.StartDate, req.EndDate, nil)
	if err != nil {
		return nil, fmt.Errorf("check overlap: %w", err)
	}
	if overlap {
		return nil, errors.New("period date range overlaps with an existing period")
	}

	p := &model.AccountingPeriod{
		PeriodNo:   req.PeriodNo,
		PeriodName: req.PeriodName,
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
		Status:     "open",
	}
	created, err := s.periodRepo.Create(ctx, tenantID, p)
	if err != nil {
		// pgx unique violation (period_no per tenant)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, errors.New("period_no already exists for this tenant")
		}
		return nil, fmt.Errorf("create period: %w", err)
	}
	return created, nil
}

// UpdatePeriod updates a period's metadata after validating constraints.
func (s *PeriodService) UpdatePeriod(ctx context.Context, tenantID uuid.UUID, id uuid.UUID, req UpdatePeriodRequest) error {
	period, err := s.periodRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("period not found")
		}
		return fmt.Errorf("get period: %w", err)
	}

	if req.StartDate != nil || req.EndDate != nil {
		if period.Status == "closed" {
			return errors.New("cannot change dates of a closed period")
		}
		start := req.StartDate
		if start == nil {
			start = &period.StartDate
		}
		end := req.EndDate
		if end == nil {
			end = &period.EndDate
		}
		if end.Before(*start) {
			return errors.New("end_date must be after start_date")
		}
		overlap, err := s.periodRepo.CheckOverlap(ctx, tenantID, *start, *end, &id)
		if err != nil {
			return fmt.Errorf("check overlap: %w", err)
		}
		if overlap {
			return errors.New("period date range overlaps with an existing period")
		}
	}

	if err := s.periodRepo.UpdateMeta(ctx, tenantID, id, req.PeriodName, req.StartDate, req.EndDate); err != nil {
		return fmt.Errorf("update period: %w", err)
	}
	return nil
}

// DeletePeriod deletes a period after verifying it is not closed and has no journal entries.
func (s *PeriodService) DeletePeriod(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error {
	period, err := s.periodRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("period not found")
		}
		return fmt.Errorf("get period: %w", err)
	}
	if period.Status == "closed" {
		return errors.New("cannot delete a closed period")
	}

	// Check for journal entries referencing this period
	periodStr := fmt.Sprintf("%d", period.PeriodNo)
	vouchers, err := s.journalRepo.GetByPeriod(ctx, tenantID, periodStr)
	if err != nil {
		return fmt.Errorf("check journal entries: %w", err)
	}
	if len(vouchers) > 0 {
		return errors.New("cannot delete period with existing journal entries")
	}

	if err := s.periodRepo.Delete(ctx, tenantID, id); err != nil {
		return fmt.Errorf("delete period: %w", err)
	}
	return nil
}

// EnablePeriod sets a closed period back to open status.
func (s *PeriodService) EnablePeriod(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error {
	period, err := s.periodRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("period not found")
		}
		return fmt.Errorf("get period: %w", err)
	}
	if period.Status != "closed" {
		return errors.New("only closed periods can be enabled")
	}

	if err := s.periodRepo.UpdateStatus(ctx, tenantID, period.PeriodNo, "open", uuid.Nil); err != nil {
		return fmt.Errorf("enable period: %w", err)
	}
	return nil
}
