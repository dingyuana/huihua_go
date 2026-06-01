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

// BankReconciliationService handles bank reconciliation business logic.
type BankReconciliationService struct {
	bankTxnRepo *repository.BankTransactionRepository
	journalRepo *repository.JournalRepository
	bankRepo    *repository.BankRepository
	glEntryRepo *repository.GLEntryRepository
}

// NewBankReconciliationService creates a new BankReconciliationService.
func NewBankReconciliationService(
	bankTxnRepo *repository.BankTransactionRepository,
	journalRepo *repository.JournalRepository,
	bankRepo *repository.BankRepository,
	glEntryRepo *repository.GLEntryRepository,
) *BankReconciliationService {
	return &BankReconciliationService{
		bankTxnRepo: bankTxnRepo,
		journalRepo: journalRepo,
		bankRepo:    bankRepo,
		glEntryRepo: glEntryRepo,
	}
}

// ReconciliationResult holds the result of a reconciliation operation.
type ReconciliationResult struct {
	BankBalance     decimal.Decimal    `json:"bank_balance"`
	BookBalance     decimal.Decimal    `json:"book_balance"`
	AdjustedBalance decimal.Decimal    `json:"adjusted_balance"`
	BankOnlyItems   []UnreconciledItem `json:"bank_only_items"`
	BookOnlyItems   []UnreconciledItem `json:"book_only_items"`
	MatchedCount    int                `json:"matched_count"`
	TotalMatched    decimal.Decimal    `json:"total_matched"`
}

// UnreconciledItem represents an item that appears in one side but not the other.
type UnreconciledItem struct {
	ID          uuid.UUID       `json:"id"`
	SourceType  string          `json:"source_type"`
	TxnDate     time.Time       `json:"txn_date"`
	Description string          `json:"description"`
	Amount      decimal.Decimal `json:"amount"`
	Direction   string          `json:"direction"`
}

