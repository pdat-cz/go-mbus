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
