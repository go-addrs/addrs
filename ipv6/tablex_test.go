package ipv6

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInsertOrUpdate(t *testing.T) {
	m := tableX{}.Table_()
	m.Insert(_a("2001::"), nil)
	m.InsertOrUpdate(_a("2001::"), 3)
	assert.Equal(t, int64(1), m.NumEntries())

	data, ok := m.Get(_a("2001::"))
	assert.True(t, ok)
	assert.Equal(t, 3, data)
}

func TestInsertOrUpdateDuplicate(t *testing.T) {
	m := newTableX_()
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

func TestGetOnlyExactMatch(t *testing.T) {
	m := newTableX_()
	m.Insert(_p("2001::/112"), 3)
	assert.Equal(t, int64(1), m.NumEntries())

	_, ok := m.Get(_a("2001::8000"))
	assert.False(t, ok)
}

func TestGetNotFound(t *testing.T) {
	m := newTableX_()
	succeeded := m.Insert(_a("2001::"), 3)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.NumEntries())

	_, ok := m.Get(_a("2001:0:0:1::"))
	assert.False(t, ok)
}

func TestGetOrInsertOnlyExactMatch(t *testing.T) {
	m := newTableX_()
	m.Insert(_p("2001::/112"), 3)
	assert.Equal(t, int64(1), m.NumEntries())

	value := m.GetOrInsert(_a("2001::8000"), 5)
	assert.Equal(t, 5, value)
	assert.Equal(t, int64(2), m.NumEntries())
}

func TestGetOrInsertNotFound(t *testing.T) {
	m := newTableX_()
	succeeded := m.Insert(_a("2001::"), 3)
	assert.True(t, succeeded)

	value := m.GetOrInsert(_a("2001:0:0:1::"), 5)
	assert.Equal(t, 5, value)
	assert.Equal(t, int64(2), m.NumEntries())
}

func TestGetOrInsertPrefixOnlyExactMatch(t *testing.T) {
	m := newTableX_()
	m.Insert(_p("2001::/112"), 3)
	assert.Equal(t, int64(1), m.NumEntries())

	value := m.GetOrInsert(_p("2001::0001/127"), 5)
	assert.Equal(t, 5, value)
	assert.Equal(t, int64(2), m.NumEntries())
}

func TestGetOrInsertPrefixNotFound(t *testing.T) {
	m := newTableX_()
	succeeded := m.Insert(_a("2001::0001"), 3)
	assert.True(t, succeeded)

	value := m.GetOrInsert(_p("2001::0002/127"), 5)
	assert.Equal(t, 5, value)
	assert.Equal(t, int64(2), m.NumEntries())
}

func TestMatchLongestPrefixMatch(t *testing.T) {
	m := newTableX_()
	m.Insert(_p("2001::1:0/112"), 3)
	assert.Equal(t, int64(1), m.NumEntries())
	m.Insert(_p("2001::/96"), 4)
	assert.Equal(t, int64(2), m.NumEntries())

	data, matched, n := m.LongestMatch(_a("2001::1:1"))
	assert.NotEqual(t, matchContains, n)
	assert.True(t, matched)
	assert.Equal(t, _p("2001::1:0/112"), n)
	assert.Equal(t, 3, data)
}

func TestMatchNotFound(t *testing.T) {
	m := newTableX_()
	succeeded := m.Insert(_a("2001::"), 3)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.NumEntries())

	_, matched, _ := m.LongestMatch(_a("2001:0:0:1::"))
	assert.False(t, matched)
}

func TestRemove(t *testing.T) {
	m := newTableX_()
	succeeded := m.Insert(_a("2001::"), 3)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.NumEntries())

	m.Remove(_a("2001::"))
	assert.Equal(t, int64(0), m.NumEntries())
}

func TestRemoveNotFound(t *testing.T) {
	m := newTableX_()
	succeeded := m.Insert(_a("2001::"), 3)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.NumEntries())

	m.Remove(_a("2001:0:0:1::"))
	assert.Equal(t, int64(1), m.NumEntries())
}

func TestInsert(t *testing.T) {
	m := newTableX_()
	succeeded := m.Insert(_p("2001::1:0/112"), 3)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.NumEntries())

	data, ok := m.Get(_p("2001::1:0/112"))
	assert.True(t, ok)
	assert.Equal(t, 3, data)

	data, ok = m.Get(_p("2001:0:1::1:0/112"))
	assert.False(t, ok)
	assert.Nil(t, data)
}

