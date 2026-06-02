package repository

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"huihua/finance/internal/model"
)

type PaymentEntryRepository struct {
	pool *pgxpool.Pool
}

func NewPaymentEntryRepository(pool *pgxpool.Pool) *PaymentEntryRepository {
	return &PaymentEntryRepository{pool: pool}
}

func (r *PaymentEntryRepository) Create(ctx context.Context, tenantID uuid.UUID, pe *model.PaymentEntry) (*model.PaymentEntry, error) {
	if pe.ID == uuid.Nil {
		pe.ID = uuid.New()
	}
	pe.TenantID = tenantID

	query := `
		INSERT INTO payment_entries (
			id, payment_no, payment_type, party_type, party_id, counterparty_name,
			paid_from_id, paid_to_id, paid_amount, received_amount,
			reference_no, reference_date, posting_date,
			company_id, tenant_id, bank_account_id, docstatus, created_by, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
		RETURNING created_at`

	err := r.pool.QueryRow(ctx, query,
		pe.ID, pe.PaymentNo, pe.PaymentType, pe.PartyType, pe.PartyID, pe.CounterpartyName,
		pe.PaidFromID, pe.PaidToID, pe.PaidAmount, pe.ReceivedAmount,
		pe.ReferenceNo, pe.ReferenceDate, pe.PostingDate,
		pe.CompanyID, pe.TenantID, pe.BankAccountID, pe.DocStatus, pe.CreatedBy, time.Now(),
	).Scan(&pe.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create payment entry: %w", err)
	}
	return pe, nil
}

func (r *PaymentEntryRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*model.PaymentEntry, error) {
	query := `
		SELECT id, payment_no, payment_type, party_type, party_id, counterparty_name,
			paid_from_id, paid_to_id, paid_amount, received_amount,
			reference_no, reference_date, posting_date,
			company_id, tenant_id, bank_account_id, docstatus, created_by, created_at
		FROM payment_entries
		WHERE id = $1 AND tenant_id = $2`

	pe := &model.PaymentEntry{}
	err := r.pool.QueryRow(ctx, query, id, tenantID).Scan(
		&pe.ID, &pe.PaymentNo, &pe.PaymentType, &pe.PartyType, &pe.PartyID, &pe.CounterpartyName,
		&pe.PaidFromID, &pe.PaidToID, &pe.PaidAmount, &pe.ReceivedAmount,
		&pe.ReferenceNo, &pe.ReferenceDate, &pe.PostingDate,
		&pe.CompanyID, &pe.TenantID, &pe.BankAccountID, &pe.DocStatus, &pe.CreatedBy, &pe.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get payment entry: %w", err)
	}
	return pe, nil
}

func (r *PaymentEntryRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, filters ...func(*pgxpool.Pool, uuid.UUID) ([]model.PaymentEntry, error)) ([]model.PaymentEntry, error) {
	query := `
		SELECT id, payment_no, payment_type, party_type, party_id, counterparty_name,
			paid_from_id, paid_to_id, paid_amount, received_amount,
			reference_no, reference_date, posting_date,
			company_id, tenant_id, bank_account_id, docstatus, created_by, created_at
		FROM payment_entries
		WHERE tenant_id = $1
		ORDER BY posting_date DESC, payment_no DESC`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list payment entries: %w", err)
	}
	defer rows.Close()

	var entries []model.PaymentEntry
	for rows.Next() {
		var pe model.PaymentEntry
		if err := rows.Scan(
			&pe.ID, &pe.PaymentNo, &pe.PaymentType, &pe.PartyType, &pe.PartyID, &pe.CounterpartyName,
			&pe.PaidFromID, &pe.PaidToID, &pe.PaidAmount, &pe.ReceivedAmount,
			&pe.ReferenceNo, &pe.ReferenceDate, &pe.PostingDate,
			&pe.CompanyID, &pe.TenantID, &pe.BankAccountID, &pe.DocStatus, &pe.CreatedBy, &pe.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan payment entry: %w", err)
		}
		entries = append(entries, pe)
	}
	return entries, rows.Err()
}

