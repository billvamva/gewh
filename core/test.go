package core

import (
	"bytes"
	"context"
	"encoding/base64"
	"log"
)


func createTestSerialisable() Serialisable {
	messageCodec := MessageCodec{}
	dummyBuffer :=  *bytes.NewBuffer([]byte{}) // Create a new buffer for each Serialisable instance


	testSerialisable := Serialisable{
		&dummyBuffer,
		&messageCodec,
	}

	return testSerialisable
}

func createAndFormatTestRequest(reqData BinaryRepresentation, id int, ctx context.Context) *Request {
	testSerialisable := createTestSerialisable()
	testSerialisable.BinaryRepresentationToByteFields(&reqData)
	testSerialisable.Encode()
	req := NewRequest(id, testSerialisable, ctx)
	empty := BinaryRepresentation{}
	empty.FormatDecodedFields(req.Message.codec.GetFields())
	return req
    }
    
func encodeStringToBase64(data []byte) string {
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	base64.StdEncoding.Encode(dst, data)
	return string(dst)
}

func getBinaryRepresentationFromSerialisable(req Serialisable) BinaryRepresentation {
	gotValues := BinaryRepresentation{
		*new(uint16),
		*new(uint16),
		[]byte{},
		[]byte{},
	}

	decodedFields, err := req.Decode()
	if err != nil {
		log.Fatalf("decoding failed: %v, for req with contents:%v \n", err, req.buf.String())
	}

	gotValues.FormatDecodedFields(decodedFields)

	return gotValues
}