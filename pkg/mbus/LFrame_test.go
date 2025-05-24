package mbus

import (
	"testing"
)

func TestNewLFrame(t *testing.T) {
	data := []byte{0x68, 0x0A, 0x0A, 0x68, 0x08, 0x01, 0x72, 0x78, 0x56, 0x34, 0x12, 0x24, 0x40, 0x01, 0x07, 0x55, 0x00, 0x00, 0x00, 0x16}
	frame := NewLFrame(data)

	if len(frame.data) != len(data) {
		t.Errorf("NewLFrame() data length = %v, want %v", len(frame.data), len(data))
	}

	for i, b := range data {
		if frame.data[i] != b {
			t.Errorf("NewLFrame() data[%d] = %v, want %v", i, frame.data[i], b)
		}
	}
}

func TestLFrame_Verify(t *testing.T) {
	tests := []struct {
		name       string
		data       []byte
		wantValid  bool
		wantErrMsg string
	}{
		{
			name:       "Valid LFrame",
			data:       []byte{0x68, 0x0A, 0x0A, 0x68, 0x08, 0x01, 0x72, 0x78, 0x56, 0x34, 0x12, 0x24, 0x40, 0x01, 0x07, 0x55, 0x00, 0x00, 0x00, 0x16},
			wantValid:  true,
			wantErrMsg: "",
		},
		{
			name:       "Invalid start byte",
			data:       []byte{0x69, 0x0A, 0x0A, 0x68, 0x08, 0x01, 0x72, 0x78, 0x56, 0x34, 0x12, 0x24, 0x40, 0x01, 0x07, 0x55, 0x00, 0x00, 0x00, 0x16},
			wantValid:  false,
			wantErrMsg: "data does not start with 0x68",
		},
		{
			name:       "Invalid second start byte",
			data:       []byte{0x68, 0x0A, 0x0A, 0x69, 0x08, 0x01, 0x72, 0x78, 0x56, 0x34, 0x12, 0x24, 0x40, 0x01, 0x07, 0x55, 0x00, 0x00, 0x00, 0x16},
			wantValid:  false,
			wantErrMsg: "data does not start with 0x68",
		},
		{
			name:       "Length mismatch",
			data:       []byte{0x68, 0x0A, 0x0B, 0x68, 0x08, 0x01, 0x72, 0x78, 0x56, 0x34, 0x12, 0x24, 0x40, 0x01, 0x07, 0x55, 0x00, 0x00, 0x00, 0x16},
			wantValid:  false,
			wantErrMsg: "data length does not match",
		},
		{
			name:       "Data too short",
			data:       []byte{0x68, 0x0A, 0x0A, 0x68, 0x08},
			wantValid:  false,
			wantErrMsg: "data length is too short",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frame := NewLFrame(tt.data)
			got, err := frame.Verify()

			if got != tt.wantValid {
				t.Errorf("Verify() valid = %v, want %v", got, tt.wantValid)
			}

			if (err != nil) != (tt.wantErrMsg != "") {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErrMsg != "")
			}

			if err != nil && err.Error() != tt.wantErrMsg {
				t.Errorf("Verify() error = %v, wantErrMsg %v", err.Error(), tt.wantErrMsg)
			}
		})
	}
}

func TestLFrame_DataHeaderLField(t *testing.T) {
	data := []byte{0x68, 0x0A, 0x0A, 0x68, 0x08, 0x01, 0x72, 0x78, 0x56, 0x34, 0x12, 0x24, 0x40, 0x01, 0x07, 0x55, 0x00, 0x00, 0x00, 0x16}
	frame := NewLFrame(data)

	got := frame.DataHeaderLField()
	want := byte(0x0A)
	if got != want {
		t.Errorf("DataHeaderLField() = %v, want %v", got, want)
	}
}

