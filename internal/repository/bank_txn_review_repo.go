package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"huihua/finance/internal/model"
)

// BankTxnStats holds the four statistics counters for the review dashboard.
type BankTxnStats struct {
	MonthlyTxns        int64
	PendingCount       int64
	AIProcessedCount   int64
	ManualPendingCount int64
}

// ListByStatus returns bank transactions filtered by status with pagination.
// It supports the review workflow AC8.
func (r *BankTransactionRepository) ListByStatus(
	ctx context.Context,
	tenantID uuid.UUID,
	status string,
	page, pageSize int,
) ([]model.BankTransaction, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}
	offset := (page - 1) * pageSize

	// Count total first.
	var total int64
	countQuery := `
		SELECT COUNT(*)
		FROM bank_transactions
		WHERE tenant_id = $1 AND status = $2`
	err := r.pool.QueryRow(ctx, countQuery, tenantID, status).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("list by status count: %w", err)
	}

	// Fetch page of results ordered by txn_date DESC.
	query := `
		SELECT id, tenant_id, bank_account_id, txn_date, description, debit, credit, direction,
			reference_no, counterparty_name, classification, matched, confirmed,
			matched_payment_entry_id, matched_gl_entry_id,
			imported_from, raw_data, company_id, created_at
		FROM bank_transactions
		WHERE tenant_id = $1 AND status = $2
		ORDER BY txn_date DESC, created_at DESC
		LIMIT $3 OFFSET $4`
	rows, err := r.pool.Query(ctx, query, tenantID, status, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list by status: %w", err)
	}
	defer rows.Close()

	var txns []model.BankTransaction
	for rows.Next() {
		var txn model.BankTransaction
		err := rows.Scan(
			&txn.ID, &txn.TenantID, &txn.BankAccountID, &txn.TxnDate, &txn.Description,
			&txn.Debit, &txn.Credit, &txn.Direction, &txn.ReferenceNo, &txn.CounterpartyName,
			&txn.Classification,
			&txn.Matched, &txn.Confirmed, &txn.MatchedPaymentEntryID, &txn.MatchedGLEntryID,
			&txn.ImportedFrom, &txn.RawData, &txn.CompanyID, &txn.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan transaction: %w", err)
		}
		txns = append(txns, txn)
	}
	return txns, total, rows.Err()
}

// GetStats returns four dashboard counters for the review workflow.
// It implements AC9: monthly_txns / pending_count / ai_processed_count / manual_pending_count.
func (r *BankTransactionRepository) GetStats(
	ctx context.Context,
	tenantID uuid.UUID,
) (*BankTxnStats, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	stats := &BankTxnStats{}

	// monthly_txns: transactions in the current month (all statuses)
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM bank_transactions
		WHERE tenant_id = $1 AND txn_date >= $2`,
		tenantID, startOfMonth).Scan(&stats.MonthlyTxns)
	if err != nil {
		return nil, fmt.Errorf("get stats monthly: %w", err)
	}

	// pending_count: transactions with status='pending'
	err = r.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM bank_transactions
		WHERE tenant_id = $1 AND status = $2`,
		tenantID, model.BankTxnReviewStatusPending).Scan(&stats.PendingCount)
	if err != nil {
		return nil, fmt.Errorf("get stats pending: %w", err)
	}

	// ai_processed_count: transactions in terminal "AI-processed" states
	// (classified, approved, voucher_generated, payment_created)
	aiStatuses := []string{
		string(model.BankTxnReviewStatusClassified),
		string(model.BankTxnReviewStatusApproved),
		string(model.BankTxnReviewStatusVoucherGenerated),
		string(model.BankTxnReviewStatusPaymentCreated),
	}
	placeholders := make([]string, len(aiStatuses))
	args := make([]interface{}, len(aiStatuses)+1)
	args[0] = tenantID
	for i, s := range aiStatuses {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = s
	}
	queryAI := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM bank_transactions
		WHERE tenant_id = $1 AND status IN (%s)`,
		strings.Join(placeholders, ","))
	err = r.pool.QueryRow(ctx, queryAI, args...).Scan(&stats.AIProcessedCount)
	if err != nil {
		return nil, fmt.Errorf("get stats ai processed: %w", err)
	}

	// manual_pending_count: transactions with status='manual_pending'
	err = r.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM bank_transactions
		WHERE tenant_id = $1 AND status = $2`,
		tenantID, model.BankTxnReviewStatusManualPending).Scan(&stats.ManualPendingCount)
	if err != nil {
		return nil, fmt.Errorf("get stats manual pending: %w", err)
	}

	return stats, nil
}

// GetByIDForUpdate acquires a row lock (FOR UPDATE) on a bank transaction.
// This is used by the SubmitReview service to lock rows before status change (AC5).
func (r *BankTransactionRepository) GetByIDForUpdate(
	ctx context.Context,
	tx pgx.Tx,
	txnID uuid.UUID,
) (*model.BankTransaction, error) {
	query := `
		SELECT id, tenant_id, bank_account_id, txn_date, description, debit, credit, direction,
			status, reference_no, counterparty_name, classification, matched, confirmed,
			matched_payment_entry_id, matched_gl_entry_id,
			imported_from, raw_data, company_id, created_at
		FROM bank_transactions
		WHERE id = $1
		FOR UPDATE`
	txn := &model.BankTransaction{}
	err := tx.QueryRow(ctx, query, txnID).Scan(
		&txn.ID, &txn.TenantID, &txn.BankAccountID, &txn.TxnDate, &txn.Description,
		&txn.Debit, &txn.Credit, &txn.Direction,
		&txn.Status,
		&txn.ReferenceNo, &txn.CounterpartyName,
		&txn.Classification,
		&txn.Matched, &txn.Confirmed, &txn.MatchedPaymentEntryID, &txn.MatchedGLEntryID,
		&txn.ImportedFrom, &txn.RawData, &txn.CompanyID, &txn.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get by id for update: %w", err)
	}
	return txn, nil
}

// BeginTx starts a new database transaction.
func (r *BankTransactionRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pool.Begin(ctx)
}