package service

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"huihua/finance/internal/model"
	"huihua/finance/internal/repository"
)

// AuditService provides business logic for recording audit log entries.
type AuditService struct {
	repo *repository.AuditRepository
}

// NewAuditService creates a new AuditService.
func NewAuditService(repo *repository.AuditRepository) *AuditService {
	return &AuditService{repo: repo}
}

// LogCreate records a create action in the audit log.
func (s *AuditService) LogCreate(ctx context.Context, tenantID, actorID uuid.UUID, actorName, objectType string, objectID uuid.UUID, data json.RawMessage) error {
	log := &model.AuditLog{
		Action:        "create",
		ObjectType:    objectType,
		ObjectID:      objectID,
		ActorID:       actorID,
		ActorName:     &actorName,
		ChangedFields: data,
	}
	return s.repo.Create(ctx, tenantID, log)
}

// LogUpdate records an update action in the audit log.
// changedFields format: {"field_name": ["old_value", "new_value"]}
func (s *AuditService) LogUpdate(ctx context.Context, tenantID, actorID uuid.UUID, actorName, objectType string, objectID uuid.UUID, changedFields json.RawMessage) error {
	log := &model.AuditLog{
		Action:        "update",
		ObjectType:    objectType,
		ObjectID:      objectID,
		ActorID:       actorID,
		ActorName:     &actorName,
		ChangedFields: changedFields,
	}
	return s.repo.Create(ctx, tenantID, log)
}

// LogDelete records a delete action in the audit log.
func (s *AuditService) LogDelete(ctx context.Context, tenantID, actorID uuid.UUID, actorName, objectType string, objectID uuid.UUID) error {
	log := &model.AuditLog{
		Action:     "delete",
		ObjectType: objectType,
		ObjectID:   objectID,
		ActorID:    actorID,
		ActorName:  &actorName,
	}
	return s.repo.Create(ctx, tenantID, log)
}

// LogSubmit records a submit action in the audit log.
func (s *AuditService) LogSubmit(ctx context.Context, tenantID, actorID uuid.UUID, actorName, objectType string, objectID uuid.UUID) error {
	log := &model.AuditLog{
		Action:     "submit",
		ObjectType: objectType,
		ObjectID:   objectID,
		ActorID:    actorID,
		ActorName:  &actorName,
	}
	return s.repo.Create(ctx, tenantID, log)
}

// LogCancel records a cancel action in the audit log with a reason stored in metadata.
func (s *AuditService) LogCancel(ctx context.Context, tenantID, actorID uuid.UUID, actorName, objectType string, objectID uuid.UUID, reason string) error {
	metadata, _ := json.Marshal(map[string]string{"reason": reason})
	log := &model.AuditLog{
		Action:     "cancel",
		ObjectType: objectType,
		ObjectID:   objectID,
		ActorID:    actorID,
		ActorName:  &actorName,
		Metadata:   metadata,
	}
	return s.repo.Create(ctx, tenantID, log)
}

// LogReverse records a reverse action in the audit log with the reversed entry ID stored in metadata.
func (s *AuditService) LogReverse(ctx context.Context, tenantID, actorID uuid.UUID, actorName, objectType string, objectID uuid.UUID, reversedID uuid.UUID) error {
	metadata, _ := json.Marshal(map[string]string{"reversed_id": reversedID.String()})
	log := &model.AuditLog{
		Action:     "reverse",
		ObjectType: objectType,
		ObjectID:   objectID,
		ActorID:    actorID,
		ActorName:  &actorName,
		Metadata:   metadata,
	}
	return s.repo.Create(ctx, tenantID, log)
}

// ListByTenant retrieves audit logs for the given tenant with filters.
func (s *AuditService) ListByTenant(ctx context.Context, tenantID uuid.UUID, filter repository.AuditFilter) ([]model.AuditLog, error) {
	return s.repo.ListByTenant(ctx, tenantID, filter)
}

// GetByObject retrieves audit logs for a specific object.
func (s *AuditService) GetByObject(ctx context.Context, tenantID uuid.UUID, objectType string, objectID uuid.UUID) ([]model.AuditLog, error) {
	return s.repo.GetByObject(ctx, tenantID, objectType, objectID)
}
