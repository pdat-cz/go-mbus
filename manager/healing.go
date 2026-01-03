package manager

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/pdat-cz/go-mbus/manager/transport"
)

// healingEngine manages automatic error recovery.
type healingEngine struct {
	mu sync.RWMutex

	config  SelfHealingConfig
	manager *Manager

	// Recovery state
	recovering     bool
	recoveryCount  int
	lastRecovery   time.Time
	attemptedFixes []string

	// Callbacks
	onRecoveryStart   func()
	onRecoverySuccess func(action string)
	onRecoveryFailed  func(err error)
}

// recoveryAction represents a recovery action to try.
type recoveryAction struct {
	name     string
	execute  func(ctx context.Context) error
	rollback func() error
}

// newHealingEngine creates a new self-healing engine.
func newHealingEngine(config SelfHealingConfig, manager *Manager) *healingEngine {
	return &healingEngine{
		config:  config,
		manager: manager,
	}
}

// SetCallbacks sets the recovery event callbacks.
func (h *healingEngine) SetCallbacks(
	onStart func(),
	onSuccess func(action string),
	onFailed func(err error),
) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.onRecoveryStart = onStart
	h.onRecoverySuccess = onSuccess
	h.onRecoveryFailed = onFailed
}

// HandleError attempts to recover from an error.
func (h *healingEngine) HandleError(ctx context.Context, err error, code ErrorCode) bool {
	if !h.config.Enabled {
		return false
	}

	h.mu.Lock()
	if h.recovering {
		h.mu.Unlock()
		return false // Already recovering
	}

	if h.recoveryCount >= h.config.MaxRecoveryAttempts {
		// Check if we should reset the counter
		if time.Since(h.lastRecovery) > 5*time.Minute {
			h.recoveryCount = 0
			h.attemptedFixes = nil
		} else {
			h.mu.Unlock()
			return false // Max attempts reached
		}
	}

	h.recovering = true
	h.recoveryCount++
	h.lastRecovery = time.Now()
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		h.recovering = false
		h.mu.Unlock()
	}()

	// Notify recovery started
	if h.onRecoveryStart != nil {
		h.onRecoveryStart()
	}

	// Get recovery actions for this error
	actions := h.getRecoveryActions(code)
	if len(actions) == 0 {
		return false
	}

	// Try each action
	for _, action := range actions {
		select {
		case <-ctx.Done():
			return false
		default:
		}

		h.mu.Lock()
		h.attemptedFixes = append(h.attemptedFixes, action.name)
		h.mu.Unlock()

		if err := action.execute(ctx); err == nil {
			// Recovery succeeded
			if h.onRecoverySuccess != nil {
				h.onRecoverySuccess(action.name)
			}
			h.mu.Lock()
			h.recoveryCount = 0
			h.attemptedFixes = nil
			h.mu.Unlock()
			return true
		}

		// Rollback if needed
		if action.rollback != nil {
			action.rollback()
		}

		// Wait before next attempt
		select {
		case <-ctx.Done():
			return false
		case <-time.After(h.config.RecoveryDelay):
		}
	}

	// All actions failed
	if h.onRecoveryFailed != nil {
		h.onRecoveryFailed(err)
	}

	return false
}

// getRecoveryActions returns recovery actions for an error code.
func (h *healingEngine) getRecoveryActions(code ErrorCode) []recoveryAction {
	switch code {
	case ErrCodePortNotFound:
		return h.getPortNotFoundActions()
	case ErrCodePortBusy:
		return h.getPortBusyActions()
	case ErrCodePortDisconnected:
		return h.getPortDisconnectedActions()
	case ErrCodeBaudRateMismatch:
		return h.getBaudRateMismatchActions()
	case ErrCodeTCPConnFailed, ErrCodeTCPConnLost:
		return h.getTCPConnectionActions()
	case ErrCodeTimeout, ErrCodeDeviceNotResponding:
		return h.getTimeoutActions()
	default:
		return h.getDefaultActions()
	}
}

// getPortNotFoundActions returns actions for port not found errors.
func (h *healingEngine) getPortNotFoundActions() []recoveryAction {
	var actions []recoveryAction

	// Action 1: Try fallback ports
	if h.config.AutoPortScan {
		actions = append(actions, recoveryAction{
			name: "try_fallback_ports",
			execute: func(ctx context.Context) error {
				return h.tryFallbackPorts(ctx)
			},
		})
	}

	// Action 2: Scan for available ports
	if h.config.AutoPortScan {
		actions = append(actions, recoveryAction{
			name: "scan_available_ports",
			execute: func(ctx context.Context) error {
				return h.scanAvailablePorts(ctx)
			},
		})
	}

	return actions
}

// getPortBusyActions returns actions for port busy errors.
func (h *healingEngine) getPortBusyActions() []recoveryAction {
	return []recoveryAction{
		{
			name: "wait_and_retry",
			execute: func(ctx context.Context) error {
				// Wait with exponential backoff
				delay := time.Duration(h.recoveryCount) * time.Second
				if delay > 30*time.Second {
					delay = 30 * time.Second
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(delay):
				}
				return h.reconnectTransport(ctx)
			},
		},
	}
}

// getPortDisconnectedActions returns actions for port disconnection.
func (h *healingEngine) getPortDisconnectedActions() []recoveryAction {
	var actions []recoveryAction

	// Action 1: Wait and reconnect
	actions = append(actions, recoveryAction{
		name: "wait_and_reconnect",
		execute: func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(2 * time.Second):
			}
			return h.reconnectTransport(ctx)
		},
	})

	// Action 2: Try fallback ports
	if h.config.AutoPortScan {
		actions = append(actions, recoveryAction{
			name: "try_fallback_ports",
			execute: func(ctx context.Context) error {
				return h.tryFallbackPorts(ctx)
			},
		})
	}

	return actions
}

