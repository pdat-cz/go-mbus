package manager

import (
	"time"

	"github.com/pdat-cz/go-mbus/pkg/mbus"
)

// DataEvent represents a successful meter reading.
type DataEvent struct {
	// Timestamp when the data was read
	Timestamp time.Time `json:"timestamp"`

	// DeviceAddr is the M-Bus address
	DeviceAddr int `json:"device_addr"`

	// DeviceName is the human-readable device name
	DeviceName string `json:"device_name"`

	// Data contains the parsed M-Bus telegram
	Data mbus.LFrameParsed `json:"data"`

	// ReadLatency is the time taken to read the device
	ReadLatency time.Duration `json:"read_latency"`
}

// ErrorEvent represents a communication error.
type ErrorEvent struct {
	// Timestamp when the error occurred
	Timestamp time.Time `json:"timestamp"`

	// DeviceAddr is the M-Bus address (0 if transport-level error)
	DeviceAddr int `json:"device_addr"`

	// DeviceName is the human-readable device name
	DeviceName string `json:"device_name"`

	// Error is the underlying error
	Error error `json:"error"`

	// Code is the error classification code
	Code ErrorCode `json:"code"`

	// Severity indicates the error severity
	Severity Severity `json:"severity"`

	// Suggestion is a human-readable fix suggestion
	Suggestion string `json:"suggestion"`

	// Context provides additional debugging information
	Context ErrorContext `json:"context"`

	// Recoverable indicates if the error can be automatically recovered
	Recoverable bool `json:"recoverable"`
}

// ErrorContext provides debugging information for errors.
type ErrorContext struct {
	// Transport type ("serial" or "tcp")
	Transport string `json:"transport"`

	// TransportAddress (port path or host:port)
	TransportAddress string `json:"transport_address"`

	// LastSuccessful is the time of last successful operation
	LastSuccessful time.Time `json:"last_successful"`

	// FailureCount is the number of consecutive failures
	FailureCount int `json:"failure_count"`

	// RawBytes contains any received data (for debugging)
	RawBytes []byte `json:"raw_bytes,omitempty"`

	// AttemptedFixes lists recovery actions tried
	AttemptedFixes []string `json:"attempted_fixes,omitempty"`
}

// HealthEvent represents system health changes.
type HealthEvent struct {
	// Timestamp when the event occurred
	Timestamp time.Time `json:"timestamp"`

	// Type of health event
	Type HealthEventType `json:"type"`

	// Component that generated the event
	Component string `json:"component"`

	// Message describes the event
	Message string `json:"message"`

	// Details contains additional information
	Details map[string]interface{} `json:"details,omitempty"`
}

// HealthEventType represents types of health events.
type HealthEventType string

const (
	HealthEventStarted          HealthEventType = "started"
	HealthEventStopped          HealthEventType = "stopped"
	HealthEventConnected        HealthEventType = "connected"
	HealthEventDisconnected     HealthEventType = "disconnected"
	HealthEventRecoveryStarted  HealthEventType = "recovery_started"
	HealthEventRecoverySuccess  HealthEventType = "recovery_success"
	HealthEventRecoveryFailed   HealthEventType = "recovery_failed"
	HealthEventDeviceOnline     HealthEventType = "device_online"
	HealthEventDeviceOffline    HealthEventType = "device_offline"
	HealthEventConfigReloaded   HealthEventType = "config_reloaded"
	HealthEventWarning          HealthEventType = "warning"
)

// Severity represents error severity levels.
type Severity int

const (
	SeverityDebug Severity = iota
	SeverityInfo
	SeverityWarning
	SeverityError
	SeverityCritical
)

