package ipv6

import (
	"fmt"
	"net"
)

const (
	// SIZE is the number of bits that an IPv6 address takes
	SIZE int = 128
)

// Address represents an IPv6 address
type Address struct {
	ui uint128
}

// AddressFromUint64 returns the IPv6 address from its two 64 bit unsigned representation
func AddressFromUint64(high uint64, low uint64) Address {
	return Address{uint128{high, low}}
}

// AddressFromBytes returns the IPv6 address from bytes
func AddressFromBytes(s []byte) Address {
	return Address{Uint128FromBytes(s)}
}

// AddressFromStdIP converts
func AddressFromStdIP(ip net.IP) (Address, error) {
	return fromSlice(ip)
}

// ParseAddress returns the ip represented by `addr` in dotted-quad notation. If
// the address cannot be parsed, then error is non-nil and the address returned
// must be ignored.
func ParseAddress(address string) (Address, error) {
	netIP := net.ParseIP(address)
	if netIP == nil {
		return Address{}, fmt.Errorf("failed to parse IPv6: %s", address)
	}
	return AddressFromStdIP(netIP)
}

// MinAddress returns the address, a or b, which comes first in lexigraphical order
func MinAddress(a, b Address) Address {
	if a.LessThan(b) {
		return a
	}
	return b
}

// MaxAddress returns the address, a or b, which comes last in lexigraphical order
func MaxAddress(a, b Address) Address {
	if a.LessThan(b) {
		return b
	}
	return a
}

// ToStdIP returns a net.IP representation of the address which always has 4 bytes
func (me Address) ToStdIP() net.IP {
	ip := me.toBytes()
	return ip[:]
}

// Equal reports whether this IPv6 address is the same as other
func (me Address) Equal(other Address) bool {
	return me.ui.Equal(other.ui)
}

// LessThan reports whether this IPv6 address comes strictly before `other`
// lexigraphically.
func (me Address) LessThan(other Address) bool {
	return me.ui.Compare(other.ui) < 0
}

/* TODO: implement once prefix is implemented
// Prefix returns a host prefix (/128) with the address
func (me Address) Prefix() Prefix {
	return Prefix{me, uint32(SIZE)}
}*/

// String returns a string representing the address in IPv6 notation
func (me Address) String() string {
	return me.ToStdIP().String()
}

// Uint128 returns the address as uint128
func (me Address) Uint128() uint128 {
	return me.ui
}

// Uint64 returns the address as two uint64
func (me Address) Uint64() (uint64, uint64) {
	return me.ui.Uint64()
}

// IsGlobalUnicast calls the same method from net.IP
func (me Address) IsGlobalUnicast() bool {
	return me.ToStdIP().IsGlobalUnicast()
}

// IsInterfaceLocalMulticast calls the same method from net.IP
func (me Address) IsInterfaceLocalMulticast() bool {
	return me.ToStdIP().IsInterfaceLocalMulticast()
}

// IsLinkLocalMulticast calls the same method from net.IP
func (me Address) IsLinkLocalMulticast() bool {
	return me.ToStdIP().IsLinkLocalMulticast()
}

// IsLinkLocalUnicast calls the same method from net.IP
func (me Address) IsLinkLocalUnicast() bool {
	return me.ToStdIP().IsLinkLocalUnicast()
}

// IsLoopback calls the same method from net.IP
func (me Address) IsLoopback() bool {
	return me.ToStdIP().IsLoopback()
}

// IsMulticast calls the same method from net.IP
func (me Address) IsMulticast() bool {
	return me.ToStdIP().IsMulticast()
}

// IsUnspecified calls the same method from net.IP
func (me Address) IsUnspecified() bool {
	return me.ToStdIP().IsUnspecified()
}

func (me Address) toBytes() []byte {
	return me.ui.ToBytes()
}

// fromSlice returns the IPv6 address from a slice with 16 bytes or an error
// if the slice is the wrong size.
func fromSlice(s []byte) (Address, error) {
	if s == nil {
		return Address{}, fmt.Errorf("failed to parse nil ip")
	}
	if len(s) != 16 {
		return Address{}, fmt.Errorf("failed to parse ip because slice size is not equal to 16")
	}
	return AddressFromBytes(s), nil
}
