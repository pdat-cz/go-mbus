package manager

import "github.com/pdat-cz/go-mbus/manager/transport"

// Type aliases for transport configurations.
// These allow users to configure transports without importing the transport package.

// SerialConfig is an alias for transport.SerialConfig.
type SerialConfig = transport.SerialConfig

// TCPConfig is an alias for transport.TCPConfig.
type TCPConfig = transport.TCPConfig