func (s Severity) String() string {
	switch s {
	case SeverityDebug:
		return "debug"
	case SeverityInfo:
		return "info"
	case SeverityWarning:
		return "warning"
	case SeverityError:
		return "error"
	case SeverityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// ErrorCode represents error classification codes.
type ErrorCode int

const (
	// Transport errors (1000-1099)
	ErrCodeUnknown           ErrorCode = 1000
	ErrCodePortNotFound      ErrorCode = 1001
	ErrCodePortBusy          ErrorCode = 1002
	ErrCodePortAccessDenied  ErrorCode = 1003
	ErrCodePortDisconnected  ErrorCode = 1004
	ErrCodeBaudRateMismatch  ErrorCode = 1005
	ErrCodeParityError       ErrorCode = 1006
	ErrCodeFrameError        ErrorCode = 1007
	ErrCodeTCPConnFailed     ErrorCode = 1010
	ErrCodeTCPConnLost       ErrorCode = 1011
	ErrCodeTCPTimeout        ErrorCode = 1012

	// Protocol errors (1100-1199)
	ErrCodeChecksumMismatch  ErrorCode = 1100
	ErrCodeInvalidFrame      ErrorCode = 1101
	ErrCodeInvalidResponse   ErrorCode = 1102
	ErrCodeTimeout           ErrorCode = 1103

	// Device errors (1200-1299)
	ErrCodeDeviceNotResponding ErrorCode = 1200
	ErrCodeDeviceError         ErrorCode = 1201
	ErrCodeDeviceBusy          ErrorCode = 1202

	// Configuration errors (1300-1399)
	ErrCodeInvalidConfig       ErrorCode = 1300
	ErrCodeInvalidAddress      ErrorCode = 1301
)

func (c ErrorCode) String() string {
	switch c {
	case ErrCodePortNotFound:
		return "PORT_NOT_FOUND"
	case ErrCodePortBusy:
		return "PORT_BUSY"
	case ErrCodePortAccessDenied:
		return "PORT_ACCESS_DENIED"
	case ErrCodePortDisconnected:
		return "PORT_DISCONNECTED"
	case ErrCodeBaudRateMismatch:
		return "BAUD_RATE_MISMATCH"
	case ErrCodeParityError:
		return "PARITY_ERROR"
	case ErrCodeFrameError:
		return "FRAME_ERROR"
	case ErrCodeTCPConnFailed:
		return "TCP_CONN_FAILED"
	case ErrCodeTCPConnLost:
		return "TCP_CONN_LOST"
	case ErrCodeTCPTimeout:
		return "TCP_TIMEOUT"
	case ErrCodeChecksumMismatch:
		return "CHECKSUM_MISMATCH"
	case ErrCodeInvalidFrame:
		return "INVALID_FRAME"
	case ErrCodeInvalidResponse:
		return "INVALID_RESPONSE"
	case ErrCodeTimeout:
		return "TIMEOUT"
	case ErrCodeDeviceNotResponding:
		return "DEVICE_NOT_RESPONDING"
	case ErrCodeDeviceError:
		return "DEVICE_ERROR"
	case ErrCodeDeviceBusy:
		return "DEVICE_BUSY"
	case ErrCodeInvalidConfig:
		return "INVALID_CONFIG"
	case ErrCodeInvalidAddress:
		return "INVALID_ADDRESS"
	default:
		return "UNKNOWN"
	}
}

// Suggestion returns a human-readable fix suggestion for the error code.
func (c ErrorCode) Suggestion() string {
	switch c {
	case ErrCodePortNotFound:
		return "Check if the USB-to-serial adapter is connected. Run 'ls /dev/tty*' to list available ports."
	case ErrCodePortBusy:
		return "The port is in use by another application. Close other M-Bus software or wait for it to release the port."
	case ErrCodePortAccessDenied:
		return "Add your user to the 'dialout' group: sudo usermod -a -G dialout $USER, then log out and back in."
	case ErrCodePortDisconnected:
		return "The serial adapter was disconnected. Check the USB cable and reconnect."
	case ErrCodeBaudRateMismatch:
		return "The baud rate may be incorrect. Standard M-Bus uses 2400 baud. Try enabling auto-detection."
	case ErrCodeParityError:
		return "Parity setting mismatch. M-Bus standard uses even parity."
	case ErrCodeFrameError:
		return "Received malformed data. Check wiring and electrical connections."
	case ErrCodeTCPConnFailed:
		return "Cannot connect to M-Bus gateway. Verify the host address and port, and check if the gateway is online."
	case ErrCodeTCPConnLost:
		return "Lost connection to M-Bus gateway. Check network stability."
	case ErrCodeTCPTimeout:
		return "Connection timed out. Check network connectivity and gateway responsiveness."
	case ErrCodeChecksumMismatch:
		return "Data corruption detected. Check M-Bus wiring for interference."
	case ErrCodeInvalidFrame:
		return "Received invalid M-Bus frame. The device may be malfunctioning."
	case ErrCodeTimeout:
		return "No response from device. Verify the device is powered and the address is correct."
	case ErrCodeDeviceNotResponding:
		return "Device not responding. Check power supply and M-Bus wiring to the device."
	case ErrCodeDeviceError:
		return "Device reported an error. Check device status and documentation."
	case ErrCodeDeviceBusy:
		return "Device is busy. Wait and retry."
	default:
		return "An unexpected error occurred. Check the logs for details."
	}
}

// DeviceStatus represents the current status of a device.
type DeviceStatus struct {
	// Address is the M-Bus address
	Address int `json:"address"`

	// Name is the human-readable device name
	Name string `json:"name"`

	// Online indicates if the device is responding
	Online bool `json:"online"`

	// LastSeen is the time of last successful read
	LastSeen time.Time `json:"last_seen"`

	// LastError is the most recent error (if any)
	LastError string `json:"last_error,omitempty"`

	// ConsecutiveFailures is the number of failures since last success
	ConsecutiveFailures int `json:"consecutive_failures"`

	// TotalReads is the total number of successful reads
	TotalReads uint64 `json:"total_reads"`

	// TotalErrors is the total number of errors
	TotalErrors uint64 `json:"total_errors"`
}

// HealthStatus represents the overall manager health.
type HealthStatus struct {
	// State is the current manager state
	State ManagerState `json:"state"`

	// StartedAt is when the manager was started
	StartedAt time.Time `json:"started_at"`

	// Uptime is the duration since start
	Uptime time.Duration `json:"uptime"`

	// Transport contains transport information
	Transport TransportHealth `json:"transport"`

	// Devices contains device statuses
	Devices []DeviceStatus `json:"devices"`

	// Stats contains aggregate statistics
	Stats ManagerStats `json:"stats"`
}

// TransportHealth represents transport health information.
type TransportHealth struct {
	// Type is the transport type
	Type string `json:"type"`

	// Address is the transport address
	Address string `json:"address"`

	// Connected indicates if the transport is connected
	Connected bool `json:"connected"`

	// LastConnected is when the transport last connected
	LastConnected time.Time `json:"last_connected"`

	// Reconnects is the number of reconnection attempts
	Reconnects uint64 `json:"reconnects"`
}

// ManagerStats contains aggregate statistics.
type ManagerStats struct {
	// TotalPolls is the total number of polling cycles
	TotalPolls uint64 `json:"total_polls"`

	// TotalReads is the total number of device reads
	TotalReads uint64 `json:"total_reads"`

	// SuccessfulReads is the number of successful reads
	SuccessfulReads uint64 `json:"successful_reads"`

	// FailedReads is the number of failed reads
	FailedReads uint64 `json:"failed_reads"`

	// RecoveryAttempts is the number of recovery attempts
	RecoveryAttempts uint64 `json:"recovery_attempts"`

	// RecoverySuccesses is the number of successful recoveries
	RecoverySuccesses uint64 `json:"recovery_successes"`

	// BytesSent is the total bytes sent
	BytesSent uint64 `json:"bytes_sent"`

	// BytesReceived is the total bytes received
	BytesReceived uint64 `json:"bytes_received"`
}

// ManagerState represents the manager's current state.
type ManagerState string

const (
	StateIdle       ManagerState = "idle"
	StateStarting   ManagerState = "starting"
	StateRunning    ManagerState = "running"
	StateStopping   ManagerState = "stopping"
	StateStopped    ManagerState = "stopped"
	StateRecovering ManagerState = "recovering"
	StateError      ManagerState = "error"
)
