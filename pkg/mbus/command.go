package mbus

import (
	"github.com/tarm/serial"
	"io"
	"time"
)

const (
	rtuMaxSize               = 256
	FRAME_ACK_START     byte = 0xE5
	FRAME_SHORT_START   byte = 0x10
	FRAME_CONTROL_START byte = 0x68
	FRAME_LONG_START    byte = 0x68
	FRAME_STOP          byte = 0x16
)

func COMMAND_SND_NKE(deviceAddress uint) []byte {
	var b = []byte{}
	var cf = CFIELD_SND_NKE.getByte()
	var ad = byte(deviceAddress)
	var crc = cf + ad // Arithmetic checksum
	b = append(b, FRAME_SHORT_START)
	b = append(b, cf)
	b = append(b, ad)
	b = append(b, crc)
	b = append(b, FRAME_STOP)
	return b
}

// COMMAND_REQ_UD2
func COMMAND_REQ_UD2(deviceAddress uint) []byte {
	var b []byte
	var cf = CFIELD_REQ_UD2_0.getByte()
	var ad = byte(deviceAddress)
	var crc = cf + ad // Arithmetic checksum
	b = append(b, FRAME_SHORT_START)
	b = append(b, cf)
	b = append(b, ad)
	b = append(b, crc)
	b = append(b, FRAME_STOP)
	return b
}

// =============================================================================
// Write Operations (SND_UD) - EN 13757-3
// =============================================================================

// BuildLongFrame constructs a long frame with proper checksum
// Format: 68 L L 68 | C A CI | DATA | CS 16
func BuildLongFrame(cField, aField, ciField byte, data []byte) []byte {
	length := byte(3 + len(data)) // C + A + CI + data length
	frame := make([]byte, 0, 6+len(data)+2)

	// Header
	frame = append(frame, FRAME_LONG_START)
	frame = append(frame, length)
	frame = append(frame, length)
	frame = append(frame, FRAME_LONG_START)

	// Control fields
	frame = append(frame, cField)
	frame = append(frame, aField)
	frame = append(frame, ciField)

	// Data
	frame = append(frame, data...)

	// Checksum: sum of C + A + CI + DATA
	var checksum byte
	checksum = cField + aField + ciField
	for _, b := range data {
		checksum += b
	}
	frame = append(frame, checksum)
	frame = append(frame, FRAME_STOP)

	return frame
}

// BuildControlFrame constructs a control frame (long frame with no data)
// Format: 68 03 03 68 | C A CI | CS 16
func BuildControlFrame(cField, aField, ciField byte) []byte {
	return BuildLongFrame(cField, aField, ciField, nil)
}

// COMMAND_SND_UD builds a generic SND_UD (Send User Data) command
// This is the base function for all write operations
func COMMAND_SND_UD(deviceAddress uint, ciField byte, data []byte) []byte {
	cf := CFIELD_SND_UD_0.getByte()
	return BuildLongFrame(cf, byte(deviceAddress), ciField, data)
}

// COMMAND_SET_PRIMARY_ADDRESS sets the primary address of a slave
// Uses broadcast address 0xFE (254) to address slave, with DIF=0x01, VIF=0x7A
// Example: 68 06 06 68 | 53 FE 51 | 01 7A 08 | 25 16 (set address to 8)
func COMMAND_SET_PRIMARY_ADDRESS(newAddress uint) []byte {
	data := []byte{
		0x01, // DIF: 8-bit integer
		0x7A, // VIF: Bus address
		byte(newAddress),
	}
	return COMMAND_SND_UD(0xFE, byte(CiFieldDataSend), data)
}

// COMMAND_SET_PRIMARY_ADDRESS_FROM sets the primary address from a known current address
func COMMAND_SET_PRIMARY_ADDRESS_FROM(currentAddress, newAddress uint) []byte {
	data := []byte{
		0x01, // DIF: 8-bit integer
		0x7A, // VIF: Bus address
		byte(newAddress),
	}
	return COMMAND_SND_UD(currentAddress, byte(CiFieldDataSend), data)
}

