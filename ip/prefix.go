package ip

import (
	"fmt"
	"net"
	"strings"

	"gopkg.in/addrs.v1/ipv4"
	"gopkg.in/addrs.v1/ipv6"
)

// Prefix is the common interface implemented by both IPv4 and IPv6 Prefix
type Prefix interface {
	Length() int
	String() string
	ToNetIPNet() *net.IPNet
}

var _ Prefix = ipv4.Prefix{}
var _ Prefix = ipv6.Prefix{}

func PrefixFromAddressMask(address Address, mask Mask) (Prefix, error) {
	switch address := address.(type) {
	case ipv6.Address:
		mask, ok := mask.(ipv6.Mask)
		if !ok {
			return nil, fmt.Errorf("address family (ipv6) doesn't match the mask family")
		}
		return ipv6.PrefixFromAddressMask(address, mask), nil
	case ipv4.Address:
		mask, ok := mask.(ipv4.Mask)
		if !ok {
			return nil, fmt.Errorf("address family (ipv4) doesn't match the mask family")
		}
		return ipv4.PrefixFromAddressMask(address, mask), nil
	}
	return nil, fmt.Errorf("unknown address family")
}

func PrefixFromString(prefix string) (Prefix, error) {
	if strings.Contains(prefix, ":") {
		return ipv6.PrefixFromString(prefix)
	}
	return ipv4.PrefixFromString(prefix)
}

func PrefixFromNetIPNet(ipn *net.IPNet) (Prefix, error) {
	if ipn == nil {
		return nil, fmt.Errorf("cannot create Prefix from nil")
	}
	switch len(ipn.IP) {
	case net.IPv6len:
		return ipv6.PrefixFromNetIPNet(ipn)
	case net.IPv4len:
		return ipv4.PrefixFromNetIPNet(ipn)
	default:
		return nil, fmt.Errorf("invalid net.IPNet with IP of size %d", len(ipn.IP))
	}
}

// AddressFromPrefix returns the address part of the given prefix. If the
// prefix passed in is not an ipv4.Prefix or ipv6.Prefix, then nil is returned.
func AddressFromPrefix(prefix Prefix) Address {
	switch prefix := prefix.(type) {
	case ipv4.Prefix:
		return prefix.Address()
	case ipv6.Prefix:
		return prefix.Address()
	default:
		return nil
	}
}

// MaskFromPrefix returns the address part of the given prefix. If the
// prefix passed in is not an ipv4.Prefix or ipv6.Prefix, then nil is returned.
func MaskFromPrefix(prefix Prefix) Mask {
	switch prefix := prefix.(type) {
	case ipv4.Prefix:
		return prefix.Mask()
	case ipv6.Prefix:
		return prefix.Mask()
	default:
		return nil
	}
}

// RangeFromPrefix returns the address part of the given prefix. If the
// prefix passed in is not an ipv4.Prefix or ipv6.Prefix, then nil is returned.
func RangeFromPrefix(prefix Prefix) Range {
	switch prefix := prefix.(type) {
	case ipv4.Prefix:
		return prefix.Range()
	case ipv6.Prefix:
		return prefix.Range()
	default:
		return nil
	}
}
