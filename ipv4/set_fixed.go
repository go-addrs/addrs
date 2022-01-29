package ipv4

// FixedSet is like a Set except that its contents are frozen
type FixedSet struct {
	trie *setNode
}

// SetI represents something that can be treated as a FixedSet by calling
// .FixedSet(). This includes the following types: Address, Prefix, Range, Set,
// and FixedSet
type SetI interface {
	FixedSet() FixedSet
}

var _ SetI = Address{}
var _ SetI = Prefix{}
var _ SetI = Range{}
var _ SetI = Set{}
var _ SetI = FixedSet{}

// NewFixedSet returns an initialized FixedSet containing the given
// initial contents. These can be any combination of Address, FixedSet,
// Prefix, Range, and Set. The result will be the union of all of them.
func NewFixedSet(initial ...SetI) FixedSet {
	set := Set{
		s: &FixedSet{},
	}
	for _, settish := range initial {
		set.Insert(settish)
	}
	return set.FixedSet()
}

// Set returns a Set initialized with the contents of the fixed set
func (me FixedSet) Set() Set {
	return Set{
		s: &FixedSet{
			trie: me.trie,
		},
	}
}

// FixedSet implements SetI
func (me FixedSet) FixedSet() FixedSet {
	return me
}

// Size returns the number of IP addresses
func (me FixedSet) Size() int64 {
	return me.trie.Size()
}

// PrefixCallback is the type of function you pass to iterate prefixes
//
// Each invocation of your callback should return true if iteration should
// continue (as long as another key / value pair exists) or false to stop
// iterating and return immediately (meaning your callback will not be called
// again).
type PrefixCallback func(Prefix) bool

// WalkPrefixes calls `callback` for each prefix stored in lexographical
// order. It stops iteration immediately if callback returns false. It always
// uses the largest prefixes possible so if two prefixes are adjacent and can
// be combined, they will be.
//
// It returns false if iteration was stopped due to a callback return false or
// true if it iterated all items.
func (me FixedSet) WalkPrefixes(callback PrefixCallback) bool {
	return me.trie.Walk(func(prefix Prefix, data interface{}) bool {
		return callback(prefix)
	})
}

// WalkAddresses calls `callback` for each address stored in lexographical
// order. It stops iteration immediately if callback returns false.
//
// It returns false if iteration was stopped due to a callback return false or
// true if it iterated all items.
func (me FixedSet) WalkAddresses(callback AddressCallback) bool {
	return me.WalkPrefixes(func(prefix Prefix) bool {
		return prefix.WalkAddresses(callback)
	})
}

// EqualInterface returns true if this set is equal to other
func (me FixedSet) EqualInterface(other interface{}) bool {
	switch o := other.(type) {
	case SetI:
		return me.Equal(o)
	default:
		return false
	}
}

// Equal returns true if this set is equal to other
func (me FixedSet) Equal(other SetI) bool {
	return me.trie.Equal(other.FixedSet().trie)
}

// Contains tests if the given prefix is entirely contained in the set
func (me FixedSet) Contains(p PrefixI) bool {
	return nil != me.trie.Match(p.Prefix())
}

// Union returns a new set with all addresses from both sets
func (me FixedSet) Union(other SetI) FixedSet {
	return FixedSet{
		trie: me.trie.Union(other.FixedSet().trie),
	}
}

// Intersection returns a new set with all addresses that appear in both sets
func (me FixedSet) Intersection(other SetI) FixedSet {
	return FixedSet{
		trie: me.trie.Intersect(other.FixedSet().trie),
	}
}

// Difference returns a new set with all addresses that appear in this set
// excluding any that also appear in the other set
func (me FixedSet) Difference(other SetI) FixedSet {
	return FixedSet{
		trie: me.trie.Difference(other.FixedSet().trie),
	}
}

func (me FixedSet) isValid() bool {
	return me.trie.isValid()
}