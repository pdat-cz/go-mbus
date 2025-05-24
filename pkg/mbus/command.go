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
