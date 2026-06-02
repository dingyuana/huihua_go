package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
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
	partySvc          *PartyService
}

// NewBankTransactionService creates a new BankTransactionService.
func NewBankTransactionService(repo *repository.BankTransactionRepository, classificationSvc *ClassificationRuleService, bankRepo *repository.BankRepository, partySvc *PartyService) *BankTransactionService {
	return &BankTransactionService{
		repo:              repo,
		classificationSvc: classificationSvc,
		bankRepo:          bankRepo,
		partySvc:          partySvc,
	}
}

// ImportFromExcel parses Excel data and imports bank transactions with auto-classification.
// When autoCreateParty is true and a row's counterparty is missing, the system extracts a
// candidate name from the description, creates (or reuses) a parties record, and uses the
// resulting name as the bank transaction's counterparty_name.
func (s *BankTransactionService) ImportFromExcel(ctx context.Context, tenantID, bankAccountID uuid.UUID, data []byte, autoCreateParty bool) (*model.ImportResult, error) {
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

	// Find the real header row (skip empty rows and title rows)
	headerRowIndex := 0
	for i, row := range rows {
		nonEmpty := 0
		for _, cell := range row {
			if strings.TrimSpace(cell) != "" {
				nonEmpty++
			}
		}
		if nonEmpty >= 5 {
			headerRowIndex = i
			break
		}
	}

	// Parse header row to find column indices (support English and Chinese names)
	headerMap := make(map[string]int)
	headerCols := make([]string, 0)
	if len(rows) > headerRowIndex {
		for i, col := range rows[headerRowIndex] {
			headerMap[strings.ToLower(strings.TrimSpace(col))] = i
			headerCols = append(headerCols, col)
		}
	}

	// Adjust total rows: count only rows with at least 1 non-blank cell.
	result.TotalRows = 0
	for _, row := range rows[headerRowIndex+1:] {
		hasContent := false
		for _, cell := range row {
			if strings.TrimSpace(cell) != "" {
				hasContent = true
				break
			}
		}
		if hasContent {
			result.TotalRows++
		}
	}

	// Helper to find a column by trying multiple possible names
	// Tries exact match first, then substring match (for bank columns like "交易日期[Transaction Date]")
	findCol := func(names ...string) (int, bool) {
		// Phase 1: exact match (fast path)
		for _, name := range names {
			key := strings.ToLower(strings.TrimSpace(name))
			if idx, ok := headerMap[key]; ok {
				return idx, true
			}
		}
		// Phase 2: substring match — does actual column name contain any candidate?
		for ci, col := range headerCols {
			colLower := strings.ToLower(col)
			for _, name := range names {
				if strings.Contains(colLower, strings.ToLower(name)) {
					return ci, true
				}
			}
		}
		return 0, false
	}

	// Required columns: date, description
	dateIdx, dateFound := findCol("date", "transaction date", "交易日期", "记账日期", "发生日期", "日期")
	descIdx, descFound := findCol("description", "摘要", "交易描述", "备注", "用途", "附言")

	if !dateFound {
		return nil, fmt.Errorf("找不到日期列，支持的列名：日期/交易日期/记账日期/发生日期/date")
	}
	if !descFound {
		return nil, fmt.Errorf("找不到摘要列，支持的列名：摘要/备注/用途/附言/description")
	}

	// Also find fallback description columns (交易附言/备注/用途)
	remarkIdx, remarkFound := findCol("交易附言", "remark", "备注", "remarks")
	purposeIdx, purposeFound := findCol("用途", "purpose")

	var txns []model.BankTransaction
	for rowIdx, row := range rows[headerRowIndex+1:] {
		rowNum := rowIdx + headerRowIndex + 2 // 1-indexed, header row found above

		if len(row) == 0 {
			continue
		}

		// Skip rows where ALL cells are empty/whitespace
		allEmpty := true
		for _, cell := range row {
			if strings.TrimSpace(cell) != "" {
				allEmpty = false
				break
			}
		}
		if allEmpty {
			continue
		}

		// Parse date
		var txnDate time.Time
		dateStr := ""
		if dateIdx < len(row) {
			dateStr = strings.TrimSpace(row[dateIdx])
			txnDate, err = parseDate(dateStr)
			if err != nil {
				result.FailedCount++
				result.FailedRows = append(result.FailedRows, rowNum)
				result.FailedReasons = append(result.FailedReasons, model.FailedRowDetail{
					Row: rowNum, Date: dateStr, Reason: "日期格式无法解析: " + dateStr,
				})
				continue
			}
		} else {
			result.FailedCount++
			result.FailedRows = append(result.FailedRows, rowNum)
			result.FailedReasons = append(result.FailedReasons, model.FailedRowDetail{
				Row: rowNum, Reason: "该行没有日期列数据",
			})
			continue
		}

		// Parse description — try main column first, then fallback to 交易附言/备注/用途
		var description *string
		if descIdx < len(row) {
			desc := strings.TrimSpace(row[descIdx])
			if desc != "" {
				description = &desc
			}
		}
		if description == nil && remarkFound && remarkIdx < len(row) {
			desc := strings.TrimSpace(row[remarkIdx])
			if desc != "" {
				description = &desc
			}
		}
		if description == nil && purposeFound && purposeIdx < len(row) {
			desc := strings.TrimSpace(row[purposeIdx])
			if desc != "" {
				description = &desc
			}
		}

		if description == nil {
			result.FailedCount++
			result.FailedRows = append(result.FailedRows, rowNum)
			result.FailedReasons = append(result.FailedReasons, model.FailedRowDetail{
				Row: rowNum, Date: dateStr, Reason: "摘要为空，需要手工补录",
			})
			continue
		}

		// Parse income (debit) and expense (credit)
		var debit, credit decimal.Decimal
		incomeIdx, _ := findCol("income", "收入金额", "贷方金额", "收入", "credit")
		expenseIdx, _ := findCol("expense", "支出金额", "借方金额", "支出", "debit")
		amountIdx, _ := findCol("amount", "金额", "发生金额", "交易金额")

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

		if debit.IsZero() && credit.IsZero() {
			result.FailedCount++
			result.FailedRows = append(result.FailedRows, rowNum)
			result.FailedReasons = append(result.FailedReasons, model.FailedRowDetail{
				Row: rowNum, Date: dateStr, Reason: "金额为空或解析为 0，跳过空白行",
			})
			continue
		}

		// Determine direction
		var direction *string
		if debit.GreaterThan(decimal.Zero) {
			directionStr := "in"
			direction = &directionStr
		} else if credit.GreaterThan(decimal.Zero) {
			directionStr := "out"
			direction = &directionStr
		} else {
			dirIdx, _ := findCol("direction", "收支方向", "方向", "借贷方向")
			if dirIdx < len(row) {
				dirVal := strings.ToLower(strings.TrimSpace(row[dirIdx]))
				if dirVal == "收入" || dirVal == "in" || dirVal == "贷方" || dirVal == "credit" || dirVal == "收" {
					directionStr := "in"
					direction = &directionStr
				} else if dirVal == "支出" || dirVal == "out" || dirVal == "借方" || dirVal == "debit" || dirVal == "支" {
					directionStr := "out"
					direction = &directionStr
				}
			}
		}

		cpIdx, cpFound := findCol(
			"counterparty", "对方户名", "对方名称", "对方账户", "交易对方",
			"付款人名称", "收款人名称",
			"付款人", "收款人",
		)
		var counterparty *string
		if cpFound && cpIdx < len(row) {
			cp := strings.TrimSpace(row[cpIdx])
			if cp != "" && !looksLikeBankCode(cp) {
				counterparty = &cp
			}
		}

		if counterparty == nil && autoCreateParty && s.partySvc != nil && description != nil {
			if candidate := extractCounterpartyName(*description); candidate != "" {
				partyType := "both"
				if direction != nil {
					if *direction == "in" {
						partyType = "customer"
					} else if *direction == "out" {
						partyType = "supplier"
					}
				}
				name, created, err := s.partySvc.EnsureParty(ctx, tenantID, &model.Party{
					PartyType: partyType,
					Name:      candidate,
					IsActive:  true,
				})
				if err == nil {
					counterparty = &name
					if created {
						result.AutoCreatedParties++
					}
				}
			}
		}

		// Parse reference no (optional)
		var referenceNo *string
		refIdx, _ := findCol("voucher no", "reference", "流水号", "凭证号", "交易流水号", "交易编号", "ref")
		if refIdx < len(row) {
			ref := strings.TrimSpace(row[refIdx])
			if ref != "" {
				referenceNo = &ref
			}
		}

		txn := model.BankTransaction{
			ID:               uuid.New(),
			TenantID:         tenantID,
			BankAccountID:    bankAccountID,
			TxnDate:          txnDate,
			Description:      description,
			Debit:            debit,
			Credit:           credit,
			Direction:        direction,
			ReferenceNo:      referenceNo,
			CounterpartyName: counterparty,
			Matched:          false,
			CompanyID:        bankAccount.CompanyID,
		}

		// Auto-classify using classification rules
		descStr := ""
		if description != nil {
			descStr = *description
		}

		counterpartyStr := ""
		if counterparty != nil {
			counterpartyStr = *counterparty
		}

		txnDirection := ""
		if debit.GreaterThan(decimal.Zero) {
			txnDirection = "in"
		} else {
			txnDirection = "out"
		}

		matchResult, err := s.classificationSvc.MatchTransaction(ctx, tenantID, descStr, counterpartyStr, txnDirection)
		if err == nil && matchResult != nil && matchResult.Matched && matchResult.RuleID != nil {
			if matchResult.Classification != nil {
				txn.Classification = matchResult.Classification
			}
			rawData, _ := json.Marshal(map[string]interface{}{
				"rule_id":        matchResult.RuleID,
				"rule_name":      matchResult.RuleName,
				"classification": matchResult.Classification,
			})
			txn.RawData = rawData
		} else {
			// Smart fallback when no rule matches
			cls := fallbackClassify(descStr, counterpartyStr, txnDirection, debit, credit)
			txn.Classification = &cls
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
		result.ImportedTxns = txns
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

	counterpartyStr := ""
	if txn.CounterpartyName != nil {
		counterpartyStr = *txn.CounterpartyName
	}

	txnDirection := ""
	if txn.Debit.GreaterThan(decimal.Zero) {
		txnDirection = "in"
	} else {
		txnDirection = "out"
	}

	matchResult, err := s.classificationSvc.MatchTransaction(ctx, tenantID, descStr, counterpartyStr, txnDirection)
	if err != nil {
		return nil, fmt.Errorf("match transaction: %w", err)
	}

	// Determine classification
	classification := "pending"
	if matchResult.Matched && matchResult.Classification != nil {
		classification = *matchResult.Classification
	} else {
		classification = fallbackClassify(descStr, counterpartyStr, txnDirection, txn.Debit, txn.Credit)
	}

	// Update transaction with match result and classification
	if matchResult.Matched && matchResult.RuleID != nil {
		err = s.repo.UpdateMatchedInfo(ctx, tenantID, txnID, *matchResult.RuleID)
		if err != nil {
			return nil, fmt.Errorf("update matched info: %w", err)
		}
	}

	// Update classification in DB
	err = s.repo.UpdateClassification(ctx, tenantID, txnID, classification)
	if err != nil {
		return nil, fmt.Errorf("update classification: %w", err)
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

		counterpartyStr := ""
		if txn.CounterpartyName != nil {
			counterpartyStr = *txn.CounterpartyName
		}

		txnDirection := ""
		if txn.Debit.GreaterThan(decimal.Zero) {
			txnDirection = "in"
		} else {
			txnDirection = "out"
		}

		matchResult, err := s.classificationSvc.MatchTransaction(ctx, tenantID, descStr, counterpartyStr, txnDirection)

		// Determine classification
		classification := "pending"
		if err == nil && matchResult != nil && matchResult.Matched && matchResult.Classification != nil {
			classification = *matchResult.Classification
		} else {
			classification = fallbackClassify(descStr, counterpartyStr, txnDirection, txn.Debit, txn.Credit)
		}

		// Update classification in DB
		_ = s.repo.UpdateClassification(ctx, tenantID, txn.ID, classification)

		if matchResult != nil && matchResult.Matched && matchResult.RuleID != nil {
			err = s.repo.UpdateMatchedInfo(ctx, tenantID, txn.ID, *matchResult.RuleID)
		}
		classifiedCount++
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
		"20060102",
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

// fallbackClassify determines a business classification using heuristics
// when no classification rule matches. It uses direction, description keywords,
// counterparty patterns, and amount thresholds.
func fallbackClassify(description, counterparty, direction string, debit, credit decimal.Decimal) string {
	desc := strings.ToLower(description)
	cp := strings.ToLower(counterparty)

	// Fee-related keywords → bank_fee (银行费用)
	feeKeywords := []string{"手续费", "服务费", "账户管理费", "年费", "短信费", "工本费", "汇费", "电报费", "网银费"}
	for _, kw := range feeKeywords {
		if strings.Contains(desc, kw) {
			return "bank_fee"
		}
	}

	// Interest-related keywords → interest_income (利息收入)
	interestKeywords := []string{"利息", "结息", "派息"}
	for _, kw := range interestKeywords {
		if strings.Contains(desc, kw) {
			return "interest_income"
		}
	}

	// Insurance-related keywords → insurance_fee (保险费用)
	insuranceKeywords := []string{"保险", "保费", "投保", "财产险", "责任险", "雇主责任险", "意外险"}
	for _, kw := range insuranceKeywords {
		if strings.Contains(desc, kw) || strings.Contains(cp, kw) {
			return "insurance_fee"
		}
	}

	// Tax payment keywords → tax_payment (税务缴费)
	taxKeywords := []string{"税款", "税金", "缴税", "扣税", "增值税", "所得税", "附加税", "社保", "公积金", "实时缴税", "国税", "地税", "税务局", "国家金库", "国库", "印花税", "城建税", "教育费附加"}
	for _, kw := range taxKeywords {
		if strings.Contains(desc, kw) || strings.Contains(cp, kw) {
			return "tax_payment"
		}
	}

	// Transfer between own accounts → internal_transfer
	transferKeywords := []string{"转账", "划转", "调拨", "转存", "转入", "转出"}
	for _, kw := range transferKeywords {
		if strings.Contains(desc, kw) {
			return "internal_transfer"
		}
	}

	// Small-amount outgoing → likely bank_fee (e.g. < 50 CNY)
	amount := decimal.Zero
	if direction == "out" {
		amount = credit
	} else {
		amount = debit
	}
	threshold := decimal.NewFromFloat(50)
	if direction == "out" && amount.GreaterThan(decimal.Zero) && amount.LessThanOrEqual(threshold) {
		return "bank_fee"
	}

	// Direction-based default
	if direction == "in" {
		return "business_receipt"
	}
	return "business_payment"
}

var (
	counterpartyTaxBureauRe = regexp.MustCompile(`(?:国家税务总局\p{Han}{0,15}税务局|\p{Han}{2,20}税务局)`)
	counterpartyGovRe       = regexp.MustCompile(`\p{Han}{2,20}(?:社保局|公积金中心|社保中心|海关)`)
	counterpartyCompanyRe   = regexp.MustCompile(`\p{Han}{2,30}(?:有限公司|股份有限公司|集团|有限责任公司|股份公司|总公司|分公司|子公司|集团公司)`)
	counterpartyShortOrgRe  = regexp.MustCompile(`\p{Han}{4,20}(?:公司|厂|店|商行|银行|事务所|医院|学校|中心)`)
)

func extractCounterpartyName(description string) string {
	desc := strings.TrimSpace(description)
	if desc == "" {
		return ""
	}
	if m := counterpartyTaxBureauRe.FindString(desc); m != "" {
		return strings.TrimSpace(m)
	}
	if m := counterpartyGovRe.FindString(desc); m != "" {
		return strings.TrimSpace(m)
	}
	if m := counterpartyCompanyRe.FindString(desc); m != "" {
		return strings.TrimSpace(m)
	}
	if m := counterpartyShortOrgRe.FindString(desc); m != "" {
		return strings.TrimSpace(m)
	}
	return ""
}

var bankCodeOnlyRe = regexp.MustCompile(`^\d{4,20}$`)

func looksLikeBankCode(s string) bool {
	return bankCodeOnlyRe.MatchString(strings.TrimSpace(s))
}
