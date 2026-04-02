//go:build go1.18
// +build go1.18

package strmap

// MapString_ is a mutable string-keyed map. It supports both exact-match and
// longest-prefix-match lookups, where "prefix" means string prefix: a stored
// key "foo" matches any query that starts with "foo".
//
// The zero value is uninitialized. Use NewMapString_() to create one.
type MapString_[T any] struct {
	m *mapStringX
}

// NewMapString_ returns a new mutable MapString_ for comparable values.
func NewMapString_[T comparable]() MapString_[T] {
	return MapString_[T]{newMapStringX()}
}

// NewMapStringCustomCompare_ returns a new mutable MapString_ using a custom
// equality comparator (for values that are not comparable with ==).
func NewMapStringCustomCompare_[T any](comparator func(a, b T) bool) MapString_[T] {
	return MapString_[T]{
		newMapStringXCustomCompare(func(a, b interface{}) bool {
			return comparator(a.(T), b.(T))
		}),
	}
}

// NumEntries returns the number of entries in the map.
func (me MapString_[T]) NumEntries() int64 {
	if me.m == nil {
		return 0
	}
	return me.m.trie.NumNodes()
}

// Insert inserts key/value. Returns false without modifying the map if the key
// already exists.
func (me MapString_[T]) Insert(key string, value T) (succeeded bool) {
	if me.m == nil {
		panic("cannot modify an uninitialized MapString_")
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
func (me MapString_[T]) Update(key string, value T) (updated bool) {
	if me.m == nil {
		panic("cannot modify an uninitialized MapString_")
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

// InsertOrUpdate inserts key/value if absent, or updates the value if present.
func (me MapString_[T]) InsertOrUpdate(key string, value T) {
	if me.m == nil {
		panic("cannot modify an uninitialized MapString_")
	}
	me.m.mutate(func() *trieNode {
		return me.m.trie.InsertOrUpdate(key, value, me.m.eq)
	})
}

// Get returns the value associated with key using an exact match. If no exact
// match is found, ok is false.
func (me MapString_[T]) Get(key string) (t T, ok bool) {
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

// GetOrInsert returns the value for key if it exists, otherwise inserts the
// given default value and returns it.
func (me MapString_[T]) GetOrInsert(key string, value T) T {
	if me.m == nil {
		panic("cannot modify an uninitialized MapString_")
	}
	var node *trieNode
	me.m.mutate(func() *trieNode {
		var newHead *trieNode
		newHead, node = me.m.trie.GetOrInsert(key, value)
		return newHead
	})
	t, _ := node.Data.(T)
	return t
}

// LongestMatch returns the value associated with the longest stored key that
// is a string prefix of the given key. For example, if "foo" and "foobar" are
// both stored and the query is "foobarbaz", "foobar" is returned.
//
// If a match is found, found is true and matchKey is the key that matched.
func (me MapString_[T]) LongestMatch(key string) (value T, found bool, matchKey string) {
	if me.m == nil {
		return value, false, ""
	}
	node := me.m.trie.Match(key)
	if node == nil {
		return value, false, ""
	}
	value, _ = node.Data.(T)
	return value, true, node.Key
}

// Remove removes the entry for key. Returns false if no entry with that exact
// key exists.
func (me MapString_[T]) Remove(key string) (succeeded bool) {
	if me.m == nil {
		panic("cannot modify an uninitialized MapString_")
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

// MapString returns an immutable snapshot of this MapString_. Due to the COW
// nature of the underlying structure, this is effectively a pointer copy.
func (me MapString_[T]) MapString() MapString[T] {
	if me.m == nil {
		return MapString[T]{}
	}
	return MapString[T]{*me.m}
}

// MapString is an immutable string-keyed map. For a mutable equivalent, see
// MapString_.
//
// The zero value is an empty map.
type MapString[T any] struct {
	x mapStringX
}

// MapString_ returns a mutable map initialized with the contents of this one.
// Due to COW, this is effectively a pointer copy.
func (me MapString[T]) MapString_() MapString_[T] {
	if me.x.eq == nil {
		me.x.eq = defaultComparator
	}
	return MapString_[T]{&me.x}
}

// Build is a convenience method for scoped mutations. It passes a mutable
// clone to the callback; if the callback returns true the result is returned as
// a new immutable snapshot, otherwise the original is returned unchanged.
func (me MapString[T]) Build(builder func(MapString_[T]) bool) MapString[T] {
	m_ := me.MapString_()
	if builder(m_) {
		return m_.MapString()
	}
	return me
}

// NumEntries returns the number of entries in the map.
func (me MapString[T]) NumEntries() int64 {
	return me.x.trie.NumNodes()
}

// Get returns the value associated with key using an exact match.
func (me MapString[T]) Get(key string) (t T, ok bool) {
	node := me.x.trie.Match(key)
	if node == nil || node.Key != key || !node.isActive {
		return t, false
	}
	t, _ = node.Data.(T)
	return t, true
}

// LongestMatch returns the value for the longest stored key that is a string
// prefix of the given key.
func (me MapString[T]) LongestMatch(key string) (value T, found bool, matchKey string) {
	node := me.x.trie.Match(key)
	if node == nil {
		return value, false, ""
	}
	value, _ = node.Data.(T)
	return value, true, node.Key
}

// Walk invokes callback for each entry in lexicographical order. Stops early
// and returns false if callback returns false.
func (me MapString[T]) Walk(callback func(string, T) bool) bool {
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
func (me MapString[T]) Diff(
	other MapString[T],
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

// Map returns a new MapString with the same keys but values transformed by
// mapper. The original is not modified.
func (me MapString[T]) Map(mapper func(string, T) T) MapString[T] {
	if mapper == nil {
		return me
	}
	return MapString[T]{
		mapStringX{
			me.x.trie.Map(func(key string, value interface{}) interface{} {
				t, _ := value.(T)
				return mapper(key, t)
			}, me.x.eq),
			me.x.eq,
		},
	}
}

// mapStringX is the internal immutable container.
type mapStringX struct {
	trie *trieNode
	eq   comparator
}

func defaultComparator(a, b interface{}) bool {
	return a == b
}

func newMapStringX() *mapStringX {
	return &mapStringX{nil, defaultComparator}
}

func newMapStringXCustomCompare(eq comparator) *mapStringX {
	return &mapStringX{nil, eq}
}

func (me *mapStringX) mutate(mutator func() *trieNode) {
	me.trie = mutator()
}
