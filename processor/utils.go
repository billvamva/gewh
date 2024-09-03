package processor

import (
	"bytes"
	"encoding/binary"
	"sort"
)

// ConvertFloat64SliceToBytes converts a slice of float64 to a single byte slice
func ConvertFloat64SliceToBytes(floats []float64) []byte {
	buf := new(bytes.Buffer)
	for _, f := range floats {
		err := binary.Write(buf, binary.LittleEndian, f)
		if err != nil {
			panic(err) // In real code, handle this error more gracefully
		}
	}
	return buf.Bytes()
}

// ConvertBytesToFloat64Slice converts a byte slice back to a slice of float64
func ConvertBytesToFloat64Slice(b []byte) []float64 {
	floats := make([]float64, 0, len(b)/8)
	buf := bytes.NewReader(b)
	for {
		var f float64
		err := binary.Read(buf, binary.LittleEndian, &f)
		if err != nil {
			break // End of buffer
		}
		floats = append(floats, f)
	}
	return floats
}

func SortKeyValues(kvs []KeyValue) {
	sort.Slice(kvs, func(i, j int) bool {
		return kvs[i].Key < kvs[j].Key
	})
}
