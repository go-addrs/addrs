package ipv4

// Set is a structure that efficiently stores sets of addresses and
// supports testing if an address or prefix is contained (entirely) in it. It
// supports the standard set operations: union, intersection, and difference.
// It supports conversion to/and from Ranges and Prefixes
// The zero value of a Set is unitialized. Reading it is equivalent to reading
// an empty set. Attempts to modify it will result in a panic. Always use
// NewSet() to get a modifyable Set.
type Set struct {
	// See the note on Table
	s *FixedSet
}

// NewSet returns a new fully-initialized Set
func NewSet() Set {
	return Set{
		s: &FixedSet{},
	}
}

// FixedSet returns the immutable set initialized with the contents of this
// set, effectively freezing it.
func (me Set) FixedSet() FixedSet {
	if me.s == nil {
		return FixedSet{}
	}
	return FixedSet{
		trie: me.s.trie,
	}
}

// mutate should be called by any method that modifies the set in any way
func (me Set) mutate(mutator func() (ok bool, newNode *setNode)) {
	oldNode := me.s.trie
	ok, newNode := mutator()
	if ok && oldNode != newNode {
		if !swapSetNodePtr(&me.s.trie, oldNode, newNode) {
			panic("concurrent modification of Set detected")
		}
	}
}

// Insert inserts all IPs from the given set into this one. It is
// effectively a Union with the other set in place.
func (me Set) Insert(other SetI) {
	if me.s == nil {
		panic("cannot modify an unitialized Set")
	}
	if other == nil {
		other = FixedSet{}
	}
	me.mutate(func() (bool, *setNode) {
		return true, me.s.trie.Union(other.FixedSet().trie)
	})
}

// Remove removes the given set (all of its addreses) from the set. It ignores
// any addresses in the other set which were not already in the set. It is
// effectively a Difference with the other set in place.
func (me Set) Remove(other SetI) {
	if me.s == nil {
		panic("cannot modify an unitialized Set")
	}
	if other == nil {
		other = FixedSet{}
	}
	me.mutate(func() (bool, *setNode) {
		return true, me.s.trie.Difference(other.FixedSet().trie)
	})
}

// Size returns the number of IP addresses
func (me Set) Size() int64 {
	if me.s == nil {
		return 0
	}
	return me.s.Size()
}

// Contains tests if the given prefix is entirely contained in the set
func (me Set) Contains(other SetI) bool {
	if me.s == nil {
		return other == nil || other.FixedSet().Size() == 0
	}
	return me.s.Contains(other)
}

// Equal returns true if this set is equal to other
func (me Set) Equal(other SetI) bool {
	if other == nil {
		other = FixedSet{}
	}
	if me.s == nil {
		return other.FixedSet().Size() == 0
	}
	return me.s.Equal(other.FixedSet())
}

// EqualInterface returns true if this set is equal to other
func (me Set) EqualInterface(other interface{}) bool {
	switch o := other.(type) {
	case SetI:
		return me.Equal(o)
	default:
		return false
	}
}

func (me Set) isValid() bool {
	return me.s.isValid()
}

// Union returns a new fixed set with all addresses from both sets
func (me Set) Union(other SetI) FixedSet {
	if other == nil {
		other = FixedSet{}
	}
	if me.s == nil {
		return other.FixedSet()
	}
	return me.s.Union(other)
}

// Intersection returns a new fixed set with all addresses that appear in both sets
func (me Set) Intersection(other SetI) FixedSet {
	if other == nil {
		other = FixedSet{}
	}
	if me.s == nil {
		return FixedSet{}
	}
	return me.s.Intersection(other)
}

// Difference returns a new fixed set with all addresses that appear in this set
// excluding any that also appear in the other set
func (me Set) Difference(other SetI) FixedSet {
	if other == nil {
		other = FixedSet{}
	}
	if me.s == nil {
		return FixedSet{}
	}
	return me.s.Difference(other)
}

// FixedSet is like a Set except that its contents are frozen
// The zero value of a FixedSet is an empty set
type FixedSet struct {
	trie *setNode
}

// SetI represents something that can be treated as a FixedSet by calling
// .FixedSet(). It is possible to be nil. In that case, it will be treated as
// if a default zero-value FixedSet{} were passed which is an empty set.
// This includes the following types: Address, Prefix, Range, Set, and FixedSet
type SetI interface {
	FixedSet() FixedSet
}

var _ SetI = Address{}
var _ SetI = Prefix{}
var _ SetI = Range{}
var _ SetI = Set{}
var _ SetI = FixedSet{}

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
		return prefix.walkAddresses(callback)
	})
}

// RangeCallback is the type of function passed to walk individual ranges
//
// Each invocation of your callback should return true if iteration should
// continue (as long as another key / value pair exists) or false to stop
// iterating and return immediately (meaning your callback will not be called
// again).
type RangeCallback func(Range) bool

// WalkRanges calls `callback` for each address stored in lexographical
// order. It stops iteration immediately if callback returns false.
//
// It returns false if iteration was stopped due to a callback return false or
// true if it iterated all items.
func (me FixedSet) WalkRanges(callback RangeCallback) bool {
	ranges := []Range{}
	finished := me.WalkPrefixes(func(p Prefix) bool {
		if len(ranges) != 0 {
			ranges = p.Range().Plus(ranges[0])
		} else {
			ranges = []Range{p.Range()}
		}
		if len(ranges) == 2 {
			if !callback(ranges[0]) {
				return false
			}
			ranges = ranges[1:]
		}
		return true
	})
	if !finished {
		return false
	}
	if len(ranges) == 1 {
		if !callback(ranges[0]) {
			return false
		}
	}
	return true
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
	if other == nil {
		other = FixedSet{}
	}
	return me.trie.Equal(other.FixedSet().trie)
}

// Contains tests if the given prefix is entirely contained in the set
func (me FixedSet) Contains(other SetI) bool {
	if other == nil {
		other = FixedSet{}
	}
	// NOTE This is the not the most efficient way to do this
	return other.FixedSet().Difference(me).Size() == 0
}

// Union returns a new set with all addresses from both sets
func (me FixedSet) Union(other SetI) FixedSet {
	if other == nil {
		other = FixedSet{}
	}
	return FixedSet{
		trie: me.trie.Union(other.FixedSet().trie),
	}
}

// Intersection returns a new set with all addresses that appear in both sets
func (me FixedSet) Intersection(other SetI) FixedSet {
	if other == nil {
		other = FixedSet{}
	}
	return FixedSet{
		trie: me.trie.Intersect(other.FixedSet().trie),
	}
}

// Difference returns a new set with all addresses that appear in this set
// excluding any that also appear in the other set
func (me FixedSet) Difference(other SetI) FixedSet {
	if other == nil {
		other = FixedSet{}
	}
	return FixedSet{
		trie: me.trie.Difference(other.FixedSet().trie),
	}
}

func (me FixedSet) isValid() bool {
	return me.trie.isValid()
}
