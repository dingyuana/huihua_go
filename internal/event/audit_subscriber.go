package event

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"huihua/finance/internal/model"
)

// AuditWriter is the minimal interface AuditLogSubscriber needs to persist
// audit log entries. Implemented by *repository.AuditRepository in main.go.
type AuditWriter interface {
	Create(ctx context.Context, tenantID uuid.UUID, log *model.AuditLog) error
}

// AuditLogSubscriber subscribes to EventAuditLog and writes audit rows.
//
// Use this to replace inline _ = s.auditRepo.Create(ctx, tenantID, auditLog)
// calls in services. After subscribing, services only need to publish an
// AuditLogEvent.
func AuditLogSubscriber(writer AuditWriter) HandlerFunc {
	return func(ctx context.Context, e Event) error {
		ev, ok := e.(AuditLogEvent)
		if !ok {
			return nil
		}
		changedFields, _ := json.Marshal(map[string]map[string]interface{}{
			"old": ev.OldValues,
			"new": ev.NewValues,
		})
		metadata, _ := json.Marshal(map[string]interface{}{
			"action": ev.Action,
			"reason": ev.Reason,
		})
		entry := &model.AuditLog{
			ID:            uuid.New(),
			Action:        ev.Action,
			ObjectType:    ev.ObjectType,
			ObjectID:      ev.ObjectID,
			TenantID:      ev.TenantID,
			ActorID:       ev.ActorID,
			ActorName:     ev.ActorName,
			ChangedFields: changedFields,
			Metadata:      metadata,
		}
		return writer.Create(ctx, ev.TenantID, entry)
	}
}
