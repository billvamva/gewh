package core

import (
	"context"
	"log"
)

type Request struct {
	Id int
	Message Serialisable
	Ctx context.Context
}

func NewRequest(id int, message Serialisable, ctx context.Context) *Request {
	return &Request{
		id,
		message,
		ctx,
	}
} 

func (req *Request) Process() {
	decodedByteFields, err := req.Message.Decode()
	if err != nil {
		log.Printf("Could not decode data on request:%v , error: %v",req.Id, err)
		return
	}
	decodedBinaryRepresentation := BinaryRepresentation{}
	decodedBinaryRepresentation.FormatDecodedFields(decodedByteFields)
	// PlaceHolder for actual processing (forwarding information?)
	decodedBinaryRepresentation.Data = []byte("Updated Message")
	req.Message.BinaryRepresentationToByteFields(&decodedBinaryRepresentation)
	req.Message.Encode()	
}