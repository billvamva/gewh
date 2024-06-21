package core

import (
	"bytes"
	"encoding/binary"
	"log"
	"reflect"
)

type Codec interface {
	AddField(ByteField)
	GetFields() ByteFields
}

type BaseSerialisable interface {
	Encode(version uint8, clientId uint16, message string)
	Decode(binaryData []byte)
}

type BinaryRepresentation struct {
	version uint8
	clientId uint16
	message string
}

type MessageCodec struct {
	fields ByteFields
}

type ByteField struct {
	Name string
	DataType reflect.Type
	Value interface {}
}

type ByteFields []ByteField

type Serialisable struct {
	buf *bytes.Buffer
	codec Codec
}

// Encode method implementation for Serialisable
func (s *Serialisable) Encode(version uint8, clientId uint16, message string)  {
	fields := s.codec.GetFields()
	if s.buf == nil {
		s.buf = new(bytes.Buffer)
	}
	for _, field := range fields {
		err := binary.Write(s.buf, binary.LittleEndian, field.Value)
		if err != nil {
			log.Fatalf("Error encoding message %v", err)
			return 
		}
	}
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
		case *uint8:
			// we don't need to encode message length
			if field.Name == "Version" {
				binaryRep.version = *v
			}
		case *uint16:
			binaryRep.clientId = *v
		case []byte:
			binaryRep.message = string(v)
		default:
			log.Printf("unsupported field type: %v", reflect.TypeOf(field.Value))
			continue
		}
	}
	return binaryRep
}


func (c *MessageCodec) AddField(field ByteField) {
	c.fields = append(c.fields, field)
}

func (c *MessageCodec) GetFields() ByteFields {
	return c.fields
}