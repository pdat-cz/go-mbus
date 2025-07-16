package mbus

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// DecodeManufacturerId Decode 2 bytes into 3 ASCII Characters - Manufacturer ID
func DecodeManufacturerId(data []byte) string {

	i := binary.LittleEndian.Uint16(data)

	a1 := ((i >> 10) & 0x001f) + 64
	a2 := ((i >> 5) & 0x001f) + 64
	a3 := (i & 0x001f) + 64
	//fmt.Println(string([]byte{byte(a1)}))
	//fmt.Println(string([]byte{byte(a2)}))
	//fmt.Println(string([]byte{byte(a3)}))

	return string([]byte{byte(a1), byte(a2), byte(a3)})
}

// BoolToInt Convert bool to int: true -> 1, false -> 0
func BoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func StringsToBytes(strings []string) []byte {
	var bytes []byte
	for _, s := range strings {
		h, err := HexStringToByte(s)
		if err == nil {
			bytes = append(bytes, h)
		}
	}
	return bytes
}

// ByteToHexString Convert byte to HEX String ie. 0XA0
func ByteToHexString(b byte) string {
	return fmt.Sprintf("0x%02X", b)
}

// HexStringToByte Convert HEX String to byte
// ie. 0xA0 -> 160
func HexStringToByte(s string) (byte, error) {
	if s == "" {
		return 0, errors.New("String is empty")
	}
	sclean := strings.Replace(s, "0x", "", -1)
	h, _ := hex.DecodeString(sclean)
	return h[0], nil
}

// HexStringToBytes Convert HEX String to byte array
//
// example: "0x01 0x02 0x03" -> []byte{1,2,3}
func HexStringToBytes(s string) []byte {
	if s == "" {
		return []byte{}
	}
	var bytes []byte
	for _, s0 := range strings.Split(s, " ") {

		h, err := HexStringToByte(s0)
		if err == nil {
			bytes = append(bytes, h)
		}
	}

	return bytes
}

func ByteTo2Bits(b byte) string {
	return fmt.Sprintf("0x%02b", b)
}

func ByteTo4Bits(b byte) string {
	return fmt.Sprintf("0x%04b", b)
}

func ByteTo8Bits(b byte) string {
	return fmt.Sprintf("0x%08b", b)
}

// / Convert bytes to HEX String ie. 0XA0
func BytesToHexString(b []byte) string {
	var s string

	for _, b0 := range b {
		s = fmt.Sprintf("%s 0x%02x", s, b0)
	}

	return s
}

// / Print Hex, Int, 8bits
func PrintByte(b byte) {
	fmt.Printf("hex: 0x%02x\t", b)
	fmt.Printf("uint8: %v\t", b)
	fmt.Printf("bits: %08b\n", b)
}

