package mbus

import "fmt"

type VIFEField struct {
	b   byte
	vif byte // VIF Code 0xFB, or 0XFD
}

func NewVIFEField(b byte, vif byte) VIFEField {
	return VIFEField{b, vif}
}

func (f *VIFEField) hex() string {
	return fmt.Sprintf("0x%02x", f.b)
}

func (f *VIFEField) hasExtension() bool {
	return HasBit(f.b, 8)
}

// ExistNextVife means that there is another VIFE field.
// This is the case when the 8th bit is set to 1.
func (f *VIFEField) ExistNextVife() bool {
	return f.hasExtension()
}

func (f *VIFEField) withoutExtension() byte {
	return f.b << 1 >> 1
}

func (f *VIFEField) unit() string {
	switch f.vif {
	case 0xFD:
		return VifVifeFdFields[f.withoutExtension()].Unit
	case 0xFB:
		return VifVifeFbFields[f.withoutExtension()].Unit
	default:
		return VifVifeFields[f.withoutExtension()].Unit
	}
}

func (f *VIFEField) exponent() float64 {
	switch f.vif {
	case 0xFB:
		return VifVifeFbFields[f.withoutExtension()].Exponent
	case 0xFD:
		return VifVifeFdFields[f.withoutExtension()].Exponent
	default:
		return VifVifeFields[f.withoutExtension()].Exponent
	}

}

func (f *VIFEField) name() string {
	switch f.vif {
	case 0xFB:
		return VifVifeFbFields[f.withoutExtension()].Name
	case 0xFD:
		return VifVifeFdFields[f.withoutExtension()].Name
	default:
		return VifVifeFields[f.withoutExtension()].Name
	}
}
