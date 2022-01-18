package ipv4

import (
	"fmt"
	"net"
)

const (
	// SIZE is the number of bits that an IPv4 address takes
	SIZE int = 32
)

// Address represents an IPv4 address
type Address struct {
	ui uint32
}

// AddressFromUint32 returns the IPv4 address from its 32 bit unsigned representation
func AddressFromUint32(ui uint32) Address {
	return Address{ui}
}

// AddressFromBytes returns the IPv4 address of the `a.b.c.d`.
func AddressFromBytes(a, b, c, d byte) Address {
	return Address{
		ui: uint32(a)<<24 | uint32(b)<<16 | uint32(c)<<8 | uint32(d),
	}
}

// AddressFromNetIP converts
func AddressFromNetIP(ip net.IP) (Address, error) {
	return fromSlice(ip.To4())
}

// ParseAddress returns the ip represented by `addr` in dotted-quad notation. If
// the address cannot be parsed, then error is non-nil and the address returned
// must be ignored.
func ParseAddress(address string) (Address, error) {
	netIP := net.ParseIP(address)
	if netIP == nil {
		return Address{}, fmt.Errorf("failed to parse IPv4: %s", address)
	}

	netIPv4 := netIP.To4()
	if netIPv4 == nil {
		return Address{}, fmt.Errorf("address is not IPv4: %s", address)
	}

	return AddressFromNetIP(netIPv4)
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

// ToNetIP returns a net.IP representation of the address which always has 4 bytes
func (me Address) ToNetIP() net.IP {
	a, b, c, d := me.toBytes()
	return net.IPv4(a, b, c, d)
}

// Equal reports whether this IPv4 address is the same as other
func (me Address) Equal(other Address) bool {
	return me == other
}

// LessThan reports whether this IPv4 address comes strictly before `other`
// lexigraphically.
func (me Address) LessThan(other Address) bool {
	return me.ui < other.ui
}

// DefaultMask returns the default IP mask for the given IP
func (me Address) DefaultMask() Mask {
	ones, _ := me.ToNetIP().DefaultMask().Size()
	return lengthToMask(ones)
}

// Prefix returns a host prefix (/32) with the address
func (me Address) Prefix() Prefix {
	return Prefix{me, uint32(SIZE)}
}

// String returns a string representing the address in dotted-quad notation
func (me Address) String() string {
	a, b, c, d := me.toBytes()
	return fmt.Sprintf("%d.%d.%d.%d", a, b, c, d)
}

// Uint32 returns the address as a uint32
func (me Address) Uint32() uint32 {
	return me.ui
}

// IsGlobalUnicast calls the same method from net.IP
func (me Address) IsGlobalUnicast() bool {
	return me.ToNetIP().IsGlobalUnicast()
}

// IsInterfaceLocalMulticast calls the same method from net.IP
func (me Address) IsInterfaceLocalMulticast() bool {
	return me.ToNetIP().IsInterfaceLocalMulticast()
}

// IsLinkLocalMulticast calls the same method from net.IP
func (me Address) IsLinkLocalMulticast() bool {
	return me.ToNetIP().IsLinkLocalMulticast()
}

// IsLinkLocalUnicast calls the same method from net.IP
func (me Address) IsLinkLocalUnicast() bool {
	return me.ToNetIP().IsLinkLocalUnicast()
}

// IsLoopback calls the same method from net.IP
func (me Address) IsLoopback() bool {
	return me.ToNetIP().IsLoopback()
}

// IsMulticast calls the same method from net.IP
func (me Address) IsMulticast() bool {
	return me.ToNetIP().IsMulticast()
}

// IsUnspecified calls the same method from net.IP
func (me Address) IsUnspecified() bool {
	return me.ToNetIP().IsUnspecified()
}

func (me Address) toBytes() (a, b, c, d byte) {
	a = byte(me.ui & 0xff000000 >> 24)
	b = byte(me.ui & 0xff0000 >> 16)
	c = byte(me.ui & 0xff00 >> 8)
	d = byte(me.ui & 0xff)
	return
}

// fromSlice returns the IPv4 address from a slice with four bytes or an error
// if the slice is the wrong size.
func fromSlice(s []byte) (Address, error) {
	if s == nil {
		return Address{}, fmt.Errorf("failed to parse nil ip")
	}
	if len(s) != 4 {
		return Address{}, fmt.Errorf("failed to parse ip because slice size is not equal to 4")
	}
	return AddressFromBytes(s[0], s[1], s[2], s[3]), nil
}
