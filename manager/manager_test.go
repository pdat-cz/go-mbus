package manager

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/pdat-cz/go-mbus/manager/transport"
)

// mockTransport is a mock implementation of the Transport interface for testing.
type mockTransport struct {
	transport.BaseTransport
	openErr     error
	sendErr     error
	sendResp    []byte
	openCalled  int
	closeCalled int
	sendCalled  int
	mu          sync.Mutex
}

func newMockTransport() *mockTransport {
	return &mockTransport{}
}

func (m *mockTransport) Open() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.openCalled++
	if m.openErr != nil {
		m.SetState(transport.StateError)
		return m.openErr
	}
	m.SetState(transport.StateOpen)
	m.MarkOpened()
	return nil
}

func (m *mockTransport) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closeCalled++
	m.SetState(transport.StateClosed)
	return nil
}

func (m *mockTransport) IsOpen() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.GetState() == transport.StateOpen
}

func (m *mockTransport) Send(command []byte, timeout time.Duration) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sendCalled++
	m.AddBytesSent(uint64(len(command)))
	if m.sendErr != nil {
		m.IncrementErrors()
		return nil, m.sendErr
	}
	if m.sendResp != nil {
		m.AddBytesReceived(uint64(len(m.sendResp)))
		return m.sendResp, nil
	}
	// Return a valid ACK response by default
	return []byte{0xE5}, nil
}

func (m *mockTransport) GetInfo() transport.Info {
	m.mu.Lock()
	defer m.mu.Unlock()
	return transport.Info{
		Type:      "mock",
		Address:   "mock://test",
		State:     m.GetState(),
		Stats:     m.GetStats(),
		LastError: m.GetLastError(),
	}
}

func (m *mockTransport) SetOpenError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.openErr = err
}

func (m *mockTransport) SetSendError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sendErr = err
}

func (m *mockTransport) SetSendResponse(resp []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sendResp = resp
}

func TestNewManager(t *testing.T) {
	config := DefaultConfig()
	config.Transport.Serial.Port = "/dev/ttyUSB0"
	config.Devices = []DeviceConfig{
		{Address: 1, Name: "Meter 1", Enabled: true},
	}

	mgr, err := NewManager(config)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	if mgr == nil {
		t.Fatal("Manager should not be nil")
	}

	devices := mgr.ListDevices()
	if len(devices) != 1 {
		t.Errorf("Device count = %v, want 1", len(devices))
	}
}

func TestNewManagerInvalidConfig(t *testing.T) {
	config := Config{
		Transport: TransportConfig{
			Type: "invalid",
		},
	}

	_, err := NewManager(config)
	if err == nil {
		t.Error("NewManager should fail with invalid config")
	}
}

func TestManagerAddRemoveDevice(t *testing.T) {
	config := DefaultConfig()
	config.Transport.Serial.Port = "/dev/ttyUSB0"

	mgr, err := NewManager(config)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Add device
	err = mgr.AddDevice(DeviceConfig{Address: 1, Name: "Meter 1", Enabled: true})
	if err != nil {
		t.Errorf("AddDevice failed: %v", err)
	}

	devices := mgr.ListDevices()
	if len(devices) != 1 {
		t.Errorf("Device count = %v, want 1", len(devices))
	}

	// Add duplicate device
	err = mgr.AddDevice(DeviceConfig{Address: 1, Name: "Meter 1 Dup", Enabled: true})
	if err == nil {
		t.Error("AddDevice should fail for duplicate address")
	}

	// Remove device
	err = mgr.RemoveDevice(1)
	if err != nil {
		t.Errorf("RemoveDevice failed: %v", err)
	}

	devices = mgr.ListDevices()
	if len(devices) != 0 {
		t.Errorf("Device count after remove = %v, want 0", len(devices))
	}

	// Remove non-existent device
	err = mgr.RemoveDevice(99)
	if err == nil {
		t.Error("RemoveDevice should fail for non-existent device")
	}
}

func TestManagerGetConfig(t *testing.T) {
	config := DefaultConfig()
	config.Transport.Serial.Port = "/dev/ttyUSB0"

	mgr, err := NewManager(config)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	got := mgr.GetConfig()
	if got.Transport.Type != config.Transport.Type {
		t.Errorf("Transport type = %v, want %v", got.Transport.Type, config.Transport.Type)
	}
}