// ReconcileBankAccount performs bank reconciliation for a bank account in a specific period.
// It matches bank transactions with journal entries by date + amount.
func (s *BankReconciliationService) ReconcileBankAccount(ctx context.Context, tenantID, bankAccountID uuid.UUID, periodNo int) (*ReconciliationResult, error) {
	// Convert periodNo to start/end dates (YYYYMM -> dates)
	startDate, endDate := periodToDateRange(periodNo)

	// 1. Get all matched bank transactions in the period
	bankTxns, err := s.bankTxnRepo.GetMatchedByPeriod(ctx, tenantID, bankAccountID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("get bank transactions: %w", err)
	}

	// 2. Get all GL entries for the bank account in the period
	// First get bank account to find the clearing account
	bankAccount, err := s.bankRepo.GetByID(ctx, tenantID, bankAccountID)
	if err != nil {
		return nil, fmt.Errorf("get bank account: %w", err)
	}

	glEntries, err := s.glEntryRepo.GetByAccountAndPeriod(ctx, tenantID, *bankAccount.ClearingAccountID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("get GL entries: %w", err)
	}

	// 3. Calculate bank balance (sum of debit - credit for matched transactions)
	bankBalance := decimal.Zero
	for _, txn := range bankTxns {
		bankBalance = bankBalance.Add(txn.Debit).Sub(txn.Credit)
	}

	// 4. Calculate book balance (sum of debit - credit for GL entries)
	bookBalance := decimal.Zero
	for _, entry := range glEntries {
		bookBalance = bookBalance.Add(entry.Debit).Sub(entry.Credit)
	}

	// 5. Match items by date + amount
	// Build lookup maps
	bankMap := make(map[string][]model.BankTransaction) // key: date+amount
	for _, txn := range bankTxns {
		amt := txn.Debit.Sub(txn.Credit).Abs()
		if amt.IsZero() {
			continue
		}
		key := txn.TxnDate.Format("2006-01-02") + "|" + amt.String()
		bankMap[key] = append(bankMap[key], txn)
	}

	glMap := make(map[string][]model.GLEntry) // key: date+amount
	for _, entry := range glEntries {
		amt := entry.Debit.Sub(entry.Credit).Abs()
		if amt.IsZero() {
			continue
		}
		key := entry.PostingDate.Format("2006-01-02") + "|" + amt.String()
		glMap[key] = append(glMap[key], entry)
	}

	// Track matched items
	matchedBank := make(map[uuid.UUID]bool)
	matchedGL := make(map[uuid.UUID]bool)

	var bankOnlyItems []UnreconciledItem
	var bookOnlyItems []UnreconciledItem
	var totalMatched decimal.Decimal

	// Find bank items that don't have matching GL entries (bank has, book doesn't)
	for key, txns := range bankMap {
		if _, ok := glMap[key]; !ok {
			// Bank only item
			for _, txn := range txns {
				amt := txn.Debit.Sub(txn.Credit).Abs()
				dir := "debit"
				if txn.Credit.GreaterThan(txn.Debit) {
					dir = "credit"
				}
				desc := ""
				if txn.Description != nil {
					desc = *txn.Description
				}
				bankOnlyItems = append(bankOnlyItems, UnreconciledItem{
					ID:          txn.ID,
					SourceType:  "bank_transaction",
					TxnDate:     txn.TxnDate,
					Description: desc,
					Amount:      amt,
					Direction:   dir,
				})
				matchedBank[txn.ID] = true
			}
		}
	}

	// Find GL items that don't have matching bank transactions (book has, bank doesn't)
	for key, entries := range glMap {
		if _, ok := bankMap[key]; !ok {
			// Book only item
			for _, entry := range entries {
				amt := entry.Debit.Sub(entry.Credit).Abs()
				dir := "debit"
				if entry.Credit.GreaterThan(entry.Debit) {
					dir = "credit"
				}
				bookOnlyItems = append(bookOnlyItems, UnreconciledItem{
					ID:          entry.ID,
					SourceType:  "gl_entry",
					TxnDate:     entry.PostingDate,
					Description: "",
					Amount:      amt,
					Direction:   dir,
				})
				matchedGL[entry.ID] = true
			}
		}
	}

	// Calculate matched amounts
	matchedCount := 0
	for key, glEntriesList := range glMap {
		if bankTxnsList, ok := bankMap[key]; ok {
			minLen := len(glEntriesList)
			if len(bankTxnsList) < minLen {
				minLen = len(bankTxnsList)
			}
			for i := 0; i < minLen; i++ {
				amt := glEntriesList[i].Debit.Sub(glEntriesList[i].Credit).Abs()
				totalMatched = totalMatched.Add(amt)
				matchedCount++
			}
		}
	}

	// Adjusted balance = bank balance - bank only items + book only items
	adjustedBalance := bankBalance
	for _, item := range bankOnlyItems {
		if item.Direction == "credit" {
			adjustedBalance = adjustedBalance.Add(item.Amount)
		} else {
			adjustedBalance = adjustedBalance.Sub(item.Amount)
		}
	}
	for _, item := range bookOnlyItems {
		if item.Direction == "debit" {
			adjustedBalance = adjustedBalance.Add(item.Amount)
		} else {
			adjustedBalance = adjustedBalance.Sub(item.Amount)
		}
	}

	return &ReconciliationResult{
		BankBalance:     bankBalance,
		BookBalance:     bookBalance,
		AdjustedBalance: adjustedBalance,
		BankOnlyItems:   bankOnlyItems,
		BookOnlyItems:   bookOnlyItems,
		MatchedCount:    matchedCount,
		TotalMatched:    totalMatched,
	}, nil
}

// GetReconciliationReport retrieves the reconciliation report for a bank account in a period.
func (s *BankReconciliationService) GetReconciliationReport(ctx context.Context, tenantID, bankAccountID uuid.UUID, periodNo int) (*model.ReconciliationReport, error) {
	// Get reconciliation record
	record, err := s.getReconciliationRecord(ctx, tenantID, bankAccountID, periodNo)
	if err != nil {
		return nil, err
	}

	report := &model.ReconciliationReport{
		ID:               record.ID,
		BankAccountID:    bankAccountID,
		PeriodNo:         periodNo,
		BankBalance:      record.BankBalance,
		BookBalance:      record.BookBalance,
		AdjustedBalance:  record.AdjustedBalance,
		Status:           record.Status,
		ReconciledBy:     record.ReconciledBy,
		ReconciledAt:     record.ReconciledAt,
	}

	// Get unreconciled items
	unreconItems, err := s.getUnreconciledItems(ctx, record.ID)
	if err != nil {
		return nil, fmt.Errorf("get unreconciled items: %w", err)
	}

	// Split into bank_only and book_only items
	var bankOnlyItems, bookOnlyItems []model.UnreconciledItem
	for _, item := range unreconItems {
		if item.ItemType == "bank_only" {
			bankOnlyItems = append(bankOnlyItems, item)
		} else {
			bookOnlyItems = append(bookOnlyItems, item)
		}
	}

	report.BankOnlyItems = bankOnlyItems
	report.BookOnlyItems = bookOnlyItems

	return report, nil
}

