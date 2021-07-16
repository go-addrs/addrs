package ipv4

import (
	"fmt"
	"net"
)

const (
	// SIZE is the number of bits that an IPv4 address takes
	SIZE uint32 = 32
)

// Addr represents an IPv4 address
type Addr struct {
	ui uint32
}

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

// Equal reports whether this IPv4 address is the same as other
func (me Addr) Equal(other Addr) bool {
	return me == other
}

// LessThan reports whether this IPv4 address comes strictly before `other`
// lexigraphically.
func (me Addr) LessThan(other Addr) bool {
	return me.ui < other.ui
}
