package event

import (
	"context"
	"log"
)

// SettlementLogSubscriber is a no-op subscriber reserved for future use.
//
// As of Phase 1, settlement logs are written inline inside allocation
// transactions (see invoice_service.AllocateToPaymentEntry and
// advance_allocation_service.Allocate) for atomicity with the allocation.
// Publishing an additional event here would be a redundant write.
//
// The subscriber is registered so that future reconciliation imports or
// external settlement sources can write logs by simply publishing events
// without changing the subscriber list. The placeholder logs the event name
// at debug level for observability.
func SettlementLogSubscriber() HandlerFunc {
	return func(ctx context.Context, e Event) error {
		ev := e.(SettlementLogEvent)
		log.Printf("[event] settlement.log observed: doc=%s id=%v (no-op; inline path used)",
			ev.DocType, ev.DocID)
		return nil
	}
}
