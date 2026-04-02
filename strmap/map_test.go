//go:build go1.18
// +build go1.18

package strmap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ieq is a simple interface equality comparator for tests.
func ieq(a, b interface{}) bool {
	return a == b
}

// --- Map_ (mutable) tests ---

func TestMapTGetOnlyExactMatch(t *testing.T) {
	m := NewMap_[int]()
	m.Insert("foo", 3)
	assert.Equal(t, int64(1), m.NumEntries())

	// "foobar" starts with "foo" but is not an exact match
	_, ok := m.Get("foobar")
	assert.False(t, ok)
}

func TestMapTGetNotFound(t *testing.T) {
	m := NewMap_[int]()
	succeeded := m.Insert("foo", 3)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.NumEntries())

	value, ok := m.Get("foo")
	assert.True(t, ok)
	assert.Equal(t, 3, value)

	_, ok = m.Get("bar")
	assert.False(t, ok)
}

func TestMapTRemove(t *testing.T) {
	m := NewMap_[int]()
	succeeded := m.Insert("foo", 3)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.NumEntries())

	succeeded = m.Remove("foo")
	assert.True(t, succeeded)
	assert.Equal(t, int64(0), m.NumEntries())
}

func TestMapTRemoveNotFound(t *testing.T) {
	m := NewMap_[int]()
	succeeded := m.Insert("foo", 3)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.NumEntries())

	succeeded = m.Remove("bar")
	assert.False(t, succeeded)
	assert.Equal(t, int64(1), m.NumEntries())
}

func TestMapTRemoveNotExact(t *testing.T) {
	m := NewMap_[int]()
	m.Insert("foo", 3)
	assert.Equal(t, int64(1), m.NumEntries())

	// "foobar" is not stored; removing it should fail
	succeeded := m.Remove("foobar")
	assert.False(t, succeeded)
	assert.Equal(t, int64(1), m.NumEntries())
}

func TestMapTInsert(t *testing.T) {
	m := NewMap_[int]()
	succeeded := m.Insert("foo", 3)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.NumEntries())

	data, ok := m.Get("foo")
	assert.True(t, ok)
	assert.Equal(t, 3, data)

	_, ok = m.Get("bar")
	assert.False(t, ok)
}

func TestMapTInsertDuplicate(t *testing.T) {
	m := NewMap_[int]()
	succeeded := m.Insert("foo", 3)
	assert.True(t, succeeded)

	succeeded = m.Insert("foo", 99)
	assert.False(t, succeeded)

	data, ok := m.Get("foo")
	assert.True(t, ok)
	assert.Equal(t, 3, data)
}

func TestMapTUpdate(t *testing.T) {
	m := NewMap_[bool]()
	assert.Equal(t, int64(0), m.m.trie.NumNodes())

	m.Insert("foo", false)

	succeeded := m.Update("foo", true)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.m.trie.NumNodes())
	value, found := m.Get("foo")
	assert.True(t, found)
	assert.True(t, value)

	succeeded = m.Update("foo", false)
	assert.True(t, succeeded)
	value, found = m.Get("foo")
	assert.True(t, found)
	assert.False(t, value)
	assert.True(t, m.m.trie.isValid())
}

func TestMapTUpdateNotFound(t *testing.T) {
	m := NewMap_[int]()
	succeeded := m.Update("foo", 3)
	assert.False(t, succeeded)
	assert.Equal(t, int64(0), m.NumEntries())
}

func TestMapTRemovePrefix(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := NewMap_[bool]()
		m.Insert("foo", true)

		succeeded := m.Remove("foo")
		assert.True(t, succeeded)
		assert.Equal(t, int64(0), m.m.trie.NumNodes())
		assert.True(t, m.m.trie.isValid())
	})

	t.Run("Not Found", func(t *testing.T) {
		m := NewMap_[bool]()
		m.Insert("foo", true)

		succeeded := m.Remove("bar")
		assert.False(t, succeeded)
		assert.Equal(t, int64(1), m.m.trie.NumNodes())
		assert.True(t, m.m.trie.isValid())
	})

	t.Run("Not Exact", func(t *testing.T) {
		m := NewMap_[bool]()
		m.Insert("foo", true)

		// "foobar" is not stored, so removing it should fail
		succeeded := m.Remove("foobar")
		assert.False(t, succeeded)
		assert.Equal(t, int64(1), m.m.trie.NumNodes())
		assert.True(t, m.m.trie.isValid())
	})
}

