package core

import (
	"context"
	"testing"
	"time"
)

// MockDispatcher for testing purposes
type MockDispatcher struct {
	queue      chan *Request
	stopCalled bool
}

func NewMockDispatcher() *MockDispatcher {
	return &MockDispatcher{
		queue: make(chan *Request, 1),
	}
}

func (md *MockDispatcher) Stop() {
	md.stopCalled = true
}

func TestNewProducer(t *testing.T) {
	p := NewProducer()
	if p.broadcastTimeout != defaultBroadcastTimeout {
		t.Errorf("Expected default broadcast timeout, got %v", p.broadcastTimeout)
	}

	customTimeout := 2 * time.Minute
	p = NewProducer(WithBroadcastTimeout[any](customTimeout))
	if p.broadcastTimeout != customTimeout {
		t.Errorf("Expected custom broadcast timeout of %v, got %v", customTimeout, p.broadcastTimeout)
	}
}

func TestSubscribe(t *testing.T) {
	p := NewProducer()
	d1 := NewMockDispatcher()
	d2 := NewMockDispatcher()

	p.Subscribe(d1)
	p.Subscribe(d2)

	if len(p.subs) != 2 {
		t.Errorf("Expected 2 subscribers, got %d", len(p.subs))
	}
}

func TestStart(t *testing.T) {
	p := NewProducer()
	d := NewMockDispatcher()
	p.Subscribe(d)

	ctx, cancel := context.WithCancel(context.Background())
	go p.Start(ctx)

	// Simulate a dispatcher being done
	p.doneListener <- p.nextID - 1

	// Give some time for the goroutine to process
	time.Sleep(10 * time.Millisecond)

	if len(p.subs) != 0 {
		t.Errorf("Expected 0 subscribers after one is done, got %d", len(p.subs))
	}

	cancel()
	time.Sleep(10 * time.Millisecond) // Give some time for the goroutine to exit
}

func TestBroadcast(t *testing.T) {
	p := NewProducer(WithBroadcastTimeout[any](100 * time.Millisecond))
	d1 := NewMockDispatcher()
	d2 := NewMockDispatcher()

	p.Subscribe(d1)
	p.Subscribe(d2)

	req := &Request{} // Assuming Request is defined elsewhere
	ctx := context.Background()

	p.Broadcast(ctx, req)

	// Check if both dispatchers received the request
	select {
	case <-d1.queue:
	case <-time.After(200 * time.Millisecond):
		t.Error("Dispatcher 1 didn't receive the request in time")
	}

	select {
	case <-d2.queue:
	case <-time.After(200 * time.Millisecond):
		t.Error("Dispatcher 2 didn't receive the request in time")
	}
}

func TestBroadcastTimeout(t *testing.T) {
	p := NewProducer(WithBroadcastTimeout[any](50 * time.Millisecond))
	d := NewMockDispatcher()
	p.Subscribe(d)

	req := &Request{}
	ctx := context.Background()

	// Block the dispatcher's queue
	d.queue <- req

	// This broadcast should timeout
	p.Broadcast(ctx, req)

	// The broadcast should have timed out without blocking
	select {
	case <-d.queue: // Try to receive the second request
		t.Error("Broadcast didn't timeout as expected")
	default:
		// This is the expected behavior
	}
}

