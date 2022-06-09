//go:build go1.18
// +build go1.18

package ipv6

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTableTInsertOrUpdate(t *testing.T) {
	m := Table[int]{}.Table_()
	m.Insert(_a("2001::"), 0)
	m.InsertOrUpdate(_a("2001::"), 3)
	assert.Equal(t, int64(1), m.NumEntries())

	data, ok := m.Get(_a("2001::"))
	assert.True(t, ok)
	assert.Equal(t, 3, data)
}

func TestTableTInsertOrUpdateDuplicate(t *testing.T) {
	m := NewTable_[int]()
	m.InsertOrUpdate(_a("2001::"), 3)
	assert.Equal(t, int64(1), m.NumEntries())
	data, ok := m.Get(_a("2001::"))
	assert.True(t, ok)
	assert.Equal(t, 3, data)

	m.InsertOrUpdate(_a("2001::"), 4)
	assert.Equal(t, int64(1), m.NumEntries())
	data, ok = m.Get(_a("2001::"))
	assert.True(t, ok)
	assert.Equal(t, 4, data)
}

func TestTableTGetOnlyExactMatch(t *testing.T) {
	m := NewTable_[int]()
	m.Insert(_p("2001::/112"), 3)
	assert.Equal(t, int64(1), m.NumEntries())

	_, ok := m.Get(_a("2001::1"))
	assert.False(t, ok)
}

func TestTableTGetNotFound(t *testing.T) {
	m := NewTable_[int]()
	succeeded := m.Insert(_a("2001::1"), 3)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.NumEntries())

	_, ok := m.Get(_a("2001:0:1::1"))
	assert.False(t, ok)
}

func TestTableTGetOrInsertOnlyExactMatch(t *testing.T) {
	m := NewTable_[int]()
	m.Insert(_p("2001::/112"), 3)
	assert.Equal(t, int64(1), m.NumEntries())

	value := m.GetOrInsert(_a("2001::1"), 5)
	assert.Equal(t, 5, value)
	assert.Equal(t, int64(2), m.NumEntries())
}

func TestTableTGetOrInsertNotFound(t *testing.T) {
	m := NewTable_[int]()
	succeeded := m.Insert(_a("2001::1"), 3)
	assert.True(t, succeeded)

	value := m.GetOrInsert(_a("2001:0:1::1"), 5)
	assert.Equal(t, 5, value)
	assert.Equal(t, int64(2), m.NumEntries())
}

func TestTableTGetOrInsertPrefixOnlyExactMatch(t *testing.T) {
	m := NewTable_[int]()
	m.Insert(_p("2001::/112"), 3)
	assert.Equal(t, int64(1), m.NumEntries())

	value := m.GetOrInsert(_p("2001::2/127"), 5)
	assert.Equal(t, 5, value)
	assert.Equal(t, int64(2), m.NumEntries())
}

func TestTableTGetOrInsertPrefixNotFound(t *testing.T) {
	m := NewTable_[int]()
	succeeded := m.Insert(_a("2001::1"), 3)
	assert.True(t, succeeded)

	value := m.GetOrInsert(_p("2001:0:1::2/127"), 5)
	assert.Equal(t, 5, value)
	assert.Equal(t, int64(2), m.NumEntries())
}

func TestTableTMatchLongestPrefixMatch(t *testing.T) {
	m := NewTable_[int]()
	m.Insert(_p("2001::/112"), 3)
	assert.Equal(t, int64(1), m.NumEntries())
	m.Insert(_p("2001::/96"), 4)
	assert.Equal(t, int64(2), m.NumEntries())

	data, matched, n := m.LongestMatch(_a("2001::1"))
	assert.True(t, matched)
	assert.Equal(t, _p("2001::/112"), n)
	assert.Equal(t, 3, data)
}

func TestMTableatchNotFound(t *testing.T) {
	m := NewTable_[int]()
	succeeded := m.Insert(_a("2001::1"), 3)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.NumEntries())

	_, matched, _ := m.LongestMatch(_a("2001:0:1::1"))
	assert.False(t, matched)
}

func TestTableTRemove(t *testing.T) {
	m := NewTable_[int]()
	succeeded := m.Insert(_a("2001::1"), 3)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.NumEntries())

	m.Remove(_a("2001::1"))
	assert.Equal(t, int64(0), m.NumEntries())
}

