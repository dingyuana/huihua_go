package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"huihua/finance/internal/repository"
)

// DashboardHandler handles dashboard aggregation HTTP requests.
type DashboardHandler struct {
	journalRepo *repository.JournalRepository
	glEntryRepo *repository.GLEntryRepository
	bankTxnRepo *repository.BankTransactionRepository
	periodRepo  *repository.PeriodRepository
}

// DashboardStats holds the key metrics for the dashboard overview.
type DashboardStats struct {
	MonthlyTxns     int64  `json:"monthly_txns"`
	PendingVouchers int64  `json:"pending_vouchers"`
	MonthlyIncome   int64  `json:"monthly_income"`
	MonthlyExpense  int64  `json:"monthly_expense"`
	MonthlyProfit   int64  `json:"monthly_profit"`
	PendingTxns     int64  `json:"pending_txns"`
	PeriodStatus    string `json:"period_status"`
}

// NewDashboardHandler creates a new DashboardHandler.
func NewDashboardHandler(
	journalRepo *repository.JournalRepository,
	glEntryRepo *repository.GLEntryRepository,
	bankTxnRepo *repository.BankTransactionRepository,
	periodRepo *repository.PeriodRepository,
) *DashboardHandler {
	return &DashboardHandler{
		journalRepo: journalRepo,
		glEntryRepo: glEntryRepo,
		bankTxnRepo: bankTxnRepo,
		periodRepo:  periodRepo,
	}
}

// GetStats returns an aggregation of key dashboard metrics for the current month.
func (h *DashboardHandler) GetStats(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	ctx := c.UserContext()

	now := time.Now()
	year, month, _ := now.Date()
	loc := now.Location()
	startOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, loc)
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)
	periodStr := startOfMonth.Format("2006-01")

	// 1. Current open period status
	periodStatus := "open"
	if currentPeriod, err := h.periodRepo.GetCurrentOpen(ctx, tenantID); err == nil && currentPeriod != nil {
		periodStatus = currentPeriod.Status
	}

	// 2. Pending vouchers (docstatus < 1, draft or rejected)
	pendingVouchers, _ := h.journalRepo.CountUnpostedByPeriod(ctx, tenantID, periodStr)

	// 3. Monthly income / expense / profit via GL entries
	income, expense, _ := h.glEntryRepo.GetIncomeExpenseSummary(ctx, tenantID, startOfMonth, endOfMonth)
	monthlyIncome := income.IntPart()
	monthlyExpense := expense.IntPart()
	monthlyProfit := income.Sub(expense).IntPart()

	// 4. Monthly transaction count (all bank accounts, current month)
	monthlyTxns, _ := h.countBankTxnsByPeriod(ctx, tenantID, startOfMonth, endOfMonth)

	// 5. Pending bank transactions (classification = 'pending', across all accounts)
	pendingTxns, _ := h.countPendingBankTxns(ctx, tenantID)

	return c.JSON(DashboardStats{
		MonthlyTxns:     monthlyTxns,
		PendingVouchers: int64(pendingVouchers),
		MonthlyIncome:   monthlyIncome,
		MonthlyExpense:  monthlyExpense,
		MonthlyProfit:   monthlyProfit,
		PendingTxns:     pendingTxns,
		PeriodStatus:    periodStatus,
	})
}

// countBankTxnsByPeriod counts all bank transactions for a tenant within a date range.
func (h *DashboardHandler) countBankTxnsByPeriod(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (int64, error) {
	var count int64
	err := h.bankTxnRepo.GetPool().QueryRow(ctx, `
		SELECT COUNT(*)
		FROM bank_transactions
		WHERE tenant_id = $1 AND txn_date >= $2 AND txn_date <= $3
	`, tenantID, startDate, endDate).Scan(&count)
	return count, err
}

// countPendingBankTxns counts all pending (unclassified) bank transactions across all accounts.
func (h *DashboardHandler) countPendingBankTxns(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	var count int64
	err := h.bankTxnRepo.GetPool().QueryRow(ctx, `
		SELECT COUNT(*)
		FROM bank_transactions
		WHERE tenant_id = $1 AND classification = 'pending'
	`, tenantID).Scan(&count)
	return count, err
}