package mbus

import (
	"fmt"
)

type VIFField byte

func NewVIFField(b byte) VIFField {
	return VIFField(b)
}

func (f *VIFField) hex() string {
	return fmt.Sprintf("0x%02x", f)
}

func (f *VIFField) hasExtension() bool {
	return HasBit(byte(*f), 8)
}

// VIFEExist means that there is another VIFE field.
// This is the case when the 8th bit is set to 1.
func (f *VIFField) VIFEExist() bool {
	return f.hasExtension()
}

func (f *VIFField) withoutExtension() byte {
	return byte(*f) << 1 >> 1
}

func (f *VIFField) unit() string {
	return VifFields[f.withoutExtension()].Unit
}

func (f *VIFField) exponent() float64 {

	return VifFields[f.withoutExtension()].Exponent
}

// / Find byte with extension bit ?
// func (f *VIFField) name(extensionBit bool) string {
func (f *VIFField) name() string {
	switch f.hasExtension() {
	case true:
		return VifFields[byte(*f)].Name
	default:
		return VifFields[f.withoutExtension()].Name
	}
}
