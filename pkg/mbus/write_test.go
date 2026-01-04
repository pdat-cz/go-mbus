package mbus

import (
	"bytes"
	"testing"
)

// =============================================================================
// Frame Builder Tests
// =============================================================================

// TestBuildLongFrame verifies long frame construction per EN 13757-2
func TestBuildLongFrame(t *testing.T) {
	tests := []struct {
		name     string
		cField   byte
		aField   byte
		ciField  byte
		data     []byte
		expected []byte
	}{
		{
			name:    "Control frame (no data)",
			cField:  0x53, // SND_UD
			aField:  0xFE, // Broadcast
			ciField: 0xBD, // 9600 baud
			data:    nil,
			// 68 03 03 68 | 53 FE BD | checksum 16
			// checksum = 0x53 + 0xFE + 0xBD = 0x0E (truncated to byte)
			expected: []byte{0x68, 0x03, 0x03, 0x68, 0x53, 0xFE, 0xBD, 0x0E, 0x16},
		},
		{
			name:    "Set primary address to 8",
			cField:  0x53, // SND_UD
			aField:  0xFE, // Broadcast
			ciField: 0x51, // DATA_SEND
			data:    []byte{0x01, 0x7A, 0x08},
			// 68 06 06 68 | 53 FE 51 | 01 7A 08 | checksum 16
			// checksum = 0x53 + 0xFE + 0x51 + 0x01 + 0x7A + 0x08 = 0x25
			expected: []byte{0x68, 0x06, 0x06, 0x68, 0x53, 0xFE, 0x51, 0x01, 0x7A, 0x08, 0x25, 0x16},
		},
		{
			name:    "Application reset",
			cField:  0x53, // SND_UD
			aField:  0xFE, // Broadcast
			ciField: 0x50, // APPLICATION_RESET
			data:    []byte{0x10},
			// 68 04 04 68 | 53 FE 50 | 10 | checksum 16
			// checksum = 0x53 + 0xFE + 0x50 + 0x10 = 0xB1
			expected: []byte{0x68, 0x04, 0x04, 0x68, 0x53, 0xFE, 0x50, 0x10, 0xB1, 0x16},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildLongFrame(tt.cField, tt.aField, tt.ciField, tt.data)
			if !bytes.Equal(got, tt.expected) {
				t.Errorf("BuildLongFrame() = %X, want %X", got, tt.expected)
			}
		})
	}
}

// TestBuildControlFrame verifies control frame construction
func TestBuildControlFrame(t *testing.T) {
	// Set baudrate to 9600 baud
	got := BuildControlFrame(0x53, 0xFE, 0xBD)
	expected := []byte{0x68, 0x03, 0x03, 0x68, 0x53, 0xFE, 0xBD, 0x0E, 0x16}
	if !bytes.Equal(got, expected) {
		t.Errorf("BuildControlFrame() = %X, want %X", got, expected)
	}
}

// =============================================================================
// Command Builder Tests - per protocol.md examples
// =============================================================================

// TestCOMMAND_SET_PRIMARY_ADDRESS verifies address setting command
func TestCOMMAND_SET_PRIMARY_ADDRESS(t *testing.T) {
	// From protocol.md: Set the slave to primary address 8
	// 68 06 06 68 | 53 FE 51 | 01 7A 08 | 25 16
	got := COMMAND_SET_PRIMARY_ADDRESS(8)
	expected := []byte{0x68, 0x06, 0x06, 0x68, 0x53, 0xFE, 0x51, 0x01, 0x7A, 0x08, 0x25, 0x16}
	if !bytes.Equal(got, expected) {
		t.Errorf("COMMAND_SET_PRIMARY_ADDRESS(8) = %X, want %X", got, expected)
	}
}

// TestCOMMAND_SET_IDENTIFICATION verifies identification setting command
func TestCOMMAND_SET_IDENTIFICATION(t *testing.T) {
	// From protocol.md: Set ID=01020304, Man=4024h (PAD), Gen=1, Med=4 (Heat)
	// 68 0D 0D 68 | 53 FE 51 | 07 79 04 03 02 01 24 40 01 04 | 95 16
	got := COMMAND_SET_IDENTIFICATION(0xFE, 0x01020304, 0x4024, 0x01, 0x04)
	expected := []byte{0x68, 0x0D, 0x0D, 0x68, 0x53, 0xFE, 0x51, 0x07, 0x79, 0x04, 0x03, 0x02, 0x01, 0x24, 0x40, 0x01, 0x04, 0x95, 0x16}
	if !bytes.Equal(got, expected) {
		t.Errorf("COMMAND_SET_IDENTIFICATION() = %X, want %X", got, expected)
	}
}