// COMMAND_SET_IDENTIFICATION sets the complete identification of a slave
// ID: 4-byte identification number (BCD)
// manufacturer: 2-byte manufacturer ID
// generation: device generation/version
// medium: device medium type
// Example: 68 0D 0D 68 | 53 FE 51 | 07 79 04 03 02 01 24 40 01 04 | 95 16
func COMMAND_SET_IDENTIFICATION(deviceAddress uint, id uint32, manufacturer uint16, generation, medium byte) []byte {
	data := []byte{
		0x07, // DIF: 64-bit integer (8 bytes) - but actually variable here
		0x79, // VIF: Enhanced identification
		// ID number (4 bytes, LSB first, BCD encoded)
		byte(id),
		byte(id >> 8),
		byte(id >> 16),
		byte(id >> 24),
		// Manufacturer ID (2 bytes, LSB first)
		byte(manufacturer),
		byte(manufacturer >> 8),
		// Generation and Medium
		generation,
		medium,
	}
	return COMMAND_SND_UD(deviceAddress, byte(CiFieldDataSend), data)
}

// COMMAND_SET_ID_NUMBER sets only the identification number of a slave
// Example: 68 08 08 68 | 53 FE 51 | 0C 79 78 56 34 12 | CS 16
func COMMAND_SET_ID_NUMBER(deviceAddress uint, id uint32) []byte {
	data := []byte{
		0x0C, // DIF: 8-digit BCD
		0x79, // VIF: Enhanced identification (ID number)
		// ID number (4 bytes, LSB first, BCD)
		byte(id),
		byte(id >> 8),
		byte(id >> 16),
		byte(id >> 24),
	}
	return COMMAND_SND_UD(deviceAddress, byte(CiFieldDataSend), data)
}

// COMMAND_SET_BAUDRATE sets the communication baudrate of a slave
// Valid baudrate codes: 0xB8=300, 0xBA=1200, 0xBB=2400, 0xBC=4800, 0xBD=9600, 0xBE=19200, 0xBF=38400
// Example: 68 03 03 68 | 53 FE BD | 0E 16 (set to 9600 baud)
func COMMAND_SET_BAUDRATE(deviceAddress uint, baudrateCI CIField) []byte {
	return BuildControlFrame(CFIELD_SND_UD_0.getByte(), byte(deviceAddress), byte(baudrateCI))
}

// COMMAND_APPLICATION_RESET sends an application reset to the slave
// Subcode specifies what type of reset:
//   0x00 = All application data
//   0x01 = User Data (all of installation)
//   0x02 = Simple billing
//   etc.
// Example: 68 04 04 68 | 53 FE 50 | 10 | B1 16
func COMMAND_APPLICATION_RESET(deviceAddress uint, subcode byte) []byte {
	data := []byte{subcode}
	return COMMAND_SND_UD(deviceAddress, byte(CiFieldApplicationReset), data)
}

// COMMAND_SET_COUNTER sets a counter value
// dif: Data Information Field (defines data type/length)
// vif: Value Information Field (defines unit)
// value: the counter value bytes (LSB first)
// Example: 68 0A 0A 68 | 53 01 51 | 0C 06 07 01 00 00 | CS 16 (set energy to 107 kWh)
func COMMAND_SET_COUNTER(deviceAddress uint, dif, vif byte, value []byte) []byte {
	data := make([]byte, 0, 2+len(value))
	data = append(data, dif)
	data = append(data, vif)
	data = append(data, value...)
	return COMMAND_SND_UD(deviceAddress, byte(CiFieldDataSend), data)
}

// COMMAND_SET_COUNTER_BCD8 sets an 8-digit BCD counter value
// vif: Value Information Field (defines unit, e.g., 0x06 for 1 kWh)
// value: the counter value as integer (will be BCD encoded)
func COMMAND_SET_COUNTER_BCD8(deviceAddress uint, vif byte, value uint32) []byte {
	// Encode as 8-digit BCD (4 bytes)
	bcd := EncodeBCD(value, 4)
	return COMMAND_SET_COUNTER(deviceAddress, 0x0C, vif, bcd)
}

