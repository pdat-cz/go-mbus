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
