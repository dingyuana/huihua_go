package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/model"
)

// GLEntryRepository provides data access for the gl_entries table.
type GLEntryRepository struct {
	pool *pgxpool.Pool
}

// NewGLEntryRepository creates a new GLEntryRepository.
func NewGLEntryRepository(pool *pgxpool.Pool) *GLEntryRepository {
	return &GLEntryRepository{pool: pool}
}

// WriteGLEntries writes GL entries from journal entry lines (called on voucher submit).
func (r *GLEntryRepository) WriteGLEntries(ctx context.Context, tenantID uuid.UUID, journalEntryID uuid.UUID, lines []model.JournalEntryLine, postingDate time.Time, voucherType *string, companyID uuid.UUID) error {
	if len(lines) == 0 {
		return nil
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, line := range lines {
		// Determine debit/credit in account currency
		var debitCcy, creditCcy decimal.Decimal
		if line.Debit.GreaterThan(decimal.Zero) {
			debitCcy = line.DebitCcy
		} else {
			creditCcy = line.CreditCcy
		}

		glEntry := &model.GLEntry{
			ID:           uuid.New(),
			AccountID:    line.AccountID,
			PostingDate:  postingDate,
			Debit:        line.Debit,
			Credit:       line.Credit,
			DebitCcy:     debitCcy,
			CreditCcy:    creditCcy,
			VoucherType:  voucherType,
			VoucherID:    &journalEntryID,
			PartyType:    line.PartyType,
			PartyID:      line.PartyID,
			CostCenterID: line.CostCenterID,
			ProjectID:    line.ProjectID,
			CompanyID:    companyID,
			TenantID:     tenantID,
			IsCancelled:  false,
			CreatedAt:    time.Now(),
		}

		_, err := tx.Exec(ctx, `
			INSERT INTO gl_entries (id, account_id, posting_date, debit, credit, debit_ccy, credit_ccy,
				voucher_type, voucher_id, party_type, party_id, cost_center_id, project_id,
				company_id, tenant_id, is_cancelled, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`,
			glEntry.ID, glEntry.AccountID, glEntry.PostingDate, glEntry.Debit, glEntry.Credit,
			glEntry.DebitCcy, glEntry.CreditCcy, glEntry.VoucherType, glEntry.VoucherID,
			glEntry.PartyType, glEntry.PartyID, glEntry.CostCenterID, glEntry.ProjectID,
			glEntry.CompanyID, glEntry.TenantID, glEntry.IsCancelled, glEntry.CreatedAt)
		if err != nil {
			return fmt.Errorf("insert gl entry for account %s: %w", line.AccountID.String(), err)
		}
	}

	return tx.Commit(ctx)
}

// CancelGLEntriesByVoucher marks all GL entries for a voucher as cancelled (called on voucher reverse/cancel).
func (r *GLEntryRepository) CancelGLEntriesByVoucher(ctx context.Context, tenantID uuid.UUID, voucherID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE gl_entries SET is_cancelled = true WHERE voucher_id = $1 AND tenant_id = $2`,
		voucherID, tenantID)
	return err
}

// WriteGLEntriesTx writes GL entries within an existing transaction.
func (r *GLEntryRepository) WriteGLEntriesTx(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID, journalEntryID uuid.UUID, lines []model.JournalEntryLine, postingDate time.Time, voucherType *string, companyID uuid.UUID) error {
	if len(lines) == 0 {
		return nil
	}

	for _, line := range lines {
		var debitCcy, creditCcy decimal.Decimal
		if line.Debit.GreaterThan(decimal.Zero) {
			debitCcy = line.DebitCcy
		} else {
			creditCcy = line.CreditCcy
		}

		glEntry := &model.GLEntry{
			ID:           uuid.New(),
			AccountID:    line.AccountID,
			PostingDate:  postingDate,
			Debit:        line.Debit,
			Credit:       line.Credit,
			DebitCcy:     debitCcy,
			CreditCcy:    creditCcy,
			VoucherType:  voucherType,
			VoucherID:    &journalEntryID,
			PartyType:    line.PartyType,
			PartyID:      line.PartyID,
			CostCenterID: line.CostCenterID,
			ProjectID:    line.ProjectID,
			CompanyID:    companyID,
			TenantID:     tenantID,
			IsCancelled:  false,
			CreatedAt:    time.Now(),
		}

		_, err := tx.Exec(ctx, `
			INSERT INTO gl_entries (id, account_id, posting_date, debit, credit, debit_ccy, credit_ccy,
				voucher_type, voucher_id, party_type, party_id, cost_center_id, project_id,
				company_id, tenant_id, is_cancelled, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`,
			glEntry.ID, glEntry.AccountID, glEntry.PostingDate, glEntry.Debit, glEntry.Credit,
			glEntry.DebitCcy, glEntry.CreditCcy, glEntry.VoucherType, glEntry.VoucherID,
			glEntry.PartyType, glEntry.PartyID, glEntry.CostCenterID, glEntry.ProjectID,
			glEntry.CompanyID, glEntry.TenantID, glEntry.IsCancelled, glEntry.CreatedAt)
		if err != nil {
			return fmt.Errorf("insert gl entry for account %s: %w", line.AccountID.String(), err)
		}
	}

	return nil
}

// CancelGLEntriesByVoucherTx marks all GL entries for a voucher as cancelled within an existing transaction.
func (r *GLEntryRepository) CancelGLEntriesByVoucherTx(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID, voucherID uuid.UUID) error {
	_, err := tx.Exec(ctx, `
		UPDATE gl_entries SET is_cancelled = true WHERE voucher_id = $1 AND tenant_id = $2`,
		voucherID, tenantID)
	return err
}

// GetByAccountAndPeriod retrieves GL entries for an account within a date range.
func (r *GLEntryRepository) GetByAccountAndPeriod(ctx context.Context, tenantID, accountID uuid.UUID, startDate, endDate time.Time) ([]model.GLEntry, error) {
	query := `
		SELECT id, account_id, posting_date, debit, credit, debit_ccy, credit_ccy,
		       account_ccy, voucher_type, voucher_id, against_voucher_type, against_voucher_id,
		       party_type, party_id, cost_center_id, project_id, company_id, tenant_id,
		       is_cancelled, created_at
		FROM gl_entries
		WHERE tenant_id = $1 AND account_id = $2 AND posting_date >= $3 AND posting_date <= $4
		ORDER BY posting_date, created_at`

	rows, err := r.pool.Query(ctx, query, tenantID, accountID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []model.GLEntry
	for rows.Next() {
		var e model.GLEntry
		err := rows.Scan(
			&e.ID, &e.AccountID, &e.PostingDate, &e.Debit, &e.Credit,
			&e.DebitCcy, &e.CreditCcy, &e.AccountCcy, &e.VoucherType, &e.VoucherID,
			&e.AgainstVoucherType, &e.AgainstVoucherID, &e.PartyType, &e.PartyID,
			&e.CostCenterID, &e.ProjectID, &e.CompanyID, &e.TenantID,
			&e.IsCancelled, &e.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// GetByTenantInRange retrieves all GL entries for a tenant within a date range.
func (r *GLEntryRepository) GetByTenantInRange(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]model.GLEntry, error) {
	query := `
		SELECT id, account_id, posting_date, debit, credit, debit_ccy, credit_ccy,
		       account_ccy, voucher_type, voucher_id, against_voucher_type, against_voucher_id,
		       party_type, party_id, cost_center_id, project_id, company_id, tenant_id,
		       is_cancelled, created_at
		FROM gl_entries
		WHERE tenant_id = $1 AND posting_date >= $2 AND posting_date <= $3
		ORDER BY account_id, posting_date, created_at`

	rows, err := r.pool.Query(ctx, query, tenantID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []model.GLEntry
	for rows.Next() {
		var e model.GLEntry
		err := rows.Scan(
			&e.ID, &e.AccountID, &e.PostingDate, &e.Debit, &e.Credit,
			&e.DebitCcy, &e.CreditCcy, &e.AccountCcy, &e.VoucherType, &e.VoucherID,
			&e.AgainstVoucherType, &e.AgainstVoucherID, &e.PartyType, &e.PartyID,
			&e.CostCenterID, &e.ProjectID, &e.CompanyID, &e.TenantID,
			&e.IsCancelled, &e.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// AccountBalance represents an aggregated balance for a single account.
type AccountBalance struct {
	AccountID uuid.UUID       `json:"account_id"`
	Code      string          `json:"code"`
	Name      string          `json:"name"`
	RootType  string          `json:"root_type"`
	Debit     decimal.Decimal `json:"debit"`
	Credit    decimal.Decimal `json:"credit"`
}

// GetBalancesByPeriod returns aggregated debit/credit per account for a date range.
func (r *GLEntryRepository) GetBalancesByPeriod(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]AccountBalance, error) {
	query := `
		SELECT g.account_id, a.code, a.name, COALESCE(a.root_type, '') as root_type,
		       COALESCE(SUM(g.debit), 0) as total_debit,
		       COALESCE(SUM(g.credit), 0) as total_credit
		FROM gl_entries g
		JOIN accounts a ON g.account_id = a.id AND a.tenant_id = $1
		WHERE g.tenant_id = $1 AND g.posting_date >= $2 AND g.posting_date <= $3 AND g.is_cancelled = false
		GROUP BY g.account_id, a.code, a.name, a.root_type
		ORDER BY a.code`
	rows, err := r.pool.Query(ctx, query, tenantID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("get balances by period: %w", err)
	}
	defer rows.Close()

	var balances []AccountBalance
	for rows.Next() {
		var b AccountBalance
		if err := rows.Scan(&b.AccountID, &b.Code, &b.Name, &b.RootType, &b.Debit, &b.Credit); err != nil {
			return nil, fmt.Errorf("scan balance: %w", err)
		}
		balances = append(balances, b)
	}
	return balances, rows.Err()
}

// GetIncomeExpenseSummary returns total debit/credit for income and expense accounts within a date range.
func (r *GLEntryRepository) GetIncomeExpenseSummary(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (incomeCredit, expenseDebit decimal.Decimal, err error) {
	query := `
		SELECT
			COALESCE(SUM(CASE WHEN a.root_type = 'income' THEN g.credit ELSE 0 END), 0) as total_income,
			COALESCE(SUM(CASE WHEN a.root_type = 'expense' THEN g.debit ELSE 0 END), 0) as total_expense
		FROM gl_entries g
		JOIN accounts a ON g.account_id = a.id AND a.tenant_id = $1
		WHERE g.tenant_id = $1 AND g.posting_date >= $2 AND g.posting_date <= $3 AND g.is_cancelled = false`
	err = r.pool.QueryRow(ctx, query, tenantID, startDate, endDate).Scan(&incomeCredit, &expenseDebit)
	if err != nil {
		return decimal.Zero, decimal.Zero, fmt.Errorf("get income expense summary: %w", err)
	}
	return
}

// GetPool returns the underlying connection pool.
func (r *GLEntryRepository) GetPool() *pgxpool.Pool {
	return r.pool
}

// GetBankGLBalance returns the net debit-credit balance for a bank account (by account code pattern) in a date range.
// It sums all GL entries for accounts whose code starts with "1002" (bank accounts).
func (r *GLEntryRepository) GetBankGLBalance(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (decimal.Decimal, error) {
	query := `
		SELECT COALESCE(SUM(g.debit) - SUM(g.credit), 0) as net_balance
		FROM gl_entries g
		JOIN accounts a ON g.account_id = a.id AND a.tenant_id = $1
		WHERE g.tenant_id = $1 AND g.posting_date >= $2 AND g.posting_date <= $3
		  AND g.is_cancelled = false AND a.code LIKE '1002%'`
	var balance decimal.NullDecimal
	err := r.pool.QueryRow(ctx, query, tenantID, startDate, endDate).Scan(&balance)
	if err != nil {
		return decimal.Zero, err
	}
	return balance.Decimal, nil
}