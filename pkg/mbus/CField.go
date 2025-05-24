package mbus

import (
	"fmt"
)

// CField CFIELD
type CField struct {
	position8 bool // reserved for future
	position7 bool // Direction true .. calling, false ... reply
	FCB       bool // CALLING successful transmission procedure
	FCV       bool // CALLING frame count bit valid . If false ... slave should ignore FCB
	ACD       bool // REPLY
	DFC       bool // REPLY
	F3        bool //
	F2        bool
	F1        bool
	F0        bool
}

func (f *CField) string() string {

	if f.position7 {
		// CALLING
		return fmt.Sprintf("8\t7\tFCB\tFCV\tF3\tF2\tF1\tF0\n%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t",
			BoolToInt(f.position8),
			BoolToInt(f.position7),
			BoolToInt(f.FCB),
			BoolToInt(f.FCV),
			BoolToInt(f.F3),
			BoolToInt(f.F2),
			BoolToInt(f.F1),
			BoolToInt(f.F0))
	}

	return fmt.Sprintf("8\t7\tACD\tDCF\tF3\tF2\tF1\tF0\n%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t",
		BoolToInt(f.position8),
		BoolToInt(f.position7),
		BoolToInt(f.ACD),
		BoolToInt(f.DFC),
		BoolToInt(f.F3),
		BoolToInt(f.F2),
		BoolToInt(f.F1),
		BoolToInt(f.F0))

}

func (f *CField) setByte(b byte) {
	if HasBit(b, 8) {
		f.position8 = true
	} else {
		f.position8 = false
	}
	if HasBit(b, 7) {
		f.position7 = true
	} else {
		f.position7 = false
	}

	if HasBit(b, 7) {
		// CALLING
		if HasBit(b, 6) {
			f.FCB = true
		} else {
			f.FCB = false
		}
		if HasBit(b, 5) {
			f.FCV = true
		} else {
			f.FCV = false
		}
	} else {
		// REPLY
		if HasBit(b, 6) {
			f.ACD = true
		} else {
			f.ACD = false
		}
		if HasBit(b, 5) {
			f.DFC = true
		} else {
			f.DFC = false
		}
	}

	if HasBit(b, 4) {
		f.F3 = true
	} else {
		f.F3 = false
	}
	if HasBit(b, 3) {
		f.F2 = true
	} else {
		f.F2 = false
	}
	if HasBit(b, 2) {
		f.F1 = true
	} else {
		f.F1 = false
	}
	if HasBit(b, 1) {
		f.F0 = true
	} else {
		f.F0 = false
	}
}

func (f *CField) getByte() byte {
	var b byte = 0 << 7
	if f.position8 {
		SetBit(&b, 8)
	}
	if f.position7 {
		SetBit(&b, 7)
	}
	// Calling
	if f.FCB && f.position7 {
		SetBit(&b, 6)
	}
	// Reply
	if f.ACD && !f.position7 {
		SetBit(&b, 6)
	}
	// Calling
	if f.FCV && f.position7 {
		SetBit(&b, 5)
	}
	// Reply
	if f.DFC && !f.position7 {
		SetBit(&b, 5)
	}
	if f.F3 {
		SetBit(&b, 4)
	}
	if f.F2 {
		SetBit(&b, 3)
	}
	if f.F1 {
		SetBit(&b, 2)
	}
	if f.F0 {
		SetBit(&b, 1)
	}
	return b
}

func NewCField(b byte) CField {
	var cf = CField{}
	cf.setByte(b)
	return cf
}

// IsFromSlave returns true if the position 7 / index 6 is 0
func (f *CField) IsFromSlave() bool {
	return !f.position7
}

// IsFromMaster returns true if the position 7 / index 6 is 1
func (f *CField) IsFromMaster() bool {
	return f.position7
}

// CFIELD_SND_NKE CField Initialization Slave
var CFIELD_SND_NKE CField = NewCField(0x40)

// CFIELD_SND_UD_0 CField Send User data to Slave
var CFIELD_SND_UD_0 CField = NewCField(0x53)

// CFIELD_SND_UD_1 CField Send User data to Slave
var CFIELD_SND_UD_1 CField = NewCField(0x73)

// CFIELD_REQ_UD2_0 CField Request for Class 2 data
var CFIELD_REQ_UD2_0 CField = NewCField(0x5B)

// CFIELD_REQ_UD2_1 CField Request for Class 2 data
var CFIELD_REQ_UD2_1 CField = NewCField(0x7B)

// CFIELD_REQ_UD1_0 CField Request for Class 1 data
var CFIELD_REQ_UD1_0 CField = NewCField(0x5A)

// CFIELD_REQ_UD1_1 CField Request for Class 1 data
var CFIELD_REQ_UD1_1 CField = NewCField(0x7A)

// CFIELD_RSP_UD_a CField data Transfer decodeFrom Slave to Master after Request
var CFIELD_RSP_UD_a CField = NewCField(0x08)

// CFIELD_RSP_UD_b CField data Transfer decodeFrom Slave to Master after Request
var CFIELD_RSP_UD_b CField = NewCField(0x18)

// CFIELD_RSP_UD_c CField data Transfer decodeFrom Slave to Master after Request
var CFIELD_RSP_UD_c CField = NewCField(0x28)

// CFIELD_RSP_UD_d CField data Transfer decodeFrom Slave to Master after Request
var CFIELD_RSP_UD_d CField = NewCField(0x38)
