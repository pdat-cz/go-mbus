package mbus

import "time"

func Ping(port string, address int) PingState {

	ps := PingState{}
	ps.Port = port
	ps.Address = address
	ps.Timestamp = time.Now()
	state, err := pingAddress(port, uint(address))
	if err != nil {
		ps.Error = err.Error()
	}
	ps.State = state
	return ps

}

func Read(port string, address int) DeviceState {
	ds := DeviceState{}
	ds.Port = port
	ds.Address = address
	ds.Timestamp = time.Now()
	data, err := readDeviceState(port, uint(address))
	if err != nil {
		ds.Error = err.Error()
	}
	ds.Data = data
	return ds
}

// =============================================================================
// Write Operations
// =============================================================================

// WriteResult represents the result of a write operation.
type WriteResult struct {
	Port      string    `json:"port"`
	Address   int       `json:"address"`
	Success   bool      `json:"success"`
	Timestamp time.Time `json:"timestamp"`
	Error     string    `json:"error"`
}

// SetAddress sets the primary address of a device (broadcast, point-to-point only).
// This uses broadcast address 0xFE, so only one device should be connected.
func SetAddress(port string, newAddress int) WriteResult {
	wr := WriteResult{
		Port:      port,
		Address:   newAddress,
		Timestamp: time.Now(),
	}
	success, err := setDeviceAddress(port, uint(newAddress))
	if err != nil {
		wr.Error = err.Error()
	}
	wr.Success = success
	return wr
}

// SetAddressFrom sets the primary address of a device from a known current address.
func SetAddressFrom(port string, currentAddress, newAddress int) WriteResult {
	wr := WriteResult{
		Port:      port,
		Address:   currentAddress,
		Timestamp: time.Now(),
	}
	success, err := setDeviceAddressFrom(port, uint(currentAddress), uint(newAddress))
	if err != nil {
		wr.Error = err.Error()
	}
	wr.Success = success
	return wr
}

// SetIdentification sets the complete identification of a device.
// id: identification number (e.g., 0x12345678)
// manufacturer: manufacturer code (e.g., 0x4024 for PAD)
// generation: device generation/version
// medium: device medium type code
func SetIdentification(port string, address int, id uint32, manufacturer uint16, generation, medium byte) WriteResult {
	wr := WriteResult{
		Port:      port,
		Address:   address,
		Timestamp: time.Now(),
	}
	success, err := setDeviceIdentification(port, uint(address), id, manufacturer, generation, medium)
	if err != nil {
		wr.Error = err.Error()
	}
	wr.Success = success
	return wr
}

// SetBaudrate sets the communication baudrate of a device.
// Valid values: CiFieldBaudrate300, CiFieldBaudrate1200, CiFieldBaudrate2400,
// CiFieldBaudrate4800, CiFieldBaudrate9600, CiFieldBaudrate19200, CiFieldBaudrate38400
func SetBaudrate(port string, address int, baudrate CIField) WriteResult {
	wr := WriteResult{
		Port:      port,
		Address:   address,
		Timestamp: time.Now(),
	}
	success, err := setDeviceBaudrate(port, uint(address), baudrate)
	if err != nil {
		wr.Error = err.Error()
	}
	wr.Success = success
	return wr
}

// Reset sends an application reset to a device.
// subcode specifies the type of reset:
//   - 0x00: All application data
//   - 0x01: User data reset
//   - 0x02: Simple billing reset
func Reset(port string, address int, subcode byte) WriteResult {
	wr := WriteResult{
		Port:      port,
		Address:   address,
		Timestamp: time.Now(),
	}
	success, err := resetDevice(port, uint(address), subcode)
	if err != nil {
		wr.Error = err.Error()
	}
	wr.Success = success
	return wr
}

// SetCounter sets a counter value on the device.
// dif: Data Information Field (defines data type/length)
// vif: Value Information Field (defines unit)
// value: the counter value bytes (LSB first)
func SetCounter(port string, address int, dif, vif byte, value []byte) WriteResult {
	wr := WriteResult{
		Port:      port,
		Address:   address,
		Timestamp: time.Now(),
	}
	success, err := setCounterValue(port, uint(address), dif, vif, value)
	if err != nil {
		wr.Error = err.Error()
	}
	wr.Success = success
	return wr
}

// SelectDataRecords configures which data records the device should respond with.
// records: the selection data including DIF/VIF specifying which records to include
func SelectDataRecords(port string, address int, records []byte) WriteResult {
	wr := WriteResult{
		Port:      port,
		Address:   address,
		Timestamp: time.Now(),
	}
	success, err := selectDataRecords(port, uint(address), records)
	if err != nil {
		wr.Error = err.Error()
	}
	wr.Success = success
	return wr
}

// SendUserData sends generic user data to the device using the DATA_SEND CI field.
func SendUserData(port string, address int, data []byte) WriteResult {
	wr := WriteResult{
		Port:      port,
		Address:   address,
		Timestamp: time.Now(),
	}
	success, err := sendUserData(port, uint(address), byte(CiFieldDataSend), data)
	if err != nil {
		wr.Error = err.Error()
	}
	wr.Success = success
	return wr
}

// SendRawCommand sends a raw command and returns the ACK status.
// Use this for custom commands not covered by other functions.
func SendRawCommand(port string, address int, ciField byte, data []byte) WriteResult {
	wr := WriteResult{
		Port:      port,
		Address:   address,
		Timestamp: time.Now(),
	}
	success, err := sendUserData(port, uint(address), ciField, data)
	if err != nil {
		wr.Error = err.Error()
	}
	wr.Success = success
	return wr
}
