package ipv4

import (
	"fmt"
	"math/bits"
)

// setNode is currently the same data structure as trieNode. However,
// its purpose is to implement a set of keys. Hence, values in the underlying
// data structure are completely ignored. Aliasing it in this way allows me to
// provide a completely different API on top of the same data structure and
// benefit from the trieNode API where needed by casting.
type setNode trieNode

func setNodeFromPrefix(p Prefix) *setNode {
	return &setNode{
		isActive: true,
		Prefix:   p,
		size:     1,
		h:        1,
	}
}

func setNodeFromRange(r Range) *setNode {
	// xor shows the bits that are different between first and last
	xor := r.first.ui ^ r.last.ui
	// The number of leading zeroes in the xor is the number of bits the two addresses have in common
	numCommonBits := bits.LeadingZeros32(xor)

	if numCommonBits == bits.OnesCount32(^xor) {
		// This range is exactly one prefix, return a node with it.
		prefix := Prefix{r.first, uint32(numCommonBits)}
		return setNodeFromPrefix(prefix)
	}

	// "pivot" is the address within the range with the most trailing zeroes.
	// Dividing and conquering on it recursively teases out all of the largest
	// prefixes in the range. The result is the smallest set of prefixes that
	// covers it. It takes Log(p) time where p is the number of prefixes in the
	// result -- bounded by 32 x 2 in the worst case
	pivot := r.first.ui & (uint32(0xffffffff) << (32 - numCommonBits))
	pivot |= uint32(0x80000000) >> numCommonBits

	a := setNodeFromRange(Range{r.first, Address{pivot - 1}})
	b := setNodeFromRange(Range{Address{pivot}, r.last})
	return a.Union(b)
}

// Insert inserts the key / value if the key didn't previously exist and then
// flattens the structure (without regard to any values) to remove nested
// prefixes resulting in a flat list of disjoint prefixes.
func (me *setNode) Insert(key Prefix) *setNode {
	newHead, _ := (*trieNode)(me).insert(&trieNode{Prefix: key, Data: nil}, insertOpts{insert: true, update: true, flatten: true})
	return (*setNode)(newHead)
}

// Delete removes a prefix from the trie and returns the new root of the trie.
// It is important to note that the root of the trie can change. Like Insert,
// this is designed for using trie as a set of keys, completely ignoring
// values. All stored prefixes that match the given prefix with LPM will be
// removed, not just exact matches.
func (me *setNode) Remove(key Prefix) *setNode {
	newHead, _ := (*trieNode)(me).del(key, deleteOpts{flatten: true})
	return (*setNode)(newHead)
}

func (me *setNode) Left() *setNode {
	return (*setNode)(me.children[0])
}

func (me *setNode) Right() *setNode {
	return (*setNode)(me.children[1])
}

// so much casting!
func (me *setNode) mutate(mutator func(*setNode)) *setNode {
	n := (*trieNode)(me)
	n = n.mutate(func(node *trieNode) {
		mutator((*setNode)(node))
	})
	return (*setNode)(n)
}

func (me *setNode) flatten() {
	(*trieNode)(me).flatten()
}

