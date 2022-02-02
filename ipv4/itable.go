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
		panic("cannot modify an unitialized Set")
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
		panic("cannot modify an unitialized Set")
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
		panic("cannot modify an unitialized Set")
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
		panic("cannot modify an unitialized Set")
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
		panic("cannot modify an unitialized Set")
	}
	var err error
	me.mutate(func() (bool, *trieNode) {
		var newHead *trieNode
		newHead, err = me.m.trie.Delete(prefix.Prefix())
		return true, newHead
	})
	return err == nil
}

// Walk invokes the given callback function for each prefix/value pair in
// the table in lexigraphical order.
//
// It returns false if iteration was stopped due to a callback return false or
// true if it iterated all items.
func (me ITable) Walk(callback ITableCallback) bool {
	if me.m == nil {
		return true
	}
	return me.m.Walk(callback)
}

// WalkAggregates invokes then given callback function for each prefix/value
// pair in the table, aggregated by value, in lexigraphical order.
//
// If two prefixes map to the same value, one contains the other, and there is
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
func (me ITable) WalkAggregates(callback ITableCallback) bool {
	if me.m == nil {
		return true
	}
	return me.m.WalkAggregates(callback)
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
