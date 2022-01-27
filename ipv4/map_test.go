package ipv4

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInsertOrUpdate(t *testing.T) {
	m := NewMap()
	m.Insert(_a("10.224.24.1"), nil)
	m.InsertOrUpdate(_a("10.224.24.1"), 3)
	assert.Equal(t, int64(1), m.Size())

	data, ok := m.Get(_a("10.224.24.1"))
	assert.True(t, ok)
	assert.Equal(t, 3, data)
}

func TestInsertOrUpdateDuplicate(t *testing.T) {
	m := NewMap()
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
	m := NewMap()
	m.Insert(_p("10.224.24.0/24"), 3)
	assert.Equal(t, int64(1), m.Size())

	_, ok := m.Get(_a("10.224.24.1"))
	assert.False(t, ok)
}

func TestGetNotFound(t *testing.T) {
	m := NewMap()
	err := m.Insert(_a("10.224.24.1"), 3)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), m.Size())

	_, ok := m.Get(_a("10.225.24.1"))
	assert.False(t, ok)
}

func TestGetOrInsertOnlyExactMatch(t *testing.T) {
	m := NewMap()
	m.Insert(_p("10.224.24.0/24"), 3)
	assert.Equal(t, int64(1), m.Size())

	value := m.GetOrInsert(_a("10.224.24.1"), 5)
	assert.Equal(t, 5, value)
	assert.Equal(t, int64(2), m.Size())
}

func TestGetOrInsertNotFound(t *testing.T) {
	m := NewMap()
	err := m.Insert(_a("10.224.24.1"), 3)
	assert.Nil(t, err)

	value := m.GetOrInsert(_a("10.225.24.1"), 5)
	assert.Equal(t, 5, value)
	assert.Equal(t, int64(2), m.Size())
}

func TestGetOrInsertPrefixOnlyExactMatch(t *testing.T) {
	m := NewMap()
	m.Insert(_p("10.224.24.0/24"), 3)
	assert.Equal(t, int64(1), m.Size())

	value := m.GetOrInsert(_p("10.224.24.2/31"), 5)
	assert.Equal(t, 5, value)
	assert.Equal(t, int64(2), m.Size())
}

func TestGetOrInsertPrefixNotFound(t *testing.T) {
	m := NewMap()
	err := m.Insert(_a("10.224.24.1"), 3)
	assert.Nil(t, err)

	value := m.GetOrInsert(_p("10.225.24.2/31"), 5)
	assert.Equal(t, 5, value)
	assert.Equal(t, int64(2), m.Size())
}

func TestMatchLongestPrefixMatch(t *testing.T) {
	m := NewMap()
	m.Insert(_p("10.224.24.0/24"), 3)
	assert.Equal(t, int64(1), m.Size())
	m.Insert(_p("10.224.0.0/16"), 4)
	assert.Equal(t, int64(2), m.Size())

	matched, n, data := m.LongestMatch(_a("10.224.24.1"))
	assert.Equal(t, MatchContains, matched)
	assert.Equal(t, _p("10.224.24.0/24"), n)
	assert.Equal(t, 3, data)
}

func TestMatchNotFound(t *testing.T) {
	m := NewMap()
	err := m.Insert(_a("10.224.24.1"), 3)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), m.Size())

	matched, _, _ := m.LongestMatch(_a("10.225.24.1"))
	assert.Equal(t, MatchNone, matched)
}

func TestRemove(t *testing.T) {
	m := NewMap()
	err := m.Insert(_a("10.224.24.1"), 3)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), m.Size())

	m.Remove(_a("10.224.24.1"))
	assert.Equal(t, int64(0), m.Size())
}

func TestRemoveNotFound(t *testing.T) {
	m := NewMap()
	err := m.Insert(_a("10.224.24.1"), 3)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), m.Size())

	m.Remove(_a("10.225.24.1"))
	assert.Equal(t, int64(1), m.Size())
}

func TestInsert(t *testing.T) {
	m := NewMap()
	err := m.Insert(_p("10.224.24.0/24"), 3)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), m.Size())

	data, ok := m.Get(_p("10.224.24.0/24"))
	assert.True(t, ok)
	assert.Equal(t, 3, data)

	data, ok = m.Get(_p("10.225.24.0/24"))
	assert.False(t, ok)
}

func TestInsertOrUpdatePrefix(t *testing.T) {
	m := NewMap()
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
	m := NewMap()
	err := m.Insert(_p("10.224.24.0/24"), 3)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), m.Size())

	m.Remove(_p("10.224.24.0/24"))
	assert.Equal(t, int64(0), m.Size())
}

func TestRemovePrefixNotFound(t *testing.T) {
	m := NewMap()
	err := m.Insert(_p("10.224.24.0/24"), 3)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), m.Size())

	m.Remove(_p("10.225.24.0/24"))
	assert.Equal(t, int64(1), m.Size())
}

