package io

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBatchReader(t *testing.T) {
	// Create a temporary CSV file for testing
	tmpfile, err := os.CreateTemp("", "test*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	// Write test data to the temp file
	testData := []string{
		"header1,header2,header3",
		"1,2,3",
		"4,5,6",
		"7,8,9",
		"10,11,12",
		"13,14,15",
	}
	for _, line := range testData {
		_, err := tmpfile.WriteString(line + "\n")
		assert.NoError(t, err)
	}
	tmpfile.Close()

	tests := []struct {
		name      string
		batchSize int
		want      [][]Record
	}{
		{
			name:      "Batch size 2",
			batchSize: 2,
			want: [][]Record{
				{{"header1", "header2", "header3"}, {"1", "2", "3"}},
				{{"4", "5", "6"}, {"7", "8", "9"}},
				{{"10", "11", "12"}, {"13", "14", "15"}},
			},
		},
		{
			name:      "Batch size 3",
			batchSize: 3,
			want: [][]Record{
				{{"header1", "header2", "header3"}, {"1", "2", "3"}, {"4", "5", "6"}},
				{{"7", "8", "9"}, {"10", "11", "12"}, {"13", "14", "15"}},
			},
		},
		{
			name:      "Batch size larger than file",
			batchSize: 10,
			want: [][]Record{
				{
					{"header1", "header2", "header3"},
					{"1", "2", "3"},
					{"4", "5", "6"},
					{"7", "8", "9"},
					{"10", "11", "12"},
					{"13", "14", "15"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, err := NewBatchReader(tmpfile.Name(), tt.batchSize)
			assert.NoError(t, err)

			var got [][]Record
			batchChan, _ := reader.ReadRecords()
			assert.NoError(t, err)
			for batch := range batchChan {
				got = append(got, batch)
			}

			assert.Equal(t, got, tt.want)
		})
	}
}
