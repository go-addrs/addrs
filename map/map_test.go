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

// --- MapString_ (mutable) tests ---

func TestMapStringTInsertOrUpdate(t *testing.T) {
	m := MapString[int]{}.MapString_()
	m.Insert("foo", 0)
	m.InsertOrUpdate("foo", 3)
	assert.Equal(t, int64(1), m.NumEntries())

	data, ok := m.Get("foo")
	assert.True(t, ok)
	assert.Equal(t, 3, data)
}

func TestMapStringTInsertOrUpdateDuplicate(t *testing.T) {
	m := NewMapString_[int]()
	m.InsertOrUpdate("foo", 3)
	assert.Equal(t, int64(1), m.NumEntries())
	data, ok := m.Get("foo")
	assert.True(t, ok)
	assert.Equal(t, 3, data)

	m.InsertOrUpdate("foo", 4)
	assert.Equal(t, int64(1), m.NumEntries())
	data, ok = m.Get("foo")
	assert.True(t, ok)
	assert.Equal(t, 4, data)
}

func TestMapStringTGetOnlyExactMatch(t *testing.T) {
	m := NewMapString_[int]()
	m.Insert("foo", 3)
	assert.Equal(t, int64(1), m.NumEntries())

	// "foobar" starts with "foo" but is not an exact match
	_, ok := m.Get("foobar")
	assert.False(t, ok)
}

func TestMapStringTGetNotFound(t *testing.T) {
	m := NewMapString_[int]()
	succeeded := m.Insert("foo", 3)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.NumEntries())

	_, ok := m.Get("bar")
	assert.False(t, ok)
}

func TestMapStringTGetOrInsertOnlyExactMatch(t *testing.T) {
	m := NewMapString_[int]()
	m.Insert("foo", 3)
	assert.Equal(t, int64(1), m.NumEntries())

	// "foobar" is not in the map; should insert and return the default
	value := m.GetOrInsert("foobar", 5)
	assert.Equal(t, 5, value)
	assert.Equal(t, int64(2), m.NumEntries())
}

func TestMapStringTGetOrInsertNotFound(t *testing.T) {
	m := NewMapString_[int]()
	succeeded := m.Insert("foo", 3)
	assert.True(t, succeeded)

	value := m.GetOrInsert("bar", 5)
	assert.Equal(t, 5, value)
	assert.Equal(t, int64(2), m.NumEntries())
}

func TestMapStringTGetOrInsertExists(t *testing.T) {
	m := NewMapString_[int]()
	m.Insert("foo", 3)
	assert.Equal(t, int64(1), m.NumEntries())

	value := m.GetOrInsert("foo", 99)
	assert.Equal(t, 3, value)
	assert.Equal(t, int64(1), m.NumEntries())
}

func TestMapStringTLongestMatch(t *testing.T) {
	m := NewMapString_[int]()
	m.Insert("foo", 3)
	assert.Equal(t, int64(1), m.NumEntries())
	m.Insert("foobar", 4)
	assert.Equal(t, int64(2), m.NumEntries())

	// "foobarbaz" matches "foobar" (longest prefix)
	data, matched, key := m.LongestMatch("foobarbaz")
	assert.True(t, matched)
	assert.Equal(t, "foobar", key)
	assert.Equal(t, 4, data)
}

func TestMapStringTLongestMatchShorter(t *testing.T) {
	m := NewMapString_[int]()
	m.Insert("foo", 3)
	assert.Equal(t, int64(1), m.NumEntries())
	m.Insert("foobar", 4)
	assert.Equal(t, int64(2), m.NumEntries())

	// "foox" matches "foo" (only "foo" is a prefix of "foox")
	data, matched, key := m.LongestMatch("foox")
	assert.True(t, matched)
	assert.Equal(t, "foo", key)
	assert.Equal(t, 3, data)
}

func TestMapStringTLongestMatchNotFound(t *testing.T) {
	m := NewMapString_[int]()
	succeeded := m.Insert("foo", 3)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.NumEntries())

	_, matched, _ := m.LongestMatch("bar")
	assert.False(t, matched)
}

func TestMapStringTLongestMatchExact(t *testing.T) {
	m := NewMapString_[int]()
	m.Insert("foo", 3)

	data, matched, key := m.LongestMatch("foo")
	assert.True(t, matched)
	assert.Equal(t, "foo", key)
	assert.Equal(t, 3, data)
}

func TestMapStringTRemove(t *testing.T) {
	m := NewMapString_[int]()
	succeeded := m.Insert("foo", 3)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.NumEntries())

	succeeded = m.Remove("foo")
	assert.True(t, succeeded)
	assert.Equal(t, int64(0), m.NumEntries())
}

