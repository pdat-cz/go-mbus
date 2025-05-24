package mbus

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// LFrame represents an M-Bus telegram in the Long Frame format.
// It contains the raw data of the telegram and provides methods to access and interpret its fields.
type LFrame struct {
	data []byte
}

type LFrameRecord struct {
	DIF  byte   `yaml:"DIF" json:"DIF"`
	DIFE []byte `yaml:"DIFE" json:"DIFE"`
	VIF  byte   `yaml:"VIF" json:"VIF"`
	VIFE []byte `yaml:"VIFE" json:"VIFE"`
	// Manufacturer VIFE
	VIFEM []byte `yaml:"VIFEM" json:"VIFEM"`
	Value string `yaml:"value" json:"value"`
	//
	Function string  `yaml:"function" json:"function"`
	Unit     string  `yaml:"unit" json:"unit"`
	Name     string  `yaml:"name" json:"name"`
	Exponent float64 `yaml:"exponent" json:"exponent"`
}

// LFrameParsed Parsed, records in DIF, DIFE, VIF, VIFE, ...
type LFrameParsed struct {
	IdentificationNumber string `yaml:"identification_number" json:"identification_number"`
	Manufacturer         string `yaml:"manufacturer" json:"manufacturer"`
	Version              uint   `yaml:"version" json:"version"`
	Medium               string `yaml:"medium" json:"medium"`
	AccessNumber         uint   `yaml:"access_number" json:"access_number"`
	Status               string `yaml:"status" json:"status"`
	Model                string `yaml:"model" json:"model"`
	Address              uint8  `yaml:"address" json:"address"`
	Signature            []byte `yaml:"signature" json:"signature"`
	Records              map[int]LFrameRecord
}

func NewLFrame(data []byte) LFrame {
	return LFrame{data: data}
}

// Verify Frame
func (lf *LFrame) Verify() (bool, error) {
	if len(lf.data) < 6 {
		return false, fmt.Errorf("data length is too short")
	}
	if lf.data[0] != 0x68 {
		return false, fmt.Errorf("data does not start with 0x68")
	}
	if lf.data[1] != lf.data[2] {
		return false, fmt.Errorf("data length does not match")
	}
	if lf.data[3] != 0x68 {
		return false, fmt.Errorf("data does not start with 0x68")
	}
	// Checksum

	return true, nil
}

// DataHeaderLField position 1
func (lf *LFrame) DataHeaderLField() byte {
	return lf.data[1]
}

// CField is in index 4
func (lf *LFrame) CField() (CField, error) {
	// Check length
	if len(lf.data) < 4 {
		return CField{}, errors.New("data length is too short")
	}
	return NewCField(lf.data[4]), nil
}

// CIField at index 6
func (lf *LFrame) CIField() (byte, string) {
	code := lf.data[6]
	ciField := CIField(code)
	return code, ciField.String()
}

// IsFromSlave from CField position 7
// TODO: Is this the same as IsFromMaster?
func (lf *LFrame) IsFromSlave() bool {
	f, err := lf.CField()
	if err != nil {
		return false
	}
	return f.IsFromSlave()
}

// IsFromMaster from CField position 7
// TODO: Is this the same as IsFromSlave?
func (lf *LFrame) IsFromMaster() bool {
	f, err := lf.CField()
	if err != nil {
		return false
	}
	return f.IsFromMaster()
}

// AField at index 5
func (lf *LFrame) AField() (AField, error) {
	// Check length
	if len(lf.data) < 5 {
		return NewAField(0), errors.New("data length is too short")
	}
	return NewAField(lf.data[5]), nil
}

// SlaveAddress or error
func (lf *LFrame) SlaveAddress() (uint8, error) {
	af, err := lf.AField()
	if err != nil {
		return 0, err
	}
	sa, errSa := af.SlaveAddress()
	if errSa != nil {
		return 0, errSa
	}
	return sa, nil
}

// IdentificationNumber position 7-10
func (lf *LFrame) IdentificationNumber() string {
	var bcd []byte = lf.data[7:11]
	var number string
	for i := len(bcd) - 1; i >= 0; i-- {
		b := bcd[i]
		number += fmt.Sprintf("%02X", b)
	}
	return number
}

