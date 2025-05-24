package mbus

import (
	"testing"
)

func TestNewSFrame(t *testing.T) {
	data := []byte{0x10, 0x40, 0x01, 0x41, 0x16}
	frame := NewSFrame(data)

	if len(frame.data) != len(data) {
		t.Errorf("NewSFrame() data length = %v, want %v", len(frame.data), len(data))
	}

	for i, b := range data {
		if frame.data[i] != b {
			t.Errorf("NewSFrame() data[%d] = %v, want %v", i, frame.data[i], b)
		}
	}
}

func TestSFrame_Verify(t *testing.T) {
	tests := []struct {
		name       string
		data       []byte
		wantValid  bool
		wantErrMsg string
	}{
		{
			name:       "Valid SFrame",
			data:       []byte{0x10, 0x40, 0x01, 0x41, 0x16},
			wantValid:  true,
			wantErrMsg: "",
		},
		{
			name:       "Invalid start byte",
			data:       []byte{0x11, 0x40, 0x01, 0x41, 0x16},
			wantValid:  false,
			wantErrMsg: "SFrame does not start with 0x10",
		},
		{
			name:       "Invalid end byte",
			data:       []byte{0x10, 0x40, 0x01, 0x41, 0x17},
			wantValid:  false,
			wantErrMsg: "SFrame does not end with 0x16",
		},
		{
			name:       "Invalid length (too short)",
			data:       []byte{0x10, 0x40, 0x01, 0x16},
			wantValid:  false,
			wantErrMsg: "SFrame length is not 5",
		},
		{
			name:       "Invalid length (too long)",
			data:       []byte{0x10, 0x40, 0x01, 0x41, 0x42, 0x16},
			wantValid:  false,
			wantErrMsg: "SFrame length is not 5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frame := NewSFrame(tt.data)
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

func TestSFrame_CField(t *testing.T) {
	tests := []struct {
		name       string
		data       []byte
		wantCField CField
		wantErr    bool
	}{
		{
			name:       "Valid SFrame",
			data:       []byte{0x10, 0x40, 0x01, 0x41, 0x16},
			wantCField: CFIELD_SND_NKE,
			wantErr:    false,
		},
		{
			name:       "Empty data",
			data:       []byte{},
			wantCField: CField{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frame := NewSFrame(tt.data)
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

func TestSFrame_AField(t *testing.T) {
	tests := []struct {
		name       string
		data       []byte
		wantAField AField
		wantErr    bool
	}{
		{
			name:       "Valid SFrame",
			data:       []byte{0x10, 0x40, 0x01, 0x41, 0x16},
			wantAField: AField(1),
			wantErr:    false,
		},
		{
			name:       "Empty data",
			data:       []byte{},
			wantAField: AField(0),
			wantErr:    true,
		},
		{
			name:       "Data too short",
			data:       []byte{0x10, 0x40},
			wantAField: AField(0),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frame := NewSFrame(tt.data)
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
