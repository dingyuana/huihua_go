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

type WriteOffRepository struct {
	db *pgxpool.Pool
}

func NewWriteOffRepository(db *pgxpool.Pool) *WriteOffRepository {
	return &WriteOffRepository{db: db}
}

func (r *WriteOffRepository) Create(ctx context.Context, record *model.WriteOffRecord) error {
	sql := `
		INSERT INTO write_off_records (
			tenant_id, write_off_no, type, receipt_payment_id, 
			receivable_payable_id, receivable_payable_type, amount, 
			diff_amount, diff_account_code, write_off_date, 
			operator, status, remark, match_rule
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRow(ctx, sql,
		record.TenantID,
		record.WriteOffNo,
		record.Type,
		record.ReceiptPaymentID,
		record.ReceivablePayableID,
		record.ReceivablePayableType,
		record.Amount,
		record.DiffAmount,
		record.DiffAccountCode,
		record.WriteOffDate,
		record.Operator,
		record.Status,
		record.Remark,
		record.MatchRule,
	).Scan(&record.ID, &record.CreatedAt, &record.UpdatedAt)
	return err
}

func (r *WriteOffRepository) CreateTx(ctx context.Context, tx pgx.Tx, record *model.WriteOffRecord) error {
	sql := `
		INSERT INTO write_off_records (
			tenant_id, write_off_no, type, receipt_payment_id, 
			receivable_payable_id, receivable_payable_type, amount, 
			diff_amount, diff_account_code, write_off_date, 
			operator, status, remark, match_rule
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at, updated_at
	`
	err := tx.QueryRow(ctx, sql,
		record.TenantID,
		record.WriteOffNo,
		record.Type,
		record.ReceiptPaymentID,
		record.ReceivablePayableID,
		record.ReceivablePayableType,
		record.Amount,
		record.DiffAmount,
		record.DiffAccountCode,
		record.WriteOffDate,
		record.Operator,
		record.Status,
		record.Remark,
		record.MatchRule,
	).Scan(&record.ID, &record.CreatedAt, &record.UpdatedAt)
	return err
}

func (r *WriteOffRepository) BatchCreateTx(ctx context.Context, tx pgx.Tx, records []*model.WriteOffRecord) error {
	batch := &pgx.Batch{}
	for _, record := range records {
		sql := `
			INSERT INTO write_off_records (
				tenant_id, write_off_no, type, receipt_payment_id, 
				receivable_payable_id, receivable_payable_type, amount, 
				diff_amount, diff_account_code, write_off_date, 
				operator, status, remark, match_rule
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		`
		batch.Queue(sql,
			record.TenantID,
			record.WriteOffNo,
			record.Type,
			record.ReceiptPaymentID,
			record.ReceivablePayableID,
			record.ReceivablePayableType,
			record.Amount,
			record.DiffAmount,
			record.DiffAccountCode,
			record.WriteOffDate,
			record.Operator,
			record.Status,
			record.Remark,
			record.MatchRule,
		)
	}
	results := tx.SendBatch(ctx, batch)
	defer results.Close()
	for range records {
		if _, err := results.Exec(); err != nil {
			return err
		}
	}
	return nil
}

func (r *WriteOffRepository) GetByID(ctx context.Context, id int64) (*model.WriteOffRecord, error) {
	sql := `
		SELECT id, tenant_id, write_off_no, type, receipt_payment_id, 
		       receivable_payable_id, receivable_payable_type, amount, 
		       diff_amount, diff_account_code, write_off_date, 
		       operator, status, remark, match_rule, created_at, updated_at
		FROM write_off_records WHERE id = $1
	`
	var record model.WriteOffRecord
	err := r.db.QueryRow(ctx, sql, id).Scan(
		&record.ID,
		&record.TenantID,
		&record.WriteOffNo,
		&record.Type,
		&record.ReceiptPaymentID,
		&record.ReceivablePayableID,
		&record.ReceivablePayableType,
		&record.Amount,
		&record.DiffAmount,
		&record.DiffAccountCode,
		&record.WriteOffDate,
		&record.Operator,
		&record.Status,
		&record.Remark,
		&record.MatchRule,
		&record.CreatedAt,
		&record.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &record, err
}

func (r *WriteOffRepository) List(ctx context.Context, tenantID uuid.UUID, params map[string]interface{}) ([]*model.WriteOffRecord, error) {
	sql := `
		SELECT id, tenant_id, write_off_no, type, receipt_payment_id, 
		       receivable_payable_id, receivable_payable_type, amount, 
		       diff_amount, diff_account_code, write_off_date, 
		       operator, status, remark, match_rule, created_at, updated_at
		FROM write_off_records 
		WHERE tenant_id = $1
	`
	args := []interface{}{tenantID}
	argIdx := 2

	if status, ok := params["status"]; ok {
		sql += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}
	if docType, ok := params["document_type"]; ok {
		sql += fmt.Sprintf(" AND type = $%d", argIdx)
		args = append(args, docType)
		argIdx++
	}
	if startDate, ok := params["start_date"]; ok {
		sql += fmt.Sprintf(" AND write_off_date >= $%d", argIdx)
		args = append(args, startDate)
		argIdx++
	}
	if endDate, ok := params["end_date"]; ok {
		sql += fmt.Sprintf(" AND write_off_date <= $%d", argIdx)
		args = append(args, endDate)
		argIdx++
	}

	sql += " ORDER BY write_off_date DESC"

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*model.WriteOffRecord
	for rows.Next() {
		var record model.WriteOffRecord
		err := rows.Scan(
			&record.ID,
			&record.TenantID,
			&record.WriteOffNo,
			&record.Type,
			&record.ReceiptPaymentID,
			&record.ReceivablePayableID,
			&record.ReceivablePayableType,
			&record.Amount,
			&record.DiffAmount,
			&record.DiffAccountCode,
			&record.WriteOffDate,
			&record.Operator,
			&record.Status,
			&record.Remark,
			&record.MatchRule,
			&record.CreatedAt,
			&record.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, &record)
	}
	return records, nil
}

func (r *WriteOffRepository) ReverseTx(ctx context.Context, tx pgx.Tx, id int64) error {
	sql := `UPDATE write_off_records SET status = 0, updated_at = $1 WHERE id = $2`
	_, err := tx.Exec(ctx, sql, time.Now(), id)
	return err
}

func (r *WriteOffRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.db.Begin(ctx)
}

func (r *WriteOffRepository) GetUnmatchedByCounterparty(ctx context.Context, tenantID uuid.UUID, counterpartyID uuid.UUID) ([]*model.WriteOffUnmatchedItem, error) {
	sql := `
		SELECT 
			pe.id, 'payment_entry' as document_type, pe.payment_no as document_no,
			pe.counterparty_id, p.name as counterparty_name,
			pe.amount, pe.remaining_amount, pe.payment_date as document_date,
			NULL as due_date, pe.description, NULL as unmatched_reason
		FROM payment_entries pe
		JOIN parties p ON pe.counterparty_id = p.id
		WHERE pe.tenant_id = $1 AND pe.counterparty_id = $2 
		  AND pe.remaining_amount > 0
		UNION ALL
		SELECT 
			ari.id, 'ar_invoice' as document_type, ari.invoice_no as document_no,
			ari.customer_id, p.name as counterparty_name,
			ari.amount, ari.outstanding_amount, ari.invoice_date as document_date,
			ari.due_date, ari.description, NULL as unmatched_reason
		FROM ar_invoices ari
		JOIN parties p ON ari.customer_id = p.id
		WHERE ari.tenant_id = $1 AND ari.customer_id = $2 
		  AND ari.outstanding_amount > 0
		UNION ALL
		SELECT 
			api.id, 'ap_invoice' as document_type, api.invoice_no as document_no,
			api.supplier_id, p.name as counterparty_name,
			api.amount, api.outstanding_amount, api.invoice_date as document_date,
			api.due_date, api.description, NULL as unmatched_reason
		FROM ap_invoices api
		JOIN parties p ON api.supplier_id = p.id
		WHERE api.tenant_id = $1 AND api.supplier_id = $2 
		  AND api.outstanding_amount > 0
		ORDER BY document_date
	`
	rows, err := r.db.Query(ctx, sql, tenantID, counterpartyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*model.WriteOffUnmatchedItem
	for rows.Next() {
		var item model.WriteOffUnmatchedItem
		err := rows.Scan(
			&item.ID,
			&item.DocumentType,
			&item.DocumentNo,
			&item.CounterpartyID,
			&item.CounterpartyName,
			&item.Amount,
			&item.RemainingAmount,
			&item.DocumentDate,
			&item.DueDate,
			&item.Description,
			&item.UnmatchedReason,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, &item)
	}
	return items, nil
}

func (r *WriteOffRepository) GetPrepaidBalance(ctx context.Context, tenantID, counterpartyID uuid.UUID) (decimal.Decimal, error) {
	sql := `
		SELECT COALESCE(SUM(remaining_amount), 0)
		FROM payment_entries
		WHERE tenant_id = $1 AND counterparty_id = $2 
		  AND type = 'receipt' AND remaining_amount > 0
		  AND document_type = 'prepaid'
	`
	var balance decimal.Decimal
	err := r.db.QueryRow(ctx, sql, tenantID, counterpartyID).Scan(&balance)
	if err != nil {
		return decimal.Zero, err
	}
	return balance, nil
}

func (r *WriteOffRepository) Update(ctx context.Context, record *model.WriteOffRecord) error {
	sql := `
		UPDATE write_off_records
		SET type = $1, receipt_payment_id = $2, receivable_payable_id = $3,
		    receivable_payable_type = $4, amount = $5, diff_amount = $6,
		    diff_account_code = $7, write_off_date = $8, status = $9,
		    remark = $10, match_rule = $11, approver = $12,
		    approved_at = $13, reject_reason = $14, updated_at = $15
		WHERE id = $16
	`
	_, err := r.db.Exec(ctx, sql,
		record.Type,
		record.ReceiptPaymentID,
		record.ReceivablePayableID,
		record.ReceivablePayableType,
		record.Amount,
		record.DiffAmount,
		record.DiffAccountCode,
		record.WriteOffDate,
		record.Status,
		record.Remark,
		record.MatchRule,
		record.Approver,
		record.ApprovedAt,
		record.RejectReason,
		time.Now(),
		record.ID,
	)
	return err
}

func (r *WriteOffRepository) UpdateTx(ctx context.Context, tx pgx.Tx, record *model.WriteOffRecord) error {
	sql := `
		UPDATE write_off_records
		SET type = $1, receipt_payment_id = $2, receivable_payable_id = $3,
		    receivable_payable_type = $4, amount = $5, diff_amount = $6,
		    diff_account_code = $7, write_off_date = $8, status = $9,
		    remark = $10, match_rule = $11, approver = $12,
		    approved_at = $13, reject_reason = $14, updated_at = $15
		WHERE id = $16
	`
	_, err := tx.Exec(ctx, sql,
		record.Type,
		record.ReceiptPaymentID,
		record.ReceivablePayableID,
		record.ReceivablePayableType,
		record.Amount,
		record.DiffAmount,
		record.DiffAccountCode,
		record.WriteOffDate,
		record.Status,
		record.Remark,
		record.MatchRule,
		record.Approver,
		record.ApprovedAt,
		record.RejectReason,
		time.Now(),
		record.ID,
	)
	return err
}

func (r *WriteOffRepository) Delete(ctx context.Context, id int64) error {
	sql := `DELETE FROM write_off_records WHERE id = $1`
	_, err := r.db.Exec(ctx, sql, id)
	return err
}