package core

import (
	"bytes"
	"reflect"
	"testing"
)

var dummyBuffer bytes.Buffer = *bytes.NewBuffer([]byte{})

func TestEncodingAndDecoding(t *testing.T) {
	t.Run("test binary encoding", func(t *testing.T) {
		version := uint16(1)
		clientId := uint16(42)
		message := "Hello, World!"

		fields := ByteFields{
			{"Version", reflect.TypeOf(version), version},
			{"ClientId", reflect.TypeOf(clientId), clientId},
			{"MessageLength", reflect.TypeOf(uint8(len(message))), uint8(len(message))},
			{"Message", reflect.TypeOf([]byte(message)), []byte(message)},
		}
		testSerialisable := createTestSerialisable(fields)

		testSerialisable.Encode()
		want := []byte{
			0x01, 0x00,                   // Version
			0x2a, 0x00,                   // ClientId (little-endian)
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
			0x0d,                         // MessageLength
			0x48, 0x65, 0x6c, 0x6c, 0x6f, // Message "Hello, World!"
			0x2c, 0x20, 0x57, 0x6f,
			0x72, 0x6c, 0x64, 0x21,
		}

		version := uint16(1)
		clientId := uint16(42)
		message := "Hello, World!"

		expectedFields := ByteFields{
			{"Version", reflect.TypeOf(version), new(uint16)},
			{"ClientId", reflect.TypeOf(clientId), new(uint16)},
			{"MessageLength", reflect.TypeOf(uint8(len(message))), new(uint8)},
			{"Message", reflect.TypeOf([]byte(message)), []byte{}},
		}

		testSerialisable := createTestSerialisable(expectedFields)

		testSerialisable.InsertDataToSerialisableBuffer(encodedData)

		decodedFields, err := testSerialisable.Decode()
		if err != nil {
			t.Fatalf("decoding failed: %v", err)
		}

		// Expected values
		expectedValues := BinaryRepresentation{
			version:  version,
			clientId: clientId,
			message:  []byte(message),
		}

		gotValues := FormatDecodedFields(decodedFields)

		if !reflect.DeepEqual(gotValues, expectedValues) {
			t.Errorf("decoded field mismatch: got %v, want %v", gotValues, expectedValues)
		}
	})

	t.Run("test binary decoding with incorrect type", func(t *testing.T) {
		// Given encoded binary data with incorrect types
		encodedData := []byte{
			0x01, 0x00,                   // Version (should be uint8)
			0x2a, 0x00,                   // ClientId (should be uint16)
			0x0d,                         // MessageLength (should be uint8)
			0x48, 0x65, 0x6c, 0x6c, 0x6f, // Message "Hello, World!" (should be []byte)
			0x2c, 0x20, 0x57, 0x6f,
			0x72, 0x6c, 0x64, 0x21,
		}

		// Define fields with incorrect types
		version := uint16(1)    // Incorrect type: should be uint8
		clientId := uint32(42)  // Incorrect type: should be uint16
		message := []uint8{72, 101, 108, 108, 111, 44, 32, 87, 111, 114, 108, 100, 33} // Correct type: just for testing

		incorrectFields := ByteFields{
			{"Version", reflect.TypeOf(version), new(uint32)},
			{"ClientId", reflect.TypeOf(clientId), new(uint32)},
			{"MessageLength", reflect.TypeOf(uint8(len(message))), new(uint8)},
			{"Message", reflect.TypeOf([]byte(message)), make([]byte, len(message))},
		}

		testSerialisable := createTestSerialisable(incorrectFields)

		testSerialisable.InsertDataToSerialisableBuffer(encodedData)

		_, err := testSerialisable.Decode()
		if err == nil {
			t.Fatalf("expected decoding to fail due to incorrect types, but it succeeded")
		} else {
			t.Logf("decoding failed as expected: %v", err)
		}
	})
}

func createTestSerialisable(fields ByteFields) Serialisable {
	messageCodec := MessageCodec{fields}

	testSerialisable := Serialisable{
		&dummyBuffer,
		&messageCodec,
	}

	return testSerialisable
}
