// Package transport provides abstraction for M-Bus communication transports.
package transport

import (
	"sync"
	"time"
)

// State represents the current state of a transport connection.
type State int

const (
	// StateClosed indicates the transport is not connected.
	StateClosed State = iota
	// StateConnecting indicates the transport is attempting to connect.
	StateConnecting
	// StateOpen indicates the transport is connected and ready.
	StateOpen
	// StateError indicates the transport encountered an error.
	StateError
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateConnecting:
		return "connecting"
	case StateOpen:
		return "open"
	case StateError:
		return "error"
	default:
		return "unknown"
	}
}

// Stats contains transport statistics.
type Stats struct {
	BytesSent     uint64
	BytesReceived uint64
	Errors        uint64
	Reconnects    uint64
	OpenedAt      time.Time
}

// Info contains transport information.
type Info struct {
	Type      string // "serial" or "tcp"
	Address   string // "/dev/ttyUSB0" or "192.168.1.1:10001"
	State     State
	Stats     Stats
	LastError error
}

// Transport defines the interface for M-Bus communication transports.
type Transport interface {
	// Open establishes the connection.
	Open() error

	// Close terminates the connection.
	Close() error

	// IsOpen returns true if the transport is connected.
	IsOpen() bool

	// Send transmits a command and receives response.
	// It handles the write and read operations atomically.
	Send(command []byte, timeout time.Duration) ([]byte, error)

	// GetInfo returns transport details and statistics.
	GetInfo() Info

	// Lock acquires exclusive access to the transport.
	// This must be called before Send() in concurrent scenarios.
	Lock()

	// Unlock releases exclusive access to the transport.
	Unlock()
}

// BaseTransport provides common functionality for transports.
type BaseTransport struct {
	mu        sync.Mutex
	state     State
	stats     Stats
	lastError error
}

// Lock acquires the transport mutex.
func (t *BaseTransport) Lock() {
	t.mu.Lock()
}

// Unlock releases the transport mutex.
func (t *BaseTransport) Unlock() {
	t.mu.Unlock()
}

// GetState returns the current transport state.
func (t *BaseTransport) GetState() State {
	return t.state
}

// SetState sets the transport state.
func (t *BaseTransport) SetState(state State) {
	t.state = state
}

// GetStats returns a copy of the transport statistics.
func (t *BaseTransport) GetStats() Stats {
	return t.stats
}

// AddBytesSent increments the bytes sent counter.
func (t *BaseTransport) AddBytesSent(n uint64) {
	t.stats.BytesSent += n
}

// AddBytesReceived increments the bytes received counter.
func (t *BaseTransport) AddBytesReceived(n uint64) {
	t.stats.BytesReceived += n
}

// IncrementErrors increments the error counter.
func (t *BaseTransport) IncrementErrors() {
	t.stats.Errors++
}

// IncrementReconnects increments the reconnect counter.
func (t *BaseTransport) IncrementReconnects() {
	t.stats.Reconnects++
}

// SetLastError sets the last error encountered.
func (t *BaseTransport) SetLastError(err error) {
	t.lastError = err
}

// GetLastError returns the last error encountered.
func (t *BaseTransport) GetLastError() error {
	return t.lastError
}

// MarkOpened sets the opened timestamp.
func (t *BaseTransport) MarkOpened() {
	t.stats.OpenedAt = time.Now()
}

// ResetStats resets the transport statistics.
func (t *BaseTransport) ResetStats() {
	t.stats = Stats{}
}
