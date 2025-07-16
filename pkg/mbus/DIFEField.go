package mbus

import "fmt"

type DIFEField byte

func NewDIFEField(b byte) DIFEField {
	return DIFEField(b)
}

func (df *DIFEField) hex() string {
	return fmt.Sprintf("0x%02x", df)
}

func (df *DIFEField) hasExtension() bool {
	return HasBit(byte(*df), 8)
}

// NextExist means that there is another DIFE field. This is the case when the
// 8th bit is set to 1.
func (df *DIFEField) NextExist() bool {
	return df.hasExtension()
}

// DeviceUnit is the bit in position in 7
func (df *DIFEField) DeviceUnit() bool {
	return HasBit(byte(*df), 7)
}

// StorageNumber is the uint in positions 1-4
func (df *DIFEField) StorageNumber() string {
	return fmt.Sprintf("%v", uint(SliceByte8(byte(*df), 1, 4)))
}

// Tariff is the uint in positions 5-6
func (df *DIFEField) Tariff() string {
	return fmt.Sprintf("%v", uint(SliceByte8(byte(*df), 5, 2)))
}
