package ipv4

// SetBuilder is a structure used to build an immutable Set
type SetBuilder struct {
	trie *setNode
}

// Set returns the immutable set effectively freezing this builder
func (me *SetBuilder) Set() Set {
	return Set{trie: me.trie}
}

// Insert inserts the given address into the set
func (me *SetBuilder) Insert(addr Address) {
	me.InsertPrefix(ipToKey(addr))
}

// InsertPrefix inserts the given prefix (all of its addreses) into the set
func (me *SetBuilder) InsertPrefix(prefix Prefix) {
	me.trie = me.trie.Insert(prefix)
}

// Remove inserts the given address into the set
func (me *SetBuilder) Remove(addr Address) {
	me.RemovePrefix(ipToKey(addr))
}

// RemovePrefix inserts the given prefix (all of its addreses) into the set
func (me *SetBuilder) RemovePrefix(prefix Prefix) {
	me.trie = me.trie.Remove(prefix)
}
