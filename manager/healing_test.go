package manager

import (
	"context"
	"testing"
	"time"
)

func TestNewHealingEngine(t *testing.T) {
	config := SelfHealingConfig{
		Enabled:             true,
		MaxRecoveryAttempts: 5,
		RecoveryDelay:       1 * time.Second,
	}

	h := newHealingEngine(config, nil)

	if h == nil {
		t.Fatal("newHealingEngine should not return nil")
	}
	if !h.config.Enabled {
		t.Error("Healing should be enabled")
	}
	if h.IsRecovering() {
		t.Error("Should not be recovering initially")
	}
}

func TestHealingEngineDisabled(t *testing.T) {
	config := SelfHealingConfig{
		Enabled: false,
	}

	h := newHealingEngine(config, nil)

	ctx := context.Background()
	result := h.HandleError(ctx, nil, ErrCodePortNotFound)

	if result {
		t.Error("HandleError should return false when disabled")
	}
}

func TestHealingEngineMaxAttempts(t *testing.T) {
	config := SelfHealingConfig{
		Enabled:             true,
		MaxRecoveryAttempts: 2,
		RecoveryDelay:       10 * time.Millisecond,
	}

	h := newHealingEngine(config, nil)

	ctx := context.Background()

	// First attempt
	h.HandleError(ctx, nil, ErrCodeUnknown)
	if h.GetRecoveryCount() != 1 {
		t.Errorf("Recovery count = %v, want 1", h.GetRecoveryCount())
	}

	// Second attempt
	h.HandleError(ctx, nil, ErrCodeUnknown)
	if h.GetRecoveryCount() != 2 {
		t.Errorf("Recovery count = %v, want 2", h.GetRecoveryCount())
	}

	// Third attempt should be blocked (max is 2)
	result := h.HandleError(ctx, nil, ErrCodeUnknown)
	if result {
		t.Error("HandleError should return false when max attempts reached")
	}
}

func TestHealingEngineReset(t *testing.T) {
	config := SelfHealingConfig{
		Enabled:             true,
		MaxRecoveryAttempts: 5,
		RecoveryDelay:       10 * time.Millisecond,
	}

	h := newHealingEngine(config, nil)

	ctx := context.Background()
	h.HandleError(ctx, nil, ErrCodeUnknown)

	if h.GetRecoveryCount() == 0 {
		t.Error("Recovery count should be > 0")
	}

	h.Reset()

	if h.GetRecoveryCount() != 0 {
		t.Errorf("Recovery count after reset = %v, want 0", h.GetRecoveryCount())
	}
	if len(h.GetAttemptedFixes()) != 0 {
		t.Error("Attempted fixes should be empty after reset")
	}
}

func TestHealingEngineCallbacks(t *testing.T) {
	config := SelfHealingConfig{
		Enabled:             true,
		MaxRecoveryAttempts: 5,
		RecoveryDelay:       10 * time.Millisecond,
	}

	h := newHealingEngine(config, nil)

	var startCalled, successCalled, failedCalled bool

	h.SetCallbacks(
		func() { startCalled = true },
		func(action string) { successCalled = true },
		func(err error) { failedCalled = true },
	)

	ctx := context.Background()
	h.HandleError(ctx, nil, ErrCodeUnknown)

	// At least start should be called
	if !startCalled {
		t.Error("Start callback should have been called")
	}
	// With nil manager, recovery will fail, so either success or failed should be called
	if !successCalled && !failedCalled {
		t.Error("Either success or failed callback should have been called")
	}
}

func TestHealingEngineUpdateConfig(t *testing.T) {
	config := SelfHealingConfig{
		Enabled:             true,
		MaxRecoveryAttempts: 5,
	}

	h := newHealingEngine(config, nil)

	newConfig := SelfHealingConfig{
		Enabled:             false,
		MaxRecoveryAttempts: 10,
	}

	h.UpdateConfig(newConfig)

	if h.config.Enabled {
		t.Error("Config should be updated to disabled")
	}
	if h.config.MaxRecoveryAttempts != 10 {
		t.Errorf("MaxRecoveryAttempts = %v, want 10", h.config.MaxRecoveryAttempts)
	}
}

func TestHealingEngineGetAttemptedFixes(t *testing.T) {
	config := SelfHealingConfig{
		Enabled:             true,
		MaxRecoveryAttempts: 5,
		RecoveryDelay:       10 * time.Millisecond,
	}

	h := newHealingEngine(config, nil)

	ctx := context.Background()
	h.HandleError(ctx, nil, ErrCodeUnknown)

	fixes := h.GetAttemptedFixes()
	if len(fixes) == 0 {
		t.Error("Should have attempted at least one fix")
	}
}

func TestHealingEngineContextCancellation(t *testing.T) {
	config := SelfHealingConfig{
		Enabled:             true,
		MaxRecoveryAttempts: 5,
		RecoveryDelay:       1 * time.Second,
	}

	h := newHealingEngine(config, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	result := h.HandleError(ctx, nil, ErrCodePortBusy)

	if result {
		t.Error("HandleError should return false when context is cancelled")
	}
}

func TestGetRecoveryActions(t *testing.T) {
	config := SelfHealingConfig{
		Enabled:             true,
		MaxRecoveryAttempts: 5,
		AutoBaudDetection:   true,
		AutoPortScan:        true,
	}

	h := newHealingEngine(config, nil)

	tests := []struct {
		code        ErrorCode
		minActions  int
		description string
	}{
		{ErrCodePortNotFound, 1, "port not found should have scan actions"},
		{ErrCodePortBusy, 1, "port busy should have wait action"},
		{ErrCodePortDisconnected, 1, "port disconnected should have reconnect actions"},
		{ErrCodeBaudRateMismatch, 1, "baud mismatch should have detection action"},
		{ErrCodeTCPConnFailed, 1, "TCP failed should have reconnect action"},
		{ErrCodeTCPConnLost, 1, "TCP lost should have reconnect action"},
		{ErrCodeTimeout, 1, "timeout should have retry action"},
		{ErrCodeDeviceNotResponding, 1, "device not responding should have retry action"},
		{ErrCodeUnknown, 1, "unknown should have default action"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			actions := h.getRecoveryActions(tt.code)
			if len(actions) < tt.minActions {
				t.Errorf("Expected at least %d actions for %v, got %d",
					tt.minActions, tt.code, len(actions))
			}
		})
	}
}

func TestGetRecoveryActionsDisabledFeatures(t *testing.T) {
	config := SelfHealingConfig{
		Enabled:             true,
		MaxRecoveryAttempts: 5,
		AutoBaudDetection:   false,
		AutoPortScan:        false,
	}

	h := newHealingEngine(config, nil)

	// Port not found without auto scan should have fewer actions
	actions := h.getRecoveryActions(ErrCodePortNotFound)
	if len(actions) > 0 {
		t.Error("Should have no port scan actions when AutoPortScan is disabled")
	}

	// Baud rate mismatch without auto detection should have no actions
	actions = h.getRecoveryActions(ErrCodeBaudRateMismatch)
	if len(actions) > 0 {
		t.Error("Should have no baud detection actions when AutoBaudDetection is disabled")
	}
}