// TestCOMMAND_SET_BAUDRATE verifies baudrate setting command
func TestCOMMAND_SET_BAUDRATE(t *testing.T) {
	// From protocol.md: Set 9600 baud
	// 68 03 03 68 | 53 FE BD | 0E 16
	got := COMMAND_SET_BAUDRATE(0xFE, CiFieldBaudrate9600)
	expected := []byte{0x68, 0x03, 0x03, 0x68, 0x53, 0xFE, 0xBD, 0x0E, 0x16}
	if !bytes.Equal(got, expected) {
		t.Errorf("COMMAND_SET_BAUDRATE(9600) = %X, want %X", got, expected)
	}
}

// TestCOMMAND_APPLICATION_RESET verifies application reset command
func TestCOMMAND_APPLICATION_RESET(t *testing.T) {
	// From protocol.md: Enhanced application reset (subcode 0x10)
	// 68 04 04 68 | 53 FE 50 | 10 | B1 16
	got := COMMAND_APPLICATION_RESET(0xFE, 0x10)
	expected := []byte{0x68, 0x04, 0x04, 0x68, 0x53, 0xFE, 0x50, 0x10, 0xB1, 0x16}
	if !bytes.Equal(got, expected) {
		t.Errorf("COMMAND_APPLICATION_RESET(0x10) = %X, want %X", got, expected)
	}
}

// TestCOMMAND_SELECT_ALL_DATA verifies global readout request
func TestCOMMAND_SELECT_ALL_DATA(t *testing.T) {
	// From protocol.md: Request all data from address 3
	// 68 04 04 68 | 53 03 51 | 7F | 26 16
	got := COMMAND_SELECT_ALL_DATA(3)
	expected := []byte{0x68, 0x04, 0x04, 0x68, 0x53, 0x03, 0x51, 0x7F, 0x26, 0x16}
	if !bytes.Equal(got, expected) {
		t.Errorf("COMMAND_SELECT_ALL_DATA(3) = %X, want %X", got, expected)
	}
}

// TestCOMMAND_SELECT_DATA_RECORDS verifies data record selection
func TestCOMMAND_SELECT_DATA_RECORDS(t *testing.T) {
	// From protocol.md: Select volume (VIF=13h) and flow temp (VIF=5Ah) from address 7
	// 68 07 07 68 | 53 07 51 | 08 13 08 5A | 28 16
	records := []byte{0x08, 0x13, 0x08, 0x5A}
	got := COMMAND_SELECT_DATA_RECORDS(7, records)
	expected := []byte{0x68, 0x07, 0x07, 0x68, 0x53, 0x07, 0x51, 0x08, 0x13, 0x08, 0x5A, 0x28, 0x16}
	if !bytes.Equal(got, expected) {
		t.Errorf("COMMAND_SELECT_DATA_RECORDS() = %X, want %X", got, expected)
	}
}

// TestCOMMAND_SET_COUNTER verifies counter setting
func TestCOMMAND_SET_COUNTER(t *testing.T) {
	// From protocol.md: Set 8-digit BCD counter (1kWh) to 107 kWh at address 1
	// 68 0A 0A 68 | 53 01 51 | 0C 86 00 07 01 00 00 | 3F 16
	// Note: protocol example uses 0x86 for VIF with bit 7 set (storage 0)
	// We test the raw COMMAND_SET_COUNTER with explicit bytes
	value := []byte{0x00, 0x07, 0x01, 0x00, 0x00} // Storage 0 + BCD 107
	got := COMMAND_SET_COUNTER(1, 0x0C, 0x86, value)
	expected := []byte{0x68, 0x0A, 0x0A, 0x68, 0x53, 0x01, 0x51, 0x0C, 0x86, 0x00, 0x07, 0x01, 0x00, 0x00, 0x3F, 0x16}
	if !bytes.Equal(got, expected) {
		t.Errorf("COMMAND_SET_COUNTER() = %X, want %X", got, expected)
	}
}

// =============================================================================
// BCD Encoding Tests
// =============================================================================

// TestEncodeBCD verifies BCD encoding
func TestEncodeBCD(t *testing.T) {
	tests := []struct {
		value    uint32
		numBytes int
		expected []byte
	}{
		{0, 1, []byte{0x00}},
		{12, 1, []byte{0x12}},
		{99, 1, []byte{0x99}},
		{1234, 2, []byte{0x34, 0x12}},
		{107, 4, []byte{0x07, 0x01, 0x00, 0x00}},          // 107 as 8-digit BCD
		{12345678, 4, []byte{0x78, 0x56, 0x34, 0x12}},     // Full 8 digits
		{99999999, 4, []byte{0x99, 0x99, 0x99, 0x99}},     // Maximum 8-digit BCD
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := EncodeBCD(tt.value, tt.numBytes)
			if !bytes.Equal(got, tt.expected) {
				t.Errorf("EncodeBCD(%d, %d) = %X, want %X", tt.value, tt.numBytes, got, tt.expected)
			}
		})
	}
}

