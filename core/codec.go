package core

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"log"
	"reflect"
	"strings"
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
	Version uint16 `json:"version"`
	ClientId uint16 `json:"clientId"`
	Token []byte `json:"token"`
	Data []byte `json:"data"`
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
		field := &fields[i]
		switch field.Name {
		case "Token", "Data":
			length, err := getLengthForDynamicSizedField(fields, field.Name)
			if err != nil {
				log.Fatalf("Error in Decoding: %v", err)
			}
			s.readDataFromDynamicSizedField(field, length)
		default:
			// Non-byte arrays: Read values into pointers
			err := binary.Read(s.buf, binary.LittleEndian, field.Value)
			if err != nil {
				return nil, err
			}
		}
	}


	return fields, nil
}

func getLengthForDynamicSizedField(fields ByteFields, fieldName string) (uint8, error) {
	var length uint8
	lengthKey := fieldName + "Length"
	for _, field := range fields {
		if strings.Compare(field.Name, lengthKey) == 0 {
			length =  *((field).Value.(*uint8))
		}
	}
	if length == 0 {
		return 0,errors.New("invalid key for dynamic shaped field")
	}
	return length, nil
}

func (s *Serialisable) readDataFromDynamicSizedField(field *ByteField, fieldSize uint8) error {
	message := make([]byte, fieldSize)
	err := binary.Read(s.buf, binary.LittleEndian, &message)
	if err != nil {
		return err
	}
	field.Value = message
	return nil
}

func (s *Serialisable) InsertDataToSerialisableBuffer(binaryData []byte) {
	if s.buf == nil {
		s.buf = bytes.NewBuffer(binaryData)
	} else {
		s.buf.Reset()
		s.buf.Write(binaryData)
	}	
}

func (s *Serialisable) BinaryRepresentationToByteFields(binaryRep *BinaryRepresentation) {
	dataLength := uint8(len(binaryRep.Data))
	tokenLength := uint8(len(binaryRep.Token))
	s.codec.AddFields(ByteFields{
		{"Version", reflect.TypeOf(binaryRep.Version), &binaryRep.Version},
		{"ClientId", reflect.TypeOf(binaryRep.ClientId), &binaryRep.ClientId},
		{"TokenLength", reflect.TypeOf(uint8(len(binaryRep.Token))), &tokenLength},
		{"Token", reflect.TypeOf([]byte(binaryRep.Token)), []byte(binaryRep.Token)},
		{"DataLength", reflect.TypeOf(uint8(len(binaryRep.Data))), &dataLength},
		{"Data", reflect.TypeOf([]byte(binaryRep.Data)), []byte(binaryRep.Data)},
	})
}

func (b *BinaryRepresentation) FormatDecodedFields(decodedFields ByteFields) {
	for _, field := range decodedFields {
		switch v := field.Value.(type) {
		case *uint8:
			continue
		case *uint16:
			switch field.Name {
			case "Version":
				b.Version = *v
			case "ClientId":
				b.ClientId = *v
			}
		case []byte:
			switch field.Name {
			case "Data":
				b.Data = v 
			case "Token":
				b.Token = v
			default:
				log.Printf("unsupported field name: %v", field.Name)
			}
		default:
			log.Printf("unsupported field type: %v", reflect.TypeOf(field.Value))
			continue
		}
	}
}

func (b BinaryRepresentation) MarshalToJson() []byte{
	jsonData, err := json.Marshal(b)

	if err != nil {
		log.Println("Error encoding JSON:", err)
		return nil
	}
	return jsonData
}

func (b *BinaryRepresentation) UnmarshalJson(data []byte){
	err := json.Unmarshal(data, b)
	if err != nil {
		log.Println("Error decoding JSON:", err)
		return
	}
}

func (c *MessageCodec) AddFields(fields ByteFields) {
	c.fields = fields
}

func (c *MessageCodec) GetFields() ByteFields {
	return c.fields
}