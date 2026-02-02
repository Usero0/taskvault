package examples

import (
	"fmt"
	"log"
	"os"

	"github.com/taskvault/taskvault/pkg/sdk"
)

// ExampleBasicUsage demonstrates the simplest usage pattern
func ExampleBasicUsage() {
	// Initialize client
	client, err := sdk.NewClient(".taskvault/config.yaml")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// Task input and output
	taskInput := []byte("input_data.csv")
	taskOutput := []byte("computed_result_data_here")

	// First run: save to cache
	cacheKey, err := client.CacheResult("my_task", taskInput, taskOutput)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Saved to cache with key: %s\n", cacheKey)

	// Second run with same input: retrieve from cache
	cached, hit, err := client.GetCachedResult("my_task", taskInput)
	if err != nil {
		log.Fatal(err)
	}

	if hit {
		fmt.Printf("✓ Cache hit! Retrieved %d bytes instantly\n", len(cached))
	} else {
		fmt.Println("✗ Cache miss - need to recompute")
	}
}

// ExampleStatsMonitoring shows how to monitor cache health
func ExampleStatsMonitoring() {
	client, err := sdk.NewClient(".taskvault/config.yaml")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	stats, err := client.GetStats()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Cache Entries: %v\n", stats["entries"])
	fmt.Printf("Total Size: %.2f MB\n", float64(stats["total_size"].(int64))/1024/1024)
	fmt.Printf("Usage: %.1f%%\n", stats["usage_percent"])
}

// ExampleFileCaching demonstrates caching file processing
func ExampleFileCaching(inputFile, outputFile string) error {
	client, err := sdk.NewClient(".taskvault/config.yaml")
	if err != nil {
		return err
	}
	defer client.Close()

	// Read input file
	inputData, err := os.ReadFile(inputFile)
	if err != nil {
		return err
	}

	// Check if result is cached
	cached, hit, err := client.GetCachedResult("process_file", inputData)
	if err != nil {
		return err
	}

	if hit {
		// Restore from cache
		if err := os.WriteFile(outputFile, cached, 0644); err != nil {
			return err
		}
		fmt.Printf("✓ Restored from cache: %s\n", outputFile)
		return nil
	}

	// Cache miss: process the file (simulated here)
	processedData := append(inputData, []byte("\n# processed")...)

	// Save result to cache
	_, err = client.CacheResult("process_file", inputData, processedData)
	if err != nil {
		return err
	}

	// Write output
	if err := os.WriteFile(outputFile, processedData, 0644); err != nil {
		return err
	}

	fmt.Printf("✓ Processed and cached: %s\n", outputFile)
	return nil
}
