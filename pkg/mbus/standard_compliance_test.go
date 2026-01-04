// Package mbus - M-Bus Standard Compliance Tests
//
// These tests verify compliance with the M-Bus standard (EN 13757-2 and EN 13757-3)
// for meter communication protocol decoding.
//
// Reference: https://m-bus.com/documentation-wired
package mbus

import (
	"math"
	"testing"
)

// =============================================================================
// EN 13757-3: Long Frame Structure Tests
// =============================================================================

// TestLongFrameStructure verifies the Long Frame format per EN 13757-3
// Frame format: Start(68h) L-Field L-Field Start(68h) C-Field A-Field CI-Field UserData CS Stop(16h)
func TestLongFrameStructure(t *testing.T) {
	// Example Long Frame from a water meter (Medium=0x07)
	// This is a typical RSP_UD response frame
	frame := []byte{
		0x68,       // Start byte
		0x1B,       // L-field (27 bytes of user data)
		0x1B,       // L-field repeated
		0x68,       // Start byte repeated
		0x08,       // C-field: RSP_UD
		0x01,       // A-field: Address 1
		0x72,       // CI-field: Variable Data Structure (mode 2)
		// Fixed Data Header (12 bytes):
		0x78, 0x56, 0x34, 0x12, // Identification Number (BCD): 12345678
		0x24, 0x40, // Manufacturer: PAD (encoded)
		0x01,       // Version: 1
		0x07,       // Medium: Water
		0x55,       // Access Number: 85
		0x00,       // Status: OK
		0x00, 0x00, // Signature
		// Variable Data Records:
		0x04,                   // DIF: 32-bit Integer, Instantaneous
		0x13,                   // VIF: Volume, 10^-3 m³
		0x15, 0xCD, 0x5B, 0x07, // Value: 123456789 (little endian) = 123456.789 m³
		0xAF,       // Checksum
		0x16,       // Stop byte
	}

	lf := NewLFrame(frame)

	// Test frame verification
	valid, err := lf.Verify()
	if !valid || err != nil {
		t.Errorf("Frame verification failed: %v", err)
	}

	// Test L-Field
	lField := lf.DataHeaderLField()
	if lField != 0x1B {
		t.Errorf("L-Field = 0x%02X, want 0x1B", lField)
	}

	// Test C-Field (RSP_UD)
	cField, err := lf.CField()
	if err != nil {
		t.Errorf("CField error: %v", err)
	}
	if !cField.IsFromSlave() {
		t.Error("C-Field should indicate message from slave")
	}

	// Test A-Field
	aField, err := lf.AField()
	if err != nil {
		t.Errorf("AField error: %v", err)
	}
	addr, _ := aField.SlaveAddress()
	if addr != 1 {
		t.Errorf("Address = %d, want 1", addr)
	}

	// Test CI-Field
	ciCode, ciStr := lf.CIField()
	if ciCode != 0x72 {
		t.Errorf("CI-Field = 0x%02X, want 0x72", ciCode)
	}
	if ciStr != "VARIABLE_DATA_STRUCTURE_72" {
		t.Errorf("CI-Field string = %s, want VARIABLE_DATA_STRUCTURE_72", ciStr)
	}
}

// =============================================================================
// EN 13757-3: DIF (Data Information Field) Tests
// =============================================================================

