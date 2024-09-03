package processor

import (
	"gewh/core"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapFunction(t *testing.T) {
	t.Run("Test simple word count for payload", func(t *testing.T) {
		payload := core.NewPayload(1, 1, []byte("token"), []byte("hello world"))

		expected := []KeyValue{
			{
				Key:   "hello",
				Value: []byte{1},
			},
			{
				Key:   "world",
				Value: []byte{1},
			},
		}

		actual := wordCountMapFunc(payload)

		assert.Equal(t, expected, actual)
	})

	t.Run("Test with sample weather data", func(t *testing.T) {
		weatherDataString := "Helsinski;15.0,London;16.2,Helsinski;13.2,Lisbon;12.1"
		payload := core.NewPayload(1, 1, []byte("token"), []byte(weatherDataString))

		expected := []KeyValue{
			{
				Key:   "Helsinski",
				Value: []byte("15.0"),
			},
			{
				Key:   "London",
				Value: []byte("16.2"),
			},
			{
				Key:   "Helsinski",
				Value: []byte("13.2"),
			},
			{
				Key:   "Lisbon",
				Value: []byte("12.1"),
			},
		}

		actual := WeatherMapFunc(payload)

		assert.Equal(t, expected, actual)
	})
}

func TestReduceFunction(t *testing.T) {
	t.Run("testing grouping by key and simple reduce for simple word count", func(t *testing.T) {
		kvs := []KeyValue{
			{
				Key:   "hello",
				Value: []byte{1},
			},
			{
				Key:   "world",
				Value: []byte{1},
			},
			{
				Key:   "hello",
				Value: []byte{1},
			},
		}

		expected := map[string][][]byte{
			"hello": {{1}, {1}},
			"world": {{1}},
		}

		groups := groupByKey(kvs)

		assert.Equal(t, expected, groups)

		sum := ReduceFunc(func(value [][]byte) ([]byte, error) {
			output := 0
			for i := 0; i < len(value); i++ {
				output++
			}
			return []byte{byte(output)}, nil
		})

		expectedKvs := []KeyValue{
			{
				Key:   "world",
				Value: []byte{1},
			},
			{
				Key:   "hello",
				Value: []byte{2},
			},
		}

		processedKvs := reduce(groups, sum)
		SortKeyValues(expectedKvs)
		SortKeyValues(processedKvs)

		assert.Equal(t, expectedKvs, processedKvs)
	})

	t.Run("testing grouping by key and simple reduce for weather data", func(t *testing.T) {
		kvs := []KeyValue{
			{
				Key:   "Helsinski",
				Value: []byte("15.0"),
			},
			{
				Key:   "London",
				Value: []byte("16.2"),
			},
			{
				Key:   "Helsinski",
				Value: []byte("13.2"),
			},
			{
				Key:   "Lisbon",
				Value: []byte("12.1"),
			},
		}

		expected := map[string][][]byte{
			"Helsinski": {[]byte("15.0"), []byte("13.2")},
			"London":    {[]byte("16.2")},
			"Lisbon":    {[]byte("12.1")},
		}

		groups := groupByKey(kvs)

		assert.Equal(t, expected, groups)

		expectedKvs := []KeyValue{
			{
				Key:   "London",
				Value: ConvertFloat64SliceToBytes([]float64{16.2}),
			},
			{
				Key:   "Lisbon",
				Value: ConvertFloat64SliceToBytes([]float64{12.1}),
			},
			{
				Key:   "Helsinski",
				Value: ConvertFloat64SliceToBytes([]float64{15.0, 13.2}),
			},
		}

		processedKvs := reduce(groups, WeatherReduce())

		SortKeyValues(expectedKvs)
		SortKeyValues(processedKvs)

		assert.Equal(t, expectedKvs, processedKvs)
	})

	t.Run("testing final reduce for weather data across multiple batches", func(t *testing.T) {
		// Simulate data from multiple batches
		helsinkiBatch1 := ConvertFloat64SliceToBytes([]float64{15.0, 13.2})
		helsinkiBatch2 := ConvertFloat64SliceToBytes([]float64{14.5, 16.8})
		londonBatch1 := ConvertFloat64SliceToBytes([]float64{16.2, 17.1})
		londonBatch2 := ConvertFloat64SliceToBytes([]float64{15.9, 18.0})
		lisbonBatch1 := ConvertFloat64SliceToBytes([]float64{22.3, 23.1})
		lisbonBatch2 := ConvertFloat64SliceToBytes([]float64{21.5, 24.2})

		finalReduceFunc := WeatherFinalReduce()

		testCases := []struct {
			name     string
			key      string
			values   [][]byte
			expected [3]float64 // [min, max, avg]
		}{
			{
				name:   "Helsinki",
				key:    "Helsinki",
				values: [][]byte{helsinkiBatch1, helsinkiBatch2},
				expected: [3]float64{
					13.2,                              // min
					16.8,                              // max
					(15.0 + 13.2 + 14.5 + 16.8) / 4.0, // avg
				},
			},
			{
				name:   "London",
				key:    "London",
				values: [][]byte{londonBatch1, londonBatch2},
				expected: [3]float64{
					15.9,                              // min
					18.0,                              // max
					(16.2 + 17.1 + 15.9 + 18.0) / 4.0, // avg
				},
			},
			{
				name:   "Lisbon",
				key:    "Lisbon",
				values: [][]byte{lisbonBatch1, lisbonBatch2},
				expected: [3]float64{
					21.5,                              // min
					24.2,                              // max
					(22.3 + 23.1 + 21.5 + 24.2) / 4.0, // avg
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result, _ := finalReduceFunc(tc.values)
				record := ParseResult(result)

				assert.InDelta(t, tc.expected[0], record.Min, 0.001, "Minimum temperature mismatch")
				assert.InDelta(t, tc.expected[1], record.Max, 0.001, "Maximum temperature mismatch")
				assert.InDelta(t, tc.expected[2], record.Avg, 0.001, "Average temperature mismatch")
			})
		}
	})
}
