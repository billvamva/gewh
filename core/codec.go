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

// Defines the contract for managing fields in a codec.
type Codec interface {
	AddFields(ByteFields) // Adds Fields for the specific codec
	GetFields() ByteFields // Gets all Fields for the specific codec
}

// Codec for base messages we will be handling
type MessageCodec struct {
	fields ByteFields // bytefields configured for the specific codec
}

// add fields to codec
func (c *MessageCodec) AddFields(fields ByteFields) {
	c.fields = fields
}

// get fields from codec
func (c *MessageCodec) GetFields() ByteFields {
	return c.fields
}

// data has to be first be transformed to a Bytefield to be encoded and written on the codec
type ByteField struct {
	Name string // name of the field
	DataType reflect.Type // data type of the field
	Value interface {} //Holds pointer of value
}

type ByteFields []ByteField

// Defines methods for encoding and decoding data for a serializable.
type BaseSerialisable interface {
	Encode() // Encode data from the codec fields to the buffer 
	Decode() (ByteFields, error)  // Decode data from buffer to byte fields
}

// anything that can be serialised and deserialised
type Serialisable struct {
	buf *bytes.Buffer // holds binary data
	codec Codec // holds information on how to decode and encode the binary data
}
// Struct representation of a serialisable's buffer content using its codec. This format is used as an interface with the serialisable.
type BinaryRepresentation struct {
	Version uint16 `json:"version"` // version of encoding
	ClientId uint16 `json:"clientId"` // origin client id
	Token []byte `json:"token"` // authentication token
	Data []byte `json:"data"` //  data sent
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

// get length for a dynamic sized field
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

// use length to read data from dynamic sized field
func (s *Serialisable) readDataFromDynamicSizedField(field *ByteField, fieldSize uint8) error {
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

// transform the binary struct represenatation into fields and then add it to the codec of the serialisable
func (s *Serialisable) BinaryRepresentationToByteFields(binaryRep *BinaryRepresentation) {
	dataLength := uint8(len(binaryRep.Data))
	tokenLength := uint8(len(binaryRep.Token))
	s.codec.AddFields(ByteFields{
		{"Version", reflect.TypeOf(binaryRep.Version), &binaryRep.Version},
		{"ClientId", reflect.TypeOf(binaryRep.ClientId), &binaryRep.ClientId},
		{"TokenLength", reflect.TypeOf(uint8(len(binaryRep.Token))), &tokenLength},
		{"Token", reflect.TypeOf(binaryRep.Token), append([]byte(nil), binaryRep.Token...)},
		{"DataLength", reflect.TypeOf(uint8(len(binaryRep.Data))), &dataLength},
		{"Data", reflect.TypeOf(binaryRep.Data), append([]byte(nil), binaryRep.Data...)},
	})
}

// transform byte fields into binary representation for processing
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

// transform binary representation into json
func (b BinaryRepresentation) MarshalToJson() []byte{
	jsonData, err := json.Marshal(b)

	if err != nil {
		log.Println("Error encoding JSON:", err)
		return nil
	}
	return jsonData
}

// unmarshal json to binary representation
func (b *BinaryRepresentation) UnmarshalJson(data []byte){
	err := json.Unmarshal(data, b)
	if err != nil {
		log.Println("Error decoding JSON:", err)
		return
	}
}