func TestBuildCommands(t *testing.T) {
	// Test REQ_UD2 command
	cmd := buildREQUD2Command(1)
	if len(cmd) != 5 {
		t.Errorf("REQ_UD2 length = %v, want 5", len(cmd))
	}
	if cmd[0] != 0x10 || cmd[4] != 0x16 {
		t.Error("REQ_UD2 frame format incorrect")
	}

	// Test SND_NKE command
	cmd = buildSNDNKECommand(1)
	if len(cmd) != 5 {
		t.Errorf("SND_NKE length = %v, want 5", len(cmd))
	}
	if cmd[0] != 0x10 || cmd[4] != 0x16 {
		t.Error("SND_NKE frame format incorrect")
	}
}

func TestClassifyError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode ErrorCode
	}{
		{"nil error", nil, ErrCodeUnknown},
		{"timeout error", errors.New("operation timeout"), ErrCodeTimeout},
		{"no response", errors.New("no response from device"), ErrCodeDeviceNotResponding},
		{"checksum error", errors.New("checksum mismatch"), ErrCodeChecksumMismatch},
		{"invalid frame", errors.New("invalid frame received"), ErrCodeInvalidFrame},
		{"unknown error", errors.New("something else"), ErrCodeUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _ := classifyError(tt.err)
			if code != tt.expectedCode {
				t.Errorf("classifyError(%v) = %v, want %v", tt.err, code, tt.expectedCode)
			}
		})
	}
}

func TestClassifyPortError(t *testing.T) {
	portErr := &transport.PortError{
		Code:    transport.ErrCodePortNotFound,
		Port:    "/dev/ttyUSB0",
		Message: "port not found",
	}

	code, severity := classifyError(portErr)
	if code != ErrCodePortNotFound {
		t.Errorf("classifyError(PortError) code = %v, want %v", code, ErrCodePortNotFound)
	}
	if severity != SeverityCritical {
		t.Errorf("classifyError(PortError) severity = %v, want %v", severity, SeverityCritical)
	}
}

func TestClassifyConnError(t *testing.T) {
	connErr := &transport.ConnError{
		Code:    transport.ErrCodeConnRefused,
		Address: "192.168.1.100:5000",
		Message: "connection refused",
	}

	code, severity := classifyError(connErr)
	if code != ErrCodeTCPConnFailed {
		t.Errorf("classifyError(ConnError) code = %v, want %v", code, ErrCodeTCPConnFailed)
	}
	if severity != SeverityCritical {
		t.Errorf("classifyError(ConnError) severity = %v, want %v", severity, SeverityCritical)
	}
}

func TestManagerEventHandlers(t *testing.T) {
	config := DefaultConfig()
	config.Transport.Serial.Port = "/dev/ttyUSB0"

	mgr, err := NewManager(config)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	var dataReceived, errorReceived, healthReceived bool
	// These are set by handlers but we only verify handler registration
	_, _, _ = dataReceived, errorReceived, healthReceived

	mgr.OnData(func(e DataEvent) {
		dataReceived = true
	})

	mgr.OnError(func(e ErrorEvent) {
		errorReceived = true
	})

	mgr.OnHealth(func(e HealthEvent) {
		healthReceived = true
	})

	// Verify handlers are registered (internal check)
	mgr.mu.RLock()
	dataHandlerCount := len(mgr.dataHandlers)
	errorHandlerCount := len(mgr.errorHandlers)
	healthHandlerCount := len(mgr.healthHandlers)
	mgr.mu.RUnlock()

	if dataHandlerCount != 1 {
		t.Errorf("Data handler count = %v, want 1", dataHandlerCount)
	}
	if errorHandlerCount != 1 {
		t.Errorf("Error handler count = %v, want 1", errorHandlerCount)
	}
	if healthHandlerCount != 1 {
		t.Errorf("Health handler count = %v, want 1", healthHandlerCount)
	}
}

