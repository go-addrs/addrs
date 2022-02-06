//go:build go1.18
// +build go1.18

package ipv4

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTableTInsertOrUpdate(t *testing.T) {
	m := NewTable[int]()
	m.Insert(_a("10.224.24.1"), 0)
	m.InsertOrUpdate(_a("10.224.24.1"), 3)
	assert.Equal(t, int64(1), m.Size())

	data, ok := m.Get(_a("10.224.24.1"))
	assert.True(t, ok)
	assert.Equal(t, 3, data)
}

func TestTableTInsertOrUpdateDuplicate(t *testing.T) {
	m := NewTable[int]()
	m.InsertOrUpdate(_a("10.224.24.1"), 3)
	assert.Equal(t, int64(1), m.Size())
	data, ok := m.Get(_a("10.224.24.1"))
	assert.True(t, ok)
	assert.Equal(t, 3, data)

	m.InsertOrUpdate(_a("10.224.24.1"), 4)
	assert.Equal(t, int64(1), m.Size())
	data, ok = m.Get(_a("10.224.24.1"))
	assert.True(t, ok)
	assert.Equal(t, 4, data)
}

func TestTableTGetOnlyExactMatch(t *testing.T) {
	m := NewTable[int]()
	m.Insert(_p("10.224.24.0/24"), 3)
	assert.Equal(t, int64(1), m.Size())

	_, ok := m.Get(_a("10.224.24.1"))
	assert.False(t, ok)
}

func TestTableTGetNotFound(t *testing.T) {
	m := NewTable[int]()
	succeeded := m.Insert(_a("10.224.24.1"), 3)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.Size())

	_, ok := m.Get(_a("10.225.24.1"))
	assert.False(t, ok)
}

func TestTableTGetOrInsertOnlyExactMatch(t *testing.T) {
	m := NewTable[int]()
	m.Insert(_p("10.224.24.0/24"), 3)
	assert.Equal(t, int64(1), m.Size())

	value := m.GetOrInsert(_a("10.224.24.1"), 5)
	assert.Equal(t, 5, value)
	assert.Equal(t, int64(2), m.Size())
}

func TestTableTGetOrInsertNotFound(t *testing.T) {
	m := NewTable[int]()
	succeeded := m.Insert(_a("10.224.24.1"), 3)
	assert.True(t, succeeded)

	value := m.GetOrInsert(_a("10.225.24.1"), 5)
	assert.Equal(t, 5, value)
	assert.Equal(t, int64(2), m.Size())
}

func TestTableTGetOrInsertPrefixOnlyExactMatch(t *testing.T) {
	m := NewTable[int]()
	m.Insert(_p("10.224.24.0/24"), 3)
	assert.Equal(t, int64(1), m.Size())

	value := m.GetOrInsert(_p("10.224.24.2/31"), 5)
	assert.Equal(t, 5, value)
	assert.Equal(t, int64(2), m.Size())
}

func TestTableTGetOrInsertPrefixNotFound(t *testing.T) {
	m := NewTable[int]()
	succeeded := m.Insert(_a("10.224.24.1"), 3)
	assert.True(t, succeeded)

	value := m.GetOrInsert(_p("10.225.24.2/31"), 5)
	assert.Equal(t, 5, value)
	assert.Equal(t, int64(2), m.Size())
}

func TestTableTMatchLongestPrefixMatch(t *testing.T) {
	m := NewTable[int]()
	m.Insert(_p("10.224.24.0/24"), 3)
	assert.Equal(t, int64(1), m.Size())
	m.Insert(_p("10.224.0.0/16"), 4)
	assert.Equal(t, int64(2), m.Size())

	data, matched, n := m.LongestMatch(_a("10.224.24.1"))
	assert.Equal(t, MatchContains, matched)
	assert.Equal(t, _p("10.224.24.0/24"), n)
	assert.Equal(t, 3, data)
}

func TestMTableatchNotFound(t *testing.T) {
	m := NewTable[int]()
	succeeded := m.Insert(_a("10.224.24.1"), 3)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.Size())

	_, matched, _ := m.LongestMatch(_a("10.225.24.1"))
	assert.Equal(t, MatchNone, matched)
}

