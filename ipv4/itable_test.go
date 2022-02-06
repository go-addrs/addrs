package ipv4

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInsertOrUpdate(t *testing.T) {
	m := NewITable()
	m.Insert(_a("10.224.24.1"), nil)
	m.InsertOrUpdate(_a("10.224.24.1"), 3)
	assert.Equal(t, int64(1), m.Size())

	data, ok := m.Get(_a("10.224.24.1"))
	assert.True(t, ok)
	assert.Equal(t, 3, data)
}

func TestInsertOrUpdateDuplicate(t *testing.T) {
	m := NewITable()
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

func TestGetOnlyExactMatch(t *testing.T) {
	m := NewITable()
	m.Insert(_p("10.224.24.0/24"), 3)
	assert.Equal(t, int64(1), m.Size())

	_, ok := m.Get(_a("10.224.24.1"))
	assert.False(t, ok)
}

func TestGetNotFound(t *testing.T) {
	m := NewITable()
	succeeded := m.Insert(_a("10.224.24.1"), 3)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.Size())

	_, ok := m.Get(_a("10.225.24.1"))
	assert.False(t, ok)
}

func TestGetOrInsertOnlyExactMatch(t *testing.T) {
	m := NewITable()
	m.Insert(_p("10.224.24.0/24"), 3)
	assert.Equal(t, int64(1), m.Size())

	value := m.GetOrInsert(_a("10.224.24.1"), 5)
	assert.Equal(t, 5, value)
	assert.Equal(t, int64(2), m.Size())
}

func TestGetOrInsertNotFound(t *testing.T) {
	m := NewITable()
	succeeded := m.Insert(_a("10.224.24.1"), 3)
	assert.True(t, succeeded)

	value := m.GetOrInsert(_a("10.225.24.1"), 5)
	assert.Equal(t, 5, value)
	assert.Equal(t, int64(2), m.Size())
}

func TestGetOrInsertPrefixOnlyExactMatch(t *testing.T) {
	m := NewITable()
	m.Insert(_p("10.224.24.0/24"), 3)
	assert.Equal(t, int64(1), m.Size())

	value := m.GetOrInsert(_p("10.224.24.2/31"), 5)
	assert.Equal(t, 5, value)
	assert.Equal(t, int64(2), m.Size())
}

func TestGetOrInsertPrefixNotFound(t *testing.T) {
	m := NewITable()
	succeeded := m.Insert(_a("10.224.24.1"), 3)
	assert.True(t, succeeded)

	value := m.GetOrInsert(_p("10.225.24.2/31"), 5)
	assert.Equal(t, 5, value)
	assert.Equal(t, int64(2), m.Size())
}

func TestMatchLongestPrefixMatch(t *testing.T) {
	m := NewITable()
	m.Insert(_p("10.224.24.0/24"), 3)
	assert.Equal(t, int64(1), m.Size())
	m.Insert(_p("10.224.0.0/16"), 4)
	assert.Equal(t, int64(2), m.Size())

	data, matched, n := m.LongestMatch(_a("10.224.24.1"))
	assert.Equal(t, MatchContains, matched)
	assert.Equal(t, _p("10.224.24.0/24"), n)
	assert.Equal(t, 3, data)
}

func TestMatchNotFound(t *testing.T) {
	m := NewITable()
	succeeded := m.Insert(_a("10.224.24.1"), 3)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.Size())

	_, matched, _ := m.LongestMatch(_a("10.225.24.1"))
	assert.Equal(t, MatchNone, matched)
}

func TestRemove(t *testing.T) {
	m := NewITable()
	succeeded := m.Insert(_a("10.224.24.1"), 3)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.Size())

	m.Remove(_a("10.224.24.1"))
	assert.Equal(t, int64(0), m.Size())
}

func TestRemoveNotFound(t *testing.T) {
	m := NewITable()
	succeeded := m.Insert(_a("10.224.24.1"), 3)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.Size())

	m.Remove(_a("10.225.24.1"))
	assert.Equal(t, int64(1), m.Size())
}