// COMMAND_SELECT_DATA_RECORDS configures which data records the slave should respond with
// records: the VIF bytes specifying which records to include
// Example: 68 07 07 68 | 53 07 51 | 08 13 08 5A | 28 16
func COMMAND_SELECT_DATA_RECORDS(deviceAddress uint, records []byte) []byte {
	return COMMAND_SND_UD(deviceAddress, byte(CiFieldDataSend), records)
}

// COMMAND_SELECT_ALL_DATA requests all data from the slave
// Example: 68 04 04 68 | 53 03 51 | 7F | 26 16
func COMMAND_SELECT_ALL_DATA(deviceAddress uint) []byte {
	data := []byte{0x7F} // Global readout request
	return COMMAND_SND_UD(deviceAddress, byte(CiFieldDataSend), data)
}

// COMMAND_SYNCHRONIZE sends a synchronize action to slaves
// This can be used to synchronize multiple slaves for simultaneous readings
func COMMAND_SYNCHRONIZE(deviceAddress uint) []byte {
	return BuildControlFrame(CFIELD_SND_UD_0.getByte(), byte(deviceAddress), byte(CiFieldSynchronizeAction))
}

// EncodeBCD encodes an integer value as BCD (Binary Coded Decimal)
// numBytes: number of bytes to produce (each byte = 2 BCD digits)
func EncodeBCD(value uint32, numBytes int) []byte {
	result := make([]byte, numBytes)
	for i := 0; i < numBytes; i++ {
		lowNibble := byte(value % 10)
		value /= 10
		highNibble := byte(value % 10)
		value /= 10
		result[i] = (highNibble << 4) | lowNibble
	}
	return result
}

// EncodeManufacturerId encodes a 3-character manufacturer ID to 2 bytes
// Formula: ID = (C1-64)*1024 + (C2-64)*32 + (C3-64)
func EncodeManufacturerId(id string) []byte {
	if len(id) != 3 {
		return []byte{0, 0}
	}
	c1 := int(id[0]) - 64
	c2 := int(id[1]) - 64
	c3 := int(id[2]) - 64
	encoded := uint16(c1*1024 + c2*32 + c3)
	return []byte{byte(encoded), byte(encoded >> 8)}
}

// sendDataRequest SendMessage Request Command to serial port and Device Address
// If the port is in use by another application, it will retry until the port becomes available
// or until the timeout is reached
func sendDataRequest(serialPort string, deviceAddress uint, command []byte) ([]byte, error) {
	// Maximum time to wait for the port to become available
	maxWaitTime := 30 * time.Second
	// Interval between retries
	retryInterval := 1 * time.Second
	// Start time to track timeout
	startTime := time.Now()

	config := serial.Config{
		Name:        serialPort,
		Baud:        2400,
		Size:        8,
		StopBits:    serial.Stop1,
		Parity:      serial.ParityEven,
		ReadTimeout: 2 * time.Second,
	}

	var port *serial.Port
	var err error

	// Try to open the port with retries if it's in use
	for {
		port, err = serial.OpenPort(&config)
		if err == nil {
			break
		}

		// Check if the error is related to the port being in use
		if err.Error() == "serial: port already open" || err.Error() == "serial: Access is denied." {
			// Check if we've exceeded the maximum wait time
			if time.Since(startTime) > maxWaitTime {
				return nil, err
			}
			time.Sleep(retryInterval)
			continue
		}

		// For other errors, log and return
		return nil, err
	}

	defer func() {
		err := port.Close()
		if err != nil {
		}
	}()

	_, err = port.Write(command)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, rtuMaxSize)
	//TODO: port.Read or io.ReadAll ?
	//_, err = port.Read(buf)
	buf, err = io.ReadAll(port)
	if err != nil {
		return nil, err
	}
	//_, err = io.ReadFull(port, buf)
	//io.ReadAtLeast(port, data[:], rtuMaxSize)
	if err != nil {
		return buf, err
	}
	return buf, err
}