func TestTableTRemove(t *testing.T) {
	m := NewTable[int]()
	succeeded := m.Insert(_a("10.224.24.1"), 3)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.Size())

	m.Remove(_a("10.224.24.1"))
	assert.Equal(t, int64(0), m.Size())
}

func TestTableTRemoveNotFound(t *testing.T) {
	m := NewTable[int]()
	succeeded := m.Insert(_a("10.224.24.1"), 3)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.Size())

	m.Remove(_a("10.225.24.1"))
	assert.Equal(t, int64(1), m.Size())
}

func TestTableTInsert(t *testing.T) {
	m := NewTable[int]()
	succeeded := m.Insert(_p("10.224.24.0/24"), 3)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.Size())

	data, ok := m.Get(_p("10.224.24.0/24"))
	assert.True(t, ok)
	assert.Equal(t, 3, data)

	data, ok = m.Get(_p("10.225.24.0/24"))
	assert.False(t, ok)
}

func TestTableTInsertOrUpdatePrefix(t *testing.T) {
	m := NewTable[int]()
	m.Insert(_p("10.224.24.0/24"), 0)
	m.InsertOrUpdate(_p("10.224.24.0/24"), 3)
	assert.Equal(t, int64(1), m.Size())

	data, ok := m.Get(_p("10.224.24.0/24"))
	assert.True(t, ok)
	assert.Equal(t, 3, data)

	data, ok = m.Get(_p("10.225.24.0/24"))
	assert.False(t, ok)
}

func TestTableTRemovePrefixNotFound(t *testing.T) {
	m := NewTable[int]()
	succeeded := m.Insert(_p("10.224.24.0/24"), 3)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.Size())

	m.Remove(_p("10.225.24.0/24"))
	assert.Equal(t, int64(1), m.Size())
}

func TestTableTMatchPrefixLongestPrefixMatch(t *testing.T) {
	m := NewTable[int]()
	m.Insert(_p("10.224.24.0/24"), 3)
	assert.Equal(t, int64(1), m.Size())
	m.Insert(_p("10.224.0.0/16"), 4)
	assert.Equal(t, int64(2), m.Size())

	data, matched, n := m.LongestMatch(_p("10.224.24.0/27"))
	assert.Equal(t, MatchContains, matched)
	assert.Equal(t, 3, data)
	assert.Equal(t, _p("10.224.24.0/24"), n)
}

func TestTableTMatchPrefixNotFound(t *testing.T) {
	m := NewTable[int]()
	m.Insert(_p("10.224.24.0/24"), 3)
	assert.Equal(t, int64(1), m.Size())

	_, matched, _ := m.LongestMatch(_p("10.225.24.0/24"))
	assert.Equal(t, MatchNone, matched)
}

func TestTableTExample1(t *testing.T) {
	m := NewTable[bool]()
	m.Insert(_p("10.224.24.2/31"), true)
	m.Insert(_p("10.224.24.1/32"), true)
	m.Insert(_p("10.224.24.0/32"), true)

	var result []string
	m.FixedTable().Walk(func(prefix Prefix, value bool) bool {
		result = append(result, prefix.String())
		return true
	})
	assert.Equal(
		t,
		[]string{
			"10.224.24.0/32",
			"10.224.24.1/32",
			"10.224.24.2/31",
		},
		result,
	)

	result = []string{}
	m.FixedTable().Aggregate().Walk(func(prefix Prefix, value bool) bool {
		result = append(result, prefix.String())
		return true
	})
	assert.Equal(
		t,
		[]string{
			"10.224.24.0/30",
		},
		result,
	)
}

type pairT struct {
	prefix string
	value  bool
}

