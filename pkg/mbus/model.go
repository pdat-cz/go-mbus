package mbus

import "time"

type PingState struct {
	Port      string    `json:"port"`
	Address   int       `json:"address"`
	State     bool      `json:"state"`
	Timestamp time.Time `json:"timestamp"`
	Error     string    `json:"error"`
}

type DeviceState struct {
	Port      string       `json:"port"`
	Address   int          `json:"address"`
	Data      LFrameParsed `json:"data"`
	Timestamp time.Time    `json:"timestamp"`
	Error     string       `json:"error"`
}
