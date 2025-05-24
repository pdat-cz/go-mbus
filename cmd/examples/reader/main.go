package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/pdat-cz/go-mbus"
)

func main() {
	// Parse command-line arguments
	port := flag.String("port", "/dev/ttyUSB0", "Serial port connected to M-Bus")
	address := flag.Int("address", 1, "M-Bus device address")
	outputFormat := flag.String("format", "json", "Output format (json or text)")
	flag.Parse()

	// Read data from the device
	fmt.Printf("Reading M-Bus device at address %d on port %s...\n", *address, *port)
	deviceState := mbus.Read(*port, *address)

	// Check for errors
	if deviceState.Error != "" {
		fmt.Fprintf(os.Stderr, "Error reading device: %s\n", deviceState.Error)
		os.Exit(1)
	}

	// Output the data in the requested format
	if *outputFormat == "json" {
		// Output as JSON
		jsonData, err := json.MarshalIndent(deviceState, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling JSON: %s\n", err)
			os.Exit(1)
		}
		fmt.Println(string(jsonData))
	} else {
		// Output as text
		fmt.Printf("Device State:\n")
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
