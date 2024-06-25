package core

import (
	"context"
	"log"
)

// request is the format that our serialisable messages are going to be sent as to the message broker.
type Request struct {
	Id int // id of the request
	Message Serialisable // message of serialisable form.
	Ctx context.Context // context to keep track of cancelled requests and remove from the message broker.
}

// creates new request
func NewRequest(id int, message Serialisable, ctx context.Context) *Request {
	return &Request{
		id,
		message,
		ctx,
	}
} 

type DataProcessingCallBack interface {
	modifyDataField([]byte) []byte
}

type DataProcessingFn func([]byte) []byte 

func (d DataProcessingFn) modifyDataField(field []byte) []byte {
	return d(field)
}

// placeholder for actual request processing.
func (req *Request) Process(cb DataProcessingFn) {
	decodedByteFields, err := req.Message.Decode()
	if err != nil {
		log.Printf("Could not decode data on request:%v , error: %v",req.Id, err)
		return
	}
	decodedBinaryRepresentation := BinaryRepresentation{}
	decodedBinaryRepresentation.FormatDecodedFields(decodedByteFields)
	// PlaceHolder for actual processing (forwarding information?)
	decodedBinaryRepresentation.Data = cb(decodedBinaryRepresentation.Data)
	req.Message.BinaryRepresentationToByteFields(&decodedBinaryRepresentation)
	req.Message.Encode()	
}