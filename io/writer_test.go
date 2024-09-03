package io

import (
	"encoding/csv"
	"fmt"
	"gewh/processor"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStationWriter(t *testing.T) {
	tests := []struct {
		name          string
		inputData     map[string]processor.Record
		expectedData  map[string]processor.Record
		expectedOrder []string
	}{
		{
			name: "Multiple stations",
			inputData: map[string]processor.Record{
				"StationB": {Min: 5.0, Max: 9.0, Avg: 7.0},
				"StationA": {Min: 10.0, Max: 20.0, Avg: 15.0},
				"StationC": {Min: 100.0, Max: 200.0, Avg: 150.0},
			},
			expectedData: map[string]processor.Record{
				"StationA": {Min: 10.0, Max: 20.0, Avg: 15.0},
				"StationB": {Min: 5.0, Max: 9.0, Avg: 7.0},
				"StationC": {Min: 100.0, Max: 200.0, Avg: 150.0},
			},
			expectedOrder: []string{"StationA", "StationB", "StationC"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := NewStationWriter()
			err := writer.ProcessData(tt.inputData)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedData, writer.stations)

			// Test WriteData
			tmpfile, err := os.CreateTemp("", "test*.csv")
			assert.NoError(t, err)
			defer os.Remove(tmpfile.Name())

			err = writer.WriteData(tmpfile.Name())
			assert.NoError(t, err)

			// Read the written file and check its contents
			file, err := os.Open(tmpfile.Name())
			assert.NoError(t, err)
			defer file.Close()

			csvReader := csv.NewReader(file)
			records, err := csvReader.ReadAll()
			assert.NoError(t, err)

			assert.Equal(t, len(tt.expectedData)+1, len(records), "Unexpected number of records") // +1 for header

			assert.Equal(t, []string{"Station", "Min", "Max", "Avg"}, records[0], "Unexpected header")

			for i, station := range tt.expectedOrder {
				expected := tt.expectedData[station]
				got := records[i+1]
				want := []string{
					station,
					fmt.Sprintf("%.2f", expected.Min),
					fmt.Sprintf("%.2f", expected.Max),
					fmt.Sprintf("%.2f", expected.Avg),
				}
				assert.Equal(t, want, got)
			}
		})
	}
}

// We don't need TestStationWriterErrors anymore because ProcessData doesn't parse strings
// and doesn't return errors in the new implementation. However, we can add a test for WriteData errors:

func TestStationWriterWriteDataError(t *testing.T) {
	writer := NewStationWriter()
	writer.stations = map[string]processor.Record{
		"StationA": {Min: 10.0, Max: 20.0, Avg: 15.0},
	}

	// Try to write to a directory that doesn't exist
	err := writer.WriteData("/nonexistent/directory/file.csv")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error creating file")
}
