package service

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
	"huihua/finance/internal/model"
	"huihua/finance/internal/repository"
)

// OpeningBalanceService handles opening balance operations.
type OpeningBalanceService struct {
	obRepo      *repository.OpeningBalanceRepository
	accountRepo *repository.AccountRepository
}

// NewOpeningBalanceService creates a new OpeningBalanceService.
func NewOpeningBalanceService(obRepo *repository.OpeningBalanceRepository, accountRepo *repository.AccountRepository) *OpeningBalanceService {
	return &OpeningBalanceService{
		obRepo:      obRepo,
		accountRepo: accountRepo,
	}
}

// ImportFromExcel parses Excel data and imports opening balances.
func (s *OpeningBalanceService) ImportFromExcel(ctx context.Context, tenantID, companyID uuid.UUID, periodNo int, data []byte) ([]model.OpeningBalanceEntry, error) {
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("open excel: %w", err)
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("get rows: %w", err)
	}

	// Skip header row, process data rows
	var entries []model.OpeningBalanceEntry
	for i, row := range rows {
		if i == 0 { // skip header
			continue
		}
		if len(row) < 4 {
			continue
		}

		accountCode := row[0]
		debitStr := row[2]
		creditStr := row[3]

		// Parse amounts
		debit := decimal.Zero
		credit := decimal.Zero

		if debitStr != "" {
			debit, err = decimal.NewFromString(debitStr)
			if err != nil {
				return nil, fmt.Errorf("row %d: invalid debit amount %s: %w", i+1, debitStr, err)
			}
		}
		if creditStr != "" {
			credit, err = decimal.NewFromString(creditStr)
			if err != nil {
				return nil, fmt.Errorf("row %d: invalid credit amount %s: %w", i+1, creditStr, err)
			}
		}

		// Skip rows with no amounts
		if debit.IsZero() && credit.IsZero() {
			continue
		}

		// Find account by code
		account, err := s.accountRepo.GetByCode(ctx, tenantID, accountCode)
		if err != nil {
			return nil, fmt.Errorf("row %d: account code %s not found: %w", i+1, accountCode, err)
		}

		// Determine balance direction based on root_type
		// For debit-type accounts, debit is the balance direction
		// For credit-type accounts, credit is the balance direction
		entry := model.OpeningBalanceEntry{
			ID:           uuid.New(),
			TenantID:     tenantID,
			CompanyID:    companyID,
			AccountID:    account.ID,
			DebitAmount:  debit,
			CreditAmount: credit,
			PeriodNo:     periodNo,
			IsVerified:   false,
		}

		entries = append(entries, entry)
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("no valid entries found in excel")
	}

	// Batch upsert
	if err := s.obRepo.UpsertBatch(ctx, tenantID, companyID, periodNo, entries); err != nil {
		return nil, fmt.Errorf("upsert batch: %w", err)
	}

	return entries, nil
}

