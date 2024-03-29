package ip

import (
	"fmt"

	"gopkg.in/addrs.v1/ipv4"
	"gopkg.in/addrs.v1/ipv6"
)

// Range is the common interface implemented by both IPv4 and IPv6 Range
type Range interface {
	String() string
}

var _ Range = ipv4.Range{}
var _ Range = ipv6.Range{}

func RangeFromAddresses(first, last Address) (Range, bool, error) {
	switch first := first.(type) {
	case ipv6.Address:
		last, ok := last.(ipv6.Address)
		if !ok {
			return nil, false, fmt.Errorf("first address family (ipv6) doesn't match the last family")
		}
		r, empty := ipv6.RangeFromAddresses(first, last)
		return r, empty, nil
	case ipv4.Address:
		last, ok := last.(ipv4.Address)
		if !ok {
			return nil, false, fmt.Errorf("first address family (ipv4) doesn't match the last family")
		}
		r, empty := ipv4.RangeFromAddresses(first, last)
		return r, empty, nil
	}
	return nil, false, fmt.Errorf("unknown first address family")
}

// FirstFromRange returns the first address of the given range. If the range
// passed in is not an ipv4.Range or ipv6.Range, then nil is returned.
func FirstFromRange(r Range) Range {
	switch r := r.(type) {
	case ipv4.Range:
		return r.First()
	case ipv6.Range:
		return r.First()
	default:
		return nil
	}
}

// LastFromRange returns the first address of the given range. If the range
// passed in is not an ipv4.Range or ipv6.Range, then nil is returned.
func LastFromRange(r Range) Range {
	switch r := r.(type) {
	case ipv4.Range:
		return r.Last()
	case ipv6.Range:
		return r.Last()
	default:
		return nil
	}
}
