package core

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type Queue interface {
	Enqueue(req Request)
	Dequeue() *Request
}

type Request struct {
	Id int
	Message Serialisable
	ResponseChan chan Serialisable
	Ctx context.Context
}

type MessageQueue struct {
	requests []*Request
	lock sync.Mutex
}

var requestPool = sync.Pool{
	New: func() interface{} {
		return &Request{}
	},
}


func NewMessageQueue() *MessageQueue {
	return &MessageQueue{
		requests: []*Request{},
	}
}

func NewRequest(id int, message Serialisable, ctx context.Context) *Request {
	responseChan := make(chan Serialisable)

	return &Request{
		id,
		message,
		responseChan,
		ctx,
	}
} 

func GetRequestFromPool() *Request {
	return requestPool.Get().(*Request)
}

func (q *MessageQueue) Enqueue(req *Request) {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.requests = append(q.requests, req)

}

func (q *MessageQueue) Dequeue() *Request {
	q.lock.Lock()
	defer q.lock.Unlock()
	if len(q.requests) == 0 {
		return nil
	}
	req := q.requests[0]
	q.requests = q.requests[1:]
	return req
}

func Worker(id int, queue Queue) {
	for {
		req := queue.Dequeue()
		if req == nil {
			time.Sleep(100 * time.Millisecond)
		}
		select {
		case <-req.Ctx.Done():
			fmt.Printf("Worker %d: Request %d cancelled\n", id, req.Id)
			continue // Skip this request if the context is done
		default:
			fmt.Printf("Worker %d: Processing Request %d\n", id, req.Id)
		}
		fmt.Printf("Worker %d: Processing Request %d\n", id, req.Id)

		if req.Message.buf == nil {
			log.Printf("Worker %d: Request %d has a nil buffer", id, req.Id)
			continue // Skip this request and continue with the next one
		}

		decodedByteFields, err := req.Message.Decode()
		if err != nil {
			log.Fatalf("Could not decode data on request, %v",err)
		}
		decodedBinaryRepresentation := FormatDecodedFields(decodedByteFields)
		processMessage(&decodedBinaryRepresentation)
		req.Message.BinaryRepresentationToByteFields(&decodedBinaryRepresentation)

		select {
		case req.ResponseChan <- req.Message:
		case <-req.Ctx.Done():
			fmt.Printf("Worker %d: Request %d cancelled before sending response\n", id, req.Id)
		}
	}
}

func processMessage(formattedDecodedFields *BinaryRepresentation) {
	// reformatting message - have to take into consideration changing message length
	formattedDecodedFields.message = []byte("Updated Message")
}