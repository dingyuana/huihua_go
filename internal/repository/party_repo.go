package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"huihua/finance/internal/model"
)

// PartyRepository handles parties table operations.
type PartyRepository struct {
	pool *pgxpool.Pool
}

// NewPartyRepository creates a new PartyRepository.
func NewPartyRepository(pool *pgxpool.Pool) *PartyRepository {
	return &PartyRepository{pool: pool}
}

// Create inserts a new party.
func (r *PartyRepository) Create(ctx context.Context, tenantID uuid.UUID, p *model.Party) (*model.Party, error) {
	p.ID = uuid.New()
	p.TenantID = tenantID
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	if !p.IsActive {
		p.IsActive = true
	}
	if p.Source == "" {
		p.Source = "manual"
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO parties (id, tenant_id, party_type, name, tax_number, source, code, bank_name, bank_account,
			contact_name, contact_phone, credit_limit, payment_days, is_active, ar_account_id, ap_account_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)`,
		p.ID, p.TenantID, p.PartyType, p.Name, p.TaxNumber, p.Source, p.Code, p.BankName, p.BankAccount,
		p.ContactName, p.ContactPhone, p.CreditLimit, p.PaymentDays, p.IsActive,
		p.ArAccountID, p.ApAccountID, p.CreatedAt, p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// List retrieves all parties for a tenant.
func (r *PartyRepository) List(ctx context.Context, tenantID uuid.UUID) ([]model.Party, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, party_type, name, tax_number, bank_name, bank_account,
			contact_name, contact_phone, credit_limit, payment_days, is_active,
			ar_account_id, ap_account_id, created_at, updated_at
		FROM parties WHERE tenant_id = $1 AND is_active = TRUE ORDER BY name`,
		tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parties []model.Party
	for rows.Next() {
		var p model.Party
		if err := rows.Scan(&p.ID, &p.TenantID, &p.PartyType, &p.Name, &p.TaxNumber, &p.BankName, &p.BankAccount,
			&p.ContactName, &p.ContactPhone, &p.CreditLimit, &p.PaymentDays, &p.IsActive,
			&p.ArAccountID, &p.ApAccountID, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		parties = append(parties, p)
	}
	return parties, rows.Err()
}

// ListByType retrieves parties by type for a tenant.
// ListByType retrieves parties by type.
func (r *PartyRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*model.Party, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, party_type, name, tax_number, bank_name, bank_account,
		       contact_name, contact_phone, credit_limit, payment_days, is_active,
		       ar_account_id, ap_account_id, created_at, updated_at
		FROM parties WHERE id = $1 AND tenant_id = $2`,
		id, tenantID)
	var p model.Party
	err := row.Scan(&p.ID, &p.TenantID, &p.PartyType, &p.Name, &p.TaxNumber, &p.BankName, &p.BankAccount,
		&p.ContactName, &p.ContactPhone, &p.CreditLimit, &p.PaymentDays, &p.IsActive,
		&p.ArAccountID, &p.ApAccountID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PartyRepository) ListByType(ctx context.Context, tenantID uuid.UUID, partyType string) ([]model.Party, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, party_type, name, tax_number, bank_name, bank_account,
			contact_name, contact_phone, credit_limit, payment_days, is_active,
			ar_account_id, ap_account_id, created_at, updated_at
		FROM parties WHERE tenant_id = $1 AND party_type = $2 AND is_active = TRUE ORDER BY name`,
		tenantID, partyType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parties []model.Party
	for rows.Next() {
		var p model.Party
		if err := rows.Scan(&p.ID, &p.TenantID, &p.PartyType, &p.Name, &p.TaxNumber, &p.BankName, &p.BankAccount,
			&p.ContactName, &p.ContactPhone, &p.CreditLimit, &p.PaymentDays, &p.IsActive,
			&p.ArAccountID, &p.ApAccountID, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		parties = append(parties, p)
	}
	return parties, rows.Err()
}

// Update updates a party.
func (r *PartyRepository) Update(ctx context.Context, tenantID, id uuid.UUID, p *model.Party) error {
	p.UpdatedAt = time.Now()
	_, err := r.pool.Exec(ctx, `
		UPDATE parties SET party_type = $3, name = $4, tax_number = $5, bank_name = $6, bank_account = $7,
			contact_name = $8, contact_phone = $9, credit_limit = $10, payment_days = $11, is_active = $12,
			ar_account_id = $13, ap_account_id = $14, updated_at = $15
		WHERE tenant_id = $1 AND id = $2`,
		tenantID, id, p.PartyType, p.Name, p.TaxNumber, p.BankName, p.BankAccount,
		p.ContactName, p.ContactPhone, p.CreditLimit, p.PaymentDays, p.IsActive,
		p.ArAccountID, p.ApAccountID, p.UpdatedAt)
	return err
}

// Delete soft-deletes a party (sets is_active = false).
func (r *PartyRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE parties SET is_active = FALSE, updated_at = $3
		WHERE tenant_id = $1 AND id = $2`, tenantID, id, time.Now())
	return err
}

// ExistsByNameAndType checks if a party with same name+type exists for tenant.
func (r *PartyRepository) ExistsByNameAndType(ctx context.Context, tenantID uuid.UUID, name, partyType string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM parties WHERE tenant_id = $1 AND name = $2 AND party_type = $3 AND is_active = TRUE)`,
		tenantID, name, partyType).Scan(&exists)
	return exists, err
}

// GetByName looks up a party by name within a tenant. Returns nil if not found.
func (r *PartyRepository) GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*model.Party, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, party_type, name, tax_number, bank_name, bank_account,
		       contact_name, contact_phone, credit_limit, payment_days, is_active,
		       ar_account_id, ap_account_id, created_at, updated_at
		FROM parties WHERE tenant_id = $1 AND name = $2 AND is_active = TRUE
		LIMIT 1`,
		tenantID, name)
	var p model.Party
	err := row.Scan(&p.ID, &p.TenantID, &p.PartyType, &p.Name, &p.TaxNumber, &p.BankName, &p.BankAccount,
		&p.ContactName, &p.ContactPhone, &p.CreditLimit, &p.PaymentDays, &p.IsActive,
		&p.ArAccountID, &p.ApAccountID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, nil // not found
	}
	return &p, nil
}

// CheckTaxNumberDuplicate checks for duplicate tax numbers within a tenant.
func (r *PartyRepository) CheckTaxNumberDuplicate(ctx context.Context, tenantID uuid.UUID, taxNumbers []string) (map[string]bool, error) {
	result := make(map[string]bool)
	for _, tn := range taxNumbers {
		var exists bool
		err := r.pool.QueryRow(ctx, `
			SELECT EXISTS(SELECT 1 FROM parties WHERE tenant_id = $1 AND tax_number = $2 AND is_active = TRUE)`,
			tenantID, tn).Scan(&exists)
		if err != nil {
			return nil, err
		}
		result[tn] = exists
	}
	return result, nil
}

// CreateBatch inserts multiple parties in one call.
func (r *PartyRepository) CreateBatch(ctx context.Context, tenantID uuid.UUID, parties []model.Party) error {
	batch := &pgx.Batch{}
	for _, p := range parties {
		p.ID = uuid.New()
		p.TenantID = tenantID
		p.CreatedAt = time.Now()
		p.UpdatedAt = time.Now()
		if !p.IsActive {
			p.IsActive = true
		}
		batch.Queue(`INSERT INTO parties (id, tenant_id, party_type, name, tax_number, bank_name, bank_account,
			contact_name, contact_phone, credit_limit, payment_days, is_active, ar_account_id, ap_account_id, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16) ON CONFLICT DO NOTHING`,
			p.ID, p.TenantID, p.PartyType, p.Name, p.TaxNumber, p.BankName, p.BankAccount,
			p.ContactName, p.ContactPhone, p.CreditLimit, p.PaymentDays, p.IsActive,
			p.ArAccountID, p.ApAccountID, p.CreatedAt, p.UpdatedAt)
	}
	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()
	for i := 0; i < len(parties); i++ {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}