func TestManagerWithMockTransport(t *testing.T) {
	config := DefaultConfig()
	config.Transport.Serial.Port = "/dev/ttyUSB0"
	config.Polling.Interval = 50 * time.Millisecond
	config.Devices = []DeviceConfig{
		{Address: 1, Name: "Meter 1", Enabled: true},
	}

	mgr, err := NewManager(config)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Replace transport with mock
	mock := newMockTransport()
	mgr.transport = mock

	// Subscribe to events
	var dataEvents []DataEvent
	var errorEvents []ErrorEvent
	var mu sync.Mutex

	mgr.OnData(func(e DataEvent) {
		mu.Lock()
		dataEvents = append(dataEvents, e)
		mu.Unlock()
	})

	mgr.OnError(func(e ErrorEvent) {
		mu.Lock()
		errorEvents = append(errorEvents, e)
		mu.Unlock()
	})

	ctx, cancel := context.WithCancel(context.Background())

	err = mgr.Start(ctx)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Wait for some polling
	time.Sleep(200 * time.Millisecond)

	cancel()
	mgr.Stop()

	// Verify transport was used
	mock.mu.Lock()
	openCount := mock.openCalled
	closeCount := mock.closeCalled
	mock.mu.Unlock()

	if openCount < 1 {
		t.Error("Transport Open should have been called")
	}
	if closeCount < 1 {
		t.Error("Transport Close should have been called")
	}
}

func TestManagerStartStop(t *testing.T) {
	config := DefaultConfig()
	config.Transport.Serial.Port = "/dev/ttyUSB0"

	mgr, err := NewManager(config)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Replace with mock
	mock := newMockTransport()
	mgr.transport = mock

	ctx := context.Background()

	// Start
	err = mgr.Start(ctx)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	health := mgr.GetHealth()
	if health.State != StateRunning {
		t.Errorf("State = %v, want %v", health.State, StateRunning)
	}

	// Start again should fail
	err = mgr.Start(ctx)
	if err == nil {
		t.Error("Second Start should fail")
	}

	// Stop
	err = mgr.Stop()
	if err != nil {
		t.Errorf("Stop failed: %v", err)
	}

	health = mgr.GetHealth()
	if health.State != StateStopped {
		t.Errorf("State after stop = %v, want %v", health.State, StateStopped)
	}

	// Stop again should fail
	err = mgr.Stop()
	if err == nil {
		t.Error("Second Stop should fail")
	}
}

func TestManagerPingDevice(t *testing.T) {
	config := DefaultConfig()
	config.Transport.Serial.Port = "/dev/ttyUSB0"
	config.Devices = []DeviceConfig{
		{Address: 1, Name: "Meter 1", Enabled: true},
	}

	mgr, err := NewManager(config)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Replace with mock that returns ACK
	mock := newMockTransport()
	mock.SetSendResponse([]byte{0xE5}) // ACK
	mgr.transport = mock

	ctx := context.Background()
	mgr.Start(ctx)
	defer mgr.Stop()

	// Ping device
	alive, err := mgr.PingDevice(1)
	if err != nil {
		t.Errorf("PingDevice failed: %v", err)
	}
	if !alive {
		t.Error("PingDevice should return true for ACK response")
	}

	// Test with no ACK
	mock.SetSendResponse([]byte{0x00})
	alive, err = mgr.PingDevice(1)
	if err != nil {
		t.Errorf("PingDevice failed: %v", err)
	}
	if alive {
		t.Error("PingDevice should return false for non-ACK response")
	}
}

func TestManagerGetHealth(t *testing.T) {
	config := DefaultConfig()
	config.Transport.Serial.Port = "/dev/ttyUSB0"
	config.Devices = []DeviceConfig{
		{Address: 1, Name: "Meter 1", Enabled: true},
		{Address: 2, Name: "Meter 2", Enabled: true},
	}

	mgr, err := NewManager(config)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	mock := newMockTransport()
	mgr.transport = mock

	ctx := context.Background()
	mgr.Start(ctx)
	defer mgr.Stop()

	health := mgr.GetHealth()

	if health.State != StateRunning {
		t.Errorf("State = %v, want %v", health.State, StateRunning)
	}
	if len(health.Devices) != 2 {
		t.Errorf("Devices count = %v, want 2", len(health.Devices))
	}
	if health.Transport.Type != "mock" {
		t.Errorf("Transport type = %v, want 'mock'", health.Transport.Type)
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		s        string
		substr   string
		expected bool
	}{
		{"hello world", "world", true},
		{"hello world", "foo", false},
		{"hello", "hello", true},
		{"", "foo", false},
		{"foo", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.s+"_"+tt.substr, func(t *testing.T) {
			if got := contains(tt.s, tt.substr); got != tt.expected {
				t.Errorf("contains(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.expected)
			}
		})
	}
}
