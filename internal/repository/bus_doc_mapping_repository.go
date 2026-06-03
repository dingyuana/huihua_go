package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"huihua/finance/internal/model"
)

// SystemTenantID is the default system tenant used for seed/shared mapping data.
var SystemTenantID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

type BusDocMappingRepository struct {
	pool *pgxpool.Pool
}

func NewBusDocMappingRepository(pool *pgxpool.Pool) *BusDocMappingRepository {
	return &BusDocMappingRepository{pool: pool}
}

// FindMapping looks up an active mapping by (doc_type, condition_key).
// First tries the specific tenant, then falls back to the system tenant.
// Returns nil if not found.
func (r *BusDocMappingRepository) FindMapping(ctx context.Context, tenantID uuid.UUID, docType, conditionKey string) (*model.BusDocMapping, error) {
	if conditionKey == "" {
		conditionKey = "default"
	}

	// Try the specific tenant first
	m, err := r.findMappingForTenant(ctx, tenantID, docType, conditionKey)
	if err != nil {
		return nil, err
	}
	if m != nil {
		return m, nil
	}

	// Fallback to system tenant if different from the requested tenant
	if tenantID != SystemTenantID {
		return r.findMappingForTenant(ctx, SystemTenantID, docType, conditionKey)
	}
	return nil, nil
}

func (r *BusDocMappingRepository) findMappingForTenant(ctx context.Context, tenantID uuid.UUID, docType, conditionKey string) (*model.BusDocMapping, error) {
	query := `
		SELECT id, tenant_id, doc_type, condition_key, condition_label,
		       debit_account_id, debit_subject_code, debit_subject_name,
		       credit_account_id, credit_subject_code, credit_subject_name,
		       is_active, sort_order, created_at, updated_at
		FROM bus_doc_mapping
		WHERE tenant_id = $1
		  AND doc_type = $2
		  AND condition_key = $3
		  AND is_active = TRUE
		ORDER BY sort_order ASC
		LIMIT 1`

	m := &model.BusDocMapping{}
	err := r.pool.QueryRow(ctx, query, tenantID, docType, conditionKey).Scan(
		&m.ID, &m.TenantID, &m.DocType, &m.ConditionKey, &m.ConditionLabel,
		&m.DebitAccountID, &m.DebitSubjectCode, &m.DebitSubjectName,
		&m.CreditAccountID, &m.CreditSubjectCode, &m.CreditSubjectName,
		&m.IsActive, &m.SortOrder, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf("find bus_doc_mapping: %w", err)
	}
	return m, nil
}

// FindDefaultMapping returns the default mapping for a doc_type, used as
// fallback when no condition_key matches.
func (r *BusDocMappingRepository) FindDefaultMapping(ctx context.Context, tenantID uuid.UUID, docType string) (*model.BusDocMapping, error) {
	return r.FindMapping(ctx, tenantID, docType, "default")
}
