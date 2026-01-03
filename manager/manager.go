package manager

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/pdat-cz/go-mbus/manager/transport"
	"github.com/pdat-cz/go-mbus/pkg/mbus"
)

const (
	// Default channel buffer sizes
	defaultDataChannelSize   = 100
	defaultErrorChannelSize  = 100
	defaultHealthChannelSize = 50
)

// Manager is the main M-Bus communication manager.
type Manager struct {
	mu sync.RWMutex

	// Configuration
	config Config

	// Components
	transport transport.Transport
	scheduler *scheduler
	healing   *healingEngine

	// State
	state     ManagerState
	startedAt time.Time
	stats     ManagerStats

	// Device state tracking
	devices map[int]*deviceState

	// Event channels
	dataEvents   chan DataEvent
	errorEvents  chan ErrorEvent
	healthEvents chan HealthEvent

	// Event handlers
	dataHandlers   []func(DataEvent)
	errorHandlers  []func(ErrorEvent)
	healthHandlers []func(HealthEvent)

	// Context and cancellation
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// deviceState tracks the runtime state of a device.
type deviceState struct {
	config              DeviceConfig
	online              bool
	lastSeen            time.Time
	lastError           error
	consecutiveFailures int
	totalReads          uint64
	totalErrors         uint64
}

// NewManager creates a new M-Bus manager with the given configuration.
func NewManager(config Config) (*Manager, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	m := &Manager{
		config:       config,
		state:        StateIdle,
		devices:      make(map[int]*deviceState),
		dataEvents:   make(chan DataEvent, defaultDataChannelSize),
		errorEvents:  make(chan ErrorEvent, defaultErrorChannelSize),
		healthEvents: make(chan HealthEvent, defaultHealthChannelSize),
	}

	// Create transport
	var err error
	m.transport, err = m.createTransport()
	if err != nil {
		return nil, fmt.Errorf("create transport: %w", err)
	}

	// Create scheduler
	m.scheduler = newScheduler(config.Polling.Interval, config.Polling.StaggerDelay)
	m.scheduler.SetOnPoll(m.pollDevice)

	// Create healing engine
	m.healing = newHealingEngine(config.SelfHealing, m)
	m.healing.SetCallbacks(
		func() {
			m.emitHealthEvent(HealthEventRecoveryStarted, "healing", "Recovery started", nil)
		},
		func(action string) {
			m.emitHealthEvent(HealthEventRecoverySuccess, "healing",
				fmt.Sprintf("Recovery succeeded: %s", action),
				map[string]interface{}{"action": action})
		},
		func(err error) {
			m.emitHealthEvent(HealthEventRecoveryFailed, "healing",
				fmt.Sprintf("Recovery failed: %v", err),
				map[string]interface{}{"error": err.Error()})
		},
	)

	// Initialize device states
	for _, deviceConfig := range config.Devices {
		m.devices[deviceConfig.Address] = &deviceState{
			config: deviceConfig,
			online: false,
		}
		m.scheduler.AddDevice(deviceConfig)
	}

	return m, nil
}

// createTransport creates the appropriate transport based on configuration.
func (m *Manager) createTransport() (transport.Transport, error) {
	switch m.config.Transport.Type {
	case "serial":
		if m.config.Transport.Serial == nil {
			return nil, errors.New("serial configuration required")
		}
		return transport.NewSerialTransport(*m.config.Transport.Serial), nil

	case "tcp":
		if m.config.Transport.TCP == nil {
			return nil, errors.New("TCP configuration required")
		}
		return transport.NewTCPTransport(*m.config.Transport.TCP), nil

	default:
		return nil, fmt.Errorf("unknown transport type: %s", m.config.Transport.Type)
	}
}

// Start begins the manager operation.
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	if m.state == StateRunning {
		m.mu.Unlock()
		return errors.New("manager already running")
	}

	m.state = StateStarting
	m.ctx, m.cancel = context.WithCancel(ctx)
	m.mu.Unlock()

	// Open transport
	if err := m.transport.Open(); err != nil {
		m.mu.Lock()
		m.state = StateError
		m.mu.Unlock()

		m.emitError(0, "", err, ErrCodePortNotFound, SeverityCritical, true)
		return fmt.Errorf("open transport: %w", err)
	}

	m.mu.Lock()
	m.state = StateRunning
	m.startedAt = time.Now()
	m.mu.Unlock()

	// Start event dispatcher
	m.wg.Add(1)
	go m.eventDispatcher()

	// Start scheduler
	m.scheduler.Start(m.ctx)

	// Start health check loop
	if m.config.SelfHealing.HealthCheckInterval > 0 {
		m.wg.Add(1)
		go m.healthCheckLoop()
	}

	m.emitHealthEvent(HealthEventStarted, "manager", "Manager started", nil)
	m.emitHealthEvent(HealthEventConnected, "transport",
		fmt.Sprintf("Connected to %s", m.transport.GetInfo().Address), nil)

	return nil
}