// Return Reversed bytes order
func ReversedBytes(s []byte) []byte {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

/*
*
Create byte decodeFrom String

Example:

	createByte("01001")
*/
func CreateByte(t string) byte {
	var b1 byte = '1'
	l := len(t)
	var b byte = 0 << l
	for i, char := range t {
		if b1 == byte(char) {
			SetBit(&b, uint(l-i)) // backward
		}
	}
	return b
}

/*
Set bit at position. First position is 1

Example:

	var b byte = 0x10
	fmt.Printf("bits: %08b\n", b)
	setBit(&b, 1)
	fmt.Printf("bits: %08b\n", b)
*/
func SetBit(b *byte, pos uint) {
	*b |= (1 << (pos - 1))
}

/*
	Clear bit at position. First position is 1

Example:

	var b1 byte = 0x01
	fmt.Printf("bits: %08b\n", b1)
	clearBit(&b1, 1)
	fmt.Printf("bits: %08b\n", b1)
*/
func ClearBit(b *byte, pos uint) {
	*b &= ^(1 << (pos - 1))
}

/*
	Is a bit at position ? First position is 1

Example:

	var b byte = 0x01
	fmt.Printf("bits: %08b\n", b)
	fmt.Printf("has a bit at position %v ? %v\n", 1, hasBit(b, 1))
	fmt.Printf("has a bit at position %v ? %v\n", 2, hasBit(b, 2))
*/
func HasBit(b byte, pos uint) bool {
	val := b & (1 << (pos - 1))
	return (val > 0)
}

// / Slice decodeFrom 8bit byte
// / Parameters:
// / 	start ... position, first postion is 1
// /		length ... number of bits
// /
// / Example:
// / SliceByte8(byte(0b001101), 2, 2) -> 0b10
func SliceByte8(b byte, startPosition int, length int) byte {
	left := 8 - (startPosition - 1) - length
	right := 8 - length
	return b << left >> right
}

func From8int(b []byte, exponent float64) string {
	fmt.Println(BytesToHexString(b))
	fmt.Println(exponent)
	data := binary.LittleEndian.Uint16(b)
	value := fmt.Sprintf("%f", float64(data)*exponent)
	return value
}

func From16int(b []byte, exponent float64) string {
	data := binary.LittleEndian.Uint16(b)
	value := fmt.Sprintf("%f", float64(data)*exponent)
	return value
}

func From24int(b []byte, exponent float64) string {
	//,Append change source, so  i am going by this per partes way
	bytes32 := []byte{b[0], b[1], b[2], 0x00}
	data := binary.LittleEndian.Uint32(bytes32)
	value := fmt.Sprintf("%f", float64(data)*exponent)
	return value
}

func From32int(b []byte, exponent float64) string {
	data := binary.LittleEndian.Uint32(b)
	value := fmt.Sprintf("%f", float64(data)*exponent)
	return value
}

func From48int(b []byte, exponent float64) string {
	bytes64 := []byte{b[0], b[1], b[2], b[3], b[4], b[5], 0x00, 0x00}
	data := binary.LittleEndian.Uint64(bytes64)
	value := fmt.Sprintf("%f", float64(data)*exponent)
	return value
}

func From64int(b []byte, exponent float64) string {
	data := binary.LittleEndian.Uint64(b)
	value := fmt.Sprintf("%f", float64(data)*exponent)
	return value
}

// From16intTimePoint VIF 0x62 Time Point (date)
// Type G: Compound CP16: Date
func From16intTimePoint(b []byte) (string, error) {
	// Type G: Compound CP16: Date
	// day UI5 [bit 0 to 4] <1 to 31>
	// month UI4 [bit 8 to 11] <1 to 12>
	// year UI7 [bit 6 to 7, 12 to 16] <0 to 99>

	// Check length
	if len(b) != 2 {
		return "", fmt.Errorf("From16intTimePoint: length of byte array must be 2 and not %v", len(b))
	}

	value := binary.LittleEndian.Uint16(b)
	day := int(value & 0x1F)
	month := int((value >> 8) & 0x0F)
	baseYear := 100 + int(((b[0]&0xE0)>>5)|
		((b[1]&0xF0)>>1))

	year := baseYear + 1900

	if year < 1980 {
		year += 100
	}

	output := fmt.Sprintf("%04d-%02d-%02d",
		year,
		month,
		day)

	return output, nil
}

// From32intTimePoint VIF 0x6D (date/time)
func From32intTimePoint(b []byte) (string, error) {
	// TYPE I = Compound CP32: Date and Time
	// Extract the date and time from the uint32
	//min: UI6 [bit 0 to 5] <0 to 59>
	//B1[6] {reserved}: <0>
	//IV: B1[7] {time invalid}: IV<0> := valid IV<0> := invalid
	//hour: UI5 [8 to 12] <0 to 23>
	//reserved: UI7 [bit 13 to 14] <0 to 1>
	//SU: B1[15] {summer time}: <0> := standard time, SU<1> := summer time
	//base year: UI7[bit 21 to 23,28 to 33] <0 to 99>
	//month: UI4 [bit 24 to 27] <1 to 12>
	//year = 1990 + base year
	//<0> B1[14] {reserved}: <0>

	// Check length
	if len(b) != 4 {
		return "", fmt.Errorf("From32intTimePoint: length of byte array must be 4 and not %v", len(b))
	}

	// value := uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])

	minute := int(b[0] & 0x3f)
	hour := int(b[1] & 0x1f)
	day := int(b[2] & 0x1f)
	mon := int(b[3] & 0x0f)
	baseYear := int(
		((b[2] & 0xe0) >> 5) |
			((b[3] & 0xf0) >> 1))

	year := 1900 + baseYear
	if year < 1980 {
		year += 100
	}
	// Return in RFC3339 format UTC
	output := fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:%02dZ",
		year,
		mon,
		day,
		hour,
		minute,
		0)

	return output, nil
}

