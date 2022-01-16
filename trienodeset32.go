package ipv4

// trieNodeSet32 is currently the same data structure as trieNode32. However,
// its purpose is to implement a set of keys. Hence, values in the underlying
// data structure are completely ignored. Aliasing it in this way allows me to
// provide a completely different API on top of the same data structure and
// benefit from the trieNode32 API where needed by casting.
type trieNodeSet32 trieNode32

func trieNodeSet32FromPrefix(p Prefix) *trieNodeSet32 {
	return &trieNodeSet32{
		isActive: true,
		Prefix:   p,
		size:     1,
		h:        1,
	}
}

func (me *trieNodeSet32) halves() (a, b *trieNodeSet32) {
	if !me.isActive {
		return
	}
	aPrefix, bPrefix := me.Prefix.Halves()
	return trieNodeSet32FromPrefix(aPrefix), trieNodeSet32FromPrefix(bPrefix)
}

// Insert inserts the key / value if the key didn't previously exist and then
// flattens the structure (without regard to any values) to remove nested
// prefixes resulting in a flat list of disjoint prefixes.
func (me *trieNodeSet32) Insert(key Prefix) *trieNodeSet32 {
	newHead, _ := (*trieNode32)(me).insert(&trieNode32{Prefix: key, Data: nil}, insertOpts{insert: true, update: true, flatten: true})
	return (*trieNodeSet32)(newHead)
}

func (me *trieNodeSet32) Left() *trieNodeSet32 {
	return (*trieNodeSet32)(me.children[0])
}

func (me *trieNodeSet32) Right() *trieNodeSet32 {
	return (*trieNodeSet32)(me.children[1])
}

// Union returns the flattened union of prefixes.
func (me *trieNodeSet32) Union(other *trieNodeSet32) (rc *trieNodeSet32) {
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
	result, reversed, common, child := compare32(me.Prefix, other.Prefix)
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
		newHead := &trieNodeSet32{
			Prefix: Prefix{
				Addr: Addr{
					ui: me.Prefix.Addr.ui,
				},
				length: me.Prefix.length,
			},
			children: [2]*trieNode32{
				(*trieNode32)(left),
				(*trieNode32)(right),
			},
		}
		newHead.setSize()
		return newHead

	case compareContains, compareIsContained:
		super, sub := me, other
		if reversed {
			super, sub = sub, super
		}
		if super.isActive {
			return super
		}

		var left, right *trieNodeSet32

		if child == 1 {
			left, right = super.Left(), super.Right().Union(sub)
		} else {
			left, right = super.Left().Union(sub), super.Right()
		}
		newHead := &trieNodeSet32{
			Prefix: Prefix{
				Addr: Addr{
					ui: super.Prefix.Addr.ui,
				},
				length: super.Prefix.length,
			},
			children: [2]*trieNode32{
				(*trieNode32)(left),
				(*trieNode32)(right),
			},
		}
		newHead.setSize()
		return newHead

	default:
		var left, right *trieNodeSet32

		if (child == 1) != reversed { // (child == 1) XOR reversed
			left, right = me, other
		} else {
			left, right = other, me
		}

		newHead := &trieNodeSet32{
			Prefix: Prefix{
				Addr: Addr{
					ui: me.Prefix.Addr.ui & ^(uint32(0xffffffff) >> common), // zero out bits not in common
				},
				length: common,
			},
			children: [2]*trieNode32{
				(*trieNode32)(left),
				(*trieNode32)(right),
			},
		}
		newHead.setSize()
		return newHead
	}
}

func (me *trieNodeSet32) Match(searchKey Prefix) *trieNodeSet32 {
	return (*trieNodeSet32)((*trieNode32)(me).Match(searchKey))
}

func (me *trieNodeSet32) isValid() bool {
	return (*trieNode32)(me).isValid()
}

func (me *trieNodeSet32) setSize() {
	(*trieNode32)(me).setSize()
}

// Size calls trieNode32 Size
func (me *trieNodeSet32) Size() int64 {
	return (*trieNode32)(me).Size()
}

// NumNodes returns the number of entries in the trie
func (me *trieNodeSet32) NumNodes() int {
	return (*trieNode32)(me).NumNodes()
}

func (me *trieNodeSet32) height() int {
	return (*trieNode32)(me).height()
}