func TestInsertOrUpdatePrefix(t *testing.T) {
	m := newTableX_()
	m.Insert(_p("2001::/112"), nil)
	m.InsertOrUpdate(_p("2001::/112"), 3)
	assert.Equal(t, int64(1), m.NumEntries())

	data, ok := m.Get(_p("2001::/112"))
	assert.True(t, ok)
	assert.Equal(t, 3, data)

	data, ok = m.Get(_p("2001:0:0:1::/112"))
	assert.False(t, ok)
	assert.Nil(t, data)
}

func TestRemovePrefix(t *testing.T) {
	m := newTableX_()
	succeeded := m.Insert(_p("2001::/112"), 3)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.NumEntries())

	m.Remove(_p("2001::/112"))
	assert.Equal(t, int64(0), m.NumEntries())
}

func TestRemovePrefixNotFound(t *testing.T) {
	m := newTableX_()
	succeeded := m.Insert(_p("2001::/112"), 3)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.NumEntries())

	m.Remove(_p("2001:0:0:1::/112"))
	assert.Equal(t, int64(1), m.NumEntries())
}

func TestMatchPrefixLongestPrefixMatch(t *testing.T) {
	m := newTableX_()
	m.Insert(_p("2001::1:0/112"), 3)
	assert.Equal(t, int64(1), m.NumEntries())
	m.Insert(_p("2001::/96"), 4)
	assert.Equal(t, int64(2), m.NumEntries())

	key := _p("2001::1:0/114")
	data, matched, n := m.LongestMatch(key)
	assert.NotEqual(t, key, n)
	assert.True(t, matched)
	assert.Equal(t, 3, data)
	assert.Equal(t, _p("2001::1:0/112"), n)
}

func TestMatchPrefixNotFound(t *testing.T) {
	m := newTableX_()
	m.Insert(_p("2001::/112"), 3)
	assert.Equal(t, int64(1), m.NumEntries())

	_, matched, _ := m.LongestMatch(_p("2001:0:0:1::/112"))
	assert.False(t, matched)
}