// ValidateOpeningBalance validates that the opening balances are correct.
// Returns validation result with any errors found.
func (s *OpeningBalanceService) ValidateOpeningBalance(ctx context.Context, tenantID uuid.UUID, periodNo int) (*model.OpeningBalanceValidationResult, error) {
	result := &model.OpeningBalanceValidationResult{
		Valid: true,
	}

	// Get all accounts for validation
	accounts, err := s.accountRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}

	// Build account map for lookup
	accountMap := make(map[string]*model.Account)
	for i := range accounts {
		accountMap[accounts[i].Code] = &accounts[i]
	}

	// Get opening balances for the period
	balances, err := s.obRepo.GetByPeriod(ctx, tenantID, periodNo)
	if err != nil {
		return nil, fmt.Errorf("get opening balances: %w", err)
	}

	// Build balance map by account ID
	balanceMap := make(map[uuid.UUID]model.OpeningBalanceEntry)
	for _, b := range balances {
		balanceMap[b.AccountID] = b
	}

	// Calculate totals and validate
	var totalDebit, totalCredit decimal.Decimal

	for _, account := range accounts {
		if account.IsGroup {
			continue // skip group accounts
		}

		bal, ok := balanceMap[account.ID]
		if !ok {
			// Check if this account has any balance
			continue
		}

		totalDebit = totalDebit.Add(bal.DebitAmount)
		totalCredit = totalCredit.Add(bal.CreditAmount)
	}

	result.TotalDebit = totalDebit
	result.TotalCredit = totalCredit
	result.BalanceDiff = totalDebit.Sub(totalCredit)

	// Check if balanced
	if !result.BalanceDiff.IsZero() {
		result.Valid = false
		result.Errors = append(result.Errors, model.OpeningBalanceError{
			AccountCode: "",
			AccountName: "",
			Message:     fmt.Sprintf("Opening balance is not balanced: debit total (%s) != credit total (%s), diff = %s",
				totalDebit.String(), totalCredit.String(), result.BalanceDiff.String()),
		})
	}

	// Check all balance entries have valid account codes
	for _, bal := range balances {
		found := false
		for _, acc := range accounts {
			if acc.ID == bal.AccountID {
				found = true
				break
			}
		}
		if !found {
			result.Valid = false
			result.Errors = append(result.Errors, model.OpeningBalanceError{
				AccountCode: "",
				AccountName: "",
				Message:     fmt.Sprintf("Opening balance references non-existent account ID: %s", bal.AccountID),
			})
		}
	}

	return result, nil
}

// GetTrialBalance returns the trial balance for a given period.
func (s *OpeningBalanceService) GetTrialBalance(ctx context.Context, tenantID uuid.UUID, periodNo int) (*model.TrialBalance, error) {
	// Get all accounts
	accounts, err := s.accountRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}

	// Get opening balances
	balances, err := s.obRepo.GetTrialBalance(ctx, tenantID, periodNo)
	if err != nil {
		return nil, fmt.Errorf("get trial balance: %w", err)
	}

	// Build balance map
	balanceMap := make(map[uuid.UUID]model.OpeningBalanceEntry)
	for _, b := range balances {
		balanceMap[b.AccountID] = b
	}

	// Build trial balance entries
	var entries []model.TrialBalanceEntry
	var totalDebit, totalCredit decimal.Decimal

	for _, account := range accounts {
		if account.IsGroup {
			continue // skip group accounts for trial balance detail
		}

		bal := balanceMap[account.ID]
		accountType := ""
		rootType := ""
		if account.AccountType != nil {
			accountType = *account.AccountType
		}
		if account.RootType != nil {
			rootType = *account.RootType
		}

		entry := model.TrialBalanceEntry{
			AccountCode:   account.Code,
			AccountName:   account.Name,
			AccountType:   accountType,
			RootType:      rootType,
			DebitBalance:  bal.DebitAmount,
			CreditBalance: bal.CreditAmount,
		}
		entries = append(entries, entry)

		totalDebit = totalDebit.Add(bal.DebitAmount)
		totalCredit = totalCredit.Add(bal.CreditAmount)
	}

	return &model.TrialBalance{
		PeriodNo:    periodNo,
		Entries:     entries,
		TotalDebit:  totalDebit,
		TotalCredit: totalCredit,
		IsBalanced:  totalDebit.Equal(totalCredit),
	}, nil
}

// GetByPeriod returns all opening balances for a given period.
func (s *OpeningBalanceService) GetByPeriod(ctx context.Context, tenantID uuid.UUID, periodNo int) ([]model.OpeningBalanceEntry, error) {
	return s.obRepo.GetByPeriod(ctx, tenantID, periodNo)
}

// GetByAccount returns opening balance for a specific account and period.
func (s *OpeningBalanceService) GetByAccount(ctx context.Context, tenantID, accountID uuid.UUID, periodNo int) (*model.OpeningBalanceEntry, error) {
	return s.obRepo.GetByAccount(ctx, tenantID, accountID, periodNo)
}

// Helper function to parse Excel file
func parseExcelStream(data io.Reader) (*excelize.File, error) {
	return excelize.OpenReader(data)
}