// Union returns the flattened union of prefixes.
func (me *setNode) Union(other *setNode) (rc *setNode) {
	if me == other {
		return me
	}
	if other == nil {
		return me
	}
	if me == nil {
		return other
	}
	// Test containership both ways
	result, reversed, common, child := compare(me.Prefix, other.Prefix)
	switch result {
	case compareSame:
		if me.isActive {
			return me
		}
		if other.isActive {
			return other
		}
		left := me.Left().Union(other.Left())
		right := me.Right().Union(other.Right())
		if left == me.Left() && right == me.Right() {
			return me
		}
		newHead := &setNode{
			Prefix: Prefix{
				addr: Address{
					ui: me.Prefix.addr.ui,
				},
				length: me.Prefix.length,
			},
			children: [2]*trieNode{
				(*trieNode)(left),
				(*trieNode)(right),
			},
		}
		return newHead.mutate(func(n *setNode) {
			n.flatten()
		})

	case compareContains, compareIsContained:
		super, sub := me, other
		if reversed {
			super, sub = sub, super
		}
		if super.isActive {
			return super
		}

		var left, right *setNode

		if child == 1 {
			left, right = super.Left(), super.Right().Union(sub)
		} else {
			left, right = super.Left().Union(sub), super.Right()
		}
		newHead := &setNode{
			Prefix: Prefix{
				addr: Address{
					ui: super.Prefix.addr.ui,
				},
				length: super.Prefix.length,
			},
			children: [2]*trieNode{
				(*trieNode)(left),
				(*trieNode)(right),
			},
		}
		return newHead.mutate(func(n *setNode) {
			n.flatten()
		})

	default:
		var left, right *setNode

		if (child == 1) != reversed { // (child == 1) XOR reversed
			left, right = me, other
		} else {
			left, right = other, me
		}

		newHead := &setNode{
			Prefix: Prefix{
				addr: Address{
					ui: me.Prefix.addr.ui & ^(uint32(0xffffffff) >> common), // zero out bits not in common
				},
				length: common,
			},
			children: [2]*trieNode{
				(*trieNode)(left),
				(*trieNode)(right),
			},
		}
		return newHead.mutate(func(n *setNode) {
			n.flatten()
		})
	}
}

func (me *setNode) Match(searchKey Prefix) *setNode {
	return (*setNode)((*trieNode)(me).Match(searchKey))
}

func (me *setNode) isValid() bool {
	return (*trieNode)(me).isValid()
}

// Difference returns the flattened difference of prefixes.
func (me *setNode) Difference(other *setNode) (rc *setNode) {
	if me == nil || other == nil {
		return me
	}

	result, _, _, child := compare(me.Prefix, other.Prefix)
	switch result {
	case compareIsContained:
		if other.isActive {
			return nil
		}
		return me.Difference((*setNode)(other.children[child]))
	case compareDisjoint:
		return me
	}

	if !me.isActive {
		left := me.Left().Difference(other)
		right := me.Right().Difference(other)
		if left == me.Left() && right == me.Right() {
			return me
		}

		return left.Union(right)
	}

	// Assumes `me` is active as checked above
	halves := func() (a, b *setNode) {
		aPrefix, bPrefix := me.Prefix.Halves()
		return setNodeFromPrefix(aPrefix), setNodeFromPrefix(bPrefix)
	}

	switch result {
	case compareSame:
		if other.isActive {
			return nil
		}
		a, b := halves()
		return a.Difference(other.Left()).Union(
			b.Difference(other.Right()),
		)

	case compareContains:
		a, b := halves()
		halves := [2]*setNode{a, b}
		whole := halves[(child+1)%2]
		partial := halves[child].Difference(other)
		return whole.Union(partial)
	}
	panic("unreachable")
}

// Intersect returns the flattened intersection of prefixes
func (me *setNode) Intersect(other *setNode) *setNode {
	if me == nil || other == nil {
		return nil
	}

	result, reversed, _, _ := compare(me.Prefix, other.Prefix)
	if result == compareDisjoint {
		return nil
	}
	if !me.isActive {
		return other.Intersect(me.Left()).Union(
			other.Intersect(me.Right()),
		)
	}
	if !other.isActive {
		return me.Intersect(other.Left()).Union(
			me.Intersect(other.Right()),
		)
	}
	// Return the smaller prefix
	if reversed {
		return me
	}
	return other
}

func (me *setNode) Equal(other *setNode) bool {
	return (*trieNode)(me).Equal((*trieNode)(other), func(a, b interface{}) bool {
		return true
	})
}

// NumAddresses calls trieNode NumAddresses
func (me *setNode) NumAddresses() int64 {
	return (*trieNode)(me).NumAddresses()
}

