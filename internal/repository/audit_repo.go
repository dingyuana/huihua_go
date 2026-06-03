package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"huihua/finance/internal/model"
)

// AuditFilter defines filtering criteria for listing audit logs.
type AuditFilter struct {
	ObjectType string
	ObjectID   uuid.UUID
	ActorID    uuid.UUID
	StartTime  *time.Time
	EndTime    *time.Time
	Limit      int
	Offset     int
}

// AuditRepository provides data access for the audit_logs table (append-only).
type AuditRepository struct {
	pool *pgxpool.Pool
}

// NewAuditRepository creates a new AuditRepository.
func NewAuditRepository(pool *pgxpool.Pool) *AuditRepository {
	return &AuditRepository{pool: pool}
}

// Create inserts a new audit log entry.
func (r *AuditRepository) Create(ctx context.Context, tenantID uuid.UUID, auditLog *model.AuditLog) error {
	query := `
		INSERT INTO audit_logs (id, action, object_type, object_id, tenant_id, actor_id, actor_name, changed_fields, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	if auditLog.ID == uuid.Nil {
		auditLog.ID = uuid.New()
	}
	if auditLog.CreatedAt.IsZero() {
		auditLog.CreatedAt = time.Now()
	}

	_, err := r.pool.Exec(ctx, query,
		auditLog.ID,
		auditLog.Action,
		auditLog.ObjectType,
		auditLog.ObjectID,
		tenantID,
		auditLog.ActorID,
		auditLog.ActorName,
		auditLog.ChangedFields,
		auditLog.Metadata,
		auditLog.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create audit log: %w", err)
	}

	auditLog.TenantID = tenantID
	return nil
}

// ListByTenant retrieves audit logs for the given tenant with optional filters.
func (r *AuditRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, filter AuditFilter) ([]model.AuditLog, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, fmt.Sprintf("tenant_id = $%d", argIdx))
	args = append(args, tenantID)
	argIdx++

	if filter.ObjectType != "" {
		conditions = append(conditions, fmt.Sprintf("object_type = $%d", argIdx))
		args = append(args, filter.ObjectType)
		argIdx++
	}

	if filter.ObjectID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("object_id = $%d", argIdx))
		args = append(args, filter.ObjectID)
		argIdx++
	}

	if filter.ActorID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("actor_id = $%d", argIdx))
		args = append(args, filter.ActorID)
		argIdx++
	}

	if filter.StartTime != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIdx))
		args = append(args, *filter.StartTime)
		argIdx++
	}

	if filter.EndTime != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIdx))
		args = append(args, *filter.EndTime)
		argIdx++
	}

	query := fmt.Sprintf(`
		SELECT id, action, object_type, object_id, tenant_id, actor_id, actor_name,
		       changed_fields, metadata, created_at
		FROM audit_logs
		WHERE %s
		ORDER BY created_at DESC`, strings.Join(conditions, " AND "))

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list audit logs: %w", err)
	}
	defer rows.Close()

	var logs []model.AuditLog
	for rows.Next() {
		var log model.AuditLog
		if err := rows.Scan(
			&log.ID, &log.Action, &log.ObjectType, &log.ObjectID, &log.TenantID,
			&log.ActorID, &log.ActorName, &log.ChangedFields, &log.Metadata, &log.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan audit log: %w", err)
		}
		logs = append(logs, log)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate audit logs: %w", err)
	}
	return logs, nil
}

// CreateTx inserts a new audit log entry within an existing transaction.
func (r *AuditRepository) CreateTx(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID, auditLog *model.AuditLog) error {
	query := `
		INSERT INTO audit_logs (id, action, object_type, object_id, tenant_id, actor_id, actor_name, changed_fields, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	if auditLog.ID == uuid.Nil {
		auditLog.ID = uuid.New()
	}
	if auditLog.CreatedAt.IsZero() {
		auditLog.CreatedAt = time.Now()
	}

	_, err := tx.Exec(ctx, query,
		auditLog.ID,
		auditLog.Action,
		auditLog.ObjectType,
		auditLog.ObjectID,
		tenantID,
		auditLog.ActorID,
		auditLog.ActorName,
		auditLog.ChangedFields,
		auditLog.Metadata,
		auditLog.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create audit log: %w", err)
	}

	auditLog.TenantID = tenantID
	return nil
}

// GetByObject retrieves audit logs for a specific object within the given tenant.
func (r *AuditRepository) GetByObject(ctx context.Context, tenantID uuid.UUID, objectType string, objectID uuid.UUID) ([]model.AuditLog, error) {
	query := `
		SELECT id, action, object_type, object_id, tenant_id, actor_id, actor_name,
		       changed_fields, metadata, created_at
		FROM audit_logs
		WHERE tenant_id = $1 AND object_type = $2 AND object_id = $3
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, tenantID, objectType, objectID)
	if err != nil {
		return nil, fmt.Errorf("get audit logs by object: %w", err)
	}
	defer rows.Close()

	var logs []model.AuditLog
	for rows.Next() {
		var log model.AuditLog
		if err := rows.Scan(
			&log.ID, &log.Action, &log.ObjectType, &log.ObjectID, &log.TenantID,
			&log.ActorID, &log.ActorName, &log.ChangedFields, &log.Metadata, &log.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan audit log: %w", err)
		}
		logs = append(logs, log)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate audit logs: %w", err)
	}
	return logs, nil
}
