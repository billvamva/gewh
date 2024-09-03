package core

import (
	"bytes"
	"crypto/rand"
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

		payload := Payload{
			version,
			clientId,
			[]byte(token),
			[]byte(data),
		}
		testSerialisable := NewSerialisable()
		fields := payload.ToFields()

		testSerialisable.Codec.AddFields(fields)

		testSerialisable.Encode()
		// Expected encoded data
		want := []byte{
			0x01, 0x00, // Version (uint16, little-endian)
			0x2a, 0x00, // ClientId (uint16, little-endian)
			0x0C, 0x00, 0x00, 0x00, // IdentifierLength (uint32, little-endian, 12 bytes)
			0x63, 0x75, 0x73, 0x74, 0x6f, // Identifier "custom token"
			0x6d, 0x20, 0x74, 0x6f,
			0x6b, 0x65, 0x6e,
			0x0d, 0x00, 0x00, 0x00, // DataLength (uint32, little-endian, 13 bytes)
			0x48, 0x65, 0x6c, 0x6c, 0x6f, // Data "Hello, World!"
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
		encodedData := []byte{
			0x01, 0x00, // Version (uint16, little-endian)
			0x2a, 0x00, // ClientId (uint16, little-endian)
			0x0C, 0x00, 0x00, 0x00, // IdentifierLength (uint32, little-endian, 12 bytes)
			0x63, 0x75, 0x73, 0x74, 0x6f, // Identifier "custom token"
			0x6d, 0x20, 0x74, 0x6f,
			0x6b, 0x65, 0x6e,
			0x0d, 0x00, 0x00, 0x00, // DataLength (uint32, little-endian, 13 bytes)
			0x48, 0x65, 0x6c, 0x6c, 0x6f, // Data "Hello, World!"
			0x2c, 0x20, 0x57, 0x6f,
			0x72, 0x6c, 0x64, 0x21,
		}

		// empty payload that specifies structure
		payload := Payload{
			*new(uint16),
			*new(uint16),
			[]byte{},
			[]byte{},
		}
		testSerialisable := NewSerialisable()
		// get fields from empty payload
		fields := payload.ToFields()
		// use payload to decode message
		testSerialisable.Codec.AddFields(fields)

		testSerialisable.InsertDataToSerialisableBuffer(encodedData)

		decodedFields, err := testSerialisable.Decode()
		if err != nil {
			t.Fatalf("decoding failed: %v", err)
		}

		version := uint16(1)
		clientId := uint16(42)
		message := "Hello, World!"
		identifier := "custom token"
		// Expected values
		expectedValues := Payload{
			Version:    version,
			ClientId:   clientId,
			Identifier: []byte(identifier),
			Data:       []byte(message),
		}

		payload.FromFields(decodedFields)

		if !reflect.DeepEqual(payload, expectedValues) {
			t.Errorf("decoded field mismatch: got %v, want %v", payload, expectedValues)
		}
	})

	t.Run("test json marshaling", func(t *testing.T) {
		// Given encoded binary data
		encodedData := []byte{
			0x01, 0x00, // Version (uint16, little-endian)
			0x2a, 0x00, // ClientId (uint16, little-endian)
			0x0C, 0x00, 0x00, 0x00, // IdentifierLength (uint32, little-endian, 12 bytes)
			0x63, 0x75, 0x73, 0x74, 0x6f, // Identifier "custom token"
			0x6d, 0x20, 0x74, 0x6f,
			0x6b, 0x65, 0x6e,
			0x0d, 0x00, 0x00, 0x00, // DataLength (uint32, little-endian, 13 bytes)
			0x48, 0x65, 0x6c, 0x6c, 0x6f, // Data "Hello, World!"
			0x2c, 0x20, 0x57, 0x6f,
			0x72, 0x6c, 0x64, 0x21,
		}

		payload := Payload{
			*new(uint16),
			*new(uint16),
			[]byte{},
			[]byte{},
		}
		testSerialisable := NewSerialisable()
		fields := payload.ToFields()
		testSerialisable.Codec.AddFields(fields)
		testSerialisable.InsertDataToSerialisableBuffer(encodedData)

		decodedFields, err := testSerialisable.Decode()
		if err != nil {
			t.Fatalf("decoding failed: %v", err)
		}

		token := encodeStringToBase64([]byte("custom token"))
		data := encodeStringToBase64([]byte("Hello, World!"))

		jsonString := fmt.Sprintf(`{"version":1,"clientId":42,"identifier":%s,"data":%s}`, token, data)
		// Expected values
		expectedJson := []byte(
			jsonString,
		)

		payload.FromFields(decodedFields)
		gotJson := payload.MarshalToJson()

		if bytes.Equal(gotJson, expectedJson) {
			t.Errorf("json mismatch: got %v, want %v", string(gotJson), string(expectedJson))
		}
	})
	t.Run("test json unmarshaling", func(t *testing.T) {
		// Given encoded binary data
		identifier := encodeStringToBase64([]byte("origin"))
		data := encodeStringToBase64([]byte("Hello, World!"))

		inputJsonString := fmt.Sprintf(`{"version":1,"clientId":42,"identifier":"%s","data":"%s"}`, identifier, data)

		payload := Payload{
			*new(uint16),
			*new(uint16),
			[]byte{},
			[]byte{},
		}

		version := uint16(1)
		clientId := uint16(42)
		message := "Hello, World!"
		identifier = "origin"

		expectedValues := Payload{
			Version:    version,
			ClientId:   clientId,
			Identifier: []byte(identifier),
			Data:       []byte(message),
		}

		payload.UnmarshalJson([]byte(inputJsonString))
		if !reflect.DeepEqual(payload, expectedValues) {
			t.Errorf("binary rep struct mismatch: got %v, want %v", payload, expectedValues)
		}
	})
}

