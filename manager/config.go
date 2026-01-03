// Package manager provides a complete M-Bus communication manager with
// self-healing capabilities and event-driven data delivery.
package manager

import (
	"fmt"
	"os"
	"time"

	"github.com/pdat-cz/go-mbus/manager/transport"
	"gopkg.in/yaml.v3"
)

// Config contains the complete manager configuration.
type Config struct {
	// Transport configuration
	Transport TransportConfig `yaml:"transport" json:"transport"`

	// Polling configuration
	Polling PollingConfig `yaml:"polling" json:"polling"`

	// Devices to monitor
	Devices []DeviceConfig `yaml:"devices" json:"devices"`

	// SelfHealing configuration
	SelfHealing SelfHealingConfig `yaml:"self_healing" json:"self_healing"`
}

// TransportConfig configures the communication transport.
type TransportConfig struct {
	// Type: "serial" or "tcp"
	Type string `yaml:"type" json:"type"`

	// Serial configuration (when Type is "serial")
	Serial *transport.SerialConfig `yaml:"serial,omitempty" json:"serial,omitempty"`

	// TCP configuration (when Type is "tcp")
	TCP *transport.TCPConfig `yaml:"tcp,omitempty" json:"tcp,omitempty"`
}

// PollingConfig configures the polling behavior.
type PollingConfig struct {
	// Interval between polling cycles
	Interval time.Duration `yaml:"interval" json:"interval"`

	// StaggerDelay between device reads to avoid bus congestion
	StaggerDelay time.Duration `yaml:"stagger_delay" json:"stagger_delay"`

	// RetryCount for failed device reads
	RetryCount int `yaml:"retry_count" json:"retry_count"`

	// RetryDelay between retry attempts
	RetryDelay time.Duration `yaml:"retry_delay" json:"retry_delay"`

	// Timeout for individual device reads
	Timeout time.Duration `yaml:"timeout" json:"timeout"`
}

// DeviceConfig configures an individual M-Bus device.
type DeviceConfig struct {
	// Address is the M-Bus primary address (1-250)
	Address int `yaml:"address" json:"address"`

	// Name is a human-readable device name
	Name string `yaml:"name" json:"name"`

	// Interval overrides the global polling interval for this device
	Interval *time.Duration `yaml:"interval,omitempty" json:"interval,omitempty"`

	// Enabled indicates if the device should be polled
	Enabled bool `yaml:"enabled" json:"enabled"`
}

// SelfHealingConfig configures the self-healing behavior.
type SelfHealingConfig struct {
	// Enabled activates self-healing
	Enabled bool `yaml:"enabled" json:"enabled"`

	// MaxRecoveryAttempts before giving up
	MaxRecoveryAttempts int `yaml:"max_recovery_attempts" json:"max_recovery_attempts"`

	// RecoveryDelay between recovery attempts
	RecoveryDelay time.Duration `yaml:"recovery_delay" json:"recovery_delay"`

	// AutoBaudDetection enables automatic baud rate detection
	AutoBaudDetection bool `yaml:"auto_baud_detection" json:"auto_baud_detection"`

	// AutoPortScan enables scanning for alternative ports
	AutoPortScan bool `yaml:"auto_port_scan" json:"auto_port_scan"`

	// HealthCheckInterval for periodic health checks
	HealthCheckInterval time.Duration `yaml:"health_check_interval" json:"health_check_interval"`
}

// DefaultConfig returns a configuration with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Transport: TransportConfig{
			Type:   "serial",
			Serial: ptrSerialConfig(transport.DefaultSerialConfig()),
		},
		Polling: PollingConfig{
			Interval:     30 * time.Second,
			StaggerDelay: 500 * time.Millisecond,
			RetryCount:   3,
			RetryDelay:   1 * time.Second,
			Timeout:      5 * time.Second,
		},
		Devices: []DeviceConfig{},
		SelfHealing: SelfHealingConfig{
			Enabled:             true,
			MaxRecoveryAttempts: 5,
			RecoveryDelay:       5 * time.Second,
			AutoBaudDetection:   true,
			AutoPortScan:        true,
			HealthCheckInterval: 60 * time.Second,
		},
	}
}

// LoadConfig loads configuration from a YAML file.
func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config file: %w", err)
	}

	return ParseConfig(data)
}

// ParseConfig parses configuration from YAML data.
func ParseConfig(data []byte) (Config, error) {
	config := DefaultConfig()

	if err := yaml.Unmarshal(data, &config); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}

	if err := config.Validate(); err != nil {
		return Config{}, fmt.Errorf("validate config: %w", err)
	}

	return config, nil
}

// Validate checks the configuration for errors.
func (c *Config) Validate() error {
	// Validate transport
	switch c.Transport.Type {
	case "serial":
		if c.Transport.Serial == nil {
			return fmt.Errorf("serial configuration required when type is 'serial'")
		}
		if c.Transport.Serial.Port == "" {
			return fmt.Errorf("serial port path is required")
		}
	case "tcp":
		if c.Transport.TCP == nil {
			return fmt.Errorf("TCP configuration required when type is 'tcp'")
		}
		if c.Transport.TCP.Host == "" {
			return fmt.Errorf("TCP host is required")
		}
		if c.Transport.TCP.Port <= 0 || c.Transport.TCP.Port > 65535 {
			return fmt.Errorf("TCP port must be between 1 and 65535")
		}
	default:
		return fmt.Errorf("invalid transport type: %s (must be 'serial' or 'tcp')", c.Transport.Type)
	}

	// Validate polling
	if c.Polling.Interval <= 0 {
		return fmt.Errorf("polling interval must be positive")
	}
	if c.Polling.RetryCount < 0 {
		return fmt.Errorf("retry count cannot be negative")
	}

	// Validate devices
	addressSeen := make(map[int]bool)
	for i, device := range c.Devices {
		if device.Address < 1 || device.Address > 250 {
			return fmt.Errorf("device %d: address must be between 1 and 250", i)
		}
		if addressSeen[device.Address] {
			return fmt.Errorf("device %d: duplicate address %d", i, device.Address)
		}
		addressSeen[device.Address] = true
	}

	// Validate self-healing
	if c.SelfHealing.Enabled {
		if c.SelfHealing.MaxRecoveryAttempts <= 0 {
			return fmt.Errorf("max recovery attempts must be positive")
		}
	}

	return nil
}

// SaveConfig saves configuration to a YAML file.
func (c *Config) SaveConfig(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	return nil
}

// GetDeviceByAddress returns the device configuration for an address.
func (c *Config) GetDeviceByAddress(address int) *DeviceConfig {
	for i := range c.Devices {
		if c.Devices[i].Address == address {
			return &c.Devices[i]
		}
	}
	return nil
}

// GetEnabledDevices returns all enabled device configurations.
func (c *Config) GetEnabledDevices() []DeviceConfig {
	var devices []DeviceConfig
	for _, device := range c.Devices {
		if device.Enabled {
			devices = append(devices, device)
		}
	}
	return devices
}

func ptrSerialConfig(c transport.SerialConfig) *transport.SerialConfig {
	return &c
}
