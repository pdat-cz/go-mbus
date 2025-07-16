package mbus

import "errors"

// MBus Telegram Format - Control Frame

type CFrame struct {
	data []byte
}

func NewCFrame(data []byte) CFrame {
	return CFrame{data}
}

// Verify that the CFrame is a control frame.
func (f *CFrame) Verify() (bool, error) {

	// Start with 0x68
	if f.data[0] != 0x68 {
		return false, errors.New("CFrame does not start with 0x68")
	}
	// End with 0x16
	if f.data[len(f.data)-1] != 0x16 {
		return false, errors.New("CFrame does not end with 0x16")
	}
	// Length must be 6
	if len(f.data) != 6 {
		return false, errors.New("CFrame length is not 6")
	}
	// Check checksum

	return true, nil
}

// CField is in index 5
func (f *CFrame) CField() (CField, error) {
	// Check length
	if len(f.data) < 6 {
		return CField{}, errors.New("data length is too short")
	}
	return NewCField(f.data[5]), nil
}

// AField is in index 6
func (f *CFrame) AField() (AField, error) {
	// Check length
	if len(f.data) < 7 {
		return NewAField(0), errors.New("data length is too short")
	}
	return NewAField(f.data[6]), nil
}
