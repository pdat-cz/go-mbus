package transport

import (
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

// TCPConfig contains configuration for TCP transport.
type TCPConfig struct {
	// Host address (IP or hostname)
	Host string `yaml:"host" json:"host"`

	// Port number
	Port int `yaml:"port" json:"port"`

	// ConnectTimeout for establishing connection
	ConnectTimeout time.Duration `yaml:"connect_timeout" json:"connect_timeout"`

	// ReadTimeout for read operations
	ReadTimeout time.Duration `yaml:"read_timeout" json:"read_timeout"`

	// WriteTimeout for write operations
	WriteTimeout time.Duration `yaml:"write_timeout" json:"write_timeout"`

	// KeepAlive enables TCP keep-alive
	KeepAlive bool `yaml:"keep_alive" json:"keep_alive"`

	// KeepAlivePeriod sets the keep-alive period
	KeepAlivePeriod time.Duration `yaml:"keep_alive_period" json:"keep_alive_period"`
}

// DefaultTCPConfig returns default TCP configuration.
func DefaultTCPConfig() TCPConfig {
	return TCPConfig{
		Host:            "127.0.0.1",
		Port:            10001,
		ConnectTimeout:  10 * time.Second,
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    5 * time.Second,
		KeepAlive:       true,
		KeepAlivePeriod: 30 * time.Second,
	}
}

// TCPTransport implements Transport for TCP/IP communication.
type TCPTransport struct {
	BaseTransport
	config TCPConfig
	conn   net.Conn
}

// NewTCPTransport creates a new TCP transport.
func NewTCPTransport(config TCPConfig) *TCPTransport {
	// Apply defaults for zero values
	if config.ConnectTimeout == 0 {
		config.ConnectTimeout = 10 * time.Second
	}
	if config.ReadTimeout == 0 {
		config.ReadTimeout = 5 * time.Second
	}
	if config.WriteTimeout == 0 {
		config.WriteTimeout = 5 * time.Second
	}
	if config.KeepAlivePeriod == 0 {
		config.KeepAlivePeriod = 30 * time.Second
	}

	return &TCPTransport{
		config: config,
	}
}

// Address returns the formatted address string.
func (t *TCPTransport) Address() string {
	return fmt.Sprintf("%s:%d", t.config.Host, t.config.Port)
}

// Open establishes the TCP connection.
func (t *TCPTransport) Open() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.state == StateOpen && t.conn != nil {
		return nil
	}

	t.state = StateConnecting

	// Create dialer with timeout
	dialer := &net.Dialer{
		Timeout:   t.config.ConnectTimeout,
		KeepAlive: t.config.KeepAlivePeriod,
	}

	conn, err := dialer.Dial("tcp", t.Address())
	if err != nil {
		t.state = StateError
		t.SetLastError(err)
		return t.wrapConnError(err)
	}

	// Configure keep-alive if supported
	if tcpConn, ok := conn.(*net.TCPConn); ok && t.config.KeepAlive {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(t.config.KeepAlivePeriod)
	}

	t.conn = conn
	t.state = StateOpen
	t.MarkOpened()

	return nil
}

// Close terminates the TCP connection.
func (t *TCPTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.conn != nil {
		err := t.conn.Close()
		t.conn = nil
		t.state = StateClosed
		return err
	}

	t.state = StateClosed
	return nil
}

// IsOpen returns true if the TCP connection is established.
func (t *TCPTransport) IsOpen() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.state == StateOpen && t.conn != nil
}

// Send transmits a command and receives response.
func (t *TCPTransport) Send(command []byte, timeout time.Duration) ([]byte, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.conn == nil || t.state != StateOpen {
		return nil, errors.New("TCP connection not open")
	}

	// Set write deadline
	writeTimeout := t.config.WriteTimeout
	if timeout > 0 {
		writeTimeout = timeout
	}
	if err := t.conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
		return nil, fmt.Errorf("set write deadline: %w", err)
	}

	// Write command
	n, err := t.conn.Write(command)
	if err != nil {
		t.IncrementErrors()
		t.SetLastError(err)
		return nil, fmt.Errorf("write error: %w", t.wrapConnError(err))
	}
	t.AddBytesSent(uint64(n))

	// Set read deadline
	readTimeout := t.config.ReadTimeout
	if timeout > 0 {
		readTimeout = timeout
	}
	if err := t.conn.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
		return nil, fmt.Errorf("set read deadline: %w", err)
	}

	// Read response
	// M-Bus frames have a maximum size of 256 bytes
	buf := make([]byte, 256)
	var response []byte

	for {
		n, err := t.conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// Timeout is expected after receiving all data
				break
			}
			t.IncrementErrors()
			t.SetLastError(err)
			return nil, fmt.Errorf("read error: %w", t.wrapConnError(err))
		}
		response = append(response, buf[:n]...)

		// Check if we have a complete M-Bus frame
		if isCompleteMBusFrame(response) {
			break
		}
	}

	t.AddBytesReceived(uint64(len(response)))
	return response, nil
}

