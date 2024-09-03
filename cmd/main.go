package main

import (
	"context"
	"flag"
	"gewh/core"
	gio "gewh/io"
	"gewh/processor"
	"log"
	"os"
	"time"
	// ... other imports
)

func main() {
	// Your existing flags
	inputPath := flag.String("input", "/Users/vasilieiosvamvakas/Documents/projects/gewh/data/weather_data.csv", "input path in .csv format")
	batchSize := flag.Int("batch", 1000000, "input path in .csv format")
	verbose := flag.Bool("v", false, "print verbose output to console")
	outputPath := flag.String("output", "/Users/vasilieiosvamvakas/Documents/projects/gewh/data/output_data.csv", "input path in .csv format")
	numWorkers := flag.Int("workers", core.MAX_WORKER, "number of workers per dispatcher")
	queueSize := flag.Int("queue", core.MAX_QUEUE, "size of the request queue")
	flag.Parse()

	// Set up logging
	if !*verbose {
		os.Stdout = nil
	}

	startTime := time.Now()

	reader, err := gio.NewBatchReader(*inputPath, *batchSize)
	if err != nil {
		log.Fatalf("err: %v ", err)
	}

	producer := core.NewProducer(core.WithBroadcastTimeout[core.Request](5 * time.Second))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go producer.Start(ctx)

	dispatcher := core.NewDispatcher(1, *numWorkers)
	queue := make(core.RequestQueue, *queueSize)
	dispatcher.AddQueue(queue)
	producer.Subscribe(dispatcher)

	p := processor.NewProcessor(processor.WeatherMapFunc, processor.WeatherReduce())
	dispatcher.Run(p)

	processingStartTime := time.Now()
	batchChan, _ := reader.ReadRecords()
	for rawBatch := range batchChan {
		batch := gio.CombineRecords(rawBatch)
		req := core.NewRequest(int(batch.Id), core.NewSerialisable(), context.Background())
		payload := core.NewPayload(uint16(1), uint16(1), []byte("origin"), []byte(batch.Value))
		req.AddPayload(payload)
		producer.Broadcast(ctx, req)
	}
	p.Wait()
	processingEndTime := time.Now()

	dispatcher.Stop()

	aggregationStartTime := time.Now()
	results := processor.AggregateFinalResults(processor.WeatherFinalReduce())
	aggregationEndTime := time.Now()

	writer := gio.NewStationWriter()
	writer.ProcessData(results)
	writer.WriteData(*outputPath)

	endTime := time.Now()

	// Print timing information
	log.Printf("Total execution time: %v", endTime.Sub(startTime))
	log.Printf("Processing time: %v", processingEndTime.Sub(processingStartTime))
	log.Printf("Aggregation time: %v", aggregationEndTime.Sub(aggregationStartTime))
}