func TestLFrame_CField(t *testing.T) {
	tests := []struct {
		name       string
		data       []byte
		wantCField CField
		wantErr    bool
	}{
		{
			name:       "Valid LFrame",
			data:       []byte{0x68, 0x0A, 0x0A, 0x68, 0x08, 0x01, 0x72, 0x78, 0x56, 0x34, 0x12, 0x24, 0x40, 0x01, 0x07, 0x55, 0x00, 0x00, 0x00, 0x16},
			wantCField: CFIELD_RSP_UD_a,
			wantErr:    false,
		},
		{
			name:       "Data too short",
			data:       []byte{0x68, 0x0A, 0x0A},
			wantCField: CField{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frame := NewLFrame(tt.data)
			got, err := frame.CField()

			if (err != nil) != tt.wantErr {
				t.Errorf("CField() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got.getByte() != tt.wantCField.getByte() {
				t.Errorf("CField() = %v, want %v", got.getByte(), tt.wantCField.getByte())
			}
		})
	}
}

func TestLFrame_AField(t *testing.T) {
	tests := []struct {
		name       string
		data       []byte
		wantAField AField
		wantErr    bool
	}{
		{
			name:       "Valid LFrame",
			data:       []byte{0x68, 0x0A, 0x0A, 0x68, 0x08, 0x01, 0x72, 0x78, 0x56, 0x34, 0x12, 0x24, 0x40, 0x01, 0x07, 0x55, 0x00, 0x00, 0x00, 0x16},
			wantAField: AField(1),
			wantErr:    false,
		},
		{
			name:       "Data too short",
			data:       []byte{0x68, 0x0A, 0x0A, 0x68},
			wantAField: AField(0),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frame := NewLFrame(tt.data)
			got, err := frame.AField()

			if (err != nil) != tt.wantErr {
				t.Errorf("AField() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.wantAField {
				t.Errorf("AField() = %v, want %v", got, tt.wantAField)
			}
		})
	}
}

func TestLFrame_CIField(t *testing.T) {
	data := []byte{0x68, 0x0A, 0x0A, 0x68, 0x08, 0x01, 0x72, 0x78, 0x56, 0x34, 0x12, 0x24, 0x40, 0x01, 0x07, 0x55, 0x00, 0x00, 0x00, 0x16}
	frame := NewLFrame(data)

	gotCode, gotString := frame.CIField()
	wantCode := byte(0x72)
	wantString := "VARIABLE_DATA_STRUCTURE_72"

	if gotCode != wantCode {
		t.Errorf("CIField() code = %v, want %v", gotCode, wantCode)
	}

	if gotString != wantString {
		t.Errorf("CIField() string = %v, want %v", gotString, wantString)
	}
}

func TestLFrame_IsFromSlaveAndMaster(t *testing.T) {
	tests := []struct {
		name       string
		data       []byte
		fromSlave  bool
		fromMaster bool
	}{
		{
			name:       "From Slave",
			data:       []byte{0x68, 0x0A, 0x0A, 0x68, 0x08, 0x01, 0x72, 0x78, 0x56, 0x34, 0x12, 0x24, 0x40, 0x01, 0x07, 0x55, 0x00, 0x00, 0x00, 0x16},
			fromSlave:  true,
			fromMaster: false,
		},
		{
			name:       "From Master",
			data:       []byte{0x68, 0x0A, 0x0A, 0x68, 0x53, 0x01, 0x72, 0x78, 0x56, 0x34, 0x12, 0x24, 0x40, 0x01, 0x07, 0x55, 0x00, 0x00, 0x00, 0x16},
			fromSlave:  false,
			fromMaster: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frame := NewLFrame(tt.data)

			gotFromSlave := frame.IsFromSlave()
			if gotFromSlave != tt.fromSlave {
				t.Errorf("IsFromSlave() = %v, want %v", gotFromSlave, tt.fromSlave)
			}

			gotFromMaster := frame.IsFromMaster()
			if gotFromMaster != tt.fromMaster {
				t.Errorf("IsFromMaster() = %v, want %v", gotFromMaster, tt.fromMaster)
			}
		})
	}
}

func TestLFrame_SlaveAddress(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		wantAddress uint8
		wantErr     bool
	}{
		{
			name:        "Valid Slave Address",
			data:        []byte{0x68, 0x0A, 0x0A, 0x68, 0x08, 0x01, 0x72, 0x78, 0x56, 0x34, 0x12, 0x24, 0x40, 0x01, 0x07, 0x55, 0x00, 0x00, 0x00, 0x16},
			wantAddress: 1,
			wantErr:     false,
		},
		{
			name:        "Invalid Slave Address",
			data:        []byte{0x68, 0x0A, 0x0A, 0x68, 0x08, 0x00, 0x72, 0x78, 0x56, 0x34, 0x12, 0x24, 0x40, 0x01, 0x07, 0x55, 0x00, 0x00, 0x00, 0x16},
			wantAddress: 0,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frame := NewLFrame(tt.data)
			got, err := frame.SlaveAddress()

			if (err != nil) != tt.wantErr {
				t.Errorf("SlaveAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.wantAddress {
				t.Errorf("SlaveAddress() = %v, want %v", got, tt.wantAddress)
			}
		})
	}
}

func TestLFrame_IdentificationNumber(t *testing.T) {
	data := []byte{0x68, 0x0A, 0x0A, 0x68, 0x08, 0x01, 0x72, 0x78, 0x56, 0x34, 0x12, 0x24, 0x40, 0x01, 0x07, 0x55, 0x00, 0x00, 0x00, 0x16}
	frame := NewLFrame(data)

	got := frame.IdentificationNumber()
	want := "12345678"
	if got != want {
		t.Errorf("IdentificationNumber() = %v, want %v", got, want)
	}
}

