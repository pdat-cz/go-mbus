package manager

import (
	"testing"
	"time"

	"github.com/pdat-cz/go-mbus/manager/transport"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	// Check transport defaults
	if config.Transport.Type != "serial" {
		t.Errorf("Default transport type = %v, want 'serial'", config.Transport.Type)
	}
	if config.Transport.Serial == nil {
		t.Error("Default serial config should not be nil")
	}

	// Check polling defaults
	if config.Polling.Interval != 30*time.Second {
		t.Errorf("Default polling interval = %v, want 30s", config.Polling.Interval)
	}
	if config.Polling.RetryCount != 3 {
		t.Errorf("Default retry count = %v, want 3", config.Polling.RetryCount)
	}

	// Check self-healing defaults
	if !config.SelfHealing.Enabled {
		t.Error("Self-healing should be enabled by default")
	}
	if config.SelfHealing.MaxRecoveryAttempts != 5 {
		t.Errorf("Default max recovery attempts = %v, want 5", config.SelfHealing.MaxRecoveryAttempts)
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		wantError bool
	}{
		{
			name: "valid serial config",
			config: Config{
				Transport: TransportConfig{
					Type:   "serial",
					Serial: &transport.SerialConfig{Port: "/dev/ttyUSB0"},
				},
				Polling: PollingConfig{
					Interval:   30 * time.Second,
					RetryCount: 3,
				},
				SelfHealing: SelfHealingConfig{
					Enabled:             true,
					MaxRecoveryAttempts: 5,
				},
			},
			wantError: false,
		},
		{
			name: "valid TCP config",
			config: Config{
				Transport: TransportConfig{
					Type: "tcp",
					TCP:  &transport.TCPConfig{Host: "192.168.1.100", Port: 5000},
				},
				Polling: PollingConfig{
					Interval:   30 * time.Second,
					RetryCount: 3,
				},
				SelfHealing: SelfHealingConfig{
					Enabled:             true,
					MaxRecoveryAttempts: 5,
				},
			},
			wantError: false,
		},
		{
			name: "invalid transport type",
			config: Config{
				Transport: TransportConfig{
					Type: "invalid",
				},
				Polling: PollingConfig{Interval: 30 * time.Second},
			},
			wantError: true,
		},
		{
			name: "serial without config",
			config: Config{
				Transport: TransportConfig{
					Type: "serial",
				},
				Polling: PollingConfig{Interval: 30 * time.Second},
			},
			wantError: true,
		},
		{
			name: "serial without port",
			config: Config{
				Transport: TransportConfig{
					Type:   "serial",
					Serial: &transport.SerialConfig{},
				},
				Polling: PollingConfig{Interval: 30 * time.Second},
			},
			wantError: true,
		},
		{
			name: "TCP without config",
			config: Config{
				Transport: TransportConfig{
					Type: "tcp",
				},
				Polling: PollingConfig{Interval: 30 * time.Second},
			},
			wantError: true,
		},
		{
			name: "TCP without host",
			config: Config{
				Transport: TransportConfig{
					Type: "tcp",
					TCP:  &transport.TCPConfig{Port: 5000},
				},
				Polling: PollingConfig{Interval: 30 * time.Second},
			},
			wantError: true,
		},
		{
			name: "TCP invalid port",
			config: Config{
				Transport: TransportConfig{
					Type: "tcp",
					TCP:  &transport.TCPConfig{Host: "192.168.1.100", Port: 0},
				},
				Polling: PollingConfig{Interval: 30 * time.Second},
			},
			wantError: true,
		},
		{
			name: "zero polling interval",
			config: Config{
				Transport: TransportConfig{
					Type:   "serial",
					Serial: &transport.SerialConfig{Port: "/dev/ttyUSB0"},
				},
				Polling: PollingConfig{Interval: 0},
			},
			wantError: true,
		},
		{
			name: "negative retry count",
			config: Config{
				Transport: TransportConfig{
					Type:   "serial",
					Serial: &transport.SerialConfig{Port: "/dev/ttyUSB0"},
				},
				Polling: PollingConfig{Interval: 30 * time.Second, RetryCount: -1},
			},
			wantError: true,
		},
		{
			name: "invalid device address low",
			config: Config{
				Transport: TransportConfig{
					Type:   "serial",
					Serial: &transport.SerialConfig{Port: "/dev/ttyUSB0"},
				},
				Polling: PollingConfig{Interval: 30 * time.Second},
				Devices: []DeviceConfig{{Address: 0, Enabled: true}},
			},
			wantError: true,
		},
		{
			name: "invalid device address high",
			config: Config{
				Transport: TransportConfig{
					Type:   "serial",
					Serial: &transport.SerialConfig{Port: "/dev/ttyUSB0"},
				},
				Polling: PollingConfig{Interval: 30 * time.Second},
				Devices: []DeviceConfig{{Address: 251, Enabled: true}},
			},
			wantError: true,
		},
		{
			name: "duplicate device addresses",
			config: Config{
				Transport: TransportConfig{
					Type:   "serial",
					Serial: &transport.SerialConfig{Port: "/dev/ttyUSB0"},
				},
				Polling: PollingConfig{Interval: 30 * time.Second},
				Devices: []DeviceConfig{
					{Address: 1, Enabled: true},
					{Address: 1, Enabled: true},
				},
			},
			wantError: true,
		},
		{
			name: "self-healing with zero attempts",
			config: Config{
				Transport: TransportConfig{
					Type:   "serial",
					Serial: &transport.SerialConfig{Port: "/dev/ttyUSB0"},
				},
				Polling: PollingConfig{Interval: 30 * time.Second},
				SelfHealing: SelfHealingConfig{
					Enabled:             true,
					MaxRecoveryAttempts: 0,
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestParseConfig(t *testing.T) {
	yamlData := []byte(`
transport:
  type: serial
  serial:
    port: /dev/ttyUSB0
    baud_rate: 2400

polling:
  interval: 30s
  retry_count: 3

devices:
  - address: 1
    name: "Meter 1"
    enabled: true
  - address: 2
    name: "Meter 2"
    enabled: true

self_healing:
  enabled: true
  max_recovery_attempts: 5
`)

	config, err := ParseConfig(yamlData)
	if err != nil {
		t.Fatalf("ParseConfig failed: %v", err)
	}

	if config.Transport.Type != "serial" {
		t.Errorf("Transport type = %v, want 'serial'", config.Transport.Type)
	}
	if config.Transport.Serial.Port != "/dev/ttyUSB0" {
		t.Errorf("Serial port = %v, want '/dev/ttyUSB0'", config.Transport.Serial.Port)
	}
	if len(config.Devices) != 2 {
		t.Errorf("Devices count = %v, want 2", len(config.Devices))
	}
	if config.Devices[0].Name != "Meter 1" {
		t.Errorf("Device 0 name = %v, want 'Meter 1'", config.Devices[0].Name)
	}
}

func TestParseConfigInvalid(t *testing.T) {
	yamlData := []byte(`
transport:
  type: invalid
`)

	_, err := ParseConfig(yamlData)
	if err == nil {
		t.Error("ParseConfig should fail for invalid config")
	}
}

func TestGetDeviceByAddress(t *testing.T) {
	config := DefaultConfig()
	config.Devices = []DeviceConfig{
		{Address: 1, Name: "Meter 1", Enabled: true},
		{Address: 5, Name: "Meter 5", Enabled: true},
	}

	// Find existing device
	dev := config.GetDeviceByAddress(1)
	if dev == nil {
		t.Fatal("GetDeviceByAddress(1) should return device")
	}
	if dev.Name != "Meter 1" {
		t.Errorf("Device name = %v, want 'Meter 1'", dev.Name)
	}

	// Find non-existing device
	dev = config.GetDeviceByAddress(99)
	if dev != nil {
		t.Error("GetDeviceByAddress(99) should return nil")
	}
}

func TestGetEnabledDevices(t *testing.T) {
	config := DefaultConfig()
	config.Devices = []DeviceConfig{
		{Address: 1, Name: "Meter 1", Enabled: true},
		{Address: 2, Name: "Meter 2", Enabled: false},
		{Address: 3, Name: "Meter 3", Enabled: true},
	}

	enabled := config.GetEnabledDevices()
	if len(enabled) != 2 {
		t.Errorf("Enabled devices count = %v, want 2", len(enabled))
	}

	// Check that disabled device is not included
	for _, dev := range enabled {
		if dev.Address == 2 {
			t.Error("Disabled device should not be in enabled list")
		}
	}
}