// Stop gracefully shuts down the manager.
func (m *Manager) Stop() error {
	m.mu.Lock()
	if m.state != StateRunning {
		m.mu.Unlock()
		return errors.New("manager not running")
	}

	m.state = StateStopping
	m.mu.Unlock()

	// Stop scheduler
	m.scheduler.Stop()

	// Cancel context
	if m.cancel != nil {
		m.cancel()
	}

	// Wait for goroutines
	m.wg.Wait()

	// Close transport
	if err := m.transport.Close(); err != nil {
		m.emitHealthEvent(HealthEventWarning, "transport",
			fmt.Sprintf("Error closing transport: %v", err), nil)
	}

	m.mu.Lock()
	m.state = StateStopped
	m.mu.Unlock()

	m.emitHealthEvent(HealthEventStopped, "manager", "Manager stopped", nil)

	// Close event channels
	close(m.dataEvents)
	close(m.errorEvents)
	close(m.healthEvents)

	return nil
}

// OnData registers a handler for data events.
func (m *Manager) OnData(handler func(DataEvent)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.dataHandlers = append(m.dataHandlers, handler)
}

// OnError registers a handler for error events.
func (m *Manager) OnError(handler func(ErrorEvent)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errorHandlers = append(m.errorHandlers, handler)
}

// OnHealth registers a handler for health events.
func (m *Manager) OnHealth(handler func(HealthEvent)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.healthHandlers = append(m.healthHandlers, handler)
}

// eventDispatcher dispatches events to registered handlers.
func (m *Manager) eventDispatcher() {
	defer m.wg.Done()

	for {
		select {
		case <-m.ctx.Done():
			// Drain remaining events
			for {
				select {
				case event := <-m.dataEvents:
					m.dispatchDataEvent(event)
				case event := <-m.errorEvents:
					m.dispatchErrorEvent(event)
				case event := <-m.healthEvents:
					m.dispatchHealthEvent(event)
				default:
					return
				}
			}

		case event := <-m.dataEvents:
			m.dispatchDataEvent(event)

		case event := <-m.errorEvents:
			m.dispatchErrorEvent(event)

		case event := <-m.healthEvents:
			m.dispatchHealthEvent(event)
		}
	}
}

func (m *Manager) dispatchDataEvent(event DataEvent) {
	m.mu.RLock()
	handlers := make([]func(DataEvent), len(m.dataHandlers))
	copy(handlers, m.dataHandlers)
	m.mu.RUnlock()

	for _, handler := range handlers {
		handler(event)
	}
}

func (m *Manager) dispatchErrorEvent(event ErrorEvent) {
	m.mu.RLock()
	handlers := make([]func(ErrorEvent), len(m.errorHandlers))
	copy(handlers, m.errorHandlers)
	m.mu.RUnlock()

	for _, handler := range handlers {
		handler(event)
	}
}

