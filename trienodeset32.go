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
