package core

import (
	"log"
)

var (
	MAX_QUEUE = 20
	MAX_WORKER = 5
)


type RequestQueue chan *Request


type Worker struct {
	WorkerPool  chan chan *Request
	RequestChannel  chan *Request
	quit    	chan bool
}

func NewWorker(workerPool chan chan *Request) Worker {
	return Worker{
		WorkerPool: workerPool,
		RequestChannel: make(chan *Request),
		quit:       make(chan bool)}
}

func (w Worker) Start(id int) {
	go func() {
		for {
			// (re)register channel in worker pool (when processing has been performed)
			w.WorkerPool <- w.RequestChannel
			select {
			// receive request in worker channel
			case req := <- w.RequestChannel:
				go func(req *Request) {
					select {
					case <-req.Ctx.Done():
						log.Printf("Worker %d: Request %d cancelled\n", id, req.Id)
						return
					default:
						log.Printf("Worker %d: Processing Request %d\n", id, req.Id)
				
						if req.Message.buf == nil {
							log.Printf("Worker %d: Request %d has a nil buffer", id, req.Id)
							return // Skip this request and continue with the next one
						}
						req.Process()
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
