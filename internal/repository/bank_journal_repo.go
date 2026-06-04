package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"huihua/finance/internal/model"
)

// BankJournalRepository provides data access for bank_journal_entries.
type BankJournalRepository struct {
	pool *pgxpool.Pool
}

// NewBankJournalRepository creates a new BankJournalRepository.
func NewBankJournalRepository(pool *pgxpool.Pool) *BankJournalRepository {
	return &BankJournalRepository{pool: pool}
}

// Create inserts a new bank journal entry.
func (r *BankJournalRepository) Create(ctx context.Context, tenantID uuid.UUID, entry *model.BankJournalEntry) (*model.BankJournalEntry, error) {
	query := `
		INSERT INTO bank_journal_entries (id, bank_account_id, txn_date, description, debit, credit, voucher_id, voucher_no, tenant_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at`

	if entry.ID == uuid.Nil {
		entry.ID = uuid.New()
	}

	err := r.pool.QueryRow(ctx, query,
		entry.ID, entry.BankAccountID, entry.TxnDate, entry.Description,
		entry.Debit, entry.Credit, entry.VoucherID, entry.VoucherNo, tenantID,
	).Scan(&entry.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create bank journal entry: %w", err)
	}
	entry.TenantID = tenantID
	return entry, nil
}

// ListByBankAccountAndPeriod retrieves bank journal entries for a bank account within a date range.
func (r *BankJournalRepository) ListByBankAccountAndPeriod(ctx context.Context, tenantID uuid.UUID, bankAccountID uuid.UUID, startDate, endDate time.Time) ([]model.BankJournalEntry, error) {
	query := `
		SELECT id, bank_account_id, txn_date, description, debit, credit, voucher_id, voucher_no, tenant_id, created_at
		FROM bank_journal_entries
		WHERE tenant_id = $1 AND bank_account_id = $2 AND txn_date >= $3 AND txn_date <= $4
		ORDER BY txn_date ASC, created_at ASC`

	rows, err := r.pool.Query(ctx, query, tenantID, bankAccountID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("list bank journal entries: %w", err)
	}
	defer rows.Close()

	var entries []model.BankJournalEntry
	for rows.Next() {
		var entry model.BankJournalEntry
		if err := rows.Scan(
			&entry.ID, &entry.BankAccountID, &entry.TxnDate, &entry.Description,
			&entry.Debit, &entry.Credit, &entry.VoucherID, &entry.VoucherNo,
			&entry.TenantID, &entry.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan bank journal entry: %w", err)
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

// GetByVoucherID retrieves bank journal entry by voucher ID.
func (r *BankJournalRepository) GetByVoucherID(ctx context.Context, tenantID uuid.UUID, voucherID uuid.UUID) (*model.BankJournalEntry, error) {
	query := `
		SELECT id, bank_account_id, txn_date, description, debit, credit, voucher_id, voucher_no, tenant_id, created_at
		FROM bank_journal_entries
		WHERE voucher_id = $1 AND tenant_id = $2`

	entry := &model.BankJournalEntry{}
	err := r.pool.QueryRow(ctx, query, voucherID, tenantID).Scan(
		&entry.ID, &entry.BankAccountID, &entry.TxnDate, &entry.Description,
		&entry.Debit, &entry.Credit, &entry.VoucherID, &entry.VoucherNo,
		&entry.TenantID, &entry.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get bank journal entry by voucher: %w", err)
	}
	return entry, nil
}

// ListByBankAccount retrieves entries for a bank account.
func (r *BankJournalRepository) ListByBankAccount(ctx context.Context, tenantID uuid.UUID, bankAccountID uuid.UUID) ([]model.BankJournalEntry, error) {
	query := `
		SELECT id, bank_account_id, txn_date, description, debit, credit, voucher_id, voucher_no, tenant_id, created_at
		FROM bank_journal_entries
		WHERE tenant_id = $1 AND bank_account_id = $2
		ORDER BY txn_date ASC, created_at ASC`

	rows, err := r.pool.Query(ctx, query, tenantID, bankAccountID)
	if err != nil {
		return nil, fmt.Errorf("list bank journal entries: %w", err)
	}
	defer rows.Close()

	var entries []model.BankJournalEntry
	for rows.Next() {
		var entry model.BankJournalEntry
		if err := rows.Scan(
			&entry.ID, &entry.BankAccountID, &entry.TxnDate, &entry.Description,
			&entry.Debit, &entry.Credit, &entry.VoucherID, &entry.VoucherNo,
			&entry.TenantID, &entry.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan bank journal entry: %w", err)
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}