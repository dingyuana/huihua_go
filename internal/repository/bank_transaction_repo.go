package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/model"
)

// BankTransactionRepository handles bank_transactions table operations.
type BankTransactionRepository struct {
	pool *pgxpool.Pool
}

// NewBankTransactionRepository creates a new BankTransactionRepository.
func NewBankTransactionRepository(pool *pgxpool.Pool) *BankTransactionRepository {
	return &BankTransactionRepository{pool: pool}
}

// ImportBatch inserts multiple bank transactions with duplicate checking.
// Duplicate check is based on: bank_account_id + transaction_date + description + amount.
// Returns the number of rows inserted (excluding duplicates).
func (r *BankTransactionRepository) ImportBatch(ctx context.Context, tenantID, bankAccountID uuid.UUID, txns []model.BankTransaction) (int, error) {
	if len(txns) == 0 {
		return 0, nil
	}

	// Use ON CONFLICT DO NOTHING to skip duplicates
	query := `
		INSERT INTO bank_transactions (id, tenant_id, bank_account_id, txn_date, description, 
			debit, credit, direction, reference_no, counterparty_name, classification, matched, company_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (tenant_id, bank_account_id, txn_date, description, debit, credit) DO NOTHING`

	for _, txn := range txns {
		if txn.ID == uuid.Nil {
			txn.ID = uuid.New()
		}

		var debit, credit decimal.Decimal
		if txn.Direction != nil && *txn.Direction == "in" {
			debit = txn.Debit
			credit = decimal.Zero
		} else if txn.Direction != nil && *txn.Direction == "out" {
			debit = decimal.Zero
			credit = txn.Credit
		} else {
			debit = txn.Debit
			credit = txn.Credit
		}

		classification := "pending"
		if txn.Classification != nil && *txn.Classification != "" {
			classification = *txn.Classification
		}

		_, err := r.pool.Exec(ctx, query,
			txn.ID, tenantID, bankAccountID, txn.TxnDate, txn.Description,
			debit, credit, txn.Direction, txn.ReferenceNo, txn.CounterpartyName,
			classification, txn.Matched, txn.CompanyID, time.Now())
		if err != nil {
			return 0, fmt.Errorf("import batch: %w", err)
		}
	}

	return len(txns), nil
}

// ListByBankAccount retrieves bank transactions with filters.
func (r *BankTransactionRepository) ListByBankAccount(ctx context.Context, tenantID, bankAccountID uuid.UUID, filter model.BankTxnFilter) ([]model.BankTransaction, int, error) {
	// Build WHERE clause
	conditions := []string{"tenant_id = $1", "bank_account_id = $2"}
	args := []interface{}{tenantID, bankAccountID}
	argIdx := 3

	if filter.StartDate != nil {
		conditions = append(conditions, fmt.Sprintf("txn_date >= $%d", argIdx))
		args = append(args, *filter.StartDate)
		argIdx++
	}
	if filter.EndDate != nil {
		conditions = append(conditions, fmt.Sprintf("txn_date <= $%d", argIdx))
		args = append(args, *filter.EndDate)
		argIdx++
	}
	if filter.MinAmount != nil {
		conditions = append(conditions, fmt.Sprintf("(debit >= $%d OR credit >= $%d)", argIdx, argIdx))
		args = append(args, *filter.MinAmount)
		argIdx++
	}
	if filter.MaxAmount != nil {
		conditions = append(conditions, fmt.Sprintf("(debit <= $%d OR credit <= $%d)", argIdx, argIdx))
		args = append(args, *filter.MaxAmount)
		argIdx++
	}
	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("matched = $%d", argIdx))
		args = append(args, *filter.Status == model.BankTxnStatusMatched)
		argIdx++
	}
	if filter.Search != nil && *filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(description ILIKE $%d OR counterparty_name ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+*filter.Search+"%")
		argIdx++
	}
	if filter.Classification != nil && *filter.Classification != "" {
		conditions = append(conditions, fmt.Sprintf("classification = $%d", argIdx))
		args = append(args, *filter.Classification)
		argIdx++
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count total
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM bank_transactions WHERE %s`, whereClause)
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count transactions: %w", err)
	}

	// Pagination
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 50
	}
	offset := (page - 1) * pageSize

	// Query with pagination
	query := fmt.Sprintf(`
		SELECT id, tenant_id, bank_account_id, txn_date, description, debit, credit, direction, 
			reference_no, counterparty_name, classification, matched, matched_payment_entry_id, matched_gl_entry_id,
			imported_from, raw_data, company_id, created_at
		FROM bank_transactions
		WHERE %s
		ORDER BY txn_date DESC, created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIdx, argIdx+1)
	args = append(args, pageSize, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list transactions: %w", err)
	}
	defer rows.Close()

	var txns []model.BankTransaction
	for rows.Next() {
		var txn model.BankTransaction
		err := rows.Scan(
			&txn.ID, &txn.TenantID, &txn.BankAccountID, &txn.TxnDate, &txn.Description,
			&txn.Debit, &txn.Credit, &txn.Direction, &txn.ReferenceNo, &txn.CounterpartyName,
			&txn.Classification,
			&txn.Matched, &txn.MatchedPaymentEntryID, &txn.MatchedGLEntryID, &txn.ImportedFrom,
			&txn.RawData, &txn.CompanyID, &txn.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan transaction: %w", err)
		}
		txns = append(txns, txn)
	}

	return txns, total, rows.Err()
}