// MarkAsReconciled marks a reconciliation as completed.
func (s *BankReconciliationService) MarkAsReconciled(ctx context.Context, tenantID, bankAccountID uuid.UUID, periodNo int, userID uuid.UUID) error {
	record, err := s.getReconciliationRecord(ctx, tenantID, bankAccountID, periodNo)
	if err != nil {
		return fmt.Errorf("get reconciliation record: %w", err)
	}

	record.Status = "reconciled"
	record.ReconciledBy = &userID
	now := time.Now()
	record.ReconciledAt = &now

	return s.updateReconciliationRecord(ctx, record)
}

// GetReconciliationStatus returns whether a bank account has been reconciled for a given period.
func (s *BankReconciliationService) GetReconciliationStatus(ctx context.Context, tenantID, bankAccountID uuid.UUID, periodNo int) (string, error) {
	record, err := s.getReconciliationRecord(ctx, tenantID, bankAccountID, periodNo)
	if err != nil {
		// If no record found, return "not_started"
		return "not_started", nil
	}
	return record.Status, nil
}

// Helper functions

func periodToDateRange(periodNo int) (time.Time, time.Time) {
	year := periodNo / 100
	month := periodNo % 100
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1)
	return startDate, endDate
}

func (s *BankReconciliationService) getReconciliationRecord(ctx context.Context, tenantID, bankAccountID uuid.UUID, periodNo int) (*model.ReconciliationRecord, error) {
	// Query for existing record
	query := `
		SELECT id, tenant_id, bank_account_id, period_no, bank_balance, book_balance,
		       adjusted_balance, bank_only_total, book_only_total, status,
		       reconciled_by, reconciled_at, created_at, updated_at
		FROM reconciliation_records
		WHERE tenant_id = $1 AND bank_account_id = $2 AND period_no = $3`

	record := &model.ReconciliationRecord{}
	err := s.bankTxnRepo.GetPool().QueryRow(ctx, query, tenantID, bankAccountID, periodNo).Scan(
		&record.ID, &record.TenantID, &record.BankAccountID, &record.PeriodNo,
		&record.BankBalance, &record.BookBalance, &record.AdjustedBalance,
		&record.BankOnlyTotal, &record.BookOnlyTotal, &record.Status,
		&record.ReconciledBy, &record.ReconciledAt, &record.CreatedAt, &record.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get reconciliation record: %w", err)
	}
	return record, nil
}

func (s *BankReconciliationService) updateReconciliationRecord(ctx context.Context, record *model.ReconciliationRecord) error {
	query := `
		UPDATE reconciliation_records
		SET bank_balance = $4, book_balance = $5, adjusted_balance = $6,
		    bank_only_total = $7, book_only_total = $8, status = $9,
		    reconciled_by = $10, reconciled_at = $11, updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2`

	_, err := s.bankTxnRepo.GetPool().Exec(ctx, query,
		record.ID, record.TenantID, record.BankBalance, record.BookBalance,
		record.AdjustedBalance, record.BankOnlyTotal, record.BookOnlyTotal,
		record.Status, record.ReconciledBy, record.ReconciledAt,
	)
	return err
}

