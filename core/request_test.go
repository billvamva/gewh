package core

import (
	"context"
	"reflect"
	"testing"
)


func TestRequestAndQueueHandling(t *testing.T) {
	t.Run("test request enqueue", func(t *testing.T) {

		version := uint16(1)
		clientId := uint16(42)
		message := "Hello, World!"

		fields := ByteFields{
			{"Version", reflect.TypeOf(version), &version},
			{"ClientId", reflect.TypeOf(clientId), &clientId},
			{"MessageLength", reflect.TypeOf(uint8(len(message))), uint8(len(message))},
			{"Message", reflect.TypeOf([]byte(message)), []byte(message)},
		}

		testSerialisable := createTestSerialisable(fields)
		ctx := context.Background()
		req := GetRequestFromPool()
		*req = *NewRequest(0, testSerialisable, ctx)
		messageQueue := NewMessageQueue() 
		messageQueue.Enqueue(req)

		got := len(messageQueue.requests) 

		if  got != 1 {
			t.Errorf("Wanted 1 request in queue, got: %v", got)
		}
		requestPool.Put(req)
	})
	t.Run("test request dequeue", func(t *testing.T) {

		version := uint16(1)
		clientId := uint16(42)
		message := "Hello, World!"

		binaryRep := BinaryRepresentation{
			version,
			clientId,
			[]byte(message),	
		}

		messageCodec := MessageCodec{}
		testSerialisable := Serialisable{
			&dummyBuffer,
			&messageCodec,
		}

		ctx := context.Background()
		testSerialisable.BinaryRepresentationToByteFields(&binaryRep)
		testSerialisable.Encode()
		req := GetRequestFromPool()
		*req = *NewRequest(0, testSerialisable, ctx)
		messageQueue := NewMessageQueue() 
		messageQueue.Enqueue(req)

		dequeuedRequest := messageQueue.Dequeue()	

		if dequeuedRequest == nil {
			t.Fatal("No request in queue.")
		}

		decodedFields, err := testSerialisable.Decode()
		if err != nil {
			t.Fatalf("Error decoding message, %v", err)
		}
		gotValues := FormatDecodedFields(decodedFields)
		got := gotValues.message
		want := []byte("Hello, World!")

		if  !reflect.DeepEqual(got, want)  {
			t.Errorf("Wanted 1 request in queue, got: %v", got)
		}
		requestPool.Put(dequeuedRequest)
	})

	// t.Run("test worker to dequeue messages and send to req channel", func(t *testing.T) {
	// 	requests := []BinaryRepresentation{
	// 		{version: uint16(1), clientId: uint16(1),message:[]byte("Hello, Worker!"),},
	// 		{version: uint16(1), clientId: uint16(2),message:[]byte("Hello, Worker!"),},
	// 		{version: uint16(1), clientId: uint16(2),message:[]byte("Hello, Worker!"),},
	// 	}
	// 	for i:=1; i<=2; i++ {
	// 		go Worker(i, MessageQueue)
	// 	}
	// })

}