func TestTableTRemoveNotFound(t *testing.T) {
	m := NewTable_[int]()
	succeeded := m.Insert(_a("2001::1"), 3)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.NumEntries())

	m.Remove(_a("2001:0:1::1"))
	assert.Equal(t, int64(1), m.NumEntries())
}

func TestTableTInsert(t *testing.T) {
	m := NewTable_[int]()
	succeeded := m.Insert(_p("2001::/112"), 3)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.NumEntries())

	data, ok := m.Get(_p("2001::/112"))
	assert.True(t, ok)
	assert.Equal(t, 3, data)

	_, ok = m.Get(_p("2001:0:1::/112"))
	assert.False(t, ok)
}

func TestTableTInsertOrUpdatePrefix(t *testing.T) {
	m := NewTable_[int]()
	m.Insert(_p("2001::/112"), 0)
	m.InsertOrUpdate(_p("2001::/112"), 3)
	assert.Equal(t, int64(1), m.NumEntries())

	data, ok := m.Get(_p("2001::/112"))
	assert.True(t, ok)
	assert.Equal(t, 3, data)

	_, ok = m.Get(_p("2001:0:1::/112"))
	assert.False(t, ok)
}

func TestTableTRemovePrefixNotFound(t *testing.T) {
	m := NewTable_[int]()
	succeeded := m.Insert(_p("2001::/112"), 3)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.NumEntries())

	m.Remove(_p("2001:0:1::/112"))
	assert.Equal(t, int64(1), m.NumEntries())
}

func TestTableTMatchPrefixLongestPrefixMatch(t *testing.T) {
	m := NewTable_[int]()
	m.Insert(_p("2001::/112"), 3)
	assert.Equal(t, int64(1), m.NumEntries())
	m.Insert(_p("2001::/96"), 4)
	assert.Equal(t, int64(2), m.NumEntries())

	prefix := _p("2001::/118")
	data, matched, n := m.LongestMatch(prefix)
	assert.NotEqual(t, prefix, n)
	assert.True(t, matched)
	assert.Equal(t, 3, data)
	assert.Equal(t, _p("2001::/112"), n)
}

func TestTableTMatchPrefixNotFound(t *testing.T) {
	m := NewTable_[int]()
	m.Insert(_p("2001::/112"), 3)
	assert.Equal(t, int64(1), m.NumEntries())

	_, matched, _ := m.LongestMatch(_p("2001:0:1::/112"))
	assert.False(t, matched)
}

func TestTableTExample1(t *testing.T) {
	m := Table[bool]{}.Table_()
	m.Insert(_p("2001::2/127"), true)
	m.Insert(_p("2001::1/128"), true)
	m.Insert(_p("2001::/128"), true)

	var result []string
	m.Table().Walk(func(prefix Prefix, value bool) bool {
		result = append(result, prefix.String())
		return true
	})
	assert.Equal(
		t,
		[]string{
			"2001::/128",
			"2001::1/128",
			"2001::2/127",
		},
		result,
	)

	result = []string{}
	m.Table().Aggregate().Walk(func(prefix Prefix, value bool) bool {
		result = append(result, prefix.String())
		return true
	})
	assert.Equal(
		t,
		[]string{
			"2001::/126",
		},
		result,
	)
}

type pairT struct {
	prefix string
	value  bool
}

func TestTableTExample2(t *testing.T) {
	m := NewTable_[bool]()
	m.Insert(_p("2001::/126"), true)
	m.Insert(_p("2001::/127"), false)
	m.Insert(_p("2001::1/128"), true)
	m.Insert(_p("2001::/128"), false)

	var result []pairT
	m.Table().Walk(func(prefix Prefix, value bool) bool {
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
			pairT{prefix: "2001::/126", value: true},
			pairT{prefix: "2001::/127", value: false},
			pairT{prefix: "2001::/128", value: false},
			pairT{prefix: "2001::1/128", value: true},
		},
		result,
	)

	result = []pairT{}
	m.Table().Aggregate().Walk(func(prefix Prefix, value bool) bool {
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
			pairT{prefix: "2001::/126", value: true},
			pairT{prefix: "2001::/127", value: false},
			pairT{prefix: "2001::1/128", value: true},
		},
		result,
	)
}

