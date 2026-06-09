package event

import (
	"context"
	"errors"
	"sync"
	"testing"
)

type testEvent struct {
	Name  string
	Value int
}

func (testEvent) EventName() string { return "test.event" }

func TestPublishDeliversToAllSubscribers(t *testing.T) {
	bus := NewInProcessBus()
	var got1, got2 int
	bus.Subscribe("test.event", func(_ context.Context, e Event) error {
		got1 = e.(testEvent).Value
		return nil
	})
	bus.Subscribe("test.event", func(_ context.Context, e Event) error {
		got2 = e.(testEvent).Value * 10
		return nil
	})
	if err := bus.Publish(context.Background(), testEvent{Value: 5}); err != nil {
		t.Fatalf("publish: %v", err)
	}
	if got1 != 5 {
		t.Errorf("handler1 got %d, want 5", got1)
	}
	if got2 != 50 {
		t.Errorf("handler2 got %d, want 50", got2)
	}
}

func TestPublishNoSubscribersIsNoop(t *testing.T) {
	bus := NewInProcessBus()
	if err := bus.Publish(context.Background(), testEvent{Value: 1}); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestPublishReturnsFirstHandlerError(t *testing.T) {
	bus := NewInProcessBus()
	handlerErr := errors.New("handler failed")
	bus.Subscribe("test.event", func(_ context.Context, _ Event) error {
		return handlerErr
	})
	bus.Subscribe("test.event", func(_ context.Context, _ Event) error {
		return nil
	})
	err := bus.Publish(context.Background(), testEvent{Value: 1})
	if err == nil || !errors.Is(err, handlerErr) {
		t.Errorf("expected handlerErr, got %v", err)
	}
}

func TestConcurrentSubscribeAndPublish(t *testing.T) {
	bus := NewInProcessBus()
	var wg sync.WaitGroup
	var count int64
	var mu sync.Mutex
	for i := 0; i < 10; i++ {
		bus.Subscribe("test.event", func(_ context.Context, _ Event) error {
			mu.Lock()
			count++
			mu.Unlock()
			return nil
		})
	}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = bus.Publish(context.Background(), testEvent{Value: i})
		}()
	}
	wg.Wait()
	mu.Lock()
	defer mu.Unlock()
	if count != 1000 {
		t.Errorf("expected 1000 invocations (10 subs * 100 events), got %d", count)
	}
}
