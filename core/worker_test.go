package core

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

var wg sync.WaitGroup

func TestRequestAndQueueHandling(t *testing.T) {
	t.Run("test worker to process reqs from queue", func(t *testing.T) {
		requests := []BinaryRepresentation{
			{Version: uint16(1), ClientId: uint16(1), Token: []byte("custom token"), Data: []byte("1. Hello, Worker!")},
			{Version: uint16(1), ClientId: uint16(2), Token: []byte("custom token"), Data: []byte("2. Hello, Worker!")},
			{Version: uint16(1), ClientId: uint16(3), Token: []byte("custom token"), Data: []byte("3. Hello, Worker!")},
		}
		ctx := context.TODO()
		reqStore := make([]*Request, len(requests))
		RequestQueue := make(chan *Request, MAX_QUEUE)
		for i, reqData := range requests {
			req := createAndFormatTestRequest(reqData, i, ctx)
			reqStore[i] = req
			RequestQueue <- reqStore[i]
		}

		// Define a DataProcessingFn that appends a string to the field
		cb := DataProcessingFn(func(field []byte) []byte {
			additionalData := []byte(" I have been added in processing!")
			field = append(field, additionalData...)
			return field
		})

		pool := make(chan chan *Request, 1)
		worker := NewWorker(pool)
		worker.Start(1, cb)

		go func() {
			for {
				select {
				case req := <-RequestQueue:
					// a job request has been received
					wg.Add(1)
					go func(req *Request) {
						// try to obtain a worker job channel that is available.
						// this will block until a worker is idle
						reqChannel := <-worker.WorkerPool

						// dispatch the job to the worker job channel
						reqChannel <- req
						wg.Done()
					}(req)
				default:
					return
				}
			}
		}()
		time.Sleep(time.Second * 2)
		wg.Wait()
		worker.Stop()

		for i, req := range reqStore {
			buffer := bytes.Buffer{}
			expectedString := fmt.Sprintf("%d. Hello, Worker! I have been added in processing!", i+1)
			buffer.WriteString(expectedString)
			got := string(getBinaryRepresentationFromSerialisable(req.Message).Data)
			want := buffer.String()
			if !strings.EqualFold(got, want) {
				t.Errorf("Expected %s, got %s", want, got)
			}
		}
	})

	t.Run("test worker to process req from queue and cancel req before processing", func(t *testing.T) {
		requests := []BinaryRepresentation{
			{Version: uint16(1), ClientId: uint16(1), Token: []byte("custom token"), Data: []byte("Hello, Worker!")},
		}
		parentCtx := context.Background()
		ctx, cancel := context.WithCancel(parentCtx)
		reqStore := make([]*Request, len(requests))
		RequestQueue := make(chan *Request, MAX_QUEUE)
		for i, reqData := range requests {
			req := createAndFormatTestRequest(reqData, i, ctx)
			reqStore[i] = req
			RequestQueue <- req
		}
		// Define a DataProcessingFn that appends a string to the field
		cb := DataProcessingFn(func(field []byte) []byte {
			additionalData := []byte(" I have been added in processing!")
			field = append(field, additionalData...)
			return field
		})

		pool := make(chan chan *Request, 1)
		worker := NewWorker(pool)
		worker.Start(1, cb)

		go func() {
			for {
				select {
				case req := <-RequestQueue:
					// a job request has been received
					wg.Add(1)
					go func(req *Request) {
						// try to obtain a worker job channel that is available.
						// this will block until a worker is idle
						reqChannel := <-worker.WorkerPool

						// dispatch the job to the worker job channel
						reqChannel <- req
						cancel()
						wg.Done()
					}(req)
				default:
					return
				}
			}
		}()
		time.Sleep(time.Second * 5)
		wg.Wait()
		worker.Stop()

		for _, req := range reqStore {
			buffer := bytes.Buffer{}
			buffer.WriteString("Hello, Worker!")
			got := string(getBinaryRepresentationFromSerialisable(req.Message).Data)
			want := buffer.String()
			if !strings.EqualFold(got, want) {
				t.Errorf("Expected %s, got %s", want, got)
			}
		}
	})

	t.Run("test worker with dispatcher", func(t *testing.T) {
		requests := []BinaryRepresentation{
			{Version: uint16(1), ClientId: uint16(1), Token: []byte("custom token"), Data: []byte("1. Hello, Worker!")},
			{Version: uint16(1), ClientId: uint16(2), Token: []byte("custom token"), Data: []byte("2. Hello, Worker!")},
			{Version: uint16(1), ClientId: uint16(3), Token: []byte("custom token"), Data: []byte("3. Hello, Worker!")},
		}
		ctx := context.TODO()
		reqStore := make([]*Request, len(requests))
		RequestQueue := make(chan *Request, MAX_QUEUE)
		for i, reqData := range requests {
			reqStore[i] = createAndFormatTestRequest(reqData, i, ctx)
			RequestQueue <- reqStore[i]
		}
		// Define a DataProcessingFn that appends a string to the field
		cb := DataProcessingFn(func(field []byte) []byte {
			additionalData := []byte(" I have been added in processing!")
			field = append(field, additionalData...)
			return field
		})

		dispatcher := NewDispatcher(1, MAX_WORKER)
		dispatcher.AddQueue(RequestQueue)
		dispatcher.Run(cb)
		time.Sleep(time.Second * 2)

		for i, req := range reqStore {
			buffer := bytes.Buffer{}
			expectedString := fmt.Sprintf("%d. Hello, Worker! I have been added in processing!", i+1)
			buffer.WriteString(expectedString)
			got := string(getBinaryRepresentationFromSerialisable(req.Message).Data)
			want := buffer.String()
			if !strings.EqualFold(got, want) {
				t.Errorf("Expected %s, got %s", want, got)
			}
		}

		fmt.Printf("%v", reqStore[1].Message.buf.String())
	})
}
