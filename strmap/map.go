//go:build go1.18
// +build go1.18

package strmap

// Map_ is a mutable string-keyed copy-on-write (COW) map type.
//
// The zero value is uninitialized. Use NewMap_() to create one.
type Map_[T any] struct {
	m *mapStringX
}

// NewMap_ returns a new mutable Map_ for comparable values.
func NewMap_[T comparable]() Map_[T] {
	return Map_[T]{newMapX()}
}

// NewMapCustomCompare_ returns a new mutable Map_ using a custom
// equality comparator (for values that are not comparable with ==).
func NewMapCustomCompare_[T any](comparator func(a, b T) bool) Map_[T] {
	return Map_[T]{
		newMapXCustomCompare(func(a, b interface{}) bool {
			return comparator(a.(T), b.(T))
		}),
	}
}

// NumEntries returns the number of entries in the map.
func (me Map_[T]) NumEntries() int64 {
	if me.m == nil {
		return 0
	}
	return me.m.trie.NumNodes()
}

// Insert inserts key/value. Returns false without modifying the map if the key
// already exists.
func (me Map_[T]) Insert(key string, value T) (succeeded bool) {
	if me.m == nil {
		panic("cannot modify an uninitialized Map_")
	}
	var err error
	me.m.mutate(func() *trieNode {
		var newHead *trieNode
		newHead, err = me.m.trie.Insert(key, value)
		if err != nil {
			return me.m.trie
		}
		return newHead
	})
	return err == nil
}

// Update updates an existing key's value. Returns false without modifying the
// map if the key does not exist.
func (me Map_[T]) Update(key string, value T) (updated bool) {
	if me.m == nil {
		panic("cannot modify an uninitialized Map_")
	}
	var err error
	me.m.mutate(func() *trieNode {
		var newHead *trieNode
		newHead, err = me.m.trie.Update(key, value, me.m.eq)
		if err != nil {
			return me.m.trie
		}
		return newHead
	})
	return err == nil
}

// Get returns the value associated with key using an exact match. If no exact
// match is found, ok is false.
func (me Map_[T]) Get(key string) (t T, ok bool) {
	if me.m == nil {
		return t, false
	}
	node := me.m.trie.Match(key)
	if node == nil || node.Key != key || !node.isActive {
		return t, false
	}
	t, _ = node.Data.(T)
	return t, true
}

// Remove removes the entry for key. Returns false if no entry with that exact
// key exists.
func (me Map_[T]) Remove(key string) (succeeded bool) {
	if me.m == nil {
		panic("cannot modify an uninitialized Map_")
	}
	var err error
	me.m.mutate(func() *trieNode {
		var newHead *trieNode
		newHead, err = me.m.trie.Delete(key)
		if err != nil {
			return me.m.trie
		}
		return newHead
	})
	return err == nil
}

// Map returns an immutable snapshot of this Map_. Due to the COW
// nature of the underlying structure, this is effectively a pointer copy.
func (me Map_[T]) Map() Map[T] {
	if me.m == nil {
		return Map[T]{}
	}
	return Map[T]{*me.m}
}

// Map is an immutable string-keyed copy-on-write map. For a mutable
// equivalent, see Map_.
//
// The zero value is an empty map.
type Map[T any] struct {
	x mapStringX
}

// Map_ returns a mutable map initialized with the contents of this one.
// Due to COW, this is effectively a pointer copy.
func (me Map[T]) Map_() Map_[T] {
	if me.x.eq == nil {
		me.x.eq = defaultComparator
	}
	return Map_[T]{&me.x}
}

// Build is a convenience method for scoped mutations. It passes a mutable
// clone to the callback; if the callback returns true the result is returned as
// a new immutable snapshot, otherwise the original is returned unchanged.
func (me Map[T]) Build(builder func(Map_[T]) bool) Map[T] {
	m_ := me.Map_()
	if builder(m_) {
		return m_.Map()
	}
	return me
}

// NumEntries returns the number of entries in the map.
func (me Map[T]) NumEntries() int64 {
	return me.x.trie.NumNodes()
}

// Get returns the value associated with key using an exact match.
func (me Map[T]) Get(key string) (t T, ok bool) {
	node := me.x.trie.Match(key)
	if node == nil || node.Key != key || !node.isActive {
		return t, false
	}
	t, _ = node.Data.(T)
	return t, true
}

// Walk invokes callback for each entry in lexicographical order. Stops early
// and returns false if callback returns false.
func (me Map[T]) Walk(callback func(string, T) bool) bool {
	return me.x.trie.Walk(func(key string, data interface{}) bool {
		var t T
		t, _ = data.(T)
		return callback(key, t)
	})
}

// Diff calls the provided callbacks for entries that differ between me and
// other. changed is called for keys present in both with different values;
// left and right are called for keys present only on the respective side;
// unchanged is called for keys present in both with equal values. Any callback
// may be nil. Returns false if iteration was stopped by a callback.
func (me Map[T]) Diff(
	other Map[T],
	changed func(key string, left, right T) bool,
	left, right, unchanged func(string, T) bool,
) bool {
	handler := trieDiffHandler{}
	if left != nil {
		handler.Removed = func(n *trieNode) bool {
			var t T
			t, _ = n.Data.(T)
			return left(n.Key, t)
		}
	}
	if right != nil {
		handler.Added = func(n *trieNode) bool {
			var t T
			t, _ = n.Data.(T)
			return right(n.Key, t)
		}
	}
	if changed != nil {
		handler.Modified = func(l, r *trieNode) bool {
			var lt, rt T
			lt, _ = l.Data.(T)
			rt, _ = r.Data.(T)
			return changed(l.Key, lt, rt)
		}
	}
	if unchanged != nil {
		handler.Same = func(n *trieNode) bool {
			var t T
			t, _ = n.Data.(T)
			return unchanged(n.Key, t)
		}
	}
	return me.x.trie.Diff(other.x.trie, handler, me.x.eq)
}

// mapStringX is the internal immutable container.
type mapStringX struct {
	trie *trieNode
	eq   comparator
}

func defaultComparator(a, b interface{}) bool {
	return a == b
}

func newMapX() *mapStringX {
	return &mapStringX{nil, defaultComparator}
}

func newMapXCustomCompare(eq comparator) *mapStringX {
	return &mapStringX{nil, eq}
}

func (me *mapStringX) mutate(mutator func() *trieNode) {
	me.trie = mutator()
}