func TestInsert(t *testing.T) {
	m := NewITable()
	succeeded := m.Insert(_p("10.224.24.0/24"), 3)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.Size())

	data, ok := m.Get(_p("10.224.24.0/24"))
	assert.True(t, ok)
	assert.Equal(t, 3, data)

	data, ok = m.Get(_p("10.225.24.0/24"))
	assert.False(t, ok)
}

func TestInsertOrUpdatePrefix(t *testing.T) {
	m := NewITable()
	m.Insert(_p("10.224.24.0/24"), nil)
	m.InsertOrUpdate(_p("10.224.24.0/24"), 3)
	assert.Equal(t, int64(1), m.Size())

	data, ok := m.Get(_p("10.224.24.0/24"))
	assert.True(t, ok)
	assert.Equal(t, 3, data)

	data, ok = m.Get(_p("10.225.24.0/24"))
	assert.False(t, ok)
}

func TestRemovePrefix(t *testing.T) {
	m := NewITable()
	succeeded := m.Insert(_p("10.224.24.0/24"), 3)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.Size())

	m.Remove(_p("10.224.24.0/24"))
	assert.Equal(t, int64(0), m.Size())
}

func TestRemovePrefixNotFound(t *testing.T) {
	m := NewITable()
	succeeded := m.Insert(_p("10.224.24.0/24"), 3)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.Size())

	m.Remove(_p("10.225.24.0/24"))
	assert.Equal(t, int64(1), m.Size())
}

func TestMatchPrefixLongestPrefixMatch(t *testing.T) {
	m := NewITable()
	m.Insert(_p("10.224.24.0/24"), 3)
	assert.Equal(t, int64(1), m.Size())
	m.Insert(_p("10.224.0.0/16"), 4)
	assert.Equal(t, int64(2), m.Size())

	data, matched, n := m.LongestMatch(_p("10.224.24.0/27"))
	assert.Equal(t, MatchContains, matched)
	assert.Equal(t, 3, data)
	assert.Equal(t, _p("10.224.24.0/24"), n)
}

func TestMatchPrefixNotFound(t *testing.T) {
	m := NewITable()
	m.Insert(_p("10.224.24.0/24"), 3)
	assert.Equal(t, int64(1), m.Size())

	_, matched, _ := m.LongestMatch(_p("10.225.24.0/24"))
	assert.Equal(t, MatchNone, matched)
}

