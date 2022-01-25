package ipv4

type sharedMap struct {
	trie *trieNode
}

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
	m *sharedMap
}

// NewMap returns a new fully-initialized Map
func NewMap() Map {
	return Map{
		m: &sharedMap{},
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
	return me.m.trie.NumNodes()
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
	return me.InsertPrefix(ip.HostPrefix(), value)
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
	return me.UpdatePrefix(ip.HostPrefix(), value)
}

// InsertOrUpdatePrefix inserts the given prefix with the given value into the map.
// If the prefix already existed, it updates the associated value in place.
func (me Map) InsertOrUpdatePrefix(prefix Prefix, value interface{}) {
	me.m.trie = me.m.trie.InsertOrUpdate(prefix, value)
}

// InsertOrUpdate is a convenient alternative to InsertOrUpdatePrefix that treats
// the given IP address as a host prefix (i.e. /32 for IPv4 and /128 for IPv6)
func (me Map) InsertOrUpdate(ip Address, value interface{}) {
	me.InsertOrUpdatePrefix(ip.HostPrefix(), value)
}

// GetPrefix returns the value in the map associated with the given network prefix
// with an exact match: both the IP and the prefix length must match. If an
// exact match is not found, found is false and value is nil and should be
// ignored.
func (me Map) GetPrefix(prefix Prefix) (interface{}, bool) {
	match, _, value := me.LongestMatchPrefix(prefix)

	if match == MatchExact {
		return value, true
	}

	return nil, false
}

// Get is a convenient alternative to GetPrefix that treats the given IP address
// as a host prefix (i.e. /32 for IPv4 and /128 for IPv6)
func (me Map) Get(ip Address) (value interface{}, found bool) {
	return me.GetPrefix(ip.HostPrefix())
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
	return me.GetOrInsertPrefix(ip.HostPrefix(), value)
}

// LongestMatchPrefix returns the value in the map associated with the given
// network prefix using a longest prefix match. If a match is found, it returns
// a Prefix representing the longest prefix matched. If a match is *not* found,
// matched is MatchNone and the other fields should be ignored
func (me Map) LongestMatchPrefix(searchPrefix Prefix) (matched Match, prefix Prefix, value interface{}) {
	var node *trieNode
	node = me.m.trie.Match(searchPrefix)
	if node == nil {
		return MatchNone, Prefix{}, nil
	}

	var resultKey Prefix
	resultKey = node.Prefix

	if node.Prefix.length == searchPrefix.length {
		return MatchExact, resultKey, node.Data
	}
	return MatchContains, resultKey, node.Data
}

// LongestMatch is a convenient alternative to MatchPrefix that treats the
// given IP address as a host prefix (i.e. /32 for IPv4 and /128 for IPv6)
func (me Map) LongestMatch(ip Address) (matched Match, prefix Prefix, value interface{}) {
	return me.LongestMatchPrefix(ip.HostPrefix())
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
	return me.RemovePrefix(ip.HostPrefix())
}

// MapCallback is the signature of the callback functions that can be passed to
// Iterate or IterateAggregates to handle each prefix/value combination.
//
// Each invocation of your callback should return true if iteration should
// continue (as long as another key / value pair exists) or false to stop
// iterating and return immediately (meaning your callback will not be called
// again).
type MapCallback func(Prefix, interface{}) bool

// Iterate invokes the given callback function for each prefix/value pair in
// the map in lexigraphical order.
//
// It returns false if iteration was stopped due to a callback return false or
// true if it iterated all items.
func (me Map) Iterate(callback MapCallback) bool {
	return me.m.trie.Iterate(trieCallback(callback))
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
	return me.m.trie.Aggregate(trieCallback(callback))
}
