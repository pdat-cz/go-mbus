package mbus

import (
	"testing"
)

func TestNewAField(t *testing.T) {
	tests := []struct {
		name     string
		input    byte
		expected AField
	}{
		{
			name:     "Valid slave address",
			input:    10,
			expected: AField(10),
		},
		{
			name:     "Unconfigured slave address",
			input:    0,
			expected: AField(0),
		},
		{
			name:     "Broadcast address with slave reply",
			input:    254,
			expected: AField(254),
		},
		{
			name:     "Broadcast address",
			input:    255,
			expected: AField(255),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewAField(tt.input)
			if got != tt.expected {
				t.Errorf("NewAField() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAField_String(t *testing.T) {
	tests := []struct {
		name     string
		afield   AField
		expected string
	}{
		{
			name:     "Valid slave address",
			afield:   AField(10),
			expected: "0x0A",
		},
		{
			name:     "Unconfigured slave address",
			afield:   AField(0),
			expected: "0x00",
		},
		{
			name:     "Broadcast address with slave reply",
			afield:   AField(254),
			expected: "0xFE",
		},
		{
			name:     "Broadcast address",
			afield:   AField(255),
			expected: "0xFF",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &tt.afield
			got := a.String()
			if got != tt.expected {
				t.Errorf("String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAField_IsSlaveAddress(t *testing.T) {
	tests := []struct {
		name     string
		afield   AField
		expected bool
	}{
		{
			name:     "Valid slave address (lower bound)",
			afield:   AField(1),
			expected: true,
		},
		{
			name:     "Valid slave address (middle)",
			afield:   AField(125),
			expected: true,
		},
		{
			name:     "Valid slave address (upper bound)",
			afield:   AField(250),
			expected: true,
		},
		{
			name:     "Unconfigured slave address",
			afield:   AField(0),
			expected: false,
		},
		{
			name:     "Broadcast address with slave reply",
			afield:   AField(254),
			expected: false,
		},
		{
			name:     "Broadcast address",
			afield:   AField(255),
			expected: false,
		},
		{
			name:     "Invalid address (251)",
			afield:   AField(251),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &tt.afield
			got := a.IsSlaveAddress()
			if got != tt.expected {
				t.Errorf("IsSlaveAddress() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAField_SlaveAddress(t *testing.T) {
	tests := []struct {
		name        string
		afield      AField
		expected    uint8
		expectError bool
	}{
		{
			name:        "Valid slave address",
			afield:      AField(10),
			expected:    10,
			expectError: false,
		},
		{
			name:        "Unconfigured slave address",
			afield:      AField(0),
			expected:    0,
			expectError: true,
		},
		{
			name:        "Broadcast address with slave reply",
			afield:      AField(254),
			expected:    0,
			expectError: true,
		},
		{
			name:        "Broadcast address",
			afield:      AField(255),
			expected:    0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &tt.afield
			got, err := a.SlaveAddress()

			if tt.expectError {
				if err == nil {
					t.Errorf("SlaveAddress() expected error, got nil")
				}
				if err != ErrInvalidSlaveAddress {
					t.Errorf("SlaveAddress() expected ErrInvalidSlaveAddress, got %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("SlaveAddress() unexpected error: %v", err)
				}
				if got != tt.expected {
					t.Errorf("SlaveAddress() = %v, want %v", got, tt.expected)
				}
			}
		})
	}
}

func TestAField_IsUnconfiguredSlave(t *testing.T) {
	tests := []struct {
		name     string
		afield   AField
		expected bool
	}{
		{
			name:     "Unconfigured slave address",
			afield:   AField(0),
			expected: true,
		},
		{
			name:     "Valid slave address",
			afield:   AField(10),
			expected: false,
		},
		{
			name:     "Broadcast address with slave reply",
			afield:   AField(254),
			expected: false,
		},
		{
			name:     "Broadcast address",
			afield:   AField(255),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &tt.afield
			got := a.IsUnconfiguredSlave()
			if got != tt.expected {
				t.Errorf("IsUnconfiguredSlave() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAField_IsBroadcastAddressWithSlaveReplyAddress(t *testing.T) {
	tests := []struct {
		name     string
		afield   AField
		expected bool
	}{
		{
			name:     "Broadcast address with slave reply",
			afield:   AFieldBroadcastAddressWithSlaveReplyAddress,
			expected: true,
		},
		{
			name:     "Valid slave address",
			afield:   AField(10),
			expected: false,
		},
		{
			name:     "Unconfigured slave address",
			afield:   AField(0),
			expected: false,
		},
		{
			name:     "Broadcast address",
			afield:   AField(255),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &tt.afield
			got := a.IsBroadcastAddressWithSlaveReplyAddress()
			if got != tt.expected {
				t.Errorf("IsBroadcastAddressWithSlaveReplyAddress() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAField_IsBroadcastAddress(t *testing.T) {
	tests := []struct {
		name     string
		afield   AField
		expected bool
	}{
		{
			name:     "Broadcast address",
			afield:   AFieldBroadcastAddress,
			expected: true,
		},
		{
			name:     "Valid slave address",
			afield:   AField(10),
			expected: false,
		},
		{
			name:     "Unconfigured slave address",
			afield:   AField(0),
			expected: false,
		},
		{
			name:     "Broadcast address with slave reply",
			afield:   AField(254),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &tt.afield
			got := a.IsBroadcastAddress()
			if got != tt.expected {
				t.Errorf("IsBroadcastAddress() = %v, want %v", got, tt.expected)
			}
		})
	}
}
