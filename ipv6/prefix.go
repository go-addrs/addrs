package ipv6

import (
	"fmt"
	"net"
)

// Prefix represents an IP prefix which is formally an Address plus a Mask. It
// is stored in a more space-efficient way by storing the number of 1s in the
// Mask as a length.
// Note that any bits in the address can be 0 or 1 regardless if they in the
// first `length` bits or not. This allows storing an IP address in CIDR
// notation with both the network and host parts of the address.
// The zero value of a Prefix is "::/0"
type Prefix struct {
	addr   Address
	length uint32
}

// PrefixI is something that can be treated as a Prefix by calling .Prefix().
// It is possible to pass nil as a PrefixI. In that case, it will be treated as
// if a default zero-value Prefix{} were passed which is equivalent to
// "::/0".
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
		return Prefix{}, fmt.Errorf("failed to convert IPNet with size != 128")
	}
	addr, err := AddressFromNetIP(net.IP)
	if err != nil {
		return Prefix{}, err
	}
	return Prefix{
		addr:   addr,
		length: uint32(ones),
	}, nil
}

// PrefixFromAddressMask combines the address and mask into a prefix
func PrefixFromAddressMask(address Address, mask Mask) Prefix {
	return Prefix{
		addr:   address,
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
		return nil, fmt.Errorf("failed to parse prefix: %s", prefix)
	}
	return &net.IPNet{IP: ip, Mask: ipNet.Mask}, nil
}

// ParsePrefix returns the net represented by `addr` in ipv6 hex colon CIDR
// notation. If the prefix cannot be parsed, then error is non-nil and the
// prefix returned must be ignored.
func ParsePrefix(prefix string) (Prefix, error) {
	ipNet, err := parseNet(prefix)
	if err != nil {
		return Prefix{}, err
	}
	return PrefixFromNetIPNet(ipNet)
}

// Address returns the address part of the Prefix, including host bits
func (me Prefix) Address() Address {
	return me.addr
}

// Prefix implements PrefixI
func (me Prefix) Prefix() Prefix {
	return me
}

// ToNetIPNet returns a *net.IPNet representation of this prefix
func (me Prefix) ToNetIPNet() *net.IPNet {
	return &net.IPNet{
		IP:   me.addr.ToNetIP(),
		Mask: net.CIDRMask(me.Length(), addressSize),
	}
}

// lessThan reports whether this Prefix comes strictly before `other`
// lexigraphically.
func (me Prefix) lessThan(other Prefix) bool {
	meNet, otherNet := me.Network(), other.Network()
	if meNet.addr != otherNet.addr {
		return meNet.addr.lessThan(otherNet.addr)
	}
	if me.length == other.length {
		return me.Host().addr.lessThan(other.Host().addr)
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
	network := me.addr.ui.and(me.Mask().ui)
	return Prefix{
		addr: Address{
			ui: network,
		},
		length: me.length,
	}
}

// Host returns a new Prefix with the first `length` bits zeroed out so
// that only the bits in the `host` part of the prefix are present
func (me Prefix) Host() Prefix {
	host := me.addr.ui.and(me.Mask().ui.complement())
	return Prefix{
		addr: Address{
			ui: host,
		},
		length: me.length,
	}
}

// String returns the string representation of this prefix in colon cidr
// format (e.g 2001::1/64)
func (me Prefix) String() string {
	return fmt.Sprintf("%s/%d", me.addr.String(), me.Length())
}

// Uint64 returns the address and mask as uint64s
func (me Prefix) Uint64() (addressHigh, addressLow, maskHigh, maskLow uint64) {
	addressHigh, addressLow = me.addr.Uint64()
	maskHigh, maskLow = me.Mask().Uint64()
	return
}

// prefixUpperLimit returns a new Prefix with all bits after `length` set to 1s. Note
// that this method ignores special cases where a broadcast address doesn't
// make sense like in a host route or point-to-point prefix (/128 and /127). It
// just does the math.
func (me Prefix) prefixUpperLimit() Prefix {
	network := me.addr.ui.or(me.Mask().ui.complement())
	return Prefix{
		addr: Address{
			ui: network,
		},
		length: me.length,
	}
}

// Range returns the range that includes the same addresses as the prefix
// It ignores any bits set in the host part of the address.
func (me Prefix) Range() Range {
	// Note: this error can be ignored by design
	r, _ := NewRange(me.Network().addr, me.prefixUpperLimit().addr)
	return r
}

// Halves returns two prefixes that add up to the current one
// if the prefix is a /128, the return value is undefined
func (me Prefix) Halves() (a, b Prefix) {
	if me.length < 128 {
		base := me.Network().addr
		a = Prefix{
			Address{base.ui},
			me.length + 1,
		}
		b = Prefix{
			Address{base.ui.or(uint128{0x8000000000000000, 0}.rightShift(int(me.length)))},
			me.length + 1,
		}
	}
	return
}
