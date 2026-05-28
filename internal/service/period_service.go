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

// PeriodService handles accounting period operations including closing.
type PeriodService struct {
	periodRepo    *repository.PeriodRepository
	journalRepo   *repository.JournalRepository
	glEntryRepo   *repository.GLEntryRepository
	accountRepo   *repository.AccountRepository
}

// NewPeriodService creates a new PeriodService.
func NewPeriodService(periodRepo *repository.PeriodRepository, journalRepo *repository.JournalRepository, glEntryRepo *repository.GLEntryRepository, accountRepo *repository.AccountRepository) *PeriodService {
	return &PeriodService{
		periodRepo:    periodRepo,
		journalRepo:   journalRepo,
		glEntryRepo:   glEntryRepo,
		accountRepo:   accountRepo,
	}
}

// ClosePeriodRequest is the request to close an accounting period.
type ClosePeriodRequest struct {
	PeriodNo   int    `json:"period_no"`
	UserID     string `json:"user_id"`
	UserName   string `json:"user_name"`
	GenerateClosingEntries bool `json:"generate_closing_entries"`
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
					Debit:    decimal.Zero,
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
		DocStatus:   1, // posted immediately
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

	// Write GL for the closing entry
	return nil
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

func stringPtr(s string) *string { return &s }