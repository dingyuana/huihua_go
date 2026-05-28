package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/model"
	"huihua/finance/internal/repository"
)

// ReportService generates financial reports (trial balance, P&L, balance sheet).
type ReportService struct {
	glEntryRepo *repository.GLEntryRepository
	obRepo      *repository.OpeningBalanceRepository
	accountRepo *repository.AccountRepository
	periodRepo  *repository.PeriodRepository
}

// NewReportService creates a new ReportService.
func NewReportService(
	glEntryRepo *repository.GLEntryRepository,
	obRepo *repository.OpeningBalanceRepository,
	accountRepo *repository.AccountRepository,
	periodRepo *repository.PeriodRepository,
) *ReportService {
	return &ReportService{
		glEntryRepo:  glEntryRepo,
		obRepo:       obRepo,
		accountRepo:  accountRepo,
		periodRepo:   periodRepo,
	}
}

// GetTrialBalance returns the trial balance for a period.
// It combines: opening balance + current period GL movements = closing balance.
func (s *ReportService) GetTrialBalance(ctx context.Context, tenantID uuid.UUID, periodNo int) (*model.TrialBalance, error) {
	// Get period dates
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
		return nil, fmt.Errorf("period %d not found", periodNo)
	}

	// Get all accounts
	accounts, err := s.accountRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}

	// Get opening balances
	obEntries, err := s.obRepo.GetByPeriod(ctx, tenantID, periodNo)
	if err != nil {
		return nil, fmt.Errorf("get opening balances: %w", err)
	}
	obMap := make(map[uuid.UUID]model.OpeningBalanceEntry)
	for _, e := range obEntries {
		obMap[e.AccountID] = e
	}

	// Get GL entries for the period
	glEntries, err := s.glEntryRepo.GetByTenantInRange(ctx, tenantID, period.StartDate, period.EndDate)
	if err != nil {
		return nil, fmt.Errorf("get gl entries: %w", err)
	}
	glMap := make(map[uuid.UUID][]model.GLEntry)
	for _, e := range glEntries {
		glMap[e.AccountID] = append(glMap[e.AccountID], e)
	}

	// Build trial balance entries
	var entries []model.TrialBalanceEntry
	var totalDebit, totalCredit decimal.Decimal

	for _, acct := range accounts {
		if acct.IsGroup {
			continue // skip group accounts
		}

		ob := obMap[acct.ID]
		acctGL := glMap[acct.ID]

		// Calculate GL movement (sum of non-cancelled entries)
		var periodDebit, periodCredit decimal.Decimal
		for _, e := range acctGL {
			if !e.IsCancelled {
				periodDebit = periodDebit.Add(e.Debit)
				periodCredit = periodCredit.Add(e.Credit)
			}
		}

		// Compute debit/credit balance based on root_type
		rootType := ""
		if acct.RootType != nil {
			rootType = *acct.RootType
		}

		entry := model.TrialBalanceEntry{
			AccountCode:   acct.Code,
			AccountName:   acct.Name,
			AccountType:   stringVal(acct.AccountType),
			RootType:      rootType,
			DebitBalance:  ob.DebitAmount.Add(periodDebit),
			CreditBalance: ob.CreditAmount.Add(periodCredit),
		}

		entries = append(entries, entry)

		totalDebit = totalDebit.Add(entry.DebitBalance)
		totalCredit = totalCredit.Add(entry.CreditBalance)
	}

	return &model.TrialBalance{
		PeriodNo:    periodNo,
		PeriodName:  period.PeriodName,
		StartDate:   period.StartDate,
		EndDate:     period.EndDate,
		Entries:     entries,
		TotalDebit:  totalDebit,
		TotalCredit: totalCredit,
		IsBalanced:  totalDebit.Equal(totalCredit),
	}, nil
}

// GetIncomeStatement returns the income statement (P&L) for a period.
func (s *ReportService) GetIncomeStatement(ctx context.Context, tenantID uuid.UUID, periodNo int) (*model.IncomeStatement, error) {
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
		return nil, fmt.Errorf("period %d not found", periodNo)
	}

	// Get all accounts
	accounts, err := s.accountRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}

	// Get GL entries for the period
	glEntries, err := s.glEntryRepo.GetByTenantInRange(ctx, tenantID, period.StartDate, period.EndDate)
	if err != nil {
		return nil, fmt.Errorf("get gl entries: %w", err)
	}
	glMap := make(map[uuid.UUID][]model.GLEntry)
	for _, e := range glEntries {
		glMap[e.AccountID] = append(glMap[e.AccountID], e)
	}

	// Aggregate by root_type
	type agg struct {
		debit, credit decimal.Decimal
	}
	rootAgg := make(map[string]agg)
	acctDetails := make(map[string][]model.IncomeStatementAccountEntry)

	for _, acct := range accounts {
		if acct.IsGroup {
			continue
		}
		rootType := stringVal(acct.RootType)
		if rootType != "income" && rootType != "expense" {
			continue
		}

		acctGL := glMap[acct.ID]
		var periodDebit, periodCredit decimal.Decimal
		for _, e := range acctGL {
			if !e.IsCancelled {
				periodDebit = periodDebit.Add(e.Debit)
				periodCredit = periodCredit.Add(e.Credit)
			}
		}

		// Income: credit - debit; Expense: debit - credit
		var net decimal.Decimal
		if rootType == "income" {
			net = periodCredit.Sub(periodDebit)
		} else {
			net = periodDebit.Sub(periodCredit)
		}

		if !net.IsZero() {
			rootAgg[rootType] = agg{
				debit:  rootAgg[rootType].debit.Add(periodDebit),
				credit: rootAgg[rootType].credit.Add(periodCredit),
			}
			acctDetails[rootType] = append(acctDetails[rootType], model.IncomeStatementAccountEntry{
				AccountCode:   acct.Code,
				AccountName:   acct.Name,
				PeriodDebit:  periodDebit,
				PeriodCredit: periodCredit,
				NetAmount:    net,
			})
		}
	}

	totalIncome := decimal.Zero
	if v, ok := rootAgg["income"]; ok {
		totalIncome = v.credit.Sub(v.debit)
	}
	totalExpense := decimal.Zero
	if v, ok := rootAgg["expense"]; ok {
		totalExpense = v.debit.Sub(v.credit)
	}
	netProfit := totalIncome.Sub(totalExpense)

	return &model.IncomeStatement{
		PeriodNo:      periodNo,
		PeriodName:    period.PeriodName,
		StartDate:     period.StartDate,
		EndDate:       period.EndDate,
		TotalIncome:   totalIncome,
		TotalExpense:  totalExpense,
		NetProfit:     netProfit,
		IncomeDetails: acctDetails["income"],
		ExpenseDetails: acctDetails["expense"],
	}, nil
}

