package event

import (
	"time"

	"github.com/google/uuid"
)

// Event names — keep stable for routing and observability.
const (
	EventVoucherRequested = "voucher.requested"
	EventAuditLog         = "audit.log"
	EventSettlementLog    = "settlement.log"
)

// VoucherRequestedEvent is published when a business operation needs a
// voucher generated (e.g. after confirming a sales invoice, or after an
// advance allocation). The auto-voucher-generation service subscribes to
// this and produces the corresponding journal entry.
type VoucherRequestedEvent struct {
	OccurredAt    time.Time
	TenantID      uuid.UUID
	CompanyID     uuid.UUID
	SourceType    string // sale_invoice | purchase_invoice | advance_receipt | advance_payment | cash_discount
	SourceID      uuid.UUID
	Action        string // confirm | allocate | discount
	UserID        uuid.UUID
	Payload       map[string]interface{}
}

func (VoucherRequestedEvent) EventName() string { return EventVoucherRequested }

// AuditLogEvent is published when an auditable action occurs.
// The audit repository subscriber writes a row to audit_logs.
//
// Replaces the inline auditRepo.Create calls that were scattered across
// services.
type AuditLogEvent struct {
	OccurredAt  time.Time
	TenantID    uuid.UUID
	Action      string // e.g. "invoice_status_change", "payment_rollback_on_voucher_delete"
	ObjectType  string // e.g. "sales_invoice", "payment_entry"
	ObjectID    uuid.UUID
	ActorID     uuid.UUID
	ActorName   *string
	Reason      *string
	OldValues   map[string]interface{}
	NewValues   map[string]interface{}
}

func (AuditLogEvent) EventName() string { return EventAuditLog }

// SettlementLogEvent is published when a settlement (write-off or reversal)
// occurs. The settlement-log repository subscriber writes the immutable log row.
//
// Replaces inline LogWriteOff calls and makes settlement events first-class
// for future analytics/observability.
type SettlementLogEvent struct {
	OccurredAt         time.Time
	TenantID           uuid.UUID
	SourceType         string // payment_allocation | advance_allocation | manual_reversal
	SourceID           uuid.UUID
	DocType            string // sales_invoice | ar_invoice | ap_invoice | advance_receipt | advance_payment
	DocID              uuid.UUID
	Direction          string // debit | credit
	Amount             string // decimal as string
	OutstandingBefore  string
	OutstandingAfter   string
	IsReversal         bool
	ReversedLogID      *uuid.UUID
	CreatedBy          *uuid.UUID
}

func (SettlementLogEvent) EventName() string { return EventSettlementLog }