func TestTableTExample2(t *testing.T) {
	m := NewTable[bool]()
	m.Insert(_p("10.224.24.0/30"), true)
	m.Insert(_p("10.224.24.0/31"), false)
	m.Insert(_p("10.224.24.1/32"), true)
	m.Insert(_p("10.224.24.0/32"), false)

	var result []pairT
	m.FixedTable().Walk(func(prefix Prefix, value bool) bool {
		result = append(
			result,
			pairT{
				prefix: prefix.String(),
				value:  value,
			},
		)
		return true
	})
	assert.Equal(
		t,
		[]pairT{
			pairT{prefix: "10.224.24.0/30", value: true},
			pairT{prefix: "10.224.24.0/31", value: false},
			pairT{prefix: "10.224.24.0/32", value: false},
			pairT{prefix: "10.224.24.1/32", value: true},
		},
		result,
	)

	result = []pairT{}
	m.FixedTable().Aggregate().Walk(func(prefix Prefix, value bool) bool {
		result = append(
			result,
			pairT{
				prefix: prefix.String(),
				value:  value,
			},
		)
		return true
	})
	assert.Equal(
		t,
		[]pairT{
			pairT{prefix: "10.224.24.0/30", value: true},
			pairT{prefix: "10.224.24.0/31", value: false},
			pairT{prefix: "10.224.24.1/32", value: true},
		},
		result,
	)
}

func TestTableTExample3(t *testing.T) {
	m := NewTable[bool]()
	m.Insert(_p("172.21.0.0/20"), false)
	m.Insert(_p("192.68.27.0/25"), false)
	m.Insert(_p("192.168.26.128/25"), false)
	m.Insert(_p("10.224.24.0/32"), false)
	m.Insert(_p("192.68.24.0/24"), false)
	m.Insert(_p("172.16.0.0/12"), false)
	m.Insert(_p("192.68.26.0/24"), false)
	m.Insert(_p("10.224.24.0/30"), false)
	m.Insert(_p("192.168.24.0/24"), false)
	m.Insert(_p("192.168.25.0/24"), false)
	m.Insert(_p("192.168.26.0/25"), false)
	m.Insert(_p("192.68.25.0/24"), false)
	m.Insert(_p("192.168.27.0/24"), false)
	m.Insert(_p("172.20.128.0/19"), false)
	m.Insert(_p("192.68.27.128/25"), false)

	var result []string
	m.FixedTable().Walk(func(prefix Prefix, value bool) bool {
		result = append(result, prefix.String())
		return true
	})
	assert.Equal(
		t,
		[]string{
			"10.224.24.0/30",
			"10.224.24.0/32",
			"172.16.0.0/12",
			"172.20.128.0/19",
			"172.21.0.0/20",
			"192.68.24.0/24",
			"192.68.25.0/24",
			"192.68.26.0/24",
			"192.68.27.0/25",
			"192.68.27.128/25",
			"192.168.24.0/24",
			"192.168.25.0/24",
			"192.168.26.0/25",
			"192.168.26.128/25",
			"192.168.27.0/24",
		},
		result,
	)
	iterations := 0
	m.FixedTable().Walk(func(prefix Prefix, value bool) bool {
		iterations++
		return false
	})
	assert.Equal(t, 1, iterations)

	result = []string{}
	m.FixedTable().Aggregate().Walk(func(prefix Prefix, value bool) bool {
		result = append(result, prefix.String())
		return true
	})
	assert.Equal(
		t,
		[]string{
			"10.224.24.0/30",
			"172.16.0.0/12",
			"192.68.24.0/22",
			"192.168.24.0/22",
		},
		result,
	)
	iterations = 0
	m.FixedTable().Aggregate().Walk(func(prefix Prefix, value bool) bool {
		iterations++
		return false
	})
	assert.Equal(t, 1, iterations)
}

func TestTableTnsert(t *testing.T) {
	m := NewTable[bool]()
	assert.Equal(t, int64(0), m.m.trie.NumNodes())

	key := Prefix{Address{0x0ae01800}, 24}
	succeeded := m.Insert(key, true)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.m.trie.NumNodes())
	assert.True(t, m.m.trie.isValid())
}

func TestTableTnsertOrUpdate(t *testing.T) {
	m := NewTable[bool]()
	assert.Equal(t, int64(0), m.m.trie.NumNodes())

	key := Prefix{Address{0x0ae01800}, 24}
	m.InsertOrUpdate(key, true)
	assert.Equal(t, int64(1), m.m.trie.NumNodes())
	value, match, matchedKey := m.LongestMatch(key)
	assert.Equal(t, MatchExact, match)
	assert.Equal(t, key, matchedKey)
	assert.True(t, value)

	m.InsertOrUpdate(key, false)
	assert.Equal(t, int64(1), m.m.trie.NumNodes())
	value, match, matchedKey = m.LongestMatch(key)
	assert.Equal(t, MatchExact, match)
	assert.Equal(t, key, matchedKey)
	assert.False(t, value)
	assert.True(t, m.m.trie.isValid())
}

