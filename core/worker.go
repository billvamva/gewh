package core

import (
	"fmt"
)

var (
	MAX_QUEUE  = 20
	MAX_WORKER = 5
)

// worker that interacts with the message broker
type Worker struct {
	WorkerPool     chan chan *Request // each worker corresponds to a worker pool that holds request channels
	RequestChannel chan *Request      // worker's request channel
	quit           chan bool          // signal to quit the worker
}

// creates new worker
func NewWorker(workerPool chan chan *Request) Worker {
	return Worker{
		WorkerPool:     workerPool,
		RequestChannel: make(chan *Request),
		quit:           make(chan bool),
	}
}

// registers worker's request channel to the pool and waits for requests or quit signal on the request channel.
func (w Worker) Start(id int, p DataProcessor) {
	go func() {
		for {
			// (re)register channel in worker pool (when processing has been performed)
			w.WorkerPool <- w.RequestChannel
			select {
			// receive request in worker channel
			case req := <-w.RequestChannel:
				go func(req *Request) {
					select {
					case <-req.Ctx.Done():
						fmt.Printf("Worker %d: Request %d cancelled\n", id, req.Id)
						return
					default:
						fmt.Printf("Worker %d: Processing Request %d\n", id, req.Id)

						if req.Message.buf == nil {
							fmt.Printf("Worker %d: Request %d has a nil buffer", id, req.Id)
							return // Skip this request and continue with the next one
						}
						// processing implementation
						p.Process(req)
					}
				}(req)

			case <-w.quit:
				return
			}
		}
	}()
}

// Stop signals the worker to stop listening for work requests.
func (w Worker) Stop() {
	go func() {
		w.quit <- true
	}()
}