func (m *Manager) dispatchHealthEvent(event HealthEvent) {
	m.mu.RLock()
	handlers := make([]func(HealthEvent), len(m.healthHandlers))
	copy(handlers, m.healthHandlers)
	m.mu.RUnlock()

	for _, handler := range handlers {
		handler(event)
	}
}

// pollDevice is called by the scheduler to poll a device.
func (m *Manager) pollDevice(address int) {
	m.mu.RLock()
	dev, ok := m.devices[address]
	if !ok {
		m.mu.RUnlock()
		return
	}
	deviceName := dev.config.Name
	m.mu.RUnlock()

	startTime := time.Now()

	// Attempt to read the device with retries
	var lastErr error
	var data mbus.LFrameParsed

	for attempt := 0; attempt <= m.config.Polling.RetryCount; attempt++ {
		if attempt > 0 {
			time.Sleep(m.config.Polling.RetryDelay)
		}

		var err error
		data, err = m.readDevice(address)
		if err == nil {
			// Success
			m.handleReadSuccess(address, deviceName, data, time.Since(startTime))
			return
		}

		lastErr = err
	}

	// All retries failed
	m.handleReadFailure(address, deviceName, lastErr)
}

// readDevice performs the actual device read.
func (m *Manager) readDevice(address int) (mbus.LFrameParsed, error) {
	m.transport.Lock()
	defer m.transport.Unlock()

	if !m.transport.IsOpen() {
		return mbus.LFrameParsed{}, errors.New("transport not open")
	}

	// Build REQ_UD2 command
	command := buildREQUD2Command(uint(address))

	// Send command and receive response
	response, err := m.transport.Send(command, m.config.Polling.Timeout)
	if err != nil {
		return mbus.LFrameParsed{}, err
	}

	if len(response) == 0 {
		return mbus.LFrameParsed{}, errors.New("no response from device")
	}

	// Parse the response
	frame := mbus.NewLFrame(response)
	return frame.Parse()
}

// handleReadSuccess processes a successful device read.
func (m *Manager) handleReadSuccess(address int, name string, data mbus.LFrameParsed, latency time.Duration) {
	m.mu.Lock()
	if dev, ok := m.devices[address]; ok {
		wasOffline := !dev.online
		dev.online = true
		dev.lastSeen = time.Now()
		dev.lastError = nil
		dev.consecutiveFailures = 0
		dev.totalReads++

		if wasOffline {
			m.mu.Unlock()
			m.emitHealthEvent(HealthEventDeviceOnline, "device",
				fmt.Sprintf("Device %s (addr=%d) is online", name, address),
				map[string]interface{}{"address": address, "name": name})
			m.mu.Lock()
		}
	}

	m.stats.TotalReads++
	m.stats.SuccessfulReads++
	m.mu.Unlock()

	// Emit data event
	event := DataEvent{
		Timestamp:   time.Now(),
		DeviceAddr:  address,
		DeviceName:  name,
		Data:        data,
		ReadLatency: latency,
	}

	select {
	case m.dataEvents <- event:
	default:
		// Channel full, drop event
	}
}

// handleReadFailure processes a failed device read.
func (m *Manager) handleReadFailure(address int, name string, err error) {
	m.mu.Lock()
	if dev, ok := m.devices[address]; ok {
		wasOnline := dev.online
		dev.consecutiveFailures++
		dev.totalErrors++
		dev.lastError = err

		// Mark offline after threshold
		if dev.consecutiveFailures >= 3 {
			dev.online = false
			if wasOnline {
				m.mu.Unlock()
				m.emitHealthEvent(HealthEventDeviceOffline, "device",
					fmt.Sprintf("Device %s (addr=%d) is offline", name, address),
					map[string]interface{}{"address": address, "name": name, "error": err.Error()})
				m.mu.Lock()
			}
		}
	}

	m.stats.TotalReads++
	m.stats.FailedReads++
	m.mu.Unlock()

	// Classify error and emit
	code, severity := classifyError(err)
	m.emitError(address, name, err, code, severity, true)

	// Attempt self-healing for transport-level errors
	if code >= ErrCodePortNotFound && code <= ErrCodeTCPTimeout {
		m.healing.HandleError(m.ctx, err, code)
	}
}

