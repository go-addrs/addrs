package ipv4

import (
	"fmt"
	"net"
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
func PrefixFromUint32(ui uint32, length int) (Prefix, error) {
	if length < 0 || SIZE < length {
		return Prefix{}, fmt.Errorf("failed to convert prefix, length %d isn't between 0 and 32", length)
	}
	return Prefix{Addr{ui}, uint32(length)}, nil
}

// PrefixFromBytes returns the IPv4 address of the `a.b.c.d`.
func PrefixFromBytes(a, b, c, d byte, length int) (Prefix, error) {
	if length < 0 || SIZE < length {
		return Prefix{}, fmt.Errorf("failed to convert prefix, length %d isn't between 0 and 32", length)
	}
	return Prefix{
		Addr:   AddrFromBytes(a, b, c, d),
		length: uint32(length),
	}, nil
}

// PrefixFromStdIPNet converts the given *net.IPNet to a Prefix
func PrefixFromStdIPNet(net *net.IPNet) (Prefix, error) {
	if net == nil {
		return Prefix{}, fmt.Errorf("failed to convert nil *net.IPNet")
	}
	ones, bits := net.Mask.Size()
	if bits != SIZE {
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

// PrefixFromAddrMask combines the address and mask into a prefix
func PrefixFromAddrMask(address Addr, mask Mask) Prefix {
	return Prefix{
		Addr:   address,
		length: uint32(mask.Length()),
	}
}

// parseNet returns a single *net.IPNet that unifies the IP address and
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

// ToStdIPNet returns a *net.IPNet representation of this prefix
func (me Prefix) ToStdIPNet() *net.IPNet {
	return &net.IPNet{
		IP:   me.Addr.ToStdIP(),
		Mask: net.CIDRMask(me.Length(), SIZE),
	}
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

// Length returns the number of leading 1s in the mask.
func (me Prefix) Length() int {
	return int(me.length)
}

// Mask returns a new Addr with 1s in the first `length` bits and then 0s
// representing the network mask for this prefix.
func (me Prefix) Mask() Mask {
	return lengthToMask(me.Length())
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
			length: uint32(SIZE),
		},
	)
}

// Size returns the number of addresses in the prefix, including network and
// broadcast addresses. It ignores any bits set in the host part of the address.
func (me Prefix) Size() int64 {
	return 1 << (SIZE - me.Length())
}

// String returns the string representation of this prefix in dotted-quad cidr
// format (e.g 10.224.24.1/24)
func (me Prefix) String() string {
	return fmt.Sprintf("%s/%d", me.Addr.String(), me.Length())
}

// Uint32 returns the address and mask as uint32s
func (me Prefix) Uint32() (address, mask uint32) {
	address = me.Addr.Uint32()
	mask = me.Mask().Uint32()
	return
}

// Range returns the range that includes the same addresses as the prefix
// It ignores any bits set in the host part of the address.
func (me Prefix) Range() Range {
	// Note: this error can be ignored by design
	r, _ := NewRange(me.Network().IP(), me.Broadcast().IP())
	return r
}