// GetInfo returns transport information.
func (t *TCPTransport) GetInfo() Info {
	t.mu.Lock()
	defer t.mu.Unlock()

	return Info{
		Type:      "tcp",
		Address:   t.Address(),
		State:     t.state,
		Stats:     t.stats,
		LastError: t.lastError,
	}
}

// GetConfig returns the current TCP configuration.
func (t *TCPTransport) GetConfig() TCPConfig {
	return t.config
}

// SetConfig updates the TCP configuration.
// The transport must be closed before calling this.
func (t *TCPTransport) SetConfig(config TCPConfig) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.state == StateOpen {
		return errors.New("cannot change config while connected")
	}

	t.config = config
	return nil
}

// Reconnect closes and reopens the connection.
func (t *TCPTransport) Reconnect() error {
	if err := t.Close(); err != nil {
		return err
	}
	t.IncrementReconnects()
	return t.Open()
}

// wrapConnError wraps TCP connection errors with more context.
func (t *TCPTransport) wrapConnError(err error) error {
	if err == nil {
		return nil
	}

	if netErr, ok := err.(net.Error); ok {
		if netErr.Timeout() {
			return &ConnError{
				Code:    ErrCodeConnTimeout,
				Address: t.Address(),
				Cause:   err,
				Message: fmt.Sprintf("connection timeout to %s", t.Address()),
			}
		}
	}

	if opErr, ok := err.(*net.OpError); ok {
		if opErr.Op == "dial" {
			return &ConnError{
				Code:    ErrCodeConnRefused,
				Address: t.Address(),
				Cause:   err,
				Message: fmt.Sprintf("connection refused to %s", t.Address()),
			}
		}
	}

	return &ConnError{
		Code:    ErrCodeConnUnknown,
		Address: t.Address(),
		Cause:   err,
		Message: fmt.Sprintf("connection error: %v", err),
	}
}

// ConnErrorCode represents TCP connection error codes.
type ConnErrorCode int

const (
	ErrCodeConnUnknown ConnErrorCode = iota
	ErrCodeConnRefused
	ErrCodeConnTimeout
	ErrCodeConnLost
	ErrCodeConnReset
)

// ConnError represents a TCP connection error with context.
type ConnError struct {
	Code    ConnErrorCode
	Address string
	Cause   error
	Message string
}

func (e *ConnError) Error() string {
	return e.Message
}

func (e *ConnError) Unwrap() error {
	return e.Cause
}

// Suggestion returns a human-readable fix suggestion.
func (e *ConnError) Suggestion() string {
	switch e.Code {
	case ErrCodeConnRefused:
		return fmt.Sprintf("M-Bus gateway at %s is not accepting connections. "+
			"Check if the gateway is powered on and configured correctly.", e.Address)
	case ErrCodeConnTimeout:
		return fmt.Sprintf("Connection to %s timed out. "+
			"Check network connectivity and firewall settings.", e.Address)
	case ErrCodeConnLost:
		return "Connection to the M-Bus gateway was lost. " +
			"Check network stability and gateway status."
	case ErrCodeConnReset:
		return "Connection was reset by the M-Bus gateway. " +
			"The gateway may be overloaded or restarting."
	default:
		return "Check network connectivity and gateway configuration."
	}
}

// isCompleteMBusFrame checks if the buffer contains a complete M-Bus frame.
func isCompleteMBusFrame(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	switch data[0] {
	case 0xE5:
		// Single character ACK
		return true

	case 0x10:
		// Short frame: start(1) + C(1) + A(1) + checksum(1) + stop(1) = 5 bytes
		return len(data) >= 5 && data[len(data)-1] == 0x16

	case 0x68:
		// Long/Control frame: start(1) + L(1) + L(1) + start(1) + ... + stop(1)
		if len(data) < 4 {
			return false
		}
		frameLen := int(data[1])
		totalLen := 4 + frameLen + 2 // header(4) + data(L) + checksum(1) + stop(1)
		return len(data) >= totalLen && data[len(data)-1] == 0x16
	}

	return false
}
