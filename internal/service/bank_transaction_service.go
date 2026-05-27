package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
	"huihua/finance/internal/model"
	"huihua/finance/internal/repository"
)

// BankTransactionService handles bank transaction business logic.
type BankTransactionService struct {
	repo              *repository.BankTransactionRepository
	classificationSvc *ClassificationRuleService
	bankRepo          *repository.BankRepository
}

// NewBankTransactionService creates a new BankTransactionService.
func NewBankTransactionService(repo *repository.BankTransactionRepository, classificationSvc *ClassificationRuleService, bankRepo *repository.BankRepository) *BankTransactionService {
	return &BankTransactionService{
		repo:              repo,
		classificationSvc: classificationSvc,
		bankRepo:          bankRepo,
	}
}

// ImportFromExcel parses Excel data and imports bank transactions with auto-classification.
func (s *BankTransactionService) ImportFromExcel(ctx context.Context, tenantID, bankAccountID uuid.UUID, data []byte) (*model.ImportResult, error) {
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("open excel: %w", err)
	}
	defer f.Close()

	// Get first sheet
	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("get rows: %w", err)
	}

	result := &model.ImportResult{
		TotalRows: len(rows) - 1, // exclude header
	}

	// Get company_id from bank account
	bankAccount, err := s.bankRepo.GetByID(ctx, tenantID, bankAccountID)
	if err != nil {
		return nil, fmt.Errorf("get bank account: %w", err)
	}

	// Parse header row to find column indices
	headerMap := make(map[string]int)
	if len(rows) > 0 {
		for i, col := range rows[0] {
			headerMap[strings.ToLower(strings.TrimSpace(col))] = i
		}
	}

	// Required columns: date, description
	dateIdx, ok := headerMap["date"]
	if !ok {
		dateIdx, ok = headerMap["transaction date"]
	}
	descIdx, ok := headerMap["description"]
	if !ok {
		descIdx, ok = headerMap["摘要"]
	}

	var txns []model.BankTransaction
	for rowIdx, row := range rows[1:] { // skip header
		rowNum := rowIdx + 2 // 1-indexed, header is row 1

		if len(row) == 0 {
			continue
		}

		// Parse date
		var txnDate time.Time
		if dateIdx < len(row) {
			dateStr := strings.TrimSpace(row[dateIdx])
			txnDate, err = parseDate(dateStr)
			if err != nil {
				result.FailedCount++
				result.FailedRows = append(result.FailedRows, rowNum)
				continue
			}
		}

		// Parse description
		var description *string
		if descIdx < len(row) {
			desc := strings.TrimSpace(row[descIdx])
			if desc != "" {
				description = &desc
			}
		}

		if description == nil {
			result.FailedCount++
			result.FailedRows = append(result.FailedRows, rowNum)
			continue
		}

		// Parse income (debit) and expense (credit)
		var debit, credit decimal.Decimal
		incomeIdx := headerMap["income"]
		expenseIdx := headerMap["expense"]
		amountIdx, _ := headerMap["amount"]

		if incomeIdx < len(row) && row[incomeIdx] != "" {
			if v, err := strconv.ParseFloat(strings.ReplaceAll(row[incomeIdx], ",", ""), 64); err == nil && v > 0 {
				debit = decimal.NewFromFloat(v)
			}
		}
		if expenseIdx < len(row) && row[expenseIdx] != "" {
			if v, err := strconv.ParseFloat(strings.ReplaceAll(row[expenseIdx], ",", ""), 64); err == nil && v > 0 {
				credit = decimal.NewFromFloat(v)
			}
		}
		if amountIdx < len(row) && row[amountIdx] != "" {
			if v, err := strconv.ParseFloat(strings.ReplaceAll(row[amountIdx], ",", ""), 64); err == nil {
				if v > 0 {
					debit = decimal.NewFromFloat(v)
				} else if v < 0 {
					credit = decimal.NewFromFloat(-v)
				}
			}
		}

		// Determine direction
		var direction *string
		if debit.GreaterThan(decimal.Zero) {
			directionStr := "debit"
			direction = &directionStr
		} else if credit.GreaterThan(decimal.Zero) {
			directionStr := "credit"
			direction = &directionStr
		}

		// Parse counterparty
		var counterparty *string
		cpIdx, ok := headerMap["counterparty"]
		if !ok {
			cpIdx, _ = headerMap["对方账户"]
		}
		if cpIdx < len(row) {
			cp := strings.TrimSpace(row[cpIdx])
			if cp != "" {
				counterparty = &cp
			}
		}

		// Parse reference no (optional)
		var referenceNo *string
		refIdx, ok := headerMap["voucher no"]
		if !ok {
			refIdx, _ = headerMap["凭证号"]
		}
		if refIdx < len(row) {
			ref := strings.TrimSpace(row[refIdx])
			if ref != "" {
				referenceNo = &ref
			}
		}

		txn := model.BankTransaction{
			ID:              uuid.New(),
			TenantID:        tenantID,
			BankAccountID:   bankAccountID,
			TxnDate:         txnDate,
			Description:     description,
			Debit:           debit,
			Credit:          credit,
			Direction:       direction,
			ReferenceNo:     referenceNo,
			CounterpartyName: counterparty,
			Matched:         false,
			CompanyID:       bankAccount.CompanyID,
		}

		// Auto-classify using classification rules
		descStr := ""
		if description != nil {
			descStr = *description
		}

		var amount decimal.Decimal
		if debit.GreaterThan(decimal.Zero) {
			amount = debit
		} else {
			amount = credit
		}

		dirStr := "credit"
		if debit.GreaterThan(decimal.Zero) {
			dirStr = "debit"
		}

		matchResult, err := s.classificationSvc.MatchTransaction(ctx, tenantID, descStr, amount, dirStr)
		if err == nil && matchResult != nil && matchResult.Matched && matchResult.RuleID != nil {
			txn.Matched = true
			// Store match info in RawData
			rawData, _ := json.Marshal(map[string]interface{}{
				"rule_id":   matchResult.RuleID,
				"rule_name": matchResult.RuleName,
				"account_id": matchResult.AccountID,
			})
			txn.RawData = rawData
		}

		txns = append(txns, txn)
		result.SuccessCount++
	}

	// Batch import
	if len(txns) > 0 {
		_, err = s.repo.ImportBatch(ctx, tenantID, bankAccountID, txns)
		if err != nil {
			return nil, fmt.Errorf("import batch: %w", err)
		}
	}

	return result, nil
}

