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

// Minus returns a slice of ranges resulting from subtracting the given range
// The slice will contain from 0 to 2 new ranges depending on how they overlap
func (me Range) Minus(other Range) []Range {
	result := []Range{}
	if me.first.LessThan(other.first) {
		result = append(result, Range{
			me.first,
			MinAddr(other.prev(), me.last),
		})
	}
	if other.last.LessThan(me.last) {
		result = append(result, Range{
			MaxAddr(me.first, other.next()),
			me.last,
		})
	}
	return result
}

// prev returns the address just before the range (or maxint) if the range
// starts at the beginning of the IP space due to overflow)
func (me Range) prev() Addr {
	return Addr{me.first.ui - 1}
}

// next returns the next address after the range (or 0 if the range goes to the
// end of the IP space due to overflow)
func (me Range) next() Addr {
	return Addr{me.last.ui + 1}
}
