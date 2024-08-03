# gewh
Message Broker implementation in Golang

### Package Overview

The `core` package provides a framework for encoding and decoding structured data into binary format. It is designed to facilitate the serialization and deserialization of messages, ensuring that data can be easily converted to and from a binary representation. It also provides a Publisher-Subcriber Message Broker to process the data efficiently. Messages will be mantained in binary format.

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

#### Queue and Worker testing
- **Purpose**: Verify that the `Worker` method correctly receives messages from the queue and processes them..
- **Approach**: Add elements to the queue and then received them from the channel and dispatch them to the processing worker. Provide context for early cancellation.


### Documentation for Types and Functions

#### Interface: `BaseSerialisable`
Defines methods for encoding and decoding data for a serializable.

```go
type BaseSerialisable interface {
    Encode()                     // Encode data from the codec fields to the buffer
    Decode() (ByteFields, error) // Decode data from buffer to byte fields
}
```

#### Struct: `BinaryRepresentation`
Struct representation of a serialisable's buffer content using its codec. This format is used as an interface with the serialisable.

```go
type BinaryRepresentation struct {
    Version  uint16 `json:"version"`  // version of encoding
    ClientId uint16 `json:"clientId"` // origin client id
    Token    []byte `json:"token"`    // authentication token
    Data     []byte `json:"data"`     // data sent
}
```

##### Method: `FormatDecodedFields`
Transforms byte fields into binary representation for processing.

```go
func (b *BinaryRepresentation) FormatDecodedFields(decodedFields ByteFields)
```

##### Method: `MarshalToJson`
Transforms binary representation into JSON.

```go
func (b BinaryRepresentation) MarshalToJson() []byte
```

##### Method: `UnmarshalJson`
Unmarshals JSON to binary representation.

```go
func (b *BinaryRepresentation) UnmarshalJson(data []byte)
```

#### Struct: `ByteField`
Data has to be first transformed to a ByteField to be encoded and written on the codec.

```go
type ByteField struct {
    Name     string       // name of the field
    DataType reflect.Type // data type of the field
    Value    interface{}  // Holds pointer of value
}
```

#### Type: `ByteFields`
A slice of `ByteField`.

```go
type ByteFields []ByteField
```

#### Interface: `Codec`
Defines the contract for managing fields in a codec.

```go
type Codec interface {
    AddFields(ByteFields)  // Adds Fields for the specific codec
    GetFields() ByteFields // Gets all Fields for the specific codec
}
```

#### Struct: `MessageCodec`
Codec for base messages we will be handling.

```go
type MessageCodec struct {
    // Has unexported fields.
}
```

##### Method: `AddFields`
Adds fields to the codec.

```go
func (c *MessageCodec) AddFields(fields ByteFields)
```

##### Method: `GetFields`
Gets fields from the codec.

```go
func (c *MessageCodec) GetFields() ByteFields
```

#### Struct: `Request`
Request is the format that our serialisable messages are going to be sent as to the message broker.

```go
type Request struct {
    Id      int             // id of the request
    Message Serialisable    // message of serialisable form
    Ctx     context.Context // context to keep track of cancelled requests and remove from the message broker
}
```

##### Function: `NewRequest`
Creates a new request.

```go
func NewRequest(id int, message Serialisable, ctx context.Context) *Request
```

##### Method: `Process`
Placeholder for actual request processing.

```go
func (req *Request) Process()
```

#### Type: `RequestQueue`
Single queue message broker, to be expanded.

```go
type RequestQueue chan *Request
```

#### Struct: `Serialisable`
Anything that can be serialised and deserialised.

```go
type Serialisable struct {
    // Has unexported fields.
}
```

##### Method: `BinaryRepresentationToByteFields`
Transforms the binary struct representation into fields and then adds it to the codec of the serialisable.

```go
func (s *Serialisable) BinaryRepresentationToByteFields(binaryRep *BinaryRepresentation)
```

##### Method: `Decode`
Decode method implementation for `Serialisable`.

```go
func (s *Serialisable) Decode() (ByteFields, error)
```

##### Method: `Encode`
Encode method implementation for `Serialisable`.

```go
func (s *Serialisable) Encode()
```

##### Method: `InsertDataToSerialisableBuffer`
Inserts raw binary data into the buffer.

```go
func (s *Serialisable) InsertDataToSerialisableBuffer(binaryData []byte)
```

#### Struct: `Worker`
Worker that interacts with the message broker.

```go
type Worker struct {
    WorkerPool     chan chan *Request // each worker corresponds to a worker pool that holds request channels
    RequestChannel chan *Request      // worker's request channel
    // Has unexported fields.
}
```

##### Function: `NewWorker`
Creates a new worker.

```go
func NewWorker(workerPool chan chan *Request) Worker
```

##### Method: `Start`
Registers worker's request channel to the pool and waits for requests or quit signal on the request channel.

```go
func (w Worker) Start(id int)
```

##### Method: `Stop`
Stops the worker from listening for work requests.

```go
func (w Worker) Stop()
```
### Summary

The `core` package offers a robust solution for encoding and decoding structured data into binary format. It defines clear interfaces and concrete implementations for managing fields, enabling easy serialization and deserialization of messages. The test cases ensure that both the encoding and decoding processes work correctly and handle various edge cases, including incorrect data types. It also provides an efficient format for handling the processing of the requests in the queue.