// ClassifyTransaction re-classifies a single bank transaction.
func (s *BankTransactionService) ClassifyTransaction(ctx context.Context, tenantID, txnID uuid.UUID) (*model.RuleMatchResult, error) {
	txn, err := s.repo.GetByID(ctx, tenantID, txnID)
	if err != nil {
		return nil, fmt.Errorf("get transaction: %w", err)
	}

	descStr := ""
	if txn.Description != nil {
		descStr = *txn.Description
	}

	var amount decimal.Decimal
	if txn.Debit.GreaterThan(decimal.Zero) {
		amount = txn.Debit
	} else {
		amount = txn.Credit
	}

	dirStr := "credit"
	if txn.Debit.GreaterThan(decimal.Zero) {
		dirStr = "debit"
	}

	matchResult, err := s.classificationSvc.MatchTransaction(ctx, tenantID, descStr, amount, dirStr)
	if err != nil {
		return nil, fmt.Errorf("match transaction: %w", err)
	}

	// Update transaction with match result
	if matchResult.Matched && matchResult.RuleID != nil {
		err = s.repo.UpdateMatchedInfo(ctx, tenantID, txnID, *matchResult.RuleID)
		if err != nil {
			return nil, fmt.Errorf("update matched info: %w", err)
		}
	}

	return matchResult, nil
}

// ClassifyAllPending re-classifies all pending (unmatched) transactions for a bank account.
func (s *BankTransactionService) ClassifyAllPending(ctx context.Context, tenantID, bankAccountID uuid.UUID) (int, error) {
	txns, err := s.repo.GetUnmatched(ctx, tenantID, bankAccountID)
	if err != nil {
		return 0, fmt.Errorf("get unmatched: %w", err)
	}

	classifiedCount := 0
	for _, txn := range txns {
		descStr := ""
		if txn.Description != nil {
			descStr = *txn.Description
		}

		var amount decimal.Decimal
		if txn.Debit.GreaterThan(decimal.Zero) {
			amount = txn.Debit
		} else {
			amount = txn.Credit
		}

		dirStr := "credit"
		if txn.Debit.GreaterThan(decimal.Zero) {
			dirStr = "debit"
		}

		matchResult, err := s.classificationSvc.MatchTransaction(ctx, tenantID, descStr, amount, dirStr)
		if err != nil || matchResult == nil || !matchResult.Matched {
			continue
		}

		if matchResult.RuleID != nil {
			err = s.repo.UpdateMatchedInfo(ctx, tenantID, txn.ID, *matchResult.RuleID)
			if err == nil {
				classifiedCount++
			}
		}
	}

	return classifiedCount, nil
}

// ListTransactions lists bank transactions with filters.
func (s *BankTransactionService) ListTransactions(ctx context.Context, tenantID, bankAccountID uuid.UUID, filter model.BankTxnFilter) ([]model.BankTransaction, int, error) {
	return s.repo.ListByBankAccount(ctx, tenantID, bankAccountID, filter)
}

// GetTransaction gets a single bank transaction.
func (s *BankTransactionService) GetTransaction(ctx context.Context, tenantID, id uuid.UUID) (*model.BankTransaction, error) {
	return s.repo.GetByID(ctx, tenantID, id)
}

// DeleteTransaction deletes a bank transaction.
func (s *BankTransactionService) DeleteTransaction(ctx context.Context, tenantID, id uuid.UUID) error {
	return s.repo.UpdateStatus(ctx, tenantID, id, false)
}

// MarkAsMatched marks transactions as matched.
func (s *BankTransactionService) MarkAsMatched(ctx context.Context, tenantID uuid.UUID, ids []uuid.UUID, journalEntryID uuid.UUID) error {
	return s.repo.MarkAsMatched(ctx, tenantID, ids, journalEntryID)
}

// MarkAsReconciled marks transactions as reconciled.
func (s *BankTransactionService) MarkAsReconciled(ctx context.Context, tenantID, bankAccountID uuid.UUID, periodNo int) error {
	return s.repo.MarkAsReconciled(ctx, tenantID, bankAccountID, periodNo)
}

// GetUnmatched gets all unmatched transactions.
func (s *BankTransactionService) GetUnmatched(ctx context.Context, tenantID, bankAccountID uuid.UUID) ([]model.BankTransaction, error) {
	return s.repo.GetUnmatched(ctx, tenantID, bankAccountID)
}

// parseDate tries to parse a date string in various formats.
func parseDate(s string) (time.Time, error) {
	formats := []string{
		"2006-01-02",
		"2006/01/02",
		"01/02/2006",
		"2006-01-02 15:04:05",
		"2006/01/02 15:04:05",
		"02/01/2006",
		"2006年01月02日",
	}

	s = strings.TrimSpace(s)
	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("cannot parse date: %s", s)
}