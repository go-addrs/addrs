package ipv6

import (
	"fmt"
	"net"
)

// Prefix represents an IPv6 prefix which is IPv6 address along with prefix
// length, or the number of bits which are significant in the network portion.
// Note that any bits in the address can be 0 or 1 regardless if they in the
// first `length` bits or not. This allows storing an IP address in CIDR
// notation with both the network and host parts of the address.
type Prefix struct {
	Address Address
	length  uint32
}

// IP returns the address part of the prefix alone
func (me Prefix) IP() Address {
	return me.Address
}

// PrefixFromUint64 returns the IPv6 address from its 128 bit unsigned representation
func PrefixFromUint64(high uint64, low uint64, length int) (Prefix, error) {
	if length < 0 || SIZE < length {
		return Prefix{}, fmt.Errorf("failed to convert prefix, length %d isn't between 0 and 128", length)
	}
	return Prefix{Address{high, low}, uint32(length)}, nil
}

// PrefixFromBytes returns the IPv6 address.
func PrefixFromBytes(bytes []byte, length int) (Prefix, error) {
	if length < 0 || SIZE < length {
		return Prefix{}, fmt.Errorf("failed to convert prefix, length %d isn't between 0 and 128", length)
	}
	return Prefix{
		Address: AddressFromBytes(bytes),
		length:  uint32(length),
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
	addr, err := AddressFromStdIP(net.IP)
	if err != nil {
		return Prefix{}, err
	}
	return Prefix{
		Address: addr,
		length:  uint32(ones),
	}, nil
}

// PrefixFromAddressMask combines the address and mask into a prefix
func PrefixFromAddressMask(address Address, mask Mask) Prefix {
	return Prefix{
		Address: address,
		length:  uint32(mask.Length()),
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
		return nil, fmt.Errorf("failed to parse IPv6 prefix: %s", prefix)
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
		IP:   me.Address.ToStdIP(),
		Mask: net.CIDRMask(me.Length(), SIZE),
	}
}

// Equal reports whether this IPv6 address is the same as other
func (me Prefix) Equal(other Prefix) bool {
	return me == other
}

// LessThan reports whether this IPv6 prefix comes strictly before `other`
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

// Mask returns a new Address with 1s in the first `length` bits and then 0s
// representing the network mask for this prefix.
func (me Prefix) Mask() Mask {
	return lengthToMask(me.Length())
}

// Network returns a new Prefix with all bits after `length` zeroed out so that
// only the bits in the `network` part of the prefix are present. Note that
// this method ignores special cases where a network address doesn't make
// sense like in a host route or point-to-point prefix (/128 and /127). It just
// does the math.
func (me Prefix) Network() Prefix {
	networkHigh := me.Address.high & me.Mask().high
	networkLow := me.Address.low & me.Mask().low
	return Prefix{
		Address: Address{
			high: networkHigh,
			low:  networkLow,
		},
		length: me.length,
	}
}

// Broadcast returns a new Prefix with all bits after `length` set to 1s. Note
// that this method ignores special cases where a broadcast address doesn't
// make sense like in a host route or point-to-point prefix (/128 and /127). It
// just does the math.
func (me Prefix) Broadcast() Prefix {
	networkHigh := me.Address.high | ^me.Mask().high
	networkLow := me.Address.low | ^me.Mask().low
	return Prefix{
		Address: Address{
			high: networkHigh,
			low:  networkLow,
		},
		length: me.length,
	}
}

// Host returns a new Prefix with the first `length` bits zeroed out so
// that only the bits in the `host` part of the prefix are present
func (me Prefix) Host() Prefix {
	hostHigh := me.Address.high & ^me.Mask().high
	hostLow := me.Address.low & ^me.Mask().low
	return Prefix{
		Address: Address{
			high: hostHigh,
			low:  hostLow,
		},
		length: me.length,
	}
}

// TODO
// ContainsPrefix returns true if the given containee is wholly contained
// within this Prefix. If the two Prefixes are equal, true is returned. The
// host bits in the address are ignored when testing containership.
func (me Prefix) ContainsPrefix(other Prefix) bool {
	if other.length < me.length {
		return false
	}
	mask := me.Mask().ui
	return me.Address.ui&mask == other.Address.ui&mask
}

// ContainsAddress returns true if the given containee is wholly contained
// within this Prefix. It is equivalent to calling ContainsPrefix with the
// given address interpreted as a host route.
func (me Prefix) ContainsAddress(other Address) bool {
	return me.ContainsPrefix(
		Prefix{
			Address: other,
			length:  uint32(SIZE),
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
	return fmt.Sprintf("%s/%d", me.Address.String(), me.Length())
}

// Uint32 returns the address and mask as uint64s
func (me Prefix) Uint64() (highAddr, lowAddr, highMask, lowMask uint64) {
	highAddr, lowAddr = me.Address.Uint64()
	highMask, lowMask = me.Mask().Uint64()
	return
}

// AddressCallback is the type of function passed to Iterate over individual addresses
type AddressCallback func(Address) bool

// TODO
// Iterate visits all of the addresses in the prefix in lexigraphical order
func (me Prefix) Iterate(callback AddressCallback) bool {
	for a := me.Network().Address.Uint64(); a <= me.Broadcast().Address.Uint64(); a++ {
		if !callback(Address{a}) {
			return false
		}
	}
	return true
}

// Range returns the range that includes the same addresses as the prefix
// It ignores any bits set in the host part of the address.
func (me Prefix) Range() Range {
	// Note: this error can be ignored by design
	r, _ := NewRange(me.Network().IP(), me.Broadcast().IP())
	return r
}

// TODO
// Halves returns two prefixes that add up to the current one
// if the prefix is a /128, the return value is undefined
func (me Prefix) Halves() (a, b Prefix) {
	if me.length < 128 {
		base := me.Network().Address
		a = Prefix{
			Address{base.high, base.low},
			me.length + 1,
		}
		b = Prefix{
			Address{base.ui | (0x80000000 >> me.length)},
			me.length + 1,
		}
	}
	return
}

// Set returns the set that includes the same addresses as the prefix
// It ignores any bits set in the host part of the address.
func (me Prefix) Set() Set {
	return Set{setNodeFromPrefix(me)}
}