func TestTableTExample3(t *testing.T) {
	m := NewTable_[bool]()
	m.Insert(_p("172:21::/40"), false)
	m.Insert(_p("192:68:27::/49"), false)
	m.Insert(_p("192:168:26:8000::/49"), false)
	m.Insert(_p("10:224:24::/64"), false)
	m.Insert(_p("192:68:24::/48"), false)
	m.Insert(_p("172:16::/24"), false)
	m.Insert(_p("192:68:26::/48"), false)
	m.Insert(_p("10:224:24::/60"), false)
	m.Insert(_p("192:168:24::/48"), false)
	m.Insert(_p("192:168:25::/48"), false)
	m.Insert(_p("192:168:26::/49"), false)
	m.Insert(_p("192:68:25::/48"), false)
	m.Insert(_p("192:168:27::/48"), false)
	m.Insert(_p("172:20:8000::/38"), false)
	m.Insert(_p("192:68:27:8000::/49"), false)

	var result []string
	m.Table().Walk(func(prefix Prefix, value bool) bool {
		result = append(result, prefix.String())
		return true
	})
	assert.Equal(
		t,
		[]string{
			"10:224:24::/60",
			"10:224:24::/64",
			"172:16::/24",
			"172:20:8000::/38",
			"172:21::/40",
			"192:68:24::/48",
			"192:68:25::/48",
			"192:68:26::/48",
			"192:68:27::/49",
			"192:68:27:8000::/49",
			"192:168:24::/48",
			"192:168:25::/48",
			"192:168:26::/49",
			"192:168:26:8000::/49",
			"192:168:27::/48",
		},
		result,
	)
	iterations := 0
	m.Table().Walk(func(prefix Prefix, value bool) bool {
		iterations++
		return false
	})
	assert.Equal(t, 1, iterations)

	result = []string{}
	m.Table().Aggregate().Walk(func(prefix Prefix, value bool) bool {
		result = append(result, prefix.String())
		return true
	})
	assert.Equal(
		t,
		[]string{
			"10:224:24::/60",
			"172:16::/24",
			"192:68:24::/46",
			"192:168:24::/46",
		},
		result,
	)
	iterations = 0
	m.Table().Aggregate().Walk(func(prefix Prefix, value bool) bool {
		iterations++
		return false
	})
	assert.Equal(t, 1, iterations)
}

func TestTableTnsert(t *testing.T) {
	m := NewTable_[bool]()
	assert.Equal(t, int64(0), m.t.m.trie.NumNodes())

	key := Prefix{Address{uint128{0x2001, 0x0180000}}, 112}
	succeeded := m.Insert(key, true)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.t.m.trie.NumNodes())
	assert.True(t, m.t.m.trie.isValid())
}

func TestTableTnsertOrUpdate(t *testing.T) {
	m := NewTable_[bool]()
	assert.Equal(t, int64(0), m.t.m.trie.NumNodes())

	key := Prefix{Address{uint128{0x2001, 0x0180000}}, 112}
	m.InsertOrUpdate(key, true)
	assert.Equal(t, int64(1), m.t.m.trie.NumNodes())
	value, match, matchedKey := m.LongestMatch(key)
	assert.Equal(t, key, matchedKey)
	assert.True(t, match)
	assert.Equal(t, key, matchedKey)
	assert.True(t, value)

	m.InsertOrUpdate(key, false)
	assert.Equal(t, int64(1), m.t.m.trie.NumNodes())
	value, match, matchedKey = m.LongestMatch(key)
	assert.Equal(t, key, matchedKey)
	assert.True(t, match)
	assert.Equal(t, key, matchedKey)
	assert.False(t, value)
	assert.True(t, m.t.m.trie.isValid())
}

