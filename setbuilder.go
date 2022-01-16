package ipv4

// SetBuilder is a structure used to build an immutable Set
type SetBuilder struct {
	trie *trieNodeSet32
}

// Set returns the immutable set effectively freezing this builder
func (me *SetBuilder) Set() Set {
	return Set{trie: me.trie}
}

// InsertPrefix inserts the given prefix (all of its addreses) into the set
func (me *SetBuilder) InsertPrefix(prefix Prefix) {
	me.trie = me.trie.Insert(prefix)
}
