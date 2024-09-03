package core

import (
	"context"
	"log"
)

// request is the format that our serialisable messages are going to be sent as to the message broker.
type Request struct {
	Id      int             // id of the request
	Message *Serialisable   // message of serialisable form.
	Ctx     context.Context // context to keep track of cancelled requests and remove from the message broker.
}

// creates new request
func NewRequest(id int, message *Serialisable, ctx context.Context) *Request {
	return &Request{
		id,
		message,
		ctx,
	}
}

func (r *Request) AddPayload(payload *Payload) {
	fields := payload.ToFields()
	r.Message.Codec.AddFields(fields)
	r.Message.Encode()
}

type DataProcessor interface {
	Process(*Request) error
}

type MockDataProcessingFn func([]byte) []byte

func (fn MockDataProcessingFn) modifyDataField(field []byte) []byte {
	return fn(field)
}

// request processing, currently needs even an empty function should be changed.
func (fn MockDataProcessingFn) Process(req *Request) error {
	// decoding current message into byte fields given the codec that is attached to the message
	decodedByteFields, err := req.Message.Decode()
	if err != nil {
		log.Printf("Could not decode data on request:%v , error: %v", req.Id, err)
		return err
	}
	// create empty payload and populate it with the decoded bytefields
	decodedPayload := Payload{}
	decodedPayload.FromFields(decodedByteFields)
	// Processing of message
	decodedPayload.Data = fn.modifyDataField(decodedPayload.Data)
	// back to byte fields
	modifiedByteFields := decodedPayload.ToFields()
	req.Message.Codec.AddFields(modifiedByteFields)
	// encode
	req.Message.Encode()
	return nil
}
