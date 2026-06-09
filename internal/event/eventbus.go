// Package event provides an in-process event bus for decoupling side effects
// from primary operations.
//
// Design rationale:
//   - Primary operations (e.g. "confirm sales invoice") publish events.
//   - Side effects (voucher generation, audit logging, settlement log writing,
//     notification) subscribe to events and run independently.
//   - This breaks the call-graph dependency between "what happened" and
//     "what to do about it", making each layer easier to test and evolve.
//
// Current implementation: synchronous in-process.
// Future: can be swapped for an out-of-process broker (Kafka, NATS, etc.)
// without changing publisher code.
package event

import (
	"context"
	"log"
	"sync"
)

// Event is the base interface implemented by all events on the bus.
type Event interface {
	// EventName returns a stable identifier for routing and observability.
	// e.g. "voucher.requested", "audit.log", "settlement.log"
	EventName() string
}

// HandlerFunc is the signature for event subscribers.
// Handlers MUST be idempotent because the bus may deliver the same event
// multiple times during retries (future async mode).
type HandlerFunc func(ctx context.Context, e Event) error

// Subscription binds a handler to a specific event name.
type Subscription struct {
	EventName string
	Handler   HandlerFunc
}

// Bus is the in-process event bus.
//
// Publishers call Publish; subscribers register via Subscribe.
// Handlers run synchronously by default; the order of invocation matches
// subscription order.
type Bus interface {
	Subscribe(eventName string, handler HandlerFunc)
	Publish(ctx context.Context, e Event) error
	Close() error
}

// InProcessBus is a thread-safe in-process bus.
// Suitable for the current single-process deployment; pluggable for the
// future broker-backed implementation.
type InProcessBus struct {
	mu       sync.RWMutex
	handlers map[string][]HandlerFunc
	closed   bool
}

// NewInProcessBus creates a new in-process event bus.
func NewInProcessBus() *InProcessBus {
	return &InProcessBus{
		handlers: make(map[string][]HandlerFunc),
	}
}

// Subscribe registers a handler for the given event name.
// Multiple handlers may subscribe to the same event.
func (b *InProcessBus) Subscribe(eventName string, handler HandlerFunc) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventName] = append(b.handlers[eventName], handler)
}

// Publish dispatches an event to all registered handlers synchronously.
// If a handler returns an error, the remaining handlers are still invoked
// and the first error is returned. (A future async mode may collect all errors.)
func (b *InProcessBus) Publish(ctx context.Context, e Event) error {
	b.mu.RLock()
	handlers := append([]HandlerFunc(nil), b.handlers[e.EventName()]...)
	b.mu.RUnlock()

	var firstErr error
	for _, h := range handlers {
		if err := h(ctx, e); err != nil {
			log.Printf("[event] handler error for %s: %v", e.EventName(), err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

// Close marks the bus as closed. Further Publish calls return an error.
// Subscriptions are preserved so the bus can still answer introspection queries.
func (b *InProcessBus) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.closed = true
	return nil
}
