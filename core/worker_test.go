package core

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"testing"
	"time"
)

var wg sync.WaitGroup
func TestRequestAndQueueHandling(t *testing.T) {
	t.Run("test worker to process reqs from queue", func(t *testing.T) {
		requests := []BinaryRepresentation{
			{Version: uint16(1), ClientId: uint16(1),Token: []byte("custom token"), Data:[]byte("Hello, Worker!"),},
			{Version: uint16(1), ClientId: uint16(2),Token: []byte("custom token"), Data:[]byte("Hello, Worker!"),},
			{Version: uint16(1), ClientId: uint16(2),Token: []byte("custom token"), Data:[]byte("Hello, Worker!"),},
		}
		ctx := context.TODO()
		reqStore := make([]*Request, len(requests))
		RequestQueue := make(chan *Request, MAX_QUEUE)
		for i, reqData := range requests {
			req := createAndFormatTestRequest(&reqData, i, ctx)
			reqStore[i] = req
			RequestQueue <- req
		}

		pool := make(chan chan *Request, 1)
		worker := NewWorker(pool)
		worker.Start(1)
		t.Logf("%v", len(RequestQueue))

		go func() {
			for {
				select {
				case req := <-RequestQueue:
					// a job request has been received
					wg.Add(1)
					go func(req *Request) {
						// try to obtain a worker job channel that is available.
						// this will block until a worker is idle
						reqChannel := <- worker.WorkerPool

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

		for _, req := range reqStore {
			buffer :=  bytes.Buffer{}
			buffer.WriteString("Updated Message")
			got := req.Message.buf.String()
			want := buffer.String()
			if strings.Compare(got, want) == 0 {
				t.Errorf("Expected %s, got %s",want,got)
			}
		}
	})

	t.Run("test worker to process req from queue and cancel req before processing", func(t *testing.T) {
		requests := []BinaryRepresentation{
			{Version: uint16(1), ClientId: uint16(1),Token: []byte("custom token"), Data:[]byte("Hello, Worker!"),},
		}
		parentCtx := context.Background()
		ctx, cancel := context.WithCancel(parentCtx)
		reqStore := make([]*Request, len(requests))
		RequestQueue := make(chan *Request, MAX_QUEUE)
		for i, reqData := range requests {
			req := createAndFormatTestRequest(&reqData, i, ctx)
			reqStore[i] = req
			RequestQueue <- req
		}

		pool := make(chan chan *Request, 1)
		worker := NewWorker(pool)
		worker.Start(1)
		t.Logf("%v", len(RequestQueue))

		go func() {
			for {
				select {
				case req := <-RequestQueue:
					// a job request has been received
					wg.Add(1)
					go func(req *Request) {
						// try to obtain a worker job channel that is available.
						// this will block until a worker is idle
						reqChannel := <- worker.WorkerPool

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
		time.Sleep(time.Second * 2)
		wg.Wait()
		worker.Stop()

		for _, req := range reqStore {
			buffer :=  bytes.Buffer{}
			buffer.WriteString("Hello Worker")
			got := req.Message.buf.String()
			want := buffer.String()
			if strings.Compare(got, want) == 0 {
				t.Errorf("Expected %s, got %s",want,got)
			}
		}
	})

}

