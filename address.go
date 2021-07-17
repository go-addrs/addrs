package ipv4

import (
	"fmt"
	"net"
)

const (
	// SIZE is the number of bits that an IPv4 address takes
	SIZE int = 32
)

// Addr represents an IPv4 address
type Addr struct {
	ui uint32
}

// Mask represents an IPv4 prefix mask. It has 0-32 leading 1s and then all
// remaining bits are 0s
type Mask Addr

// AddrFromUint32 returns the IPv4 address from its 32 bit unsigned representation
func AddrFromUint32(ui uint32) Addr {
	return Addr{ui}
}

// AddrFromBytes returns the IPv4 address of the `a.b.c.d`.
func AddrFromBytes(a, b, c, d byte) Addr {
	return Addr{
		ui: uint32(a)<<24 | uint32(b)<<16 | uint32(c)<<8 | uint32(d),
	}
}

// fromSlice returns the IPv4 address from a slice with four bytes or an error
// if the slice is the wrong size.
func fromSlice(s []byte) (Addr, error) {
	if s == nil {
		return Addr{}, fmt.Errorf("failed to parse nil ip")
	}
	if len(s) != 4 {
		return Addr{}, fmt.Errorf("failed to parse ip because slice size is not equal to 4")
	}
	return AddrFromBytes(s[0], s[1], s[2], s[3]), nil
}

// AddrFromStdIP converts
func AddrFromStdIP(ip net.IP) (Addr, error) {
	return fromSlice(ip.To4())
}

// ParseAddr returns the ip represented by `addr` in dotted-quad notation. If
// the address cannot be parsed, then error is non-nil and the address returned
// must be ignored.
func ParseAddr(address string) (Addr, error) {
	netIP := net.ParseIP(address)
	if netIP == nil {
		return Addr{}, fmt.Errorf("failed to parse IPv4: %s", address)
	}

	netIPv4 := netIP.To4()
	if netIPv4 == nil {
		return Addr{}, fmt.Errorf("address is not IPv4: %s", address)
	}

	return AddrFromStdIP(netIPv4)
}

// Min returns the address, a or b, which comes first in lexigraphical order
func Min(a, b Addr) Addr {
	if a.LessThan(b) {
		return a
	}
	return b
}

// Max returns the address, a or b, which comes last in lexigraphical order
func Max(a, b Addr) Addr {
	if a.LessThan(b) {
		return b
	}
	return a
}

func lengthToMask(length int) Mask {
	return Mask{
		ui: ^uint32(0) << (SIZE - length),
	}
}

// CreateMask converts the given length (0-32) into a mask with that number of leading 1s
func CreateMask(length int) (Mask, error) {
	if length < 0 || SIZE < length {
		return Mask{}, fmt.Errorf("failed to create Mask where length %d isn't between 0 and 32", length)
	}

	return lengthToMask(length), nil
}

// ToStdIP returns a net.IP representation of the address which always has 4 bytes
func (me Addr) ToStdIP() net.IP {
	a := byte(me.ui & 0xff000000 >> 24)
	b := byte(me.ui & 0xff0000 >> 16)
	c := byte(me.ui & 0xff00 >> 8)
	d := byte(me.ui & 0xff)
	return net.IPv4(a, b, c, d)
}

// Equal reports whether this IPv4 address is the same as other
func (me Addr) Equal(other Addr) bool {
	return me == other
}

// LessThan reports whether this IPv4 address comes strictly before `other`
// lexigraphically.
func (me Addr) LessThan(other Addr) bool {
	return me.ui < other.ui
}

// DefaultMask returns the default IP mask for the given IP
func (me Addr) DefaultMask() Mask {
	ones, _ := me.ToStdIP().DefaultMask().Size()
	return lengthToMask(ones)
}
