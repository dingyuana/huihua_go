package repository

import (
	"context"
	"fmt"
	"time"

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

// UpdateStatus updates the docstatus of a journal entry and records the transition.
func (r *JournalRepository) UpdateStatus(ctx context.Context, tenantID uuid.UUID, journalID uuid.UUID, newStatus int16, changedBy uuid.UUID, changedByName *string, action model.VoucherAction, reason *string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Get current status
	var oldStatus int16
	var createdAt time.Time
	err = tx.QueryRow(ctx, `SELECT docstatus, created_at FROM journal_entries WHERE id = $1 AND tenant_id = $2`, journalID, tenantID).Scan(&oldStatus, &createdAt)
	if err != nil {
		return fmt.Errorf("get current status: %w", err)
	}

	// Update status
	_, err = tx.Exec(ctx, `UPDATE journal_entries SET docstatus = $1, updated_at = NOW() WHERE id = $2 AND tenant_id = $3`, newStatus, journalID, tenantID)
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	// Record state transition
	transition := &model.VoucherStateTransition{
		ID:             uuid.New(),
		JournalID:      journalID,
		TenantID:       tenantID,
		FromStatus:     model.VoucherStatus(fmt.Sprintf("%d", oldStatus)),
		ToStatus:       model.VoucherStatus(fmt.Sprintf("%d", newStatus)),
		Action:         action,
		ChangedBy:      changedBy,
		ChangedByName:  changedByName,
		Reason:         reason,
		CreatedAt:      time.Now(),
	}
	_, err = tx.Exec(ctx, `INSERT INTO voucher_state_transitions (id, journal_id, tenant_id, from_status, to_status, action, changed_by, changed_by_name, reason, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		transition.ID, transition.JournalID, transition.TenantID, transition.FromStatus, transition.ToStatus,
		transition.Action, transition.ChangedBy, transition.ChangedByName, transition.Reason, transition.CreatedAt)
	if err != nil {
		return fmt.Errorf("record state transition: %w", err)
	}

	return tx.Commit(ctx)
}

// GetStatus retrieves the current docstatus of a journal entry.
func (r *JournalRepository) GetStatus(ctx context.Context, tenantID uuid.UUID, journalID uuid.UUID) (int16, error) {
	var status int16
	err := r.pool.QueryRow(ctx, `SELECT docstatus FROM journal_entries WHERE id = $1 AND tenant_id = $2`, journalID, tenantID).Scan(&status)
	if err != nil {
		return 0, fmt.Errorf("get status: %w", err)
	}
	return status, nil
}

// GetPostedByPeriod retrieves all posted (docstatus=1) journal entries for a given period.
func (r *JournalRepository) GetPostedByPeriod(ctx context.Context, tenantID uuid.UUID, periodNo string) ([]model.JournalEntry, error) {
	query := `
		SELECT id, voucher_no, voucher_type, posting_date, company_id, tenant_id, remark,
		       docstatus, reversed_id, reversal_id, submitted_by, submitted_at, created_by,
		       created_at, updated_at
		FROM journal_entries
		WHERE tenant_id = $1 AND docstatus = 1 AND TO_CHAR(posting_date, 'YYYY-MM') = $2
		ORDER BY posting_date DESC, voucher_no DESC`

	rows, err := r.pool.Query(ctx, query, tenantID, periodNo)
	if err != nil {
		return nil, fmt.Errorf("get posted by period: %w", err)
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

// GetLines retrieves all journal entry lines for a given journal entry.
func (r *JournalRepository) GetLines(ctx context.Context, tenantID uuid.UUID, journalEntryID uuid.UUID) ([]model.JournalEntryLine, error) {
	query := `
		SELECT id, journal_entry_id, account_id, debit, credit, debit_ccy, credit_ccy,
		       account_ccy, exchange_rate, party_type, party_id, cost_center_id, project_id,
		       user_remark, reconciled, tenant_id
		FROM journal_entry_lines
		WHERE journal_entry_id = $1 AND tenant_id = $2`

	rows, err := r.pool.Query(ctx, query, journalEntryID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get lines: %w", err)
	}
	defer rows.Close()

	var lines []model.JournalEntryLine
	for rows.Next() {
		var line model.JournalEntryLine
		if err := rows.Scan(
			&line.ID, &line.JournalEntryID, &line.AccountID, &line.Debit, &line.Credit,
			&line.DebitCcy, &line.CreditCcy, &line.AccountCcy, &line.ExchangeRate,
			&line.PartyType, &line.PartyID, &line.CostCenterID, &line.ProjectID,
			&line.UserRemark, &line.Reconciled, &line.TenantID,
		); err != nil {
			return nil, fmt.Errorf("scan line: %w", err)
		}
		lines = append(lines, line)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate lines: %w", err)
	}
	return lines, nil
}

// ListVouchers retrieves vouchers with optional filters (for manual CRUD UI).
func (r *JournalRepository) ListVouchers(ctx context.Context, tenantID uuid.UUID, startDate, endDate *time.Time, voucherType *string, docStatus *int16, accountID *uuid.UUID, limit, offset int) ([]model.JournalEntry, error) {
	query := `
		SELECT DISTINCT je.id, je.voucher_no, je.voucher_type, je.posting_date, je.company_id,
		       je.tenant_id, je.remark, je.docstatus, je.reversed_id, je.reversal_id,
		       je.submitted_by, je.submitted_at, je.created_by, je.created_at, je.updated_at
		FROM journal_entries je`
	var args []interface{}
	argIdx := 1

	if accountID != nil {
		query += fmt.Sprintf(" INNER JOIN journal_entry_lines jel ON je.id = jel.journal_entry_id AND jel.tenant_id = $%d", argIdx)
		args = append(args, tenantID)
		argIdx++
		query += fmt.Sprintf(" AND jel.account_id = $%d", argIdx)
		args = append(args, *accountID)
		argIdx++
	} else {
		query += " WHERE je.tenant_id = $" + fmt.Sprintf("%d", argIdx)
		args = append(args, tenantID)
		argIdx++
	}

	if startDate != nil {
		query += fmt.Sprintf(" AND je.posting_date >= $%d", argIdx)
		args = append(args, *startDate)
		argIdx++
	}
	if endDate != nil {
		query += fmt.Sprintf(" AND je.posting_date <= $%d", argIdx)
		args = append(args, *endDate)
		argIdx++
	}
	if voucherType != nil {
		query += fmt.Sprintf(" AND je.voucher_type = $%d", argIdx)
		args = append(args, *voucherType)
		argIdx++
	}
	if docStatus != nil {
		query += fmt.Sprintf(" AND je.docstatus = $%d", argIdx)
		args = append(args, *docStatus)
		argIdx++
	}

	query += " ORDER BY je.posting_date DESC, je.voucher_no DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list vouchers: %w", err)
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
			return nil, fmt.Errorf("scan voucher: %w", err)
		}
		entries = append(entries, je)
	}
	return entries, rows.Err()
}

// Update updates the header fields of a journal entry (not lines).
func (r *JournalRepository) Update(ctx context.Context, tenantID uuid.UUID, je *model.JournalEntry) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE journal_entries
		SET voucher_type = $1, posting_date = $2, remark = $3, updated_at = $4
		WHERE id = $5 AND tenant_id = $6`,
		je.VoucherType, je.PostingDate, je.Remark, je.UpdatedAt, je.ID, tenantID)
	if err != nil {
		return fmt.Errorf("update journal entry: %w", err)
	}
	return nil
}

// DeleteLines deletes all lines for a given journal entry.
func (r *JournalRepository) DeleteLines(ctx context.Context, tenantID, journalEntryID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM journal_entry_lines WHERE journal_entry_id = $1 AND tenant_id = $2`, journalEntryID, tenantID)
	if err != nil {
		return fmt.Errorf("delete lines: %w", err)
	}
	return nil
}

// DeleteVoucher deletes a journal entry and its lines.
func (r *JournalRepository) DeleteVoucher(ctx context.Context, tenantID, journalEntryID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `DELETE FROM journal_entry_lines WHERE journal_entry_id = $1 AND tenant_id = $2`, journalEntryID, tenantID)
	if err != nil {
		return fmt.Errorf("delete lines: %w", err)
	}
	_, err = tx.Exec(ctx, `DELETE FROM voucher_state_transitions WHERE journal_id = $1 AND tenant_id = $2`, journalEntryID, tenantID)
	if err != nil {
		return fmt.Errorf("delete transitions: %w", err)
	}
	_, err = tx.Exec(ctx, `DELETE FROM journal_entries WHERE id = $1 AND tenant_id = $2`, journalEntryID, tenantID)
	if err != nil {
		return fmt.Errorf("delete journal entry: %w", err)
	}

	return tx.Commit(ctx)
}

// GetTransitions retrieves all state transitions for a given journal entry.
func (r *JournalRepository) GetTransitions(ctx context.Context, tenantID uuid.UUID, journalID uuid.UUID) ([]model.VoucherStateTransition, error) {
	query := `
		SELECT id, journal_id, tenant_id, from_status, to_status, action, changed_by, changed_by_name, reason, created_at
		FROM voucher_state_transitions
		WHERE journal_id = $1 AND tenant_id = $2
		ORDER BY created_at ASC`

	rows, err := r.pool.Query(ctx, query, journalID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get transitions: %w", err)
	}
	defer rows.Close()

	var transitions []model.VoucherStateTransition
	for rows.Next() {
		var t model.VoucherStateTransition
		if err := rows.Scan(
			&t.ID, &t.JournalID, &t.TenantID, &t.FromStatus, &t.ToStatus,
			&t.Action, &t.ChangedBy, &t.ChangedByName, &t.Reason, &t.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan transition: %w", err)
		}
		transitions = append(transitions, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate transitions: %w", err)
	}
	return transitions, nil
}