func From32real(b []byte, exponent float64) string {
	data := binary.LittleEndian.Uint32(b)
	dataF := math.Float32frombits(data)
	value := float64(dataF) * exponent
	return fmt.Sprintf("%f", value)
}

// DecodedTime Decoded Time Point
type DecodedTime struct {
	sec   int // 0-59
	min   int // 0-59
	hour  int // 0-23
	mday  int // 1-31
	mon   int // 1-12
	year  int // year since 1900
	wday  int // 0-6
	yday  int // 0-365
	isdst int // daylight saving time flag. true if DST is in effect
}

// DecodeFrom Decode time form bytes
func (dt *DecodedTime) DecodeFrom(b []byte) {
	/*
		if len(b) == 6 {
			// TYPE I = Compound CP48: Date and Time
			dt.sec = int(b[0] & 0x3f)
			dt.min = int(b[1] & 0x3f)
			dt.hour = int(b[2] & 0x1f)
			dt.mday = int(b[3] & 0x1f)
			dt.mon = int(b[4] & 0x0f)
			dt.year = int(b[2] & 0x7f)
			if (b[0] & 0x40) != 0 {
				dt.isdst = 1
			}
		}
	*/
	if len(b) == 4 {
		// TYPE I = Compound CP32: Date and Time
		dt.min = int(b[0] & 0x3f)

		dt.hour = int(b[1] & 0x1f)

		dt.mday = int(b[2] & 0x1f)

		dt.mon = int(b[3] & 0x0f)

		// Month is 0 based

		dt.year = 1900 + int(
			((b[2]&0xe0)>>5)|
				((b[3]&0xf0)>>1))
		// TODO: day saving time calculation ?
		dt.isdst = 0

	}
	if len(b) == 2 {
		// TYPE I = Compound CP48: Date and Time
		dt.mday = int(b[0] & 0x1f)
		dt.mon = int(b[1] & 0x0f)
		dt.year = 1900 + int(
			((b[0]&0xe0)>>5)|
				((b[1]&0xf0)>>1),
		)
	}
}