// Manufacturer position 11-12
func (lf *LFrame) Manufacturer() string {
	return DecodeManufacturerId(lf.data[11:13])
}

// Version position 13
func (lf *LFrame) Version() uint {
	return uint(lf.data[13])
}

// Medium of device position 14
// Return: (MediumType, string)
func (lf *LFrame) Medium() MediumType {
	mediumByte := lf.data[14]
	medium := MediumType(mediumByte)
	return medium
}

// AccessNumber incremental how many is accessed ... position 15
func (lf *LFrame) AccessNumber() uint {
	return uint(lf.data[15])
}

func (lf *LFrame) Status() (byte, error) {
	b := lf.data[16]
	if HasBit(b, 0) && HasBit(b, 1) {
		return b, nil
	}
	if HasBit(b, 0) && HasBit(b, 1) {
		return b, errors.New("Reserved, .. not defined")
	}

	if HasBit(b, 0) {
		return b, errors.New("Device Is Busy")
	}

	return b, nil
}

// Signature ... position 17,18
func (lf *LFrame) Signature() []byte {
	return lf.data[17:19]
}

// VariableDataRecord First Record start in index 20
func (lf *LFrame) VariableDataRecord(firstPosition1 int) (LFrameRecord, int, error) {

	// Helper function to discard string values
	discard := func(s string) {}

	//var output = make(map[string]interface{})
	record := LFrameRecord{}
	record.DIFE = []byte{}
	record.VIFE = []byte{}
	record.VIFEM = []byte{}

	index := firstPosition1 - 1

	// 1. Get DIF
	// Check if exist index
	if len(lf.data) < index {
		return record, index, errors.New(fmt.Sprintf("Index out of range. data length: %v, index: %v", len(lf.data), index))
	}

	var dif = NewDIFField(lf.data[index])
	record.DIF = byte(dif)

	// 2. Get DIFE if existed, and if so, get all DIFE
	if dif.hasExtension() {
		parseDife := true
		for ok := true; ok; ok = parseDife {
			index += 1
			dife := NewDIFEField(lf.data[index])

			parseDife = dife.hasExtension()
			record.DIFE = append(record.DIFE)
		}

	}

	// VIF
	index += 1
	var vif VIFField = NewVIFField(lf.data[index])
	record.VIF = byte(vif)
	// MANUFACTURER SPECIFIC VIF
	// END OF USER DATA
	if record.DIF == 0x0f || record.DIF == 0x1f {
		// SPECIAL FUNCTION - Manufacturer specific data structures to end of user data
		endOfData := lf.LastDataPosition()
		// Skip the value bytes as they're not used
		return record, endOfData, nil
	}

	// SWITCH VIF
	switch vif {
	case 0xFD:
		// VIF is in next byte
		index += 1
		vif = NewVIFField(lf.data[index])
		record.VIF = byte(vif)
		record.Unit = vif.unit()
		record.Name = vif.name()
	case
		0xFB:
		// VIF is in next byte
		index += 1
		vif = NewVIFField(lf.data[index])
		record.VIF = byte(vif)

		if vif.VIFEExist() {
			parseVife := true
			for ok := true; ok; ok = parseVife {
				index += 1
				vife := VIFEField{lf.data[index], 0xFB}
				// MANUFACTURER VIFE
				// If has Extension, then find  next VIFE
				parseVife = vife.hasExtension()

				if vife.b == 0xFF {
					// Manufacturer specific. Next byte is Manufacturer specific VIFE
					index += 1
					manufacturerVIFE := lf.data[index]
					// MANUFACTURER VIFE
					record.VIFEM = append(record.VIFEM, manufacturerVIFE)
					parseVife = HasBit(manufacturerVIFE, 8)
				} else {
					record.VIFE = append(record.VIFE, vife.b)
					record.Unit = vife.unit()
					record.Name = vife.name()
				}
			}
		}
	case 0x7f, 0xff:
		// Manufacturer specific. Next byte is Manufacturer specific VIFE
		index += 1
		manufacturerVIFE := lf.data[index]
		record.VIFEM = append(record.VIFEM, manufacturerVIFE)

	default:
		record.VIF = byte(vif)
		// Unit and Unit Name
		record.Unit = vif.unit()
		record.Name = vif.name()
		//
		if vif.hasExtension() {
			parseVife := true
			for ok := true; ok; ok = parseVife {
				index += 1
				vife := NewVIFEField(lf.data[index], byte(vif))

				// If it has Extension, then find  next VIFE
				parseVife = vife.hasExtension()

				if vife.b == 0xFF {
					// Manufacturer specific. Next byte is Manufacturer specific VIFE
					index += 1
					manufacturerVIFE := lf.data[index]
					record.VIFEM = append(record.VIFEM, manufacturerVIFE)
					parseVife = HasBit(manufacturerVIFE, 8)
				} else {
					record.VIFE = append(record.VIFE, vife.b)
					record.Unit = vife.unit()
					record.Name = vife.name()
				}
			}
		}
	}

	// GET data
	index += 1

	var value string
	var exponent float64

	// True VIF is next byte after VIF 0xFB, 0XFD
	if vif == 0xFD || vif == 0xFB {
		index += 1
		trueVIF := NewVIFField(lf.data[index])
		exponent = trueVIF.exponent()
	} else {
		exponent = vif.exponent()
	}

	switch dif.dataLengthName() {
	case "BIT_8_INTEGER":
		valueLengthBytes := dif.dataLength()
		dataOfRecord := lf.data[index : index+valueLengthBytes]
		// valueBytes = BytesToHexString(dataOfRecord) - removed unused assignment
		value = From8int(dataOfRecord, exponent)
		index += dif.dataLength()
	case "BIT_16_INTEGER":
		valueLengthBytes := dif.dataLength()
		dataOfRecord := lf.data[index : index+valueLengthBytes]
		_ = BytesToHexString(dataOfRecord) // Intentionally discarded

		if vif == 0x6c {
			/// TIME POINT (date)
			value, _ = From16intTimePoint(dataOfRecord)
		} else {
			value = From16int(dataOfRecord, exponent)
		}
		index += dif.dataLength()
	case "BIT_24_INTEGER":
		valueLengthBytes := dif.dataLength()
		dataOfRecord := lf.data[index : index+valueLengthBytes]
		discard(BytesToHexString(dataOfRecord))
		value = From24int(dataOfRecord, exponent)
		index += dif.dataLength()
	case "BIT_32_REAL":
		valueLengthBytes := dif.dataLength()
		dataOfRecord := lf.data[index : index+valueLengthBytes]
		discard(BytesToHexString(dataOfRecord))
		value = From32real(dataOfRecord, exponent)
		index += dif.dataLength()
	case "BIT_32_INTEGER":
		valueLengthBytes := dif.dataLength()
		dataOfRecord := lf.data[index : index+valueLengthBytes]
		discard(BytesToHexString(dataOfRecord))

		if vif == 0x6d {
			/// TIME POINT (date/time)
			/// START (date/time)
			/// Battery change (date/time)

			value, _ = From32intTimePoint(dataOfRecord)
		} else {
			value = From32int(dataOfRecord, exponent)
		}

		// output["value"] = value
		index += dif.dataLength()
	case "BIT_48_INTEGER":
		valueLengthBytes := dif.dataLength()
		dataOfRecord := lf.data[index : index+valueLengthBytes]
		discard(BytesToHexString(dataOfRecord))
		value = From48int(dataOfRecord, exponent)
		// output["value"] = value
		index += dif.dataLength()
	case "BIT_64_INTEGER":
		valueLengthBytes := dif.dataLength()
		dataOfRecord := lf.data[index : index+valueLengthBytes]
		discard(BytesToHexString(dataOfRecord))
		value = From64int(dataOfRecord, exponent)
		// output["value"] = value
		index += dif.dataLength()
	case "BCD_2_DIGIT",
		"BCD_4_DIGIT",
		"BCD_6_DIGIT",
		"BCD_8_DIGIT",
		"BCD_12_DIGIT":
		valueLengthBytes := dif.dataLength()
		dataOfRecord := lf.data[index : index+valueLengthBytes]
		discard(BytesToHexString(dataOfRecord))
		value = FromBCD(dataOfRecord, exponent)
		index += dif.dataLength()
	case "VARIABLE_LENGTH":
		lvar := lf.data[index]
		//fmt.Printf("LVAR: 0x%02x\n", lvar)
		//fmt.Println(bytesToHexString(lf.data[index : index+int(lvar)]))
		switch true {
		case lvar <= 0xBF:
			// VARIABLE ASCII
			dataOfRecord := lf.data[index : index+1+int(lvar)]
			asciiR := ReversedBytes(dataOfRecord)
			discard(BytesToHexString(dataOfRecord))
			value = string(asciiR)
			index += int(lvar) + 1

		case lvar >= 0xC0 && lvar <= 0xCF:
			// POSITIVE BCD = (LVAR - 0xC0)
			length := int(lvar - 0xC0)
			bytes := lf.data[index : index+length]
			discard(BytesToHexString(bytes))
			value = FromBCD(bytes, exponent)

		case lvar >= 0xD0 && lvar <= 0xDF:
			// NEGATIVE BCD = (LVAR - 0xD0)
			length := int(lvar - 0xD0)
			bytes := lf.data[index : index+length]
			discard(BytesToHexString(bytes))
			value = FromBCD(bytes, exponent)

		case lvar >= 0xE0 && lvar <= 0xEF:
			// Binary number = (LVAR - 0xE0)
			length := int(lvar - 0xE0)
			bytes := lf.data[index : index+length]
			discard(BytesToHexString(bytes))

		case lvar >= 0xF0 && lvar <= 0xFA:
			// Floating point number = (LVAR - 0xE0)
			length := int(lvar - 0xF0)
			bytes := lf.data[index : index+length]
			discard(BytesToHexString(bytes))
		}

	}
	record.Value = strings.TrimSpace(value)
	record.Exponent = exponent
	record.Function = dif.dataTypeName()
	// convert decodeFrom index starting 0, to position starting at 1
	nextPosition1 := index + 1
	//fmt.Printf("--- [END] DATA ----\n--- next position %v\n", nextPosition1)
	return record, nextPosition1, nil
}

