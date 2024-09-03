package core

// single queue message broker, to be expanded
type RequestQueue chan *Request

// dispatches requests to available workers - interface with workers
type Dispatcher struct {
	id         uint64
	WorkerPool chan chan *Request // A pool of workers channels that are registered with the dispatcher
	maxWorkers int                // maxWorker count
	queue      RequestQueue       // where the dispatcher will get the requests from
	quit       chan bool          // bool to stop the dispatcher
}

// creates NewDispatcher
func NewDispatcher(id uint64, maxWorkers int) *Dispatcher {
	pool := make(chan chan *Request, maxWorkers)

	return &Dispatcher{id: id, WorkerPool: pool, maxWorkers: maxWorkers}
}

func (d *Dispatcher) AddQueue(queue RequestQueue) {
	d.queue = queue
}

// starting n number of workers
func (d *Dispatcher) Run(p DataProcessor) {
	for i := 0; i < d.maxWorkers; i++ {
		worker := NewWorker(d.WorkerPool)
		worker.Start(i, p)
	}

	go d.dispatch()
}

// goroutine to dispatch requests to workers
func (d *Dispatcher) dispatch() {
	for {
		select {
		case req := <-d.queue:
			go func(req *Request) {
				requestChannel := <-d.WorkerPool

				requestChannel <- req
			}(req)
		case <-d.quit:
			return
		}
	}
}

func (d *Dispatcher) Stop() {
	go func() {
		d.quit <- true
	}()
}
