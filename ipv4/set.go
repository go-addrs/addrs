package ipv4

import (
	"strings"
)

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

// NumAddresses returns the number of IP addresses
func (me Set_) NumAddresses() int64 {
	if me.s == nil {
		return 0
	}
	return me.s.NumAddresses()
}

// Contains tests if the given prefix is entirely contained in the set
func (me Set_) Contains(other SetI) bool {
	if me.s == nil {
		return other == nil || other.Set().NumAddresses() == 0
	}
	return me.s.Contains(other)
}

// Equal returns true if this set is equal to other
func (me Set_) Equal(other Set_) bool {
	if me.s == nil {
		return other.NumAddresses() == 0
	}
	return me.s.Equal(other.Set())
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

// Build is a convenience method for making modifications to a set within a
// defined scope. It calls the given callback passing a modifiable clone of
// itself. The callback can make any changes to it. After it returns true, Build
// returns the fixed snapshot of the result.
//
// If the callback returns false, modifications are aborted and the original
// fixed table is returned.
func (me Set) Build(builder func(Set_) bool) Set {
	s_ := me.Set_()
	if builder(s_) {
		return s_.Set()
	}
	return me
}

// Set implements SetI
func (me Set) Set() Set {
	return me
}

// NumAddresses returns the number of IP addresses
func (me Set) NumAddresses() int64 {
	return me.trie.NumAddresses()
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

// String returns a string representation of the set showing the minimal set of
// maximally sized prefixes that exactly cover the addresses in the set.
func (me Set) String() string {
	builder := strings.Builder{}
	builder.WriteString("[")
	var comma bool
	me.WalkPrefixes(func(p Prefix) bool {
		if comma {
			builder.WriteString(", ")
		} else {
			comma = true
		}
		builder.WriteString(p.String())
		return true
	})
	builder.WriteString("]")
	return builder.String()
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

// WalkRanges calls `callback` for each IP range in lexographical order. It
// stops iteration immediately if callback returns false. It always uses the
// largest ranges possible so if two ranges are adjacent and can be combined,
// they will be.
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

// Equal returns true if this set is equal to other
func (me Set) Equal(other Set) bool {
	return me.trie.Equal(other.trie)
}

// Contains tests if the given prefix is entirely contained in the set
func (me Set) Contains(other SetI) bool {
	if other == nil {
		other = Set{}
	}
	// NOTE This is the not the most efficient way to do this
	return other.Set().Difference(me).NumAddresses() == 0
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
