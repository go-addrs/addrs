package ipv4

// ITable is a structure that maps IP prefixes to values. For example, you can
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
// The zero value of a ITable is unitialized. Reading it is equivalent to
// reading an empty ITable. Attempts to modify it will result in a panic. Always
// use NewITable() to get a modifyable ITable.
type ITable struct {
	// This is an abuse of FixedITable because it uses its package privileges
	// to turn it into a mutable one. This could be refactored to be cleaner
	// without changing the interface.

	// Be careful not to take an FixedITable from outside the package and turn
	// it into a mutable one. That would break the contract.
	m *FixedITable
}

// NewITable returns a new fully-initialized ITable
func NewITable() ITable {
	return ITable{
		m: &FixedITable{},
	}
}

// Match indicates how closely the given key matches the search result
type Match int

const (
	// MatchNone indicates that no match was found
	MatchNone Match = iota
	// MatchContains indicates that a match was found that contains the search key but isn't exact
	MatchContains
	// MatchExact indicates that a match with the same prefix
	MatchExact
)

// Size returns the number of exact prefixes stored in the table
func (me ITable) Size() int64 {
	if me.m == nil {
		return 0
	}
	return me.m.Size()
}

// mutate should be called by any method that modifies the table in any way
func (me ITable) mutate(mutator func() (ok bool, node *trieNode)) {
	oldNode := me.m.trie
	ok, newNode := mutator()
	if ok && oldNode != newNode {
		if !swapTrieNodePtr(&me.m.trie, oldNode, newNode) {
			panic("concurrent modification of Table detected")
		}
	}
}

