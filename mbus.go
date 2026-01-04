// Package mbus provides functionality for working with M-Bus devices.
//
// This package is a wrapper around the pkg/mbus package to allow importing from
// "github.com/pdat-cz/go-mbus" instead of "github.com/pdat-cz/go-mbus/pkg/mbus".
package mbus

import (
	"time"

	"github.com/pdat-cz/go-mbus/pkg/mbus"
)

// Ping checks if a device is alive at the given address.
func Ping(port string, address int) mbus.PingState {
	return mbus.Ping(port, address)
}

// Read reads data from a device at the given address.
func Read(port string, address int) DeviceState {
	ds := mbus.Read(port, address)

	// Convert to our DeviceState with our LFrameRecord
	result := DeviceState{
		Port:      ds.Port,
		Address:   ds.Address,
		Timestamp: ds.Timestamp,
		Error:     ds.Error,
		Data: LFrameParsed{
			IdentificationNumber: ds.Data.IdentificationNumber,
			Manufacturer:         ds.Data.Manufacturer,
			Version:              ds.Data.Version,
			Medium:               ds.Data.Medium,
			AccessNumber:         ds.Data.AccessNumber,
			Status:               ds.Data.Status,
			Model:                ds.Data.Model,
			Address:              ds.Data.Address,
			Signature:            ds.Data.Signature,
			Records:              make(map[int]LFrameRecord),
		},
	}

	// Convert each record, setting Description to Name
	for i, record := range ds.Data.Records {
		result.Data.Records[i] = LFrameRecord{
			DIF:         record.DIF,
			DIFE:        record.DIFE,
			VIF:         record.VIF,
			VIFE:        record.VIFE,
			VIFEM:       record.VIFEM,
			Value:       record.Value,
			Function:    record.Function,
			Unit:        record.Unit,
			Name:        record.Name,
			Exponent:    record.Exponent,
			Description: record.Name, // Set Description to Name
		}
	}

	return result
}

// PingState represents the state of a ping operation.
type PingState = mbus.PingState

// DeviceState represents the state of a device.
type DeviceState struct {
	Port      string       `json:"port"`
	Address   int          `json:"address"`
	Data      LFrameParsed `json:"data"`
	Timestamp time.Time    `json:"timestamp"`
	Error     string       `json:"error"`
}

// LFrameParsed represents a parsed M-Bus telegram.
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

// LFrameRecord represents a record in an M-Bus telegram.
type LFrameRecord struct {
	DIF         byte    `yaml:"DIF" json:"DIF"`
	DIFE        []byte  `yaml:"DIFE" json:"DIFE"`
	VIF         byte    `yaml:"VIF" json:"VIF"`
	VIFE        []byte  `yaml:"VIFE" json:"VIFE"`
	VIFEM       []byte  `yaml:"VIFEM" json:"VIFEM"`
	Value       string  `yaml:"value" json:"value"`
	Function    string  `yaml:"function" json:"function"`
	Unit        string  `yaml:"unit" json:"unit"`
	Name        string  `yaml:"name" json:"name"`
	Exponent    float64 `yaml:"exponent" json:"exponent"`
	Description string  `yaml:"description" json:"description"`
}

// =============================================================================
// Write Operations
// =============================================================================

// WriteResult represents the result of a write operation.
type WriteResult = mbus.WriteResult

// SetAddress sets the primary address of a device (broadcast, point-to-point only).
// This uses broadcast address 0xFE, so only one device should be connected.
func SetAddress(port string, newAddress int) WriteResult {
	return mbus.SetAddress(port, newAddress)
}

// SetAddressFrom sets the primary address of a device from a known current address.
func SetAddressFrom(port string, currentAddress, newAddress int) WriteResult {
	return mbus.SetAddressFrom(port, currentAddress, newAddress)
}

// SetIdentification sets the complete identification of a device.
// id: identification number (e.g., 0x12345678)
// manufacturer: manufacturer code (e.g., 0x4024 for PAD)
// generation: device generation/version
// medium: device medium type code
func SetIdentification(port string, address int, id uint32, manufacturer uint16, generation, medium byte) WriteResult {
	return mbus.SetIdentification(port, address, id, manufacturer, generation, medium)
}

// SetBaudrate sets the communication baudrate of a device.
// Use the CIField baudrate constants from the mbus package.
func SetBaudrate(port string, address int, baudrate mbus.CIField) WriteResult {
	return mbus.SetBaudrate(port, address, baudrate)
}

// Reset sends an application reset to a device.
// subcode specifies the type of reset:
//   - 0x00: All application data
//   - 0x01: User data reset
//   - 0x02: Simple billing reset
func Reset(port string, address int, subcode byte) WriteResult {
	return mbus.Reset(port, address, subcode)
}

// SetCounter sets a counter value on the device.
// dif: Data Information Field (defines data type/length)
// vif: Value Information Field (defines unit)
// value: the counter value bytes (LSB first)
func SetCounter(port string, address int, dif, vif byte, value []byte) WriteResult {
	return mbus.SetCounter(port, address, dif, vif, value)
}

// SelectDataRecords configures which data records the device should respond with.
// records: the selection data including DIF/VIF specifying which records to include
func SelectDataRecords(port string, address int, records []byte) WriteResult {
	return mbus.SelectDataRecords(port, address, records)
}

// SendUserData sends generic user data to the device using the DATA_SEND CI field.
func SendUserData(port string, address int, data []byte) WriteResult {
	return mbus.SendUserData(port, address, data)
}

// SendRawCommand sends a raw command and returns the ACK status.
// Use this for custom commands not covered by other functions.
func SendRawCommand(port string, address int, ciField byte, data []byte) WriteResult {
	return mbus.SendRawCommand(port, address, ciField, data)
}

// CIField type alias for baudrate constants
type CIField = mbus.CIField

// Baudrate constants for SetBaudrate
var (
	CiFieldBaudrate300   = mbus.CiFieldBaudrate300
	CiFieldBaudrate1200  = mbus.CiFieldBaudrate1200
	CiFieldBaudrate2400  = mbus.CiFieldBaudrate2400
	CiFieldBaudrate4800  = mbus.CiFieldBaudrate4800
	CiFieldBaudrate9600  = mbus.CiFieldBaudrate9600
	CiFieldBaudrate19200 = mbus.CiFieldBaudrate19200
	CiFieldBaudrate38400 = mbus.CiFieldBaudrate38400
)