func TestTableTUpdate(t *testing.T) {
	m := NewTable[bool]()
	assert.Equal(t, int64(0), m.m.trie.NumNodes())

	key := Prefix{Address{0x0ae01800}, 24}
	m.Insert(key, false)

	succeeded := m.Update(key, true)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.m.trie.NumNodes())
	value, match, matchedKey := m.LongestMatch(key)
	assert.Equal(t, MatchExact, match)
	assert.Equal(t, key, matchedKey)
	assert.True(t, value)

	succeeded = m.Update(key, false)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.m.trie.NumNodes())
	value, match, matchedKey = m.LongestMatch(key)
	assert.Equal(t, MatchExact, match)
	assert.Equal(t, key, matchedKey)
	assert.False(t, value)
	assert.True(t, m.m.trie.isValid())
}

func TestTableTGetOrInsert(t *testing.T) {
	m := NewTable[bool]()
	assert.Equal(t, int64(0), m.m.trie.NumNodes())

	key := Prefix{Address{0x0ae01800}, 24}
	value := m.GetOrInsert(key, true)
	assert.True(t, value)
	assert.Equal(t, int64(1), m.m.trie.NumNodes())
	assert.True(t, m.m.trie.isValid())
}

func TestTableTMatch(t *testing.T) {
	m := NewTable[bool]()

	insertKey := Prefix{Address{0x0ae01800}, 24}
	m.Insert(insertKey, true)

	t.Run("None", func(t *testing.T) {
		_, level, _ := m.LongestMatch(Prefix{Address{0x0ae01000}, 24})
		assert.Equal(t, MatchNone, level)
		assert.True(t, m.m.trie.isValid())
	})

	t.Run("Exact", func(t *testing.T) {
		value, level, key := m.LongestMatch(Prefix{Address{0x0ae01800}, 24})
		assert.Equal(t, MatchExact, level)
		assert.Equal(t, insertKey, key)
		assert.True(t, value)
		assert.True(t, m.m.trie.isValid())
	})

	t.Run("Contains", func(t *testing.T) {
		value, level, key := m.LongestMatch(Prefix{Address{0x0ae01817}, 32})
		assert.Equal(t, MatchContains, level)
		assert.Equal(t, insertKey, key)
		assert.True(t, value)
		assert.True(t, m.m.trie.isValid())
	})
}

func TestTableTRemovePrefix(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := NewTable[bool]()

		insertKey := Prefix{Address{0x0ae01800}, 24}
		m.Insert(insertKey, true)

		key := Prefix{Address{0x0ae01800}, 24}
		succeeded := m.Remove(key)
		assert.True(t, succeeded)
		assert.Equal(t, int64(0), m.m.trie.NumNodes())
		assert.True(t, m.m.trie.isValid())
	})

	t.Run("Not Found", func(t *testing.T) {
		m := NewTable[bool]()

		insertKey := Prefix{Address{0x0ae01800}, 24}
		m.Insert(insertKey, true)

		key := Prefix{Address{0x0ae01000}, 24}
		succeeded := m.Remove(key)
		assert.False(t, succeeded)
		assert.Equal(t, int64(1), m.m.trie.NumNodes())
		assert.True(t, m.m.trie.isValid())
	})

	t.Run("Not Exact", func(t *testing.T) {
		m := NewTable[bool]()

		insertKey := Prefix{Address{0x0ae01800}, 24}
		m.Insert(insertKey, true)

		key := Prefix{Address{0x0ae01817}, 32}
		succeeded := m.Remove(key)
		assert.False(t, succeeded)
		assert.Equal(t, int64(1), m.m.trie.NumNodes())
		assert.True(t, m.m.trie.isValid())
	})
}