func TestTableTUpdate(t *testing.T) {
	m := NewTable_[bool]()
	assert.Equal(t, int64(0), m.t.m.trie.NumNodes())

	key := Prefix{Address{uint128{0x2001, 0x0180000}}, 112}
	m.Insert(key, false)

	succeeded := m.Update(key, true)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.t.m.trie.NumNodes())
	value, match, matchedKey := m.LongestMatch(key)
	assert.Equal(t, key, matchedKey)
	assert.True(t, match)
	assert.Equal(t, key, matchedKey)
	assert.True(t, value)

	succeeded = m.Update(key, false)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.t.m.trie.NumNodes())
	value, match, matchedKey = m.LongestMatch(key)
	assert.Equal(t, key, matchedKey)
	assert.True(t, match)
	assert.Equal(t, key, matchedKey)
	assert.False(t, value)
	assert.True(t, m.t.m.trie.isValid())
}

func TestTableTGetOrInsert(t *testing.T) {
	m := NewTable_[bool]()
	assert.Equal(t, int64(0), m.t.m.trie.NumNodes())

	key := Prefix{Address{uint128{0x2001, 0x0180000}}, 112}
	value := m.GetOrInsert(key, true)
	assert.True(t, value)
	assert.Equal(t, int64(1), m.t.m.trie.NumNodes())
	assert.True(t, m.t.m.trie.isValid())
}

func TestTableTMatch(t *testing.T) {
	m := NewTable_[bool]()

	insertKey := Prefix{Address{uint128{0x2001, 0x0180000}}, 112}
	m.Insert(insertKey, true)

	t.Run("None", func(t *testing.T) {
		_, found, _ := m.LongestMatch(Prefix{Address{uint128{0x2001, 0x0100000}}, 112})
		assert.False(t, found)
		assert.True(t, m.t.m.trie.isValid())
	})

	t.Run("Exact", func(t *testing.T) {
		prefix := Prefix{Address{uint128{0x2001, 0x0180000}}, 112}
		value, found, key := m.LongestMatch(prefix)
		assert.Equal(t, prefix, key)
		assert.True(t, found)
		assert.Equal(t, insertKey, key)
		assert.True(t, value)
		assert.True(t, m.t.m.trie.isValid())
	})

	t.Run("Contains", func(t *testing.T) {
		prefix := Prefix{Address{uint128{0x2001, 0x0180017}}, 128}
		value, found, key := m.LongestMatch(prefix)
		assert.NotEqual(t, prefix, key)
		assert.True(t, found)
		assert.Equal(t, insertKey, key)
		assert.True(t, value)
		assert.True(t, m.t.m.trie.isValid())
	})
}

func TestTableTRemovePrefix(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := NewTable_[bool]()

		insertKey := Prefix{Address{uint128{0x2001, 0x0180000}}, 112}
		m.Insert(insertKey, true)

		key := Prefix{Address{uint128{0x2001, 0x0180000}}, 112}
		succeeded := m.Remove(key)
		assert.True(t, succeeded)
		assert.Equal(t, int64(0), m.t.m.trie.NumNodes())
		assert.True(t, m.t.m.trie.isValid())
	})

	t.Run("Not Found", func(t *testing.T) {
		m := NewTable_[bool]()

		insertKey := Prefix{Address{uint128{0x2001, 0x0180000}}, 112}
		m.Insert(insertKey, true)

		key := Prefix{Address{uint128{0x2001, 0x0100000}}, 112}
		succeeded := m.Remove(key)
		assert.False(t, succeeded)
		assert.Equal(t, int64(1), m.t.m.trie.NumNodes())
		assert.True(t, m.t.m.trie.isValid())
	})

	t.Run("Not Exact", func(t *testing.T) {
		m := NewTable_[bool]()

		insertKey := Prefix{Address{uint128{0x2001, 0x0180000}}, 112}
		m.Insert(insertKey, true)

		key := Prefix{Address{uint128{0x2001, 0x0180017}}, 128}
		succeeded := m.Remove(key)
		assert.False(t, succeeded)
		assert.Equal(t, int64(1), m.t.m.trie.NumNodes())
		assert.True(t, m.t.m.trie.isValid())
	})
}

func TestTableTWalk(t *testing.T) {
	m := NewTable_[bool]()

	insertKey := Prefix{Address{uint128{0x2001, 0x0180000}}, 112}
	m.Insert(insertKey, true)

	found := false
	m.Table().Walk(func(key Prefix, value bool) bool {
		assert.Equal(t, insertKey, key)
		assert.True(t, value)
		found = true
		return true
	})
	assert.True(t, found)
	assert.True(t, m.t.m.trie.isValid())
}

