package ipv4

// Set is a structure that efficiently stores sets of IPv4 addresses and
// supports testing if an address or prefix is contained (entirely) in it.
// It supports the standard set operations: union, intersection, and difference.
// It supports conversion to/and from Ranges and Prefixes
// Sets are immutable and can be built using a SetBuilder
type Set struct {
	trie *setNode
}

// Builder returns a SetBuilder which starts with the contents of the set
func (me Set) Builder() SetBuilder {
	return SetBuilder{
		sb: &sharedSetBuilder{
			trie: me.trie,
		},
	}
}

// Size returns the number of IP addresses
func (me Set) Size() int64 {
	return me.trie.Size()
}

// PrefixCallback is the type of function you pass to iterate prefixes
//
// Each invocation of your callback should return true if iteration should
// continue (as long as another key / value pair exists) or false to stop
// iterating and return immediately (meaning your callback will not be called
// again).
type PrefixCallback func(Prefix) bool

// IteratePrefixes calls `callback` for each prefix stored in lexographical
// order. It stops iteration immediately if callback returns false. It always
// uses the largest prefixes possible so if two prefixes are adjacent and can
// be combined, they will be.
//
// It returns false if iteration was stopped due to a callback return false or
// true if it iterated all items.
func (me Set) IteratePrefixes(callback PrefixCallback) bool {
	return me.trie.Iterate(func(prefix Prefix, data interface{}) bool {
		return callback(prefix)
	})
}

// Iterate calls `callback` for each address stored in lexographical order. It
// stops iteration immediately if callback returns false.
//
// It returns false if iteration was stopped due to a callback return false or
// true if it iterated all items.
func (me Set) Iterate(callback AddressCallback) bool {
	return me.IteratePrefixes(func(prefix Prefix) bool {
		return prefix.Iterate(callback)
	})
}

// EqualInterface returns true if this set is equal to other
func (me Set) EqualInterface(other interface{}) bool {
	switch o := other.(type) {
	case Set:
		return me.Equal(o)
	default:
		return false
	}
}

// Equal returns true if this set is equal to other
func (me Set) Equal(other Set) bool {
	return me.trie.Equal(other.trie)
}

// Contains tests if the given address is in the set
func (me Set) Contains(addr Address) bool {
	return me.ContainsPrefix(addr.HostPrefix())
}

// ContainsPrefix tests if the given prefix is entirely contained in the set
func (me Set) ContainsPrefix(prefix Prefix) bool {
	return nil != me.trie.Match(prefix)
}

// Union returns a new set with all addresses from both sets
func (me Set) Union(other Set) Set {
	return Set{
		trie: me.trie.Union(other.trie),
	}
}

// Intersection returns a new set with all addresses that appear in both sets
func (me Set) Intersection(other Set) Set {
	return Set{
		trie: me.trie.Intersect(other.trie),
	}
}

// Difference returns a new set with all addresses that appear in this set
// excluding any that also appear in the other set
func (me Set) Difference(other Set) Set {
	return Set{
		trie: me.trie.Difference(other.trie),
	}
}

func (me Set) isValid() bool {
	return me.trie.isValid()
}