func TestTableTWalk(t *testing.T) {
	m := NewTable[bool]()

	insertKey := Prefix{Address{0x0ae01800}, 24}
	m.Insert(insertKey, true)

	found := false
	m.FixedTable().Walk(func(key Prefix, value bool) bool {
		assert.Equal(t, insertKey, key)
		assert.True(t, value)
		found = true
		return true
	})
	assert.True(t, found)
	assert.True(t, m.m.trie.isValid())
}

func TestTableTWalkAggregates(t *testing.T) {
	m := NewTable[bool]()

	insertKey := Prefix{Address{0x0ae01800}, 24}
	m.Insert(insertKey, true)

	secondKey := Prefix{Address{0x0ae01817}, 32}
	m.Insert(secondKey, true)

	found := false
	m.FixedTable().Aggregate().Walk(func(key Prefix, value bool) bool {
		assert.Equal(t, insertKey, key)
		assert.True(t, value)
		found = true
		return true
	})
	assert.True(t, found)
	assert.True(t, m.m.trie.isValid())
}

func TestTableTEqual(t *testing.T) {
	a := NewTable[bool]()
	b := NewTable[bool]()

	assert.True(t, a.m.trie.Equal(b.m.trie))
	assert.True(t, b.m.trie.Equal(a.m.trie))

	a.Insert(Prefix{Address{0x0ae01801}, 24}, true)
	assert.False(t, a.m.trie.Equal(b.m.trie))
	assert.False(t, b.m.trie.Equal(a.m.trie))

	b.Insert(Prefix{Address{0x0ae01800}, 24}, true)
	assert.False(t, a.m.trie.Equal(b.m.trie))
	assert.False(t, b.m.trie.Equal(a.m.trie))
}

// Test that Tables, when passed and copied, refer to the same data
func TestTableTAsReferenceType(t *testing.T) {
	m := NewTable[int]()

	manipulate := func(m Table[int]) {
		m.Insert(_a("10.224.24.1"), 0)
		m.InsertOrUpdate(_a("10.224.24.1"), 3)
	}
	manipulate(m)
	assert.Equal(t, int64(1), m.Size())
	data, ok := m.Get(_a("10.224.24.1"))
	assert.True(t, ok)
	assert.Equal(t, 3, data)
}

func TestTableTConcurrentModification(t *testing.T) {
	m := NewTable[bool]()

	wg := new(sync.WaitGroup)
	wg.Add(2)

	var panicked int
	wrap := func() {
		if r := recover(); r != nil {
			panicked++
		}
		wg.Done()
	}

	// Simulate two goroutines modifying at the same time using a channel to
	// freeze one in the middle and start the other.
	ch := make(chan bool)
	go func() {
		defer wrap()
		(ITable)(m).mutate(func() (bool, *trieNode) {
			ch <- true

			newHead, _ := m.m.trie.Insert(_p("10.0.0.0/24"), nil)
			return true, newHead
		})
	}()
	go func() {
		defer wrap()
		(ITable)(m).mutate(func() (bool, *trieNode) {
			<-ch
			newHead, _ := m.m.trie.Insert(_p("10.0.1.0/24"), nil)
			return true, newHead
		})
	}()
	wg.Wait()
	assert.Equal(t, 1, panicked)
}

func TestNilTable(t *testing.T) {
	var table Table[bool]

	// On-offs
	assert.Equal(t, int64(0), table.Size())
	assert.Equal(t, int64(0), table.FixedTable().Size())
	_, found := table.Get(_a("203.0.113.0"))
	assert.False(t, found)
	_, matched, _ := table.LongestMatch(_a("203.0.113.0"))
	assert.Equal(t, MatchNone, matched)

	// Walk
	assert.True(t, table.FixedTable().Walk(func(Prefix, bool) bool {
		panic("should not be called")
	}))
	assert.True(t, table.FixedTable().Aggregate().Walk(func(Prefix, bool) bool {
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
		testPanic(func() {
			table.Insert(_a("203.0.113.0"), false)
		})
	})
	t.Run("update panics", func(t *testing.T) {
		testPanic(func() {
			table.Update(_a("203.0.113.0"), false)
		})
	})
	t.Run("insert or update panics", func(t *testing.T) {
		testPanic(func() {
			table.InsertOrUpdate(_a("203.0.113.0"), false)
		})
	})
	t.Run("get or insert panics", func(t *testing.T) {
		testPanic(func() {
			table.GetOrInsert(_a("203.0.113.0"), false)
		})
	})
	t.Run("remove panics", func(t *testing.T) {
		testPanic(func() {
			table.Remove(_a("203.0.113.0"))
		})
	})
}

func TestTableInsertNil(t *testing.T) {
	m := NewTable[int]()
	succeeded := m.Insert(nil, 3)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.Size())
	value, found := m.Get(_p("0.0.0.0/0"))
	assert.True(t, found)
	assert.Equal(t, 3, value)
}