// GetByID retrieves a bank transaction by ID.
func (r *BankTransactionRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*model.BankTransaction, error) {
	query := `
		SELECT id, tenant_id, bank_account_id, txn_date, description, debit, credit, direction,
			reference_no, counterparty_name, classification, matched, matched_payment_entry_id, matched_gl_entry_id,
			imported_from, raw_data, company_id, created_at
		FROM bank_transactions
		WHERE tenant_id = $1 AND id = $2`

	txn := &model.BankTransaction{}
	err := r.pool.QueryRow(ctx, query, tenantID, id).Scan(
		&txn.ID, &txn.TenantID, &txn.BankAccountID, &txn.TxnDate, &txn.Description,
		&txn.Debit, &txn.Credit, &txn.Direction, &txn.ReferenceNo, &txn.CounterpartyName,
		&txn.Classification,
		&txn.Matched, &txn.MatchedPaymentEntryID, &txn.MatchedGLEntryID, &txn.ImportedFrom,
		&txn.RawData, &txn.CompanyID, &txn.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get transaction by id: %w", err)
	}
	return txn, nil
}

// UpdateStatus updates the status of a bank transaction.
func (r *BankTransactionRepository) UpdateStatus(ctx context.Context, tenantID, id uuid.UUID, matched bool) error {
	query := `UPDATE bank_transactions SET matched = $3, updated_at = NOW() WHERE tenant_id = $1 AND id = $2`
	_, err := r.pool.Exec(ctx, query, tenantID, id, matched)
	return err
}

// MarkAsMatched marks multiple transactions as matched with rule and journal entry info.
func (r *BankTransactionRepository) MarkAsMatched(ctx context.Context, tenantID uuid.UUID, ids []uuid.UUID, journalEntryID uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	query := `
		UPDATE bank_transactions 
		SET matched = TRUE, matched_gl_entry_id = $3, updated_at = NOW()
		WHERE tenant_id = $1 AND id = ANY($2)`

	_, err := r.pool.Exec(ctx, query, tenantID, ids, journalEntryID)
	return err
}

// MarkAsReconciled marks all transactions for a bank account in a period as reconciled.
func (r *BankTransactionRepository) MarkAsReconciled(ctx context.Context, tenantID, bankAccountID uuid.UUID, periodNo int) error {
	query := `
		UPDATE bank_transactions 
		SET reconciled = TRUE, reconciled_period = $4, updated_at = NOW()
		WHERE tenant_id = $1 AND bank_account_id = $2 AND matched = TRUE`

	_, err := r.pool.Exec(ctx, query, tenantID, bankAccountID, periodNo)
	return err
}

// ListUnmatched returns all unmatched bank transactions for a tenant (across all accounts).
func (r *BankTransactionRepository) ListUnmatched(ctx context.Context, tenantID uuid.UUID) ([]model.BankTransaction, error) {
	query := `
		SELECT id, tenant_id, bank_account_id, txn_date, description, debit, credit, direction,
		       reference_no, counterparty_name, classification, matched, matched_payment_entry_id, matched_gl_entry_id,
		       imported_from, raw_data, company_id, created_at
		FROM bank_transactions
		WHERE tenant_id = $1 AND matched = FALSE
		ORDER BY txn_date`
	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list unmatched: %w", err)
	}
	defer rows.Close()
	var txns []model.BankTransaction
	for rows.Next() {
		var t model.BankTransaction
		if err := rows.Scan(&t.ID, &t.TenantID, &t.BankAccountID, &t.TxnDate, &t.Description,
			&t.Debit, &t.Credit, &t.Direction, &t.ReferenceNo, &t.CounterpartyName,
			&t.Classification,
			&t.Matched, &t.MatchedPaymentEntryID, &t.MatchedGLEntryID, &t.ImportedFrom,
			&t.RawData, &t.CompanyID, &t.CreatedAt); err != nil {
			return nil, err
		}
		txns = append(txns, t)
	}
	return txns, rows.Err()
}

