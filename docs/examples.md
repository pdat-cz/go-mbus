# go-mbus Usage Examples

This document provides examples of how to use the go-mbus library for various M-Bus communication tasks.

## Basic Communication

### Pinging a Device

To check if an M-Bus device is alive:

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
    } else {
        fmt.Printf("Device at address %d is not responding: %s\n", 
            pingState.Address, pingState.Error)
    }
}
```

### Reading Data from a Device

To read data from an M-Bus device:

```go
package main

import (
    "fmt"
    "github.com/pdat-cz/go-mbus"
)

func main() {
    // Read data from a device at address 1 on port /dev/ttyUSB0
    deviceState := mbus.Read("/dev/ttyUSB0", 1)

    if deviceState.Error == "" {
        // Print the raw data
        fmt.Printf("Device data: %+v\n", deviceState.Data)

        // Access specific fields if available
        if len(deviceState.Data.Records) > 0 {
            for i, record := range deviceState.Data.Records {
                fmt.Printf("Record %d: %s = %s %s\n", 
                    i, record.Description, record.Value, record.Unit)
            }
        }
    } else {
        fmt.Printf("Error reading device: %s\n", deviceState.Error)
    }
}
```

## Device Discovery

### Scanning for Devices

To scan for M-Bus devices on a bus:

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

            // Optionally read device data
            deviceState := mbus.Read("/dev/ttyUSB0", address)
            if deviceState.Error == "" {
                fmt.Printf("Device data: %+v\n", deviceState.Data)
            }
        }
    }
}
```

## Working with Telegrams

### Parsing a Raw Telegram

To parse a raw M-Bus telegram:

```go
package main

import (
    "encoding/hex"
    "fmt"
    "github.com/pdat-cz/go-mbus"
)

func main() {
    // Example telegram in hex format
    hexData := "68 1F 1F 68 08 02 72 78 56 34 12 24 40 01 07 55 00 00 00 03 13 15 31 00 DA 02 3B 13 01 8B 60 04 37 18 02 18 16"

    // Convert hex string to bytes
    bytes := mbus.HexStringToBytes(hexData)

    // Parse the telegram
    telegram, err := mbus.ParseTelegram(bytes)
    if err != nil {
        fmt.Printf("Error parsing telegram: %s\n", err)
        return
    }

    // Access telegram data
    fmt.Printf("Telegram type: %s\n", telegram.Type)
    fmt.Printf("Manufacturer: %s\n", telegram.Manufacturer)
    fmt.Printf("Serial number: %s\n", telegram.SerialNumber)

    // Access data records
    for i, record := range telegram.DataRecords {
        fmt.Printf("Record %d: %s = %s %s\n", 
            i, record.Description, record.Value, record.Unit)
    }
}
```

## Advanced Usage

### Setting Device Parameters

To set parameters on an M-Bus device:

```go
package main

import (
    "fmt"
    "github.com/pdat-cz/go-mbus"
)

func main() {
    // Set the primary address of a device
    // Change device at address 1 to address 2
    err := mbus.SetPrimaryAddress("/dev/ttyUSB0", 1, 2)
    if err != nil {
        fmt.Printf("Error setting primary address: %s\n", err)
        return
    }

    fmt.Println("Primary address changed successfully")
}
```

### Reading Specific Data Points

To request specific data points from a device:

```go
package main

import (
    "fmt"
    "github.com/pdat-cz/go-mbus"
)

func main() {
    // Request specific data points (e.g., volume and flow temperature)
    // from a device at address 7
    dataPoints := []byte{0x13, 0x5A} // VIF codes for volume and flow temperature

    err := mbus.RequestSpecificData("/dev/ttyUSB0", 7, dataPoints)
    if err != nil {
        fmt.Printf("Error requesting specific data: %s\n", err)
        return
    }

    // Read the response
    deviceState := mbus.Read("/dev/ttyUSB0", 7)
    if deviceState.Error == "" {
        fmt.Printf("Device data: %+v\n", deviceState.Data)
    } else {
        fmt.Printf("Error reading device: %s\n", deviceState.Error)
    }
}
```

## Error Handling

Always check for errors when communicating with M-Bus devices:

```go
package main

import (
    "fmt"
    "github.com/pdat-cz/go-mbus"
    "time"
)

func main() {
    // Attempt to read with retries
    maxRetries := 3
    var deviceState mbus.DeviceState

    for retry := 0; retry < maxRetries; retry++ {
        deviceState = mbus.Read("/dev/ttyUSB0", 1)

        if deviceState.Error == "" {
            break
        }

        fmt.Printf("Retry %d: Error reading device: %s\n", retry+1, deviceState.Error)
        time.Sleep(1 * time.Second)
    }

    if deviceState.Error != "" {
        fmt.Printf("Failed to read device after %d retries\n", maxRetries)
        return
    }

    fmt.Printf("Device data: %+v\n", deviceState.Data)
}
```

## Note on Function Availability

Not all functions shown in these examples may be implemented in the current version of the library. These examples are intended to demonstrate the intended usage patterns of the library. Check the actual implementation for available functions.
