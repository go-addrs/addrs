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
	Address Address
	length  uint32
}

// PrefixI is something that can be treated as a Prefix by calling .Prefix().
// This includes the following types: Address and Prefix
type PrefixI interface {
	Prefix() Prefix
}

var _ PrefixI = Address{}
var _ PrefixI = Prefix{}

// PrefixFromNetIPNet converts the given *net.IPNet to a Prefix
func PrefixFromNetIPNet(net *net.IPNet) (Prefix, error) {
	if net == nil {
		return Prefix{}, fmt.Errorf("failed to convert nil *net.IPNet")
	}
	ones, bits := net.Mask.Size()
	if bits != addressSize {
		return Prefix{}, fmt.Errorf("failed to convert IPNet with size != 32")
	}
	addr, err := AddressFromNetIP(net.IP)
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
	return PrefixFromNetIPNet(ipNet)
}

// Prefix implements PrefixI
func (me Prefix) Prefix() Prefix {
	return me
}

// ToNetIPNet returns a *net.IPNet representation of this prefix
func (me Prefix) ToNetIPNet() *net.IPNet {
	return &net.IPNet{
		IP:   me.Address.ToNetIP(),
		Mask: net.CIDRMask(me.Length(), addressSize),
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
	if !meNet.Address.Equal(otherNet.Address) {
		return meNet.Address.LessThan(otherNet.Address)
	}
	if me.length == other.length {
		return me.Host().Address.LessThan(other.Host().Address)
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
// sense like in a host route or point-to-point prefix (/32 and /31). It just
// does the math.
func (me Prefix) Network() Prefix {
	network := me.Address.ui & me.Mask().ui
	return Prefix{
		Address: Address{
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
	network := me.Address.ui | ^me.Mask().ui
	return Prefix{
		Address: Address{
			ui: network,
		},
		length: me.length,
	}
}

// Host returns a new Prefix with the first `length` bits zeroed out so
// that only the bits in the `host` part of the prefix are present
func (me Prefix) Host() Prefix {
	host := me.Address.ui & ^me.Mask().ui
	return Prefix{
		Address: Address{
			ui: host,
		},
		length: me.length,
	}
}

// Contains returns true if the given containee is wholly contained within this
// Prefix. If the two Prefixes are equal, true is returned. The host bits in
// the address are ignored when testing containership.
func (me Prefix) Contains(other PrefixI) bool {
	prefix := other.Prefix()
	if prefix.length < me.length {
		return false
	}
	mask := me.Mask().ui
	return me.Address.ui&mask == prefix.Address.ui&mask
}

// Size returns the number of addresses in the prefix, including network and
// broadcast addresses. It ignores any bits set in the host part of the address.
func (me Prefix) Size() int64 {
	return 1 << (addressSize - me.Length())
}

// String returns the string representation of this prefix in dotted-quad cidr
// format (e.g 10.224.24.1/24)
func (me Prefix) String() string {
	return fmt.Sprintf("%s/%d", me.Address.String(), me.Length())
}

// Uint32 returns the address and mask as uint32s
func (me Prefix) Uint32() (address, mask uint32) {
	address = me.Address.Uint32()
	mask = me.Mask().Uint32()
	return
}

// AddressCallback is the type of function passed to walk individual addresses
//
// Each invocation of your callback should return true if iteration should
// continue (as long as another key / value pair exists) or false to stop
// iterating and return immediately (meaning your callback will not be called
// again).
type AddressCallback func(Address) bool

// WalkAddresses visits all of the addresses in the prefix in lexigraphical
// order
//
// It returns false if iteration was stopped due to a callback return false or
// true if it iterated all items.
func (me Prefix) WalkAddresses(callback AddressCallback) bool {
	for a := me.Network().Address.Uint32(); a <= me.Broadcast().Address.Uint32(); a++ {
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
	r, _ := NewRange(me.Network().Address, me.Broadcast().Address)
	return r
}

// Halves returns two prefixes that add up to the current one
// if the prefix is a /32, the return value is undefined
func (me Prefix) Halves() (a, b Prefix) {
	if me.length < 32 {
		base := me.Network().Address
		a = Prefix{
			Address{base.ui},
			me.length + 1,
		}
		b = Prefix{
			Address{base.ui | (0x80000000 >> me.length)},
			me.length + 1,
		}
	}
	return
}

// FixedSet returns the set that includes the same addresses as the prefix
// It ignores any bits set in the host part of the address.
func (me Prefix) FixedSet() FixedSet {
	return FixedSet{
		trie: setNodeFromPrefix(me),
	}
}