func TestTableTWalkAggregates(t *testing.T) {
	m := NewTable_[bool]()

	insertKey := Prefix{Address{uint128{0x2001, 0x0180000}}, 112}
	m.Insert(insertKey, true)

	secondKey := Prefix{Address{uint128{0x2001, 0x0180017}}, 128}
	m.Insert(secondKey, true)

	found := false
	m.Table().Aggregate().Walk(func(key Prefix, value bool) bool {
		assert.Equal(t, insertKey, key)
		assert.True(t, value)
		found = true
		return true
	})
	assert.True(t, found)
	assert.True(t, m.t.m.trie.isValid())
}

func TestTableTEqual(t *testing.T) {
	a := NewTable_[bool]()
	b := NewTable_[bool]()

	assert.True(t, a.t.m.trie.Equal(b.t.m.trie, ieq))
	assert.True(t, b.t.m.trie.Equal(a.t.m.trie, ieq))

	a.Insert(Prefix{Address{uint128{0x2001, 0x0180001}}, 112}, true)
	assert.False(t, a.t.m.trie.Equal(b.t.m.trie, ieq))
	assert.False(t, b.t.m.trie.Equal(a.t.m.trie, ieq))

	b.Insert(Prefix{Address{uint128{0x2001, 0x0180000}}, 112}, true)
	assert.False(t, a.t.m.trie.Equal(b.t.m.trie, ieq))
	assert.False(t, b.t.m.trie.Equal(a.t.m.trie, ieq))
}

// Test that Tables, when passed and copied, refer to the same data
func TestTableTAsReferenceType(t *testing.T) {
	m := NewTable_[int]()

	manipulate := func(m Table_[int]) {
		m.Insert(_a("2001::1"), 0)
		m.InsertOrUpdate(_a("2001::1"), 3)
	}
	manipulate(m)
	assert.Equal(t, int64(1), m.NumEntries())
	data, ok := m.Get(_a("2001::1"))
	assert.True(t, ok)
	assert.Equal(t, 3, data)
}

func TestTableTConcurrentModification(t *testing.T) {
	m := NewTable_[bool]()

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
		m.t.mutate(func() (bool, *trieNode) {
			ch <- true

			newHead, _ := m.t.m.trie.Insert(_p("2001::/112"), nil)
			return true, newHead
		})
	}()
	go func() {
		defer wrap()
		m.t.mutate(func() (bool, *trieNode) {
			<-ch
			newHead, _ := m.t.m.trie.Insert(_p("2001::1:0/112"), nil)
			return true, newHead
		})
	}()
	wg.Wait()
	assert.Equal(t, 1, panicked)
}

func TestNilTable(t *testing.T) {
	var table Table_[bool]

	// On-offs
	assert.Equal(t, int64(0), table.NumEntries())
	assert.Equal(t, int64(0), table.Table().NumEntries())
	_, found := table.Get(_a("2001::1234:0"))
	assert.False(t, found)
	_, matched, _ := table.LongestMatch(_a("2001::1234:0"))
	assert.False(t, matched)

	// Walk
	assert.True(t, table.Table().Walk(func(Prefix, bool) bool {
		panic("should not be called")
	}))
	assert.True(t, table.Table().Aggregate().Walk(func(Prefix, bool) bool {
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
			table.Insert(_a("2001::1234:0"), false)
		})
	})
	t.Run("update panics", func(t *testing.T) {
		testPanic(func() {
			table.Update(_a("2001::1234:0"), false)
		})
	})
	t.Run("insert or update panics", func(t *testing.T) {
		testPanic(func() {
			table.InsertOrUpdate(_a("2001::1234:0"), false)
		})
	})
	t.Run("get or insert panics", func(t *testing.T) {
		testPanic(func() {
			table.GetOrInsert(_a("2001::1234:0"), false)
		})
	})
	t.Run("remove panics", func(t *testing.T) {
		testPanic(func() {
			table.Remove(_a("2001::1234:0"))
		})
	})
}

func TestTableInsertNil(t *testing.T) {
	m := NewTable_[int]()
	succeeded := m.Insert(nil, 3)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.NumEntries())
	value, found := m.Get(_p("::/0"))
	assert.True(t, found)
	assert.Equal(t, 3, value)
}

