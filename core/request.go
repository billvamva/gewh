package core

import (
	"sync"
)

type Queue interface {
	Enqueue(req Request)
	Dequeue() *Request
}

type Request struct {
	Id int
	Message Serialisable
	ResponseChan chan Serialisable
}

type MessageQueue struct {
	requests []Request
	lock sync.Mutex
}

func NewMessageQueue() *MessageQueue {
	return &MessageQueue{
		requests: []Request{},
	}
}

func (q *MessageQueue) Enqueue(req Request) {
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
	return &req
}

// func Worker(id int, queue Queue) {
// 	for {
// 		req := queue.Dequeue()
// 		if req == nil {
// 			time.Sleep(100 * time.Millisecond)
// 		}
// 		fmt.Printf("Worker %d: Processing Request %d\n", id, req.Id)

// 		var decodedBinaryRepresentation BinaryRepresentation
// 		req.Message.Decode()
// 	}
// }