// TestEncodeBCD_RoundTrip tests encoding and decoding match
func TestEncodeBCD_RoundTrip(t *testing.T) {
	values := []uint32{0, 1, 10, 99, 100, 1234, 9999, 12345678}

	for _, v := range values {
		// Encode
		encoded := EncodeBCD(v, 4)

		// Decode (using existing decode logic from DIFField)
		decoded := decodeBCD(encoded)

		if decoded != int64(v) {
			t.Errorf("RoundTrip BCD: encoded %d, decoded %d", v, decoded)
		}
	}
}

// decodeBCD decodes BCD bytes to integer (for testing)
func decodeBCD(data []byte) int64 {
	var result int64
	multiplier := int64(1)
	for _, b := range data {
		low := int64(b & 0x0F)
		high := int64((b >> 4) & 0x0F)
		result += low * multiplier
		multiplier *= 10
		result += high * multiplier
		multiplier *= 10
	}
	return result
}

// =============================================================================
// Manufacturer ID Encoding Tests
// =============================================================================

// TestEncodeManufacturerId verifies manufacturer ID encoding
func TestEncodeManufacturerId(t *testing.T) {
	tests := []struct {
		id       string
		expected []byte
	}{
		{"PAD", []byte{0x24, 0x40}}, // (P-64)*1024 + (A-64)*32 + (D-64) = 16420 = 0x4024
		{"ACW", []byte{0x77, 0x04}}, // (A-64)*1024 + (C-64)*32 + (W-64) = 1143 = 0x0477
		{"KAM", []byte{0x2D, 0x2C}}, // (K-64)*1024 + (A-64)*32 + (M-64) = 11309 = 0x2C2D
		{"SLB", []byte{0x82, 0x4D}}, // (S-64)*1024 + (L-64)*32 + (B-64) = 19842 = 0x4D82
		{"ABB", []byte{0x42, 0x04}}, // (A-64)*1024 + (B-64)*32 + (B-64) = 1090 = 0x0442
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			got := EncodeManufacturerId(tt.id)
			if !bytes.Equal(got, tt.expected) {
				t.Errorf("EncodeManufacturerId(%q) = %X, want %X", tt.id, got, tt.expected)
			}
		})
	}
}

// TestManufacturerIdRoundTrip tests encoding and decoding match
func TestManufacturerIdRoundTrip(t *testing.T) {
	manufacturers := []string{"PAD", "ACW", "KAM", "SLB", "ABB", "ABC", "ZZZ"}

	for _, m := range manufacturers {
		encoded := EncodeManufacturerId(m)
		decoded := DecodeManufacturerId(encoded)
		if decoded != m {
			t.Errorf("RoundTrip Manufacturer: encoded %q, decoded %q", m, decoded)
		}
	}
}

// =============================================================================
// Frame Structure Validation Tests
// =============================================================================

// TestWriteLongFrameStructure validates the long frame format per EN 13757-2
func TestWriteLongFrameStructure(t *testing.T) {
	frame := BuildLongFrame(0x53, 0xFE, 0x51, []byte{0x01, 0x7A, 0x08})

	// Check start byte
	if frame[0] != 0x68 {
		t.Errorf("Start byte = %X, want 0x68", frame[0])
	}

	// Check L-field (length)
	expectedLen := byte(3 + 3) // C + A + CI + data length
	if frame[1] != expectedLen || frame[2] != expectedLen {
		t.Errorf("L-field = %X %X, want %X %X", frame[1], frame[2], expectedLen, expectedLen)
	}

	// Check second start byte
	if frame[3] != 0x68 {
		t.Errorf("Second start byte = %X, want 0x68", frame[3])
	}

	// Check stop byte
	if frame[len(frame)-1] != 0x16 {
		t.Errorf("Stop byte = %X, want 0x16", frame[len(frame)-1])
	}

	// Verify checksum
	var checksum byte
	for i := 4; i < len(frame)-2; i++ {
		checksum += frame[i]
	}
	if frame[len(frame)-2] != checksum {
		t.Errorf("Checksum = %X, calculated %X", frame[len(frame)-2], checksum)
	}
}

// TestCFieldValues verifies C-Field values for write operations
func TestCFieldValues(t *testing.T) {
	tests := []struct {
		name     string
		cfield   CField
		expected byte
	}{
		{"SND_UD_0", CFIELD_SND_UD_0, 0x53},
		{"SND_UD_1", CFIELD_SND_UD_1, 0x73},
		{"SND_NKE", CFIELD_SND_NKE, 0x40},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cfield.getByte()
			if got != tt.expected {
				t.Errorf("%s.getByte() = %X, want %X", tt.name, got, tt.expected)
			}
		})
	}
}