func TestMapStringTRemoveNotFound(t *testing.T) {
	m := NewMapString_[int]()
	succeeded := m.Insert("foo", 3)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.NumEntries())

	succeeded = m.Remove("bar")
	assert.False(t, succeeded)
	assert.Equal(t, int64(1), m.NumEntries())
}

func TestMapStringTRemoveNotExact(t *testing.T) {
	m := NewMapString_[int]()
	m.Insert("foo", 3)
	assert.Equal(t, int64(1), m.NumEntries())

	// "foobar" is not stored; removing it should fail
	succeeded := m.Remove("foobar")
	assert.False(t, succeeded)
	assert.Equal(t, int64(1), m.NumEntries())
}

func TestMapStringTInsert(t *testing.T) {
	m := NewMapString_[int]()
	succeeded := m.Insert("foo", 3)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.NumEntries())

	data, ok := m.Get("foo")
	assert.True(t, ok)
	assert.Equal(t, 3, data)

	_, ok = m.Get("bar")
	assert.False(t, ok)
}

func TestMapStringTInsertDuplicate(t *testing.T) {
	m := NewMapString_[int]()
	succeeded := m.Insert("foo", 3)
	assert.True(t, succeeded)

	succeeded = m.Insert("foo", 99)
	assert.False(t, succeeded)

	data, ok := m.Get("foo")
	assert.True(t, ok)
	assert.Equal(t, 3, data)
}

func TestMapStringTUpdate(t *testing.T) {
	m := NewMapString_[bool]()
	assert.Equal(t, int64(0), m.m.trie.NumNodes())

	m.Insert("foo", false)

	succeeded := m.Update("foo", true)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.m.trie.NumNodes())
	value, found, key := m.LongestMatch("foo")
	assert.True(t, found)
	assert.Equal(t, "foo", key)
	assert.True(t, value)

	succeeded = m.Update("foo", false)
	assert.True(t, succeeded)
	value, found, key = m.LongestMatch("foo")
	assert.True(t, found)
	assert.Equal(t, "foo", key)
	assert.False(t, value)
	assert.True(t, m.m.trie.isValid())
}

func TestMapStringTUpdateNotFound(t *testing.T) {
	m := NewMapString_[int]()
	succeeded := m.Update("foo", 3)
	assert.False(t, succeeded)
	assert.Equal(t, int64(0), m.NumEntries())
}

func TestMapStringTInsertOrUpdateIsValid(t *testing.T) {
	m := NewMapString_[bool]()
	assert.Equal(t, int64(0), m.m.trie.NumNodes())

	m.InsertOrUpdate("foo", true)
	assert.Equal(t, int64(1), m.m.trie.NumNodes())
	value, match, key := m.LongestMatch("foo")
	assert.True(t, match)
	assert.Equal(t, "foo", key)
	assert.True(t, value)

	m.InsertOrUpdate("foo", false)
	assert.Equal(t, int64(1), m.m.trie.NumNodes())
	value, match, key = m.LongestMatch("foo")
	assert.True(t, match)
	assert.Equal(t, "foo", key)
	assert.False(t, value)
	assert.True(t, m.m.trie.isValid())
}

func TestMapStringTGetOrInsertIsValid(t *testing.T) {
	m := NewMapString_[bool]()
	assert.Equal(t, int64(0), m.m.trie.NumNodes())

	value := m.GetOrInsert("foo", true)
	assert.True(t, value)
	assert.Equal(t, int64(1), m.m.trie.NumNodes())
	assert.True(t, m.m.trie.isValid())
}

func TestMapStringTMatch(t *testing.T) {
	m := NewMapString_[bool]()
	m.Insert("foo", true)

	t.Run("None", func(t *testing.T) {
		_, found, _ := m.LongestMatch("bar")
		assert.False(t, found)
		assert.True(t, m.m.trie.isValid())
	})

	t.Run("Exact", func(t *testing.T) {
		value, found, key := m.LongestMatch("foo")
		assert.True(t, found)
		assert.Equal(t, "foo", key)
		assert.True(t, value)
		assert.True(t, m.m.trie.isValid())
	})

	t.Run("Contains", func(t *testing.T) {
		// "foo" is a prefix of "foobar", so LongestMatch("foobar") should return "foo"
		value, found, key := m.LongestMatch("foobar")
		assert.NotEqual(t, "foobar", key)
		assert.True(t, found)
		assert.Equal(t, "foo", key)
		assert.True(t, value)
		assert.True(t, m.m.trie.isValid())
	})
}

