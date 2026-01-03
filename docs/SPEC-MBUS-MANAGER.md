# M-Bus Manager Specification

## Overview

A complete, self-healing M-Bus communication manager that provides:
- Unified interface for Serial and TCP/IP communication
- Configurable polling with automatic scheduling
- Self-healing capabilities with rich diagnostics
- Event-driven data delivery for flexible integration

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        MBusManager                               │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │   Config    │  │  Scheduler  │  │    Self-Healing Engine  │  │
│  │   Loader    │  │  (Polling)  │  │   (Diagnostics/Recovery)│  │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘  │
├─────────────────────────────────────────────────────────────────┤
│                    Transport Layer                               │
│  ┌─────────────────────┐    ┌─────────────────────┐             │
│  │   SerialTransport   │    │    TCPTransport     │             │
│  │   /dev/ttyUSB0      │    │   192.168.1.1:10001 │             │
│  └─────────────────────┘    └─────────────────────┘             │
├─────────────────────────────────────────────────────────────────┤
│                    Device Registry                               │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐           │
│  │ Device 1 │ │ Device 2 │ │ Device 3 │ │ Device N │           │
│  │ addr: 1  │ │ addr: 2  │ │ addr: 5  │ │ addr: N  │           │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘           │
├─────────────────────────────────────────────────────────────────┤
│                    Event Bus (Channels)                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │
│  │ DataEvents  │  │ ErrorEvents │  │HealthEvents│              │
│  └─────────────┘  └─────────────┘  └─────────────┘              │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
              ┌───────────────────────────────┐
              │     Developer Application     │
              │  (Persistence, NATS, MQTT...) │
              └───────────────────────────────┘
```

## Core Components

### 1. Configuration

```go
type Config struct {
    // Transport configuration
    Transport TransportConfig `yaml:"transport"`

    // Polling configuration
    Polling PollingConfig `yaml:"polling"`

    // Devices to monitor
    Devices []DeviceConfig `yaml:"devices"`

    // Self-healing configuration
    SelfHealing SelfHealingConfig `yaml:"self_healing"`
}

type TransportConfig struct {
    // Type: "serial" or "tcp"
    Type string `yaml:"type"`

    // Serial port settings
    Serial *SerialConfig `yaml:"serial,omitempty"`

    // TCP settings
    TCP *TCPConfig `yaml:"tcp,omitempty"`
}

type SerialConfig struct {
    // Port path: "/dev/ttyUSB0", "COM1", etc.
    Port string `yaml:"port"`

    // Baud rate (default: 2400 for M-Bus)
    BaudRate int `yaml:"baud_rate"`

    // Auto-detect settings if connection fails
    AutoDetect bool `yaml:"auto_detect"`

    // Alternative ports to try on failure
    FallbackPorts []string `yaml:"fallback_ports"`
}

type TCPConfig struct {
    // Host address
    Host string `yaml:"host"`

    // Port number
    Port int `yaml:"port"`

    // Connection timeout
    Timeout time.Duration `yaml:"timeout"`

    // Keep-alive settings
    KeepAlive bool `yaml:"keep_alive"`
}

type PollingConfig struct {
    // Polling interval
    Interval time.Duration `yaml:"interval"`

    // Stagger device reads to avoid bus congestion
    StaggerDelay time.Duration `yaml:"stagger_delay"`

    // Retry count per device before marking as failed
    RetryCount int `yaml:"retry_count"`

    // Retry delay between attempts
    RetryDelay time.Duration `yaml:"retry_delay"`
}

type DeviceConfig struct {
    // Device address (1-250)
    Address int `yaml:"address"`

    // Human-readable name
    Name string `yaml:"name"`

    // Device-specific polling interval (overrides global)
    Interval *time.Duration `yaml:"interval,omitempty"`

    // Enable/disable device
    Enabled bool `yaml:"enabled"`
}

