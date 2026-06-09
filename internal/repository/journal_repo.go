package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
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
		INSERT INTO journal_entries (id, voucher_no, voucher_type, posting_date, company_id, tenant_id, remark, docstatus, reversed_id, reversal_id, submitted_by, submitted_at, created_by, counterparty_name, source_doc_type, source_doc_id, source_doc_no, source_type, source_id, source_invoice_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
		RETURNING created_at, updated_at`

	if je.ID == uuid.Nil {
		je.ID = uuid.New()
	}

	err := r.pool.QueryRow(ctx, query,
		je.ID, je.VoucherNo, je.VoucherType, je.PostingDate, je.CompanyID,
		tenantID, je.Remark, je.DocStatus, je.ReversedID, je.ReversalID,
		je.SubmittedBy, je.SubmittedAt, je.CreatedBy, je.CounterpartyName,
		je.SourceDocType, je.SourceDocID, je.SourceDocNo,
		je.SourceType, je.SourceID, je.SourceInvoiceID,
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
		       created_at, updated_at, counterparty_name,
		       source_doc_type, source_doc_id, source_doc_no,
		       source_type, source_id, source_invoice_id, version
		FROM journal_entries
		WHERE id = $1 AND tenant_id = $2`

	je := &model.JournalEntry{}
	err := r.pool.QueryRow(ctx, query, id, tenantID).Scan(
		&je.ID, &je.VoucherNo, &je.VoucherType, &je.PostingDate, &je.CompanyID,
		&je.TenantID, &je.Remark, &je.DocStatus, &je.ReversedID, &je.ReversalID,
		&je.SubmittedBy, &je.SubmittedAt, &je.CreatedBy, &je.CreatedAt, &je.UpdatedAt,
		&je.CounterpartyName,
		&je.SourceDocType, &je.SourceDocID, &je.SourceDocNo,
		&je.SourceType, &je.SourceID, &je.SourceInvoiceID, &je.Version,
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

// BeginTx starts a new database transaction.
func (r *JournalRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pool.Begin(ctx)
}

// CreateTx inserts a new journal entry within an existing transaction and returns it.
func (r *JournalRepository) CreateTx(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID, je *model.JournalEntry) (*model.JournalEntry, error) {
	query := `
		INSERT INTO journal_entries (id, voucher_no, voucher_type, posting_date, company_id, tenant_id, remark, docstatus, reversed_id, reversal_id, submitted_by, submitted_at, created_by, counterparty_name, source_doc_type, source_doc_id, source_doc_no, source_type, source_id, source_invoice_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
		RETURNING created_at, updated_at`

	if je.ID == uuid.Nil {
		je.ID = uuid.New()
	}

	err := tx.QueryRow(ctx, query,
		je.ID, je.VoucherNo, je.VoucherType, je.PostingDate, je.CompanyID,
		tenantID, je.Remark, je.DocStatus, je.ReversedID, je.ReversalID,
		je.SubmittedBy, je.SubmittedAt, je.CreatedBy, je.CounterpartyName,
		je.SourceDocType, je.SourceDocID, je.SourceDocNo,
		je.SourceType, je.SourceID, je.SourceInvoiceID,
	).Scan(&je.CreatedAt, &je.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create journal entry: %w", err)
	}

	je.TenantID = tenantID
	return je, nil
}

// AddLinesTx inserts multiple journal entry lines within an existing transaction.
func (r *JournalRepository) AddLinesTx(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID, journalEntryID uuid.UUID, lines []model.JournalEntryLine) ([]model.JournalEntryLine, error) {
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

		_, err := tx.Exec(ctx, query,
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

// UpdateStatusTx updates the docstatus of a journal entry and records the transition within an existing transaction.
func (r *JournalRepository) UpdateStatusTx(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID, journalID uuid.UUID, newStatus int16, changedBy uuid.UUID, changedByName *string, action model.VoucherAction, reason *string) error {
	// Get current status
	var oldStatus int16
	var createdAt time.Time
	err := tx.QueryRow(ctx, `SELECT docstatus, created_at FROM journal_entries WHERE id = $1 AND tenant_id = $2`, journalID, tenantID).Scan(&oldStatus, &createdAt)
	if err != nil {
		return fmt.Errorf("get current status: %w", err)
	}

	// Update status + approved_by/approved_at when transitioning 0→1 (submit → posted/approve)
	approvedBy := changedBy
	approvedAt := time.Now()
	_, err = tx.Exec(ctx, `UPDATE journal_entries SET docstatus = $1, updated_at = NOW(), approved_by = $2, approved_at = $3 WHERE id = $4 AND tenant_id = $5`, newStatus, approvedBy, approvedAt, journalID, tenantID)
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	// Record state transition
	transition := &model.VoucherStateTransition{
		ID:            uuid.New(),
		VoucherID:     journalID,
		TenantID:      tenantID,
		FromStatus:    model.VoucherStatus(fmt.Sprintf("%d", oldStatus)),
		ToStatus:      model.VoucherStatus(fmt.Sprintf("%d", newStatus)),
		Action:        action,
		ChangedBy:     changedBy,
		ChangedByName: changedByName,
		Reason:        reason,
		CreatedAt:     time.Now(),
	}
	_, err = tx.Exec(ctx, `INSERT INTO voucher_state_transitions (id, voucher_id, tenant_id, from_status, to_status, action, changed_by, changed_by_name, reason, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		transition.ID, transition.VoucherID, transition.TenantID, transition.FromStatus, transition.ToStatus,
		transition.Action, transition.ChangedBy, transition.ChangedByName, transition.Reason, transition.CreatedAt)
	if err != nil {
		return fmt.Errorf("record state transition: %w", err)
	}

	return nil
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

	// Update status + approved_by/approved_at when transitioning 0→1 (submit → posted/approve)
	approvedBy := changedBy
	approvedAt := time.Now()
	_, err = tx.Exec(ctx, `UPDATE journal_entries SET docstatus = $1, updated_at = NOW(), approved_by = $2, approved_at = $3 WHERE id = $4 AND tenant_id = $5`, newStatus, approvedBy, approvedAt, journalID, tenantID)
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	// Record state transition
	transition := &model.VoucherStateTransition{
		ID:            uuid.New(),
		VoucherID:     journalID,
		TenantID:      tenantID,
		FromStatus:    model.VoucherStatus(fmt.Sprintf("%d", oldStatus)),
		ToStatus:      model.VoucherStatus(fmt.Sprintf("%d", newStatus)),
		Action:        action,
		ChangedBy:     changedBy,
		ChangedByName: changedByName,
		Reason:        reason,
		CreatedAt:     time.Now(),
	}
	_, err = tx.Exec(ctx, `INSERT INTO voucher_state_transitions (id, voucher_id, tenant_id, from_status, to_status, action, changed_by, changed_by_name, reason, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		transition.ID, transition.VoucherID, transition.TenantID, transition.FromStatus, transition.ToStatus,
		transition.Action, transition.ChangedBy, transition.ChangedByName, transition.Reason, transition.CreatedAt)
	if err != nil {
		return fmt.Errorf("record state transition: %w", err)
	}

	return tx.Commit(ctx)
}

// UpdateTx updates a journal entry within an existing transaction.
func (r *JournalRepository) UpdateTx(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID, je *model.JournalEntry) error {
	_, err := tx.Exec(ctx, `
		UPDATE journal_entries
		SET voucher_type = $1, posting_date = $2, remark = $3, updated_at = $4
		WHERE id = $5 AND tenant_id = $6`,
		je.VoucherType, je.PostingDate, je.Remark, je.UpdatedAt, je.ID, tenantID)
	if err != nil {
		return fmt.Errorf("update journal entry: %w", err)
	}
	return nil
}

// DeleteLinesTx deletes all lines for a given journal entry within an existing transaction.
func (r *JournalRepository) DeleteLinesTx(ctx context.Context, tx pgx.Tx, tenantID, journalEntryID uuid.UUID) error {
	_, err := tx.Exec(ctx, `DELETE FROM journal_entry_lines WHERE journal_entry_id = $1 AND tenant_id = $2`, journalEntryID, tenantID)
	if err != nil {
		return fmt.Errorf("delete lines: %w", err)
	}
	return nil
}

// DeleteVoucherTx deletes a journal entry and its transitions within an existing transaction.
func (r *JournalRepository) DeleteVoucherTx(ctx context.Context, tx pgx.Tx, tenantID, journalEntryID uuid.UUID) error {
	_, err := tx.Exec(ctx, `DELETE FROM voucher_state_transitions WHERE voucher_id = $1 AND tenant_id = $2`, journalEntryID, tenantID)
	if err != nil {
		return fmt.Errorf("delete transitions: %w", err)
	}
	_, err = tx.Exec(ctx, `DELETE FROM journal_entries WHERE id = $1 AND tenant_id = $2`, journalEntryID, tenantID)
	if err != nil {
		return fmt.Errorf("delete journal entry: %w", err)
	}
	return nil
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

// GetByPeriod retrieves all journal entries for a given period (YYYY-MM).
func (r *JournalRepository) GetByPeriod(ctx context.Context, tenantID uuid.UUID, periodNo string) ([]model.JournalEntry, error) {
	query := `
		SELECT id, voucher_no, voucher_type, posting_date, company_id, tenant_id, remark,
		       docstatus, reversed_id, reversal_id, submitted_by, submitted_at, created_by,
		       created_at, updated_at
		FROM journal_entries
		WHERE tenant_id = $1 AND TO_CHAR(posting_date, 'YYYY-MM') = $2
		ORDER BY voucher_no ASC`

	rows, err := r.pool.Query(ctx, query, tenantID, periodNo)
	if err != nil {
		return nil, fmt.Errorf("get vouchers by period: %w", err)
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

// CountUnpostedByPeriod counts vouchers with docstatus not 1 (draft/rejected) in a period.
func (r *JournalRepository) CountUnpostedByPeriod(ctx context.Context, tenantID uuid.UUID, periodStr string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM journal_entries
		WHERE tenant_id = $1 AND TO_CHAR(posting_date, 'YYYY-MM') = $2 AND docstatus < 1
	`, tenantID, periodStr).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count unposted by period: %w", err)
	}
	return count, nil
}

// CountClosingByPeriod counts closing-type vouchers in a period.
func (r *JournalRepository) CountClosingByPeriod(ctx context.Context, tenantID uuid.UUID, periodStr string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM journal_entries
		WHERE tenant_id = $1 AND TO_CHAR(posting_date, 'YYYY-MM') = $2 AND voucher_type = 'closing'
	`, tenantID, periodStr).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count closing by period: %w", err)
	}
	return count, nil
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
		SELECT jel.id, jel.journal_entry_id, jel.account_id, jel.debit, jel.credit, jel.debit_ccy, jel.credit_ccy,
		       jel.account_ccy, jel.exchange_rate, jel.party_type, jel.party_id, jel.cost_center_id, jel.project_id,
		       jel.user_remark, jel.reconciled, jel.tenant_id,
		       a.code AS account_code, a.name AS account_name
		FROM journal_entry_lines jel
		LEFT JOIN accounts a ON a.id = jel.account_id
		WHERE jel.journal_entry_id = $1 AND jel.tenant_id = $2`

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
			&line.AccountCode, &line.AccountName,
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
func (r *JournalRepository) ListVouchers(ctx context.Context, tenantID uuid.UUID, startDate, endDate *time.Time, voucherType *string, docStatus *int16, accountID *uuid.UUID, amountMin, amountMax *decimal.Decimal, keyword *string, limit, offset int) ([]model.JournalEntry, error) {
	query := `
		SELECT je.id, je.voucher_no, je.voucher_type, je.posting_date, je.company_id,
		       je.tenant_id, je.remark, je.docstatus, je.reversed_id, je.reversal_id,
		       je.submitted_by, je.submitted_at, je.created_by, je.created_at, je.updated_at,
		       COALESCE(jl.debit_total, 0) AS debit_total, COALESCE(jl.credit_total, 0) AS credit_total,
		       je.counterparty_name, je.source_doc_type, je.source_doc_id, je.source_doc_no,
		       (SELECT a.code FROM journal_entry_lines jel
		         JOIN accounts a ON a.id = jel.account_id
		         WHERE jel.journal_entry_id = je.id
		         ORDER BY jel.debit DESC, jel.credit DESC
		         LIMIT 1) AS first_account_code,
		       (SELECT a.name FROM journal_entry_lines jel
		         JOIN accounts a ON a.id = jel.account_id
		         WHERE jel.journal_entry_id = je.id
		         ORDER BY jel.debit DESC, jel.credit DESC
		         LIMIT 1) AS first_account_name
		FROM journal_entries je
		LEFT JOIN LATERAL (
		    SELECT SUM(jel.debit) AS debit_total, SUM(jel.credit) AS credit_total
		    FROM journal_entry_lines jel WHERE jel.journal_entry_id = je.id
		) jl ON true`
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

	// Amount range filter — use the lateral subquery totals
	if amountMin != nil {
		query += fmt.Sprintf(" AND (COALESCE(jl.debit_total, 0) >= $%d OR COALESCE(jl.credit_total, 0) >= $%d)", argIdx, argIdx)
		args = append(args, *amountMin)
		argIdx++
	}
	if amountMax != nil {
		query += fmt.Sprintf(" AND (COALESCE(jl.debit_total, 0) <= $%d OR COALESCE(jl.credit_total, 0) <= $%d)", argIdx, argIdx)
		args = append(args, *amountMax)
		argIdx++
	}
	// Keyword filter on remark
	if keyword != nil {
		query += fmt.Sprintf(" AND je.remark ILIKE '%%' || $%d || '%%'", argIdx)
		args = append(args, *keyword)
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
			&je.DebitTotal, &je.CreditTotal,
			&je.CounterpartyName, &je.SourceDocType, &je.SourceDocID, &je.SourceDocNo,
			&je.FirstAccountCode, &je.FirstAccountName,
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
	_, err = tx.Exec(ctx, `DELETE FROM voucher_state_transitions WHERE voucher_id = $1 AND tenant_id = $2`, journalEntryID, tenantID)
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
		SELECT id, voucher_id, tenant_id, from_status, to_status, action, changed_by, changed_by_name, reason, created_at
		FROM voucher_state_transitions
		WHERE voucher_id = $1 AND tenant_id = $2
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
			&t.ID, &t.VoucherID, &t.TenantID, &t.FromStatus, &t.ToStatus,
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

// ListUnmatched returns journal entries that have been submitted but not matched to any bank transaction.
// Amount must be fetched separately by summing lines.
func (r *JournalRepository) ListUnmatched(ctx context.Context, tenantID uuid.UUID) ([]model.JournalEntry, error) {
	query := `
		SELECT id, voucher_no, voucher_type, posting_date, company_id, tenant_id, remark, docstatus,
		       reversed_id, reversal_id, submitted_by, submitted_at, created_by,
		       updated_at, created_at
		FROM journal_entries
		WHERE tenant_id = $1 AND docstatus >= 2 AND (bank_transaction_id IS NULL OR bank_transaction_id = '00000000-0000-0000-0000-000000000000')
		ORDER BY posting_date`
	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list unmatched journals: %w", err)
	}
	defer rows.Close()
	var entries []model.JournalEntry
	for rows.Next() {
		var e model.JournalEntry
		if err := rows.Scan(&e.ID, &e.VoucherNo, &e.VoucherType, &e.PostingDate, &e.CompanyID,
			&e.TenantID, &e.Remark, &e.DocStatus, &e.ReversedID, &e.ReversalID,
			&e.SubmittedBy, &e.SubmittedAt, &e.CreatedBy,
			&e.UpdatedAt, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan unmatched journal: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// ErrConcurrentModification is returned by *WithVersion update methods when the
// stored version does not match the expected version, indicating a lost-update
// race condition.
var ErrConcurrentModification = errors.New("concurrent modification detected")

// UpdateFieldsWithVersion updates editable fields on a journal entry with optimistic
// locking. Returns ErrConcurrentModification if the stored version does not match
// expectedVersion. The version is auto-incremented on success.
//
// Use this in concurrent edit flows (e.g. voucher edit page) where two users might
// read the same voucher, then both submit conflicting updates.
func (r *JournalRepository) UpdateFieldsWithVersion(
	ctx context.Context,
	tenantID, id uuid.UUID,
	expectedVersion int64,
	fields map[string]interface{},
) error {
	if len(fields) == 0 {
		return fmt.Errorf("no fields to update")
	}
	setClauses := []string{"updated_at = NOW()", "version = version + 1"}
	args := []interface{}{}
	idx := 1
	for col, val := range fields {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, idx))
		args = append(args, val)
		idx++
	}
	args = append(args, id, tenantID, expectedVersion)
	query := fmt.Sprintf(`
		UPDATE journal_entries
		SET %s
		WHERE id = $%d AND tenant_id = $%d AND version = $%d`,
		strings.Join(setClauses, ", "), idx, idx+1, idx+2)
	tag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update journal fields with version: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrConcurrentModification
	}
	return nil
}

// UpdateStatusWithVersion updates docstatus with optimistic locking.
// Use for voucher submit/approve/reject/cancel flows where concurrent state
// transitions are possible.
func (r *JournalRepository) UpdateStatusWithVersion(
	ctx context.Context,
	tenantID, id uuid.UUID,
	expectedVersion int64,
	newStatus int16,
) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE journal_entries
		SET docstatus = $1, updated_at = NOW(), version = version + 1
		WHERE id = $2 AND tenant_id = $3 AND version = $4`,
		newStatus, id, tenantID, expectedVersion)
	if err != nil {
		return fmt.Errorf("update journal status with version: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrConcurrentModification
	}
	return nil
}
