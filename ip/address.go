package ip

import (
	"fmt"
	"net"
	"strings"

	"gopkg.in/addrs.v1/ipv4"
	"gopkg.in/addrs.v1/ipv6"
)

// Address is the common interface implemented by both IPv4 and IPv6 addresses
type Address interface {
	String() string
	ToNetIP() net.IP
	NumBits() int
}

var _ Address = ipv4.Address{}
var _ Address = ipv6.Address{}

// AddressFromString returns an instance of an ipv4.Address or ipv6.Address
func AddressFromString(address string) (Address, error) {
	if strings.Contains(address, ":") {
		return ipv6.AddressFromString(address)
	}
	return ipv4.AddressFromString(address)
}

// AddressFromNetIP returns an instance of an ipv4.Address or ipv6.Address
func AddressFromNetIP(ip net.IP) (Address, error) {
	switch len(ip) {
	case net.IPv6len:
		return ipv6.AddressFromNetIP(ip)
	case net.IPv4len:
		return ipv4.AddressFromNetIP(ip)
	default:
		return nil, fmt.Errorf("invalid net.IP of size %d", len(ip))
	}
}

// PrefixFromAddress returns the host route prefix (i.e. /32) from the given
// route. If the address passed in is not an ipv4.Address or ipv6.Address, then
// nil is returned.
func PrefixFromAddress(address Address) Prefix {
	switch address := address.(type) {
	case ipv4.Address:
		return address.Prefix()
	case ipv6.Address:
		return address.Prefix()
	default:
		return nil
	}
}
