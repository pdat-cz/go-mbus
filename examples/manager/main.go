// Example usage of the M-Bus Manager
//
// This example demonstrates how to use the M-Bus Manager to poll devices
// and receive data events.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pdat-cz/go-mbus/manager"
)

func main() {
	// Command line flags
	configPath := flag.String("config", "", "Path to configuration file (YAML)")
	port := flag.String("port", "/dev/ttyUSB0", "Serial port path")
	tcpHost := flag.String("tcp-host", "", "TCP host (if using TCP transport)")
	tcpPort := flag.Int("tcp-port", 10001, "TCP port")
	interval := flag.Duration("interval", 30*time.Second, "Polling interval")
	flag.Parse()

	var config manager.Config
	var err error

	// Load configuration
	if *configPath != "" {
		config, err = manager.LoadConfig(*configPath)
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}
		log.Printf("Loaded configuration from %s", *configPath)
	} else {
		// Create default configuration
		config = manager.DefaultConfig()

		// Configure transport based on flags
		if *tcpHost != "" {
			config.Transport.Type = "tcp"
			config.Transport.TCP = &manager.TCPConfig{
				Host:           *tcpHost,
				Port:           *tcpPort,
				ConnectTimeout: 10 * time.Second,
				ReadTimeout:    5 * time.Second,
				KeepAlive:      true,
			}
			config.Transport.Serial = nil
		} else {
			config.Transport.Type = "serial"
			config.Transport.Serial = &manager.SerialConfig{
				Port:          *port,
				BaudRate:      2400,
				Parity:        "even",
				ReadTimeout:   2 * time.Second,
				FallbackPorts: []string{"/dev/ttyUSB1", "/dev/ttyACM0"},
				AutoDetect:    true,
			}
		}

		config.Polling.Interval = *interval

		// Add some example devices (you would configure these for your setup)
		config.Devices = []manager.DeviceConfig{
			{Address: 1, Name: "Meter 1", Enabled: true},
			{Address: 2, Name: "Meter 2", Enabled: true},
		}
	}

	// Create the manager
	mgr, err := manager.NewManager(config)
	if err != nil {
		log.Fatalf("Failed to create manager: %v", err)
	}

	// Subscribe to data events
	mgr.OnData(func(event manager.DataEvent) {
		log.Printf("[DATA] Device %s (addr=%d):", event.DeviceName, event.DeviceAddr)
		log.Printf("  Manufacturer: %s", event.Data.Manufacturer)
		log.Printf("  Medium: %s", event.Data.Medium)
		log.Printf("  ID: %s", event.Data.IdentificationNumber)
		log.Printf("  Records: %d", len(event.Data.Records))
		log.Printf("  Latency: %v", event.ReadLatency)

		// Print each record
		for i, record := range event.Data.Records {
			log.Printf("  Record %d: %s = %s %s",
				i, record.Name, record.Value, record.Unit)
		}

		// Here you would typically:
		// - Store in database
		// - Publish to NATS/MQTT
		// - Send to monitoring system
		// Example:
		// nats.Publish("mbus.data", event)
		// db.Save(event)
	})

	// Subscribe to error events
	mgr.OnError(func(event manager.ErrorEvent) {
		log.Printf("[ERROR] %s", event.Error)
		log.Printf("  Device: %s (addr=%d)", event.DeviceName, event.DeviceAddr)
		log.Printf("  Code: %s (%d)", event.Code.String(), event.Code)
		log.Printf("  Severity: %s", event.Severity)
		log.Printf("  Suggestion: %s", event.Suggestion)

		if event.Recoverable {
			log.Printf("  Status: Attempting automatic recovery...")
		}

		// Here you would typically:
		// - Send alert
		// - Log to monitoring system
		// Example:
		// alerting.Send(event)
	})

	// Subscribe to health events
	mgr.OnHealth(func(event manager.HealthEvent) {
		log.Printf("[HEALTH] %s: %s", event.Type, event.Message)

		// Here you would typically:
		// - Update status dashboard
		// - Log health metrics
		// Example:
		// dashboard.UpdateHealth(event)
	})

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the manager
	log.Printf("Starting M-Bus Manager...")
	log.Printf("  Transport: %s", config.Transport.Type)
	if config.Transport.Type == "serial" {
		log.Printf("  Port: %s", config.Transport.Serial.Port)
	} else {
		log.Printf("  Host: %s:%d", config.Transport.TCP.Host, config.Transport.TCP.Port)
	}
	log.Printf("  Polling interval: %v", config.Polling.Interval)
	log.Printf("  Devices: %d", len(config.Devices))

	if err := mgr.Start(ctx); err != nil {
		log.Fatalf("Failed to start manager: %v", err)
	}

	// Handle shutdown signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Print status periodically
	statusTicker := time.NewTicker(60 * time.Second)
	defer statusTicker.Stop()

	log.Println("Manager running. Press Ctrl+C to stop.")

	for {
		select {
		case sig := <-sigCh:
			log.Printf("Received signal %v, shutting down...", sig)
			cancel()

			if err := mgr.Stop(); err != nil {
				log.Printf("Error stopping manager: %v", err)
			}

			log.Println("Manager stopped.")
			return

		case <-statusTicker.C:
			// Print health status
			health := mgr.GetHealth()
			fmt.Println("\n--- Status Report ---")
			fmt.Printf("State: %s\n", health.State)
			fmt.Printf("Uptime: %v\n", health.Uptime)
			fmt.Printf("Transport: %s (%s)\n", health.Transport.Type, health.Transport.Address)
			fmt.Printf("Connected: %v\n", health.Transport.Connected)
			fmt.Printf("Total Reads: %d (Success: %d, Failed: %d)\n",
				health.Stats.TotalReads,
				health.Stats.SuccessfulReads,
				health.Stats.FailedReads)
			fmt.Println("Devices:")
			for _, dev := range health.Devices {
				status := "offline"
				if dev.Online {
					status = "online"
				}
				fmt.Printf("  - %s (addr=%d): %s, reads=%d, errors=%d\n",
					dev.Name, dev.Address, status, dev.TotalReads, dev.TotalErrors)
			}
			fmt.Println("---")
		}
	}
}