func TestMapTWalk(t *testing.T) {
	m := NewMap_[bool]()
	m.Insert("foo", true)

	found := false
	m.Map().Walk(func(key string, value bool) bool {
		assert.Equal(t, "foo", key)
		assert.True(t, value)
		found = true
		return true
	})
	assert.True(t, found)
	assert.True(t, m.m.trie.isValid())
}

func TestMapTEqual(t *testing.T) {
	a := NewMap_[bool]()
	b := NewMap_[bool]()

	assert.True(t, a.m.trie.Equal(b.m.trie, ieq))
	assert.True(t, b.m.trie.Equal(a.m.trie, ieq))

	a.Insert("foo", true)
	assert.False(t, a.m.trie.Equal(b.m.trie, ieq))
	assert.False(t, b.m.trie.Equal(a.m.trie, ieq))

	b.Insert("bar", true)
	assert.False(t, a.m.trie.Equal(b.m.trie, ieq))
	assert.False(t, b.m.trie.Equal(a.m.trie, ieq))
}

// --- Walk ordering ---

func TestMapWalkOrder(t *testing.T) {
	// Walk should visit shorter (broader) keys before longer (more-specific) ones,
	// and disjoint keys in lexicographical order.
	m := Map[bool]{}.Map_()
	m.Insert("foobar", true)
	m.Insert("foobaz", true)
	m.Insert("foo", true)

	var result []string
	m.Map().Walk(func(key string, _ bool) bool {
		result = append(result, key)
		return true
	})
	assert.Equal(t, []string{"foo", "foobar", "foobaz"}, result)
}

func TestMapWalkDisjoint(t *testing.T) {
	m := Map[bool]{}.Map_()
	m.Insert("bbb", true)
	m.Insert("aaa", true)
	m.Insert("ccc", true)

	var result []string
	m.Map().Walk(func(key string, _ bool) bool {
		result = append(result, key)
		return true
	})
	assert.Equal(t, []string{"aaa", "bbb", "ccc"}, result)
}

func TestMapWalkEarlyStop(t *testing.T) {
	m := NewMap_[bool]()
	m.Insert("aaa", true)
	m.Insert("bbb", true)
	m.Insert("ccc", true)

	iterations := 0
	cont := m.Map().Walk(func(string, bool) bool {
		iterations++
		return false
	})
	assert.Equal(t, 1, iterations)
	assert.False(t, cont)
}

// --- Immutable snapshot ---

func TestMapImmutableSnapshot(t *testing.T) {
	m := NewMap_[int]()
	m.Insert("foo", 1)

	snap := m.Map()
	m.Insert("bar", 2)

	m2 := snap.Map_()
	m2.Insert("baz", 3)

	var found bool
	var value int

	// m has foo and bar
	value, found = m.Get("foo")
	assert.True(t, found)
	assert.Equal(t, 1, value)
	value, found = m.Get("bar")
	assert.True(t, found)
	assert.Equal(t, 2, value)
	_, found = m.Get("baz")
	assert.False(t, found)

	// snap only has foo
	assert.Equal(t, int64(1), snap.NumEntries())
	value, found = snap.Get("foo")
	assert.True(t, found)
	assert.Equal(t, 1, value)
	_, found = snap.Get("bar")
	assert.False(t, found)

	// m2 has foo and baz
	assert.Equal(t, int64(2), m2.NumEntries())
	value, found = m2.Get("foo")
	assert.True(t, found)
	assert.Equal(t, 1, value)
	_, found = m2.Get("bar")
	assert.False(t, found)
	value, found = m2.Get("baz")
	assert.True(t, found)
	assert.Equal(t, 3, value)
}