func (s *BankReconciliationService) getUnreconciledItems(ctx context.Context, reconciliationRecordID uuid.UUID) ([]model.UnreconciledItem, error) {
	query := `
		SELECT id, reconciliation_record_id, item_type, source_type, source_id,
		       txn_date, description, debit, credit, amount, tenant_id, created_at
		FROM unreconciled_items
		WHERE reconciliation_record_id = $1
		ORDER BY txn_date`

	rows, err := s.bankTxnRepo.GetPool().Query(ctx, query, reconciliationRecordID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.UnreconciledItem
	for rows.Next() {
		var item model.UnreconciledItem
		err := rows.Scan(
			&item.ID, &item.ReconciliationRecordID, &item.ItemType, &item.SourceType,
			&item.SourceID, &item.TxnDate, &item.Description, &item.Debit, &item.Credit,
			&item.Amount, &item.TenantID, &item.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// SaveReconciliationResult saves the reconciliation result to the database.
func (s *BankReconciliationService) SaveReconciliationResult(ctx context.Context, tenantID, bankAccountID uuid.UUID, periodNo int, result *ReconciliationResult) error {
	// Check if record exists
	existing, _ := s.getReconciliationRecord(ctx, tenantID, bankAccountID, periodNo)

	var bankOnlyTotal, bookOnlyTotal decimal.Decimal
	for _, item := range result.BankOnlyItems {
		bankOnlyTotal = bankOnlyTotal.Add(item.Amount)
	}
	for _, item := range result.BookOnlyItems {
		bookOnlyTotal = bookOnlyTotal.Add(item.Amount)
	}

	if existing != nil {
		// Update
		query := `
			UPDATE reconciliation_records
			SET bank_balance = $4, book_balance = $5, adjusted_balance = $6,
			    bank_only_total = $7, book_only_total = $8, status = $9, updated_at = NOW()
			WHERE tenant_id = $1 AND bank_account_id = $2 AND period_no = $3`
		_, err := s.bankTxnRepo.GetPool().Exec(ctx, query,
			tenantID, bankAccountID, periodNo,
			result.BankBalance, result.BookBalance, result.AdjustedBalance,
			bankOnlyTotal, bookOnlyTotal, "in_progress",
		)
		if err != nil {
			return fmt.Errorf("update reconciliation record: %w", err)
		}
	} else {
		// Insert
		query := `
			INSERT INTO reconciliation_records (id, tenant_id, bank_account_id, period_no,
				bank_balance, book_balance, adjusted_balance, bank_only_total, book_only_total, status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			ON CONFLICT (tenant_id, bank_account_id, period_no) DO UPDATE
			SET bank_balance = $5, book_balance = $6, adjusted_balance = $7,
			    bank_only_total = $8, book_only_total = $9, status = $10, updated_at = NOW()`
		_, err := s.bankTxnRepo.GetPool().Exec(ctx, query,
			uuid.New(), tenantID, bankAccountID, periodNo,
			result.BankBalance, result.BookBalance, result.AdjustedBalance,
			bankOnlyTotal, bookOnlyTotal, "in_progress",
		)
		if err != nil {
			return fmt.Errorf("insert reconciliation record: %w", err)
		}
	}

	return nil
}

type BalanceCheckItem struct {
	BankAccountID   uuid.UUID       `json:"bank_account_id"`
	BankName        string          `json:"bank_name"`
	OpeningBalance  decimal.Decimal `json:"opening_balance"`
	TxnInflowTotal  decimal.Decimal `json:"txn_inflow_total"`
	TxnOutflowTotal decimal.Decimal `json:"txn_outflow_total"`
	ExpectedBalance decimal.Decimal `json:"expected_balance"`
	StoredBalance   decimal.Decimal `json:"stored_balance"`
	Difference      decimal.Decimal `json:"difference"`
	TxnCount        int             `json:"txn_count"`
	HasInconsistency bool           `json:"has_inconsistency"`
}

type BalanceCheckResult struct {
	TotalAccounts     int                `json:"total_accounts"`
	ConsistentCount   int                `json:"consistent_count"`
	InconsistentCount int                `json:"inconsistent_count"`
	Items             []BalanceCheckItem `json:"items"`
}

func (s *BankReconciliationService) BalanceCheck(ctx context.Context, tenantID uuid.UUID) (*BalanceCheckResult, error) {
	accounts, err := s.bankRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list bank accounts: %w", err)
	}

	result := &BalanceCheckResult{TotalAccounts: len(accounts), Items: []BalanceCheckItem{}}

	for _, acct := range accounts {
		netChange, txnCount, err := s.bankTxnRepo.GetNetChangeByAccount(ctx, tenantID, acct.ID)
		if err != nil {
			return nil, fmt.Errorf("sum txns for %s: %w", acct.ID, err)
		}

		expected := acct.OpeningBalance.Add(netChange)
		diff := expected.Sub(acct.CurrentBalance)

		hasIssue := !diff.IsZero()
		if hasIssue {
			result.InconsistentCount++
		} else {
			result.ConsistentCount++
		}

		inflow := decimal.Zero
		outflow := decimal.Zero
		if !acct.OpeningBalance.IsZero() || txnCount > 0 {
			inflow, outflow, _ = s.bankTxnRepo.GetDirectionTotalsByAccount(ctx, tenantID, acct.ID)
		}

		bankName := acct.BankName
		if acct.BankAccountType != nil && *acct.BankAccountType == "cash" {
			bankName = "[现金] " + bankName
		}

		result.Items = append(result.Items, BalanceCheckItem{
			BankAccountID:    acct.ID,
			BankName:         bankName,
			OpeningBalance:   acct.OpeningBalance,
			TxnInflowTotal:   inflow,
			TxnOutflowTotal:  outflow,
			ExpectedBalance:  expected,
			StoredBalance:    acct.CurrentBalance,
			Difference:       diff,
			TxnCount:         txnCount,
			HasInconsistency: hasIssue,
		})
	}
	return result, nil
}

func (s *BankReconciliationService) GetBankAccountForReport(ctx context.Context, tenantID, bankAccountID uuid.UUID) (string, error) {
	acct, err := s.bankRepo.GetByID(ctx, tenantID, bankAccountID)
	if err != nil {
		return "", err
	}
	return acct.BankName, nil
}
