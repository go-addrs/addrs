package ipv4

import (
	"fmt"
	"net"
)

const (
	allOnes uint32 = ^uint32(0)
)

// Prefix represents an IPv4 prefix which is IPv4 address along with prefix
// length, or the number of bits which are significant in the network portion.
// Note that any bits in the address can be 0 or 1 regardless if they in the
// first `length` bits or not. This allows storing an IP address in CIDR
// notation with both the network and host parts of the address.
type Prefix struct {
	Addr
	length uint32
}

// IP returns the address part of the prefix alone
func (me Prefix) IP() Addr {
	return me.Addr
}

// PrefixFromUint32 returns the IPv4 address from its 32 bit unsigned representation
func PrefixFromUint32(ui, length uint32) (Prefix, error) {
	if SIZE < length {
		return Prefix{}, fmt.Errorf("failed to convert prefix, length %d is greater than 32", length)
	}
	return Prefix{Addr{ui}, length}, nil
}

// PrefixFromBytes returns the IPv4 address of the `a.b.c.d`.
func PrefixFromBytes(a, b, c, d byte, length uint32) (Prefix, error) {
	if SIZE < length {
		return Prefix{}, fmt.Errorf("failed to convert prefix, length %d is greater than 32", length)
	}
	return Prefix{
		Addr:   AddrFromBytes(a, b, c, d),
		length: length,
	}, nil
}

// PrefixFromStdIPNet converts the given *net.IPNet to a Prefix
func PrefixFromStdIPNet(net *net.IPNet) (Prefix, error) {
	if net == nil {
		return Prefix{}, fmt.Errorf("failed to convert nil *net.IPNet")
	}
	ones, bits := net.Mask.Size()
	if bits != int(SIZE) {
		return Prefix{}, fmt.Errorf("failed to convert IPNet with size != 32")
	}
	addr, err := AddrFromStdIP(net.IP)
	if err != nil {
		return Prefix{}, err
	}
	return Prefix{
		Addr:   addr,
		length: uint32(ones),
	}, nil
}

// ParseCIDRToNet returns a single *net.IPNet that unifies the IP address and
// the mask. It leaves out the network address which net.ParseCIDR returns.
// This may be considered an abuse of the IPNet construct as it is documented
// that IP is supposed to be the "network number". However, the public IPNet
// interface does not dissallow it and this usage has been spotted in the wild.
func parseNet(prefix string) (*net.IPNet, error) {
	ip, ipNet, err := net.ParseCIDR(prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to parse IPv4 prefix: %s", prefix)
	}
	return &net.IPNet{IP: ip, Mask: ipNet.Mask}, nil
}

// ParsePrefix returns the net represented by `addr` in dotted-quad CIDR
// notation. If the prefix cannot be parsed, then error is non-nil and the
// prefix returned must be ignored.
func ParsePrefix(prefix string) (Prefix, error) {
	ipNet, err := parseNet(prefix)
	if err != nil {
		return Prefix{}, err
	}
	return PrefixFromStdIPNet(ipNet)
}

// Equal reports whether this IPv4 address is the same as other
func (me Prefix) Equal(other Prefix) bool {
	return me == other
}

// LessThan reports whether this IPv4 prefix comes strictly before `other`
// lexigraphically.
func (me Prefix) LessThan(other Prefix) bool {
	meNet, otherNet := me.Network(), other.Network()
	if !meNet.IP().Equal(otherNet.IP()) {
		return meNet.IP().LessThan(otherNet.IP())
	}
	if me.length == other.length {
		return me.Host().IP().LessThan(other.Host().IP())
	}
	return me.length < other.length
}

// Mask returns a new Addr with 1s in the first `length` bits and then 0s
// representing the network mask for this prefix.
func (me Prefix) Mask() Addr {
	return Addr{
		ui: allOnes << (SIZE - me.length),
	}
}

// Network returns a new Prefix with all bits after `length` zeroed out so that
// only the bits in the `network` part of the prefix are present. Note that
// this method ignores special cases where a network address doesn't make
// sense like in a host route or point-to-point prefix (/32 and /31). It just
// does the math.
func (me Prefix) Network() Prefix {
	network := me.Addr.ui & me.Mask().ui
	return Prefix{
		Addr: Addr{
			ui: network,
		},
		length: me.length,
	}
}

// Broadcast returns a new Prefix with all bits after `length` set to 1s. Note
// that this method ignores special cases where a broadcast address doesn't
// make sense like in a host route or point-to-point prefix (/32 and /31). It
// just does the math.
func (me Prefix) Broadcast() Prefix {
	network := me.Addr.ui | ^me.Mask().ui
	return Prefix{
		Addr: Addr{
			ui: network,
		},
		length: me.length,
	}
}

// Host returns a new Prefix with the first `length` bits zeroed out so
// that only the bits in the `host` part of the prefix are present
func (me Prefix) Host() Prefix {
	host := me.Addr.ui & ^me.Mask().ui
	return Prefix{
		Addr: Addr{
			ui: host,
		},
		length: me.length,
	}
}

// ContainsPrefix returns true if the given containee is wholly contained
// within this Prefix. If the two Prefixes are equal, true is returned. The
// host bits in the address are ignored when testing containership.
func (me Prefix) ContainsPrefix(other Prefix) bool {
	if other.length < me.length {
		return false
	}
	mask := me.Mask().ui
	return me.Addr.ui&mask == other.Addr.ui&mask
}

// ContainsAddr returns true if the given containee is wholly contained
// within this Prefix. It is equivalent to calling ContainsPrefix with the
// given address interpreted as a host route.
func (me Prefix) ContainsAddr(other Addr) bool {
	return me.ContainsPrefix(
		Prefix{
			Addr:   other,
			length: SIZE,
		},
	)
}
