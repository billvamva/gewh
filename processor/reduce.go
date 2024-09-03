package processor

import (
	"encoding/binary"
	"math"
	"sort"
	"strconv"
	"sync"
)

type (
	ReduceFunc func([][]byte) ([]byte, error)
)

// groupByKey groups the KeyValue pairs by their keys
func groupByKey(kvs []KeyValue) map[string][][]byte {
	groups := make(map[string][][]byte)
	for _, kv := range kvs {
		groups[kv.Key] = append(groups[kv.Key], kv.Value)
	}
	return groups
}

// reduce applies the reduce function to each group
func reduce(groups map[string][][]byte, reduceFunc ReduceFunc) []KeyValue {
	var results []KeyValue
	var wg sync.WaitGroup
	resultChan := make(chan KeyValue, len(groups))

	for key, values := range groups {
		wg.Add(1)
		go func(k string, v [][]byte) {
			defer wg.Done()
			reduced, _ := reduceFunc(v)
			resultChan <- KeyValue{Key: k, Value: reduced}
		}(key, values)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for result := range resultChan {
		results = append(results, result)
	}

	return results
}

// combineResults merges the reduced results back into a single byte slice
func combineResults(results []KeyValue) []byte {
	// Sort results by key for consistency
	sort.Slice(results, func(i, j int) bool {
		return results[i].Key < results[j].Key
	})
	var combined []byte
	for _, result := range results {
		combined = append(combined, result.Value...)
	}
	return combined
}

func convertTemperatures(value [][]byte) ([]byte, error) {
	output := make([]float64, 0, len(value))
	var val float64
	var err error
	for i := 0; i < len(value); i++ {
		if val, err = strconv.ParseFloat(string(value[i]), 64); err != nil {
			return nil, err
		}

		output = append(output, val)
	}

	return ConvertFloat64SliceToBytes(output), nil
}

func weatherFinalReduceFunc(values [][]byte) ([]byte, error) {
	temperatures := make([]float64, 0, len(values)*4)

	for _, value := range values {
		temps := ConvertBytesToFloat64Slice(value)
		temperatures = append(temperatures, temps...)
	}

	if len(temperatures) == 0 {
		return []byte{}, nil
	}

	sort.Float64s(temperatures)

	min := temperatures[0]
	max := temperatures[len(temperatures)-1]
	sum := 0.0
	for _, temp := range temperatures {
		sum += temp
	}
	avg := sum / float64(len(temperatures))

	// Pack min, max, avg into a byte slice
	result := make([]byte, 24) // 3 float64 values, 8 bytes each
	binary.LittleEndian.PutUint64(result[0:8], math.Float64bits(min))
	binary.LittleEndian.PutUint64(result[8:16], math.Float64bits(max))
	binary.LittleEndian.PutUint64(result[16:24], math.Float64bits(avg))

	return result, nil
}

func ParseResult(data []byte) Record {
	min := math.Float64frombits(binary.LittleEndian.Uint64(data[0:8]))
	max := math.Float64frombits(binary.LittleEndian.Uint64(data[8:16]))
	avg := math.Float64frombits(binary.LittleEndian.Uint64(data[16:24]))
	return Record{
		min,
		max,
		avg,
	}
}

// Wrapper function to use weatherFinalReduceFunc as a ReduceFunc
func WeatherReduce() ReduceFunc {
	return func(values [][]byte) ([]byte, error) {
		return convertTemperatures(values)
	}
}

// Wrapper function to use weatherFinalReduceFunc as a ReduceFunc
func WeatherFinalReduce() ReduceFunc {
	return func(values [][]byte) ([]byte, error) {
		return weatherFinalReduceFunc(values)
	}
}
