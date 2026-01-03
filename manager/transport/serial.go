package transport

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/tarm/serial"
)

// SerialConfig contains configuration for serial transport.
type SerialConfig struct {
	// Port path: "/dev/ttyUSB0", "COM1", etc.
	Port string `yaml:"port" json:"port"`

	// BaudRate (default: 2400 for M-Bus)
	BaudRate int `yaml:"baud_rate" json:"baud_rate"`

	// DataBits (default: 8)
	DataBits int `yaml:"data_bits" json:"data_bits"`

	// StopBits (default: 1)
	StopBits int `yaml:"stop_bits" json:"stop_bits"`

	// Parity: "none", "even", "odd" (default: "even" for M-Bus)
	Parity string `yaml:"parity" json:"parity"`

	// ReadTimeout for read operations
	ReadTimeout time.Duration `yaml:"read_timeout" json:"read_timeout"`

	// FallbackPorts to try if primary port fails
	FallbackPorts []string `yaml:"fallback_ports" json:"fallback_ports"`

	// AutoDetect enables automatic baud rate detection
	AutoDetect bool `yaml:"auto_detect" json:"auto_detect"`
}

// DefaultSerialConfig returns default M-Bus serial configuration.
func DefaultSerialConfig() SerialConfig {
	return SerialConfig{
		Port:        "/dev/ttyUSB0",
		BaudRate:    2400,
		DataBits:    8,
		StopBits:    1,
		Parity:      "even",
		ReadTimeout: 2 * time.Second,
	}
}

// SerialTransport implements Transport for serial communication.
type SerialTransport struct {
	BaseTransport
	config     SerialConfig
	port       *serial.Port
	activePort string // Currently connected port
}

// NewSerialTransport creates a new serial transport.
func NewSerialTransport(config SerialConfig) *SerialTransport {
	// Apply defaults for zero values
	if config.BaudRate == 0 {
		config.BaudRate = 2400
	}
	if config.DataBits == 0 {
		config.DataBits = 8
	}
	if config.StopBits == 0 {
		config.StopBits = 1
	}
	if config.Parity == "" {
		config.Parity = "even"
	}
	if config.ReadTimeout == 0 {
		config.ReadTimeout = 2 * time.Second
	}

	return &SerialTransport{
		config: config,
	}
}

// Open establishes the serial connection.
func (t *SerialTransport) Open() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.state == StateOpen {
		return nil
	}

	t.state = StateConnecting

	// Try primary port first
	portsToTry := []string{t.config.Port}
	portsToTry = append(portsToTry, t.config.FallbackPorts...)

	var lastErr error
	for _, portPath := range portsToTry {
		if portPath == "" {
			continue
		}

		err := t.openPort(portPath)
		if err == nil {
			t.activePort = portPath
			t.state = StateOpen
			t.MarkOpened()
			return nil
		}
		lastErr = err
	}

	t.state = StateError
	t.SetLastError(lastErr)
	return fmt.Errorf("failed to open serial port: %w", lastErr)
}

func (t *SerialTransport) openPort(portPath string) error {
	parity := serial.ParityEven
	switch strings.ToLower(t.config.Parity) {
	case "none":
		parity = serial.ParityNone
	case "odd":
		parity = serial.ParityOdd
	case "even":
		parity = serial.ParityEven
	}

	stopBits := serial.Stop1
	if t.config.StopBits == 2 {
		stopBits = serial.Stop2
	}

	cfg := &serial.Config{
		Name:        portPath,
		Baud:        t.config.BaudRate,
		Size:        byte(t.config.DataBits),
		Parity:      parity,
		StopBits:    stopBits,
		ReadTimeout: t.config.ReadTimeout,
	}

	port, err := serial.OpenPort(cfg)
	if err != nil {
		return t.wrapPortError(err, portPath)
	}

	t.port = port
	return nil
}

// Close terminates the serial connection.
func (t *SerialTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.port != nil {
		err := t.port.Close()
		t.port = nil
		t.state = StateClosed
		t.activePort = ""
		return err
	}

	t.state = StateClosed
	return nil
}

// IsOpen returns true if the serial port is connected.
func (t *SerialTransport) IsOpen() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.state == StateOpen && t.port != nil
}

// Send transmits a command and receives response.
func (t *SerialTransport) Send(command []byte, timeout time.Duration) ([]byte, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.port == nil || t.state != StateOpen {
		return nil, errors.New("serial port not open")
	}

	// Write command
	n, err := t.port.Write(command)
	if err != nil {
		t.IncrementErrors()
		t.SetLastError(err)
		return nil, fmt.Errorf("write error: %w", err)
	}
	t.AddBytesSent(uint64(n))

	// Read response
	buf, err := io.ReadAll(t.port)
	if err != nil {
		t.IncrementErrors()
		t.SetLastError(err)
		return nil, fmt.Errorf("read error: %w", err)
	}
	t.AddBytesReceived(uint64(len(buf)))

	return buf, nil
}

// GetInfo returns transport information.
func (t *SerialTransport) GetInfo() Info {
	t.mu.Lock()
	defer t.mu.Unlock()

	addr := t.activePort
	if addr == "" {
		addr = t.config.Port
	}

	return Info{
		Type:      "serial",
		Address:   addr,
		State:     t.state,
		Stats:     t.stats,
		LastError: t.lastError,
	}
}

