package ipv4

import (
	"sync/atomic"
	"unsafe"
)

// Set is a structure that efficiently stores sets of IPv4 addresses and
// supports testing if an address or prefix is contained (entirely) in it. It
// supports the standard set operations: union, intersection, and difference.
// It supports conversion to/and from Ranges and Prefixes
type Set struct {
	// See the note on Table
	s *FixedSet
}

// NewSet returns a new fully-initialized Set
func NewSet(initial ...SetI) Set {
	im := NewFixedSet(initial...)
	return Set{
		s: &im,
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
		swapped := atomic.CompareAndSwapPointer(
			(*unsafe.Pointer)(
				unsafe.Pointer(&me.s.trie),
			),
			unsafe.Pointer(oldNode),
			unsafe.Pointer(newNode),
		)
		if !swapped {
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
func (me Set) Contains(s SetI) bool {
	if me.s == nil {
		return false
	}
	return me.s.Contains(s)
}

// Equal returns true if this set is equal to other
func (me Set) Equal(other SetI) bool {
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
	if me.s == nil {
		return other.FixedSet()
	}
	return me.s.Union(other)
}

// Intersection returns a new fixed set with all addresses that appear in both sets
func (me Set) Intersection(other SetI) FixedSet {
	if me.s == nil {
		return FixedSet{}
	}
	return me.s.Intersection(other)
}

// Difference returns a new fixed set with all addresses that appear in this set
// excluding any that also appear in the other set
func (me Set) Difference(other SetI) FixedSet {
	if me.s == nil {
		return FixedSet{}
	}
	return me.s.Difference(other)
}

// WalkPrefixes calls `callback` for each prefix stored in lexographical
// order. It stops iteration immediately if callback returns false. It always
// uses the largest prefixes possible so if two prefixes are adjacent and can
// be combined, they will be.
//
// It returns false if iteration was stopped due to a callback return false or
// true if it iterated all items.
func (me Set) WalkPrefixes(callback PrefixCallback) bool {
	if me.s == nil {
		return true
	}
	return me.s.WalkPrefixes(callback)
}

// WalkAddresses calls `callback` for each address stored in lexographical
// order. It stops iteration immediately if callback returns false.
//
// It returns false if iteration was stopped due to a callback return false or
// true if it iterated all items.
func (me Set) WalkAddresses(callback AddressCallback) bool {
	if me.s == nil {
		return true
	}
	return me.s.WalkAddresses(callback)
}

// WalkRanges calls `callback` for each address stored in lexographical
// order. It stops iteration immediately if callback returns false.
//
// It returns false if iteration was stopped due to a callback return false or
// true if it iterated all items.
func (me Set) WalkRanges(callback RangeCallback) bool {
	if me.s == nil {
		return true
	}
	return me.s.WalkRanges(callback)
}