func TestExample1(t *testing.T) {
	m := tableX{}.Table_()
	m.Insert(_p("2001::2/127"), true)
	m.Insert(_p("2001::1/128"), true)
	m.Insert(_p("2001::/128"), true)

	var result []string
	m.Table().Walk(func(prefix Prefix, value interface{}) bool {
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
	m.Table().Aggregate().Walk(func(prefix Prefix, value interface{}) bool {
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

type pair struct {
	prefix string
	value  interface{}
}

func TestExample2(t *testing.T) {
	m := newTableX_()
	m.Insert(_p("2001::/126"), true)
	m.Insert(_p("2001::/127"), false)
	m.Insert(_p("2001::1/128"), true)
	m.Insert(_p("2001::/128"), false)

	var result []pair
	m.Table().Walk(func(prefix Prefix, value interface{}) bool {
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
			pair{prefix: "2001::/126", value: true},
			pair{prefix: "2001::/127", value: false},
			pair{prefix: "2001::/128", value: false},
			pair{prefix: "2001::1/128", value: true},
		},
		result,
	)

	result = []pair{}
	m.Table().Aggregate().Walk(func(prefix Prefix, value interface{}) bool {
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
			pair{prefix: "2001::/126", value: true},
			pair{prefix: "2001::/127", value: false},
			pair{prefix: "2001::1/128", value: true},
		},
		result,
	)
}

func TestExample3(t *testing.T) {
	m := newTableX_()
	m.Insert(_p("172:21::/40"), nil)
	m.Insert(_p("192:68:27::/49"), nil)
	m.Insert(_p("192:168:26:8000::/49"), nil)
	m.Insert(_p("10:224:24::/64"), nil)
	m.Insert(_p("192:68:24::/48"), nil)
	m.Insert(_p("172:16::/24"), nil)
	m.Insert(_p("192:68:26::/48"), nil)
	m.Insert(_p("10:224:24::/60"), nil)
	m.Insert(_p("192:168:24::/48"), nil)
	m.Insert(_p("192:168:25::/48"), nil)
	m.Insert(_p("192:168:26::/49"), nil)
	m.Insert(_p("192:68:25::/48"), nil)
	m.Insert(_p("192:168:27::/48"), nil)
	m.Insert(_p("172:20:8000::/38"), nil)
	m.Insert(_p("192:68:27:8000::/49"), nil)

	var result []string
	m.Table().Walk(func(prefix Prefix, value interface{}) bool {
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
	m.Table().Walk(func(prefix Prefix, value interface{}) bool {
		iterations++
		return false
	})
	assert.Equal(t, 1, iterations)

	result = []string{}
	m.Table().Aggregate().Walk(func(prefix Prefix, value interface{}) bool {
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
	m.Table().Aggregate().Walk(func(prefix Prefix, value interface{}) bool {
		iterations++
		return false
	})
	assert.Equal(t, 1, iterations)
}

func TestTableXnsert(t *testing.T) {
	m := newTableX_()
	assert.Equal(t, int64(0), m.m.trie.NumNodes())

	key := Prefix{Address{uint128{0x2001, 0x0}}, 112}
	succeeded := m.Insert(key, true)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.m.trie.NumNodes())
	assert.True(t, m.m.trie.isValid())
}

func TestTableXnsertOrUpdate(t *testing.T) {
	m := newTableX_()
	assert.Equal(t, int64(0), m.m.trie.NumNodes())

	key := Prefix{Address{uint128{0x2001, 0x0}}, 112}
	m.InsertOrUpdate(key, true)
	assert.Equal(t, int64(1), m.m.trie.NumNodes())
	value, match, matchedKey := m.LongestMatch(key)
	assert.Equal(t, key, matchedKey)
	assert.True(t, match)
	assert.Equal(t, key, matchedKey)
	assert.True(t, value.(bool))

	m.InsertOrUpdate(key, false)
	assert.Equal(t, int64(1), m.m.trie.NumNodes())
	value, match, matchedKey = m.LongestMatch(key)
	assert.Equal(t, key, matchedKey)
	assert.True(t, match)
	assert.Equal(t, key, matchedKey)
	assert.False(t, value.(bool))
	assert.True(t, m.m.trie.isValid())
}

func TestTableUpdate(t *testing.T) {
	m := newTableX_()
	assert.Equal(t, int64(0), m.m.trie.NumNodes())

	key := Prefix{Address{uint128{0x2001, 0x0}}, 112}
	m.Insert(key, false)

	succeeded := m.Update(key, true)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.m.trie.NumNodes())
	value, match, matchedKey := m.LongestMatch(key)
	assert.Equal(t, key, matchedKey)
	assert.True(t, match)
	assert.Equal(t, key, matchedKey)
	assert.True(t, value.(bool))

	succeeded = m.Update(key, false)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.m.trie.NumNodes())
	value, match, matchedKey = m.LongestMatch(key)
	assert.Equal(t, key, matchedKey)
	assert.True(t, match)
	assert.Equal(t, key, matchedKey)
	assert.False(t, value.(bool))
	assert.True(t, m.m.trie.isValid())
}

func TestTableGetOrInsert(t *testing.T) {
	m := newTableX_()
	assert.Equal(t, int64(0), m.m.trie.NumNodes())

	key := Prefix{Address{uint128{0x2001, 0x0}}, 112}
	value := m.GetOrInsert(key, true)
	assert.True(t, value.(bool))
	assert.Equal(t, int64(1), m.m.trie.NumNodes())
	assert.True(t, m.m.trie.isValid())
}

func TestTableMatch(t *testing.T) {
	m := newTableX_()

	insertKey := Prefix{Address{uint128{0x2001, 0x0180000}}, 112}
	m.Insert(insertKey, true)

	t.Run("None", func(t *testing.T) {
		_, found, _ := m.LongestMatch(Prefix{Address{uint128{0x2001, 0x010000}}, 112})
		assert.False(t, found)
		assert.True(t, m.m.trie.isValid())
	})

	t.Run("Exact", func(t *testing.T) {
		prefix := Prefix{Address{uint128{0x2001, 0x0180000}}, 112}
		value, found, key := m.LongestMatch(prefix)
		assert.Equal(t, prefix, key)
		assert.True(t, found)
		assert.Equal(t, insertKey, key)
		assert.True(t, value.(bool))
		assert.True(t, m.m.trie.isValid())
	})

	t.Run("Contains", func(t *testing.T) {
		prefix := Prefix{Address{uint128{0x2001, 0x0180017}}, 128}
		value, found, key := m.LongestMatch(prefix)
		assert.True(t, found)
		assert.NotEqual(t, prefix, key)
		assert.Equal(t, insertKey, key)
		assert.True(t, value.(bool))
		assert.True(t, m.m.trie.isValid())
	})
}

func TestTableRemovePrefix(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := newTableX_()

		insertKey := Prefix{Address{uint128{0x2001, 0x0180000}}, 112}
		m.Insert(insertKey, true)

		key := Prefix{Address{uint128{0x2001, 0x0180000}}, 112}
		succeeded := m.Remove(key)
		assert.True(t, succeeded)
		assert.Equal(t, int64(0), m.m.trie.NumNodes())
		assert.True(t, m.m.trie.isValid())
	})

	t.Run("Not Found", func(t *testing.T) {
		m := newTableX_()

		insertKey := Prefix{Address{uint128{0x2001, 0x0180000}}, 112}
		m.Insert(insertKey, true)

		key := Prefix{Address{uint128{0x2001, 0x0100000}}, 112}
		succeeded := m.Remove(key)
		assert.False(t, succeeded)
		assert.Equal(t, int64(1), m.m.trie.NumNodes())
		assert.True(t, m.m.trie.isValid())
	})

	t.Run("Not Exact", func(t *testing.T) {
		m := newTableX_()

		insertKey := Prefix{Address{uint128{0x2001, 0x0180000}}, 112}
		m.Insert(insertKey, true)

		key := Prefix{Address{uint128{0x2001, 0x0180017}}, 128}
		succeeded := m.Remove(key)
		assert.False(t, succeeded)
		assert.Equal(t, int64(1), m.m.trie.NumNodes())
		assert.True(t, m.m.trie.isValid())
	})
}

func TestTableWalk(t *testing.T) {
	m := newTableX_()

	insertKey := Prefix{Address{uint128{0x2001, 0x0180000}}, 112}
	m.Insert(insertKey, true)

	found := false
	m.Table().Walk(func(key Prefix, value interface{}) bool {
		assert.Equal(t, insertKey, key)
		assert.True(t, value.(bool))
		found = true
		return true
	})
	assert.True(t, found)
	assert.True(t, m.m.trie.isValid())
}

func TestTableWalkAggregates(t *testing.T) {
	m := newTableX_()

	insertKey := Prefix{Address{uint128{0x2001, 0x0180000}}, 112}
	m.Insert(insertKey, true)

	secondKey := Prefix{Address{uint128{0x2001, 0x0180017}}, 128}
	m.Insert(secondKey, true)

	found := false
	m.Table().Aggregate().Walk(func(key Prefix, value interface{}) bool {
		assert.Equal(t, insertKey, key)
		assert.True(t, value.(bool))
		found = true
		return true
	})
	assert.True(t, found)
	assert.True(t, m.m.trie.isValid())
}

func ieq(a, b interface{}) bool {
	return a == b
}

func TestTableEqual(t *testing.T) {
	a := newTableX_()
	b := newTableX_()

	assert.True(t, a.m.trie.Equal(b.m.trie, ieq))
	assert.True(t, b.m.trie.Equal(a.m.trie, ieq))

	a.Insert(Prefix{Address{uint128{0x2001, 0x0180001}}, 112}, true)
	assert.False(t, a.m.trie.Equal(b.m.trie, ieq))
	assert.False(t, b.m.trie.Equal(a.m.trie, ieq))

	b.Insert(Prefix{Address{uint128{0x2001, 0x0180000}}, 112}, true)
	assert.False(t, a.m.trie.Equal(b.m.trie, ieq))
	assert.False(t, b.m.trie.Equal(a.m.trie, ieq))
}

// Test that Tables, when passed and copied, refer to the same data
func TestTableAsReferenceType(t *testing.T) {
	m := newTableX_()

	manipulate := func(m tableX_) {
		m.Insert(_a("2001::1"), nil)
		m.InsertOrUpdate(_a("2001::1"), 3)
	}
	manipulate(m)
	assert.Equal(t, int64(1), m.NumEntries())
	data, ok := m.Get(_a("2001::1"))
	assert.True(t, ok)
	assert.Equal(t, 3, data)
}

func TestTableConcurrentModification(t *testing.T) {
	m := newTableX_()

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

			newHead, _ := m.m.trie.Insert(_p("2001::/112"), nil)
			return true, newHead
		})
	}()
	go func() {
		defer wrap()
		m.mutate(func() (bool, *trieNode) {
			<-ch
			newHead, _ := m.m.trie.Insert(_p("2001:0:1::/112"), nil)
			return true, newHead
		})
	}()
	wg.Wait()
	assert.Equal(t, 1, panicked)
}

func TestNilTableX(t *testing.T) {
	var table tableX_

	// On-offs
	assert.Equal(t, int64(0), table.NumEntries())
	assert.Equal(t, int64(0), table.Table().NumEntries())
	_, found := table.Get(_a("2001:0:0:123::"))
	assert.False(t, found)
	_, matched, _ := table.LongestMatch(_a("2001:0:0:123::"))
	assert.False(t, matched)

	// Walk
	assert.True(t, table.Table().Walk(func(Prefix, interface{}) bool {
		panic("should not be called")
	}))
	assert.True(t, table.Table().Aggregate().Walk(func(Prefix, interface{}) bool {
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
			table.Insert(_a("2001:0:0:123::"), nil)
		})
	})
	t.Run("update panics", func(t *testing.T) {
		testPanic(func() {
			table.Update(_a("2001:0:0:123::"), nil)
		})
	})
	t.Run("insert or update panics", func(t *testing.T) {
		testPanic(func() {
			table.InsertOrUpdate(_a("2001:0:0:123::"), nil)
		})
	})
	t.Run("get or insert panics", func(t *testing.T) {
		testPanic(func() {
			table.GetOrInsert(_a("2001:0:0:123::"), nil)
		})
	})
	t.Run("remove panics", func(t *testing.T) {
		testPanic(func() {
			table.Remove(_a("2001:0:0:123::"))
		})
	})
}