// GetConfig returns the current serial configuration.
func (t *SerialTransport) GetConfig() SerialConfig {
	return t.config
}

// SetConfig updates the serial configuration.
// The transport must be closed before calling this.
func (t *SerialTransport) SetConfig(config SerialConfig) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.state == StateOpen {
		return errors.New("cannot change config while connected")
	}

	t.config = config
	return nil
}

// GetActivePort returns the currently connected port path.
func (t *SerialTransport) GetActivePort() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.activePort
}

// TryBaudRate attempts to communicate at a specific baud rate.
func (t *SerialTransport) TryBaudRate(baudRate int) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.state == StateOpen {
		return errors.New("cannot change baud rate while connected")
	}

	t.config.BaudRate = baudRate
	return nil
}

// DetectBaudRate tries common baud rates to find a working one.
func (t *SerialTransport) DetectBaudRate(testFunc func() bool) (int, error) {
	commonBaudRates := []int{2400, 9600, 19200, 4800, 1200, 300}

	originalBaud := t.config.BaudRate

	for _, baud := range commonBaudRates {
		t.config.BaudRate = baud

		if err := t.Open(); err != nil {
			continue
		}

		if testFunc() {
			return baud, nil
		}

		t.Close()
	}

	t.config.BaudRate = originalBaud
	return 0, errors.New("no working baud rate found")
}

// ListAvailablePorts returns available serial ports on the system.
func ListAvailablePorts() ([]string, error) {
	var ports []string

	switch runtime.GOOS {
	case "linux":
		// Check /dev/ttyUSB*, /dev/ttyACM*, /dev/ttyAMA*, /dev/ttyS*
		patterns := []string{
			"/dev/ttyUSB*",
			"/dev/ttyACM*",
			"/dev/ttyAMA*",
		}
		for _, pattern := range patterns {
			matches, err := filepath.Glob(pattern)
			if err == nil {
				ports = append(ports, matches...)
			}
		}

	case "darwin":
		// Check /dev/tty.* and /dev/cu.*
		patterns := []string{
			"/dev/tty.usb*",
			"/dev/cu.usb*",
		}
		for _, pattern := range patterns {
			matches, err := filepath.Glob(pattern)
			if err == nil {
				ports = append(ports, matches...)
			}
		}

	case "windows":
		// Check COM1 through COM256
		for i := 1; i <= 256; i++ {
			port := fmt.Sprintf("COM%d", i)
			if _, err := os.Stat(fmt.Sprintf("\\\\.\\%s", port)); err == nil {
				ports = append(ports, port)
			}
		}
	}

	return ports, nil
}

// wrapPortError wraps serial port errors with more context.
func (t *SerialTransport) wrapPortError(err error, portPath string) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	// Check for common error patterns
	if strings.Contains(errStr, "no such file") || strings.Contains(errStr, "does not exist") {
		return &PortError{
			Code:    ErrCodePortNotFound,
			Port:    portPath,
			Cause:   err,
			Message: fmt.Sprintf("serial port %s not found", portPath),
		}
	}

	if strings.Contains(errStr, "permission denied") || strings.Contains(errStr, "access denied") {
		return &PortError{
			Code:    ErrCodePortAccessDenied,
			Port:    portPath,
			Cause:   err,
			Message: fmt.Sprintf("permission denied for port %s", portPath),
		}
	}

	if strings.Contains(errStr, "already open") || strings.Contains(errStr, "device or resource busy") {
		return &PortError{
			Code:    ErrCodePortBusy,
			Port:    portPath,
			Cause:   err,
			Message: fmt.Sprintf("serial port %s is busy", portPath),
		}
	}

	return &PortError{
		Code:    ErrCodePortUnknown,
		Port:    portPath,
		Cause:   err,
		Message: fmt.Sprintf("serial port error: %v", err),
	}
}

// PortErrorCode represents serial port error codes.
type PortErrorCode int

const (
	ErrCodePortUnknown PortErrorCode = iota
	ErrCodePortNotFound
	ErrCodePortBusy
	ErrCodePortAccessDenied
	ErrCodePortDisconnected
)

// PortError represents a serial port error with context.
type PortError struct {
	Code    PortErrorCode
	Port    string
	Cause   error
	Message string
}

func (e *PortError) Error() string {
	return e.Message
}

func (e *PortError) Unwrap() error {
	return e.Cause
}

// Suggestion returns a human-readable fix suggestion.
func (e *PortError) Suggestion() string {
	switch e.Code {
	case ErrCodePortNotFound:
		return fmt.Sprintf("Check if the device is connected. Run 'ls /dev/tty*' to list available ports. "+
			"Common ports: /dev/ttyUSB0, /dev/ttyACM0, /dev/ttyAMA0. Port %s was not found.", e.Port)
	case ErrCodePortBusy:
		return fmt.Sprintf("Port %s is in use by another application. "+
			"Check for other M-Bus software or close conflicting applications.", e.Port)
	case ErrCodePortAccessDenied:
		return fmt.Sprintf("Add your user to the 'dialout' group: sudo usermod -a -G dialout $USER, "+
			"then log out and back in. Or run with sudo. Port: %s", e.Port)
	case ErrCodePortDisconnected:
		return "The USB-to-serial adapter was disconnected. Check the cable and reconnect."
	default:
		return "Check the serial port configuration and connections."
	}
}