type SelfHealingConfig struct {
    // Enable self-healing
    Enabled bool `yaml:"enabled"`

    // Maximum recovery attempts before giving up
    MaxRecoveryAttempts int `yaml:"max_recovery_attempts"`

    // Delay between recovery attempts
    RecoveryDelay time.Duration `yaml:"recovery_delay"`

    // Auto-scan for baud rate
    AutoBaudDetection bool `yaml:"auto_baud_detection"`

    // Auto-scan for alternative ports
    AutoPortScan bool `yaml:"auto_port_scan"`
}
```

### 2. Event System

```go
// DataEvent represents successful meter reading
type DataEvent struct {
    Timestamp   time.Time
    DeviceAddr  int
    DeviceName  string
    Data        LFrameParsed
    ReadLatency time.Duration
}

// ErrorEvent represents a communication error
type ErrorEvent struct {
    Timestamp   time.Time
    DeviceAddr  int
    DeviceName  string
    Error       error
    ErrorCode   ErrorCode
    Severity    Severity
    Suggestion  string      // Human-readable fix suggestion
    Context     ErrorContext
}

// HealthEvent represents system health changes
type HealthEvent struct {
    Timestamp    time.Time
    EventType    HealthEventType
    Transport    string
    Message      string
    Details      map[string]interface{}
}

// Error codes for rich diagnostics
type ErrorCode int

const (
    ErrPortNotFound ErrorCode = iota + 1000
    ErrPortBusy
    ErrPortAccessDenied
    ErrPortDisconnected
    ErrBaudRateMismatch
    ErrParityError
    ErrFrameError
    ErrChecksumMismatch
    ErrTimeout
    ErrDeviceNotResponding
    ErrInvalidResponse
    ErrTCPConnectionFailed
    ErrTCPConnectionLost
    ErrTCPTimeout
)

// ErrorContext provides debugging information
type ErrorContext struct {
    Transport      string
    LastSuccessful time.Time
    FailureCount   int
    RawBytes       []byte
    AttemptedFixes []string
}
```

### 3. Manager Interface

```go
type MBusManager interface {
    // Start begins polling and event processing
    Start(ctx context.Context) error

    // Stop gracefully shuts down the manager
    Stop() error

    // Subscribe to data events
    OnData(handler func(DataEvent))

    // Subscribe to error events
    OnError(handler func(ErrorEvent))

    // Subscribe to health events
    OnHealth(handler func(HealthEvent))

    // Manual device operations
    ReadDevice(address int) (DataEvent, error)
    PingDevice(address int) (bool, error)

    // Device management
    AddDevice(config DeviceConfig) error
    RemoveDevice(address int) error
    ListDevices() []DeviceStatus

    // Health and diagnostics
    GetHealth() HealthStatus
    RunDiagnostics() DiagnosticsReport

    // Configuration
    UpdateConfig(config Config) error
    GetConfig() Config
}
```

### 4. Self-Healing Engine

The self-healing engine automatically diagnoses and recovers from common issues:

#### Serial Port Issues

| Problem | Detection | Recovery Action |
|---------|-----------|-----------------|
| Port not found | `ENOENT` error | Scan for available ports, try fallbacks |
| Port busy | `EBUSY` error | Wait and retry with exponential backoff |
| Permission denied | `EACCES` error | Report with fix suggestion (chmod/udev) |
| Port disconnected | Read/Write fails | Attempt reconnect, try fallback ports |
| Wrong baud rate | Garbage/timeout | Auto-detect baud rate (2400, 9600, 19200) |
| Parity mismatch | Frame errors | Try different parity settings |

#### TCP/IP Issues

| Problem | Detection | Recovery Action |
|---------|-----------|-----------------|
| Connection refused | `ECONNREFUSED` | Retry with backoff, report gateway issue |
| Connection timeout | Dial timeout | Increase timeout, check network |
| Connection lost | Read/Write fails | Reconnect with exponential backoff |
| Gateway overloaded | Slow responses | Reduce polling frequency |

#### Device Issues

| Problem | Detection | Recovery Action |
|---------|-----------|-----------------|
| Device not responding | No ACK | Retry, then mark device offline |
| Invalid checksum | CRC mismatch | Retry read, report if persistent |
| Corrupted data | Parse failure | Retry, log raw bytes for analysis |
| Device address changed | Device missing | Optional: scan for device |

#### Recovery Strategy

```go
type RecoveryStrategy struct {
    // Exponential backoff settings
    InitialDelay   time.Duration
    MaxDelay       time.Duration
    Multiplier     float64

    // Recovery actions to try in order
    Actions []RecoveryAction
}

