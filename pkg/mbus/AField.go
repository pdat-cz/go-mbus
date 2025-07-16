package mbus

import "errors"

// AField is a byte.
type AField byte

func NewAField(b byte) AField {
	return AField(b)
}

func (a *AField) String() string {
	b := byte(*a)
	return ByteToHexString(b)
}

// IsSlaveAddress returns true if the AField is a slave address. 1-250 are valid slave addresses.
func (a *AField) IsSlaveAddress() bool {
	b := byte(*a)
	return b >= 1 && b <= 250
}

// SlaveAddress Slave Address is uint8 in the range 1-250.
func (a *AField) SlaveAddress() (uint8, error) {
	b := byte(*a)
	if b >= 1 && b <= 250 {
		return uint8(b), nil
	}
	return 0, ErrInvalidSlaveAddress
}

var ErrInvalidSlaveAddress = errors.New("invalid / unconfigured slave address")

// IsUnconfiguredSlave returns true if the AField is an unconfigured slave address. 0 is the unconfigured slave address.
func (a *AField) IsUnconfiguredSlave() bool {
	b := byte(*a)
	return b == 0
}

// IsBroadcastAddressWithSlaveReplyAddress returns true if the AField is a broadcast address. 254 is the broadcast address.
// All slaves reply with their own addresses.
func (a *AField) IsBroadcastAddressWithSlaveReplyAddress() bool {
	return byte(*a) == byte(AFieldBroadcastAddressWithSlaveReplyAddress)
}

var AFieldBroadcastAddressWithSlaveReplyAddress = NewAField(254)

// IsBroadcastAddress returns true if the AField is a broadcast address. 255 is the broadcast address.
// None of the slaves reply.
func (a *AField) IsBroadcastAddress() bool {
	return byte(*a) == byte(AFieldBroadcastAddress)
}

var AFieldBroadcastAddress = NewAField(255)