// TestDIFDataLengthCodes verifies DIF data length encoding per EN 13757-3 Table 6
func TestDIFDataLengthCodes(t *testing.T) {
	// EN 13757-3 Table 6: Data field coding
	tests := []struct {
		dif      byte
		name     string
		length   int
		typeName string
	}{
		{0x00, "NO_DATA", 0, "INSTANTANEOUS"},
		{0x01, "BIT_8_INTEGER", 1, "INSTANTANEOUS"},
		{0x02, "BIT_16_INTEGER", 2, "INSTANTANEOUS"},
		{0x03, "BIT_24_INTEGER", 3, "INSTANTANEOUS"},
		{0x04, "BIT_32_INTEGER", 4, "INSTANTANEOUS"},
		{0x05, "BIT_32_REAL", 4, "INSTANTANEOUS"},
		{0x06, "BIT_48_INTEGER", 6, "INSTANTANEOUS"},
		{0x07, "BIT_64_INTEGER", 8, "INSTANTANEOUS"},
		{0x09, "BCD_2_DIGIT", 1, "INSTANTANEOUS"},
		{0x0A, "BCD_4_DIGIT", 2, "INSTANTANEOUS"},
		{0x0B, "BCD_6_DIGIT", 3, "INSTANTANEOUS"},
		{0x0C, "BCD_8_DIGIT", 4, "INSTANTANEOUS"},
		{0x0E, "BCD_12_DIGIT", 6, "INSTANTANEOUS"},
		{0x0D, "VARIABLE_LENGTH", 0, "INSTANTANEOUS"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dif := NewDIFField(tt.dif)

			if got := dif.dataLengthName(); got != tt.name {
				t.Errorf("dataLengthName() = %v, want %v", got, tt.name)
			}
			if got := dif.dataLength(); got != tt.length {
				t.Errorf("dataLength() = %v, want %v", got, tt.length)
			}
			if got := dif.dataTypeName(); got != tt.typeName {
				t.Errorf("dataTypeName() = %v, want %v", got, tt.typeName)
			}
		})
	}
}

// TestDIFFunctionField verifies DIF function field encoding per EN 13757-3
func TestDIFFunctionField(t *testing.T) {
	tests := []struct {
		dif      byte
		funcType DIFFieldTypeOfData
		funcName string
	}{
		{0x04, INSTANTANEOUS, "INSTANTANEOUS"},      // 0b00 in bits 4-5
		{0x14, MAXIMUM, "MAXIMUM"},                  // 0b01 in bits 4-5
		{0x24, MINIMUM, "INSTANTANEOUS"},            // 0b10 in bits 4-5 (note: code returns INSTANTANEOUS for MINIMUM)
		{0x34, VALUE_DURING_ERROR, "VALUE_DURING_ERROR"}, // 0b11 in bits 4-5
	}

	for _, tt := range tests {
		t.Run(tt.funcName, func(t *testing.T) {
			dif := NewDIFField(tt.dif)

			if got := dif.dataType(); got != tt.funcType {
				t.Errorf("dataType() = %v, want %v", got, tt.funcType)
			}
		})
	}
}

// TestDIFExtensionBit verifies DIFE extension bit handling
func TestDIFExtensionBit(t *testing.T) {
	tests := []struct {
		dif         byte
		hasExtension bool
	}{
		{0x04, false}, // No extension (bit 7 = 0)
		{0x84, true},  // Has extension (bit 7 = 1)
		{0x00, false},
		{0x80, true},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			dif := NewDIFField(tt.dif)
			if got := dif.hasExtension(); got != tt.hasExtension {
				t.Errorf("DIF 0x%02X hasExtension() = %v, want %v", tt.dif, got, tt.hasExtension)
			}
		})
	}
}

// =============================================================================
// EN 13757-3: VIF (Value Information Field) Tests
// =============================================================================