// GetUnmatched retrieves all unmatched transactions for a bank account.
func (r *BankTransactionRepository) GetUnmatched(ctx context.Context, tenantID, bankAccountID uuid.UUID) ([]model.BankTransaction, error) {
	query := `
		SELECT id, tenant_id, bank_account_id, txn_date, description, debit, credit, direction,
			reference_no, counterparty_name, classification, matched, matched_payment_entry_id, matched_gl_entry_id,
			imported_from, raw_data, company_id, created_at
		FROM bank_transactions
		WHERE tenant_id = $1 AND bank_account_id = $2 AND matched = FALSE
		ORDER BY txn_date`

	rows, err := r.pool.Query(ctx, query, tenantID, bankAccountID)
	if err != nil {
		return nil, fmt.Errorf("get unmatched: %w", err)
	}
	defer rows.Close()

	var txns []model.BankTransaction
	for rows.Next() {
		var txn model.BankTransaction
		err := rows.Scan(
			&txn.ID, &txn.TenantID, &txn.BankAccountID, &txn.TxnDate, &txn.Description,
			&txn.Debit, &txn.Credit, &txn.Direction, &txn.ReferenceNo, &txn.CounterpartyName,
			&txn.Classification,
			&txn.Matched, &txn.MatchedPaymentEntryID, &txn.MatchedGLEntryID, &txn.ImportedFrom,
			&txn.RawData, &txn.CompanyID, &txn.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan unmatched: %w", err)
		}
		txns = append(txns, txn)
	}

	return txns, rows.Err()
}

// GetByIDSimple retrieves a bank transaction without tenant check (for internal use).
func (r *BankTransactionRepository) GetByIDSimple(ctx context.Context, id uuid.UUID) (*model.BankTransaction, error) {
	query := `
		SELECT id, tenant_id, bank_account_id, txn_date, description, debit, credit, direction,
			reference_no, counterparty_name, classification, matched, matched_payment_entry_id, matched_gl_entry_id,
			imported_from, raw_data, company_id, created_at
		FROM bank_transactions WHERE id = $1`

	txn := &model.BankTransaction{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&txn.ID, &txn.TenantID, &txn.BankAccountID, &txn.TxnDate, &txn.Description,
		&txn.Debit, &txn.Credit, &txn.Direction, &txn.ReferenceNo, &txn.CounterpartyName,
		&txn.Classification,
		&txn.Matched, &txn.MatchedPaymentEntryID, &txn.MatchedGLEntryID, &txn.ImportedFrom,
		&txn.RawData, &txn.CompanyID, &txn.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get by id simple: %w", err)
	}
	return txn, nil
}

// UpdateMatchedInfo updates the matched rule info for a transaction.
func (r *BankTransactionRepository) UpdateMatchedInfo(ctx context.Context, tenantID, id, glEntryID uuid.UUID) error {
	query := `
		UPDATE bank_transactions 
		SET matched = TRUE, matched_gl_entry_id = $3, updated_at = NOW()
		WHERE tenant_id = $1 AND id = $2`
	_, err := r.pool.Exec(ctx, query, tenantID, id, glEntryID)
	return err
}

// GetMatchedByPeriod retrieves all matched bank transactions within a date range.
func (r *BankTransactionRepository) GetMatchedByPeriod(ctx context.Context, tenantID, bankAccountID uuid.UUID, startDate, endDate time.Time) ([]model.BankTransaction, error) {
	query := `
		SELECT id, tenant_id, bank_account_id, txn_date, description, debit, credit, direction,
		       reference_no, counterparty_name, classification, matched, matched_payment_entry_id, matched_gl_entry_id,
		       imported_from, raw_data, company_id, created_at
		FROM bank_transactions
		WHERE tenant_id = $1 AND bank_account_id = $2 AND matched = TRUE AND txn_date >= $3 AND txn_date <= $4
		ORDER BY txn_date, created_at`

	rows, err := r.pool.Query(ctx, query, tenantID, bankAccountID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txns []model.BankTransaction
	for rows.Next() {
		var txn model.BankTransaction
		err := rows.Scan(
			&txn.ID, &txn.TenantID, &txn.BankAccountID, &txn.TxnDate, &txn.Description,
			&txn.Debit, &txn.Credit, &txn.Direction, &txn.ReferenceNo, &txn.CounterpartyName,
			&txn.Classification,
			&txn.Matched, &txn.MatchedPaymentEntryID, &txn.MatchedGLEntryID, &txn.ImportedFrom,
			&txn.RawData, &txn.CompanyID, &txn.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		txns = append(txns, txn)
	}
	return txns, rows.Err()
}

// UpdateClassification updates the classification of a bank transaction.
func (r *BankTransactionRepository) UpdateClassification(ctx context.Context, tenantID, id uuid.UUID, classification string) error {
	query := `UPDATE bank_transactions SET classification = $3, updated_at = NOW() WHERE tenant_id = $1 AND id = $2`
	_, err := r.pool.Exec(ctx, query, tenantID, id, classification)
	return err
}

// GetPool returns the underlying connection pool.
func (r *BankTransactionRepository) GetPool() *pgxpool.Pool {
	return r.pool
}