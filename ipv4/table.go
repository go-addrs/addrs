//go:build go1.18
// +build go1.18

package ipv4

// Table_ is a mutable version of Table, allowing inserting, replacing, or
// removing elements in various ways. You can use it as a Table builder or on
// its own.
//
// The zero value of a Table_ is unitialized. Reading it is equivalent to
// reading an empty Table_. Attempts to modify it will result in a panic.
// Always use NewTable_() to get an initialized Table_.
type Table_[T any] struct {
	t ITable_
}

// NewTable_ returns a new fully-initialized Table_
func NewTable_[T any]() Table_[T] {
	return Table_[T]{NewITable_()}
}

// Size returns the number of exact prefixes stored in the table
func (me Table_[T]) Size() int64 {
	if me.t.m == nil {
		return 0
	}
	return me.t.Size()
}

// Insert inserts the given prefix with the given value into the table.
// If an entry with the same prefix already exists, it will not overwrite it
// and return false.
func (me Table_[T]) Insert(prefix PrefixI, value T) (succeeded bool) {
	return me.t.Insert(prefix, value)
}

// Update inserts the given prefix with the given value into the table. If the
// prefix already existed, it updates the associated value in place and return
// true. Otherwise, it returns false.
func (me Table_[T]) Update(prefix PrefixI, value T) (succeeded bool) {
	return me.t.Update(prefix, value)
}

// InsertOrUpdate inserts the given prefix with the given value into the table.
// If the prefix already existed, it updates the associated value in place.
func (me Table_[T]) InsertOrUpdate(prefix PrefixI, value T) {
	me.t.InsertOrUpdate(prefix, value)
}

// Get returns the value in the table associated with the given network prefix
// with an exact match: both the IP and the prefix length must match. If an
// exact match is not found, found is false and value is nil and should be
// ignored.
func (me Table_[T]) Get(prefix PrefixI) (t T, ok bool) {
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
func (me Table_[T]) GetOrInsert(prefix PrefixI, value T) T {
	var rv T
	rv, _ = me.t.GetOrInsert(prefix, value).(T)
	return rv
}

// LongestMatch returns the value associated with the given network prefix
// using a longest prefix match. If a match is found, it returns true and the
// Prefix matched, which may be equal to or shorter than the one passed. If no
// match is found, returns false and the other fields must be ignored.
func (me Table_[T]) LongestMatch(searchPrefix PrefixI) (value T, found bool, prefix Prefix) {
	if me.t.m == nil {
		return value, found, Prefix{}
	}
	var v interface{}
	v, found, prefix = me.t.LongestMatch(searchPrefix)
	value, _ = v.(T)
	return value, found, prefix
}

// Remove removes the given prefix from the table with its associated value and
// returns true if it was found. Only a prefix with an exact match will be
// removed. If no entry with the given prefix exists, it will do nothing and
// return false.
func (me Table_[T]) Remove(prefix PrefixI) (succeeded bool) {
	return me.t.Remove(prefix)
}

// Table returns an immutable snapshot of this Table_. Due to the COW
// nature of the underlying datastructure, it is very cheap to create these --
// effectively a pointer copy.
func (me Table_[T]) Table() Table[T] {
	if me.t.m == nil {
		return Table[T]{}
	}
	return Table[T]{
		*me.t.m,
	}
}

// Table is a structure that maps IP prefixes to values. For example, the
// following values can all exist as distinct prefix/value pairs in the table.
//
//     10.0.0.0/16 -> 1
//     10.0.0.0/24 -> 1
//     10.0.0.0/32 -> 2
//
// The table supports looking up values based on a longest prefix match and also
// supports efficient aggregation of prefix/value pairs based on equality of
// values. See the README.md file for a more detailed discussion.
//
// The zero value of a Table is an empty table
// Table is immutable. For a mutable equivalent, see Table_.
type Table[T any] struct {
	t ITable
}

// Table_ returns a mutable table initialized with the contents of this one. Due to
// the COW nature of the underlying datastructure, it is very cheap to copy
// these -- effectively a pointer copy.
func (me Table[T]) Table_() Table_[T] {
	return Table_[T]{
		ITable_{
			&me.t,
		},
	}
}

// Size returns the number of exact prefixes stored in the table
func (me Table[T]) Size() int64 {
	return me.t.Size()
}

// Get returns the value in the table associated with the given network prefix
// with an exact match: both the IP and the prefix length must match. If an
// exact match is not found, found is false and value is nil and should be
// ignored.
func (me Table[T]) Get(prefix PrefixI) (t T, ok bool) {
	i, b := me.t.Get(prefix)
	if !b {
		return t, b
	}
	t, _ = i.(T)
	return t, b
}

// LongestMatch returns the value associated with the given network prefix
// using a longest prefix match. If a match is found, it returns true and the
// Prefix matched, which may be equal to or shorter than the one passed. If no
// match is found, returns false and the other fields must be ignored.
func (me Table[T]) LongestMatch(searchPrefix PrefixI) (value T, found bool, prefix Prefix) {
	v, found, prefix := me.t.LongestMatch(searchPrefix)
	if !found {
		return value, false, prefix
	}
	value, _ = v.(T)
	return value, true, prefix
}

// Aggregate returns a new aggregated table as described below.
//
// It combines aggregable prefixes that are either adjacent to each other with
// the same prefix length or contained within another prefix with a shorter
// length.
//
// Prefixes are only considered aggregable if their values compare equal. This
// is useful for aggregating prefixes where the next hop is the same but not
// where they're different. Values that can be compared with == or implement
// `IEqual(interface{})` can be used in aggregation.
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
func (me Table[T]) Aggregate() Table[T] {
	return Table[T]{
		me.t.Aggregate(),
	}
}

// Walk invokes the given callback function for each prefix/value pair in
// the table in lexigraphical order.
//
// It returns false if iteration was stopped due to a callback return false or
// true if it iterated all items.
func (me Table[T]) Walk(callback func(Prefix, T) bool) bool {
	return me.t.Walk(func(p Prefix, i interface{}) bool {
		var t T
		t, _ = i.(T)
		return callback(p, t)
	})
}

// Diff invokes the given callback functions for each prefix/value pair in the
// table in lexigraphical order.
//
// It takes three callbacks, the first two handle prefixes that only exist on
// the left and right side tables respectively. The third callback handles
// prefixes that exist in both tables but with different values. No callback is
// called for prefixes that exist in both tables with the same values.
//
// It returns false if iteration was stopped due to a callback returning false
// or true if it iterated all items.
func (me Table[T]) Diff(other Table[T], left, right func(Prefix, T) bool, changed func(p Prefix, left, right T) bool) bool {
	trieHandler := trieDiffHandler{}
	if left != nil {
		trieHandler.Removed = func(n *trieNode) bool {
			var t T
			t, _ = n.Data.(T)
			return left(n.Prefix, t)
		}
	}
	if right != nil {
		trieHandler.Added = func(n *trieNode) bool {
			var t T
			t, _ = n.Data.(T)
			return right(n.Prefix, t)
		}
	}
	if changed != nil {
		trieHandler.Modified = func(l, r *trieNode) bool {
			var lt, rt T
			lt, _ = l.Data.(T)
			rt, _ = r.Data.(T)
			return changed(l.Prefix, lt, rt)
		}
	}
	return me.t.trie.Diff(other.t.trie, trieHandler)
}
