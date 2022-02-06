//go:build go1.18
// +build go1.18

package ipv4

// FixedTable is like a Table except this its contents are frozen
// The zero value of a FixedTable is an empty table
type FixedTable[T any] FixedITable

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

// Aggregate returns a new aggregated table as described below.
//
// It combines aggregable prefixes that are either adjacent to each other with
// the same prefix length or contained within another prefix with a shorter
// length.
//
// Prefixes are only considered aggregable if their values compare equal. This is
// useful for aggregating prefixes where the next hop is the same but not where
// they're different. Values that can be compared with == or implement the
// EqualComparable interface can be used.
//
// The aggregated table has the minimum set of prefix/value pairs needed to
// return the same value for any longest prefix match using a host route  as
// would be returned by the the original trie, non-aggregated. This can be
// useful, for example, to minimize the number of prefixes needed to install
// into a router's datapath to guarantee that all of the next hops are correct.
//
// If two prefixes in the original table map to the same value, one contains
// the other, and there is no intermediate prefix between them with a different
// value then only the broader prefix will appear in the resulting table.
//
// In general, routing protocols should not aggregate and then pass on the
// aggregates to neighbors as this will likely lead to poor comparisions by
// neighboring routers who receive routes aggregated differently from different
// peers.
func (me FixedTable[T]) Aggregate() FixedTable[T] {
	return (FixedTable[T])(
		(FixedITable)(me).Aggregate(),
	)
}

// TableCallback is the signature of the callback functions that can be passed to
// Walk to handle each prefix/value combination.
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