func TestMapStringTRemovePrefix(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := NewMapString_[bool]()
		m.Insert("foo", true)

		succeeded := m.Remove("foo")
		assert.True(t, succeeded)
		assert.Equal(t, int64(0), m.m.trie.NumNodes())
		assert.True(t, m.m.trie.isValid())
	})

	t.Run("Not Found", func(t *testing.T) {
		m := NewMapString_[bool]()
		m.Insert("foo", true)

		succeeded := m.Remove("bar")
		assert.False(t, succeeded)
		assert.Equal(t, int64(1), m.m.trie.NumNodes())
		assert.True(t, m.m.trie.isValid())
	})

	t.Run("Not Exact", func(t *testing.T) {
		m := NewMapString_[bool]()
		m.Insert("foo", true)

		// "foobar" is not stored, so removing it should fail
		succeeded := m.Remove("foobar")
		assert.False(t, succeeded)
		assert.Equal(t, int64(1), m.m.trie.NumNodes())
		assert.True(t, m.m.trie.isValid())
	})
}

func TestMapStringTWalk(t *testing.T) {
	m := NewMapString_[bool]()
	m.Insert("foo", true)

	found := false
	m.MapString().Walk(func(key string, value bool) bool {
		assert.Equal(t, "foo", key)
		assert.True(t, value)
		found = true
		return true
	})
	assert.True(t, found)
	assert.True(t, m.m.trie.isValid())
}

func TestMapStringTEqual(t *testing.T) {
	a := NewMapString_[bool]()
	b := NewMapString_[bool]()

	assert.True(t, a.m.trie.Equal(b.m.trie, ieq))
	assert.True(t, b.m.trie.Equal(a.m.trie, ieq))

	a.Insert("foo", true)
	assert.False(t, a.m.trie.Equal(b.m.trie, ieq))
	assert.False(t, b.m.trie.Equal(a.m.trie, ieq))

	b.Insert("bar", true)
	assert.False(t, a.m.trie.Equal(b.m.trie, ieq))
	assert.False(t, b.m.trie.Equal(a.m.trie, ieq))
}

// Test that MapString_, when passed by value, refers to the same underlying data.
func TestMapStringTAsReferenceType(t *testing.T) {
	m := NewMapString_[int]()

	manipulate := func(m MapString_[int]) {
		m.Insert("foo", 0)
		m.InsertOrUpdate("foo", 3)
	}
	manipulate(m)
	assert.Equal(t, int64(1), m.NumEntries())
	data, ok := m.Get("foo")
	assert.True(t, ok)
	assert.Equal(t, 3, data)
}

// --- String-specific prefix semantics ---

func TestMapStringEmptyKeyMatchesAll(t *testing.T) {
	m := NewMapString_[int]()
	m.Insert("", 1)
	assert.Equal(t, int64(1), m.NumEntries())

	// "" is a prefix of everything
	value, found, key := m.LongestMatch("anything")
	assert.True(t, found)
	assert.Equal(t, "", key)
	assert.Equal(t, 1, value)

	// exact match works too
	v, ok := m.Get("")
	assert.True(t, ok)
	assert.Equal(t, 1, v)
}

func TestMapStringPrefixHierarchy(t *testing.T) {
	m := NewMapString_[int]()
	m.Insert("", 1)
	m.Insert("foo", 2)
	m.Insert("foobar", 3)

	// exact matches
	v, ok := m.Get("")
	assert.True(t, ok)
	assert.Equal(t, 1, v)
	v, ok = m.Get("foo")
	assert.True(t, ok)
	assert.Equal(t, 2, v)
	v, ok = m.Get("foobar")
	assert.True(t, ok)
	assert.Equal(t, 3, v)

	// longest match returns the most specific stored prefix
	value, found, key := m.LongestMatch("foobarbaz")
	assert.True(t, found)
	assert.Equal(t, "foobar", key)
	assert.Equal(t, 3, value)

	value, found, key = m.LongestMatch("foox")
	assert.True(t, found)
	assert.Equal(t, "foo", key)
	assert.Equal(t, 2, value)

	value, found, key = m.LongestMatch("baz")
	assert.True(t, found)
	assert.Equal(t, "", key)
	assert.Equal(t, 1, value)
}

// --- Walk ordering ---

func TestMapStringWalkOrder(t *testing.T) {
	// Walk should visit shorter (broader) keys before longer (more-specific) ones,
	// and disjoint keys in lexicographical order.
	m := MapString[bool]{}.MapString_()
	m.Insert("foobar", true)
	m.Insert("foobaz", true)
	m.Insert("foo", true)

	var result []string
	m.MapString().Walk(func(key string, _ bool) bool {
		result = append(result, key)
		return true
	})
	assert.Equal(t, []string{"foo", "foobar", "foobaz"}, result)
}

func TestMapStringWalkDisjoint(t *testing.T) {
	m := MapString[bool]{}.MapString_()
	m.Insert("bbb", true)
	m.Insert("aaa", true)
	m.Insert("ccc", true)

	var result []string
	m.MapString().Walk(func(key string, _ bool) bool {
		result = append(result, key)
		return true
	})
	assert.Equal(t, []string{"aaa", "bbb", "ccc"}, result)
}

