package core

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type subId uint64

const (
	defaultBroadcastTimeout = time.Minute
)

// manages dispatcher subscriptions and broadcasting requests 
type Producer struct {
	sync.RWMutex
	subs 	map[subId]*Dispatcher
	nextID subId
	doneListener chan subId
	broadcastTimeout time.Duration
}

type ProducerOpt func(*Producer)

func WithBroadcastTimeout[T any](timeout time.Duration) ProducerOpt {
	return func(ep *Producer) {
		ep.broadcastTimeout = timeout
	}
}

// creates new producer with options
func NewProducer (opts ...ProducerOpt) *Producer {
	producer := &Producer{
		subs: make(map[subId]*Dispatcher),
		doneListener: make(chan subId, 100),
		broadcastTimeout: defaultBroadcastTimeout,
	}
	for _, opt := range opts {
		opt(producer)
	}

	return producer
}

// Start begins listening for dispatcher cancelation requests or context cancelation.
func (ep *Producer) Start(ctx context.Context) {
	for {
		select {
		case id := <-ep.doneListener:
			ep.Lock()
			if dp, exists := ep.subs[id]; exists {
				dp.Stop()
				delete(ep.subs, id)
			}
			ep.Unlock()
		case <-ctx.Done():
			close(ep.doneListener)
			return
		}
	}
}

// Dispatcher subcribes to Producer, listens to requests emitted by Producer
func (ep *Producer) Subscribe(dp *Dispatcher) {
	ep.Lock()
	defer ep.Unlock()
	id := ep.nextID
	ep.subs[id] = dp
}


// Producer broadcasts requests to all listening dispatchers
func (ep *Producer) Broadcast(ctx context.Context, req *Request) {
	ep.RLock()
	defer ep.RUnlock()
	var wg sync.WaitGroup
	for _, sub := range ep.subs {
		wg.Add(1)
		go func (listener *Dispatcher, w *sync.WaitGroup) {
			defer w.Done()
			select {
			case listener.queue <- req:
			case <- time.After(ep.broadcastTimeout):
				fmt.Print("Broadcast to listener timed out.")	
			case <-ctx.Done():
			}
		} (sub, &wg)
	}
}