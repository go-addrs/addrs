package ipv4

// Set_ is the mutable version of a Set, allowing insertion and deletion of
// elements.
// The zero value of a Set_ is unitialized. Reading it is equivalent to reading
// an empty set. Attempts to modify it will result in a panic. Always use
// NewSet_() to get an initialized Set_.
type Set_ struct {
	// See the note on Table_
	s *Set
}

// NewSet_ returns a new fully-initialized Set_
func NewSet_() Set_ {
	return Set_{
		s: &Set{},
	}
}

// Set returns the immutable set initialized with the contents of this
// set, effectively freezing it.
func (me Set_) Set() Set {
	if me.s == nil {
		return Set{}
	}
	return Set{
		trie: me.s.trie,
	}
}

// mutate should be called by any method that modifies the set in any way
func (me Set_) mutate(mutator func() (ok bool, newNode *setNode)) {
	oldNode := me.s.trie
	ok, newNode := mutator()
	if ok && oldNode != newNode {
		if !swapSetNodePtr(&me.s.trie, oldNode, newNode) {
			panic("concurrent modification of Set_ detected")
		}
	}
}

// Insert inserts all IPs from the given set into this one. It is
// effectively a Union with the other set in place.
func (me Set_) Insert(other SetI) {
	if me.s == nil {
		panic("cannot modify an unitialized Set_")
	}
	if other == nil {
		other = Set{}
	}
	me.mutate(func() (bool, *setNode) {
		return true, me.s.trie.Union(other.Set().trie)
	})
}

// Remove removes the given set (all of its addreses) from the set. It ignores
// any addresses in the other set which were not already in the set. It is
// effectively a Difference with the other set in place.
func (me Set_) Remove(other SetI) {
	if me.s == nil {
		panic("cannot modify an unitialized Set_")
	}
	if other == nil {
		other = Set{}
	}
	me.mutate(func() (bool, *setNode) {
		return true, me.s.trie.Difference(other.Set().trie)
	})
}

// Size returns the number of IP addresses
func (me Set_) Size() int64 {
	if me.s == nil {
		return 0
	}
	return me.s.Size()
}

// Contains tests if the given prefix is entirely contained in the set
func (me Set_) Contains(other SetI) bool {
	if me.s == nil {
		return other == nil || other.Set().Size() == 0
	}
	return me.s.Contains(other)
}

// Equal returns true if this set is equal to other
func (me Set_) Equal(other SetI) bool {
	if other == nil {
		other = Set{}
	}
	if me.s == nil {
		return other.Set().Size() == 0
	}
	return me.s.Equal(other.Set())
}

// EqualInterface returns true if this set is equal to other
func (me Set_) EqualInterface(other interface{}) bool {
	switch o := other.(type) {
	case SetI:
		return me.Equal(o)
	default:
		return false
	}
}

func (me Set_) isValid() bool {
	return me.s.isValid()
}

// Union returns a new fixed set with all addresses from both sets
func (me Set_) Union(other SetI) Set {
	if other == nil {
		other = Set{}
	}
	if me.s == nil {
		return other.Set()
	}
	return me.s.Union(other)
}

// Intersection returns a new fixed set with all addresses that appear in both sets
func (me Set_) Intersection(other SetI) Set {
	if other == nil {
		other = Set{}
	}
	if me.s == nil {
		return Set{}
	}
	return me.s.Intersection(other)
}

// Difference returns a new fixed set with all addresses that appear in this set
// excluding any that also appear in the other set
func (me Set_) Difference(other SetI) Set {
	if other == nil {
		other = Set{}
	}
	if me.s == nil {
		return Set{}
	}
	return me.s.Difference(other)
}

// Set is a structure that efficiently stores sets of addresses and supports
// testing if an address or prefix is contained (entirely) in it. It supports
// the standard set operations: union, intersection, and difference. It
// supports conversion to/and from Ranges and Prefixes. The zero value of a Set
// is an empty set
// Set is immutable. For a mutable equivalent, see Set_.
type Set struct {
	trie *setNode
}

// SetI represents something that can be treated as a Set by calling .Set() --
// Address, Prefix, Range, Set, and Set_. It is possible to be nil in which
// case, it will be treated as a zero-value Set{} which is empty.
type SetI interface {
	Set() Set
}

var _ SetI = Address{}
var _ SetI = Prefix{}
var _ SetI = Range{}
var _ SetI = Set_{}
var _ SetI = Set{}

// Set_ returns a Set_ initialized with the contents of the fixed set
func (me Set) Set_() Set_ {
	return Set_{
		s: &Set{
			trie: me.trie,
		},
	}
}

// Set implements SetI
func (me Set) Set() Set {
	return me
}

// Size returns the number of IP addresses
func (me Set) Size() int64 {
	return me.trie.Size()
}

// WalkPrefixes calls `callback` for each prefix stored in lexographical
// order. It stops iteration immediately if callback returns false. It always
// uses the largest prefixes possible so if two prefixes are adjacent and can
// be combined, they will be.
//
// It returns false if iteration was stopped due to a callback return false or
// true if it iterated all items.
func (me Set) WalkPrefixes(callback func(Prefix) bool) bool {
	return me.trie.Walk(func(prefix Prefix, data interface{}) bool {
		return callback(prefix)
	})
}

// WalkAddresses calls `callback` for each address stored in lexographical
// order. It stops iteration immediately if callback returns false.
//
// It returns false if iteration was stopped due to a callback return false or
// true if it iterated all items.
func (me Set) WalkAddresses(callback func(Address) bool) bool {
	return me.WalkPrefixes(func(prefix Prefix) bool {
		return prefix.walkAddresses(callback)
	})
}

// WalkRanges calls `callback` for each address stored in lexographical
// order. It stops iteration immediately if callback returns false.
//
// It returns false if iteration was stopped due to a callback return false or
// true if it iterated all items.
func (me Set) WalkRanges(callback func(Range) bool) bool {
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
func (me Set) EqualInterface(other interface{}) bool {
	switch o := other.(type) {
	case SetI:
		return me.Equal(o)
	default:
		return false
	}
}

// Equal returns true if this set is equal to other
func (me Set) Equal(other SetI) bool {
	if other == nil {
		other = Set{}
	}
	return me.trie.Equal(other.Set().trie)
}

// Contains tests if the given prefix is entirely contained in the set
func (me Set) Contains(other SetI) bool {
	if other == nil {
		other = Set{}
	}
	// NOTE This is the not the most efficient way to do this
	return other.Set().Difference(me).Size() == 0
}

// Union returns a new set with all addresses from both sets
func (me Set) Union(other SetI) Set {
	if other == nil {
		other = Set{}
	}
	return Set{
		trie: me.trie.Union(other.Set().trie),
	}
}

// Intersection returns a new set with all addresses that appear in both sets
func (me Set) Intersection(other SetI) Set {
	if other == nil {
		other = Set{}
	}
	return Set{
		trie: me.trie.Intersect(other.Set().trie),
	}
}

// Difference returns a new set with all addresses that appear in this set
// excluding any that also appear in the other set
func (me Set) Difference(other SetI) Set {
	if other == nil {
		other = Set{}
	}
	return Set{
		trie: me.trie.Difference(other.Set().trie),
	}
}

func (me Set) isValid() bool {
	return me.trie.isValid()
}
