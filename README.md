# gewh
Message Broker implementation in Golang

### Package Overview

The `core` package provides a framework for encoding and decoding structured data into binary format. It is designed to facilitate the serialization and deserialization of messages, ensuring that data can be easily converted to and from a binary representation.

### Key Components

#### 1. `Codec` Interface
- **Purpose**: Defines the contract for managing fields in a codec.
- **Methods**:
  - `AddField(ByteField)`: Adds a field to the codec.
  - `GetFields() ByteFields`: Retrieves all fields from the codec.

#### 2. `BaseSerialisable` Interface
- **Purpose**: Defines methods for encoding and decoding data.
- **Methods**:
  - `Encode(version uint8, clientId uint16, message string)`: Encodes the given data into binary format.
  - `Decode(binaryData []byte) (ByteFields, error)`: Decodes binary data into structured fields.

#### 3. `MessageCodec` Struct
- **Purpose**: A concrete implementation of the `Codec` interface, holding a collection of fields.
- **Fields**:
  - `fields ByteFields`: A slice of `ByteField` structs.
- **Methods**:
  - `AddField(field ByteField)`: Adds a field to the codec.
  - `GetFields() ByteFields`: Returns all fields in the codec.

#### 4. `ByteField` Struct
- **Purpose**: Represents a single field with a name, data type, and value.
- **Fields**:
  - `Name string`: The name of the field.
  - `DataType reflect.Type`: The data type of the field.
  - `Value interface{}`: The value of the field.

#### 5. `ByteFields` Type
- **Purpose**: A slice of `ByteField` structs, representing a collection of fields.

#### 6. `Serialisable` Struct
- **Purpose**: Provides methods to encode and decode data using a buffer and a codec.
- **Fields**:
  - `buf *bytes.Buffer`: A buffer to hold binary data.
  - `codec Codec`: A codec to manage `ByteField` instances.
- **Methods**:
  - `Encode(version uint8, clientId uint16, message string)`: Encodes the given data into binary format.
  - `Decode(binaryData []byte) (ByteFields, error)`: Decodes binary data into fields.

#### 7. `Request` Struct
- **Purpose**: Represents a request with an ID, a message, and a response channel.
- **Fields**:
  - `Id int`: The identifier for the request.
  - `Message Serialisable`: A `Serialisable` instance containing the message.
  - `ResponseChan chan Serialisable`: A channel for sending the response.

### Encoding and Decoding Process

#### Encoding
- The `Encode` method of the `Serialisable` struct converts structured data (such as version, clientId, and message) into binary format.
- It uses a buffer to store the binary data and writes each field in little-endian format.

#### Decoding
- The `Decode` method reads binary data from a buffer and reconstructs the original fields.
- It uses reflection to dynamically handle type assertions based on the expected data types.
- The method ensures that the decoded data matches the expected structure and types.

### Example Test Cases

#### Encoding Test
- **Purpose**: Verify that the `Encode` method correctly converts fields into binary format.
- **Approach**: Initialize fields, encode them, and compare the output with the expected binary data.

#### Decoding Test
- **Purpose**: Verify that the `Decode` method correctly interprets binary data and reconstructs the original fields.
- **Approach**: Provide encoded binary data, decode it, and compare the resulting fields with the expected values. Include tests to validate error handling for incorrect types.

### Summary

The `core` package offers a robust solution for encoding and decoding structured data into binary format. It defines clear interfaces and concrete implementations for managing fields, enabling easy serialization and deserialization of messages. The test cases ensure that both the encoding and decoding processes work correctly and handle various edge cases, including incorrect data types.