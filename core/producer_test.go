package core

import (
	"context"
	"fmt"
	"testing"
	"time"
)

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
	q1 := make(chan *Request, MAX_QUEUE)
	q2 := make(chan *Request, MAX_QUEUE)
	d1 := NewDispatcher(1, MAX_WORKER)
	d1.AddQueue(q1)
	d2 := NewDispatcher(2, MAX_WORKER)
	d2.AddQueue(q2)

	p.Subscribe(d1)
	p.Subscribe(d2)

	if len(p.subs) != 2 {
		t.Errorf("Expected 2 subscribers, got %d", len(p.subs))
	}
}

func TestStart(t *testing.T) {
	t.Run("Testing Basic Start Functionality", func(t *testing.T) {
		p := NewProducer()
		q := make(chan *Request, MAX_QUEUE)
		d := NewDispatcher(1, MAX_WORKER)
		d.AddQueue(q)
		p.Subscribe(d)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go p.Start(ctx)
		// Simulate a dispatcher being done
		p.doneListener <- 1

		// Give some time for the goroutine to process
		time.Sleep(10 * time.Millisecond)

		if len(p.subs) != 0 {
			t.Errorf("Expected 0 subscribers after one is done, got %d", len(p.subs))
		}

		time.Sleep(10 * time.Millisecond) // Give some time for the goroutine to exit
	})
	t.Run("Testing Context cancellation", func(t *testing.T) {
		p := NewProducer()
		q := make(chan *Request, MAX_QUEUE)
		d := NewDispatcher(1, MAX_WORKER)
		d.AddQueue(q)
		p.Subscribe(d)
		ctx, cancel := context.WithCancel(context.Background())

		done := make(chan bool)

		go func() {
			p.Start(ctx)
			done <- true
		}()
		cancel()
		// Wait for Start to return
		select {
		case <-done:
			// Success: Start returned
		case <-time.After(100 * time.Millisecond):
			t.Errorf("Start did not return after context cancellation")
		}

		// Check if doneListener channel is closed
		_, open := <-p.doneListener
		if open {
			t.Errorf("doneListener channel was not closed")
		}
	})
}

func TestBroadcast(t *testing.T) {
	p := NewProducer(WithBroadcastTimeout[any](50 * time.Millisecond))

	q := make(chan *Request, 2)
	d := NewDispatcher(2, MAX_WORKER)
	d.AddQueue(q)
	p.Subscribe(d)

	req := &Request{}
	ctx := context.Background()

	t.Run("Single Broadcast", func(t *testing.T) {
		p.Broadcast(ctx, req)

		// Wait a short time for the goroutine to complete
		time.Sleep(10 * time.Millisecond)

		select {
		case received := <-q:
			if received != req {
				t.Errorf("Received request does not match sent request")
			}
		default:
			t.Errorf("No request received in the queue")
		}

		flushChannel(q)
	})

	t.Run("Multiple Broadcasts", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			p.Broadcast(ctx, req)
		}

		// Wait a short time for the goroutines to complete
		time.Sleep(10 * time.Millisecond)

		count := 0
		for i := 0; i < 3; i++ {
			select {
			case <-q:
				count++
			default:
				// Queue is empty
			}
		}

		if count != 2 {
			t.Errorf("Expected 2 items in queue, got %d", count)
		}
		flushChannel(q)
	})

	t.Run("Broadcast Timeout", func(t *testing.T) {
		// Fill the queue
		for i := 0; i < cap(q); i++ {
			q <- req
		}

		fmt.Printf("Queue capacity: %d, current length: %d\n", cap(q), len(q))

		start := time.Now()
		p.Broadcast(ctx, req)
		duration := time.Since(start)

		fmt.Printf("Broadcast duration: %v\n", duration)

		if duration < 45*time.Millisecond {
			t.Errorf("Expected broadcast to take at least 45ms, but it took %v", duration)
		}

		flushChannel(q)
	})
}

func flushChannel(ch chan *Request) {
	for len(ch) > 0 {
		<-ch
	}
}