func (r *PaymentEntryRepository) ListByBankAccount(ctx context.Context, tenantID, bankAccountID uuid.UUID) ([]model.PaymentEntry, error) {
	query := `
		SELECT id, payment_no, payment_type, party_type, party_id, counterparty_name,
			paid_from_id, paid_to_id, paid_amount, received_amount,
			reference_no, reference_date, posting_date,
			company_id, tenant_id, bank_account_id, docstatus, created_by, created_at
		FROM payment_entries
		WHERE tenant_id = $1 AND bank_account_id = $2
		ORDER BY posting_date DESC, payment_no DESC`

	rows, err := r.pool.Query(ctx, query, tenantID, bankAccountID)
	if err != nil {
		return nil, fmt.Errorf("list by bank account: %w", err)
	}
	defer rows.Close()

	var entries []model.PaymentEntry
	for rows.Next() {
		var pe model.PaymentEntry
		if err := rows.Scan(
			&pe.ID, &pe.PaymentNo, &pe.PaymentType, &pe.PartyType, &pe.PartyID, &pe.CounterpartyName,
			&pe.PaidFromID, &pe.PaidToID, &pe.PaidAmount, &pe.ReceivedAmount,
			&pe.ReferenceNo, &pe.ReferenceDate, &pe.PostingDate,
			&pe.CompanyID, &pe.TenantID, &pe.BankAccountID, &pe.DocStatus, &pe.CreatedBy, &pe.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan payment entry: %w", err)
		}
		entries = append(entries, pe)
	}
	return entries, rows.Err()
}

func (r *PaymentEntryRepository) Update(ctx context.Context, tenantID uuid.UUID, pe *model.PaymentEntry) error {
	query := `
		UPDATE payment_entries
		SET payment_type = $3, party_type = $4, party_id = $5,
			paid_from_id = $6, paid_to_id = $7, paid_amount = $8, received_amount = $9,
			reference_no = $10, reference_date = $11, posting_date = $12,
			docstatus = $13
		WHERE id = $1 AND tenant_id = $2`

	_, err := r.pool.Exec(ctx, query,
		pe.ID, tenantID,
		pe.PaymentType, pe.PartyType, pe.PartyID,
		pe.PaidFromID, pe.PaidToID, pe.PaidAmount, pe.ReceivedAmount,
		pe.ReferenceNo, pe.ReferenceDate, pe.PostingDate,
		pe.DocStatus,
	)
	return err
}

func (r *PaymentEntryRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM payment_entries WHERE id = $1 AND tenant_id = $2`
	_, err := r.pool.Exec(ctx, query, id, tenantID)
	return err
}

func (r *PaymentEntryRepository) GetNextPaymentNo(ctx context.Context, tenantID uuid.UUID, paymentType string) (string, error) {
	prefix := paymentTypePrefix(paymentType)

	var maxNo string
	query := `
		SELECT payment_no FROM payment_entries
		WHERE tenant_id = $1 AND payment_no LIKE $2
		ORDER BY payment_no DESC LIMIT 1`

	err := r.pool.QueryRow(ctx, query, tenantID, prefix+"%").Scan(&maxNo)
	if err != nil && err != pgx.ErrNoRows {
		return "", fmt.Errorf("get max payment no: %w", err)
	}

	if maxNo == "" {
		return fmt.Sprintf("%s-000001", prefix), nil
	}

	var seq int
	_, scanErr := fmt.Sscanf(maxNo, "%s-%d", new(string), &seq)
	if scanErr != nil {
		seqStr := strings.TrimPrefix(maxNo, prefix+"-")
		seq, _ = strconv.Atoi(seqStr)
	}
	return fmt.Sprintf("%s-%06d", prefix, seq+1), nil
}

func paymentTypePrefix(paymentType string) string {
	switch paymentType {
	case "receive":
		return "REC"
	case "pay":
		return "PAY"
	case "expense":
		return "EXP"
	case "interest":
		return "INT"
	case "transfer":
		return "TRF"
	default:
		return "DOC"
	}
}

func IsValidPaymentType(t string) bool {
	switch t {
	case "receive", "pay", "expense", "interest", "transfer":
		return true
	default:
		return false
	}
}