func TestMapBuild(t *testing.T) {
	base := Map[int]{}.Build(func(m Map_[int]) bool {
		m.Insert("foo", 1)
		m.Insert("bar", 2)
		return true
	})

	// Build with return false leaves base unchanged
	result := base.Build(func(m Map_[int]) bool {
		m.Insert("baz", 3)
		return false
	})

	assert.Equal(t, int64(2), result.NumEntries())

	// Build with return true applies changes
	result2 := base.Build(func(m Map_[int]) bool {
		m.Insert("baz", 3)
		return true
	})
	assert.Equal(t, int64(3), result2.NumEntries())
	v, ok := result2.Get("baz")
	assert.True(t, ok)
	assert.Equal(t, 3, v)
}

// --- Nil / zero value safety ---

func TestNilMap(t *testing.T) {
	var m Map_[bool]

	assert.Equal(t, int64(0), m.NumEntries())
	assert.Equal(t, int64(0), m.Map().NumEntries())
	_, found := m.Get("foo")
	assert.False(t, found)

	// Walk on zero-value immutable map
	assert.True(t, m.Map().Walk(func(string, bool) bool {
		panic("should not be called")
	}))

	testPanic := func(run func()) {
		var panicked bool
		func() {
			defer func() {
				if r := recover(); r != nil {
					panicked = true
				}
			}()
			run()
		}()
		assert.True(t, panicked)
	}

	t.Run("insert panics", func(t *testing.T) {
		testPanic(func() { m.Insert("foo", false) })
	})
	t.Run("update panics", func(t *testing.T) {
		testPanic(func() { m.Update("foo", false) })
	})
	t.Run("remove panics", func(t *testing.T) {
		testPanic(func() { m.Remove("foo") })
	})
}

// --- Diff ---

func TestMapDiff(t *testing.T) {
	a := Map[bool]{}.Build(func(m Map_[bool]) bool {
		m.Insert("foo", true)
		m.Insert("foobar", true)
		m.Insert("foobaz", true)
		return true
	})

	b := Map[bool]{}.Build(func(m Map_[bool]) bool {
		m.Insert("foo", true)
		m.Insert("foobar", false) // changed
		m.Insert("qux", true)    // only in b
		return true
	})

	type action struct {
		key          string
		before, after bool
	}

	var actions []action
	getHandlers := func() (left, right func(string, bool) bool, changed func(string, bool, bool) bool) {
		actions = nil
		left = func(k string, v bool) bool {
			actions = append(actions, action{k, v, false})
			return true
		}
		right = func(k string, v bool) bool {
			actions = append(actions, action{k, false, v})
			return true
		}
		changed = func(k string, l, r bool) bool {
			actions = append(actions, action{k, l, r})
			return true
		}
		return
	}

	t.Run("forward", func(t *testing.T) {
		lf, rf, cf := getHandlers()
		a.Diff(b, cf, lf, rf, nil)
		assert.Equal(t, []action{
			{"foobar", true, false}, // changed
			{"foobaz", true, false}, // only in a
			{"qux", false, true},    // only in b
		}, actions)
	})

	t.Run("backward", func(t *testing.T) {
		lf, rf, cf := getHandlers()
		b.Diff(a, cf, lf, rf, nil)
		assert.Equal(t, []action{
			{"foobar", false, true}, // changed (reversed)
			{"foobaz", false, true}, // only in a (now on right)
			{"qux", true, false},    // only in b (now on left)
		}, actions)
	})
}

// --- Large / structural validity ---

func TestMapManyEntriesValid(t *testing.T) {
	keys := []string{
		"alpha", "beta", "gamma", "delta", "epsilon",
		"foo", "foobar", "foobaz", "fooqax",
		"zoo", "zookeeper",
	}

	m := NewMap_[int]()
	for i, k := range keys {
		succeeded := m.Insert(k, i)
		assert.True(t, succeeded)
	}
	assert.Equal(t, int64(len(keys)), m.NumEntries())
	assert.True(t, m.m.trie.isValid())

	// All keys retrievable
	for i, k := range keys {
		v, ok := m.Get(k)
		assert.True(t, ok)
		assert.Equal(t, i, v)
	}

	// Remove half, check still valid
	for _, k := range keys[:len(keys)/2] {
		succeeded := m.Remove(k)
		assert.True(t, succeeded)
	}
	assert.True(t, m.m.trie.isValid())
}