func DecodeDateTimeCP32(cp32Value []byte) (time.Time, error) {
	if len(cp32Value) != 4 {
		return time.Time{}, fmt.Errorf("cp32Value must be 4 bytes long")
	}
	// Convert the first 4 bytes of cp32Value to uint32
	value := uint32(cp32Value[0])<<24 | uint32(cp32Value[1])<<16 | uint32(cp32Value[2])<<8 | uint32(cp32Value[3])

	// Extract the date and time from the uint32
	//min: UI6 [bit 0 to 5] <0 to 59>
	//B1[6] {reserved}: <0>
	//IV: B1[7] {time invalid}: IV<0> := valid IV<0> := invalid
	//hour: UI5 [8 to 12] <0 to 23>
	//hundred years: UI7 [bit 13 to 14] <0 to 1>
	//SU: B1[15] {summer time}: <0> := standard time, SU<1> := summer time
	//base year: UI7[bit 21 to 23,28 to 33] <0 to 99>
	//month: UI4 [bit 24 to 27] <1 to 12>
	//year = 1990 + base year
	//<0> B1[14] {reserved}: <0>

	// bit index and type
	//var bitType = map[int]string{
	//	0:  "minute",
	//	6:  "reserved",
	//	7:  "timeValid",
	//	8:  "hour",
	//	13: "hundredYears",
	//	15: "summerTime",
	//	16: "day",
	//	21: "year",
	//	24: "month",
	//	28: "year",
	//}

	// minute bits 0 to 5
	minute := value & 0x3f

	// hour bits 8 to 12
	hour := (value & 0x1F00) >> 8 & 0x1f

	// day bits 16 to 20
	day := (value & 0x001F0000) >> 16 & 0x1f

	// month bits 24 to 27
	month := (value & 0x0F000000) >> 24 & 0x0f

	// base year bits 21 to 23 and 28 to 33
	baseYear := (value & 0x00E00000) >> 21 & 0x07
	baseYear |= (value & 0x3F000000) >> 28 & 0x3f

	// hundred years bits 13 to 14
	hundredYears := (value & 0x00006000) >> 13 & 0x03

	// year = 1990 + hundred years * 100 + base year
	year := 1990 + hundredYears*100 + baseYear

	// tinmeValid bit 7
	//timeValid := (value & 0x00000080) >> 7 & 0x01

	// summerTime bit 15
	//summerTime := (value & 0x00008000) >> 15 & 0x01

	//yearVersion2 := ((cp32Value[2] & 0xE0) >> 5) | ((cp32Value[3] & 0xF0) >> 1)

	// Create a time.Time from the extracted date and time
	return time.Date(int(year), time.Month(month), int(day), int(hour), int(minute), 0, 0, time.UTC), nil

}

// FromBCD convert BCD to string
func FromBCD(bcd []byte, exponent float64) string {
	var number string
	for i := len(bcd) - 1; i >= 0; i-- {
		b := bcd[i]
		number += fmt.Sprintf("%02X", b)
	}
	value, err := strconv.Atoi(number)
	if err != nil {
		fmt.Printf("Error to convet %s. %s", number, err)
	}
	valueF := float64(value) * exponent
	return fmt.Sprintf("%f", valueF)
}

// / Are []bytes equall ? Position can be different
func BytesAreEqual(x []byte, y []byte) bool {

	if x == nil && y == nil {
		return true
	}
	if x == nil || y == nil {

		return false
	}
	if len(x) != len(y) {

		return false
	}

	for _, v := range x {
		if !IncludesByte(y, v) {
			return false
		}
	}

	for _, v := range y {
		if !IncludesByte(x, v) {
			return false
		}
	}

	return true
}

// / If byte is in []byte, then return true
func IncludesByte(m []byte, b byte) bool {
	for _, x := range m {
		if x == b {
			return true
		}
	}
	return false
}

// / If string is in []string, then return true
func IncludesString(m []string, s string) bool {
	for _, x := range m {
		if x == s {
			return true
		}
	}
	return false
}

func IncludeKey(m map[string]interface{}, s string) bool {
	for key, _ := range m {
		if key == s {
		}
		return true
	}
	return false
}

// cleanString remove all unwanted code
func cleanString(s string) string {
	var replacer = strings.NewReplacer(
		"\r\n", "",
		"\r", "",
		"\n", "",
		"\v", "",
		"\f", "",
		"\u0008", "",
		"\u0085", "",
		"\u2028", "",
		"\u2029", "",
	)

	return replacer.Replace(s)
}

// timeDuration parse 10 or 15s to time duration. If it is not possible to parsee, return 0
func timeDuration(s string) time.Duration {

	number, err := strconv.Atoi(s)
	if err != nil {
		// It is not integer
		duration, err := time.ParseDuration(s)
		if err != nil {
			// Cannot parse, send 0
			return time.Duration(0)
		}

		// It is in string ie. 15s
		return duration
	}

	// It is int. Convert to seconds
	return time.Duration(number) * time.Second
}