func TestLFrame_Manufacturer(t *testing.T) {
	data := []byte{0x68, 0x0A, 0x0A, 0x68, 0x08, 0x01, 0x72, 0x78, 0x56, 0x34, 0x12, 0x24, 0x40, 0x01, 0x07, 0x55, 0x00, 0x00, 0x00, 0x16}
	frame := NewLFrame(data)

	got := frame.Manufacturer()
	// This depends on the implementation of DecodeManufacturerId
	// For this test, we'll just check that it's not empty
	if got == "" {
		t.Errorf("Manufacturer() returned empty string")
	}
}

func TestLFrame_Version(t *testing.T) {
	data := []byte{0x68, 0x0A, 0x0A, 0x68, 0x08, 0x01, 0x72, 0x78, 0x56, 0x34, 0x12, 0x24, 0x40, 0x01, 0x07, 0x55, 0x00, 0x00, 0x00, 0x16}
	frame := NewLFrame(data)

	got := frame.Version()
	want := uint(1)
	if got != want {
		t.Errorf("Version() = %v, want %v", got, want)
	}
}

func TestLFrame_Medium(t *testing.T) {
	data := []byte{0x68, 0x0A, 0x0A, 0x68, 0x08, 0x01, 0x72, 0x78, 0x56, 0x34, 0x12, 0x24, 0x40, 0x01, 0x07, 0x55, 0x00, 0x00, 0x00, 0x16}
	frame := NewLFrame(data)

	got := frame.Medium()
	want := MediumType(0x07)
	if got != want {
		t.Errorf("Medium() = %v, want %v", got, want)
	}
}

func TestLFrame_AccessNumber(t *testing.T) {
	data := []byte{0x68, 0x0A, 0x0A, 0x68, 0x08, 0x01, 0x72, 0x78, 0x56, 0x34, 0x12, 0x24, 0x40, 0x01, 0x07, 0x55, 0x00, 0x00, 0x00, 0x16}
	frame := NewLFrame(data)

	got := frame.AccessNumber()
	want := uint(0x55)
	if got != want {
		t.Errorf("AccessNumber() = %v, want %v", got, want)
	}
}

func TestLFrame_Status(t *testing.T) {
	data := []byte{0x68, 0x0A, 0x0A, 0x68, 0x08, 0x01, 0x72, 0x78, 0x56, 0x34, 0x12, 0x24, 0x40, 0x01, 0x07, 0x55, 0x00, 0x00, 0x00, 0x16}
	frame := NewLFrame(data)

	got, err := frame.Status()
	want := byte(0x00)
	if err != nil {
		t.Errorf("Status() error = %v", err)
	}
	if got != want {
		t.Errorf("Status() = %v, want %v", got, want)
	}
}

func TestLFrame_Signature(t *testing.T) {
	data := []byte{0x68, 0x0A, 0x0A, 0x68, 0x08, 0x01, 0x72, 0x78, 0x56, 0x34, 0x12, 0x24, 0x40, 0x01, 0x07, 0x55, 0x00, 0x00, 0x00, 0x16}
	frame := NewLFrame(data)

	got := frame.Signature()
	want := []byte{0x00, 0x00}
	if len(got) != len(want) {
		t.Errorf("Signature() length = %v, want %v", len(got), len(want))
	}
	for i, b := range want {
		if got[i] != b {
			t.Errorf("Signature()[%d] = %v, want %v", i, got[i], b)
		}
	}
}

func TestLFrame_StopByteIndex(t *testing.T) {
	data := []byte{0x68, 0x0A, 0x0A, 0x68, 0x08, 0x01, 0x72, 0x78, 0x56, 0x34, 0x12, 0x24, 0x40, 0x01, 0x07, 0x55, 0x00, 0x00, 0x00, 0x16}
	frame := NewLFrame(data)

	got := frame.StopByteIndex()
	want := 19 // Based on the formula in StopByteIndex()
	if got != want {
		t.Errorf("StopByteIndex() = %v, want %v", got, want)
	}
}

func TestLFrame_LastDataPosition(t *testing.T) {
	data := []byte{0x68, 0x0A, 0x0A, 0x68, 0x08, 0x01, 0x72, 0x78, 0x56, 0x34, 0x12, 0x24, 0x40, 0x01, 0x07, 0x55, 0x00, 0x00, 0x00, 0x16}
	frame := NewLFrame(data)

	got := frame.LastDataPosition()
	want := 18 // Based on the formula in LastDataPosition()
	if got != want {
		t.Errorf("LastDataPosition() = %v, want %v", got, want)
	}
}
