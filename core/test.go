package core

import (
	"context"
	"encoding/base64"
	"log"
)

func createAndFormatTestRequest(payload *Payload, id int, ctx context.Context) *Request {
	serialisable := NewSerialisable()
	req := NewRequest(id, serialisable, ctx)
	req.AddPayload(payload)
	return req
}

func encodeStringToBase64(data []byte) string {
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	base64.StdEncoding.Encode(dst, data)
	return string(dst)
}

func getPayloadFromSerialisable(req *Serialisable) Payload {
	gotValues := Payload{
		*new(uint16),
		*new(uint16),
		[]byte{},
		[]byte{},
	}

	decodedFields, err := req.Decode()
	if err != nil {
		log.Fatalf("decoding failed: %v, for req with contents:%v \n", err, req.buf.String())
	}

	gotValues.FromFields(decodedFields)

	return gotValues
}