func TestTableXInsertNil(t *testing.T) {
	m := newTableX_()
	succeeded := m.Insert(nil, 3)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.NumEntries())
	value, found := m.Get(_p("::/0"))
	assert.True(t, found)
	assert.Equal(t, 3, value)
}

func TestTableXUpdateNil(t *testing.T) {
	m := newTableX_()
	m.Insert(_p("::/0"), 10)

	succeeded := m.Update(nil, 3)
	assert.True(t, succeeded)
	assert.Equal(t, int64(1), m.NumEntries())
	value, found := m.Get(_p("::/0"))
	assert.True(t, found)
	assert.Equal(t, 3, value)
}

func TestTableXRemoveNil(t *testing.T) {
	m := newTableX_()
	m.Insert(_p("::/0"), 10)

	succeeded := m.Remove(nil)
	assert.True(t, succeeded)

	_, found := m.Get(_p("::/0"))
	assert.False(t, found)
}

func TestTableXLongestMatch(t *testing.T) {
	m := newTableX_()
	m.Insert(_p("::/0"), 10)

	value, matched, prefix := m.LongestMatch(nil)
	assert.Equal(t, 10, value)
	assert.True(t, matched)
	assert.Equal(t, prefix, Prefix{})
	assert.Equal(t, _p("::/0"), prefix)
}

