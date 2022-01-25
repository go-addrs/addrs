package ipv4

type sharedSetBuilder struct {
	trie *setNode
}

// SetBuilder is a structure used to build an immutable Set
type SetBuilder struct {
	sb *sharedSetBuilder
}

// NewSetBuilder returns a new fully-initialized SetBuilder
func NewSetBuilder() SetBuilder {
	return SetBuilder{
		sb: &sharedSetBuilder{},
	}
}

// Set returns the immutable set effectively freezing this builder
func (me *SetBuilder) Set() Set {
	return Set{trie: me.sb.trie}
}

// Insert inserts the given address into the set
func (me SetBuilder) Insert(addr Address) {
	me.InsertPrefix(addr.HostPrefix())
}

// InsertPrefix inserts the given prefix (all of its addreses) into the set
func (me SetBuilder) InsertPrefix(prefix Prefix) {
	me.sb.trie = me.sb.trie.Insert(prefix)
}

// Remove inserts the given address into the set
func (me SetBuilder) Remove(addr Address) {
	me.RemovePrefix(addr.HostPrefix())
}

// RemovePrefix inserts the given prefix (all of its addreses) into the set
func (me SetBuilder) RemovePrefix(prefix Prefix) {
	me.sb.trie = me.sb.trie.Remove(prefix)
}
