package core

import (
	"bytes"
	"context"
	"encoding/base64"
)


var dummyBuffer bytes.Buffer = *bytes.NewBuffer([]byte{})

func createTestSerialisable() Serialisable {
	messageCodec := MessageCodec{}

	testSerialisable := Serialisable{
		&dummyBuffer,
		&messageCodec,
	}

	return testSerialisable
}

func createAndFormatTestRequest(reqData *BinaryRepresentation, id int, ctx context.Context) *Request{
	testSerialisable := createTestSerialisable()
	testSerialisable.BinaryRepresentationToByteFields(reqData)
	testSerialisable.Encode()
	req := NewRequest(id, testSerialisable, ctx)
	return req
}

func encodeStringToBase64(data []byte) string {
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	base64.StdEncoding.Encode(dst, data)
	return string(dst)
}