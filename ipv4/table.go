//go:build go1.18
// +build go1.18

package ipv4

// Table is a structure that maps IP prefixes to values. For example, you can
// insert the following values and they will all exist as distinct prefix/value
// pairs in the table.
//
//     10.0.0.0/16 -> 1
//     10.0.0.0/24 -> 1
//     10.0.0.0/32 -> 2
//
// The table supports looking up values based on a longest prefix match and also
// supports efficient aggregation of prefix/value pairs based on equality of
// values. See the README.md file for a more detailed discussion..
//
// The zero value of a Table is unitialized. Reading it is equivalent to
// reading an empty Table. Attempts to modify it will result in a panic. Always
// use NewTable() to get a modifyable Table.
type Table[T any] struct {
	t ITable
}

// NewTable returns a new fully-initialized Table
func NewTable[T any]() Table[T] {
	return Table[T]{NewITable()}
}

// Size returns the number of exact prefixes stored in the table
func (me Table[T]) Size() int64 {
	if me.t.m == nil {
		return 0
	}
	return me.t.Size()
}

// Insert inserts the given prefix with the given value into the table.
// If an entry with the same prefix already exists, it will not overwrite it
// and return false.
func (me Table[T]) Insert(prefix PrefixI, value T) (succeeded bool) {
	return me.t.Insert(prefix, value)
}

// Update inserts the given prefix with the given value into the table. If the
// prefix already existed, it updates the associated value in place and return
// true. Otherwise, it returns false.
func (me Table[T]) Update(prefix PrefixI, value T) (succeeded bool) {
	return me.t.Update(prefix, value)
}

// InsertOrUpdate inserts the given prefix with the given value into the table.
// If the prefix already existed, it updates the associated value in place.
func (me Table[T]) InsertOrUpdate(prefix PrefixI, value T) {
	me.t.InsertOrUpdate(prefix, value)
}

// Get returns the value in the table associated with the given network prefix
// with an exact match: both the IP and the prefix length must match. If an
// exact match is not found, found is false and value is nil and should be
// ignored.
func (me Table[T]) Get(prefix PrefixI) (t T, ok bool) {
	if me.t.m == nil {
		return t, false
	}
	var value interface{}
	value, ok = me.t.Get(prefix)
	t, _ = value.(T)
	return t, ok
}

// GetOrInsert returns the value associated with the given prefix if it already
// exists. If it does not exist, it inserts it with the given value and returns
// that.
func (me Table[T]) GetOrInsert(prefix PrefixI, value T) T {
	var rv T
	rv, _ = me.t.GetOrInsert(prefix, value).(T)
	return rv
}

// LongestMatch returns the value in the table associated with the given network
// prefix using a longest prefix match. If a match is found, it returns a
// Prefix representing the longest prefix matched. If a match is *not* found,
// matched is MatchNone and the other fields should be ignored
func (me Table[T]) LongestMatch(searchPrefix PrefixI) (t T, matched Match, prefix Prefix) {
	if me.t.m == nil {
		return t, MatchNone, Prefix{}
	}
	var value interface{}
	value, matched, prefix = me.t.LongestMatch(searchPrefix)
	t, _ = value.(T)
	return t, matched, prefix
}

// Remove removes the given prefix from the table with its associated value and
// returns true if it was found. Only a prefix with an exact match will be
// removed. If no entry with the given prefix exists, it will do nothing and
// return false.
func (me Table[T]) Remove(prefix PrefixI) (succeeded bool) {
	return me.t.Remove(prefix)
}

// FixedTable returns an immutable snapshot of this Table. Due to the COW
// nature of the underlying datastructure, it is very cheap to create these --
// effectively a pointer copy.
func (me Table[T]) FixedTable() FixedTable[T] {
	if me.t.m == nil {
		return FixedTable[T]{}
	}
	return FixedTable[T]{
		*me.t.m,
	}
}

// FixedTable is like a Table except this its contents are frozen
// The zero value of a FixedTable is an empty table
type FixedTable[T any] struct {
	t FixedITable
}

// Table returns a mutable table initialized with the contents of this one. Due to
// the COW nature of the underlying datastructure, it is very cheap to copy
// these -- effectively a pointer copy.
func (me FixedTable[T]) Table() Table[T] {
	return Table[T]{
		ITable{
			&me.t,
		},
	}
}

// Size returns the number of exact prefixes stored in the table
func (me FixedTable[T]) Size() int64 {
	return me.t.Size()
}

// Get returns the value in the table associated with the given network prefix
// with an exact match: both the IP and the prefix length must match. If an
// exact match is not found, found is false and value is nil and should be
// ignored.
func (me FixedTable[T]) Get(prefix PrefixI) (T, bool) {
	i, b := me.t.Get(prefix)
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
	i, matched, prefix := me.t.LongestMatch(searchPrefix)
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
	return FixedTable[T]{
		me.t.Aggregate(),
	}
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
	return me.t.Walk(func(p Prefix, i interface{}) bool {
		var t T
		t, _ = i.(T)
		return callback(p, t)
	})
}

// TableModifiedCallback is the signature of the callback functions to handle
// a modified entry when diffing. It is passed the prefix and the values before
// and after the change.
//
// Each invocation of your callback should return true if iteration should
// continue (as long as another key / value pair exists) or false to stop
// iterating and return immediately (meaning your callback will not be called
// again).
type TableModifiedCallback[T any] func(p Prefix, left, right T) bool

// DiffHandler is a struct passed to Diff to handle changes found between the
// left and right tables. Removed is called for prefixes that appear in
// the left table but not the right, Added is called for prefixes that appear
// in the right but not the left, and Modified is called for prefixes that
// appear in both but have different values.
//
// Any of the handlers can be left out safely -- they will default to nil. In
// that case, Diff will skip those cases.
type DiffHandler[T any] struct {
	Removed  TableCallback[T]
	Added    TableCallback[T]
	Modified TableModifiedCallback[T]
}

// Diff invokes the given callback functions for each prefix/value pair in the
// table in lexigraphical order.
//
// It returns false if iteration was stopped due to a callback returning false
// or true if it iterated all items.
func (me FixedTable[T]) Diff(other FixedTable[T], handler DiffHandler[T]) bool {
	trieHandler := trieDiffHandler{}
	if handler.Removed != nil {
		trieHandler.Removed = func(n *trieNode) bool {
			return handler.Removed(n.Prefix, n.Data.(T))
		}
	}
	if handler.Added != nil {
		trieHandler.Added = func(n *trieNode) bool {
			return handler.Added(n.Prefix, n.Data.(T))
		}
	}
	if handler.Modified != nil {
		trieHandler.Modified = func(l, r *trieNode) bool {
			return handler.Modified(l.Prefix, l.Data.(T), r.Data.(T))
		}
	}
	return me.t.trie.Diff(other.t.trie, trieHandler)
}
