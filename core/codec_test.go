package core

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
)


func TestEncodingAndDecoding(t *testing.T) {
	t.Run("test binary encoding", func(t *testing.T) {
		version := uint16(1)
		clientId := uint16(42)
		data := "Hello, World!"
		token := "custom token"


		binaryRep := BinaryRepresentation{
			version,
			clientId,
			[]byte(token),
			[]byte(data),
		}
		testSerialisable := createTestSerialisable()
		testSerialisable.BinaryRepresentationToByteFields(&binaryRep) 

		testSerialisable.Encode()
		want := []byte{
			0x01, 0x00,                   // Version
			0x2a, 0x00,                   // ClientId (little-endian)
			0x0C,                         // TokenLength (12 bytes)
			0x63, 0x75, 0x73, 0x74, 0x6f, // Token "custom token"
			0x6d, 0x20, 0x74, 0x6f,
			0x6b, 0x65, 0x6e,
			0x0d,                         // MessageLength
			0x48, 0x65, 0x6c, 0x6c, 0x6f, // Message "Hello, World!"
			0x2c, 0x20, 0x57, 0x6f,
			0x72, 0x6c, 0x64, 0x21,
		}
	

		got := testSerialisable.buf.Bytes()

		// Print the binary data
		if !bytes.Equal(got, want) {
			t.Fatalf("encoded output mismatch: got %#x, want %#x", got, want)
		}
	})
	t.Run("test binary decoding", func(t *testing.T) {
		// Given encoded binary data
		encodedData := []byte{
			0x01, 0x00,                   // Version
			0x2a, 0x00,                   // ClientId (little-endian)
			0x0C,                         // TokenLength (12 bytes)
			0x63, 0x75, 0x73, 0x74, 0x6f, // Token "custom token"
			0x6d, 0x20, 0x74, 0x6f,
			0x6b, 0x65, 0x6e,
			0x0d,                         // MessageLength
			0x48, 0x65, 0x6c, 0x6c, 0x6f, // Message "Hello, World!"
			0x2c, 0x20, 0x57, 0x6f,
			0x72, 0x6c, 0x64, 0x21,
		}

		version := uint16(1)
		clientId := uint16(42)
		message := "Hello, World!"
		token := "custom token"

		gotValues := BinaryRepresentation{
			*new(uint16),
			*new(uint16),
			[]byte{},
			[]byte{},
		}
		testSerialisable := createTestSerialisable()
		testSerialisable.BinaryRepresentationToByteFields(&gotValues) 

		testSerialisable.InsertDataToSerialisableBuffer(encodedData)

		decodedFields, err := testSerialisable.Decode()
		if err != nil {
			t.Fatalf("decoding failed: %v", err)
		}

		// Expected values
		expectedValues := BinaryRepresentation{
			Version:  version,
			ClientId: clientId,
			Token: []byte(token),
			Data:  []byte(message),
		}

		gotValues.FormatDecodedFields(decodedFields)

		if !reflect.DeepEqual(gotValues, expectedValues) {
			t.Errorf("decoded field mismatch: got %v, want %v", gotValues, expectedValues)
		}
	})

	t.Run("test json marshaling", func(t *testing.T) {
		// Given encoded binary data
		encodedData := []byte{
			0x01, 0x00,                   // Version
			0x2a, 0x00,                   // ClientId (little-endian)
			0x0C,                         // TokenLength (12 bytes)
			0x63, 0x75, 0x73, 0x74, 0x6f, // Token "custom token"
			0x6d, 0x20, 0x74, 0x6f,
			0x6b, 0x65, 0x6e,
			0x0d,                         // MessageLength
			0x48, 0x65, 0x6c, 0x6c, 0x6f, // Message "Hello, World!"
			0x2c, 0x20, 0x57, 0x6f,
			0x72, 0x6c, 0x64, 0x21,
		}

		gotValues := BinaryRepresentation{
			*new(uint16),
			*new(uint16),
			[]byte{},
			[]byte{},
		}
		testSerialisable := createTestSerialisable()
		testSerialisable.BinaryRepresentationToByteFields(&gotValues) 

		testSerialisable.InsertDataToSerialisableBuffer(encodedData)

		decodedFields, err := testSerialisable.Decode()
		if err != nil {
			t.Fatalf("decoding failed: %v", err)
		}

		token := encodeStringToBase64([]byte("custom token"))
		data := encodeStringToBase64([]byte("Hello, World!"))

		jsonString := fmt.Sprintf(`{"version":1,"clientId":42,"token":%s,"data":%s}`, token, data)
		// Expected values
		expectedJson := []byte(
			jsonString,
		)

		gotValues.FormatDecodedFields(decodedFields)
		gotJson := gotValues.MarshalToJson()

		if  bytes.Equal(gotJson, expectedJson){
			t.Errorf("json mismatch: got %v, want %v", string(gotJson), string(expectedJson))
		}
	})
	t.Run("test json unmarshaling", func(t *testing.T) {
		// Given encoded binary data
		token := encodeStringToBase64([]byte("custom token"))
		data := encodeStringToBase64([]byte("Hello, World!"))

		inputJsonString := fmt.Sprintf(`{"version":1,"clientId":42,"token":"%s","data":"%s"}`, token, data)

		gotValues := BinaryRepresentation{
			*new(uint16),
			*new(uint16),
			[]byte{},
			[]byte{},
		}

		version := uint16(1)
		clientId := uint16(42)
		message := "Hello, World!"
		expectedToken := "custom token"

		expectedValues := BinaryRepresentation{
			Version:  version,
			ClientId: clientId,
			Token: []byte(expectedToken),
			Data:  []byte(message),
		} 

		gotValues.UnmarshalJson([]byte(inputJsonString))
		if  !reflect.DeepEqual(gotValues, expectedValues){
			t.Errorf("binary rep struct mismatch: got %v, want %v", gotValues, expectedValues)
		}
	})
}

