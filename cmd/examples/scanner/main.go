package main

import (
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/pdat-cz/go-mbus"
)

func main() {
	// Parse command-line arguments
	port := flag.String("port", "/dev/ttyUSB0", "Serial port connected to M-Bus")
	startAddr := flag.Int("start", 1, "Start address for scanning")
	endAddr := flag.Int("end", 250, "End address for scanning")
	concurrent := flag.Int("concurrent", 1, "Number of concurrent scans")
	timeout := flag.Duration("timeout", 5*time.Second, "Timeout for each device scan")
	readData := flag.Bool("read", false, "Read data from found devices")
	flag.Parse()

	// Validate input
	if *startAddr < 1 || *startAddr > 250 {
		fmt.Fprintf(os.Stderr, "Start address must be between 1 and 250\n")
		os.Exit(1)
	}
	if *endAddr < *startAddr || *endAddr > 250 {
		fmt.Fprintf(os.Stderr, "End address must be between start address and 250\n")
		os.Exit(1)
	}
	if *concurrent < 1 {
		fmt.Fprintf(os.Stderr, "Concurrent scans must be at least 1\n")
		os.Exit(1)
	}

	fmt.Printf("Scanning M-Bus devices on port %s (addresses %d-%d)...\n",
		*port, *startAddr, *endAddr)

	// Create a channel for addresses to scan
	addresses := make(chan int, *endAddr-*startAddr+1)

	// Create a channel for results
	results := make(chan int, *endAddr-*startAddr+1)

	// Create a wait group to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < *concurrent; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for addr := range addresses {
				// Set a timeout for the scan
				done := make(chan bool, 1)
				go func() {
					pingState := mbus.Ping(*port, addr)
					if pingState.State {
						results <- addr
					}
					done <- true
				}()

				// Wait for the scan to complete or timeout
				select {
				case <-done:
					// Scan completed
				case <-time.After(*timeout):
					fmt.Printf("Timeout scanning address %d\n", addr)
				}
			}
		}()
	}

	// Send addresses to scan
	go func() {
		for addr := *startAddr; addr <= *endAddr; addr++ {
			addresses <- addr
		}
		close(addresses)
	}()

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect and display results
	foundDevices := []int{}
	for addr := range results {
		foundDevices = append(foundDevices, addr)
		fmt.Printf("Found device at address %d\n", addr)
	}

	fmt.Printf("Scan complete. Found %d devices.\n", len(foundDevices))

	// Read data from found devices if requested
	if *readData && len(foundDevices) > 0 {
		fmt.Println("\nReading data from found devices:")
		for _, addr := range foundDevices {
			fmt.Printf("\nReading device at address %d...\n", addr)
			deviceState := mbus.Read(*port, addr)

			if deviceState.Error != "" {
				fmt.Printf("Error reading device: %s\n", deviceState.Error)
				continue
			}

			fmt.Printf("Device data:\n")
			fmt.Printf("  Port: %s\n", deviceState.Port)
			fmt.Printf("  Address: %d\n", deviceState.Address)
			fmt.Printf("  Timestamp: %s\n", deviceState.Timestamp.Format("2006-01-02 15:04:05"))

			if len(deviceState.Data.Records) > 0 {
				fmt.Printf("  Data Records:\n")
				for i, record := range deviceState.Data.Records {
					fmt.Printf("    Record %d:\n", i+1)
					fmt.Printf("      Description: %s\n", record.Description)
					fmt.Printf("      Value: %s\n", record.Value)
					fmt.Printf("      Unit: %s\n", record.Unit)
				}
			} else {
				fmt.Printf("  No data records available\n")
			}
		}
	}
}
