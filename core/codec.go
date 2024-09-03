package core

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"
)

// Defines the contract for managing fields in a Codec.
type Codec interface {
	AddFields(fields ByteFields) // Adds Fields for the specific Codec
	GetFields() ByteFields       // Gets all Fields for the specific Codec
}

// Codec for base messages we will be handling
type MessageCodec struct {
	fields ByteFields // bytefields configured for the specific Codec
}

// add fields to Codec
func (c *MessageCodec) AddFields(fields ByteFields) {
	c.fields = fields
}

// get fields from Codec
func (c *MessageCodec) GetFields() ByteFields {
	return c.fields
}

// data has to be first be transformed to a Bytefield to be encoded and written on the Codec
type ByteField struct {
	Name     string       // name of the field
	DataType reflect.Type // data type of the field
	Value    interface{}  // Holds pointer of value
}

type ByteFields []ByteField

// Defines methods for encoding and decoding data for a serializable.
type BaseSerialisable interface {
	Encode()                     // Encode data from the Codec fields to the buffer
	Decode() (ByteFields, error) // Decode data from buffer to byte fields
}

// anything that can be serialised and deserialised
type Serialisable struct {
	buf   *bytes.Buffer // holds binary data
	Codec Codec         // holds information on how to decode and encode the binary data
}

func NewSerialisable() *Serialisable {
	return &Serialisable{
		bytes.NewBuffer([]byte{}),
		&MessageCodec{},
	}
}

// Struct representation of a serialisable's buffer content using its Codec. This format is used as an interface with the serialisable.
type Payload struct {
	Version    uint16 `json:"version"`    // version of encoding
	ClientId   uint16 `json:"clientId"`   // origin client id
	Identifier []byte `json:"identifier"` // Identifier
	Data       []byte `json:"data"`       //  data sent
}

func NewPayload(version uint16, clientId uint16, token []byte, data []byte) *Payload {
	return &Payload{
		version,
		clientId,
		token,
		data,
	}
}

// Encode method implementation for Serialisable
func (s *Serialisable) Encode() error {
	fields := s.Codec.GetFields()
	tempBuf := new(bytes.Buffer)
	for _, field := range fields {
		err := binary.Write(tempBuf, binary.LittleEndian, field.Value)
		if err != nil {
			log.Fatalf("Error encoding message %v", err)
			return err
		}
	}
	s.InsertDataToSerialisableBuffer(tempBuf.Bytes())
	return nil
}

func (s *Serialisable) GetBufString() string {
	return s.buf.String()
}

// Decode method implementation for Serialisable
func (s *Serialisable) Decode() (ByteFields, error) {
	if s.buf == nil {
		log.Fatal("Insert Data into buffer of serialisable to decode.")
	}
	fields := s.Codec.GetFields()

	for i := range fields {
		field := &fields[i]
		switch field.Name {
		case "Identifier", "Data":
			length, err := getLengthForDynamicSizedField(fields, field.Name)
			if err != nil {
				log.Printf("Error in Decoding: %v, field value %v", err, field.Value)
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

// get length for a dynamic sized field
func getLengthForDynamicSizedField(fields ByteFields, fieldName string) (uint32, error) {
	var length uint32
	lengthKey := fieldName + "Length"
	for _, field := range fields {
		if strings.Compare(field.Name, lengthKey) == 0 {
			length = *((field).Value.(*uint32))
		}
	}
	if length == 0 {
		return 0, fmt.Errorf("invalid key for dynamic shaped field, %s", fieldName)
	}
	return length, nil
}

// use length to read data from dynamic sized field
func (s *Serialisable) readDataFromDynamicSizedField(field *ByteField, fieldSize uint32) error {
	message := make([]byte, fieldSize)
	err := binary.Read(s.buf, binary.LittleEndian, &message)
	if err != nil {
		return err
	}
	field.Value = message
	return nil
}

// insert raw binary data into the buffer
func (s *Serialisable) InsertDataToSerialisableBuffer(binaryData []byte) {
	if s.buf == nil {
		s.buf = bytes.NewBuffer(binaryData)
	} else {
		s.buf.Reset()
		s.buf.Write(binaryData)
	}
}

// transform the binary struct represenatation into fields and then add it to the Codec of the serialisable
// ToFields converts a Payload to a slice of Fields
func (p *Payload) ToFields() ByteFields {
	tokenLength := uint32(len(p.Identifier))
	dataLength := uint32(len(p.Data))

	return ByteFields{
		{"Version", reflect.TypeOf(p.Version), &p.Version},
		{"ClientId", reflect.TypeOf(p.ClientId), &p.ClientId},
		{"IdentifierLength", reflect.TypeOf(uint32(0)), &tokenLength},
		{"Identifier", reflect.TypeOf(p.Identifier), &p.Identifier},
		{"DataLength", reflect.TypeOf(uint32(0)), &dataLength},
		{"Data", reflect.TypeOf(p.Data), &p.Data},
	}
}

// transform byte fields into binary representation for processing
func (b *Payload) FromFields(fields ByteFields) {
	for _, field := range fields {
		switch v := field.Value.(type) {
		case *uint32:
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
			case "Identifier":
				b.Identifier = v
			default:
				log.Printf("unsupported field name: %v", field.Name)
			}
		default:
			log.Printf("unsupported field type: %v", reflect.TypeOf(field.Value))
			continue
		}
	}
}

// transform binary representation into json
func (b Payload) MarshalToJson() []byte {
	jsonData, err := json.Marshal(b)
	if err != nil {
		log.Println("Error encoding JSON:", err)
		return nil
	}
	return jsonData
}

// unmarshal json to binary representation
func (b *Payload) UnmarshalJson(data []byte) {
	err := json.Unmarshal(data, b)
	if err != nil {
		log.Println("Error decoding JSON:", err)
		return
	}
}
