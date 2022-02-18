package ipv4

import (
	"fmt"
)

// Range represents a range of addresses that don't have to be aligned to
// powers of 2 like a Prefix does.
//
// Note that there is no instantiation of an empty range (.NumAddresses() == 0)
// because .First() and .Last() would not make sense.
//
// The zero value of a Range is "[0.0.0.0, 0.0.0.0]" (.NumAddresses() == 1)
type Range struct {
	first, last Address
}

// NewRange returns a new range from the given first and last addresses. If
// first > last, then empty is set to true and the returned range must be
// ignored. There is no valid instantiation of an empty range.
func NewRange(first, last Address) (r Range, empty bool) {
	if last.lessThan(first) {
		return Range{}, true
	}
	return Range{
		first: first,
		last:  last,
	}, false
}

// NumAddresses returns the number of addresses in the range
func (me Range) NumAddresses() int64 {
	return 1 + int64(me.last.ui-me.first.ui)
}

// First returns the first address in the range
func (me Range) First() Address {
	return me.first
}

// Last returns the last address in the range
func (me Range) Last() Address {
	return me.last
}

func (me Range) String() string {
	return fmt.Sprintf("[%s,%s]", me.first, me.last)
}

// Contains returns true iff this range entirely contains the given other range
func (me Range) Contains(other SetI) bool {
	return me.Set().Contains(other)
}

// Minus returns a slice of ranges resulting from subtracting the given range
// The slice will contain from 0 to 2 new ranges depending on how they overlap
func (me Range) Minus(other Range) []Range {
	result := []Range{}
	if me.first.lessThan(other.first) {
		result = append(result, Range{
			me.first,
			minAddress(other.prev(), me.last),
		})
	}
	if other.last.lessThan(me.last) {
		result = append(result, Range{
			maxAddress(me.first, other.next()),
			me.last,
		})
	}
	return result
}

// Plus returns a slice of ranges resulting from adding the given range to this
// one. The slice will contain 1 or 2 new ranges depending on how/if they
// overlap. If two ranges are returned, they will be returned sorted
// lexigraphically by their first address.
func (me Range) Plus(other Range) []Range {
	plus := func(a, b Range) []Range {
		if a.next().lessThan(b.first) {
			return []Range{a, b}
		}
		return []Range{
			Range{a.first, maxAddress(a.last, b.last)},
		}
	}
	if me.first.lessThan(other.first) {
		return plus(me, other)
	}
	return plus(other, me)
}

// Set returns a Set_ containing the same ips as this range
func (me Range) Set() Set {
	return Set{
		trie: setNodeFromRange(me),
	}
}

// prev returns the address just before the range (or maxint) if the range
// starts at the beginning of the IP space due to overflow)
func (me Range) prev() Address {
	return Address{me.first.ui - 1}
}

// next returns the next address after the range (or 0 if the range goes to the
// end of the IP space due to overflow)
func (me Range) next() Address {
	return Address{me.last.ui + 1}
}
