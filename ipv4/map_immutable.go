package ipv4

// ImmutableMap is like a Map except without the ability to modify it.
type ImmutableMap struct {
	trie *trieNode
}

// Map returns a mutable map initialized with the contents of this one. Due to
// the COW nature of the underlying datastructure, it is very cheap to copy
// these -- effectively a pointer copy.
func (me ImmutableMap) Map() Map {
	return Map{
		m: &ImmutableMap{
			trie: me.trie,
		},
	}
}

// Size returns the number of exact prefixes stored in the map
func (me ImmutableMap) Size() int64 {
	return me.trie.NumNodes()
}

// GetPrefix returns the value in the map associated with the given network prefix
// with an exact match: both the IP and the prefix length must match. If an
// exact match is not found, found is false and value is nil and should be
// ignored.
func (me ImmutableMap) GetPrefix(prefix Prefix) (interface{}, bool) {
	match, _, value := me.LongestMatchPrefix(prefix)

	if match == MatchExact {
		return value, true
	}

	return nil, false
}

// Get is a convenient alternative to GetPrefix that treats the given IP address
// as a host prefix (i.e. /32 for IPv4 and /128 for IPv6)
func (me ImmutableMap) Get(ip Address) (value interface{}, found bool) {
	return me.GetPrefix(ip.HostPrefix())
}

// LongestMatchPrefix returns the value in the map associated with the given
// network prefix using a longest prefix match. If a match is found, it returns
// a Prefix representing the longest prefix matched. If a match is *not* found,
// matched is MatchNone and the other fields should be ignored
func (me ImmutableMap) LongestMatchPrefix(searchPrefix Prefix) (matched Match, prefix Prefix, value interface{}) {
	var node *trieNode
	node = me.trie.Match(searchPrefix)
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
func (me ImmutableMap) LongestMatch(ip Address) (matched Match, prefix Prefix, value interface{}) {
	return me.LongestMatchPrefix(ip.HostPrefix())
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
func (me ImmutableMap) Iterate(callback MapCallback) bool {
	return me.trie.Iterate(trieCallback(callback))
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
func (me ImmutableMap) IterateAggregates(callback MapCallback) bool {
	return me.trie.Aggregate(trieCallback(callback))
}
