package ipv4

// FixedTableI is like a TableI except this its contents are frozen
type FixedTableI struct {
	trie *trieNode
}

// NewFixedTableI returns an initialized but empty FixedTableI
func NewFixedTableI() FixedTableI {
	return FixedTableI{}
}

// Table returns a mutable table initialized with the contents of this one. Due to
// the COW nature of the underlying datastructure, it is very cheap to copy
// these -- effectively a pointer copy.
func (me FixedTableI) Table() TableI {
	return TableI{
		m: &FixedTableI{
			trie: me.trie,
		},
	}
}

// Size returns the number of exact prefixes stored in the table
func (me FixedTableI) Size() int64 {
	return me.trie.NumNodes()
}

// Get returns the value in the table associated with the given network prefix
// with an exact match: both the IP and the prefix length must match. If an
// exact match is not found, found is false and value is nil and should be
// ignored.
func (me FixedTableI) Get(prefix PrefixI) (interface{}, bool) {
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
func (me FixedTableI) LongestMatch(searchPrefix PrefixI) (value interface{}, matched Match, prefix Prefix) {
	sp := searchPrefix.Prefix()
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

// TableICallback is the signature of the callback functions that can be passed to
// Walk or WalkAggregates to handle each prefix/value combination.
//
// Each invocation of your callback should return true if iteration should
// continue (as long as another key / value pair exists) or false to stop
// iterating and return immediately (meaning your callback will not be called
// again).
type TableICallback func(Prefix, interface{}) bool

// Walk invokes the given callback function for each prefix/value pair in
// the table in lexigraphical order.
//
// It returns false if iteration was stopped due to a callback return false or
// true if it iterated all items.
func (me FixedTableI) Walk(callback TableICallback) bool {
	return me.trie.Walk(trieCallback(callback))
}

// WalkAggregates invokes then given callback function for each prefix/value
// pair in the table, aggregated by value, in lexigraphical order.
//
// If two prefixes table to the same value, one contains the other, and there is
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
func (me FixedTableI) WalkAggregates(callback TableICallback) bool {
	return me.trie.Aggregate(trieCallback(callback))
}