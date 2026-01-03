package transport

import (
	"testing"
	"time"
)

func TestStateString(t *testing.T) {
	tests := []struct {
		state    State
		expected string
	}{
		{StateClosed, "closed"},
		{StateConnecting, "connecting"},
		{StateOpen, "open"},
		{StateError, "error"},
		{State(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.state.String(); got != tt.expected {
				t.Errorf("State.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBaseTransport(t *testing.T) {
	bt := &BaseTransport{}

	// Test initial state
	if bt.GetState() != StateClosed {
		t.Errorf("Initial state should be StateClosed, got %v", bt.GetState())
	}

	// Test state changes
	bt.SetState(StateOpen)
	if bt.GetState() != StateOpen {
		t.Errorf("State should be StateOpen, got %v", bt.GetState())
	}

	// Test stats
	bt.AddBytesSent(100)
	bt.AddBytesReceived(200)
	bt.IncrementErrors()
	bt.IncrementReconnects()

	stats := bt.GetStats()
	if stats.BytesSent != 100 {
		t.Errorf("BytesSent = %v, want 100", stats.BytesSent)
	}
	if stats.BytesReceived != 200 {
		t.Errorf("BytesReceived = %v, want 200", stats.BytesReceived)
	}
	if stats.Errors != 1 {
		t.Errorf("Errors = %v, want 1", stats.Errors)
	}
	if stats.Reconnects != 1 {
		t.Errorf("Reconnects = %v, want 1", stats.Reconnects)
	}

	// Test MarkOpened
	bt.MarkOpened()
	stats = bt.GetStats()
	if stats.OpenedAt.IsZero() {
		t.Error("OpenedAt should be set after MarkOpened")
	}

	// Test last error
	testErr := &PortError{Code: ErrCodePortNotFound, Message: "test error"}
	bt.SetLastError(testErr)
	if bt.GetLastError() != testErr {
		t.Errorf("GetLastError() = %v, want %v", bt.GetLastError(), testErr)
	}

	// Test reset stats
	bt.ResetStats()
	stats = bt.GetStats()
	if stats.BytesSent != 0 || stats.BytesReceived != 0 {
		t.Error("Stats should be reset")
	}
}

func TestDefaultSerialConfig(t *testing.T) {
	config := DefaultSerialConfig()

	if config.BaudRate != 2400 {
		t.Errorf("Default BaudRate = %v, want 2400", config.BaudRate)
	}
	if config.DataBits != 8 {
		t.Errorf("Default DataBits = %v, want 8", config.DataBits)
	}
	if config.StopBits != 1 {
		t.Errorf("Default StopBits = %v, want 1", config.StopBits)
	}
	if config.Parity != "even" {
		t.Errorf("Default Parity = %v, want 'even'", config.Parity)
	}
	if config.ReadTimeout != 2*time.Second {
		t.Errorf("Default ReadTimeout = %v, want 2s", config.ReadTimeout)
	}
}

func TestNewSerialTransport(t *testing.T) {
	config := SerialConfig{
		Port: "/dev/ttyUSB0",
	}

	st := NewSerialTransport(config)

	// Check defaults are applied
	if st.config.BaudRate != 2400 {
		t.Errorf("BaudRate = %v, want 2400", st.config.BaudRate)
	}
	if st.config.Parity != "even" {
		t.Errorf("Parity = %v, want 'even'", st.config.Parity)
	}

	// Check initial state
	if st.IsOpen() {
		t.Error("New transport should not be open")
	}

	info := st.GetInfo()
	if info.Type != "serial" {
		t.Errorf("Type = %v, want 'serial'", info.Type)
	}
	if info.Address != "/dev/ttyUSB0" {
		t.Errorf("Address = %v, want '/dev/ttyUSB0'", info.Address)
	}
}

func TestSerialTransportSetConfig(t *testing.T) {
	config := SerialConfig{Port: "/dev/ttyUSB0"}
	st := NewSerialTransport(config)

	// Should be able to set config when closed
	newConfig := SerialConfig{Port: "/dev/ttyUSB1", BaudRate: 9600}
	if err := st.SetConfig(newConfig); err != nil {
		t.Errorf("SetConfig failed: %v", err)
	}

	if st.config.Port != "/dev/ttyUSB1" {
		t.Errorf("Port = %v, want '/dev/ttyUSB1'", st.config.Port)
	}
}

func TestDefaultTCPConfig(t *testing.T) {
	config := DefaultTCPConfig()

	if config.Host != "127.0.0.1" {
		t.Errorf("Default Host = %v, want '127.0.0.1'", config.Host)
	}
	if config.Port != 10001 {
		t.Errorf("Default Port = %v, want 10001", config.Port)
	}
	if config.ConnectTimeout != 10*time.Second {
		t.Errorf("Default ConnectTimeout = %v, want 10s", config.ConnectTimeout)
	}
	if config.KeepAlive != true {
		t.Errorf("Default KeepAlive = %v, want true", config.KeepAlive)
	}
}

func TestNewTCPTransport(t *testing.T) {
	config := TCPConfig{
		Host: "192.168.1.100",
		Port: 5000,
	}

	tt := NewTCPTransport(config)

	// Check defaults are applied
	if tt.config.ConnectTimeout != 10*time.Second {
		t.Errorf("ConnectTimeout = %v, want 10s", tt.config.ConnectTimeout)
	}

	// Check address
	if tt.Address() != "192.168.1.100:5000" {
		t.Errorf("Address() = %v, want '192.168.1.100:5000'", tt.Address())
	}

	// Check initial state
	if tt.IsOpen() {
		t.Error("New transport should not be open")
	}

	info := tt.GetInfo()
	if info.Type != "tcp" {
		t.Errorf("Type = %v, want 'tcp'", info.Type)
	}
}

func TestTCPTransportSetConfig(t *testing.T) {
	config := TCPConfig{Host: "192.168.1.100", Port: 5000}
	tt := NewTCPTransport(config)

	// Should be able to set config when closed
	newConfig := TCPConfig{Host: "192.168.1.200", Port: 6000}
	if err := tt.SetConfig(newConfig); err != nil {
		t.Errorf("SetConfig failed: %v", err)
	}

	if tt.config.Host != "192.168.1.200" {
		t.Errorf("Host = %v, want '192.168.1.200'", tt.config.Host)
	}
}

func TestPortError(t *testing.T) {
	err := &PortError{
		Code:    ErrCodePortNotFound,
		Port:    "/dev/ttyUSB0",
		Message: "port not found",
	}

	if err.Error() != "port not found" {
		t.Errorf("Error() = %v, want 'port not found'", err.Error())
	}

	suggestion := err.Suggestion()
	if suggestion == "" {
		t.Error("Suggestion should not be empty")
	}
}

func TestPortErrorSuggestions(t *testing.T) {
	tests := []struct {
		code     PortErrorCode
		contains string
	}{
		{ErrCodePortNotFound, "ls /dev/tty"},
		{ErrCodePortBusy, "in use"},
		{ErrCodePortAccessDenied, "dialout"},
		{ErrCodePortDisconnected, "disconnected"},
		{ErrCodePortUnknown, "configuration"},
	}

	for _, tt := range tests {
		t.Run(tt.code.String(), func(t *testing.T) {
			err := &PortError{Code: tt.code, Port: "/dev/ttyUSB0"}
			suggestion := err.Suggestion()
			if !containsString(suggestion, tt.contains) {
				t.Errorf("Suggestion for %v should contain '%v', got '%v'",
					tt.code, tt.contains, suggestion)
			}
		})
	}
}

func TestConnError(t *testing.T) {
	err := &ConnError{
		Code:    ErrCodeConnRefused,
		Address: "192.168.1.100:5000",
		Message: "connection refused",
	}

	if err.Error() != "connection refused" {
		t.Errorf("Error() = %v, want 'connection refused'", err.Error())
	}

	suggestion := err.Suggestion()
	if suggestion == "" {
		t.Error("Suggestion should not be empty")
	}
}

func TestConnErrorSuggestions(t *testing.T) {
	tests := []struct {
		code     ConnErrorCode
		contains string
	}{
		{ErrCodeConnRefused, "gateway"},
		{ErrCodeConnTimeout, "timed out"},
		{ErrCodeConnLost, "lost"},
		{ErrCodeConnReset, "reset"},
		{ErrCodeConnUnknown, "connectivity"},
	}

	for _, tt := range tests {
		t.Run(tt.contains, func(t *testing.T) {
			err := &ConnError{Code: tt.code, Address: "192.168.1.100:5000"}
			suggestion := err.Suggestion()
			if !containsString(suggestion, tt.contains) {
				t.Errorf("Suggestion for code %v should contain '%v', got '%v'",
					tt.code, tt.contains, suggestion)
			}
		})
	}
}

func TestIsCompleteMBusFrame(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected bool
	}{
		{"empty", []byte{}, false},
		{"ACK", []byte{0xE5}, true},
		{"short frame incomplete", []byte{0x10, 0x5B, 0x01}, false},
		{"short frame complete", []byte{0x10, 0x5B, 0x01, 0x5C, 0x16}, true},
		{"long frame incomplete", []byte{0x68, 0x03, 0x03, 0x68}, false},
		{"long frame complete", []byte{0x68, 0x03, 0x03, 0x68, 0x08, 0x01, 0x72, 0x7B, 0x16}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isCompleteMBusFrame(tt.data); got != tt.expected {
				t.Errorf("isCompleteMBusFrame(%v) = %v, want %v", tt.data, got, tt.expected)
			}
		})
	}
}

func (c PortErrorCode) String() string {
	switch c {
	case ErrCodePortNotFound:
		return "PORT_NOT_FOUND"
	case ErrCodePortBusy:
		return "PORT_BUSY"
	case ErrCodePortAccessDenied:
		return "PORT_ACCESS_DENIED"
	case ErrCodePortDisconnected:
		return "PORT_DISCONNECTED"
	default:
		return "UNKNOWN"
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
