package ipv4

// Map is a structure that maps IP prefixes to values. For example, you can
// insert the following values and they will all exist as distinct prefix/value
// pairs in the map.
//
// 10.0.0.0/16 -> 1
// 10.0.0.0/24 -> 1
// 10.0.0.0/32 -> 2
//
// The map supports looking up values based on a longest prefix match and also
// supports efficient aggregation of prefix/value pairs based on equality of
// values. See the README.md file for a more detailed discussion..
type Map struct {
	// This is an abuse of FixedMap because it uses its package privileges
	// to turn it into a mutable one. This could be refactored to be cleaner
	// without changing the interface.

	// Be careful not to take an FixedMap from outside the package and turn
	// it into a mutable one. That would break the contract.
	m *FixedMap
}

// NewMap returns a new fully-initialized Map
func NewMap() Map {
	return Map{
		m: &FixedMap{},
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

// Size returns the number of exact prefixes stored in the map
func (me Map) Size() int64 {
	return me.m.Size()
}

// Insert inserts the given prefix with the given value into the map
func (me Map) Insert(p PrefixI, value interface{}) error {
	var err error
	var newHead *trieNode
	newHead, err = me.m.trie.Insert(p.Prefix(), value)
	if err != nil {
		return err
	}

	me.m.trie = newHead
	return nil
}

// Update inserts the given prefix with the given value into the map. If the
// prefix already existed, it updates the associated value in place.
func (me Map) Update(p PrefixI, value interface{}) error {
	var err error
	var newHead *trieNode
	newHead, err = me.m.trie.Update(p.Prefix(), value)
	if err != nil {
		return err
	}

	me.m.trie = newHead
	return nil
}

// InsertOrUpdate inserts the given prefix with the given value into the map.
// If the prefix already existed, it updates the associated value in place.
func (me Map) InsertOrUpdate(p PrefixI, value interface{}) {
	me.m.trie = me.m.trie.InsertOrUpdate(p.Prefix(), value)
}

// Get returns the value in the map associated with the given network prefix
// with an exact match: both the IP and the prefix length must match. If an
// exact match is not found, found is false and value is nil and should be
// ignored.
func (me Map) Get(prefix PrefixI) (interface{}, bool) {
	return me.m.Get(prefix)
}

// GetOrInsert returns the value associated with the given prefix if it already
// exists. If it does not exist, it inserts it with the given value and returns
// that.
func (me Map) GetOrInsert(p PrefixI, value interface{}) interface{} {
	var newHead, node *trieNode
	newHead, node = me.m.trie.GetOrInsert(p.Prefix(), value)
	me.m.trie = newHead
	return node.Data
}

// LongestMatch returns the value in the map associated with the given network
// prefix using a longest prefix match. If a match is found, it returns a
// Prefix representing the longest prefix matched. If a match is *not* found,
// matched is MatchNone and the other fields should be ignored
func (me Map) LongestMatch(searchPrefix PrefixI) (matched Match, prefix Prefix, value interface{}) {
	return me.m.LongestMatch(searchPrefix)
}

// Remove removes the given prefix from the map with its associated value. Only
// a prefix with an exact match will be removed.
func (me Map) Remove(p PrefixI) (err error) {
	me.m.trie, err = me.m.trie.Delete(p.Prefix())
	return
}

// Iterate invokes the given callback function for each prefix/value pair in
// the map in lexigraphical order.
//
// It returns false if iteration was stopped due to a callback return false or
// true if it iterated all items.
func (me Map) Iterate(callback MapCallback) bool {
	return me.m.Iterate(callback)
}

// IterateAggregates invokes then given callback function for each prefix/value
// pair in the map, aggregated by value, in lexigraphical order.
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
func (me Map) IterateAggregates(callback MapCallback) bool {
	return me.m.IterateAggregates(callback)
}

// FixedMap returns an immutable snapshot of this Map. Due to the COW
// nature of the underlying datastructure, it is very cheap to create these --
// effectively a pointer copy.
func (me Map) FixedMap() FixedMap {
	return FixedMap{
		trie: me.m.trie,
	}
}