func TestExample1(t *testing.T) {
	m := NewITable()
	m.Insert(_p("10.224.24.2/31"), true)
	m.Insert(_p("10.224.24.1/32"), true)
	m.Insert(_p("10.224.24.0/32"), true)

	var result []string
	m.FixedTable().Walk(func(prefix Prefix, value interface{}) bool {
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
	m.FixedTable().Aggregate().Walk(func(prefix Prefix, value interface{}) bool {
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

type pair struct {
	prefix string
	value  interface{}
}

func TestExample2(t *testing.T) {
	m := NewITable()
	m.Insert(_p("10.224.24.0/30"), true)
	m.Insert(_p("10.224.24.0/31"), false)
	m.Insert(_p("10.224.24.1/32"), true)
	m.Insert(_p("10.224.24.0/32"), false)

	var result []pair
	m.FixedTable().Walk(func(prefix Prefix, value interface{}) bool {
		result = append(
			result,
			pair{
				prefix: prefix.String(),
				value:  value,
			},
		)
		return true
	})
	assert.Equal(
		t,
		[]pair{
			pair{prefix: "10.224.24.0/30", value: true},
			pair{prefix: "10.224.24.0/31", value: false},
			pair{prefix: "10.224.24.0/32", value: false},
			pair{prefix: "10.224.24.1/32", value: true},
		},
		result,
	)

	result = []pair{}
	m.FixedTable().Aggregate().Walk(func(prefix Prefix, value interface{}) bool {
		result = append(
			result,
			pair{
				prefix: prefix.String(),
				value:  value,
			},
		)
		return true
	})
	assert.Equal(
		t,
		[]pair{
			pair{prefix: "10.224.24.0/30", value: true},
			pair{prefix: "10.224.24.0/31", value: false},
			pair{prefix: "10.224.24.1/32", value: true},
		},
		result,
	)
}

func TestExample3(t *testing.T) {
	m := NewITable()
	m.Insert(_p("172.21.0.0/20"), nil)
	m.Insert(_p("192.68.27.0/25"), nil)
	m.Insert(_p("192.168.26.128/25"), nil)
	m.Insert(_p("10.224.24.0/32"), nil)
	m.Insert(_p("192.68.24.0/24"), nil)
	m.Insert(_p("172.16.0.0/12"), nil)
	m.Insert(_p("192.68.26.0/24"), nil)
	m.Insert(_p("10.224.24.0/30"), nil)
	m.Insert(_p("192.168.24.0/24"), nil)
	m.Insert(_p("192.168.25.0/24"), nil)
	m.Insert(_p("192.168.26.0/25"), nil)
	m.Insert(_p("192.68.25.0/24"), nil)
	m.Insert(_p("192.168.27.0/24"), nil)
	m.Insert(_p("172.20.128.0/19"), nil)
	m.Insert(_p("192.68.27.128/25"), nil)

	var result []string
	m.FixedTable().Walk(func(prefix Prefix, value interface{}) bool {
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
	m.FixedTable().Walk(func(prefix Prefix, value interface{}) bool {
		iterations++
		return false
	})
	assert.Equal(t, 1, iterations)

	result = []string{}
	m.FixedTable().Aggregate().Walk(func(prefix Prefix, value interface{}) bool {
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
	m.FixedTable().Aggregate().Walk(func(prefix Prefix, value interface{}) bool {
		iterations++
		return false
	})
	assert.Equal(t, 1, iterations)
}

func TestITablensert(t *testing.T) {
	m := NewITable()
	assert.Equal(t, int64(0), m.m.trie.NumNodes())

	key := Prefix{Address{0x0ae01800}, 24}
	succeeded := m.Insert(key, true)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.m.trie.NumNodes())
	assert.True(t, m.m.trie.isValid())
}

func TestITablensertOrUpdate(t *testing.T) {
	m := NewITable()
	assert.Equal(t, int64(0), m.m.trie.NumNodes())

	key := Prefix{Address{0x0ae01800}, 24}
	m.InsertOrUpdate(key, true)
	assert.Equal(t, int64(1), m.m.trie.NumNodes())
	value, match, matchedKey := m.LongestMatch(key)
	assert.Equal(t, MatchExact, match)
	assert.Equal(t, key, matchedKey)
	assert.True(t, value.(bool))

	m.InsertOrUpdate(key, false)
	assert.Equal(t, int64(1), m.m.trie.NumNodes())
	value, match, matchedKey = m.LongestMatch(key)
	assert.Equal(t, MatchExact, match)
	assert.Equal(t, key, matchedKey)
	assert.False(t, value.(bool))
	assert.True(t, m.m.trie.isValid())
}

func TestTableUpdate(t *testing.T) {
	m := NewITable()
	assert.Equal(t, int64(0), m.m.trie.NumNodes())

	key := Prefix{Address{0x0ae01800}, 24}
	m.Insert(key, false)

	succeeded := m.Update(key, true)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.m.trie.NumNodes())
	value, match, matchedKey := m.LongestMatch(key)
	assert.Equal(t, MatchExact, match)
	assert.Equal(t, key, matchedKey)
	assert.True(t, value.(bool))

	succeeded = m.Update(key, false)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.m.trie.NumNodes())
	value, match, matchedKey = m.LongestMatch(key)
	assert.Equal(t, MatchExact, match)
	assert.Equal(t, key, matchedKey)
	assert.False(t, value.(bool))
	assert.True(t, m.m.trie.isValid())
}

func TestTableGetOrInsert(t *testing.T) {
	m := NewITable()
	assert.Equal(t, int64(0), m.m.trie.NumNodes())

	key := Prefix{Address{0x0ae01800}, 24}
	value := m.GetOrInsert(key, true)
	assert.True(t, value.(bool))
	assert.Equal(t, int64(1), m.m.trie.NumNodes())
	assert.True(t, m.m.trie.isValid())
}

func TestTableMatch(t *testing.T) {
	m := NewITable()

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
		assert.True(t, value.(bool))
		assert.True(t, m.m.trie.isValid())
	})

	t.Run("Contains", func(t *testing.T) {
		value, level, key := m.LongestMatch(Prefix{Address{0x0ae01817}, 32})
		assert.Equal(t, MatchContains, level)
		assert.Equal(t, insertKey, key)
		assert.True(t, value.(bool))
		assert.True(t, m.m.trie.isValid())
	})
}

func TestTableRemovePrefix(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := NewITable()

		insertKey := Prefix{Address{0x0ae01800}, 24}
		m.Insert(insertKey, true)

		key := Prefix{Address{0x0ae01800}, 24}
		succeeded := m.Remove(key)
		assert.True(t, succeeded)
		assert.Equal(t, int64(0), m.m.trie.NumNodes())
		assert.True(t, m.m.trie.isValid())
	})

	t.Run("Not Found", func(t *testing.T) {
		m := NewITable()

		insertKey := Prefix{Address{0x0ae01800}, 24}
		m.Insert(insertKey, true)

		key := Prefix{Address{0x0ae01000}, 24}
		succeeded := m.Remove(key)
		assert.False(t, succeeded)
		assert.Equal(t, int64(1), m.m.trie.NumNodes())
		assert.True(t, m.m.trie.isValid())
	})

	t.Run("Not Exact", func(t *testing.T) {
		m := NewITable()

		insertKey := Prefix{Address{0x0ae01800}, 24}
		m.Insert(insertKey, true)

		key := Prefix{Address{0x0ae01817}, 32}
		succeeded := m.Remove(key)
		assert.False(t, succeeded)
		assert.Equal(t, int64(1), m.m.trie.NumNodes())
		assert.True(t, m.m.trie.isValid())
	})
}

func TestTableWalk(t *testing.T) {
	m := NewITable()

	insertKey := Prefix{Address{0x0ae01800}, 24}
	m.Insert(insertKey, true)

	found := false
	m.FixedTable().Walk(func(key Prefix, value interface{}) bool {
		assert.Equal(t, insertKey, key)
		assert.True(t, value.(bool))
		found = true
		return true
	})
	assert.True(t, found)
	assert.True(t, m.m.trie.isValid())
}

func TestTableWalkAggregates(t *testing.T) {
	m := NewITable()

	insertKey := Prefix{Address{0x0ae01800}, 24}
	m.Insert(insertKey, true)

	secondKey := Prefix{Address{0x0ae01817}, 32}
	m.Insert(secondKey, true)

	found := false
	m.FixedTable().Aggregate().Walk(func(key Prefix, value interface{}) bool {
		assert.Equal(t, insertKey, key)
		assert.True(t, value.(bool))
		found = true
		return true
	})
	assert.True(t, found)
	assert.True(t, m.m.trie.isValid())
}

func TestTableEqual(t *testing.T) {
	a := NewITable()
	b := NewITable()

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
func TestTableAsReferenceType(t *testing.T) {
	m := NewITable()

	manipulate := func(m ITable) {
		m.Insert(_a("10.224.24.1"), nil)
		m.InsertOrUpdate(_a("10.224.24.1"), 3)
	}
	manipulate(m)
	assert.Equal(t, int64(1), m.Size())
	data, ok := m.Get(_a("10.224.24.1"))
	assert.True(t, ok)
	assert.Equal(t, 3, data)
}

func TestTableConcurrentModification(t *testing.T) {
	m := NewITable()

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
		m.mutate(func() (bool, *trieNode) {
			ch <- true

			newHead, _ := m.m.trie.Insert(_p("10.0.0.0/24"), nil)
			return true, newHead
		})
	}()
	go func() {
		defer wrap()
		m.mutate(func() (bool, *trieNode) {
			<-ch
			newHead, _ := m.m.trie.Insert(_p("10.0.1.0/24"), nil)
			return true, newHead
		})
	}()
	wg.Wait()
	assert.Equal(t, 1, panicked)
}

func TestNilITable(t *testing.T) {
	var table ITable

	// On-offs
	assert.Equal(t, int64(0), table.Size())
	assert.Equal(t, int64(0), table.FixedTable().Size())
	_, found := table.Get(_a("203.0.113.0"))
	assert.False(t, found)
	_, matched, _ := table.LongestMatch(_a("203.0.113.0"))
	assert.Equal(t, MatchNone, matched)

	// Walk
	assert.True(t, table.FixedTable().Walk(func(Prefix, interface{}) bool {
		panic("should not be called")
	}))
	assert.True(t, table.FixedTable().Aggregate().Walk(func(Prefix, interface{}) bool {
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
			table.Insert(_a("203.0.113.0"), nil)
		})
	})
	t.Run("update panics", func(t *testing.T) {
		testPanic(func() {
			table.Update(_a("203.0.113.0"), nil)
		})
	})
	t.Run("insert or update panics", func(t *testing.T) {
		testPanic(func() {
			table.InsertOrUpdate(_a("203.0.113.0"), nil)
		})
	})
	t.Run("get or insert panics", func(t *testing.T) {
		testPanic(func() {
			table.GetOrInsert(_a("203.0.113.0"), nil)
		})
	})
	t.Run("remove panics", func(t *testing.T) {
		testPanic(func() {
			table.Remove(_a("203.0.113.0"))
		})
	})
}

func TestITableInsertNil(t *testing.T) {
	m := NewITable()
	succeeded := m.Insert(nil, 3)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.Size())
	value, found := m.Get(_p("0.0.0.0/0"))
	assert.True(t, found)
	assert.Equal(t, 3, value)
}