// emitError emits an error event.
func (m *Manager) emitError(address int, name string, err error, code ErrorCode, severity Severity, recoverable bool) {
	event := ErrorEvent{
		Timestamp:   time.Now(),
		DeviceAddr:  address,
		DeviceName:  name,
		Error:       err,
		Code:        code,
		Severity:    severity,
		Suggestion:  code.Suggestion(),
		Recoverable: recoverable,
		Context: ErrorContext{
			Transport:        m.config.Transport.Type,
			TransportAddress: m.transport.GetInfo().Address,
			AttemptedFixes:   m.healing.GetAttemptedFixes(),
		},
	}

	// Add device-specific context
	m.mu.RLock()
	if dev, ok := m.devices[address]; ok {
		event.Context.LastSuccessful = dev.lastSeen
		event.Context.FailureCount = dev.consecutiveFailures
	}
	m.mu.RUnlock()

	select {
	case m.errorEvents <- event:
	default:
		// Channel full, drop event
	}
}

// emitHealthEvent emits a health event.
func (m *Manager) emitHealthEvent(eventType HealthEventType, component, message string, details map[string]interface{}) {
	event := HealthEvent{
		Timestamp: time.Now(),
		Type:      eventType,
		Component: component,
		Message:   message,
		Details:   details,
	}

	select {
	case m.healthEvents <- event:
	default:
		// Channel full, drop event
	}
}

// healthCheckLoop periodically checks system health.
func (m *Manager) healthCheckLoop() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.config.SelfHealing.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.performHealthCheck()
		}
	}
}

// performHealthCheck checks the health of all components.
func (m *Manager) performHealthCheck() {
	// Check transport
	if !m.transport.IsOpen() {
		m.emitHealthEvent(HealthEventDisconnected, "transport",
			"Transport connection lost", nil)

		// Attempt recovery
		if m.config.SelfHealing.Enabled {
			m.healing.HandleError(m.ctx, errors.New("transport disconnected"), ErrCodePortDisconnected)
		}
	}
}

// ReadDevice performs an immediate read of a specific device.
func (m *Manager) ReadDevice(address int) (DataEvent, error) {
	m.mu.RLock()
	dev, ok := m.devices[address]
	if !ok {
		m.mu.RUnlock()
		return DataEvent{}, fmt.Errorf("device at address %d not found", address)
	}
	deviceName := dev.config.Name
	m.mu.RUnlock()

	startTime := time.Now()

	data, err := m.readDevice(address)
	if err != nil {
		return DataEvent{}, err
	}

	return DataEvent{
		Timestamp:   time.Now(),
		DeviceAddr:  address,
		DeviceName:  deviceName,
		Data:        data,
		ReadLatency: time.Since(startTime),
	}, nil
}

// PingDevice checks if a device is responding.
func (m *Manager) PingDevice(address int) (bool, error) {
	m.transport.Lock()
	defer m.transport.Unlock()

	if !m.transport.IsOpen() {
		return false, errors.New("transport not open")
	}

	// Build SND_NKE command
	command := buildSNDNKECommand(uint(address))

	// Send command and receive response
	response, err := m.transport.Send(command, 500*time.Millisecond)
	if err != nil {
		return false, err
	}

	// Check for ACK (0xE5)
	if len(response) > 0 && response[0] == 0xE5 {
		return true, nil
	}

	return false, nil
}

// AddDevice adds a new device to the manager.
func (m *Manager) AddDevice(config DeviceConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.devices[config.Address]; exists {
		return fmt.Errorf("device at address %d already exists", config.Address)
	}

	m.devices[config.Address] = &deviceState{
		config: config,
		online: false,
	}
	m.scheduler.AddDevice(config)

	return nil
}

