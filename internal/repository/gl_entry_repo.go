package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
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

// GetPool returns the underlying connection pool.
func (r *GLEntryRepository) GetPool() *pgxpool.Pool {
	return r.pool
}