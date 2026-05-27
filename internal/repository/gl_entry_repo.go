package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
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

// GetPool returns the underlying connection pool.
func (r *GLEntryRepository) GetPool() *pgxpool.Pool {
	return r.pool
}