type RecoveryAction interface {
    Name() string
    Execute(ctx context.Context) error
    Rollback() error
}

// Example recovery actions:
// - TryAlternativePort
// - TryAlternativeBaudRate
// - TryAlternativeParity
// - ResetConnection
// - RestartTransport
// - ScanForDevices
```

### 5. Transport Abstraction

```go
type Transport interface {
    // Open establishes the connection
    Open() error

    // Close terminates the connection
    Close() error

    // IsOpen returns connection state
    IsOpen() bool

    // Send transmits a command and receives response
    Send(command []byte) ([]byte, error)

    // GetInfo returns transport details
    GetInfo() TransportInfo

    // SetConfig updates transport configuration
    SetConfig(config interface{}) error
}

type TransportInfo struct {
    Type        string            // "serial" or "tcp"
    Address     string            // "/dev/ttyUSB0" or "192.168.1.1:10001"
    State       TransportState
    Stats       TransportStats
    LastError   error
}

type TransportStats struct {
    BytesSent     uint64
    BytesReceived uint64
    Errors        uint64
    Reconnects    uint64
    Uptime        time.Duration
}
```

## Usage Example

```go
package main

import (
    "context"
    "log"
    "time"

    mbus "github.com/pdat-cz/go-mbus/manager"
)

func main() {
    // Create configuration
    config := mbus.Config{
        Transport: mbus.TransportConfig{
            Type: "serial",
            Serial: &mbus.SerialConfig{
                Port:          "/dev/ttyUSB0",
                BaudRate:      2400,
                AutoDetect:    true,
                FallbackPorts: []string{"/dev/ttyUSB1", "/dev/ttyACM0"},
            },
        },
        Polling: mbus.PollingConfig{
            Interval:     30 * time.Second,
            StaggerDelay: 500 * time.Millisecond,
            RetryCount:   3,
            RetryDelay:   1 * time.Second,
        },
        Devices: []mbus.DeviceConfig{
            {Address: 1, Name: "Water Meter - Building A", Enabled: true},
            {Address: 2, Name: "Gas Meter - Building A", Enabled: true},
            {Address: 5, Name: "Heat Meter - Building B", Enabled: true},
        },
        SelfHealing: mbus.SelfHealingConfig{
            Enabled:             true,
            MaxRecoveryAttempts: 5,
            RecoveryDelay:       5 * time.Second,
            AutoBaudDetection:   true,
            AutoPortScan:        true,
        },
    }

    // Create manager
    manager, err := mbus.NewManager(config)
    if err != nil {
        log.Fatal(err)
    }

    // Subscribe to data events
    manager.OnData(func(event mbus.DataEvent) {
        log.Printf("[DATA] Device %s (addr=%d): %d records received",
            event.DeviceName, event.DeviceAddr, len(event.Data.Records))

        // Send to NATS, persist to DB, etc.
        // nats.Publish("mbus.data", event)
        // db.Save(event)
    })

    // Subscribe to error events
    manager.OnError(func(event mbus.ErrorEvent) {
        log.Printf("[ERROR] Device %s: %s (Code: %d)",
            event.DeviceName, event.Error, event.ErrorCode)
        log.Printf("  Suggestion: %s", event.Suggestion)

        // Alert monitoring system
        // alerting.Send(event)
    })

    // Subscribe to health events
    manager.OnHealth(func(event mbus.HealthEvent) {
        log.Printf("[HEALTH] %s: %s", event.EventType, event.Message)

        // Update status dashboard
        // dashboard.UpdateHealth(event)
    })

    // Start the manager
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    if err := manager.Start(ctx); err != nil {
        log.Fatal(err)
    }

    // Run until interrupted
    select {}
}
```

## Configuration File Example

```yaml
# mbus-config.yaml
transport:
  type: serial
  serial:
    port: /dev/ttyUSB0
    baud_rate: 2400
    auto_detect: true
    fallback_ports:
      - /dev/ttyUSB1
      - /dev/ttyACM0
      - /dev/ttyAMA0

polling:
  interval: 30s
  stagger_delay: 500ms
  retry_count: 3
  retry_delay: 1s

