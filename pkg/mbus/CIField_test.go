package mbus

import (
	"testing"
)

func TestNewCIField(t *testing.T) {
	tests := []struct {
		name     string
		input    byte
		expected CIField
	}{
		{
			name:     "Data Send",
			input:    0x51,
			expected: CiFieldDataSend,
		},
		{
			name:     "Selection of Slaves",
			input:    0x52,
			expected: CiFieldSelectionOfSlaves,
		},
		{
			name:     "Application Reset",
			input:    0x50,
			expected: CiFieldApplicationReset,
		},
		{
			name:     "Synchronize Action",
			input:    0x54,
			expected: CiFieldSynchronizeAction,
		},
		{
			name:     "Variable Data Structure 72",
			input:    0x72,
			expected: CiFieldVariable72,
		},
		{
			name:     "Variable Data Structure 76",
			input:    0x76,
			expected: CiFieldVariable76,
		},
		{
			name:     "Baudrate 300",
			input:    0xB8,
			expected: CiFieldBaudrate300,
		},
		{
			name:     "Baudrate 1200",
			input:    0xBA,
			expected: CiFieldBaudrate1200,
		},
		{
			name:     "Baudrate 2400",
			input:    0xBB,
			expected: CiFieldBaudrate2400,
		},
		{
			name:     "Baudrate 4800",
			input:    0xBC,
			expected: CiFieldBaudrate4800,
		},
		{
			name:     "Baudrate 9600",
			input:    0xBD,
			expected: CiFieldBaudrate9600,
		},
		{
			name:     "Baudrate 19200",
			input:    0xBE,
			expected: CiFieldBaudrate19200,
		},
		{
			name:     "Baudrate 38400",
			input:    0xBF,
			expected: CiFieldBaudrate38400,
		},
		{
			name:     "Request Readout Complete RAM",
			input:    0xB1,
			expected: CiFieldRequestReadoutCompleteRam,
		},
		{
			name:     "Send User Data Not Standardized RAM Write",
			input:    0xB2,
			expected: CiFieldSendUserDataNotStandardizedRamWrite,
		},
		{
			name:     "Initialize Test Calibration Mode",
			input:    0xB3,
			expected: CiFieldInitializeTestCalibrationMode,
		},
		{
			name:     "EEPROM Read",
			input:    0xB4,
			expected: CiFieldEepromRead,
		},
		{
			name:     "Start Software Test",
			input:    0xB6,
			expected: CiFieldStartSoftwareTest,
		},
		{
			name:     "Codes Used For Hashing 0",
			input:    0x90,
			expected: CiFieldCodesUsedForHashing0,
		},
		{
			name:     "Codes Used For Hashing 1",
			input:    0x91,
			expected: CiFieldCodesUsedForHashing1,
		},
		{
			name:     "Codes Used For Hashing 2",
			input:    0x92,
			expected: CiFieldCodesUsedForHashing2,
		},
		{
			name:     "Codes Used For Hashing 3",
			input:    0x93,
			expected: CiFieldCodesUsedForHashing3,
		},
		{
			name:     "Codes Used For Hashing 4",
			input:    0x94,
			expected: CiFieldCodesUsedForHashing4,
		},
		{
			name:     "Codes Used For Hashing 5",
			input:    0x95,
			expected: CiFieldCodesUsedForHashing5,
		},
		{
			name:     "Codes Used For Hashing 6",
			input:    0x96,
			expected: CiFieldCodesUsedForHashing6,
		},
		{
			name:     "Codes Used For Hashing 7",
			input:    0x97,
			expected: CiFieldCodesUsedForHashing7,
		},
		{
			name:     "Undefined CI Field",
			input:    0x00,
			expected: CIField(0x00),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewCIField(tt.input)
			if got != tt.expected {
				t.Errorf("NewCIField() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCIField_String(t *testing.T) {
	tests := []struct {
		name     string
		cifield  CIField
		expected string
	}{
		{
			name:     "Data Send",
			cifield:  CiFieldDataSend,
			expected: "DATA_SEND",
		},
		{
			name:     "Selection of Slaves",
			cifield:  CiFieldSelectionOfSlaves,
			expected: "SELECTION_OF_SLAVES",
		},
		{
			name:     "Application Reset",
			cifield:  CiFieldApplicationReset,
			expected: "APPLICATION_RESET",
		},
		{
			name:     "Synchronize Action",
			cifield:  CiFieldSynchronizeAction,
			expected: "SYNCHRONIZE_ACTION",
		},
		{
			name:     "Variable Data Structure 72",
			cifield:  CiFieldVariable72,
			expected: "VARIABLE_DATA_STRUCTURE_72",
		},
		{
			name:     "Variable Data Structure 76",
			cifield:  CiFieldVariable76,
			expected: "VARIABLE_DATA_STRUCTURE_76",
		},
		{
			name:     "Baudrate 300",
			cifield:  CiFieldBaudrate300,
			expected: "BAUDRATE_300",
		},
		{
			name:     "Baudrate 1200",
			cifield:  CiFieldBaudrate1200,
			expected: "BAUDRATE_1200",
		},
		{
			name:     "Baudrate 2400",
			cifield:  CiFieldBaudrate2400,
			expected: "BAUDRATE_2400",
		},
		{
			name:     "Baudrate 4800",
			cifield:  CiFieldBaudrate4800,
			expected: "BAUDRATE_4800",
		},
		{
			name:     "Baudrate 9600",
			cifield:  CiFieldBaudrate9600,
			expected: "BAUDRATE_9600",
		},
		{
			name:     "Baudrate 19200",
			cifield:  CiFieldBaudrate19200,
			expected: "BAUDRATE_19200",
		},
		{
			name:     "Baudrate 38400",
			cifield:  CiFieldBaudrate38400,
			expected: "BAUDRATE_38400",
		},
		{
			name:     "Request Readout Complete RAM",
			cifield:  CiFieldRequestReadoutCompleteRam,
			expected: "REQUEST_READOUT_COMPLETE_RAM",
		},
		{
			name:     "Send User Data Not Standardized RAM Write",
			cifield:  CiFieldSendUserDataNotStandardizedRamWrite,
			expected: "SEND_USER_DATA_NOT_STANDARDIZED_RAM_WRITE",
		},
		{
			name:     "Initialize Test Calibration Mode",
			cifield:  CiFieldInitializeTestCalibrationMode,
			expected: "INITIALIZE_TEST_CALIBRATION_MODE",
		},
		{
			name:     "EEPROM Read",
			cifield:  CiFieldEepromRead,
			expected: "EEPROM_READ",
		},
		{
			name:     "Start Software Test",
			cifield:  CiFieldStartSoftwareTest,
			expected: "START_SOFTWARE_TEST",
		},
		{
			name:     "Codes Used For Hashing 0",
			cifield:  CiFieldCodesUsedForHashing0,
			expected: "CODES_USED_FOR_HASHING_0",
		},
		{
			name:     "Codes Used For Hashing 1",
			cifield:  CiFieldCodesUsedForHashing1,
			expected: "CODES_USED_FOR_HASHING_1",
		},
		{
			name:     "Codes Used For Hashing 2",
			cifield:  CiFieldCodesUsedForHashing2,
			expected: "CODES_USED_FOR_HASHING_2",
		},
		{
			name:     "Codes Used For Hashing 3",
			cifield:  CiFieldCodesUsedForHashing3,
			expected: "CODES_USED_FOR_HASHING_3",
		},
		{
			name:     "Codes Used For Hashing 4",
			cifield:  CiFieldCodesUsedForHashing4,
			expected: "CODES_USED_FOR_HASHING_4",
		},
		{
			name:     "Codes Used For Hashing 5",
			cifield:  CiFieldCodesUsedForHashing5,
			expected: "CODES_USED_FOR_HASHING_5",
		},
		{
			name:     "Codes Used For Hashing 6",
			cifield:  CiFieldCodesUsedForHashing6,
			expected: "CODES_USED_FOR_HASHING_6",
		},
		{
			name:     "Codes Used For Hashing 7",
			cifield:  CiFieldCodesUsedForHashing7,
			expected: "CODES_USED_FOR_HASHING_7",
		},
		{
			name:     "Undefined CI Field",
			cifield:  CIField(0x00),
			expected: "UNDEFINED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cifield.String()
			if got != tt.expected {
				t.Errorf("String() = %v, want %v", got, tt.expected)
			}
		})
	}
}