// TestCIFieldValues verifies CI-Field values for write operations
func TestCIFieldValues(t *testing.T) {
	tests := []struct {
		name     string
		cifield  CIField
		expected byte
	}{
		{"DATA_SEND", CiFieldDataSend, 0x51},
		{"APPLICATION_RESET", CiFieldApplicationReset, 0x50},
		{"SELECTION_OF_SLAVES", CiFieldSelectionOfSlaves, 0x52},
		{"SYNCHRONIZE_ACTION", CiFieldSynchronizeAction, 0x54},
		{"BAUDRATE_9600", CiFieldBaudrate9600, 0xBD},
		{"BAUDRATE_2400", CiFieldBaudrate2400, 0xBB},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := byte(tt.cifield)
			if got != tt.expected {
				t.Errorf("%s = %X, want %X", tt.name, got, tt.expected)
			}
		})
	}
}

// =============================================================================
// Edge Cases and Error Handling Tests
// =============================================================================

// TestEncodeManufacturerId_Invalid tests invalid manufacturer ID handling
func TestEncodeManufacturerId_Invalid(t *testing.T) {
	tests := []struct {
		name string
		id   string
	}{
		{"Empty", ""},
		{"TooShort", "AB"},
		{"TooLong", "ABCD"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EncodeManufacturerId(tt.id)
			expected := []byte{0, 0}
			if !bytes.Equal(got, expected) {
				t.Errorf("EncodeManufacturerId(%q) = %X, want %X for invalid input", tt.id, got, expected)
			}
		})
	}
}

// TestBuildLongFrame_EmptyData tests frame building with no data
func TestBuildLongFrame_EmptyData(t *testing.T) {
	frame := BuildLongFrame(0x53, 0xFE, 0x50, nil)

	// Should be a valid control frame
	if len(frame) != 9 {
		t.Errorf("Empty data frame length = %d, want 9", len(frame))
	}

	// L-field should be 3 (C + A + CI)
	if frame[1] != 0x03 || frame[2] != 0x03 {
		t.Errorf("L-field for empty data = %X %X, want 03 03", frame[1], frame[2])
	}
}

// TestBuildLongFrame_MaxData tests frame building with maximum data
func TestBuildLongFrame_MaxData(t *testing.T) {
	// Maximum L-field is 252 (0xFC), so max data is 252-3=249 bytes
	data := make([]byte, 249)
	for i := range data {
		data[i] = byte(i)
	}

	frame := BuildLongFrame(0x53, 0x01, 0x51, data)

	// L-field should be 252
	if frame[1] != 0xFC || frame[2] != 0xFC {
		t.Errorf("L-field for max data = %X %X, want FC FC", frame[1], frame[2])
	}
}

// TestCOMMAND_SET_COUNTER_BCD8 tests BCD counter setting
func TestCOMMAND_SET_COUNTER_BCD8(t *testing.T) {
	// Set energy counter to 107 kWh
	got := COMMAND_SET_COUNTER_BCD8(1, 0x06, 107)

	// Frame should contain: 68 0A 0A 68 | 53 01 51 | 0C 06 07 01 00 00 | CS 16
	// Verify structure
	if got[0] != 0x68 {
		t.Errorf("Start byte = %X, want 0x68", got[0])
	}

	// DIF should be 0x0C (8-digit BCD)
	if got[7] != 0x0C {
		t.Errorf("DIF = %X, want 0x0C", got[7])
	}

	// VIF should be 0x06 (1 kWh)
	if got[8] != 0x06 {
		t.Errorf("VIF = %X, want 0x06", got[8])
	}

	// BCD value for 107 should be 07 01 00 00
	bcdValue := got[9:13]
	expected := []byte{0x07, 0x01, 0x00, 0x00}
	if !bytes.Equal(bcdValue, expected) {
		t.Errorf("BCD value = %X, want %X", bcdValue, expected)
	}
}

// TestCOMMAND_SYNCHRONIZE tests synchronize command
func TestCOMMAND_SYNCHRONIZE(t *testing.T) {
	got := COMMAND_SYNCHRONIZE(0xFF) // Broadcast to all slaves

	// Should be a control frame with CI=0x54
	if got[6] != 0x54 {
		t.Errorf("CI-field = %X, want 0x54 (SYNCHRONIZE_ACTION)", got[6])
	}

	// Length should be 3
	if got[1] != 0x03 {
		t.Errorf("L-field = %X, want 0x03", got[1])
	}
}