// StopByteIndex Position of stop byte
func (lf *LFrame) StopByteIndex() int {
	// The stop byte is the last byte in the data
	return len(lf.data) - 1
}

// LastDataPosition Position of last data byte. Next position is Checksum and then Stop byte
func (lf *LFrame) LastDataPosition() int {
	// The last data byte is the second-to-last byte in the data
	// (the last byte is the stop byte, and the second-to-last is the checksum)
	return len(lf.data) - 2
}

func (lf *LFrame) Records() (map[int]LFrameRecord, error) {

	records := make(map[int]LFrameRecord)

	// Check if stop byte is 0x16
	if byte(lf.data[lf.StopByteIndex()]) != byte(0x16) {
		err := errors.New(fmt.Sprintf("Bad length of data in LField. LField: 0x%02x, stop bit founded: 0x%02x and I should find 0x16\n", lf.DataHeaderLField(), lf.data[lf.StopByteIndex()]))
		return records, err
	}
	// data start in index 20
	position := 20
	recordNumber := 0
	var err error

	// Go through all data
	for ok := true; ok; ok = position < lf.LastDataPosition() {
		record := LFrameRecord{}
		record, position, err = lf.VariableDataRecord(position)
		if err != nil {
			err := errors.New(fmt.Sprintf("Error parse data start at:%v  error:%s", position, err))
			return records, err
		}
		records[recordNumber] = record
		recordNumber += 1
	}

	return records, nil
}