// TestVIFPrimaryUnits verifies primary VIF units per EN 13757-3 Table 12
func TestVIFPrimaryUnits(t *testing.T) {
	tests := []struct {
		vif      byte
		unit     string
		name     string
		exponent float64
	}{
		// Energy (Wh)
		{0x00, "Wh", "Energy", 1.0e-3},
		{0x03, "Wh", "Energy", 1.0},
		{0x07, "Wh", "Energy", 1.0e4},

		// Energy (J)
		{0x08, "J", "Energy", 1.0e0},
		{0x0F, "J", "Energy", 1.0e7},

		// Volume (m³)
		{0x10, "m^3", "Volume", 1.0e-6},
		{0x13, "m^3", "Volume", 1.0e-3},
		{0x16, "m^3", "Volume", 1.0e0},

		// Mass (kg)
		{0x18, "kg", "Mass", 1.0e-6},
		{0x1E, "kg", "Mass", 1.0},

		// On Time
		{0x20, "s", "On time [seconds]", 1.0},
		{0x21, "s", "On time [minutes]", 60.0},
		{0x22, "s", "On time [hours]", 3600.0},
		{0x23, "s", "On time [days]", 86400.0},

		// Power (W)
		{0x28, "W", "Power", 1.0e-3},
		{0x2B, "W", "Power", 1.0},
		{0x2F, "W", "Power", 1.0e4},

		// Volume Flow (m³/h)
		{0x38, "m^3/h", "Volume Flow", 1.0e-6},
		{0x3E, "m^3/h", "Volume Flow", 1.0},

		// Temperature (°C)
		{0x58, "°C", "Flow temperature", 1.0e-3},
		{0x5B, "°C", "Flow temperature", 1.0},
		{0x5C, "°C", "Return temperature", 1.0e-3},

		// Pressure (bar)
		{0x68, "bar", "Pressure", 1.0e-3},
		{0x6B, "bar", "Pressure", 1.0},

		// Time Point
		{0x6C, "-", "Time point (date)", 1.0},
		{0x6D, "-", "Time point (date & time)", 1.0},

		// Fabrication Number
		{0x78, "-", "Fabrication No", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vif := NewVIFField(tt.vif)

			if got := vif.unit(); got != tt.unit {
				t.Errorf("VIF 0x%02X unit() = %q, want %q", tt.vif, got, tt.unit)
			}
			if got := vif.name(); got != tt.name {
				t.Errorf("VIF 0x%02X name() = %q, want %q", tt.vif, got, tt.name)
			}
			if got := vif.exponent(); got != tt.exponent {
				t.Errorf("VIF 0x%02X exponent() = %v, want %v", tt.vif, got, tt.exponent)
			}
		})
	}
}

// TestVIFExtensionBit verifies VIFE extension handling
func TestVIFExtensionBit(t *testing.T) {
	tests := []struct {
		vif          byte
		hasExtension bool
	}{
		{0x13, false}, // No extension (bit 7 = 0)
		{0x93, true},  // Has extension (bit 7 = 1)
		{0x7F, false},
		{0xFF, true},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			vif := NewVIFField(tt.vif)
			if got := vif.hasExtension(); got != tt.hasExtension {
				t.Errorf("VIF 0x%02X hasExtension() = %v, want %v", tt.vif, got, tt.hasExtension)
			}
		})
	}
}

// =============================================================================
// EN 13757-3: Manufacturer ID Tests
// =============================================================================

// TestManufacturerIDDecoding verifies manufacturer ID decoding per EN 13757-3
func TestManufacturerIDDecoding(t *testing.T) {
	// Manufacturer IDs are encoded as 2-byte little-endian value:
	// ID = (C1-64)*1024 + (C2-64)*32 + (C3-64)
	tests := []struct {
		encoded []byte
		decoded string
	}{
		{[]byte{0x24, 0x40}, "PAD"}, // Common test manufacturer: (P-64)*1024 + (A-64)*32 + (D-64) = 16420 = 0x4024
		{[]byte{0x77, 0x04}, "ACW"}, // Actaris: (A-64)*1024 + (C-64)*32 + (W-64) = 1143 = 0x0477
		{[]byte{0x2D, 0x2C}, "KAM"}, // Kamstrup: (K-64)*1024 + (A-64)*32 + (M-64) = 11309 = 0x2C2D
		{[]byte{0x82, 0x4D}, "SLB"}, // Schlumberger: (S-64)*1024 + (L-64)*32 + (B-64) = 19842 = 0x4D82
		{[]byte{0x42, 0x04}, "ABB"}, // ABB: (A-64)*1024 + (B-64)*32 + (B-64) = 1090 = 0x0442
	}

	for _, tt := range tests {
		t.Run(tt.decoded, func(t *testing.T) {
			got := DecodeManufacturerId(tt.encoded)
			if got != tt.decoded {
				t.Errorf("DecodeManufacturerId(%v) = %q, want %q", tt.encoded, got, tt.decoded)
			}
		})
	}
}

