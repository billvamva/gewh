package processor

import (
	"bytes"
	"gewh/core"
	"strings"
)

type (
	MapFunc func(*core.Payload) []KeyValue
)

type KeyValue struct {
	Key   string
	Value []byte
}

func wordCountMapFunc(payload *core.Payload) []KeyValue {
	words := bytes.Fields(payload.Data)

	kvs := make([]KeyValue, len(words))

	for i, word := range words {
		kvs[i] = KeyValue{
			Key:   string(word),
			Value: []byte{1},
		}
	}
	return kvs
}

func WeatherMapFunc(payload *core.Payload) []KeyValue {
	pairs := strings.Split(string(payload.Data), ",")

	kvs := make([]KeyValue, len(pairs))

	for i, v := range pairs {
		info := strings.SplitN(v, ";", 2)
		kvs[i] = KeyValue{
			Key:   info[0],
			Value: []byte(info[1]),
		}
	}

	return kvs
}
