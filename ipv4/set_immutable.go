package ipv4

// ImmutableSet is like a Set except that its contents are frozen
type ImmutableSet struct {
	trie *setNode
}

// NewImmutableSet returns an initialized but empty ImmutableSet
func NewImmutableSet() ImmutableSet {
	return ImmutableSet{}
}

// Set returns a Set initialized with the contents of the immutable set
func (me ImmutableSet) Set() Set {
	return Set{
		s: &ImmutableSet{
			trie: me.trie,
		},
	}
}

// Size returns the number of IP addresses
func (me ImmutableSet) Size() int64 {
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
func (me ImmutableSet) IteratePrefixes(callback PrefixCallback) bool {
	return me.trie.Iterate(func(prefix Prefix, data interface{}) bool {
		return callback(prefix)
	})
}

// Iterate calls `callback` for each address stored in lexographical order. It
// stops iteration immediately if callback returns false.
//
// It returns false if iteration was stopped due to a callback return false or
// true if it iterated all items.
func (me ImmutableSet) Iterate(callback AddressCallback) bool {
	return me.IteratePrefixes(func(prefix Prefix) bool {
		return prefix.Iterate(callback)
	})
}

// EqualInterface returns true if this set is equal to other
func (me ImmutableSet) EqualInterface(other interface{}) bool {
	switch o := other.(type) {
	case ImmutableSet:
		return me.Equal(o)
	default:
		return false
	}
}

// Equal returns true if this set is equal to other
func (me ImmutableSet) Equal(other ImmutableSet) bool {
	return me.trie.Equal(other.trie)
}

// Contains tests if the given prefix is entirely contained in the set
func (me ImmutableSet) Contains(p Prefixish) bool {
	return nil != me.trie.Match(p.Prefix())
}

// Union returns a new set with all addresses from both sets
func (me ImmutableSet) Union(other ImmutableSet) ImmutableSet {
	return ImmutableSet{
		trie: me.trie.Union(other.trie),
	}
}

// Intersection returns a new set with all addresses that appear in both sets
func (me ImmutableSet) Intersection(other ImmutableSet) ImmutableSet {
	return ImmutableSet{
		trie: me.trie.Intersect(other.trie),
	}
}

// Difference returns a new set with all addresses that appear in this set
// excluding any that also appear in the other set
func (me ImmutableSet) Difference(other ImmutableSet) ImmutableSet {
	return ImmutableSet{
		trie: me.trie.Difference(other.trie),
	}
}

func (me ImmutableSet) isValid() bool {
	return me.trie.isValid()
}
