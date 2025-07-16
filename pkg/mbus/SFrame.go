package mbus

import "errors"

// MBus Telegram Format - Short Frame

type SFrame struct {
	data []byte
}

func NewSFrame(data []byte) SFrame {
	return SFrame{data}
}

// Verify that the SFrame is a short frame.
func (f *SFrame) Verify() (bool, error) {

	// Start with 0x10
	if f.data[0] != 0x10 {
		return false, errors.New("SFrame does not start with 0x10")
	}
	// End with 0x16
	if f.data[len(f.data)-1] != 0x16 {
		return false, errors.New("SFrame does not end with 0x16")
	}
	// Length must be 5
	if len(f.data) != 5 {
		return false, errors.New("SFrame length is not 5")
	}
	// Check checksum

	return true, nil
}

// CField is on index 1
func (f *SFrame) CField() (CField, error) {
	// Check length
	if len(f.data) < 1 {
		return CField{}, errors.New("data length is too short")
	}
	return NewCField(f.data[1]), nil
}

// AField is on index 2
func (f *SFrame) AField() (AField, error) {
	// Check length
	if len(f.data) < 3 {
		return NewAField(0), errors.New("data length is too short")
	}
	return NewAField(f.data[2]), nil
}