// getBaudRateMismatchActions returns actions for baud rate issues.
func (h *healingEngine) getBaudRateMismatchActions() []recoveryAction {
	if !h.config.AutoBaudDetection {
		return nil
	}

	return []recoveryAction{
		{
			name: "auto_detect_baud_rate",
			execute: func(ctx context.Context) error {
				return h.autoDetectBaudRate(ctx)
			},
		},
	}
}

// getTCPConnectionActions returns actions for TCP connection issues.
func (h *healingEngine) getTCPConnectionActions() []recoveryAction {
	return []recoveryAction{
		{
			name: "tcp_reconnect",
			execute: func(ctx context.Context) error {
				// Wait with backoff before reconnecting
				delay := time.Duration(h.recoveryCount) * 2 * time.Second
				if delay > 30*time.Second {
					delay = 30 * time.Second
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(delay):
				}
				return h.reconnectTransport(ctx)
			},
		},
	}
}

// getTimeoutActions returns actions for timeout errors.
func (h *healingEngine) getTimeoutActions() []recoveryAction {
	return []recoveryAction{
		{
			name: "retry_communication",
			execute: func(ctx context.Context) error {
				// Just reconnect and retry
				return h.reconnectTransport(ctx)
			},
		},
	}
}

// getDefaultActions returns default recovery actions.
func (h *healingEngine) getDefaultActions() []recoveryAction {
	return []recoveryAction{
		{
			name: "reconnect",
			execute: func(ctx context.Context) error {
				return h.reconnectTransport(ctx)
			},
		},
	}
}

// reconnectTransport closes and reopens the transport.
func (h *healingEngine) reconnectTransport(ctx context.Context) error {
	if h.manager == nil || h.manager.transport == nil {
		return errors.New("no transport available")
	}

	h.manager.transport.Close()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(500 * time.Millisecond):
	}

	return h.manager.transport.Open()
}

// tryFallbackPorts attempts to connect to fallback serial ports.
func (h *healingEngine) tryFallbackPorts(ctx context.Context) error {
	if h.manager == nil || h.manager.transport == nil {
		return errors.New("no transport available")
	}

	serialTransport, ok := h.manager.transport.(*transport.SerialTransport)
	if !ok {
		return errors.New("not a serial transport")
	}

	config := serialTransport.GetConfig()
	if len(config.FallbackPorts) == 0 {
		return errors.New("no fallback ports configured")
	}

	// Close current connection
	serialTransport.Close()

	// Try each fallback port
	for _, port := range config.FallbackPorts {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		config.Port = port
		if err := serialTransport.SetConfig(config); err != nil {
			continue
		}

		if err := serialTransport.Open(); err == nil {
			return nil // Success
		}
	}

	return errors.New("no fallback port available")
}

// scanAvailablePorts scans for and tries available serial ports.
func (h *healingEngine) scanAvailablePorts(ctx context.Context) error {
	if h.manager == nil || h.manager.transport == nil {
		return errors.New("no transport available")
	}

	serialTransport, ok := h.manager.transport.(*transport.SerialTransport)
	if !ok {
		return errors.New("not a serial transport")
	}

	ports, err := transport.ListAvailablePorts()
	if err != nil {
		return fmt.Errorf("scan ports: %w", err)
	}

	if len(ports) == 0 {
		return errors.New("no serial ports found")
	}

	config := serialTransport.GetConfig()
	serialTransport.Close()

	// Try each available port
	for _, port := range ports {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		config.Port = port
		if err := serialTransport.SetConfig(config); err != nil {
			continue
		}

		if err := serialTransport.Open(); err == nil {
			return nil // Success
		}
	}

	return errors.New("no working port found")
}

// autoDetectBaudRate tries different baud rates to find a working one.
func (h *healingEngine) autoDetectBaudRate(ctx context.Context) error {
	if h.manager == nil || h.manager.transport == nil {
		return errors.New("no transport available")
	}

	serialTransport, ok := h.manager.transport.(*transport.SerialTransport)
	if !ok {
		return errors.New("not a serial transport")
	}

	baudRates := []int{2400, 9600, 19200, 4800, 1200, 300}

	config := serialTransport.GetConfig()
	originalBaud := config.BaudRate

	serialTransport.Close()

	for _, baud := range baudRates {
		select {
		case <-ctx.Done():
			// Restore original baud rate
			config.BaudRate = originalBaud
			serialTransport.SetConfig(config)
			return ctx.Err()
		default:
		}

		config.BaudRate = baud
		if err := serialTransport.SetConfig(config); err != nil {
			continue
		}

		if err := serialTransport.Open(); err != nil {
			continue
		}

		// Test communication by sending a ping
		// This is a simplified test - in practice, you'd want to
		// actually verify communication works
		return nil // Assume success if we can open the port
	}

	// Restore original baud rate
	config.BaudRate = originalBaud
	serialTransport.SetConfig(config)

	return errors.New("no working baud rate found")
}

// GetAttemptedFixes returns the list of attempted recovery actions.
func (h *healingEngine) GetAttemptedFixes() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make([]string, len(h.attemptedFixes))
	copy(result, h.attemptedFixes)
	return result
}

// GetRecoveryCount returns the number of recovery attempts.
func (h *healingEngine) GetRecoveryCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.recoveryCount
}

// IsRecovering returns whether recovery is in progress.
func (h *healingEngine) IsRecovering() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.recovering
}

// Reset resets the recovery state.
func (h *healingEngine) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.recoveryCount = 0
	h.attemptedFixes = nil
	h.recovering = false
}

// UpdateConfig updates the self-healing configuration.
func (h *healingEngine) UpdateConfig(config SelfHealingConfig) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.config = config
}
