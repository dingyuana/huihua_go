package service

import (
	"context"
	"fmt"
	"math"
	"strings"
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

// MatchScore holds the score for a bank transaction ↔ GL entry match candidate.
type MatchScore struct {
	TotalScore   float64    `json:"total_score"`
	AmountScore  float64    `json:"amount_score"`  // 0-50
	DateScore    float64    `json:"date_score"`    // 0-20
	NameScore    float64    `json:"name_score"`    // 0-15
	DescScore    float64    `json:"desc_score"`    // 0-10
	RefNoScore   float64    `json:"ref_no_score"`  // 0-5
	IsAutoMatched bool     `json:"is_auto_matched"` // ≥85分
	NeedConfirm  bool       `json:"need_confirm"`    // 60-85分
	BankTxnID    uuid.UUID  `json:"bank_txn_id"`
	GLEntryID    uuid.UUID  `json:"gl_entry_id"`
}

// PendingConfirmItem represents an item pending manual confirmation.
type PendingConfirmItem struct {
	BankTxnID   uuid.UUID       `json:"bank_txn_id"`
	BankTxnDesc string          `json:"bank_txn_desc"`
	BankTxnAmt  decimal.Decimal `json:"bank_txn_amt"`
	GLEntryID   uuid.UUID       `json:"gl_entry_id"`
	GLEntryDesc string          `json:"gl_entry_desc"`
	GLEntryAmt  decimal.Decimal `json:"gl_entry_amt"`
	Score       MatchScore      `json:"score"`
}

// ReconciliationResult holds the result of a reconciliation operation.
type MatchItemResult struct {
	ID          uuid.UUID  `json:"id"`
	Score       float64    `json:"score"`
	BankTxnDesc string     `json:"bank_txn"`
	GLEntryDesc string     `json:"gl_entry"`
	NeedConfirm bool       `json:"needConfirm"`
}

type ReconciliationResult struct {
	BankBalance       decimal.Decimal       `json:"bank_balance"`
	BookBalance       decimal.Decimal       `json:"book_balance"`
	AdjustedBalance   decimal.Decimal       `json:"adjusted_balance"`
	BankOnlyItems     []UnreconciledItem    `json:"bank_only_items"`
	BookOnlyItems     []UnreconciledItem    `json:"book_only_items"`
	MatchedCount      int                   `json:"matched_count"`
	TotalMatched      decimal.Decimal       `json:"total_matched"`
	PendingItems      []PendingConfirmItem   `json:"pending_items"`
	AutoMatchedCount  int                   `json:"auto_matched_count"`
	MatchItems        []MatchItemResult      `json:"match_items"`
}

// UnreconciledItem represents an item that appears in one side but not the other.
type UnreconciledItem struct {
	ID          uuid.UUID       `json:"id"`
	SourceType  string          `json:"source_type"`
	TxnDate     time.Time       `json:"txn_date"`
	Description string          `json:"description"`
	Amount      decimal.Decimal `json:"amount"`
	Direction   string          `json:"direction"`
	ItemType    string          `json:"item_type"` // four-direction type
}

// Item type constants for four-direction unreconciled items
const (
	ItemTypeBankReceiptNotInGL = "bank_receipt_not_in_gl" // bank credit, book missing
	ItemTypeBankPaymentNotInGL = "bank_payment_not_in_gl" // bank debit, book missing
	ItemTypeGLReceiptNotInBank = "gl_receipt_not_in_bank" // book credit, bank missing
	ItemTypeGLPaymentNotInBank = "gl_payment_not_in_bank"  // book debit, bank missing
)

// keywords for description matching
var descKeywords = []string{
	"货款", "采购", "付款", "收款", "退款", "手续费", "服务费",
	"工资", "奖金", "报销", "房租", "水电", "运费", "税费",
	"提成", "佣金", "赔偿", "还款", "分红", "投资",
}

// CalculateMatchScore computes the 5-dimension match score for a bank txn ↔ GL entry pair.
func CalculateMatchScore(bankTxn *model.BankTransaction, glEntry *model.GLEntry) MatchScore {
	score := MatchScore{
		BankTxnID: bankTxn.ID,
		GLEntryID: glEntry.ID,
	}

	// 1. Amount Score (50pts max): 完全一致=50, 差异1%内按比例衰减
	bankAmt := bankTxn.Debit.Sub(bankTxn.Credit).Abs()
	glAmt := glEntry.Debit.Sub(glEntry.Credit).Abs()

	if bankAmt.IsZero() && glAmt.IsZero() {
		score.AmountScore = 50
	} else if !bankAmt.IsZero() {
		diff := bankAmt.Sub(glAmt).Abs()
		pct := diff.Div(bankAmt).InexactFloat64()
		if pct < 0.01 {
			// 差异1%内，按比例给分
			score.AmountScore = 50 * (1 - pct*100)
		} else if pct < 0.05 {
			// 差异5%内，按比例衰减但最低30分
			score.AmountScore = 30 + 20*(1-pct*20)
		}
		// else 差异>=5%得0分
	}

	// 2. Date Score (20pts max): 同一天=20, ±1天按比例衰减, ±3天以上=0
	daysDiff := math.Abs(bankTxn.TxnDate.Sub(glEntry.PostingDate).Hours() / 24)
	if daysDiff == 0 {
		score.DateScore = 20
	} else if daysDiff <= 1 {
		score.DateScore = 20 * (1 - float64(daysDiff))
	} else if daysDiff <= 3 {
		score.DateScore = 20 * (1 - float64(daysDiff)/3 * 0.5)
	}
	// daysDiff > 3 → 0分

	// 3. Name Score (15pts max): Levenshtein similarity
	bankName := strValueOf(bankTxn.CounterpartyName)
	glName := glEntry.PartyID.String() // fallback, actual PartyName would need lookup
	nameSim := levenshteinSimilarity(bankName, glName)
	if nameSim >= 0.9 {
		score.NameScore = 15
	} else if nameSim >= 0.8 {
		score.NameScore = 10
	} else if nameSim >= 0.7 {
		score.NameScore = 5
	}

	// 4. Desc Score (10pts max): 关键词匹配
	bankDesc := strValueOf(bankTxn.Description)
	glDesc := "" // GL description not available in GLEntry model directly
	descMatch := containsKeywords(bankDesc, descKeywords) + containsKeywords(glDesc, descKeywords)
	if descMatch >= 2 {
		score.DescScore = 10
	} else if descMatch >= 1 {
		score.DescScore = 5
	}

	// 5. RefNo Score (5pts max): reference匹配
	bankRef := strValueOf(bankTxn.ReferenceNo)
	if bankRef != "" && bankRef == glEntry.VoucherID.String() {
		score.RefNoScore = 5
	}

	// Calculate total
	score.TotalScore = score.AmountScore + score.DateScore + score.NameScore + score.DescScore + score.RefNoScore
	score.IsAutoMatched = score.TotalScore >= 85
	score.NeedConfirm = score.TotalScore >= 60 && score.TotalScore < 85

	return score
}

// strValueOf safely extracts string value from *string
func strValueOf(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// levenshteinSimilarity computes normalized similarity between two strings (0-1)
func levenshteinSimilarity(a, b string) float64 {
	if a == "" && b == "" {
		return 1
	}
	if a == "" || b == "" {
		return 0
	}
	// Simple Levenshtein distance
	alen, blen := len(a), len(b)
	// Use a simplified version for Chinese/short strings
	dist := levenshteinDistance(a, b)
	maxLen := alen
	if blen > maxLen {
		maxLen = blen
	}
	return 1 - float64(dist)/float64(maxLen)
}

// levenshteinDistance computes edit distance between two strings
func levenshteinDistance(a, b string) int {
	alen, blen := len(a), len(b)
	dp := make([][]int, alen+1)
	for i := range dp {
		dp[i] = make([]int, blen+1)
		dp[i][0] = i
	}
	for j := 0; j <= blen; j++ {
		dp[0][j] = j
	}
	for i := 1; i <= alen; i++ {
		for j := 1; j <= blen; j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			dp[i][j] = minInt(dp[i-1][j]+1, dp[i][j-1]+1, dp[i-1][j-1]+cost)
		}
	}
	return dp[alen][blen]
}

func minInt(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// containsKeywords counts how many keywords appear in text
func containsKeywords(text string, keywords []string) int {
	count := 0
	lowerText := strings.ToLower(text)
	for _, kw := range keywords {
		if strings.Contains(lowerText, strings.ToLower(kw)) {
			count++
		}
	}
	return count
}

// ReconcileBankAccount performs bank reconciliation for a bank account in a specific period.
// Uses a 5-dimension scoring system: ≥85 auto-matched, 60-85 pending confirm, <60 goes to reconciliation items.
func (s *BankReconciliationService) ReconcileBankAccount(ctx context.Context, tenantID, bankAccountID uuid.UUID, periodNo int) (*ReconciliationResult, error) {
	// Convert periodNo to start/end dates (YYYYMM -> dates)
	startDate, endDate := periodToDateRange(periodNo)

	// 1. Get all unmatched bank transactions in the period
	bankTxns, err := s.bankTxnRepo.GetUnmatched(ctx, tenantID, bankAccountID)
	if err != nil {
		return nil, fmt.Errorf("get bank transactions: %w", err)
	}

	// Filter by date range
	var filteredBankTxns []model.BankTransaction
	for _, txn := range bankTxns {
		if !txn.TxnDate.Before(startDate) && !txn.TxnDate.After(endDate) {
			filteredBankTxns = append(filteredBankTxns, txn)
		}
	}
	bankTxns = filteredBankTxns

	// 2. Get all GL entries for the bank account in the period
	bankAccount, err := s.bankRepo.GetByID(ctx, tenantID, bankAccountID)
	if err != nil {
		return nil, fmt.Errorf("get bank account: %w", err)
	}

	if bankAccount.ClearingAccountID == nil {
		return nil, fmt.Errorf("bank account %s has no clearing account configured", bankAccountID)
	}
	glEntries, err := s.glEntryRepo.GetByAccountAndPeriod(ctx, tenantID, *bankAccount.ClearingAccountID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("get GL entries: %w", err)
	}

	// 3. Calculate bank balance and book balance (all transactions)
	bankBalance := decimal.Zero
	for _, txn := range bankTxns {
		bankBalance = bankBalance.Add(txn.Debit).Sub(txn.Credit)
	}

	bookBalance := decimal.Zero
	for _, entry := range glEntries {
		bookBalance = bookBalance.Add(entry.Debit).Sub(entry.Credit)
	}

	// 4. Score-based matching: find best GL entry candidate for each bank txn
	// Candidates must be within date±3 days and amount±1%
	matchedBank := make(map[uuid.UUID]bool)
	matchedGL := make(map[uuid.UUID]bool)
	autoMatchedCount := 0
	var pendingItems []PendingConfirmItem
	var bankOnlyItems []UnreconciledItem
	var bookOnlyItems []UnreconciledItem
	var totalMatched decimal.Decimal
	var matchItems []MatchItemResult

	// For each bank txn, find the best scoring GL entry
	for _, txn := range bankTxns {
		if matchedBank[txn.ID] {
			continue
		}

		bankAmt := txn.Debit.Sub(txn.Credit).Abs()
		if bankAmt.IsZero() {
			continue
		}

		var bestScore *MatchScore
		var bestGLEntry *model.GLEntry

		// Find candidates: date within ±3 days, amount within ±1%
		for i := range glEntries {
			entry := &glEntries[i]
			if matchedGL[entry.ID] {
				continue
			}

			glAmt := entry.Debit.Sub(entry.Credit).Abs()
			if glAmt.IsZero() {
				continue
			}

			// Date filter: ±3 days
			daysDiff := math.Abs(txn.TxnDate.Sub(entry.PostingDate).Hours() / 24)
			if daysDiff > 3 {
				continue
			}

			// Amount filter: ±1%
			if !bankAmt.IsZero() {
				diff := bankAmt.Sub(glAmt).Abs()
				pct := diff.Div(bankAmt).InexactFloat64()
				if pct > 0.01 {
					continue
				}
			}

			// Calculate score
			score := CalculateMatchScore(&txn, entry)
			if bestScore == nil || score.TotalScore > bestScore.TotalScore {
				bestScore = &score
				bestGLEntry = entry
			}
		}

		if bestScore != nil && bestGLEntry != nil {
			if bestScore.IsAutoMatched {
				// Auto-match: mark as matched
				matchedBank[txn.ID] = true
				matchedGL[bestGLEntry.ID] = true
				autoMatchedCount++
				// Accumulate matched amount, not score
				matchedAmt := bestGLEntry.Debit.Sub(bestGLEntry.Credit).Abs()
				totalMatched = totalMatched.Add(matchedAmt)
				// Update bank txn matched_gl_entry_id
				s.bankTxnRepo.UpdateMatchedGLEntryID(ctx, txn.ID, bestGLEntry.ID)
				matchItems = append(matchItems, MatchItemResult{
					ID:          txn.ID,
					Score:       bestScore.TotalScore,
					BankTxnDesc: strValueOf(txn.Description),
					GLEntryDesc: "", // Would need GL entry description lookup
					NeedConfirm: false,
				})
			} else if bestScore.NeedConfirm {
				// Pending confirmation
				pendingItems = append(pendingItems, PendingConfirmItem{
					BankTxnID:   txn.ID,
					BankTxnDesc: strValueOf(txn.Description),
					BankTxnAmt:  bankAmt,
					GLEntryID:   bestGLEntry.ID,
					GLEntryDesc: "", // Would need party lookup
					GLEntryAmt:  bestGLEntry.Debit.Sub(bestGLEntry.Credit).Abs(),
					Score:       *bestScore,
				})
				matchItems = append(matchItems, MatchItemResult{
					ID:          txn.ID,
					Score:       bestScore.TotalScore,
					BankTxnDesc: strValueOf(txn.Description),
					GLEntryDesc: "",
					NeedConfirm: true,
				})
			} else {
				// Score < 60: goes to unreconciled items (bank only)
				dir := "debit"
				if txn.Credit.GreaterThan(txn.Debit) {
					dir = "credit"
				}
				itemType := ItemTypeBankPaymentNotInGL
				if dir == "credit" {
					itemType = ItemTypeBankReceiptNotInGL
				}
				bankOnlyItems = append(bankOnlyItems, UnreconciledItem{
					ID:          txn.ID,
					SourceType:  "bank_transaction",
					TxnDate:     txn.TxnDate,
					Description: strValueOf(txn.Description),
					Amount:      bankAmt,
					Direction:   dir,
					ItemType:    itemType,
				})
				matchItems = append(matchItems, MatchItemResult{
					ID:          txn.ID,
					Score:       bestScore.TotalScore,
					BankTxnDesc: strValueOf(txn.Description),
					GLEntryDesc: "",
					NeedConfirm: false,
				})
			}
		} else {
			// No GL entry found → bank only
			dir := "debit"
			if txn.Credit.GreaterThan(txn.Debit) {
				dir = "credit"
			}
			itemType := ItemTypeBankPaymentNotInGL
			if dir == "credit" {
				itemType = ItemTypeBankReceiptNotInGL
			}
			bankOnlyItems = append(bankOnlyItems, UnreconciledItem{
				ID:          txn.ID,
				SourceType:  "bank_transaction",
				TxnDate:     txn.TxnDate,
				Description: strValueOf(txn.Description),
				Amount:      bankAmt,
				Direction:   dir,
				ItemType:    itemType,
			})
			matchItems = append(matchItems, MatchItemResult{
				ID:          txn.ID,
				Score:       0,
				BankTxnDesc: strValueOf(txn.Description),
				GLEntryDesc: "",
				NeedConfirm: false,
			})
		}
	}

	// Find GL entries without matching bank txns → book only
	for _, entry := range glEntries {
		if matchedGL[entry.ID] {
			continue
		}
		amt := entry.Debit.Sub(entry.Credit).Abs()
		if amt.IsZero() {
			continue
		}
		dir := "debit"
		if entry.Credit.GreaterThan(entry.Debit) {
			dir = "credit"
		}
		itemType := ItemTypeGLPaymentNotInBank
		if dir == "credit" {
			itemType = ItemTypeGLReceiptNotInBank
		}
		bookOnlyItems = append(bookOnlyItems, UnreconciledItem{
			ID:          entry.ID,
			SourceType:  "gl_entry",
			TxnDate:     entry.PostingDate,
			Description: "", // Would need GL description lookup
			Amount:      amt,
			Direction:   dir,
			ItemType:    itemType,
		})
	}

	// Calculate adjusted balance
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

	matchedCount := autoMatchedCount + len(pendingItems)

	return &ReconciliationResult{
		BankBalance:     bankBalance,
		BookBalance:     bookBalance,
		AdjustedBalance: adjustedBalance,
		BankOnlyItems:   bankOnlyItems,
		BookOnlyItems:   bookOnlyItems,
		MatchedCount:    matchedCount,
		TotalMatched:    totalMatched,
		PendingItems:     pendingItems,
		AutoMatchedCount: autoMatchedCount,
		MatchItems:       matchItems,
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
		ID:              record.ID,
		BankAccountID:   bankAccountID,
		PeriodNo:        periodNo,
		BankBalance:     record.BankBalance,
		BookBalance:     record.BookBalance,
		AdjustedBalance: record.AdjustedBalance,
		Status:          record.Status,
		ReconciledBy:    record.ReconciledBy,
		ReconciledAt:    record.ReconciledAt,
	}

	// Get unreconciled items
	unreconItems, err := s.getUnreconciledItems(ctx, record.ID)
	if err != nil {
		return nil, fmt.Errorf("get unreconciled items: %w", err)
	}

	// Split into bank_only and book_only items (legacy mapping)
	var bankOnlyItems, bookOnlyItems []model.UnreconciledItem
	for _, item := range unreconItems {
		if item.ItemType == "bank_only" || item.ItemType == "bank_receipt_not_in_gl" || item.ItemType == "bank_payment_not_in_gl" {
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

// LockReconciliation locks a reconciliation record for exclusive access.
func (s *BankReconciliationService) LockReconciliation(ctx context.Context, tenantID, bankAccountID uuid.UUID, periodNo int, userID uuid.UUID) error {
	record, err := s.getReconciliationRecord(ctx, tenantID, bankAccountID, periodNo)
	if err != nil {
		// If no record exists, create one
		record = &model.ReconciliationRecord{
			ID:            uuid.New(),
			TenantID:      tenantID,
			BankAccountID: bankAccountID,
			PeriodNo:      periodNo,
			Status:        "in_progress",
		}
	}

	if record.Locked && record.LockedBy != nil && *record.LockedBy != userID {
		return fmt.Errorf("reconciliation already locked by another user")
	}

	now := time.Now()
	record.Locked = true
	record.LockedBy = &userID
	record.LockedAt = &now

	if record.ID == uuid.Nil {
		return s.insertReconciliationRecord(ctx, record)
	}
	return s.updateReconciliationRecord(ctx, record)
}

// UnlockReconciliation unlocks a reconciliation record.
func (s *BankReconciliationService) UnlockReconciliation(ctx context.Context, tenantID, bankAccountID uuid.UUID, periodNo int, userID uuid.UUID) error {
	record, err := s.getReconciliationRecord(ctx, tenantID, bankAccountID, periodNo)
	if err != nil {
		return fmt.Errorf("get reconciliation record: %w", err)
	}

	if !record.Locked {
		return nil // already unlocked
	}

	if record.LockedBy != nil && *record.LockedBy != userID {
		return fmt.Errorf("cannot unlock: locked by different user")
	}

	now := time.Now()
	record.Locked = false
	record.LockedBy = nil
	record.LockedAt = nil
	record.UnlockApprovedBy = &userID
	record.UnlockApprovedAt = &now

	return s.updateReconciliationRecord(ctx, record)
}

// IsLocked checks if a reconciliation record is locked.
func (s *BankReconciliationService) IsLocked(ctx context.Context, tenantID, bankAccountID uuid.UUID, periodNo int) (bool, error) {
	record, err := s.getReconciliationRecord(ctx, tenantID, bankAccountID, periodNo)
	if err != nil {
		return false, nil // not found = not locked
	}
	return record.Locked, nil
}

// GetPendingConfirmItems returns items pending manual confirmation.
func (s *BankReconciliationService) GetPendingConfirmItems(ctx context.Context, tenantID, bankAccountID uuid.UUID, periodNo int) ([]PendingConfirmItem, error) {
	record, err := s.getReconciliationRecord(ctx, tenantID, bankAccountID, periodNo)
	if err != nil {
		return nil, fmt.Errorf("get reconciliation record: %w", err)
	}

	// Query pending confirmation pairs from reconciliation_pairs table
	query := `
		SELECT rp.source_id, rp.target_id, rp.amount, rp.status, rp.match_level
		FROM reconciliation_pairs rp
		WHERE rp.tenant_id = $1 AND rp.status = 'pending'
		AND rp.created_at >= $2 AND rp.created_at <= $3
	`
	startDate, endDate := periodToDateRange(periodNo)

	rows, err := s.bankTxnRepo.GetPool().Query(ctx, query, tenantID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []PendingConfirmItem
	for rows.Next() {
		// This is a simplified placeholder - actual implementation would need
		// to properly query and join with bank_transactions and gl_entries
		_ = record // avoid unused
	}

	return items, nil
}

// ConfirmMatch manually confirms a match between bank txn and GL entry.
func (s *BankReconciliationService) ConfirmMatch(ctx context.Context, tenantID, bankTxnID, glEntryID uuid.UUID, userID uuid.UUID) error {
	// Update bank transaction matched_gl_entry_id
	err := s.bankTxnRepo.UpdateMatchedGLEntryID(ctx, bankTxnID, glEntryID)
	if err != nil {
		return fmt.Errorf("update matched gl entry: %w", err)
	}

	// Delete from pending confirmation if exists
	query := `DELETE FROM reconciliation_pairs WHERE source_id = $1 AND target_id = $2 AND tenant_id = $3`
	_, err = s.bankTxnRepo.GetPool().Exec(ctx, query, bankTxnID, glEntryID, tenantID)
	if err != nil {
		return fmt.Errorf("delete pending pair: %w", err)
	}

	return nil
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
	query := `
		SELECT id, tenant_id, bank_account_id, period_no, bank_balance, book_balance,
		       adjusted_balance, bank_only_total, book_only_total, status,
		       reconciled_by, reconciled_at, locked, locked_by, locked_at,
		       unlock_approved_by, unlock_approved_at, created_at, updated_at
		FROM reconciliation_records
		WHERE tenant_id = $1 AND bank_account_id = $2 AND period_no = $3`

	record := &model.ReconciliationRecord{}
	err := s.bankTxnRepo.GetPool().QueryRow(ctx, query, tenantID, bankAccountID, periodNo).Scan(
		&record.ID, &record.TenantID, &record.BankAccountID, &record.PeriodNo,
		&record.BankBalance, &record.BookBalance, &record.AdjustedBalance,
		&record.BankOnlyTotal, &record.BookOnlyTotal, &record.Status,
		&record.ReconciledBy, &record.ReconciledAt, &record.Locked, &record.LockedBy, &record.LockedAt,
		&record.UnlockApprovedBy, &record.UnlockApprovedAt, &record.CreatedAt, &record.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get reconciliation record: %w", err)
	}
	return record, nil
}

func (s *BankReconciliationService) insertReconciliationRecord(ctx context.Context, record *model.ReconciliationRecord) error {
	query := `
		INSERT INTO reconciliation_records (id, tenant_id, bank_account_id, period_no,
			bank_balance, book_balance, adjusted_balance, bank_only_total, book_only_total, status,
			locked, locked_by, locked_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW(), NOW())`

	_, err := s.bankTxnRepo.GetPool().Exec(ctx, query,
		record.ID, record.TenantID, record.BankAccountID, record.PeriodNo,
		record.BankBalance, record.BookBalance, record.AdjustedBalance,
		record.BankOnlyTotal, record.BookOnlyTotal, record.Status,
		record.Locked, record.LockedBy, record.LockedAt,
	)
	return err
}

func (s *BankReconciliationService) updateReconciliationRecord(ctx context.Context, record *model.ReconciliationRecord) error {
	query := `
		UPDATE reconciliation_records
		SET bank_balance = $4, book_balance = $5, adjusted_balance = $6,
		    bank_only_total = $7, book_only_total = $8, status = $9,
		    reconciled_by = $10, reconciled_at = $11,
		    locked = $12, locked_by = $13, locked_at = $14,
		    unlock_approved_by = $15, unlock_approved_at = $16,
		    updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2`

	_, err := s.bankTxnRepo.GetPool().Exec(ctx, query,
		record.ID, record.TenantID, record.BankBalance, record.BookBalance,
		record.AdjustedBalance, record.BankOnlyTotal, record.BookOnlyTotal,
		record.Status, record.ReconciledBy, record.ReconciledAt,
		record.Locked, record.LockedBy, record.LockedAt,
		record.UnlockApprovedBy, record.UnlockApprovedAt,
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
// Writes unreconciled items with four-direction item_type classification.
func (s *BankReconciliationService) SaveReconciliationResult(ctx context.Context, tenantID, bankAccountID uuid.UUID, periodNo int, result *ReconciliationResult) error {
	// Check if record exists
	existing, err := s.getReconciliationRecord(ctx, tenantID, bankAccountID, periodNo)
	if err != nil {
		existing = nil
	}

	var recordID uuid.UUID
	var bankOnlyTotal, bookOnlyTotal decimal.Decimal
	for _, item := range result.BankOnlyItems {
		bankOnlyTotal = bankOnlyTotal.Add(item.Amount)
	}
	for _, item := range result.BookOnlyItems {
		bookOnlyTotal = bookOnlyTotal.Add(item.Amount)
	}

	if existing != nil {
		recordID = existing.ID
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
		recordID = uuid.New()
		// Insert
		query := `
			INSERT INTO reconciliation_records (id, tenant_id, bank_account_id, period_no,
				bank_balance, book_balance, adjusted_balance, bank_only_total, book_only_total, status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'in_progress')
			ON CONFLICT (tenant_id, bank_account_id, period_no) DO UPDATE
			SET bank_balance = $5, book_balance = $6, adjusted_balance = $7,
			    bank_only_total = $8, book_only_total = $9, status = 'in_progress', updated_at = NOW()`
		_, err := s.bankTxnRepo.GetPool().Exec(ctx, query,
			recordID, tenantID, bankAccountID, periodNo,
			result.BankBalance, result.BookBalance, result.AdjustedBalance,
			bankOnlyTotal, bookOnlyTotal,
		)
		if err != nil {
			return fmt.Errorf("insert reconciliation record: %w", err)
		}
	}

	// Delete existing unreconciled_items for this record
	_, err = s.bankTxnRepo.GetPool().Exec(ctx,
		`DELETE FROM unreconciled_items WHERE reconciliation_record_id = $1`,
		recordID)
	if err != nil {
		return fmt.Errorf("delete existing unreconciled items: %w", err)
	}

	// Insert bank_only items with four-direction type
	for _, item := range result.BankOnlyItems {
		debit := decimal.Zero
		credit := decimal.Zero
		if item.Direction == "debit" {
			debit = item.Amount
		} else {
			credit = item.Amount
		}
		desc := item.Description
		query := `
			INSERT INTO unreconciled_items (id, reconciliation_record_id, item_type, source_type, source_id,
				txn_date, description, debit, credit, amount, tenant_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
		_, err := s.bankTxnRepo.GetPool().Exec(ctx, query,
			uuid.New(), recordID, item.ItemType, item.SourceType, item.ID,
			item.TxnDate, &desc, debit, credit, item.Amount, tenantID,
		)
		if err != nil {
			return fmt.Errorf("insert bank only item: %w", err)
		}
	}

	// Insert book_only items with four-direction type
	for _, item := range result.BookOnlyItems {
		debit := decimal.Zero
		credit := decimal.Zero
		if item.Direction == "debit" {
			debit = item.Amount
		} else {
			credit = item.Amount
		}
		desc := item.Description
		query := `
			INSERT INTO unreconciled_items (id, reconciliation_record_id, item_type, source_type, source_id,
				txn_date, description, debit, credit, amount, tenant_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
		_, err := s.bankTxnRepo.GetPool().Exec(ctx, query,
			uuid.New(), recordID, item.ItemType, item.SourceType, item.ID,
			item.TxnDate, &desc, debit, credit, item.Amount, tenantID,
		)
		if err != nil {
			return fmt.Errorf("insert book only item: %w", err)
		}
	}

	return nil
}

// GetReconciliationItems returns unreconciled items grouped by type for a period.
func (s *BankReconciliationService) GetReconciliationItems(ctx context.Context, tenantID, bankAccountID uuid.UUID, periodNo int) ([]UnreconciledItem, error) {
	record, err := s.getReconciliationRecord(ctx, tenantID, bankAccountID, periodNo)
	if err != nil {
		return nil, fmt.Errorf("get reconciliation record: %w", err)
	}

	items, err := s.getUnreconciledItems(ctx, record.ID)
	if err != nil {
		return nil, err
	}

	result := make([]UnreconciledItem, 0, len(items))
	for _, item := range items {
		dir := "credit"
		if item.Debit.GreaterThan(item.Credit) {
			dir = "debit"
		}
		desc := ""
		if item.Description != nil {
			desc = *item.Description
		}
		result = append(result, UnreconciledItem{
			ID:          item.ID,
			SourceType:  item.SourceType,
			TxnDate:     item.TxnDate,
			Description: desc,
			Amount:      item.Amount,
			Direction:   dir,
			ItemType:    item.ItemType,
		})
	}
	return result, nil
}

// BalanceCheckItem and BalanceCheckResult (existing code preserved)
type BalanceCheckItem struct {
	BankAccountID    uuid.UUID       `json:"bank_account_id"`
	BankName         string          `json:"bank_name"`
	OpeningBalance   decimal.Decimal `json:"opening_balance"`
	TxnInflowTotal   decimal.Decimal `json:"txn_inflow_total"`
	TxnOutflowTotal  decimal.Decimal `json:"txn_outflow_total"`
	ExpectedBalance  decimal.Decimal `json:"expected_balance"`
	StoredBalance    decimal.Decimal `json:"stored_balance"`
	Difference       decimal.Decimal `json:"difference"`
	TxnCount         int             `json:"txn_count"`
	HasInconsistency bool            `json:"has_inconsistency"`
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