devices:
  - address: 1
    name: "Water Meter - Main Building"
    enabled: true
  - address: 2
    name: "Gas Meter - Main Building"
    enabled: true
  - address: 5
    name: "Heat Meter - Warehouse"
    enabled: true
    interval: 60s  # Less frequent polling

self_healing:
  enabled: true
  max_recovery_attempts: 5
  recovery_delay: 5s
  auto_baud_detection: true
  auto_port_scan: true
```

## Error Messages and Solutions

The library provides actionable error messages:

```
ERROR: Serial port /dev/ttyUSB0 not found
  Code: 1000 (ErrPortNotFound)
  Severity: Critical

  Possible causes:
    - USB-to-serial adapter disconnected
    - Wrong port path configured
    - Device not recognized by system

  Suggestions:
    1. Check USB cable connection
    2. Run 'ls /dev/ttyUSB*' to list available ports
    3. Check 'dmesg | tail' for device detection
    4. Try fallback ports: /dev/ttyUSB1, /dev/ttyACM0

  Auto-recovery: Scanning for available ports...
```

```
ERROR: Device at address 5 not responding
  Code: 1009 (ErrDeviceNotResponding)
  Severity: Warning

  Possible causes:
    - Device powered off
    - Incorrect address
    - Bus wiring issue
    - Device malfunction

  Suggestions:
    1. Verify device has power
    2. Check M-Bus wiring connections
    3. Verify device address matches configuration
    4. Try pinging device manually

  Context:
    - Last successful read: 2024-01-15 10:30:00
    - Failed attempts: 3
    - Other devices on bus: responding normally
```

## Thread Safety

The manager is fully thread-safe:

- Internal mutex protection for shared state
- Channel-based event delivery
- Safe concurrent access to configuration
- Atomic transport operations

```go
type manager struct {
    mu          sync.RWMutex
    config      Config
    transport   Transport
    devices     map[int]*deviceState

    dataEvents   chan DataEvent
    errorEvents  chan ErrorEvent
    healthEvents chan HealthEvent

    ctx    context.Context
    cancel context.CancelFunc
    wg     sync.WaitGroup
}
```

## Metrics and Observability

```go
type Metrics struct {
    // Polling metrics
    PollsTotal      uint64
    PollsSuccessful uint64
    PollsFailed     uint64

    // Timing metrics
    AvgReadLatency  time.Duration
    MaxReadLatency  time.Duration

    // Device metrics
    DevicesOnline   int
    DevicesOffline  int

    // Transport metrics
    BytesSent       uint64
    BytesReceived   uint64
    Reconnects      uint64

    // Self-healing metrics
    RecoveryAttempts   uint64
    RecoverySuccessful uint64
}
```

## Implementation Phases

### Phase 1: Core Infrastructure
- [ ] Transport abstraction (Serial + TCP)
- [ ] Configuration system
- [ ] Basic polling scheduler
- [ ] Event system (channels)

### Phase 2: Self-Healing
- [ ] Error classification system
- [ ] Recovery strategies
- [ ] Auto-detection (baud rate, ports)
- [ ] Health monitoring

### Phase 3: Advanced Features
- [ ] Metrics and observability
- [ ] Device discovery/scanning
- [ ] Hot configuration reload
- [ ] Graceful degradation

### Phase 4: Production Hardening
- [ ] Comprehensive tests
- [ ] Documentation
- [ ] Examples
- [ ] Performance optimization

## File Structure

```
go-mbus/
├── manager/
│   ├── manager.go          # Main MBusManager implementation
│   ├── config.go           # Configuration types and loading
│   ├── events.go           # Event types and handling
│   ├── errors.go           # Error codes and rich errors
│   ├── scheduler.go        # Polling scheduler
│   ├── healing.go          # Self-healing engine
│   ├── diagnostics.go      # Diagnostics and reporting
│   ├── metrics.go          # Metrics collection
│   └── transport/
│       ├── transport.go    # Transport interface
│       ├── serial.go       # Serial transport
│       └── tcp.go          # TCP transport
├── pkg/mbus/               # Existing protocol implementation
└── examples/
    └── manager/
        └── main.go         # Example usage
```
