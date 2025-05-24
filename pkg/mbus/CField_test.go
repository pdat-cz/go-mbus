package mbus

import (
	"testing"
)

func TestNewCField(t *testing.T) {
	tests := []struct {
		name     string
		input    byte
		expected CField
	}{
		{
			name:  "Master to Slave (SND_NKE)",
			input: 0x40,
			expected: CField{
				position7: true,
				position8: false,
				FCB:       false,
				FCV:       false,
				F3:        false,
				F2:        false,
				F1:        false,
				F0:        false,
			},
		},
		{
			name:  "Master to Slave (SND_UD)",
			input: 0x53,
			expected: CField{
				position7: true,
				position8: false,
				FCB:       false,
				FCV:       true,
				F3:        false,
				F2:        false,
				F1:        true,
				F0:        true,
			},
		},
		{
			name:  "Slave to Master (RSP_UD)",
			input: 0x08,
			expected: CField{
				position7: false,
				position8: false,
				ACD:       false,
				DFC:       false,
				F3:        true,
				F2:        false,
				F1:        false,
				F0:        false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewCField(tt.input)

			if got.position7 != tt.expected.position7 {
				t.Errorf("position7 = %v, want %v", got.position7, tt.expected.position7)
			}
			if got.position8 != tt.expected.position8 {
				t.Errorf("position8 = %v, want %v", got.position8, tt.expected.position8)
			}

			// Check direction-specific fields
			if got.position7 { // Master to Slave
				if got.FCB != tt.expected.FCB {
					t.Errorf("FCB = %v, want %v", got.FCB, tt.expected.FCB)
				}
				if got.FCV != tt.expected.FCV {
					t.Errorf("FCV = %v, want %v", got.FCV, tt.expected.FCV)
				}
			} else { // Slave to Master
				if got.ACD != tt.expected.ACD {
					t.Errorf("ACD = %v, want %v", got.ACD, tt.expected.ACD)
				}
				if got.DFC != tt.expected.DFC {
					t.Errorf("DFC = %v, want %v", got.DFC, tt.expected.DFC)
				}
			}

			// Check function bits
			if got.F3 != tt.expected.F3 {
				t.Errorf("F3 = %v, want %v", got.F3, tt.expected.F3)
			}
			if got.F2 != tt.expected.F2 {
				t.Errorf("F2 = %v, want %v", got.F2, tt.expected.F2)
			}
			if got.F1 != tt.expected.F1 {
				t.Errorf("F1 = %v, want %v", got.F1, tt.expected.F1)
			}
			if got.F0 != tt.expected.F0 {
				t.Errorf("F0 = %v, want %v", got.F0, tt.expected.F0)
			}
		})
	}
}

func TestCField_getByte(t *testing.T) {
	tests := []struct {
		name     string
		cfield   CField
		expected byte
	}{
		{
			name: "Master to Slave (SND_NKE)",
			cfield: CField{
				position7: true,
				position8: false,
				FCB:       false,
				FCV:       false,
				F3:        false,
				F2:        false,
				F1:        false,
				F0:        false,
			},
			expected: 0x40,
		},
		{
			name: "Master to Slave (SND_UD)",
			cfield: CField{
				position7: true,
				position8: false,
				FCB:       false,
				FCV:       true,
				F3:        false,
				F2:        false,
				F1:        true,
				F0:        true,
			},
			expected: 0x53,
		},
		{
			name: "Slave to Master (RSP_UD)",
			cfield: CField{
				position7: false,
				position8: false,
				ACD:       false,
				DFC:       false,
				F3:        false,
				F2:        false,
				F1:        false,
				F0:        false,
			},
			expected: 0x00,
		},
		{
			name: "Slave to Master with ACD set",
			cfield: CField{
				position7: false,
				position8: false,
				ACD:       true,
				DFC:       false,
				F3:        false,
				F2:        false,
				F1:        false,
				F0:        false,
			},
			expected: 0x20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cfield.getByte()
			if got != tt.expected {
				t.Errorf("getByte() = %#02x, want %#02x", got, tt.expected)
			}
		})
	}
}

func TestCField_IsFromSlaveAndMaster(t *testing.T) {
	tests := []struct {
		name       string
		input      byte
		fromSlave  bool
		fromMaster bool
	}{
		{
			name:       "Master to Slave (SND_NKE)",
			input:      0x40,
			fromSlave:  false,
			fromMaster: true,
		},
		{
			name:       "Master to Slave (SND_UD)",
			input:      0x53,
			fromSlave:  false,
			fromMaster: true,
		},
		{
			name:       "Slave to Master (RSP_UD)",
			input:      0x08,
			fromSlave:  true,
			fromMaster: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cf := NewCField(tt.input)

			if got := cf.IsFromSlave(); got != tt.fromSlave {
				t.Errorf("IsFromSlave() = %v, want %v", got, tt.fromSlave)
			}

			if got := cf.IsFromMaster(); got != tt.fromMaster {
				t.Errorf("IsFromMaster() = %v, want %v", got, tt.fromMaster)
			}
		})
	}
}

func TestCFieldConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant CField
		expected byte
	}{
		{
			name:     "CFIELD_SND_NKE",
			constant: CFIELD_SND_NKE,
			expected: 0x40,
		},
		{
			name:     "CFIELD_SND_UD_0",
			constant: CFIELD_SND_UD_0,
			expected: 0x53,
		},
		{
			name:     "CFIELD_SND_UD_1",
			constant: CFIELD_SND_UD_1,
			expected: 0x73,
		},
		{
			name:     "CFIELD_REQ_UD2_0",
			constant: CFIELD_REQ_UD2_0,
			expected: 0x5B,
		},
		{
			name:     "CFIELD_REQ_UD2_1",
			constant: CFIELD_REQ_UD2_1,
			expected: 0x7B,
		},
		{
			name:     "CFIELD_REQ_UD1_0",
			constant: CFIELD_REQ_UD1_0,
			expected: 0x5A,
		},
		{
			name:     "CFIELD_REQ_UD1_1",
			constant: CFIELD_REQ_UD1_1,
			expected: 0x7A,
		},
		{
			name:     "CFIELD_RSP_UD_a",
			constant: CFIELD_RSP_UD_a,
			expected: 0x08,
		},
		{
			name:     "CFIELD_RSP_UD_b",
			constant: CFIELD_RSP_UD_b,
			expected: 0x18,
		},
		{
			name:     "CFIELD_RSP_UD_c",
			constant: CFIELD_RSP_UD_c,
			expected: 0x28,
		},
		{
			name:     "CFIELD_RSP_UD_d",
			constant: CFIELD_RSP_UD_d,
			expected: 0x38,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.constant.getByte()
			if got != tt.expected {
				t.Errorf("%s.getByte() = %#02x, want %#02x", tt.name, got, tt.expected)
			}
		})
	}
}
