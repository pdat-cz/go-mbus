package mbus

import (
	"testing"
)

func TestBoolToInt(t *testing.T) {
	type tsts struct {
		name  string
		input bool
		want  int
	}
	tests := []tsts{
		{"true", true, 1},
		{"false", false, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BoolToInt(tt.input)
			if got != tt.want {
				t.Errorf("BoolToInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestByteToHexString(t *testing.T) {
	type tst struct {
		name  string
		input byte
		want  string
	}
	tests := []tst{
		{"0", 0, "0x00"},
		{"1", 1, "0x01"},
		{"2", 2, "0x02"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ByteToHexString(tt.input)
			if got != tt.want {
				t.Errorf("ByteToHexString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHexStringToByte(t *testing.T) {
	type tst struct {
		name  string
		input string
		want  byte
	}
	tests := []tst{
		{"0", "0x00", 0},
		{"1", "0x01", 1},
		{"2", "0x02", 2},
		{"0x0A", "0x0A", 10},
		{"0xA0", "0xA0", 160},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := HexStringToByte(tt.input)
			if got != tt.want {
				t.Errorf("HexStringToByte() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHexStringToBytes(t *testing.T) {
	type tst struct {
		name  string
		input string
		want  []byte
	}
	tests := []tst{
		{"0", "0x00 00", []byte{0, 0}},
		{"1", "0x01 01", []byte{1, 1}},
		{"2", "0x02 02", []byte{2, 2}},
		{"0A 01", "0x0A 0x01", []byte{10, 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HexStringToBytes(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("HexStringToBytes() = %v, want %v", got, tt.want)
			}
			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("HexStringToBytes() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestFrom32intTimePoint(t *testing.T) {

	tData := []struct {
		desc      string
		tData     []byte
		expectedT string
	}{
		{
			desc:      "Compound CP32: Date and Time 01",
			tData:     []byte{0x00, 0x0C, 0xE3, 0xB3},
			expectedT: "1995-03-03T12:00:00Z",
		},
		{
			desc:      "Compound CP32: Date and Time 02",
			tData:     []byte{0x32, 0x0B, 0xE3, 0xB3},
			expectedT: "1995-03-03T11:50:00Z",
		},
		{
			desc:      "Compound CP32: Date and Time 03",
			tData:     []byte{0x19, 0x0F, 0x8A, 0x17},
			expectedT: "2012-07-10T15:25:00Z",
		},
		{
			desc:      "Compound CP32: Date and Time 04",
			tData:     []byte{0x0B, 0x0B, 0xCD, 0x13},
			expectedT: "2014-03-13T11:11:00Z",
		},
	}
	for _, td := range tData {
		t.Run(td.desc, func(t *testing.T) {

			get, _ := From32intTimePoint(td.tData)
			if get != td.expectedT {
				t.Errorf("From32intTimePoint() = %v, want %v", get, td.expectedT)
			}

		})
	}
}

func TestFrom16intTimePoint(t *testing.T) {

	//flgDebug = true

	tData := []struct {
		desc      string
		tData     []byte
		expectedT string
	}{
		{
			desc:      "Compound CP16: Date and Time 01",
			tData:     []byte{0xBF, 0x1C},
			expectedT: "2013-12-31",
		},
		{
			desc:      "Compound CP16: Date and Time 01",
			tData:     []byte{0xDF, 0x1C},
			expectedT: "2014-12-31",
		},
		{
			desc:      "Compound CP16: Date and Time 01",
			tData:     []byte{0xBF, 0x15},
			expectedT: "2013-05-31",
		},
	}

	for _, td := range tData {
		t.Run(td.desc, func(t *testing.T) {

			get, _ := From16intTimePoint(td.tData)
			if get != td.expectedT {
				t.Errorf("From16intTimePoint() = %v, want %v", get, td.expectedT)
			}

		})
	}

}
func TestIncludeKey(t *testing.T) {
	m := map[string]interface{}{"a": 1, "b": "two"}
	if !IncludeKey(m, "a") {
		t.Errorf("IncludeKey() returned false for existing key")
	}
	if IncludeKey(m, "c") {
		t.Errorf("IncludeKey() returned true for non-existing key")
	}
}
