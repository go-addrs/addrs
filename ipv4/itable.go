package ipv4

// ITable_ is a mutable version of ITable, allowing inserting, replacing, or
// removing elements in various ways. You can use it as an ITable builder or on
// its own.
//
// The zero value of a ITable_ is unitialized. Reading it is equivalent to
// reading an empty ITable_. Attempts to modify it will result in a panic.
// Always use NewITable_() to get an initialized ITable_.
type ITable_ struct {
	// This is an abuse of ITable because it uses its package privileges
	// to turn it into a mutable one. This could be refactored to be cleaner
	// without changing the interface.

	// Be careful not to take an ITable from outside the package and turn
	// it into a mutable one. That would break the contract.
	m  *ITable
	eq comparator
}

// NewITable_ returns a new fully-initialized Table_ optimized for values that
// are comparable with ==.
func NewITable_() ITable_ {
	return ITable_{
		&ITable{},
		func(a, b interface{}) bool {
			return a == b
		},
	}
}

// NewITableCustomCompare_ returns a new fully-initialized Table_ optimized for
// data that can be compared used a comparator that you pass.
func NewITableCustomCompare_(comparator func(a, b interface{}) bool) ITable_ {
	return ITable_{
		&ITable{},
		comparator,
	}
}

// match indicates how closely the given key matches the search result
type match int

const (
	// matchNone indicates that no match was found
	matchNone match = iota
	// matchContains indicates that a match was found that contains the search key but isn't exact
	matchContains
	// matchExact indicates that a match with the same prefix
	matchExact
)

// NumEntries returns the number of exact prefixes stored in the table
func (me ITable_) NumEntries() int64 {
	if me.m == nil {
		return 0
	}
	return me.m.NumEntries()
}

// mutate should be called by any method that modifies the table in any way
func (me ITable_) mutate(mutator func() (ok bool, node *trieNode)) {
	oldNode := me.m.trie
	ok, newNode := mutator()
	if ok && oldNode != newNode {
		if !swapTrieNodePtr(&me.m.trie, oldNode, newNode) {
			panic("concurrent modification of Table_ detected")
		}
	}
}

// Insert inserts the given prefix with the given value into the table.
// If an entry with the same prefix already exists, it will not overwrite it
// and return false.
func (me ITable_) Insert(prefix PrefixI, value interface{}) (succeeded bool) {
	if me.m == nil {
		panic("cannot modify an unitialized Table_")
	}
	if prefix == nil {
		prefix = Prefix{}
	}
	var err error
	me.mutate(func() (bool, *trieNode) {
		var newHead *trieNode
		newHead, err = me.m.trie.Insert(prefix.Prefix(), value)
		if err != nil {
			return false, nil
		}
		return true, newHead
	})
	return err == nil
}

// Update inserts the given prefix with the given value into the table. If the
// prefix already existed, it updates the associated value in place and return
// true. Otherwise, it returns false.
func (me ITable_) Update(prefix PrefixI, value interface{}) (succeeded bool) {
	if me.m == nil {
		panic("cannot modify an unitialized Table_")
	}
	if prefix == nil {
		prefix = Prefix{}
	}
	var err error
	me.mutate(func() (bool, *trieNode) {
		var newHead *trieNode
		newHead, err = me.m.trie.Update(prefix.Prefix(), value)
		if err != nil {
			return false, nil
		}
		return true, newHead
	})
	return err == nil
}

// InsertOrUpdate inserts the given prefix with the given value into the table.
// If the prefix already existed, it updates the associated value in place.
func (me ITable_) InsertOrUpdate(prefix PrefixI, value interface{}) {
	if me.m == nil {
		panic("cannot modify an unitialized Table_")
	}
	if prefix == nil {
		prefix = Prefix{}
	}
	me.mutate(func() (bool, *trieNode) {
		return true, me.m.trie.InsertOrUpdate(prefix.Prefix(), value)
	})
}

// Get returns the value in the table associated with the given network prefix
// with an exact match: both the IP and the prefix length must match. If an
// exact match is not found, found is false and value is nil and should be
// ignored.
func (me ITable_) Get(prefix PrefixI) (interface{}, bool) {
	if me.m == nil {
		return nil, false
	}
	return me.m.Get(prefix)
}

// GetOrInsert returns the value associated with the given prefix if it already
// exists. If it does not exist, it inserts it with the given value and returns
// that.
func (me ITable_) GetOrInsert(prefix PrefixI, value interface{}) interface{} {
	if me.m == nil {
		panic("cannot modify an unitialized Table_")
	}
	if prefix == nil {
		prefix = Prefix{}
	}
	var node *trieNode
	me.mutate(func() (bool, *trieNode) {
		var newHead *trieNode
		newHead, node = me.m.trie.GetOrInsert(prefix.Prefix(), value)
		return true, newHead
	})
	return node.Data
}

// LongestMatch returns the value associated with the given network prefix
// using a longest prefix match. If a match is found, it returns true and the
// Prefix matched, which may be equal to or shorter than the one passed. If no
// match is found, returns false and the other fields must be ignored.
func (me ITable_) LongestMatch(prefix PrefixI) (value interface{}, found bool, matchPrefix Prefix) {
	if me.m == nil {
		return nil, false, Prefix{}
	}
	return me.m.LongestMatch(prefix)
}

