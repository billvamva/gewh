package io

import (
	"encoding/csv"
	"io"
	"os"
	"strings"
	"sync/atomic"
)

var batchIdCounter uint64

// reader interface, can be extended to various formats for different data formats (e.g. json)
type Reader interface {
	// using filenames for now can be extended to use ReadWriter instead to extend data sources
	ReadRecords() <-chan []Record
}

// BatchReader reads CSV records in batches
type BatchReader struct {
	reader    *csv.Reader
	batchSize int
	batchChan chan []Record
	errChan   chan error
}

// Batch that holds batch id and string of values
type Batch struct {
	Id    uint64
	Value string
}

func NewBatch(value string) *Batch {
	return &Batch{
		Id:    atomic.AddUint64(&batchIdCounter, 1),
		Value: value,
	}
}

// NewBatchReader creates a new BatchReader
func NewBatchReader(filename string, batchSize int) (*BatchReader, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, ErrOpeningFile
	}

	reader := csv.NewReader(file)
	return &BatchReader{
		reader:    reader,
		batchSize: batchSize,
		batchChan: make(chan []Record),
		errChan:   make(chan error),
	}, nil
}

func (br *BatchReader) ReadRecords() (<-chan []Record, <-chan error) {
	go func() {
		defer close(br.batchChan)
		defer close(br.errChan)
		batch := make([]Record, 0, br.batchSize)

		for {
			record, err := br.reader.Read()
			if err == io.EOF {
				if len(batch) > 0 {
					br.batchChan <- batch
				}
				return
			}
			if err != nil {
				br.errChan <- err
			}

			batch = append(batch, record)
			if len(batch) == br.batchSize {
				br.batchChan <- batch
				batch = make([]Record, 0, br.batchSize)
			}
		}
	}()

	return br.batchChan, br.errChan
}

func CombineRecords(records []Record) *Batch {
	var combinedRecords []string
	for _, record := range records {
		combinedRecords = append(combinedRecords, strings.Join(record, ";"))
	}
	return NewBatch(strings.Join(combinedRecords, ","))
}
