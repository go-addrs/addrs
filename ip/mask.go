package ip

import (
	"fmt"
	"net"

	"gopkg.in/addrs.v1/ipv4"
	"gopkg.in/addrs.v1/ipv6"
)

// Mask is the common interface implemented by both IPv4 and IPv6 netmasks
type Mask interface {
	Length() int
	String() string
	ToNetIPMask() net.IPMask
}

var _ Mask = ipv4.Mask{}
var _ Mask = ipv6.Mask{}

func MaskFromNetIPMask(mask net.IPMask) (Mask, error) {
	ones, bits := mask.Size()
	if ones < 0 || bits < 0 {
		return nil, fmt.Errorf("invalid net.IPMask /%d (%d bits)", ones, bits)
	}
	if ones > bits {
		return nil, fmt.Errorf("invalid net.IPMask /%d (%d bits)", ones, bits)
	}
	switch bits {
	case 8 * net.IPv6len:
		return ipv6.MaskFromLength(ones)
	case 8 * net.IPv4len:
		return ipv4.MaskFromLength(ones)
	}
	return nil, fmt.Errorf("invalid net.IPMask size: %d bits", bits)
}