func TestMatchPrefixLongestPrefixMatch(t *testing.T) {
	m := NewMap()
	m.Insert(_p("10.224.24.0/24"), 3)
	assert.Equal(t, int64(1), m.Size())
	m.Insert(_p("10.224.0.0/16"), 4)
	assert.Equal(t, int64(2), m.Size())

	matched, n, data := m.LongestMatch(_p("10.224.24.0/27"))
	assert.Equal(t, MatchContains, matched)
	assert.Equal(t, 3, data)
	assert.Equal(t, _p("10.224.24.0/24"), n)
}

func TestMatchPrefixNotFound(t *testing.T) {
	m := NewMap()
	m.Insert(_p("10.224.24.0/24"), 3)
	assert.Equal(t, int64(1), m.Size())

	matched, _, _ := m.LongestMatch(_p("10.225.24.0/24"))
	assert.Equal(t, MatchNone, matched)
}

func TestExample1(t *testing.T) {
	m := NewMap()
	m.Insert(_p("10.224.24.2/31"), true)
	m.Insert(_p("10.224.24.1/32"), true)
	m.Insert(_p("10.224.24.0/32"), true)

	var result []string
	m.Iterate(func(prefix Prefix, value interface{}) bool {
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
	m.IterateAggregates(func(prefix Prefix, value interface{}) bool {
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
	m := NewMap()
	m.Insert(_p("10.224.24.0/30"), true)
	m.Insert(_p("10.224.24.0/31"), false)
	m.Insert(_p("10.224.24.1/32"), true)
	m.Insert(_p("10.224.24.0/32"), false)

	var result []pair
	m.Iterate(func(prefix Prefix, value interface{}) bool {
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
	m.IterateAggregates(func(prefix Prefix, value interface{}) bool {
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
	m := NewMap()
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
	m.Iterate(func(prefix Prefix, value interface{}) bool {
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
	m.Iterate(func(prefix Prefix, value interface{}) bool {
		iterations++
		return false
	})
	assert.Equal(t, 1, iterations)

	result = []string{}
	m.IterateAggregates(func(prefix Prefix, value interface{}) bool {
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
	m.IterateAggregates(func(prefix Prefix, value interface{}) bool {
		iterations++
		return false
	})
	assert.Equal(t, 1, iterations)
}

func TestMapInsert(t *testing.T) {
	m := NewMap()
	assert.Equal(t, int64(0), m.m.trie.NumNodes())

	key := Prefix{Address{0x0ae01800}, 24}
	err := m.Insert(key, true)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), m.m.trie.NumNodes())
	assert.True(t, m.m.trie.isValid())
}

func TestMapInsertOrUpdate(t *testing.T) {
	m := NewMap()
	assert.Equal(t, int64(0), m.m.trie.NumNodes())

	key := Prefix{Address{0x0ae01800}, 24}
	m.InsertOrUpdate(key, true)
	assert.Equal(t, int64(1), m.m.trie.NumNodes())
	match, matchedKey, value := m.LongestMatch(key)
	assert.Equal(t, MatchExact, match)
	assert.Equal(t, key, matchedKey)
	assert.True(t, value.(bool))

	m.InsertOrUpdate(key, false)
	assert.Equal(t, int64(1), m.m.trie.NumNodes())
	match, matchedKey, value = m.LongestMatch(key)
	assert.Equal(t, MatchExact, match)
	assert.Equal(t, key, matchedKey)
	assert.False(t, value.(bool))
	assert.True(t, m.m.trie.isValid())
}

func TestMapUpdate(t *testing.T) {
	m := NewMap()
	assert.Equal(t, int64(0), m.m.trie.NumNodes())

	key := Prefix{Address{0x0ae01800}, 24}
	m.Insert(key, false)

	err := m.Update(key, true)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), m.m.trie.NumNodes())
	match, matchedKey, value := m.LongestMatch(key)
	assert.Equal(t, MatchExact, match)
	assert.Equal(t, key, matchedKey)
	assert.True(t, value.(bool))

	err = m.Update(key, false)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), m.m.trie.NumNodes())
	match, matchedKey, value = m.LongestMatch(key)
	assert.Equal(t, MatchExact, match)
	assert.Equal(t, key, matchedKey)
	assert.False(t, value.(bool))
	assert.True(t, m.m.trie.isValid())
}

func TestMapGetOrInsert(t *testing.T) {
	m := NewMap()
	assert.Equal(t, int64(0), m.m.trie.NumNodes())

	key := Prefix{Address{0x0ae01800}, 24}
	value := m.GetOrInsert(key, true)
	assert.True(t, value.(bool))
	assert.Equal(t, int64(1), m.m.trie.NumNodes())
	assert.True(t, m.m.trie.isValid())
}

func TestMapMatch(t *testing.T) {
	m := NewMap()

	insertKey := Prefix{Address{0x0ae01800}, 24}
	m.Insert(insertKey, true)

	t.Run("None", func(t *testing.T) {
		level, _, _ := m.LongestMatch(Prefix{Address{0x0ae01000}, 24})
		assert.Equal(t, MatchNone, level)
		assert.True(t, m.m.trie.isValid())
	})

	t.Run("Exact", func(t *testing.T) {
		level, key, value := m.LongestMatch(Prefix{Address{0x0ae01800}, 24})
		assert.Equal(t, MatchExact, level)
		assert.Equal(t, insertKey, key)
		assert.True(t, value.(bool))
		assert.True(t, m.m.trie.isValid())
	})

	t.Run("Contains", func(t *testing.T) {
		level, key, value := m.LongestMatch(Prefix{Address{0x0ae01817}, 32})
		assert.Equal(t, MatchContains, level)
		assert.Equal(t, insertKey, key)
		assert.True(t, value.(bool))
		assert.True(t, m.m.trie.isValid())
	})
}

func TestMapRemovePrefix(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := NewMap()

		insertKey := Prefix{Address{0x0ae01800}, 24}
		m.Insert(insertKey, true)

		key := Prefix{Address{0x0ae01800}, 24}
		err := m.Remove(key)
		assert.Nil(t, err)
		assert.Equal(t, int64(0), m.m.trie.NumNodes())
		assert.True(t, m.m.trie.isValid())
	})

	t.Run("Not Found", func(t *testing.T) {
		m := NewMap()

		insertKey := Prefix{Address{0x0ae01800}, 24}
		m.Insert(insertKey, true)

		key := Prefix{Address{0x0ae01000}, 24}
		err := m.Remove(key)
		assert.NotNil(t, err)
		assert.Equal(t, int64(1), m.m.trie.NumNodes())
		assert.True(t, m.m.trie.isValid())
	})

	t.Run("Not Exact", func(t *testing.T) {
		m := NewMap()

		insertKey := Prefix{Address{0x0ae01800}, 24}
		m.Insert(insertKey, true)

		key := Prefix{Address{0x0ae01817}, 32}
		err := m.Remove(key)
		assert.NotNil(t, err)
		assert.Equal(t, int64(1), m.m.trie.NumNodes())
		assert.True(t, m.m.trie.isValid())
	})
}

func TestMapIterate(t *testing.T) {
	m := NewMap()

	insertKey := Prefix{Address{0x0ae01800}, 24}
	m.Insert(insertKey, true)

	found := false
	m.Iterate(func(key Prefix, value interface{}) bool {
		assert.Equal(t, insertKey, key)
		assert.True(t, value.(bool))
		found = true
		return true
	})
	assert.True(t, found)
	assert.True(t, m.m.trie.isValid())
}

func TestMapIterateAggregates(t *testing.T) {
	m := NewMap()

	insertKey := Prefix{Address{0x0ae01800}, 24}
	m.Insert(insertKey, true)

	secondKey := Prefix{Address{0x0ae01817}, 32}
	m.Insert(secondKey, true)

	found := false
	m.IterateAggregates(func(key Prefix, value interface{}) bool {
		assert.Equal(t, insertKey, key)
		assert.True(t, value.(bool))
		found = true
		return true
	})
	assert.True(t, found)
	assert.True(t, m.m.trie.isValid())
}

func TestMapEqual(t *testing.T) {
	a := NewMap()
	b := NewMap()

	assert.True(t, a.m.trie.Equal(b.m.trie))
	assert.True(t, b.m.trie.Equal(a.m.trie))

	a.Insert(Prefix{Address{0x0ae01801}, 24}, true)
	assert.False(t, a.m.trie.Equal(b.m.trie))
	assert.False(t, b.m.trie.Equal(a.m.trie))

	b.Insert(Prefix{Address{0x0ae01800}, 24}, true)
	assert.False(t, a.m.trie.Equal(b.m.trie))
	assert.False(t, b.m.trie.Equal(a.m.trie))
}

// Test that Maps, when passed and copied, refer to the same data
func TestMapAsReferenceType(t *testing.T) {
	m := NewMap()

	manipulate := func(m Map) {
		m.Insert(_a("10.224.24.1"), nil)
		m.InsertOrUpdate(_a("10.224.24.1"), 3)
	}
	manipulate(m)
	assert.Equal(t, int64(1), m.Size())
	data, ok := m.Get(_a("10.224.24.1"))
	assert.True(t, ok)
	assert.Equal(t, 3, data)
}