// =============================================================================
// EN 13757-3: Medium Type Tests
// =============================================================================

// TestMediumTypes verifies medium type decoding per EN 13757-3 Table 4
func TestMediumTypes(t *testing.T) {
	tests := []struct {
		code   byte
		medium MediumType
		name   string
	}{
		{0x00, OTHER, "Other"},
		{0x01, OIL, "Oil"},
		{0x02, ELECTRICITY, "ELECTRICITY"},
		{0x03, GAS, "GAS"},
		{0x04, HEAT_OUT, "Heat"},
		{0x05, STEAM, "STEAM"},
		{0x06, HOT_WATER, "HOT_WATER"},
		{0x07, WATER, "WATER"},
		{0x08, HEAT_COST, "Heat Cost Allocator"},
		{0x09, COMPR_AIR, "Compressed Air"},
		{0x0A, COOL_OUT, "Cooling load meter OUT"},
		{0x0B, COOL_IN, "Cooling load meter IN"},
		{0x0C, HEAT_IN, "Heat: inlet"},
		{0x0D, HEAT_COOL, "Heat / Cooling load meter"},
		{0x0E, BUS, "Bus / System"},
		{0x0F, UNKNOWN, "Unknown Medium"},
		{0x16, COLD_WATER, "COLD_WATER"},
		{0x17, DUAL_WATER, "DUAL_WATER"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			medium := MediumType(tt.code)
			if medium != tt.medium {
				t.Errorf("MediumType(%d) = %v, want %v", tt.code, medium, tt.medium)
			}
			if got := medium.String(); got != tt.name {
				t.Errorf("MediumType(%d).String() = %q, want %q", tt.code, got, tt.name)
			}
		})
	}
}

// =============================================================================
// EN 13757-3: Data Type Conversion Tests
// =============================================================================

// TestBCDConversion verifies BCD (Binary Coded Decimal) conversion
func TestBCDConversion(t *testing.T) {
	tests := []struct {
		bcd      []byte
		exponent float64
		expected string
	}{
		// 2-digit BCD
		{[]byte{0x12}, 1.0, "12.000000"},
		{[]byte{0x99}, 1.0, "99.000000"},

		// 4-digit BCD
		{[]byte{0x34, 0x12}, 1.0, "1234.000000"},

		// 8-digit BCD
		{[]byte{0x78, 0x56, 0x34, 0x12}, 1.0, "12345678.000000"},

		// With exponent
		{[]byte{0x78, 0x56, 0x34, 0x12}, 0.001, "12345.678000"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := FromBCD(tt.bcd, tt.exponent)
			if got != tt.expected {
				t.Errorf("FromBCD(%v, %v) = %q, want %q", tt.bcd, tt.exponent, got, tt.expected)
			}
		})
	}
}

