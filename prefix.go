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
