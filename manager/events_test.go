package manager

import (
	"testing"
)

func TestSeverityString(t *testing.T) {
	tests := []struct {
		severity Severity
		expected string
	}{
		{SeverityDebug, "debug"},
		{SeverityInfo, "info"},
		{SeverityWarning, "warning"},
		{SeverityError, "error"},
		{SeverityCritical, "critical"},
		{Severity(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.severity.String(); got != tt.expected {
				t.Errorf("Severity.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestErrorCodeString(t *testing.T) {
	tests := []struct {
		code     ErrorCode
		expected string
	}{
		{ErrCodePortNotFound, "PORT_NOT_FOUND"},
		{ErrCodePortBusy, "PORT_BUSY"},
		{ErrCodePortAccessDenied, "PORT_ACCESS_DENIED"},
		{ErrCodePortDisconnected, "PORT_DISCONNECTED"},
		{ErrCodeBaudRateMismatch, "BAUD_RATE_MISMATCH"},
		{ErrCodeParityError, "PARITY_ERROR"},
		{ErrCodeFrameError, "FRAME_ERROR"},
		{ErrCodeTCPConnFailed, "TCP_CONN_FAILED"},
		{ErrCodeTCPConnLost, "TCP_CONN_LOST"},
		{ErrCodeTCPTimeout, "TCP_TIMEOUT"},
		{ErrCodeChecksumMismatch, "CHECKSUM_MISMATCH"},
		{ErrCodeInvalidFrame, "INVALID_FRAME"},
		{ErrCodeInvalidResponse, "INVALID_RESPONSE"},
		{ErrCodeTimeout, "TIMEOUT"},
		{ErrCodeDeviceNotResponding, "DEVICE_NOT_RESPONDING"},
		{ErrCodeDeviceError, "DEVICE_ERROR"},
		{ErrCodeDeviceBusy, "DEVICE_BUSY"},
		{ErrCodeInvalidConfig, "INVALID_CONFIG"},
		{ErrCodeInvalidAddress, "INVALID_ADDRESS"},
		{ErrCodeUnknown, "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.code.String(); got != tt.expected {
				t.Errorf("ErrorCode.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestErrorCodeSuggestion(t *testing.T) {
	tests := []struct {
		code     ErrorCode
		contains string
	}{
		{ErrCodePortNotFound, "connected"},
		{ErrCodePortBusy, "in use"},
		{ErrCodePortAccessDenied, "dialout"},
		{ErrCodePortDisconnected, "disconnected"},
		{ErrCodeBaudRateMismatch, "baud"},
		{ErrCodeParityError, "parity"},
		{ErrCodeFrameError, "wiring"},
		{ErrCodeTCPConnFailed, "gateway"},
		{ErrCodeTCPConnLost, "network"},
		{ErrCodeTCPTimeout, "connectivity"},
		{ErrCodeChecksumMismatch, "corruption"},
		{ErrCodeInvalidFrame, "invalid"},
		{ErrCodeTimeout, "response"},
		{ErrCodeDeviceNotResponding, "power"},
		{ErrCodeDeviceError, "device"},
		{ErrCodeDeviceBusy, "retry"},
		{ErrCodeUnknown, "unexpected"},
	}

	for _, tt := range tests {
		t.Run(tt.code.String(), func(t *testing.T) {
			suggestion := tt.code.Suggestion()
			if !containsStr(suggestion, tt.contains) {
				t.Errorf("Suggestion for %v should contain '%v', got '%v'",
					tt.code, tt.contains, suggestion)
			}
		})
	}
}

func TestHealthEventType(t *testing.T) {
	tests := []struct {
		eventType HealthEventType
		expected  string
	}{
		{HealthEventStarted, "started"},
		{HealthEventStopped, "stopped"},
		{HealthEventConnected, "connected"},
		{HealthEventDisconnected, "disconnected"},
		{HealthEventRecoveryStarted, "recovery_started"},
		{HealthEventRecoverySuccess, "recovery_success"},
		{HealthEventRecoveryFailed, "recovery_failed"},
		{HealthEventDeviceOnline, "device_online"},
		{HealthEventDeviceOffline, "device_offline"},
		{HealthEventConfigReloaded, "config_reloaded"},
		{HealthEventWarning, "warning"},
	}

	for _, tt := range tests {
		t.Run(string(tt.eventType), func(t *testing.T) {
			if string(tt.eventType) != tt.expected {
				t.Errorf("HealthEventType = %v, want %v", tt.eventType, tt.expected)
			}
		})
	}
}

func TestManagerState(t *testing.T) {
	tests := []struct {
		state    ManagerState
		expected string
	}{
		{StateIdle, "idle"},
		{StateStarting, "starting"},
		{StateRunning, "running"},
		{StateStopping, "stopping"},
		{StateStopped, "stopped"},
		{StateRecovering, "recovering"},
		{StateError, "error"},
	}

	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			if string(tt.state) != tt.expected {
				t.Errorf("ManagerState = %v, want %v", tt.state, tt.expected)
			}
		})
	}
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
