package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"huihua/finance/internal/model"
)

// JournalRepository provides data access for the journal_entries and journal_entry_lines tables.
type JournalRepository struct {
	pool *pgxpool.Pool
}

// NewJournalRepository creates a new JournalRepository.
func NewJournalRepository(pool *pgxpool.Pool) *JournalRepository {
	return &JournalRepository{pool: pool}
}

// Create inserts a new journal entry and returns it.
func (r *JournalRepository) Create(ctx context.Context, tenantID uuid.UUID, je *model.JournalEntry) (*model.JournalEntry, error) {
	query := `
		INSERT INTO journal_entries (id, voucher_no, voucher_type, posting_date, company_id, tenant_id, remark, docstatus, reversed_id, reversal_id, submitted_by, submitted_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING created_at, updated_at`

	if je.ID == uuid.Nil {
		je.ID = uuid.New()
	}

	err := r.pool.QueryRow(ctx, query,
		je.ID, je.VoucherNo, je.VoucherType, je.PostingDate, je.CompanyID,
		tenantID, je.Remark, je.DocStatus, je.ReversedID, je.ReversalID,
		je.SubmittedBy, je.SubmittedAt, je.CreatedBy,
	).Scan(&je.CreatedAt, &je.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create journal entry: %w", err)
	}

	je.TenantID = tenantID
	return je, nil
}

// GetByID retrieves a journal entry by its ID within the given tenant.
func (r *JournalRepository) GetByID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*model.JournalEntry, error) {
	query := `
		SELECT id, voucher_no, voucher_type, posting_date, company_id, tenant_id, remark,
		       docstatus, reversed_id, reversal_id, submitted_by, submitted_at, created_by,
		       created_at, updated_at
		FROM journal_entries
		WHERE id = $1 AND tenant_id = $2`

	je := &model.JournalEntry{}
	err := r.pool.QueryRow(ctx, query, id, tenantID).Scan(
		&je.ID, &je.VoucherNo, &je.VoucherType, &je.PostingDate, &je.CompanyID,
		&je.TenantID, &je.Remark, &je.DocStatus, &je.ReversedID, &je.ReversalID,
		&je.SubmittedBy, &je.SubmittedAt, &je.CreatedBy, &je.CreatedAt, &je.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get journal entry by id: %w", err)
	}
	return je, nil
}

// ListByTenant retrieves journal entries for the given tenant, ordered by posting_date DESC.
func (r *JournalRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]model.JournalEntry, error) {
	query := `
		SELECT id, voucher_no, voucher_type, posting_date, company_id, tenant_id, remark,
		       docstatus, reversed_id, reversal_id, submitted_by, submitted_at, created_by,
		       created_at, updated_at
		FROM journal_entries
		WHERE tenant_id = $1
		ORDER BY posting_date DESC, voucher_no DESC`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list journal entries by tenant: %w", err)
	}
	defer rows.Close()

	var entries []model.JournalEntry
	for rows.Next() {
		var je model.JournalEntry
		if err := rows.Scan(
			&je.ID, &je.VoucherNo, &je.VoucherType, &je.PostingDate, &je.CompanyID,
			&je.TenantID, &je.Remark, &je.DocStatus, &je.ReversedID, &je.ReversalID,
			&je.SubmittedBy, &je.SubmittedAt, &je.CreatedBy, &je.CreatedAt, &je.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan journal entry: %w", err)
		}
		entries = append(entries, je)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate journal entries: %w", err)
	}
	return entries, nil
}

// AddLines inserts multiple journal entry lines for a given journal entry.
func (r *JournalRepository) AddLines(ctx context.Context, tenantID uuid.UUID, journalEntryID uuid.UUID, lines []model.JournalEntryLine) ([]model.JournalEntryLine, error) {
	if len(lines) == 0 {
		return nil, nil
	}

	query := `
		INSERT INTO journal_entry_lines (id, journal_entry_id, account_id, debit, credit, debit_ccy, credit_ccy,
			account_ccy, exchange_rate, party_type, party_id, cost_center_id, project_id, user_remark, reconciled, tenant_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`

	result := make([]model.JournalEntryLine, 0, len(lines))
	for i := range lines {
		line := &lines[i]
		if line.ID == uuid.Nil {
			line.ID = uuid.New()
		}
		line.JournalEntryID = journalEntryID
		line.TenantID = tenantID

		_, err := r.pool.Exec(ctx, query,
			line.ID, line.JournalEntryID, line.AccountID, line.Debit, line.Credit,
			line.DebitCcy, line.CreditCcy, line.AccountCcy, line.ExchangeRate,
			line.PartyType, line.PartyID, line.CostCenterID, line.ProjectID,
			line.UserRemark, line.Reconciled, tenantID,
		)
		if err != nil {
			return nil, fmt.Errorf("add journal entry line %d: %w", i, err)
		}
		result = append(result, *line)
	}

	return result, nil
}
