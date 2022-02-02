//go:build go1.18
// +build go1.18

package ipv4

// Table is a structure that tables IP prefixes to values. For example, you can
// insert the following values and they will all exist as distinct prefix/value
// pairs in the table.
//
// 10.0.0.0/16 -> 1
// 10.0.0.0/24 -> 1
// 10.0.0.0/32 -> 2
//
// The table supports looking up values based on a longest prefix match and also
// supports efficient aggregation of prefix/value pairs based on equality of
// values. See the README.md file for a more detailed discussion..
//
// The zero value of a Table is unitialized. Reading it is equivalent to
// reading an empty Table. Attempts to modify it will result in a panic. Always
// use NewTable() to get a modifyable Table.
type Table[T any] ITable

// NewTable returns a new fully-initialized Table
func NewTable[T any]() Table[T] {
	return (Table[T])(NewITable())
}

// Size returns the number of exact prefixes stored in the table
func (me Table[T]) Size() int64 {
	if me.m == nil {
		return 0
	}
	return (FixedTable[T])(*me.m).Size()
}

// Insert inserts the given prefix with the given value into the table.
// If an entry with the same prefix already exists, it will not overwrite it
// and return false.
func (me Table[T]) Insert(prefix PrefixI, value T) (succeeded bool) {
	return (ITable)(me).Insert(prefix, value)
}

// Update inserts the given prefix with the given value into the table. If the
// prefix already existed, it updates the associated value in place and return
// true. Otherwise, it returns false.
func (me Table[T]) Update(prefix PrefixI, value T) (succeeded bool) {
	return (ITable)(me).Update(prefix, value)
}

// InsertOrUpdate inserts the given prefix with the given value into the table.
// If the prefix already existed, it updates the associated value in place.
func (me Table[T]) InsertOrUpdate(prefix PrefixI, value T) {
	(ITable)(me).InsertOrUpdate(prefix, value)
}

// Get returns the value in the table associated with the given network prefix
// with an exact match: both the IP and the prefix length must match. If an
// exact match is not found, found is false and value is nil and should be
// ignored.
func (me Table[T]) Get(prefix PrefixI) (T, bool) {
	if me.m == nil {
		var t T
		return t, false
	}
	return (FixedTable[T])(*me.m).Get(prefix)
}

// GetOrInsert returns the value associated with the given prefix if it already
// exists. If it does not exist, it inserts it with the given value and returns
// that.
func (me Table[T]) GetOrInsert(prefix PrefixI, value T) T {
	var rv T
	rv, _ = (ITable)(me).GetOrInsert(prefix, value).(T)
	return rv
}

// LongestMatch returns the value in the table associated with the given network
// prefix using a longest prefix match. If a match is found, it returns a
// Prefix representing the longest prefix matched. If a match is *not* found,
// matched is MatchNone and the other fields should be ignored
func (me Table[T]) LongestMatch(searchPrefix PrefixI) (value T, matched Match, prefix Prefix) {
	if me.m == nil {
		var t T
		return t, MatchNone, Prefix{}
	}
	return (FixedTable[T])(*me.m).LongestMatch(searchPrefix)
}

// Remove removes the given prefix from the table with its associated value and
// returns true if it was found. Only a prefix with an exact match will be
// removed. If no entry with the given prefix exists, it will do nothing and
// return false.
func (me Table[T]) Remove(prefix PrefixI) (succeeded bool) {
	return (ITable)(me).Remove(prefix)
}

// FixedTable returns an immutable snapshot of this Table. Due to the COW
// nature of the underlying datastructure, it is very cheap to create these --
// effectively a pointer copy.
func (me Table[T]) FixedTable() FixedTable[T] {
	return (FixedTable[T])(
		(ITable)(me).FixedTable(),
	)
}