func TestMapStringWalkEarlyStop(t *testing.T) {
	m := NewMapString_[bool]()
	m.Insert("aaa", true)
	m.Insert("bbb", true)
	m.Insert("ccc", true)

	iterations := 0
	cont := m.MapString().Walk(func(string, bool) bool {
		iterations++
		return false
	})
	assert.Equal(t, 1, iterations)
	assert.False(t, cont)
}

// --- Immutable snapshot ---

func TestMapStringImmutableSnapshot(t *testing.T) {
	m := NewMapString_[int]()
	m.Insert("foo", 1)

	snap := m.MapString()
	m.Insert("bar", 2)

	m2 := snap.MapString_()
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

func TestMapStringBuild(t *testing.T) {
	base := MapString[int]{}.Build(func(m MapString_[int]) bool {
		m.Insert("foo", 1)
		m.Insert("bar", 2)
		return true
	})

	// Build with return false leaves base unchanged
	result := base.Build(func(m MapString_[int]) bool {
		m.Insert("baz", 3)
		return false
	})
	assert.Equal(t, base, result)
	assert.Equal(t, int64(2), result.NumEntries())

	// Build with return true applies changes
	result2 := base.Build(func(m MapString_[int]) bool {
		m.Insert("baz", 3)
		return true
	})
	assert.Equal(t, int64(3), result2.NumEntries())
	v, ok := result2.Get("baz")
	assert.True(t, ok)
	assert.Equal(t, 3, v)
}

// --- Nil / zero value safety ---

func TestNilMapString(t *testing.T) {
	var m MapString_[bool]

	assert.Equal(t, int64(0), m.NumEntries())
	assert.Equal(t, int64(0), m.MapString().NumEntries())
	_, found := m.Get("foo")
	assert.False(t, found)
	_, matched, _ := m.LongestMatch("foo")
	assert.False(t, matched)

	// Walk on zero-value immutable map
	assert.True(t, m.MapString().Walk(func(string, bool) bool {
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
	t.Run("insert or update panics", func(t *testing.T) {
		testPanic(func() { m.InsertOrUpdate("foo", false) })
	})
	t.Run("get or insert panics", func(t *testing.T) {
		testPanic(func() { m.GetOrInsert("foo", false) })
	})
	t.Run("remove panics", func(t *testing.T) {
		testPanic(func() { m.Remove("foo") })
	})
}

// --- Diff ---

func TestMapStringDiff(t *testing.T) {
	a := MapString[bool]{}.Build(func(m MapString_[bool]) bool {
		m.Insert("foo", true)
		m.Insert("foobar", true)
		m.Insert("foobaz", true)
		return true
	})

	b := MapString[bool]{}.Build(func(m MapString_[bool]) bool {
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

// --- Map ---

func TestMapStringMap(t *testing.T) {
	var a MapString[bool]
	assert.Equal(t, a, a.Map(nil))
	assert.Equal(t, a, a.Map(func(string, bool) bool {
		panic("should not be called")
	}))

	a = MapString[bool]{}.Build(func(m MapString_[bool]) bool {
		m.Insert("foo", true)
		m.Insert("bar", true)
		m.Insert("baz", true)
		return true
	})

	result := a.Map(func(_ string, _ bool) bool {
		return false
	})

	assert.Equal(t, int64(3), result.NumEntries())

	for _, key := range []string{"foo", "bar", "baz"} {
		value, ok := result.Get(key)
		assert.True(t, ok)
		assert.False(t, value)
	}

	_, ok := result.Get("qux")
	assert.False(t, ok)
}

// --- Custom comparator ---

func TestMapStringCustomComparator(t *testing.T) {
	// Use a custom comparator that treats all positive ints as equal
	m := NewMapStringCustomCompare_[int](func(a, b int) bool {
		return (a > 0) == (b > 0)
	})

	m.Insert("foo", 1)
	// InsertOrUpdate with same-class value: should be a no-op (value kept as original)
	m.InsertOrUpdate("foo", 2)

	value, ok := m.Get("foo")
	assert.True(t, ok)
	assert.Equal(t, 1, value) // original retained because comparator says equal

	// Different class (negative) should update
	m.InsertOrUpdate("foo", -1)
	value, ok = m.Get("foo")
	assert.True(t, ok)
	assert.Equal(t, -1, value)
}

// --- Large / structural validity ---

func TestMapStringManyEntriesValid(t *testing.T) {
	keys := []string{
		"alpha", "beta", "gamma", "delta", "epsilon",
		"foo", "foobar", "foobaz", "fooqax",
		"zoo", "zookeeper",
	}

	m := NewMapString_[int]()
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