func TestTableUpdateNil(t *testing.T) {
	m := NewTable_[int]()
	m.Insert(nil, 10)

	succeeded := m.Update(nil, 3)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.NumEntries())
	value, found := m.Get(_p("::/0"))
	assert.True(t, found)
	assert.Equal(t, 3, value)
}

func TestTableRemoveNil(t *testing.T) {
	m := NewTable_[int]()
	m.Insert(_p("::/0"), 10)

	succeeded := m.Remove(nil)
	assert.True(t, succeeded)

	_, found := m.Get(_p("::/0"))
	assert.False(t, found)
}

func TestTableLongestMatch(t *testing.T) {
	m := NewTable_[int]()
	m.Insert(_p("::/0"), 10)

	value, matched, prefix := m.LongestMatch(nil)
	assert.Equal(t, 10, value)
	assert.Equal(t, prefix, Prefix{})
	assert.True(t, matched)
	assert.Equal(t, _p("::/0"), prefix)
}

func TestTableInsertOrUpdateNil(t *testing.T) {
	m := NewTable_[int]()
	m.InsertOrUpdate(nil, 3)

	assert.Equal(t, int64(1), m.NumEntries())
	value, found := m.Get(_p("::/0"))
	assert.True(t, found)
	assert.Equal(t, 3, value)
}

func TestTableGetOrInsertNil(t *testing.T) {
	m := NewTable_[int]()
	result := m.GetOrInsert(nil, 11)
	assert.Equal(t, 11, result)

	value, found := m.Get(_p("::/0"))
	assert.True(t, found)
	assert.Equal(t, 11, value)
}

func TestTableTDiff(t *testing.T) {
	a := Table[bool]{}.Build(func(a_ Table_[bool]) bool {
		a_.Insert(_p("2001::1234:0/115"), true)
		a_.Insert(_p("2001::1234:6400/115"), true)
		a_.Insert(_p("2001::1234:0/113"), true)
		return true
	})

	a = a.Build(func(a_ Table_[bool]) bool {
		a_.Insert(_p("1900::1234:0/113"), true)
		return false
	})

	b := Table[bool]{}.Build(func(b_ Table_[bool]) bool {
		b_.Insert(_p("2001::1234:0/115"), true)
		b_.Insert(_p("2001::1234:9600/115"), true)
		b_.Insert(_p("2001::1234:0/113"), false)
		return true
	})

	type action struct {
		prefix        Prefix
		before, after bool
	}

	var actions []action
	getHandlers := func() (left, right func(Prefix, bool) bool, changed func(p Prefix, left, right bool) bool) {
		actions = nil
		left = func(p Prefix, v bool) bool {
			actions = append(actions, action{p, v, false})
			return true
		}
		right = func(p Prefix, v bool) bool {
			actions = append(actions, action{p, false, v})
			return true
		}
		changed = func(p Prefix, l, r bool) bool {
			actions = append(actions, action{p, l, r})
			return true
		}
		return
	}

	t.Run("forward", func(t *testing.T) {
		left, right, changed := getHandlers()
		a.Diff(b, changed, left, right, nil)
		assert.Equal(t, []action{
			action{_p("2001::1234:0/113"), true, false},
			action{_p("2001::1234:6400/115"), true, false},
			action{_p("2001::1234:9600/115"), false, true},
		}, actions)
	})

	t.Run("backward", func(t *testing.T) {
		left, right, changed := getHandlers()
		b.Diff(a, changed, left, right, nil)
		assert.Equal(t, []action{
			action{_p("2001::1234:0/113"), false, true},
			action{_p("2001::1234:6400/115"), false, true},
			action{_p("2001::1234:9600/115"), true, false},
		}, actions)
	})
}

