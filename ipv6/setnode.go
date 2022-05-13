package ipv6

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
	xor := r.first.ui.xor(r.last.ui)
	// The number of leading zeroes in the xor is the number of bits the two addresses have in common
	numCommonBits := xor.leadingZeros()

	if numCommonBits == xor.complement().onesCount() {
		// This range is exactly one prefix, return a node with it.
		prefix := Prefix{r.first, uint32(numCommonBits)}
		return setNodeFromPrefix(prefix)
	}

	// "pivot" is the address within the range with the most trailing zeroes.
	// Dividing and conquering on it recursively teases out all of the largest
	// prefixes in the range. The result is the smallest set of prefixes that
	// covers it. It takes Log(p) time where p is the number of prefixes in the
	// result -- bounded by 128 x 2 in the worst case
	pivot := r.first.ui.and(uint128{0xffffffffffffffff, 0xffffffffffffffff}.leftShift(128 - numCommonBits))
	pivot = pivot.or(uint128{0x8000000000000000, 0}.rightShift(numCommonBits))

	a := setNodeFromRange(Range{r.first, Address{pivot.subtractUint64(1)}})
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
					ui: me.Prefix.addr.ui.and(uint128{0xffffffffffffffff, 0xffffffffffffffff}.rightShift(int(common)).complement()), // zero out bits not in common
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

// IsEmpty calls trieNode IsEmpty
func (me *setNode) IsEmpty() bool {
	return (*trieNode)(me).IsEmpty()
}

// NumNodes returns the number of entries in the trie
func (me *setNode) NumNodes() int64 {
	return (*trieNode)(me).NumNodes()
}

func (me *setNode) height() int {
	return (*trieNode)(me).height()
}

func (me *setNode) active() bool {
	return (*trieNode)(me).active()
}

func (me *setNode) Walk(callback func(Prefix, interface{}) bool) bool {
	return (*trieNode)(me).Walk(callback)
}
