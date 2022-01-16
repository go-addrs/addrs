package ipv4

// Set is a structure that efficiently stores sets of IPv4 addresses and
// supports testing if an address or prefix is contained (entirely) in it.
// It supports the standard set operations: union, intersection, and difference.
// It supports conversion to/and from Ranges and Prefixes
// Sets are immutable and can be built using a SetBuilder
type Set struct {
	trie *trieNodeSet32
}

// Size returns the number of IP addresses
func (me Set) Size() int64 {
	return me.trie.Size()
}

// Contains tests if the given address is in the set
func (me Set) Contains(addr Addr) bool {
	return me.ContainsPrefix(ipToKey(addr))
}

// ContainsPrefix tests if the given prefix is entirely contained in the set
func (me Set) ContainsPrefix(prefix Prefix) bool {
	return nil != me.trie.Match(prefix)
}