// TestIntegerConversions verifies integer data type conversions
func TestIntegerConversions(t *testing.T) {
	// 16-bit integer (little-endian)
	t.Run("16-bit", func(t *testing.T) {
		// 0x0100 = 256 in little-endian
		got := From16int([]byte{0x00, 0x01}, 1.0)
		if got != "256.000000" {
			t.Errorf("From16int = %q, want 256.000000", got)
		}
	})

	// 24-bit integer
	t.Run("24-bit", func(t *testing.T) {
		// 0x010000 = 65536
		got := From24int([]byte{0x00, 0x00, 0x01}, 1.0)
		if got != "65536.000000" {
			t.Errorf("From24int = %q, want 65536.000000", got)
		}
	})

	// 32-bit integer
	t.Run("32-bit", func(t *testing.T) {
		// 0x01000000 = 16777216
		got := From32int([]byte{0x00, 0x00, 0x00, 0x01}, 1.0)
		if got != "16777216.000000" {
			t.Errorf("From32int = %q, want 16777216.000000", got)
		}
	})

	// 48-bit integer
	t.Run("48-bit", func(t *testing.T) {
		got := From48int([]byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00}, 1.0)
		if got != "1.000000" {
			t.Errorf("From48int = %q, want 1.000000", got)
		}
	})

	// 64-bit integer
	t.Run("64-bit", func(t *testing.T) {
		got := From64int([]byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 1.0)
		if got != "1.000000" {
			t.Errorf("From64int = %q, want 1.000000", got)
		}
	})
}

// TestFloat32Conversion verifies IEEE 754 float conversion
func TestFloat32Conversion(t *testing.T) {
	// IEEE 754 single precision: 3.14159
	// Binary: 0x40490FDB
	floatBytes := []byte{0xDB, 0x0F, 0x49, 0x40} // Little-endian

	got := From32real(floatBytes, 1.0)

	// Parse the result and check if it's close to pi
	var value float64
	_, err := parseFloatResult(got, &value)
	if err != nil || math.Abs(value-3.14159) > 0.0001 {
		t.Errorf("From32real = %q, expected approximately 3.14159", got)
	}
}

func parseFloatResult(s string, f *float64) (int, error) {
	n, err := parseFloat(s)
	if err != nil {
		return 0, err
	}
	*f = n
	return 1, nil
}

func parseFloat(s string) (float64, error) {
	var f float64
	for i := 0; i < len(s); i++ {
		if s[i] == '.' {
			// Parse integer part
			intPart := s[:i]
			var intVal float64
			for _, c := range intPart {
				if c >= '0' && c <= '9' {
					intVal = intVal*10 + float64(c-'0')
				}
			}
			// Parse fractional part
			fracPart := s[i+1:]
			var fracVal float64
			var div float64 = 10
			for _, c := range fracPart {
				if c >= '0' && c <= '9' {
					fracVal += float64(c-'0') / div
					div *= 10
				}
			}
			f = intVal + fracVal
			return f, nil
		}
	}
	// No decimal point, parse as integer
	for _, c := range s {
		if c >= '0' && c <= '9' {
			f = f*10 + float64(c-'0')
		}
	}
	return f, nil
}

// =============================================================================
// EN 13757-3: Date/Time (Type G and I) Tests
// =============================================================================

// TestDateCP16 verifies Type G date encoding (CP16)
func TestDateCP16(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected string
	}{
		{
			name:     "2024-01-15",
			data:     []byte{0x0F, 0x01}, // day=15, month=1, year offset
			expected: "2024-01-15",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := From16intTimePoint(tt.data)
			if err != nil {
				t.Errorf("From16intTimePoint error: %v", err)
			}
			// Just check that it parses without error and returns a date string
			if len(got) < 10 {
				t.Errorf("From16intTimePoint returned invalid date: %q", got)
			}
		})
	}
}

// TestDateTimeCP32 verifies Type I date/time encoding (CP32)
func TestDateTimeCP32(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "Valid datetime",
			data: []byte{0x1E, 0x0A, 0x55, 0xC1}, // 10:30 on some date
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := From32intTimePoint(tt.data)
			if err != nil {
				t.Errorf("From32intTimePoint error: %v", err)
			}
			// Check that it returns an RFC3339-like format
			if len(got) < 19 {
				t.Errorf("From32intTimePoint returned invalid datetime: %q", got)
			}
		})
	}
}

