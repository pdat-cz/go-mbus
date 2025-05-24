package mbus

import (
	"fmt"
)

type DIFField byte

func NewDIFField(b byte) DIFField {
	return DIFField(b)
}

func (df *DIFField) hex() string {
	return fmt.Sprintf("0x%02x", df)
}

// hasExtension bit 8 is 1
func (df *DIFField) hasExtension() bool {
	return HasBit(byte(*df), 8)
}

// DIFEFieldExist means that there is DIFE field. Extension bit 8 is 1
func (df *DIFField) DIFEFieldExist() bool {
	return df.hasExtension()
}

// dataLengthCode Get data Length Code decodeFrom DIF bits 1-4
func (df *DIFField) dataLengthCode() byte {
	//bString := fmt.Sprintf("%08b", df)[4:8]
	b := SliceByte8(byte(*df), 1, 4)
	return b
}

// / Get Type of data decodeFrom DIF bit 5-4
func (df *DIFField) dataType() DIFFieldTypeOfData {
	var sb byte = SliceByte8(byte(*df), 5, 2)
	var dt = DIFFieldTypeOfData(sb)
	return dt
}

// DIFFieldTypeOfData data Type ... maximum, minimum, ...
type DIFFieldTypeOfData byte

// data Type ... maximum, minimum, ...
const (
	INSTANTANEOUS      DIFFieldTypeOfData = 0b00
	MAXIMUM                               = 0b01
	MINIMUM                               = 0b10
	VALUE_DURING_ERROR                    = 0b11
)

func (dtp DIFFieldTypeOfData) string() string {
	switch dtp {
	case INSTANTANEOUS:
		return "INSTANTANEOUS"
	case MAXIMUM:
		return "MAXIMUM"
	case MINIMUM:
		return "MINIMUM"
	case VALUE_DURING_ERROR:
		return "VALUE_DURING_ERROR"
	}
	return "UNDEFINED"
}

func (df *DIFField) dataTypeName() string {
	switch df.dataType() {
	case INSTANTANEOUS:
		return "INSTANTANEOUS"
	case MAXIMUM:
		return "MAXIMUM"
	case MINIMUM:
		return "INSTANTANEOUS"
	case VALUE_DURING_ERROR:
		return "VALUE_DURING_ERROR"
	default:
		return "UNDEFINED"
	}
}

func (df *DIFField) dataLengthName() string {
	return DIFFieldLengthTable[byte(df.dataLengthCode())].name
}

func (df *DIFField) dataLength() int {
	return DIFFieldLengthTable[byte(df.dataLengthCode())].length
}

// DIFFieldLengthRecord Length of data and type ie. 24Int, 32Real, ...
type DIFFieldLengthRecord struct {
	name   string
	length int
}

var DIFFieldLengthTable = map[byte]DIFFieldLengthRecord{
	0b0000: {name: "NO_DATA", length: 0},
	0b1000: {name: "SELECTION_FOR_READOUT", length: 0},
	0b0001: {name: "BIT_8_INTEGER", length: 1},
	0b0010: {name: "BIT_16_INTEGER", length: 2},
	0b0011: {name: "BIT_24_INTEGER", length: 3},
	0b0100: {name: "BIT_32_INTEGER", length: 4},
	0b0101: {name: "BIT_32_REAL", length: 4},
	0b0110: {name: "BIT_48_INTEGER", length: 6},
	0b0111: {name: "BIT_64_INTEGER", length: 8},
	0b1001: {name: "BCD_2_DIGIT", length: 1},
	0b1010: {name: "BCD_4_DIGIT", length: 2},
	0b1011: {name: "BCD_6_DIGIT", length: 3},
	0b1100: {name: "BCD_8_DIGIT", length: 4},
	0b1101: {name: "VARIABLE_LENGTH", length: 0}, // length stored in data field
	0b1110: {name: "BCD_12_DIGIT", length: 6},
	0b1111: {name: "SPECIAL_FUNCTIONS", length: 8},
}