// NumNodes returns the number of entries in the trie
func (me *setNode) NumNodes() int64 {
	return (*trieNode)(me).NumNodes()
}

func (me *setNode) height() int {
	return (*trieNode)(me).height()
}

func (me *setNode) Walk(callback func(Prefix, interface{}) bool) bool {
	return (*trieNode)(me).Walk(callback)
}

func best(left, right func() (Prefix, error), length uint32) (Prefix, error) {
	lPrefix, lErr := left()
	if lErr == nil {
		if lPrefix.length == length {
			return lPrefix, nil
		}
		rPrefix, rErr := right()
		if rErr == nil {
			if lPrefix.length < rPrefix.length {
				return rPrefix, nil
			} else {
				return lPrefix, nil
			}
		}
		return lPrefix, nil
	}

	rPrefix, rErr := right()
	if rErr == nil {
		return rPrefix, nil
	}
	return Prefix{}, fmt.Errorf("cannot find containing prefix")
}

func (me *setNode) findSmallestContainingPrefix(length uint32) (Prefix, error) {
	if me == nil || length < me.Prefix.length {
		return Prefix{}, fmt.Errorf("cannot find containing prefix")
	}
	if length == me.Prefix.length {
		if me.isActive {
			return me.Prefix, nil
		}
	}

	l, r := (*setNode)(me.children[0]), (*setNode)(me.children[1])
	bestPrefix, err := best(
		func() (Prefix, error) { return l.findSmallestContainingPrefix(length) },
		func() (Prefix, error) { return r.findSmallestContainingPrefix(length) },
		length,
	)
	if err == nil {
		return bestPrefix, nil
	}
	if !me.isActive {
		return Prefix{}, fmt.Errorf("cannot find containing prefix")
	}
	return me.Prefix, nil
}

func (me *setNode) FindSmallestContainingPrefix(reserved *setNode, length uint32) (Prefix, error) {
	if me == nil || length < me.Prefix.length {
		return Prefix{}, fmt.Errorf("cannot find containing prefix")
	}
	if reserved == nil {
		return me.findSmallestContainingPrefix(length)
	}

	result, _, _, child := compare(me.Prefix, reserved.Prefix)
	switch result {
	case compareIsContained:
		if reserved.isActive {
			return Prefix{}, fmt.Errorf("cannot find containing prefix")
		}
		return me.FindSmallestContainingPrefix((*setNode)(reserved.children[child]), length)
	case compareDisjoint:
		return me.findSmallestContainingPrefix(length)
	}

	if !me.isActive {
		return best(
			func() (Prefix, error) { return me.Left().FindSmallestContainingPrefix(reserved, length) },
			func() (Prefix, error) { return me.Right().FindSmallestContainingPrefix(reserved, length) },
			length,
		)
	}

	// Assumes `me` is active as checked above
	halves := func() (a, b *setNode) {
		aPrefix, bPrefix := me.Prefix.Halves()
		return setNodeFromPrefix(aPrefix), setNodeFromPrefix(bPrefix)
	}

	switch result {
	case compareSame:
		if reserved.isActive {
			return Prefix{}, fmt.Errorf("cannot find containing prefix")
		}
		left, right := halves()
		return best(
			func() (Prefix, error) { return left.FindSmallestContainingPrefix(reserved.Left(), length) },
			func() (Prefix, error) { return right.FindSmallestContainingPrefix(reserved.Right(), length) },
			length,
		)

	case compareContains:
		left, right := halves()
		halves := [2]*setNode{left, right}
		whole, partial := halves[(child+1)%2], halves[child]
		return best(
			func() (Prefix, error) { return whole.findSmallestContainingPrefix(length) },
			func() (Prefix, error) { return partial.FindSmallestContainingPrefix(reserved, length) },
			length,
		)
	}
	panic("unreachable")
}