func TestLongMessageEncodingDecoding(t *testing.T) {
	// Create a large payload
	version := uint16(1)
	clientID := uint16(42)
	identifier := []byte("test-identifier")

	// Generate 5 MB of random data
	dataSize := 5 * 1024 * 1024 // 5 MB
	data := make([]byte, dataSize)
	_, err := rand.Read(data)
	if err != nil {
		t.Fatalf("Failed to generate random data: %v", err)
	}

	// Create the payload
	originalPayload := NewPayload(version, clientID, identifier, data)

	// Create a new Serialisable
	s := NewSerialisable()

	// Set the codec fields
	s.Codec.AddFields(originalPayload.ToFields())

	// Encode the payload
	err = s.Encode()
	if err != nil {
		t.Fatalf("Encoding failed: %v", err)
	}

	// Create a new Serialisable for decoding
	decodedS := NewSerialisable()
	decodedS.Codec.AddFields(originalPayload.ToFields())

	// Insert the encoded data into the buffer
	decodedS.InsertDataToSerialisableBuffer(s.buf.Bytes())

	// Decode the data
	decodedFields, err := decodedS.Decode()
	if err != nil {
		t.Fatalf("Decoding failed: %v", err)
	}

	// Create a new Payload from the decoded fields
	decodedPayload := &Payload{}
	decodedPayload.FromFields(decodedFields)

	// Compare the original and decoded payloads
	if decodedPayload.Version != originalPayload.Version {
		t.Errorf("Version mismatch: got %d, want %d", decodedPayload.Version, originalPayload.Version)
	}
	if decodedPayload.ClientId != originalPayload.ClientId {
		t.Errorf("ClientId mismatch: got %d, want %d", decodedPayload.ClientId, originalPayload.ClientId)
	}
	if !bytes.Equal(decodedPayload.Identifier, originalPayload.Identifier) {
		t.Errorf("Identifier mismatch: got %s, want %s", decodedPayload.Identifier, originalPayload.Identifier)
	}
	if !bytes.Equal(decodedPayload.Data, originalPayload.Data) {
		t.Errorf("Data mismatch: lengths got %d, want %d", len(decodedPayload.Data), len(originalPayload.Data))
	}

	// Check if the data field was correctly preserved
	if !reflect.DeepEqual(decodedPayload.Data, originalPayload.Data) {
		t.Errorf("Data content mismatch")
	}

	t.Logf("Successfully encoded and decoded a payload with %d bytes of data", len(originalPayload.Data))
}
