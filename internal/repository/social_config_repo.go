package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"huihua/finance/internal/model"
)

// SocialConfigRepository provides data access for wage_social_config.
type SocialConfigRepository struct {
	pool *pgxpool.Pool
}

// NewSocialConfigRepository creates a new SocialConfigRepository.
func NewSocialConfigRepository(pool *pgxpool.Pool) *SocialConfigRepository {
	return &SocialConfigRepository{pool: pool}
}

// ListByTenant retrieves all social config records for a tenant.
func (r *SocialConfigRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]model.SocialConfig, error) {
	query := `
		SELECT id, tenant_id, insurance_type, insurance_name, company_rate, personal_rate, is_active, remark, created_at, updated_at
		FROM wage_social_config
		WHERE tenant_id = $1
		ORDER BY insurance_type`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list social config: %w", err)
	}
	defer rows.Close()

	var list []model.SocialConfig
	for rows.Next() {
		var c model.SocialConfig
		if err := rows.Scan(
			&c.ID, &c.TenantID, &c.InsuranceType, &c.InsuranceName,
			&c.CompanyRate, &c.PersonalRate, &c.IsActive,
			&c.Remark, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan social config: %w", err)
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

// GetByID retrieves a social config record by ID and tenant.
func (r *SocialConfigRepository) GetByID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*model.SocialConfig, error) {
	query := `
		SELECT id, tenant_id, insurance_type, insurance_name, company_rate, personal_rate, is_active, remark, created_at, updated_at
		FROM wage_social_config
		WHERE id = $1 AND tenant_id = $2`

	c := &model.SocialConfig{}
	err := r.pool.QueryRow(ctx, query, id, tenantID).Scan(
		&c.ID, &c.TenantID, &c.InsuranceType, &c.InsuranceName,
		&c.CompanyRate, &c.PersonalRate, &c.IsActive,
		&c.Remark, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get social config by id: %w", err)
	}
	return c, nil
}

// Update updates an existing social config record.
func (r *SocialConfigRepository) Update(ctx context.Context, tenantID uuid.UUID, config *model.SocialConfig) error {
	query := `
		UPDATE wage_social_config
		SET insurance_name = $1, company_rate = $2, personal_rate = $3, is_active = $4, remark = $5, updated_at = NOW()
		WHERE id = $6 AND tenant_id = $7`

	_, err := r.pool.Exec(ctx, query,
		config.InsuranceName, config.CompanyRate, config.PersonalRate,
		config.IsActive, config.Remark, config.ID, tenantID,
	)
	if err != nil {
		return fmt.Errorf("update social config: %w", err)
	}
	return nil
}

// InitDefaults inserts the 6 default social config records for a tenant.
// Uses ON CONFLICT DO NOTHING so it's safe to call multiple times.
func (r *SocialConfigRepository) InitDefaults(ctx context.Context, tenantID uuid.UUID) error {
	query := `
		INSERT INTO wage_social_config (id, tenant_id, insurance_type, insurance_name, company_rate, personal_rate)
		VALUES
			(gen_random_uuid(), $1, 'pension', '养老保险', 0.1600, 0.0800),
			(gen_random_uuid(), $1, 'medical', '医疗保险', 0.0800, 0.0200),
			(gen_random_uuid(), $1, 'unemployment', '失业保险', 0.0050, 0.0050),
			(gen_random_uuid(), $1, 'injury', '工伤保险', 0.0040, 0),
			(gen_random_uuid(), $1, 'maternity', '生育保险', 0.0080, 0),
			(gen_random_uuid(), $1, 'housing', '住房公积金', 0.1200, 0.1200)
		ON CONFLICT (tenant_id, insurance_type) DO NOTHING`

	_, err := r.pool.Exec(ctx, query, tenantID)
	if err != nil {
		return fmt.Errorf("init social config defaults: %w", err)
	}
	return nil
}
