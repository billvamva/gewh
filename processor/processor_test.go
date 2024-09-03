package processor

import (
	"context"
	"gewh/core"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcess(t *testing.T) {
	// Define test data
	testData := []struct {
		batchId int
		data    string
	}{
		{1, "Helsinki;15.0,London;16.2,Lisbon;22.3"},
		{2, "Helsinki;13.2,London;17.1,Lisbon;23.1"},
		{3, "Helsinki;14.5,London;15.9,Lisbon;21.5"},
	}

	reduceFunc := WeatherReduce()

	// Define FinalReduceFunc
	finalReduceFunc := WeatherFinalReduce()

	// Process each batch
	var err error
	p := NewProcessor(
		WeatherMapFunc,
		reduceFunc,
	)
	for _, batch := range testData {
		req := core.NewRequest(batch.batchId, core.NewSerialisable(), context.Background())
		payload := core.NewPayload(uint16(1), uint16(1), []byte("origin"), []byte(batch.data))
		req.AddPayload(payload)
		err = p.Process(req)
		assert.NoError(t, err)
	}

	p.Wait()

	results := AggregateFinalResults(finalReduceFunc)

	expectedResults := map[string][3]float64{
		"Helsinki": {13.2, 15.0, 14.233333},
		"London":   {15.9, 17.1, 16.4},
		"Lisbon":   {21.5, 23.1, 22.3},
	}

	for city, expected := range expectedResults {
		actual, exists := results[city]
		assert.True(t, exists, "City %s not found in results", city)
		if exists {
			assert.InDelta(t, expected[1], actual.Max, 0.001, "Max temperature mismatch for %s", city)
			assert.InDelta(t, expected[0], actual.Min, 0.001, "Min temperature mismatch for %s", city)
			assert.InDelta(t, expected[2], actual.Avg, 0.001, "Avg temperature mismatch for %s", city)
		}
	}
}
