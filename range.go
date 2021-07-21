package ipv4

import (
	"fmt"
)

// Range represents a range of addresses that don't have to be aligned to
// powers of 2 like prefixes
type Range struct {
	first, last Addr
}

// NewRange returns a new range from the given first and last addresses. If
// first > last, then an empty range is returned
func NewRange(first, last Addr) (Range, error) {
	if last.LessThan(first) {
		return Range{}, fmt.Errorf("failed to create invalid range: [%s,%s]", first, last)
	}
	return Range{
		first: first,
		last:  last,
	}, nil
}

// Size returns the number of addresses in the range
func (me Range) Size() int {
	return 1 + int(me.last.ui-me.first.ui)
}

// First returns the first address in the range
func (me Range) First() Addr {
	return me.first
}

// Last returns the last address in the range
func (me Range) Last() Addr {
	return me.last
}

func (me Range) String() string {
	return fmt.Sprintf("[%s,%s]", me.first, me.last)
}

// Contains returns true iff this range entirely contains the given other range
func (me Range) Contains(other Range) bool {
	if me.Size() == 0 {
		return other.Size() == 0
	}

	return me.first.ui <= other.first.ui && other.last.ui <= me.last.ui
}
