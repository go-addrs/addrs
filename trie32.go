package ipv4

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

type trie32 struct {
	top *trieNode32
}

// Size returns the number of entries
func (me *trie32) Size() int {
	return me.top.Size()
}

// Insert adds the given key / value pair. If the new key cannot be inserted or
// already exists, an error is returned.
func (me *trie32) Insert(key Prefix, value interface{}) error {
	var err error
	var newHead *trieNode32
	newHead, err = me.top.Insert(key, value)
	if err != nil {
		return err
	}

	me.top = newHead
	return nil
}

// InsertOrUpdate adds the given key / value pair. If the new key cannot be
// inserted or already exists, an error is returned.
func (me *trie32) InsertOrUpdate(key Prefix, value interface{}) error {
	var err error
	var newHead *trieNode32
	newHead, err = me.top.InsertOrUpdate(key, value)
	if err != nil {
		return err
	}

	me.top = newHead
	return nil
}

// Update adds the given key / value pair. If the new key cannot be inserted or
// already exists, an error is returned.
func (me *trie32) Update(key Prefix, value interface{}) error {
	var err error
	var newHead *trieNode32
	newHead, err = me.top.Update(key, value)
	if err != nil {
		return err
	}

	me.top = newHead
	return nil
}

// GetOrInsert returns the value for the given key. If the key is not found,
// then value is inserted and returned. If the new key cannot be inserted, an
// error is returned.
func (me *trie32) GetOrInsert(key Prefix, value interface{}) (interface{}, error) {
	var newHead, node *trieNode32
	newHead, node = me.top.GetOrInsert(key, value)
	me.top = newHead
	return node.Data, nil
}

// Match returns the existing key / value pair with the longest prefix that
// fully contains the given key or nil if none match.
//
// "contains" means that the first "Length" bits in the entry's key are exactly
// the same as the same number of first bits in the given search key. This
// implies the search key is at least as long as any matching node's prefix.
//
// Some examples include the following ipv4 and ipv6 matches:
//     10.0.0.0/24 contains 10.0.0.0/24, 10.0.0.0/25, and 10.0.0.0/32
//     2001:cafe:beef::/64 contains 2001:cafe:beef::a/124
//
// "longest" means that if multiple existing entries in the trie match the one
// with the longest length will be returned. It is the most specific match.
func (me *trie32) Match(key Prefix) (Match, Prefix, interface{}) {
	var node *trieNode32
	node = me.top.Match(key)
	if node == nil {
		return MatchNone, Prefix{}, nil
	}

	var resultKey Prefix
	resultKey = node.Prefix

	if node.Prefix.length == key.length {
		return MatchExact, resultKey, node.Data
	}
	return MatchContains, resultKey, node.Data
}

// Delete removes a key from the trie with its associated value.
func (me *trie32) Delete(key Prefix) error {
	var err error
	me.top, err = me.top.Delete(key)
	return err
}

// Iterate walks the entire trie and calls the given function for each key /
// value pair. The order of visiting nodes is essentially lexigraphical:
// - disjoint prefixes are visited in lexigraphical order
// - shorter prefixes are visited immediately before longer prefixes that they contain
func (me *trie32) Iterate(callback trie32Callback) bool {
	return me.top.Iterate(func(key Prefix, value interface{}) bool {
		return callback((Prefix)(key), value)
	})
}

// Aggregate is like iterate except that it has the capability of aggregating
// prefixes that are either adjacent to each other with the same prefix length
// or contained within another prefix with a shorter length.

// Aggregation visits the minimum set of key/value pairs needed to return the
// same value for any longest prefix match as would be returned by the the
// original trie, non-aggregated. This can be useful, for example, to minimize
// the number of prefixes needed to install into a router's datapath to
// guarantee that all of the next hops are correct.
//
// In general, routing protocols should not aggregate and then pass on the
// aggregates to neighbors as this will likely lead to poor comparisions by
// neighboring routers who receive routes aggregated differently from different
// peers.
//
// Prefixes are only considered aggregable if their value compare equal. This is
// useful for aggregating prefixes where the next hop is the same but not where
// they're different.
func (me *trie32) Aggregate(callback trie32Callback) bool {
	return me.top.Aggregate(func(key Prefix, value interface{}) bool {
		return callback((Prefix)(key), value)
	})
}
