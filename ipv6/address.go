package ipv6

import (
	"fmt"
	"net"
)

const (
	// SIZE is the number of bits that an IPv6 address takes
	SIZE                int = 128
	REPRESENTATIVE_SIZE int = 64
)

// Address represents an IPv6 address
type Address struct {
	high, low uint64
}

// AddressFromUint64 returns the IPv6 address from its two 64 bit unsigned representation
func AddressFromUint64(high uint64, low uint64) Address {
	return Address{high, low}
}

// AddressFromBytes returns the IPv6 address of the `a.b.c.d`.
func AddressFromBytes(s []byte) Address {
	return Address{
		high: uint64(s[0])<<56 | uint64(s[1])<<48 | uint64(s[2])<<40 | uint64(s[3])<<32 | uint64(s[4])<<24 | uint64(s[5])<<16 | uint64(s[6])<<8 | uint64(s[7]),
		low:  uint64(s[8])<<56 | uint64(s[9])<<48 | uint64(s[10])<<40 | uint64(s[11])<<32 | uint64(s[12])<<24 | uint64(s[13])<<16 | uint64(s[14])<<8 | uint64(s[15]),
	}
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
	return me == other
}

// LessThan reports whether this IPv6 address comes strictly before `other`
// lexigraphically.
func (me Address) LessThan(other Address) bool {
	if me.high == other.high {
		return me.low < other.low
	}
	return me.high < other.high
}

// DefaultMask returns the default IP mask for the given IP
func (me Address) DefaultMask() Mask {
	ones, _ := me.ToStdIP().DefaultMask().Size()
	return lengthToMask(ones)
}

// Prefix returns a host prefix (/128) with the address
func (me Address) Prefix() Prefix {
	return Prefix{me, uint32(SIZE)}
}

// TODO
// String returns a string representing the address in IPv6 notation
func (me Address) String() string {
	ip := me.toBytes()

	//make hex string for each octet
	//trim leading zeros
	//if equals zero remove hex value
	//trim any consqeutive : with length > 2
	return fmt.Sprintf("%x%x:%x%x:%x%x:%x%x:%x%x:%x%x:%x%x:%x%x")
}

// Uint64 returns the address as two uint64
func (me Address) Uint64() (uint64, uint64) {
	return me.high, me.low
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

// TODO
func (me Address) toBytes() [16]byte {
	ip := [16]byte{
		byte(0xff & me.high),
		byte(0xff & (me.high >> 8)),
		byte(0xff & (me.high >> 16)),
		byte(0xff & (me.high >> 24)),
		byte(0xff & (me.high >> 32)),
		byte(0xff & (me.high >> 40)),
		byte(0xff & (me.high >> 48)),
		byte(0xff & (me.high >> 56)),
		byte(0xff & me.low),
		byte(0xff & (me.low >> 8)),
		byte(0xff & (me.low >> 16)),
		byte(0xff & (me.low >> 24)),
		byte(0xff & (me.low >> 32)),
		byte(0xff & (me.low >> 40)),
		byte(0xff & (me.low >> 48)),
		byte(0xff & (me.low >> 56)),
	}
	return ip
}

// fromSlice returns the IPv6 address from a slice with four bytes or an error
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