func TestFixedTableT(t *testing.T) {
	addrOne := _a("2001::1")
	addrTwo := _a("2001::2")
	addrThree := _a("2001::3")

	m := NewTable_[int]()
	succeeded := m.Insert(addrOne, 1)
	assert.True(t, succeeded)

	im := m.Table()
	succeeded = m.Insert(addrTwo, 2)
	assert.True(t, succeeded)

	m2 := im.Table_()
	succeeded = m2.Insert(addrThree, 3)
	assert.True(t, succeeded)

	var found bool
	var value int

	value, found = m.Get(addrOne)
	assert.True(t, found)
	assert.Equal(t, 1, value)
	value, found = m.Get(addrTwo)
	assert.True(t, found)
	assert.Equal(t, 2, value)
	_, found = m.Get(addrThree)
	assert.False(t, found)

	assert.Equal(t, int64(1), im.NumEntries())
	value, found = im.Get(addrOne)
	assert.True(t, found)
	assert.Equal(t, 1, value)
	_, found = im.Get(addrTwo)
	assert.False(t, found)
	_, found = im.Get(addrThree)
	assert.False(t, found)

	assert.Equal(t, int64(2), m2.NumEntries())
	value, found = m2.Get(addrOne)
	assert.True(t, found)
	assert.Equal(t, 1, value)
	_, found = m2.Get(addrTwo)
	assert.False(t, found)
	value, found = m2.Get(addrThree)
	assert.True(t, found)
	assert.Equal(t, 3, value)
}

func TestTableMap(t *testing.T) {
	var a Table[bool]
	assert.Equal(t, a, a.Map(nil))
	assert.Equal(t, a, a.Map(func(Prefix, bool) bool {
		panic("this should not be run")
	}))

	a = func() Table[bool] {
		a := Table[bool]{}.Table_()
		a.Insert(_p("2001::1234:0/115"), true)
		a.Insert(_p("2001::1234:6400/115"), true)
		a.Insert(_p("2001::1234:0/113"), true)
		return a.Table()
	}()

	result := a.Map(func(Prefix, bool) bool {
		return false
	})

	assert.Equal(t, int64(3), result.NumEntries())

	value, ok := result.Get(_p("2001::1234:0/115"))
	assert.True(t, ok)
	assert.False(t, value)
	value, ok = result.Get(_p("2001::1234:6400/115"))
	assert.True(t, ok)
	assert.False(t, value)
	value, ok = result.Get(_p("2001::1234:0/113"))
	assert.True(t, ok)
	assert.False(t, value)

	_, ok = result.Get(_p("::/0"))
	assert.False(t, ok)
}

func TestTableVariousComparators(t *testing.T) {
	tests := []struct {
		description string
		table       Table_[creativeComparable]
		expected    []string
	}{
		{
			description: "comparable",
			table:       Table[creativeComparable]{}.Table_(),
			expected: []string{
				"2001::1234:0/118",
				"2001::1234:0/122",
				"2001::1234:0/124",
				"2001::1234:0/126",
				"2001::1234:0/128",
			},
		}, {
			description: "not_comparable",
			table: NewTableCustomCompare_[creativeComparable](
				func(a, b creativeComparable) bool {
					return false
				},
			),
			expected: []string{
				"2001::1234:0/118",
				"2001::1234:0/120",
				"2001::1234:0/122",
				"2001::1234:0/124",
				"2001::1234:0/126",
				"2001::1234:0/128",
			},
		}, {
			description: "equal_comparable",
			table: NewTableCustomCompare_[creativeComparable](
				func(a, b creativeComparable) bool {
					return a.Equal(b)
				},
			),
			expected: []string{
				"2001::1234:0/118",
				"2001::1234:0/126",
				"2001::1234:0/128",
			},
		}, {
			description: "custom_comparable",
			table: NewTableCustomCompare_[creativeComparable](func(a, b creativeComparable) bool {
				return a.i <= 3 && b.i <= 3
			}),
			expected: []string{
				"2001::1234:0/118",
				"2001::1234:0/128",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			tt.table.Insert(_p("2001::1234:0/118"), creativeComparable{0})
			tt.table.Insert(_p("2001::1234:0/120"), creativeComparable{0})
			tt.table.Insert(_p("2001::1234:0/122"), creativeComparable{1})
			tt.table.Insert(_p("2001::1234:0/124"), creativeComparable{2})
			tt.table.Insert(_p("2001::1234:0/126"), creativeComparable{3})
			tt.table.Insert(_p("2001::1234:0/128"), creativeComparable{4})

			result := []string{}
			tt.table.Table().Aggregate().Walk(func(prefix Prefix, value creativeComparable) bool {
				result = append(result, prefix.String())
				return true
			})
			assert.Equal(t, tt.expected, result)
		})
	}
}