func TestTableXInsertOrUpdateNil(t *testing.T) {
	m := newTableX_()
	m.InsertOrUpdate(nil, 3)

	assert.Equal(t, int64(1), m.NumEntries())
	value, found := m.Get(_p("::/0"))
	assert.True(t, found)
	assert.Equal(t, 3, value)
}

func TestTableXGetOrInsertNil(t *testing.T) {
	m := newTableX_()
	result := m.GetOrInsert(nil, 11)
	assert.Equal(t, 11, result)

	value, found := m.Get(_p("::/0"))
	assert.True(t, found)
	assert.Equal(t, 11, value)
}

func TestTableXDiff(t *testing.T) {
	a := tableX{}.Build(func(a_ tableX_) bool {
		a_.Insert(_p("2001::1234:0/115"), true)
		a_.Insert(_p("2001::1234:6400/115"), true)
		a_.Insert(_p("2001::1234:0/113"), true)
		return true
	})

	a = a.Build(func(a_ tableX_) bool {
		a_.Insert(_p("1900::1234:0/113"), true)
		return false
	})

	b := tableX{}.Build(func(b_ tableX_) bool {
		b_.Insert(_p("2001::1234:0/115"), true)
		b_.Insert(_p("2001::1234:9600/115"), true)
		b_.Insert(_p("2001::1234:0/113"), false)
		return true
	})

	type action struct {
		prefix        Prefix
		before, after interface{}
	}

	var actions []action
	getHandlers := func() (left, right func(Prefix, interface{}) bool, changed func(p Prefix, left, right interface{}) bool) {
		actions = nil
		left = func(p Prefix, v interface{}) bool {
			actions = append(actions, action{p, v, nil})
			return true
		}
		right = func(p Prefix, v interface{}) bool {
			actions = append(actions, action{p, nil, v})
			return true
		}
		changed = func(p Prefix, l, r interface{}) bool {
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
			action{_p("2001::1234:6400/115"), true, nil},
			action{_p("2001::1234:9600/115"), nil, true},
		}, actions)
	})

	t.Run("backward", func(t *testing.T) {
		left, right, changed := getHandlers()
		b.Diff(a, changed, left, right, nil)
		assert.Equal(t, []action{
			action{_p("2001::1234:0/113"), false, true},
			action{_p("2001::1234:6400/115"), nil, true},
			action{_p("2001::1234:9600/115"), true, nil},
		}, actions)
	})
}