// Insert inserts the given prefix with the given value into the table.
// If an entry with the same prefix already exists, it will not overwrite it
// and return false.
func (me ITable) Insert(prefix PrefixI, value interface{}) (succeeded bool) {
	if me.m == nil {
		panic("cannot modify an unitialized Table")
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
func (me ITable) Update(prefix PrefixI, value interface{}) (succeeded bool) {
	if me.m == nil {
		panic("cannot modify an unitialized Table")
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
func (me ITable) InsertOrUpdate(prefix PrefixI, value interface{}) {
	if me.m == nil {
		panic("cannot modify an unitialized Table")
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
func (me ITable) Get(prefix PrefixI) (interface{}, bool) {
	if me.m == nil {
		return nil, false
	}
	return me.m.Get(prefix)
}

// GetOrInsert returns the value associated with the given prefix if it already
// exists. If it does not exist, it inserts it with the given value and returns
// that.
func (me ITable) GetOrInsert(prefix PrefixI, value interface{}) interface{} {
	if me.m == nil {
		panic("cannot modify an unitialized Table")
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

// LongestMatch returns the value in the table associated with the given network
// prefix using a longest prefix match. If a match is found, it returns a
// Prefix representing the longest prefix matched. If a match is *not* found,
// matched is MatchNone and the other fields should be ignored
func (me ITable) LongestMatch(prefix PrefixI) (value interface{}, matched Match, matchPrefix Prefix) {
	if me.m == nil {
		return nil, MatchNone, Prefix{}
	}
	return me.m.LongestMatch(prefix)
}

// Remove removes the given prefix from the table with its associated value and
// returns true if it was found. Only a prefix with an exact match will be
// removed. If no entry with the given prefix exists, it will do nothing and
// return false.
func (me ITable) Remove(prefix PrefixI) (succeeded bool) {
	if me.m == nil {
		panic("cannot modify an unitialized Table")
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

// FixedTable returns an immutable snapshot of this ITable. Due to the COW
// nature of the underlying datastructure, it is very cheap to create these --
// effectively a pointer copy.
func (me ITable) FixedTable() FixedITable {
	if me.m == nil {
		return FixedITable{}
	}
	return FixedITable{
		trie: me.m.trie,
	}
}

// FixedITable is like a ITable except this its contents are frozen
// The zero value of a FixedITable is an empty table
type FixedITable struct {
	trie *trieNode
}

// Table returns a mutable table initialized with the contents of this one. Due to
// the COW nature of the underlying datastructure, it is very cheap to copy
// these -- effectively a pointer copy.
func (me FixedITable) Table() ITable {
	return ITable{
		m: &FixedITable{
			trie: me.trie,
		},
	}
}

// Size returns the number of exact prefixes stored in the table
func (me FixedITable) Size() int64 {
	return me.trie.NumNodes()
}

// Get returns the value in the table associated with the given network prefix
// with an exact match: both the IP and the prefix length must match. If an
// exact match is not found, found is false and value is nil and should be
// ignored.
func (me FixedITable) Get(prefix PrefixI) (interface{}, bool) {
	value, match, _ := me.LongestMatch(prefix)

	if match == MatchExact {
		return value, true
	}

	return nil, false
}

// LongestMatch returns the value in the table associated with the given network
// prefix using a longest prefix match. If a match is found, it returns a
// Prefix representing the longest prefix matched. If a match is *not* found,
// matched is MatchNone and the other fields should be ignored
func (me FixedITable) LongestMatch(prefix PrefixI) (value interface{}, matched Match, matchPrefix Prefix) {
	if prefix == nil {
		prefix = Prefix{}
	}
	sp := prefix.Prefix()
	var node *trieNode
	node = me.trie.Match(sp)
	if node == nil {
		return nil, MatchNone, Prefix{}
	}

	var resultKey Prefix
	resultKey = node.Prefix

	if node.Prefix.length == sp.length {
		return node.Data, MatchExact, resultKey
	}
	return node.Data, MatchContains, resultKey
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
func (me FixedITable) Aggregate() FixedITable {
	return FixedITable{
		trie: me.trie.Aggregate(),
	}
}

// ITableCallback is the signature of the callback functions that can be passed
// to Walk to handle each prefix/value combination.
//
// Each invocation of your callback should return true if iteration should
// continue (as long as another key / value pair exists) or false to stop
// iterating and return immediately (meaning your callback will not be called
// again).
type ITableCallback func(Prefix, interface{}) bool

// Walk invokes the given callback function for each prefix/value pair in
// the table in lexigraphical order.
//
// It returns false if iteration was stopped due to a callback returning false
// or true if it iterated all items.
func (me FixedITable) Walk(callback ITableCallback) bool {
	return me.trie.Walk(trieCallback(callback))
}

// ITableModifiedCallback is the signature of the callback functions to handle
// a modified entry when diffing. It is passed the prefix and the values before
// and after the change.
//
// Each invocation of your callback should return true if iteration should
// continue (as long as another key / value pair exists) or false to stop
// iterating and return immediately (meaning your callback will not be called
// again).
type ITableModifiedCallback func(p Prefix, left, right interface{}) bool

// IDiffHandler is a struct passed to Diff to handle changes found between the
// left and right tables. Removed is called for prefixes that appear in
// the left table but not the right, Added is called for prefixes that appear
// in the right but not the left, and Modified is called for prefixes that
// appear in both but have different values.
//
// Any of the handlers can be left out safely -- they will default to nil. In
// that case, Diff will skip those cases.
type IDiffHandler struct {
	Removed  ITableCallback
	Added    ITableCallback
	Modified ITableModifiedCallback
}

// Diff invokes the given callback functions for each prefix/value pair in the
// table in lexigraphical order.
//
// It returns false if iteration was stopped due to a callback returning false
// or true if it iterated all items.
func (me FixedITable) Diff(other FixedITable, handler IDiffHandler) bool {
	trieHandler := trieDiffHandler{}
	if handler.Removed != nil {
		trieHandler.Removed = func(n *trieNode) bool {
			return handler.Removed(n.Prefix, n.Data)
		}
	}
	if handler.Added != nil {
		trieHandler.Added = func(n *trieNode) bool {
			return handler.Added(n.Prefix, n.Data)
		}
	}
	if handler.Modified != nil {
		trieHandler.Modified = func(l, r *trieNode) bool {
			return handler.Modified(l.Prefix, l.Data, r.Data)
		}
	}
	return me.trie.Diff(other.trie, trieHandler)
}
