package ipv6

import (
	"fmt"
	"net"
)

const (
	// addressSize is the number of bits that an IPv6 address takes
	addressSize int = 128
)

// Address represents an IPv6 address
type Address struct {
	ui uint128
}

// AddressFromUint64 returns the Address from its unsigned int representation
func AddressFromUint64(high, low uint64) Address {
	return Address{uint128{high, low}}
}

// AddressFromUint16 returns the Address from its eight 16 bit unsigned representation
func AddressFromUint16(a, b, c, d, e, f, g, h uint16) Address {
	high := uint64(a)<<48 |
		uint64(b)<<32 |
		uint64(c)<<16 |
		uint64(d)
	low := uint64(e)<<48 |
		uint64(f)<<32 |
		uint64(g)<<16 |
		uint64(h)
	return Address{uint128{high, low}}
}

// AddressFromNetIP converts a NetIP to an Address
func AddressFromNetIP(ip net.IP) (Address, error) {
	return fromSlice(ip)
}

// ParseAddress returns the Address represented by `addr` in colon
// notation. If it cannot be parsed, then error is non-nil and the Address
// returned must be ignored.
func ParseAddress(address string) (Address, error) {
	netIP := net.ParseIP(address)
	if netIP == nil {
		return Address{}, fmt.Errorf("failed to parse address: %s", address)
	}

	netIPv4 := netIP.To4()
	if netIPv4 != nil {
		return Address{}, fmt.Errorf("address is not IPv6: %s", address)
	}

	return AddressFromNetIP(netIP)
}

// minAddress returns the address, a or b, which comes first in lexigraphical order
func minAddress(a, b Address) Address {
	if a.lessThan(b) {
		return a
	}
	return b
}

// maxAddress returns the address, a or b, which comes last in lexigraphical order
func maxAddress(a, b Address) Address {
	if a.lessThan(b) {
		return b
	}
	return a
}

// ToNetIP returns a net.IP representation of the address which always has 4 bytes
func (me Address) ToNetIP() net.IP {
	return me.ui.toBytes()
}

// lessThan reports whether this Address comes strictly before `other`
// lexigraphically.
func (me Address) lessThan(other Address) bool {
	return me.ui.compare(other.ui) < 0
}

// String returns a string representing the address in IPv6 notation
func (me Address) String() string {
	return me.ToNetIP().String()
}

// Size returns the size of an address (always 128)
func (me Address) Size() int {
	return addressSize
}

// Uint64 returns the address as two uint64
func (me Address) Uint64() (uint64, uint64) {
	return me.ui.uint64()
}

// fromSlice returns the Address from a slice or an error if the slice is the
// wrong length.
func fromSlice(s []byte) (Address, error) {
	if s == nil {
		return Address{}, fmt.Errorf("failed to parse nil ip")
	}
	if len(s) != 16 {
		return Address{}, fmt.Errorf("failed to parse ip because slice size is not equal to 16")
	}
	val, err := uint128FromBytes(s)
	return Address{val}, err
}
