package io

import (
	"encoding/csv"
	"fmt"
	"gewh/processor"
	"os"
	"sort"
)

// writer interface, same extendability as reader
type Writer interface {
	ProcessData(batch []Record) error
	WriteData(fileName string) error
}

type StationWriter struct {
	stations map[string]processor.Record
}

func NewStationWriter() *StationWriter {
	return &StationWriter{
		stations: make(map[string]processor.Record),
	}
}

// ProcessData now directly accepts map[string]Record
func (sw *StationWriter) ProcessData(data map[string]processor.Record) error {
	for station, record := range data {
		sw.stations[station] = record
	}
	return nil
}

func (sw *StationWriter) WriteData(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"Station", "Min", "Max", "Avg"}); err != nil {
		return fmt.Errorf("error writing header: %w", err)
	}

	// Sort station names
	var stationNames []string
	for name := range sw.stations {
		stationNames = append(stationNames, name)
	}
	sort.Strings(stationNames)

	// Write data
	for _, name := range stationNames {
		data := sw.stations[name]
		if err := writer.Write([]string{
			name,
			fmt.Sprintf("%.2f", data.Min),
			fmt.Sprintf("%.2f", data.Max),
			fmt.Sprintf("%.2f", data.Avg),
		}); err != nil {
			return fmt.Errorf("error writing data for station %s: %w", name, err)
		}
	}

	return nil
}