// Remove removes the given prefix from the table with its associated value and
// returns true if it was found. Only a prefix with an exact match will be
// removed. If no entry with the given prefix exists, it will do nothing and
// return false.
func (me ITable_) Remove(prefix PrefixI) (succeeded bool) {
	if me.m == nil {
		panic("cannot modify an unitialized Table_")
	}
	if prefix == nil {
		prefix = Prefix{}
	}
	var err error
	me.mutate(func() (bool, *trieNode) {
		var newHead *trieNode
		newHead, err = me.m.trie.Delete(prefix.Prefix())
		return true, newHead
	})
	return err == nil
}

// Table returns an immutable snapshot of this ITable_. Due to the COW
// nature of the underlying datastructure, it is very cheap to create these --
// effectively a pointer copy.
func (me ITable_) Table() ITable {
	if me.m == nil {
		return ITable{}
	}
	return ITable{
		me.m.trie,
		me.eq,
	}
}

// ITable is a structure that maps IP prefixes to values. For example, the
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
// The zero value of a ITable is an empty table
// ITable is immutable. For a mutable equivalent, see ITable_.
type ITable struct {
	trie *trieNode
	eq   comparator
}

// Table_ returns a mutable table initialized with the contents of this one. Due to
// the COW nature of the underlying datastructure, it is very cheap to copy
// these -- effectively a pointer copy.
func (me ITable) Table_() ITable_ {
	return ITable_{
		m: &ITable{
			me.trie,
			me.eq,
		},
	}
}

// NumEntries returns the number of exact prefixes stored in the table
func (me ITable) NumEntries() int64 {
	return me.trie.NumNodes()
}

// Get returns the value in the table associated with the given network prefix
// with an exact match: both the IP and the prefix length must match. If an
// exact match is not found, found is false and value is nil and should be
// ignored.
func (me ITable) Get(prefix PrefixI) (interface{}, bool) {
	value, matched, _ := me.longestMatch(prefix)

	if matched == matchExact {
		return value, true
	}

	return nil, false
}

// LongestMatch returns the value associated with the given network prefix
// using a longest prefix match. If a match is found, it returns true and the
// Prefix matched, which may be equal to or shorter than the one passed. If no
// match is found, returns false and the other fields must be ignored.
func (me ITable) LongestMatch(prefix PrefixI) (value interface{}, found bool, matchPrefix Prefix) {
	var matched match
	value, matched, matchPrefix = me.longestMatch(prefix)
	if matched != matchNone {
		return value, true, matchPrefix
	}
	return nil, false, Prefix{}
}

func (me ITable) longestMatch(prefix PrefixI) (value interface{}, matched match, matchPrefix Prefix) {
	if prefix == nil {
		prefix = Prefix{}
	}
	sp := prefix.Prefix()
	var node *trieNode
	node = me.trie.Match(sp)
	if node == nil {
		return nil, matchNone, Prefix{}
	}

	var resultKey Prefix
	resultKey = node.Prefix

	if node.Prefix.length == sp.length {
		return node.Data, matchExact, resultKey
	}
	return node.Data, matchContains, resultKey
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
// a custom compare can be used in aggregation.
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
func (me ITable) Aggregate() ITable {
	return ITable{
		me.trie.Aggregate(me.eq),
		me.eq,
	}
}

// Walk invokes the given callback function for each prefix/value pair in
// the table in lexigraphical order.
//
// It returns false if iteration was stopped due to a callback returning false
// or true if it iterated all items.
func (me ITable) Walk(callback func(Prefix, interface{}) bool) bool {
	return me.trie.Walk(callback)
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
func (me ITable) Diff(other ITable, left, right func(Prefix, interface{}) bool, changed func(p Prefix, left, right interface{}) bool) bool {
	trieHandler := trieDiffHandler{}
	if left != nil {
		trieHandler.Removed = func(n *trieNode) bool {
			return left(n.Prefix, n.Data)
		}
	}
	if right != nil {
		trieHandler.Added = func(n *trieNode) bool {
			return right(n.Prefix, n.Data)
		}
	}
	if changed != nil {
		trieHandler.Modified = func(l, r *trieNode) bool {
			return changed(l.Prefix, l.Data, r.Data)
		}
	}
	return me.trie.Diff(other.trie, trieHandler, me.eq)
}

// Map invokes the given mapper function for each prefix/value pair in the
// table in lexigraphical order. The resulting table has the same Prefix
// entries as the original but the values are modified by the mapper for each.
//
// A similar result can be obtained by calling Walk on the table, mapping each
// result, and inserting it into a new table or updating a mutable clone of the
// original. However, Map is more efficient than that.
//
// The walk method is inefficient in the following ways.
// 1. If inserting into a new map, a new entry is created even if the values
//    compare equal.
// 2. Each step in the walk produces an intermediate result that is eventually
//    thrown away (except the final result).
// 3. Each insert or update must traverse the result map.
//
// Map avoids all of these inefficiencies by building the resulting table in
// place takking time that is linear in the number of entries. It also avoids
// modifying anything if any values compare equal to the original.
func (me ITable) Map(mapper func(Prefix, interface{}) interface{}) ITable {
	if mapper == nil {
		return me
	}
	return ITable{
		me.trie.Map(mapper, me.eq),
		me.eq,
	}
}