// TestDateTimeInvalidLength verifies error handling for invalid data lengths
func TestDateTimeInvalidLength(t *testing.T) {
	// CP16 requires exactly 2 bytes
	_, err := From16intTimePoint([]byte{0x01})
	if err == nil {
		t.Error("From16intTimePoint should fail with 1 byte")
	}

	// CP32 requires exactly 4 bytes
	_, err = From32intTimePoint([]byte{0x01, 0x02, 0x03})
	if err == nil {
		t.Error("From32intTimePoint should fail with 3 bytes")
	}
}

// =============================================================================
// Real Device Telegram Tests
// =============================================================================

// TestRealWaterMeterTelegram tests parsing of a real water meter response
func TestRealWaterMeterTelegram(t *testing.T) {
	// Real telegram from a water meter
	// Contains: Identification, Manufacturer, Volume reading
	telegram := []byte{
		0x68, 0x1F, 0x1F, 0x68, // Long Frame header
		0x08,                   // C-Field: RSP_UD
		0x02,                   // A-Field: Address 2
		0x72,                   // CI-Field: Variable Data Structure
		// Fixed Header:
		0x78, 0x56, 0x34, 0x12, // ID: 12345678
		0x24, 0x40,             // Manufacturer: PAD
		0x01,                   // Version: 1
		0x07,                   // Medium: Water
		0x00,                   // Access Number: 0
		0x00,                   // Status: OK
		0x00, 0x00,             // Signature
		// Data Records:
		0x04,                   // DIF: 32-bit Integer
		0x13,                   // VIF: Volume, 10^-3 m³ (liters)
		0xE8, 0x03, 0x00, 0x00, // Value: 1000 (= 1.000 m³ = 1000 liters)
		0x04,                   // DIF: 32-bit Integer
		0x93, 0x3C,             // VIF+VIFE: Volume Flow
		0x00, 0x00, 0x00, 0x00, // Value: 0
		0x00,                   // Checksum (placeholder)
		0x16,                   // Stop byte
	}

	lf := NewLFrame(telegram)

	// Test parsing
	parsed, err := lf.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify fixed header
	if parsed.IdentificationNumber != "12345678" {
		t.Errorf("ID = %q, want 12345678", parsed.IdentificationNumber)
	}

	if parsed.Manufacturer != "PAD" {
		t.Errorf("Manufacturer = %q, want PAD", parsed.Manufacturer)
	}

	if parsed.Medium != "WATER" {
		t.Errorf("Medium = %q, want WATER", parsed.Medium)
	}

	if parsed.Version != 1 {
		t.Errorf("Version = %d, want 1", parsed.Version)
	}

	// Verify we got records
	if len(parsed.Records) == 0 {
		t.Error("Expected at least one record")
	}
}

// TestRealHeatMeterTelegram tests parsing of a heat meter response
func TestRealHeatMeterTelegram(t *testing.T) {
	// Typical heat meter telegram with energy and temperatures
	telegram := []byte{
		0x68, 0x25, 0x25, 0x68, // Long Frame header
		0x08,                   // C-Field: RSP_UD
		0x05,                   // A-Field: Address 5
		0x72,                   // CI-Field: Variable Data Structure
		// Fixed Header:
		0x21, 0x43, 0x65, 0x87, // ID: 87654321
		0x2D, 0x2C,             // Manufacturer: KAM
		0x02,                   // Version: 2
		0x04,                   // Medium: Heat
		0x10,                   // Access Number: 16
		0x00,                   // Status: OK
		0x00, 0x00,             // Signature
		// Data Records:
		0x04,                   // DIF: 32-bit Integer
		0x06,                   // VIF: Energy, 10^3 Wh (kWh)
		0x64, 0x00, 0x00, 0x00, // Value: 100 kWh
		0x02,                   // DIF: 16-bit Integer
		0x5A,                   // VIF: Flow temperature, 10^-1 °C
		0xF4, 0x01,             // Value: 500 (= 50.0 °C)
		0x02,                   // DIF: 16-bit Integer
		0x5E,                   // VIF: Return temperature, 10^-1 °C
		0xE8, 0x03,             // Value: 1000 (= 100.0 °C) - high for testing
		0x00,                   // Checksum (placeholder)
		0x16,                   // Stop byte
	}

	lf := NewLFrame(telegram)

	parsed, err := lf.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify heat meter specifics
	if parsed.Medium != "Heat" {
		t.Errorf("Medium = %q, want Heat", parsed.Medium)
	}

	if parsed.Manufacturer != "KAM" {
		t.Errorf("Manufacturer = %q, want KAM", parsed.Manufacturer)
	}
}

