//go:build go1.18

package ipv4

// FixedTable is like a Table except this its contents are frozen
type FixedTable[T any] FixedITable

// NewFixedTable returns an initialized but empty FixedTable
func NewFixedTable[T any]() FixedTable[T] {
	return (FixedTable[T])(NewFixedITable())
}

// Table returns a mutable table initialized with the contents of this one. Due to
// the COW nature of the underlying datastructure, it is very cheap to copy
// these -- effectively a pointer copy.
func (me FixedTable[T]) Table() Table[T] {
	return (Table[T])(
		(FixedITable)(me).Table(),
	)
}

// Size returns the number of exact prefixes stored in the table
func (me FixedTable[T]) Size() int64 {
	return (FixedITable)(me).Size()
}

// Get returns the value in the table associated with the given network prefix
// with an exact match: both the IP and the prefix length must match. If an
// exact match is not found, found is false and value is nil and should be
// ignored.
func (me FixedTable[T]) Get(prefix PrefixI) (T, bool) {
	i, b := (FixedITable)(me).Get(prefix)
	if !b {
		var t T
		return t, b
	}
	return i.(T), b
}

// LongestMatch returns the value in the table associated with the given network
// prefix using a longest prefix match. If a match is found, it returns a
// Prefix representing the longest prefix matched. If a match is *not* found,
// matched is MatchNone and the other fields should be ignored
func (me FixedTable[T]) LongestMatch(searchPrefix PrefixI) (value T, matched Match, prefix Prefix) {
	i, matched, prefix := (FixedITable)(me).LongestMatch(searchPrefix)
	if matched == MatchNone {
		var t T
		return t, matched, prefix
	}
	return i.(T), matched, prefix
}

// TableCallback is the signature of the callback functions that can be passed to
// Walk or WalkAggregates to handle each prefix/value combination.
//
// Each invocation of your callback should return true if iteration should
// continue (as long as another key / value pair exists) or false to stop
// iterating and return immediately (meaning your callback will not be called
// again).
type TableCallback[T any] func(Prefix, T) bool

// Walk invokes the given callback function for each prefix/value pair in
// the table in lexigraphical order.
//
// It returns false if iteration was stopped due to a callback return false or
// true if it iterated all items.
func (me FixedTable[T]) Walk(callback TableCallback[T]) bool {
	return (FixedITable)(me).Walk(func(p Prefix, i interface{}) bool {
		return callback(p, i.(T))
	})
}

// WalkAggregates invokes then given callback function for each prefix/value
// pair in the table, aggregated by value, in lexigraphical order.
//
// If two prefixes table to the same value, one contains the other, and there is
// no intermediate prefix between the two with a different value then only the
// broader prefix will be visited with the value.
//
// 1. The values stored must be comparable to be aggregable. Prefixes get
//    aggregated only where their values compare equal.
// 2. The set of prefix/value pairs visited is the minimal set such that any
//    longest prefix match against the aggregated set will always return the
//    same value as the same match against the non-aggregated set.
// 3. The aggregated and non-aggregated sets of prefixes may be disjoint.
//
// It returns false if iteration was stopped due to a callback return false or
// true if it iterated all items.
func (me FixedTable[T]) WalkAggregates(callback TableCallback[T]) bool {
	return (FixedITable)(me).WalkAggregates(func(p Prefix, i interface{}) bool {
		return callback(p, i.(T))
	})
}