func TestFixedTable(t *testing.T) {
	addrOne := _a("2001::1")
	addrTwo := _a("2001::2")
	addrThree := _a("2001::3")

	m := newTableX_()
	succeeded := m.Insert(addrOne, nil)
	assert.True(t, succeeded)

	im := m.Table()
	succeeded = m.Insert(addrTwo, nil)
	assert.True(t, succeeded)

	m2 := im.Table_()
	succeeded = m2.Insert(addrThree, nil)
	assert.True(t, succeeded)

	var found bool

	_, found = m.Get(addrOne)
	assert.True(t, found)
	_, found = m.Get(addrTwo)
	assert.True(t, found)
	_, found = m.Get(addrThree)
	assert.False(t, found)

	assert.Equal(t, int64(1), im.NumEntries())
	_, found = im.Get(addrOne)
	assert.True(t, found)
	_, found = im.Get(addrTwo)
	assert.False(t, found)
	_, found = im.Get(addrThree)
	assert.False(t, found)

	assert.Equal(t, int64(2), m2.NumEntries())
	_, found = m2.Get(addrOne)
	assert.True(t, found)
	_, found = m2.Get(addrTwo)
	assert.False(t, found)
	_, found = m2.Get(addrThree)
	assert.True(t, found)
}

func TestTableXMap(t *testing.T) {
	var a tableX
	assert.Equal(t, a, a.Map(nil))
	assert.Equal(t, a, a.Map(func(Prefix, interface{}) interface{} {
		panic("this should not be run")
	}))

	a = func() tableX {
		a := tableX{}.Table_()
		a.Insert(_p("2001::1234:0/115"), true)
		a.Insert(_p("2001::1234:6400/115"), true)
		a.Insert(_p("2001::1234:0/113"), true)
		return a.Table()
	}()

	result := a.Map(func(Prefix, interface{}) interface{} {
		return false
	})

	assert.Equal(t, int64(3), result.NumEntries())

	value, ok := result.Get(_p("2001::1234:0/115"))
	assert.True(t, ok)
	assert.False(t, value.(bool))
	value, ok = result.Get(_p("2001::1234:6400/115"))
	assert.True(t, ok)
	assert.False(t, value.(bool))
	value, ok = result.Get(_p("2001::1234:0/113"))
	assert.True(t, ok)
	assert.False(t, value.(bool))

	value, ok = result.Get(_p("::/0"))
	assert.False(t, ok)
	assert.Nil(t, value)
}

type creativeComparable struct {
	i int
}

func (me creativeComparable) Equal(other creativeComparable) bool {
	return me.i <= 2 && other.i <= 2
}

func TestTableXVariousComparators(t *testing.T) {
	tests := []struct {
		description string
		table       tableX_
		expected    []string
	}{
		{
			description: "comparable",
			table:       tableX{}.Table_(),
			expected: []string{
				"2001::1234:0/118",
				"2001::1234:0/122",
				"2001::1234:0/124",
				"2001::1234:0/126",
				"2001::1234:0/128",
			},
		}, {
			description: "not_comparable",
			table: newTableXCustomCompare_(func(a, b interface{}) bool {
				return false
			}),
			expected: []string{
				"2001::1234:0/118",
				"2001::1234:0/120",
				"2001::1234:0/122",
				"2001::1234:0/124",
				"2001::1234:0/126",
				"2001::1234:0/128",
			},
		}, {
			description: "custom_comparable",
			table: newTableXCustomCompare_(func(a, b interface{}) bool {
				return a.(creativeComparable).i <= 3 && b.(creativeComparable).i <= 3
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
			tt.table.Table().Aggregate().Walk(func(prefix Prefix, _ interface{}) bool {
				result = append(result, prefix.String())
				return true
			})
			assert.Equal(t, tt.expected, result)
		})
	}
}