func TestTableUpdateNil(t *testing.T) {
	m := NewTable[int]()
	m.Insert(nil, 10)

	succeeded := m.Update(nil, 3)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.Size())
	value, found := m.Get(_p("0.0.0.0/0"))
	assert.True(t, found)
	assert.Equal(t, 3, value)
}

func TestTableRemoveNil(t *testing.T) {
	m := NewTable[int]()
	m.Insert(_p("0.0.0.0/0"), 10)

	succeeded := m.Remove(nil)
	assert.True(t, succeeded)

	_, found := m.Get(_p("0.0.0.0/0"))
	assert.False(t, found)
}

func TestTableLongestMatch(t *testing.T) {
	m := NewTable[int]()
	m.Insert(_p("0.0.0.0/0"), 10)

	value, matched, prefix := m.LongestMatch(nil)
	assert.Equal(t, 10, value)
	assert.Equal(t, MatchExact, matched)
	assert.Equal(t, _p("0.0.0.0/0"), prefix)
}

func TestTableInsertOrUpdateNil(t *testing.T) {
	m := NewTable[int]()
	m.InsertOrUpdate(nil, 3)

	assert.Equal(t, int64(1), m.Size())
	value, found := m.Get(_p("0.0.0.0/0"))
	assert.True(t, found)
	assert.Equal(t, 3, value)
}

func TestTableGetOrInsertNil(t *testing.T) {
	m := NewTable[int]()
	result := m.GetOrInsert(nil, 11)
	assert.Equal(t, 11, result)

	value, found := m.Get(_p("0.0.0.0/0"))
	assert.True(t, found)
	assert.Equal(t, 11, value)
}

func TestTableTDiff(t *testing.T) {
	a := NewTable[bool]()
	a.Insert(_p("203.0.113.0/27"), true)
	a.Insert(_p("203.0.113.64/27"), true)
	a.Insert(_p("203.0.113.0/25"), true)

	b := NewTable[bool]()
	b.Insert(_p("203.0.113.0/27"), true)
	b.Insert(_p("203.0.113.96/27"), true)
	b.Insert(_p("203.0.113.0/25"), false)

	type action struct {
		prefix        Prefix
		before, after bool
	}

	var actions []action
	getHandler := func() DiffHandler[bool] {
		actions = nil
		return DiffHandler[bool]{
			Removed: func(p Prefix, v bool) bool {
				actions = append(actions, action{p, v, false})
				return true
			},
			Added: func(p Prefix, v bool) bool {
				actions = append(actions, action{p, false, v})
				return true
			},
			Modified: func(p Prefix, l, r bool) bool {
				actions = append(actions, action{p, l, r})
				return true
			},
		}
	}

	t.Run("forward", func(t *testing.T) {
		a.FixedTable().Diff(b.FixedTable(), getHandler())
		assert.Equal(t, []action{
			action{_p("203.0.113.0/25"), true, false},
			action{_p("203.0.113.64/27"), true, false},
			action{_p("203.0.113.96/27"), false, true},
		}, actions)
	})

	t.Run("backward", func(t *testing.T) {
		b.FixedTable().Diff(a.FixedTable(), getHandler())
		assert.Equal(t, []action{
			action{_p("203.0.113.0/25"), false, true},
			action{_p("203.0.113.64/27"), false, true},
			action{_p("203.0.113.96/27"), true, false},
		}, actions)
	})
}