func (lf *LFrame) parse() (LFrameParsed, error) {
	var err error
	/// [START] Header
	normalized := LFrameParsed{}
	normalized.Records = make(map[int]LFrameRecord)
	normalized.IdentificationNumber = lf.IdentificationNumber()
	normalized.Manufacturer = lf.Manufacturer()
	normalized.Medium = lf.Medium().String()
	normalized.Version = lf.Version()
	normalized.AccessNumber = lf.AccessNumber()
	normalized.Address, err = lf.SlaveAddress()
	normalized.Signature = lf.Signature()

	/// [END] Header

	// [START] Records

	records, err := lf.Records()
	if err != nil {
		return normalized, err
	}
	normalized.Records = records

	return normalized, nil
}

// normalizedPayload return normalized records in json format
func (lf *LFrame) json(pretty bool) (string, error) {
	parsed, err := lf.parse()
	if err != nil {
		return "", err
	}
	var output []byte
	if pretty {
		output, err = json.MarshalIndent(parsed, "", "  ")
	} else {
		output, err = json.Marshal(parsed)
	}
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// MediumType type of medium like electricity, gas, water, etc
type MediumType byte

const (
	OTHER         MediumType = 0x00
	OIL                      = 0x01
	ELECTRICITY              = 0x02
	GAS                      = 0x03
	HEAT_OUT                 = 0x04
	STEAM                    = 0x05
	HOT_WATER                = 0x06
	WATER                    = 0x07
	HEAT_COST                = 0x08
	COMPR_AIR                = 0x09
	COOL_OUT                 = 0x0A
	COOL_IN                  = 0x0B
	HEAT_IN                  = 0x0C
	HEAT_COOL                = 0x0D
	BUS                      = 0x0E
	UNKNOWN                  = 0x0F
	IRRIGATION               = 0x10
	WATER_LOGGER             = 0x11
	GAS_LOGGER               = 0x12
	GAS_CONV                 = 0x13
	COLORIFIC                = 0x14
	BOIL_WATER               = 0x15
	COLD_WATER               = 0x16
	DUAL_WATER               = 0x17
	PRESSURE                 = 0x18
	ADC                      = 0x19
	SMOKE                    = 0x1A
	ROOM_SENSOR              = 0x1B
	GAS_DETECTOR             = 0x1C
	BREAKER_E                = 0x20
	VALVE                    = 0x21
	CUSTOMER_UNIT            = 0x25
	WASTE_WATER              = 0x28
	GARBAGE                  = 0x29
	SERVICE_UNIT             = 0x30
	RC_SYSTEM                = 0x36
	RC_METER                 = 0x37
)

func (m MediumType) String() string {
	switch m {
	case OTHER:
		return "Other"
	case OIL:
		return "Oil"
	case ELECTRICITY:
		return "ELECTRICITY"
	case GAS:
		return "GAS"
	case HEAT_OUT:
		return "Heat"
	case STEAM:
		return "STEAM"
	case HOT_WATER:
		return "HOT_WATER"
	case WATER:
		return "WATER"
	case HEAT_COST:
		return "Heat Cost Allocator"
	case COMPR_AIR:
		return "Compressed Air"
	case COOL_OUT:
		return "Cooling load meter OUT"
	case COOL_IN:
		return "Cooling load meter IN"
	case HEAT_IN:
		return "Heat: inlet"
	case HEAT_COOL:
		return "Heat / Cooling load meter"
	case BUS:
		return "Bus / System"
	case UNKNOWN:
		return "Unknown Medium"
	case IRRIGATION:
		return "Irrigation Water"
	case WATER_LOGGER:
		return "Water data logger"
	case GAS_LOGGER:
		return "Gas data logger"
	case GAS_CONV:
		return "Gas converter"
	case COLORIFIC:
		return "Heat Value"
	case BOIL_WATER:
		return "Hot Water (>=90°C) "
	case COLD_WATER:
		return "COLD_WATER"
	case DUAL_WATER:
		return "DUAL_WATER"
	case PRESSURE:
		return "PRESSURE"
	case ADC:
		return "A/D Converter"
	case SMOKE:
		return "SMOKE"
	case ROOM_SENSOR:
		return "ROOM_SENSOR"
	case GAS_DETECTOR:
		return "GAS_DETECTOR"
	case BREAKER_E:
		return "Breaker (Electricity)"
	case VALVE:
		return "Valve (Gas or Water)"
	case CUSTOMER_UNIT:
		return "Customer Unit (Display)"
	case WASTE_WATER:
		return "WASTE_WATER"
	case GARBAGE:
		return "GARBAGE"
	case SERVICE_UNIT:
		return "SERVICE_UNIT"
	case RC_SYSTEM:
		return "Radio Control Unit (System Side)"
	case RC_METER:
		return "Radio Control Unit (Meter Side)"
	}
	return "- undefined -"
}

func (lf *LFrame) setBytes(data []byte) error {
	lf.data = data
	return nil
}

func (lf *LFrame) description() string {
	var output string
	return output
}