func TestITableUpdateNil(t *testing.T) {
	m := NewITable()
	m.Insert(_p("0.0.0.0/0"), 10)

	succeeded := m.Update(nil, 3)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.Size())
	value, found := m.Get(_p("0.0.0.0/0"))
	assert.True(t, found)
	assert.Equal(t, 3, value)
}

func TestITableRemoveNil(t *testing.T) {
	m := NewITable()
	m.Insert(_p("0.0.0.0/0"), 10)

	succeeded := m.Remove(nil)
	assert.True(t, succeeded)

	_, found := m.Get(_p("0.0.0.0/0"))
	assert.False(t, found)
}

func TestITableLongestMatch(t *testing.T) {
	m := NewITable()
	m.Insert(_p("0.0.0.0/0"), 10)

	value, matched, prefix := m.LongestMatch(nil)
	assert.Equal(t, 10, value)
	assert.Equal(t, MatchExact, matched)
	assert.Equal(t, _p("0.0.0.0/0"), prefix)
}

func TestITableInsertOrUpdateNil(t *testing.T) {
	m := NewITable()
	m.InsertOrUpdate(nil, 3)

	assert.Equal(t, int64(1), m.Size())
	value, found := m.Get(_p("0.0.0.0/0"))
	assert.True(t, found)
	assert.Equal(t, 3, value)
}

func TestITableGetOrInsertNil(t *testing.T) {
	m := NewITable()
	result := m.GetOrInsert(nil, 11)
	assert.Equal(t, 11, result)

	value, found := m.Get(_p("0.0.0.0/0"))
	assert.True(t, found)
	assert.Equal(t, 11, value)
}