// GetBalanceSheet returns the balance sheet as of a period end date.
func (s *ReportService) GetBalanceSheet(ctx context.Context, tenantID uuid.UUID, periodNo int) (*model.BalanceSheet, error) {
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
		return nil, fmt.Errorf("period %d not found", periodNo)
	}

	// Get all accounts
	accounts, err := s.accountRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}

	// Get opening balances
	obEntries, err := s.obRepo.GetByPeriod(ctx, tenantID, periodNo)
	if err != nil {
		return nil, fmt.Errorf("get opening balances: %w", err)
	}
	obMap := make(map[uuid.UUID]model.OpeningBalanceEntry)
	for _, e := range obEntries {
		obMap[e.AccountID] = e
	}

	// Get ALL GL entries up to period end (for running balance)
	glEntries, err := s.glEntryRepo.GetByTenantInRange(ctx, tenantID, period.StartDate, period.EndDate)
	if err != nil {
		return nil, fmt.Errorf("get gl entries: %w", err)
	}
	glMap := make(map[uuid.UUID][]model.GLEntry)
	for _, e := range glEntries {
		glMap[e.AccountID] = append(glMap[e.AccountID], e)
	}

	type bsEntry struct {
		account model.Account
		balance decimal.Decimal
	}
	var assetEntries, liabilityEntries, equityEntries []model.BalanceSheetAccountEntry

	for _, acct := range accounts {
		if acct.IsGroup {
			continue
		}
		rootType := stringVal(acct.RootType)
		if rootType != "asset" && rootType != "liability" && rootType != "equity" {
			continue
		}

		ob := obMap[acct.ID]
		acctGL := glMap[acct.ID]

		var periodDebit, periodCredit decimal.Decimal
		for _, e := range acctGL {
			if !e.IsCancelled {
				periodDebit = periodDebit.Add(e.Debit)
				periodCredit = periodCredit.Add(e.Credit)
			}
		}

		// Calculate closing balance based on root_type
		var closingBalance decimal.Decimal
		switch rootType {
		case "asset":
			closingBalance = ob.DebitAmount.Add(periodDebit).Sub(ob.CreditAmount.Add(periodCredit))
		case "liability", "equity":
			closingBalance = ob.CreditAmount.Add(periodCredit).Sub(ob.DebitAmount.Add(periodDebit))
		}

		bsEntry := model.BalanceSheetAccountEntry{
			AccountCode: acct.Code,
			AccountName: acct.Name,
			Balance:     closingBalance,
		}

		switch rootType {
		case "asset":
			assetEntries = append(assetEntries, bsEntry)
		case "liability":
			liabilityEntries = append(liabilityEntries, bsEntry)
		case "equity":
			equityEntries = append(equityEntries, bsEntry)
		}
	}

	totalAssets := decimal.Zero
	for _, e := range assetEntries {
		totalAssets = totalAssets.Add(e.Balance)
	}
	totalLiabilities := decimal.Zero
	for _, e := range liabilityEntries {
		totalLiabilities = totalLiabilities.Add(e.Balance)
	}
	totalEquity := decimal.Zero
	for _, e := range equityEntries {
		totalEquity = totalEquity.Add(e.Balance)
	}

	// Equity should = assets - liabilities (verify)
	derivedEquity := totalAssets.Sub(totalLiabilities)

	return &model.BalanceSheet{
		PeriodNo:         periodNo,
		PeriodName:       period.PeriodName,
		AsOfDate:         period.EndDate,
		TotalAssets:      totalAssets,
		TotalLiabilities: totalLiabilities,
		TotalEquity:      totalEquity,
		DerivedEquity:    derivedEquity,
		IsBalanced:       totalEquity.Equal(derivedEquity),
		AssetEntries:     assetEntries,
		LiabilityEntries: liabilityEntries,
		EquityEntries:    equityEntries,
	}, nil
}

func stringVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}