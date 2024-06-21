package core

import (
	"bytes"
	"encoding/binary"
	"log"
	"reflect"
)

type Codec interface {
	AddFields(ByteFields)
	GetFields() ByteFields
}

type BaseSerialisable interface {
	Encode(version uint16, clientId uint16, message string)
	Decode(binaryData []byte)
}

type BinaryRepresentation struct {
	version uint16
	clientId uint16
	message []byte
}

type MessageCodec struct {
	fields ByteFields
}

type ByteField struct {
	Name string
	DataType reflect.Type
	Value interface {} //Holds pointer of value
}

type ByteFields []ByteField

type Serialisable struct {
	buf *bytes.Buffer
	codec Codec
}

// Encode method implementation for Serialisable
func (s *Serialisable) Encode()  {
	fields := s.codec.GetFields()
	tempBuf := new(bytes.Buffer)
	for _, field := range fields {
		err := binary.Write(tempBuf, binary.LittleEndian, field.Value)
		if err != nil {
			log.Fatalf("Error encoding message %v", err)
			return 
		}
	}
	s.InsertDataToSerialisableBuffer(tempBuf.Bytes())
}

// Decode method implementation for Serialisable
func (s *Serialisable) Decode() (ByteFields, error) {
	if s.buf == nil {
		log.Fatal("Insert Data into buffer of serialisable to decode.")
	}	
	fields := s.codec.GetFields()

	for i := range fields {
		if fields[i].Name != "Message" {
			// Read values into pointers
			err := binary.Read(s.buf, binary.LittleEndian, fields[i].Value)
			if err != nil {
				return nil, err
			}
		}
	}

	// Read the message based on MessageLength
	messageLength := *(fields[2].Value.(*uint8))
	message := make([]byte, messageLength)
	err := binary.Read(s.buf, binary.LittleEndian, &message)
	if err != nil {
		return nil, err
	}
	fields[3].Value = message

	return fields, nil
}

func (s *Serialisable) InsertDataToSerialisableBuffer(binaryData []byte) {
	if s.buf == nil {
		s.buf = bytes.NewBuffer(binaryData)
	} else {
		s.buf.Reset()
		s.buf.Write(binaryData)
	}	
}

func FormatDecodedFields(decodedFields ByteFields) BinaryRepresentation {
	binaryRep := BinaryRepresentation{}
	for _, field := range decodedFields {
		switch v := field.Value.(type) {
		case *uint16:
			if field.Name == "Version" {
				binaryRep.version = *v
			} else if field.Name == "ClientId" {
				binaryRep.clientId = *v
			}
		case []byte:
			binaryRep.message = v 
		default:
			log.Printf("unsupported field type: %v", reflect.TypeOf(field.Value))
			continue
		}
	}
	return binaryRep
}

func (s *Serialisable) BinaryRepresentationToByteFields(binaryRep *BinaryRepresentation) {
	messageLength := uint8(len(binaryRep.message))
	s.codec.AddFields(ByteFields{
		{"Version", reflect.TypeOf(binaryRep.version), &binaryRep.version},
		{"ClientId", reflect.TypeOf(binaryRep.clientId), &binaryRep.clientId},
		{"MessageLength", reflect.TypeOf(uint8(len(binaryRep.message))), &messageLength},
		{"Message", reflect.TypeOf([]byte(binaryRep.message)), []byte(binaryRep.message)},
	})
}

func (c *MessageCodec) AddFields(fields ByteFields) {
	c.fields = fields
}

func (c *MessageCodec) GetFields() ByteFields {
	return c.fields
}