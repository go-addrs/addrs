package ipv4

import (
	"fmt"
)

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
	trie *trieNode
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
func (m *Map) Size() int64 {
	return m.trie.NumNodes()
}

// InsertPrefix inserts the given prefix with the given value into the map
func (m *Map) InsertPrefix(prefix Prefix, value interface{}) error {
	var err error
	var newHead *trieNode
	newHead, err = m.trie.Insert(prefix, value)
	if err != nil {
		return err
	}

	m.trie = newHead
	return nil
}

// Insert is a convenient alternative to InsertPrefix that treats the given IP
// address as a host prefix (i.e. /32 for IPv4 and /128 for IPv6)
func (m *Map) Insert(ip Address, value interface{}) error {
	return m.InsertPrefix(ipToKey(ip), value)
}

// UpdatePrefix inserts the given prefix with the given value into the map.
// If the prefix already existed, it updates the associated value in place.
func (m *Map) UpdatePrefix(prefix Prefix, value interface{}) error {
	var err error
	var newHead *trieNode
	newHead, err = m.trie.Update(prefix, value)
	if err != nil {
		return err
	}

	m.trie = newHead
	return nil
}

// Update is a convenient alternative to UpdatePrefix that treats
// the given IP address as a host prefix (i.e. /32 for IPv4 and /128 for IPv6)
func (m *Map) Update(ip Address, value interface{}) error {
	return m.UpdatePrefix(ipToKey(ip), value)
}

// InsertOrUpdatePrefix inserts the given prefix with the given value into the map.
// If the prefix already existed, it updates the associated value in place.
func (m *Map) InsertOrUpdatePrefix(prefix Prefix, value interface{}) {
	var err error
	var newHead *trieNode
	newHead, err = m.trie.InsertOrUpdate(prefix, value)
	if err != nil {
		panic(fmt.Errorf("this error shouldn't happen: %w", err))
	}

	m.trie = newHead
}

// InsertOrUpdate is a convenient alternative to InsertOrUpdatePrefix that treats
// the given IP address as a host prefix (i.e. /32 for IPv4 and /128 for IPv6)
func (m *Map) InsertOrUpdate(ip Address, value interface{}) {
	m.InsertOrUpdatePrefix(ipToKey(ip), value)
}

// GetPrefix returns the value in the map associated with the given network prefix
// with an exact match: both the IP and the prefix length must match. If an
// exact match is not found, found is false and value is nil and should be
// ignored.
func (m *Map) GetPrefix(prefix Prefix) (interface{}, bool) {
	match, _, value := m.LongestMatchPrefix(prefix)

	if match == MatchExact {
		return value, true
	}

	return nil, false
}

// Get is a convenient alternative to GetPrefix that treats the given IP address
// as a host prefix (i.e. /32 for IPv4 and /128 for IPv6)
func (m *Map) Get(ip Address) (value interface{}, found bool) {
	return m.GetPrefix(ipToKey(ip))
}

// GetOrInsertPrefix returns the value associated with the given prefix if it
// already exists. If it does not exist, it inserts it with the given value and
// returns that.
func (m *Map) GetOrInsertPrefix(prefix Prefix, value interface{}) interface{} {
	var newHead, node *trieNode
	newHead, node = m.trie.GetOrInsert(prefix, value)
	m.trie = newHead
	return node.Data
}

// GetOrInsert is a convenient alternative to GetOrInsertPrefix that treats the
// given IP address as a host prefix (i.e. /32 for IPv4 and /128 for IPv6)
func (m *Map) GetOrInsert(ip Address, value interface{}) interface{} {
	return m.GetOrInsertPrefix(ipToKey(ip), value)
}

// LongestMatchPrefix returns the value in the map associated with the given
// network prefix using a longest prefix match. If a match is found, it returns
// a Prefix representing the longest prefix matched. If a match is *not* found,
// matched is MatchNone and the other fields should be ignored
func (m *Map) LongestMatchPrefix(searchPrefix Prefix) (matched Match, prefix Prefix, value interface{}) {
	var node *trieNode
	node = m.trie.Match(searchPrefix)
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
func (m *Map) LongestMatch(ip Address) (matched Match, prefix Prefix, value interface{}) {
	return m.LongestMatchPrefix(ipToKey(ip))
}

// RemovePrefix removes the given prefix from the map with its associated value.
// Only a prefix with an exact match will be removed.
func (m *Map) RemovePrefix(prefix Prefix) (err error) {
	m.trie, err = m.trie.Delete(prefix)
	return
}

// Remove is a convenient alternative to RemovePrefix that treats the given IP
// address as a host prefix (i.e. /32 for IPv4 and /128 for IPv6)
func (m *Map) Remove(ip Address) error {
	return m.RemovePrefix(ipToKey(ip))
}

// MapCallback is the signature of the callback functions that can be passed to
// Iterate or IterateAggregates to handle each prefix/value combination.
type MapCallback func(Prefix, interface{}) bool

// Iterate invokes the given callback function for each prefix/value pair in
// the map in lexigraphical order.
func (m *Map) Iterate(callback MapCallback) bool {
	return m.trie.Iterate(trieCallback(callback))
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
func (m *Map) IterateAggregates(callback MapCallback) bool {
	return m.trie.Aggregate(trieCallback(callback))
}

func ipToKey(ip Address) Prefix {
	return Prefix{
		Address: ip,
		length:  uint32(SIZE),
	}
}
