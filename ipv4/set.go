package ipv4

// Set is a structure that efficiently stores sets of IPv4 addresses and
// supports testing if an address or prefix is contained (entirely) in it. It
// supports the standard set operations: union, intersection, and difference.
// It supports conversion to/and from Ranges and Prefixes
type Set struct {
	// See the note on Map
	s *ImmutableSet
}

// NewSet returns a new fully-initialized Set
func NewSet() Set {
	return Set{
		s: &ImmutableSet{},
	}
}

// ImmutableSet returns the immutable set initialized with the contents of this
// set, effectively freezing it.
func (me Set) ImmutableSet() ImmutableSet {
	return ImmutableSet{
		trie: me.s.trie,
	}
}

// Insert set inserts all IPs from the given set into this one. It is
// effectively a Union with the other set in place.
func (me Set) Insert(other Settish) {
	me.s.trie = me.s.trie.Union(other.ImmutableSet().trie)
}

// Remove removes the given prefix (all of its addreses) from the set. It
// ignores any addresses in the prefix which were not already in the set.
func (me Set) Remove(p Prefixish) {
	me.s.trie = me.s.trie.Remove(p.Prefix())
}

// RemoveSet removes the given set (all of its addreses) from the set. It
// ignores any addresses in the other set which were not already in the set. It
// is effectively a Difference with the other set in place.
func (me Set) RemoveSet(other Settish) {
	me.s.trie = me.s.trie.Difference(other.ImmutableSet().trie)
}

// Size returns the number of IP addresses
func (me Set) Size() int64 {
	return me.s.Size()
}

// Contains tests if the given prefix is entirely contained in the set
func (me Set) Contains(p Prefixish) bool {
	return me.s.Contains(p)
}

// Equal returns true if this set is equal to other
func (me Set) Equal(other Settish) bool {
	return me.s.Equal(other.ImmutableSet())
}

// EqualInterface returns true if this set is equal to other
func (me Set) EqualInterface(other interface{}) bool {
	switch o := other.(type) {
	case Settish:
		return me.Equal(o)
	default:
		return false
	}
}

func (me Set) isValid() bool {
	return me.s.isValid()
}

// Union returns a new set with all addresses from both sets
func (me Set) Union(other Settish) Set {
	is := me.s.Union(other)
	return Set{
		s: &is,
	}
}

// Intersection returns a new set with all addresses that appear in both sets
func (me Set) Intersection(other Settish) Set {
	is := me.s.Intersection(other)
	return Set{
		s: &is,
	}
}

// Difference returns a new set with all addresses that appear in this set
// excluding any that also appear in the other set
func (me Set) Difference(other Settish) Set {
	is := me.s.Difference(other)
	return Set{
		s: &is,
	}
}

// IteratePrefixes calls `callback` for each prefix stored in lexographical
// order. It stops iteration immediately if callback returns false. It always
// uses the largest prefixes possible so if two prefixes are adjacent and can
// be combined, they will be.
//
// It returns false if iteration was stopped due to a callback return false or
// true if it iterated all items.
func (me Set) IteratePrefixes(callback PrefixCallback) bool {
	return me.s.IteratePrefixes(callback)
}

// Iterate calls `callback` for each address stored in lexographical order. It
// stops iteration immediately if callback returns false.
//
// It returns false if iteration was stopped due to a callback return false or
// true if it iterated all items.
func (me Set) Iterate(callback AddressCallback) bool {
	return me.s.Iterate(callback)
}
