package processor

import (
	"gewh/core"
	"log"
	"sync"
)

var (
	batchResults     sync.Map
	BatchResultsLock sync.RWMutex
)

type Record struct {
	Min float64
	Max float64
	Avg float64
}

type Processor struct {
	mapFunc    MapFunc
	reduceFunc ReduceFunc
	sync.WaitGroup
}

func NewProcessor(mapFunc MapFunc, reduceFunc ReduceFunc) *Processor {
	return &Processor{
		mapFunc:    mapFunc,
		reduceFunc: reduceFunc,
	}
}

// adapter for process interface in core for request processing
func (p *Processor) Process(req *core.Request) error {
	p.Add(1)
	var err error

	go func(req *core.Request) {
		defer p.Done()
		err = ProcessRawData(req, p.mapFunc, p.reduceFunc)
	}(req)

	if err != nil {
		return err
	}
	return nil
}

// processing raw incoming requests
func ProcessRawData(req *core.Request, mapFunc MapFunc, reduceFunc ReduceFunc) error {
	// Steps 1-2: Decode the message and extract payload (unchanged)
	fields, err := req.Message.Decode()
	if err != nil {
		log.Printf("Could not decode data on request: %v, error: %v", req.Id, err)
		return err
	}
	var payload core.Payload
	payload.FromFields(fields)

	// Step 3: Map operation
	batchResults, err := mapReduceOptimized(&payload, mapFunc, reduceFunc)
	if err != nil {
		return err
	}

	// Step 6: Store batch results
	storeBatchResults(req.Id, batchResults)

	return nil
}

// group map and reduce so they are done in one operation
func mapReduceOptimized(data *core.Payload, mapFunc MapFunc, reduceFunc ReduceFunc) ([]KeyValue, error) {
	// Perform mapping
	kvs := mapFunc(data)

	// Group by key and reduce concurrently
	groups := make(map[string][][]byte)
	for _, kv := range kvs {
		groups[kv.Key] = append(groups[kv.Key], kv.Value)
	}

	results := make([]KeyValue, 0, len(groups))
	var wg sync.WaitGroup
	resultChan := make(chan KeyValue, len(groups))
	errChan := make(chan error, 1)

	for key, values := range groups {
		wg.Add(1)
		go func(k string, v [][]byte) {
			defer wg.Done()
			reduced, err := reduceFunc(v)
			if err != nil {
				select {
				case errChan <- err:
				default:
				}
				return
			}
			resultChan <- KeyValue{Key: k, Value: reduced}
		}(key, values)
	}

	go func() {
		wg.Wait()
		close(resultChan)
		close(errChan)
	}()

	for {
		select {
		case result, ok := <-resultChan:
			if !ok {
				return results, nil
			}
			results = append(results, result)
		case err := <-errChan:
			return nil, err
		}
	}
}

// aggreagating into records
func AggregateFinalResults(finalReduceFunc ReduceFunc) map[string]Record {
	// Retrieve all batch results
	allBatchResults := getAllBatchResults()

	// Group all batch results
	finalGroups := groupByKey(allBatchResults)

	// Perform final reduction
	finalResults := reduce(finalGroups, finalReduceFunc)

	results := make(map[string]Record)
	for _, data := range finalResults {
		results[data.Key] = ParseResult(data.Value)
	}

	return results
}

// storing results in memory
func storeBatchResults(batchId int, results []KeyValue) {
	BatchResultsLock.Lock()
	defer BatchResultsLock.Unlock()
	batchResults.Store(batchId, results)
}

// retrieving all with a thread safe map
func getAllBatchResults() []KeyValue {
	var allResults []KeyValue
	batchResults.Range(func(_, value interface{}) bool {
		results := value.([]KeyValue)
		allResults = append(allResults, results...)
		return true
	})
	return allResults
}