// sendSingle SendMessage Command and request will be 1 byte ie. ping if device exist on serial port and device address
// If the port is in use by another application, it will retry until the port becomes available
// or until the timeout is reached
func sendSingle(serialPort string, command []byte) ([1]byte, error) {

	config := serial.Config{
		Name:        serialPort,
		Baud:        2400,
		Size:        8,
		StopBits:    1,
		Parity:      serial.ParityEven,
		ReadTimeout: 150 * time.Millisecond,
		//Timeout:  100 * time.Millisecond,
	}

	port, err := serial.OpenPort(&config)
	if err != nil {
		return [1]byte{0}, err
	}

	defer func() {
		err := port.Close()
		if err != nil {
		}
	}()

	_, err = port.Write(command)
	if err != nil {
		return [1]byte{0}, err
	}

	buf, err := io.ReadAll(port)
	if err != nil {
		return [1]byte{}, err
	}
	// Convert buf to [1]byte
	var data [1]byte
	copy(data[:], buf[:1])
	return data, nil
}

// pingAddress Ping if device exist on serial port and address
// If the port is in use by another application, it will retry until the port becomes available
// or until the timeout is reached
func pingAddress(serialPort string, deviceAddress uint) (bool, error) {

	answer, err := sendSingle(serialPort, COMMAND_SND_NKE(deviceAddress))

	if err == nil {
		alive := answer[0] == 0xE5
		return alive, nil
	}

	return false, err
}

func readDeviceState(serialPort string, deviceAddress uint) (LFrameParsed, error) {
	rawData, err := sendDataRequest(serialPort, deviceAddress, COMMAND_REQ_UD2(deviceAddress))
	if err != nil {
		return LFrameParsed{}, err
	}

	frame := NewLFrame(rawData)

	return frame.parse()
}

// sendWriteCommand sends a write command and expects an ACK response
// Returns true if ACK received, false otherwise
func sendWriteCommand(serialPort string, command []byte) (bool, error) {
	response, err := sendSingle(serialPort, command)
	if err != nil {
		return false, err
	}
	return response[0] == FRAME_ACK_START, nil
}

// setDeviceAddress sets the primary address of a device
// Uses broadcast address to set the address (point-to-point connection required)
func setDeviceAddress(serialPort string, newAddress uint) (bool, error) {
	command := COMMAND_SET_PRIMARY_ADDRESS(newAddress)
	return sendWriteCommand(serialPort, command)
}

// setDeviceAddressFrom sets the primary address from a known current address
func setDeviceAddressFrom(serialPort string, currentAddress, newAddress uint) (bool, error) {
	command := COMMAND_SET_PRIMARY_ADDRESS_FROM(currentAddress, newAddress)
	return sendWriteCommand(serialPort, command)
}

// setDeviceIdentification sets the complete identification of a device
func setDeviceIdentification(serialPort string, address uint, id uint32, manufacturer uint16, generation, medium byte) (bool, error) {
	command := COMMAND_SET_IDENTIFICATION(address, id, manufacturer, generation, medium)
	return sendWriteCommand(serialPort, command)
}

// setDeviceBaudrate sets the communication baudrate of a device
func setDeviceBaudrate(serialPort string, address uint, baudrate CIField) (bool, error) {
	command := COMMAND_SET_BAUDRATE(address, baudrate)
	return sendWriteCommand(serialPort, command)
}

// resetDevice sends an application reset to the device
func resetDevice(serialPort string, address uint, subcode byte) (bool, error) {
	command := COMMAND_APPLICATION_RESET(address, subcode)
	return sendWriteCommand(serialPort, command)
}

// setCounterValue sets a counter value on the device
func setCounterValue(serialPort string, address uint, dif, vif byte, value []byte) (bool, error) {
	command := COMMAND_SET_COUNTER(address, dif, vif, value)
	return sendWriteCommand(serialPort, command)
}

// selectDataRecords configures which data records the device should respond with
func selectDataRecords(serialPort string, address uint, records []byte) (bool, error) {
	command := COMMAND_SELECT_DATA_RECORDS(address, records)
	return sendWriteCommand(serialPort, command)
}

// sendUserData sends generic user data to the device
func sendUserData(serialPort string, address uint, ciField byte, data []byte) (bool, error) {
	command := COMMAND_SND_UD(address, ciField, data)
	return sendWriteCommand(serialPort, command)
}
