# go-mbus

A Go implementation of the M-Bus (Meter-Bus) protocol for remote reading of utility meters.

## Overview

go-mbus is a library that provides functionality for communicating with devices using the M-Bus (Meter-Bus) protocol, which is a European standard (EN 13757-2, EN 13757-3) for remote reading of utility meters (water, gas, electricity, heat, etc.).

The library allows you to:
- Ping M-Bus devices to check if they are alive
- Read data from M-Bus devices
- Parse and interpret M-Bus telegrams
- Work with various M-Bus frame types (Short Frame, Long Frame, Control Frame)

## Installation

```bash
go get github.com/pdat-cz/go-mbus
```

### Dependencies

This library depends on the following packages:
- `github.com/tarm/serial` - For serial port communication
- `tycho-edge/pkg/log` - For logging (you'll need to provide this package or modify the code to use a different logging mechanism)
- `gopkg.in/yaml.v3` - For YAML parsing

## Usage

### Basic Example

```go
package main

import (
    "fmt"
    "github.com/pdat-cz/go-mbus"
)

func main() {
    // Ping a device at address 1 on port /dev/ttyUSB0
    pingState := mbus.Ping("/dev/ttyUSB0", 1)

    if pingState.State {
        fmt.Printf("Device at address %d is alive\n", pingState.Address)

        // Read data from the device
        deviceState := mbus.Read("/dev/ttyUSB0", 1)

        if deviceState.Error == "" {
            fmt.Printf("Device data: %+v\n", deviceState.Data)
        } else {
            fmt.Printf("Error reading device: %s\n", deviceState.Error)
        }
    } else {
        fmt.Printf("Device at address %d is not responding: %s\n", 
            pingState.Address, pingState.Error)
    }
}
```

### Scanning for Devices

```go
package main

import (
    "fmt"
    "github.com/pdat-cz/go-mbus"
)

func main() {
    // Scan for devices on port /dev/ttyUSB0, addresses 1-250
    fmt.Println("Scanning for M-Bus devices...")

    for address := 1; address <= 250; address++ {
        pingState := mbus.Ping("/dev/ttyUSB0", address)

        if pingState.State {
            fmt.Printf("Found device at address %d\n", address)
        }
    }
}
```

## Documentation

For more detailed information about the M-Bus protocol and how to use this library, see:

- [Protocol Documentation](docs/protocol.md) - Details about the M-Bus protocol
- [Examples](docs/examples.md) - More usage examples

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## References

- [M-Bus Documentation](https://m-bus.com/documentation-wired) - Official M-Bus protocol documentation
- [EN 13757-2](https://www.en-standard.eu/csn-en-13757-2-communication-systems-for-meters-and-remote-reading-of-meters-part-2-physical-and-link-layer/) - Physical and link layer
- [EN 13757-3](https://www.en-standard.eu/csn-en-13757-3-communication-systems-for-meters-and-remote-reading-of-meters-part-3-dedicated-application-layer/) - Application layer