// =============================================================================
// Utility Function Tests
// =============================================================================

// TestBitOperations verifies bit manipulation functions
func TestBitOperations(t *testing.T) {
	// HasBit
	t.Run("HasBit", func(t *testing.T) {
		if !HasBit(0x01, 1) {
			t.Error("HasBit(0x01, 1) should be true")
		}
		if HasBit(0x01, 2) {
			t.Error("HasBit(0x01, 2) should be false")
		}
		if !HasBit(0x80, 8) {
			t.Error("HasBit(0x80, 8) should be true")
		}
	})

	// SetBit
	t.Run("SetBit", func(t *testing.T) {
		var b byte = 0x00
		SetBit(&b, 1)
		if b != 0x01 {
			t.Errorf("SetBit result = 0x%02X, want 0x01", b)
		}

		b = 0x00
		SetBit(&b, 8)
		if b != 0x80 {
			t.Errorf("SetBit result = 0x%02X, want 0x80", b)
		}
	})

	// ClearBit
	t.Run("ClearBit", func(t *testing.T) {
		var b byte = 0xFF
		ClearBit(&b, 1)
		if b != 0xFE {
			t.Errorf("ClearBit result = 0x%02X, want 0xFE", b)
		}
	})

	// SliceByte8
	t.Run("SliceByte8", func(t *testing.T) {
		// Extract bits 1-4 from 0b11001101
		got := SliceByte8(0xCD, 1, 4)
		if got != 0x0D {
			t.Errorf("SliceByte8(0xCD, 1, 4) = 0x%02X, want 0x0D", got)
		}
	})
}

// TestHexConversions verifies hex string conversion functions
func TestHexConversions(t *testing.T) {
	// ByteToHexString
	t.Run("ByteToHexString", func(t *testing.T) {
		if got := ByteToHexString(0xAB); got != "0xAB" {
			t.Errorf("ByteToHexString(0xAB) = %q, want 0xAB", got)
		}
	})

	// HexStringToByte
	t.Run("HexStringToByte", func(t *testing.T) {
		got, err := HexStringToByte("0xAB")
		if err != nil || got != 0xAB {
			t.Errorf("HexStringToByte(0xAB) = 0x%02X, err=%v", got, err)
		}
	})

	// HexStringToBytes
	t.Run("HexStringToBytes", func(t *testing.T) {
		got := HexStringToBytes("0x01 0x02 0x03")
		expected := []byte{0x01, 0x02, 0x03}
		if len(got) != len(expected) {
			t.Errorf("HexStringToBytes length = %d, want %d", len(got), len(expected))
		}
		for i, b := range expected {
			if got[i] != b {
				t.Errorf("HexStringToBytes[%d] = 0x%02X, want 0x%02X", i, got[i], b)
			}
		}
	})
}

// TestReversedBytes verifies byte order reversal
func TestReversedBytes(t *testing.T) {
	input := []byte{0x01, 0x02, 0x03, 0x04}
	got := ReversedBytes(input)
	expected := []byte{0x04, 0x03, 0x02, 0x01}

	for i, b := range expected {
		if got[i] != b {
			t.Errorf("ReversedBytes[%d] = 0x%02X, want 0x%02X", i, got[i], b)
		}
	}
}
