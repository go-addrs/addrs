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
	// This is an abuse of ImmutableMap because it uses its package privileges
	// to turn it into a mutable one. This could be refactored to be cleaner
	// without changing the interface.

	// Be careful not to take an ImmutableMap from outside the package and turn
	// it into a mutable one. That would break the contract.
	m *ImmutableMap
}

// NewMap returns a new fully-initialized Map
func NewMap() Map {
	return Map{
		m: &ImmutableMap{},
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

// InsertPrefix inserts the given prefix with the given value into the map
func (me Map) InsertPrefix(prefix Prefix, value interface{}) error {
	var err error
	var newHead *trieNode
	newHead, err = me.m.trie.Insert(prefix, value)
	if err != nil {
		return err
	}

	me.m.trie = newHead
	return nil
}

// Insert is a convenient alternative to InsertPrefix that treats the given IP
// address as a host prefix (i.e. /32 for IPv4 and /128 for IPv6)
func (me Map) Insert(ip Address, value interface{}) error {
	return me.InsertPrefix(ip.Prefix(), value)
}

// UpdatePrefix inserts the given prefix with the given value into the map.
// If the prefix already existed, it updates the associated value in place.
func (me Map) UpdatePrefix(prefix Prefix, value interface{}) error {
	var err error
	var newHead *trieNode
	newHead, err = me.m.trie.Update(prefix, value)
	if err != nil {
		return err
	}

	me.m.trie = newHead
	return nil
}

// Update is a convenient alternative to UpdatePrefix that treats
// the given IP address as a host prefix (i.e. /32 for IPv4 and /128 for IPv6)
func (me Map) Update(ip Address, value interface{}) error {
	return me.UpdatePrefix(ip.Prefix(), value)
}

// InsertOrUpdatePrefix inserts the given prefix with the given value into the map.
// If the prefix already existed, it updates the associated value in place.
func (me Map) InsertOrUpdatePrefix(prefix Prefix, value interface{}) {
	me.m.trie = me.m.trie.InsertOrUpdate(prefix, value)
}

// InsertOrUpdate is a convenient alternative to InsertOrUpdatePrefix that treats
// the given IP address as a host prefix (i.e. /32 for IPv4 and /128 for IPv6)
func (me Map) InsertOrUpdate(ip Address, value interface{}) {
	me.InsertOrUpdatePrefix(ip.Prefix(), value)
}

// GetPrefix returns the value in the map associated with the given network prefix
// with an exact match: both the IP and the prefix length must match. If an
// exact match is not found, found is false and value is nil and should be
// ignored.
func (me Map) GetPrefix(prefix Prefix) (interface{}, bool) {
	return me.m.GetPrefix(prefix)
}

// Get is a convenient alternative to GetPrefix that treats the given IP address
// as a host prefix (i.e. /32 for IPv4 and /128 for IPv6)
func (me Map) Get(ip Address) (value interface{}, found bool) {
	return me.m.GetPrefix(ip.Prefix())
}

// GetOrInsertPrefix returns the value associated with the given prefix if it
// already exists. If it does not exist, it inserts it with the given value and
// returns that.
func (me Map) GetOrInsertPrefix(prefix Prefix, value interface{}) interface{} {
	var newHead, node *trieNode
	newHead, node = me.m.trie.GetOrInsert(prefix, value)
	me.m.trie = newHead
	return node.Data
}

// GetOrInsert is a convenient alternative to GetOrInsertPrefix that treats the
// given IP address as a host prefix (i.e. /32 for IPv4 and /128 for IPv6)
func (me Map) GetOrInsert(ip Address, value interface{}) interface{} {
	return me.GetOrInsertPrefix(ip.Prefix(), value)
}

// LongestMatchPrefix returns the value in the map associated with the given
// network prefix using a longest prefix match. If a match is found, it returns
// a Prefix representing the longest prefix matched. If a match is *not* found,
// matched is MatchNone and the other fields should be ignored
func (me Map) LongestMatchPrefix(searchPrefix Prefix) (matched Match, prefix Prefix, value interface{}) {
	return me.m.LongestMatchPrefix(searchPrefix)
}

// LongestMatch is a convenient alternative to MatchPrefix that treats the
// given IP address as a host prefix (i.e. /32 for IPv4 and /128 for IPv6)
func (me Map) LongestMatch(ip Address) (matched Match, prefix Prefix, value interface{}) {
	return me.m.LongestMatchPrefix(ip.Prefix())
}

// RemovePrefix removes the given prefix from the map with its associated value.
// Only a prefix with an exact match will be removed.
func (me Map) RemovePrefix(prefix Prefix) (err error) {
	me.m.trie, err = me.m.trie.Delete(prefix)
	return
}

// Remove is a convenient alternative to RemovePrefix that treats the given IP
// address as a host prefix (i.e. /32 for IPv4 and /128 for IPv6)
func (me Map) Remove(ip Address) error {
	return me.RemovePrefix(ip.Prefix())
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

// Immutable returns an immutable snapshot of this Map. Due to the COW nature
// of the underlying datastructure, it is very cheap to create these --
// effectively a pointer copy.
func (me Map) Immutable() ImmutableMap {
	return ImmutableMap{
		trie: me.m.trie,
	}
}