// RemoveDevice removes a device from the manager.
func (m *Manager) RemoveDevice(address int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.devices[address]; !exists {
		return fmt.Errorf("device at address %d not found", address)
	}

	delete(m.devices, address)
	m.scheduler.RemoveDevice(address)

	return nil
}

// ListDevices returns the status of all devices.
func (m *Manager) ListDevices() []DeviceStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []DeviceStatus
	for _, dev := range m.devices {
		status := DeviceStatus{
			Address:             dev.config.Address,
			Name:                dev.config.Name,
			Online:              dev.online,
			LastSeen:            dev.lastSeen,
			ConsecutiveFailures: dev.consecutiveFailures,
			TotalReads:          dev.totalReads,
			TotalErrors:         dev.totalErrors,
		}
		if dev.lastError != nil {
			status.LastError = dev.lastError.Error()
		}
		result = append(result, status)
	}

	return result
}

// GetHealth returns the current health status.
func (m *Manager) GetHealth() HealthStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	info := m.transport.GetInfo()

	return HealthStatus{
		State:     m.state,
		StartedAt: m.startedAt,
		Uptime:    time.Since(m.startedAt),
		Transport: TransportHealth{
			Type:      info.Type,
			Address:   info.Address,
			Connected: info.State == transport.StateOpen,
			Reconnects: info.Stats.Reconnects,
		},
		Devices: m.ListDevices(),
		Stats:   m.stats,
	}
}

// GetConfig returns the current configuration.
func (m *Manager) GetConfig() Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// buildREQUD2Command builds a REQ_UD2 M-Bus command.
func buildREQUD2Command(address uint) []byte {
	cf := byte(0x5B) // REQ_UD2, FCB=1, FCV=1
	ad := byte(address)
	crc := cf + ad

	return []byte{0x10, cf, ad, crc, 0x16}
}

// buildSNDNKECommand builds a SND_NKE M-Bus command.
func buildSNDNKECommand(address uint) []byte {
	cf := byte(0x40) // SND_NKE
	ad := byte(address)
	crc := cf + ad

	return []byte{0x10, cf, ad, crc, 0x16}
}

// classifyError determines the error code and severity for an error.
func classifyError(err error) (ErrorCode, Severity) {
	if err == nil {
		return ErrCodeUnknown, SeverityInfo
	}

	errStr := err.Error()

	// Check for port errors
	var portErr *transport.PortError
	if errors.As(err, &portErr) {
		switch portErr.Code {
		case transport.ErrCodePortNotFound:
			return ErrCodePortNotFound, SeverityCritical
		case transport.ErrCodePortBusy:
			return ErrCodePortBusy, SeverityWarning
		case transport.ErrCodePortAccessDenied:
			return ErrCodePortAccessDenied, SeverityCritical
		case transport.ErrCodePortDisconnected:
			return ErrCodePortDisconnected, SeverityError
		}
	}

	// Check for connection errors
	var connErr *transport.ConnError
	if errors.As(err, &connErr) {
		switch connErr.Code {
		case transport.ErrCodeConnRefused:
			return ErrCodeTCPConnFailed, SeverityCritical
		case transport.ErrCodeConnTimeout:
			return ErrCodeTCPTimeout, SeverityWarning
		case transport.ErrCodeConnLost:
			return ErrCodeTCPConnLost, SeverityError
		}
	}

	// Check error message for common patterns
	switch {
	case contains(errStr, "timeout"):
		return ErrCodeTimeout, SeverityWarning
	case contains(errStr, "no response"):
		return ErrCodeDeviceNotResponding, SeverityWarning
	case contains(errStr, "checksum"):
		return ErrCodeChecksumMismatch, SeverityError
	case contains(errStr, "invalid frame"):
		return ErrCodeInvalidFrame, SeverityError
	default:
		return ErrCodeUnknown, SeverityError
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsImpl(s, substr))
}

func containsImpl(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
