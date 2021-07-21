package ipv4

// IPMap is a structure that maps IP prefixes to values. For example, you can
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
type IPMap struct {
	trie trie32
}

// Size returns the number of exact prefixes stored in the map
func (m *IPMap) Size() int {
	return m.trie.Size()
}

// InsertPrefix inserts the given prefix with the given value into the map
func (m *IPMap) InsertPrefix(prefix Prefix, value interface{}) error {
	return m.trie.Insert(prefix, value)
}

// Insert is a convenient alternative to InsertPrefix that treats the given IP
// address as a host prefix (i.e. /32 for IPv4 and /128 for IPv6)
func (m *IPMap) Insert(ip Addr, value interface{}) error {
	return m.trie.Insert(
		ipToKey(ip),
		value,
	)
}

// InsertOrUpdatePrefix inserts the given prefix with the given value into the map.
// If the prefix already existed, it updates the associated value in place.
func (m *IPMap) InsertOrUpdatePrefix(prefix Prefix, value interface{}) error {
	return m.trie.InsertOrUpdate(prefix, value)
}

// InsertOrUpdate is a convenient alternative to InsertOrUpdatePrefix that treats
// the given IP address as a host prefix (i.e. /32 for IPv4 and /128 for IPv6)
func (m *IPMap) InsertOrUpdate(ip Addr, value interface{}) error {
	return m.trie.InsertOrUpdate(
		ipToKey(ip),
		value,
	)
}

// GetPrefix returns the value in the map associated with the given network prefix
// with an exact match: both the IP and the prefix length must match. If an
// exact match is not found, found is false and value is nil and should be
// ignored.
func (m *IPMap) GetPrefix(prefix Prefix) (interface{}, bool) {
	match, _, value := m.trie.Match(prefix)

	if match == matchExact {
		return value, true
	}

	return nil, false
}

// Get is a convenient alternative to GetPrefix that treats the given IP address
// as a host prefix (i.e. /32 for IPv4 and /128 for IPv6)
func (m *IPMap) Get(ip Addr) (interface{}, bool) {
	key := ipToKey(ip)
	match, _, value := m.trie.Match(key)

	if match == matchExact {
		return value, true
	}

	return nil, false
}

// GetOrInsertPrefix returns the value associated with the given prefix if it
// already exists. If it does not exist, it inserts it with the given value and
// returns that.
func (m *IPMap) GetOrInsertPrefix(prefix Prefix, value interface{}) (interface{}, error) {
	return m.trie.GetOrInsert(prefix, value)
}

// GetOrInsert is a convenient alternative to GetOrInsertPrefix that treats the
// given IP address as a host prefix (i.e. /32 for IPv4 and /128 for IPv6)
func (m *IPMap) GetOrInsert(ip Addr, value interface{}) (interface{}, error) {
	key := ipToKey(ip)
	return m.trie.GetOrInsert(key, value)
}

// MatchPrefix returns the value in the map associated with the given network
// prefix using a longest prefix match. If a match is found, it returns a
// Prefix representing the longest prefix matched. If a match is *not*
// found, ok is false and the other fields should be ignored
func (m *IPMap) MatchPrefix(searchPrefix Prefix) (ok bool, prefix Prefix, value interface{}) {
	match, matchKey, value := m.trie.Match(searchPrefix)

	if match == matchNone {
		return false, Prefix{}, nil
	}

	return true, matchKey, value
}

// Match is a convenient alternative to MatchPrefix that treats the given IP
// address as a host prefix (i.e. /32 for IPv4 and /128 for IPv6)
func (m *IPMap) Match(ip Addr) (ok bool, prefix Prefix, value interface{}) {
	key := ipToKey(ip)
	match, matchKey, value := m.trie.Match(key)

	if match == matchNone {
		return false, Prefix{}, nil
	}

	return true, matchKey, value
}

// RemovePrefix removes the given prefix from the map with its associated value.
// Only a prefix with an exact match will be removed.
func (m *IPMap) RemovePrefix(prefix Prefix) {
	m.trie.Delete(prefix)
}

// Remove is a convenient alternative to RemovePrefix that treats the given IP
// address as a host prefix (i.e. /32 for IPv4 and /128 for IPv6)
func (m *IPMap) Remove(ip Addr) {
	m.trie.Delete(ipToKey(ip))
}

// MapCallback is the signature of the callback functions that can be passed to
// Iterate or Aggregate to handle each prefix/value combination.
type MapCallback trie32Callback

// Iterate invokes the given callback function for each prefix/value pair in
// the map in lexigraphical order.
func (m *IPMap) Iterate(callback MapCallback) bool {
	return m.trie.Iterate(trie32Callback(callback))
}

// Aggregate invokes then given callback function for each prefix/value pair in
// the map, aggregated by value, in lexigraphical order.
//
// 1. The values stored must be comparable to be aggregable. Prefixes get
//    aggregated only where their values compare equal.
// 2. The set of prefix/value pairs visited is the minimal set such that any
//    longest prefix match against the aggregated set will always return the
//    same value as the same match against the non-aggregated set.
// 3. The aggregated and non-aggregated sets of prefixes may be disjoint.
func (m *IPMap) Aggregate(callback MapCallback) bool {
	return m.trie.Aggregate(trie32Callback(callback))
}

func ipToKey(ip Addr) Prefix {
	return Prefix{
		Addr:   ip,
		length: uint32(SIZE),
	}
}
