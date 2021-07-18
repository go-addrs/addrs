package ipv4

import (
	"fmt"
)

// Range represents a range of addresses that don't have to be aligned to
// powers of 2 like prefixes
type Range struct {
	first, last Addr
}

// EmptyRange returns a range with 0 IP addresses in it.
func EmptyRange() Range {
	return Range{
		first: Addr{ui: 0xffffffff},
		last:  Addr{ui: 0x0},
	}
}

// NewRange returns a new range from the given first and last addresses. If
// first > last, then an empty range is returned
func NewRange(first, last Addr) Range {
	if last.LessThan(first) {
		return EmptyRange()
	}
	return Range{
		first: first,
		last:  last,
	}
}

// Size returns the number of addresses in the range
func (me Range) Size() int {
	if me.last.ui < me.first.ui {
		return 0
	}
	return 1 + int(me.last.ui-me.first.ui)
}

// First returns the first address in the range. If the range is empty, exists
// is false and the address must be ignored.
func (me Range) First() (exists bool, addr Addr) {
	if 0 < me.Size() {
		exists, addr = true, me.first
	}
	return
}

// Last returns the last address in the range. If the range is empty, exists
// is false and the address must be ignored.
func (me Range) Last() (exists bool, addr Addr) {
	if 0 < me.Size() {
		exists, addr = true, me.last
	}
	return
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
