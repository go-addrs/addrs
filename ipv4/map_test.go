package ipv4

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInsertOrUpdate(t *testing.T) {
	m := Map{}
	m.Insert(unsafeParseAddress("10.224.24.1"), nil)
	err := m.InsertOrUpdate(unsafeParseAddress("10.224.24.1"), 3)
	assert.Nil(t, err)
	assert.Equal(t, 1, m.Size())

	data, ok := m.Get(unsafeParseAddress("10.224.24.1"))
	assert.True(t, ok)
	assert.Equal(t, 3, data)
}

func TestInsertOrUpdateDuplicate(t *testing.T) {
	m := Map{}
	err := m.InsertOrUpdate(unsafeParseAddress("10.224.24.1"), 3)
	assert.Nil(t, err)
	assert.Equal(t, 1, m.Size())
	data, ok := m.Get(unsafeParseAddress("10.224.24.1"))
	assert.True(t, ok)
	assert.Equal(t, 3, data)

	err = m.InsertOrUpdate(unsafeParseAddress("10.224.24.1"), 4)
	assert.Nil(t, err)
	assert.Equal(t, 1, m.Size())
	data, ok = m.Get(unsafeParseAddress("10.224.24.1"))
	assert.True(t, ok)
	assert.Equal(t, 4, data)
}

func TestGetOnlyExactMatch(t *testing.T) {
	m := Map{}
	m.InsertPrefix(unsafeParsePrefix("10.224.24.0/24"), 3)
	assert.Equal(t, 1, m.Size())

	_, ok := m.Get(unsafeParseAddress("10.224.24.1"))
	assert.False(t, ok)
}

func TestGetNotFound(t *testing.T) {
	m := Map{}
	err := m.Insert(unsafeParseAddress("10.224.24.1"), 3)
	assert.Nil(t, err)
	assert.Equal(t, 1, m.Size())

	_, ok := m.Get(unsafeParseAddress("10.225.24.1"))
	assert.False(t, ok)
}

func TestGetOrInsertOnlyExactMatch(t *testing.T) {
	m := Map{}
	m.InsertPrefix(unsafeParsePrefix("10.224.24.0/24"), 3)
	assert.Equal(t, 1, m.Size())

	value, err := m.GetOrInsert(unsafeParseAddress("10.224.24.1"), 5)
	assert.Nil(t, err)
	assert.Equal(t, 5, value)
	assert.Equal(t, 2, m.Size())
}

func TestGetOrInsertNotFound(t *testing.T) {
	m := Map{}
	err := m.Insert(unsafeParseAddress("10.224.24.1"), 3)
	assert.Nil(t, err)

	value, err := m.GetOrInsert(unsafeParseAddress("10.225.24.1"), 5)
	assert.Nil(t, err)
	assert.Equal(t, 5, value)
	assert.Equal(t, 2, m.Size())
}

func TestGetOrInsertPrefixOnlyExactMatch(t *testing.T) {
	m := Map{}
	m.InsertPrefix(unsafeParsePrefix("10.224.24.0/24"), 3)
	assert.Equal(t, 1, m.Size())

	value, err := m.GetOrInsertPrefix(unsafeParsePrefix("10.224.24.2/31"), 5)
	assert.Nil(t, err)
	assert.Equal(t, 5, value)
	assert.Equal(t, 2, m.Size())
}

func TestGetOrInsertPrefixNotFound(t *testing.T) {
	m := Map{}
	err := m.Insert(unsafeParseAddress("10.224.24.1"), 3)
	assert.Nil(t, err)

	value, err := m.GetOrInsertPrefix(unsafeParsePrefix("10.225.24.2/31"), 5)
	assert.Nil(t, err)
	assert.Equal(t, 5, value)
	assert.Equal(t, 2, m.Size())
}

func TestMatchLongestPrefixMatch(t *testing.T) {
	m := Map{}
	m.InsertPrefix(unsafeParsePrefix("10.224.24.0/24"), 3)
	assert.Equal(t, 1, m.Size())
	m.InsertPrefix(unsafeParsePrefix("10.224.0.0/16"), 4)
	assert.Equal(t, 2, m.Size())

	matched, n, data := m.LongestMatch(unsafeParseAddress("10.224.24.1"))
	assert.Equal(t, MatchContains, matched)
	assert.Equal(t, unsafeParsePrefix("10.224.24.0/24"), n)
	assert.Equal(t, 3, data)
}

func TestMatchNotFound(t *testing.T) {
	m := Map{}
	err := m.Insert(unsafeParseAddress("10.224.24.1"), 3)
	assert.Nil(t, err)
	assert.Equal(t, 1, m.Size())

	matched, _, _ := m.LongestMatch(unsafeParseAddress("10.225.24.1"))
	assert.Equal(t, MatchNone, matched)
}

func TestRemove(t *testing.T) {
	m := Map{}
	err := m.Insert(unsafeParseAddress("10.224.24.1"), 3)
	assert.Nil(t, err)
	assert.Equal(t, 1, m.Size())

	m.Remove(unsafeParseAddress("10.224.24.1"))
	assert.Equal(t, 0, m.Size())
}

func TestRemoveNotFound(t *testing.T) {
	m := Map{}
	err := m.Insert(unsafeParseAddress("10.224.24.1"), 3)
	assert.Nil(t, err)
	assert.Equal(t, 1, m.Size())

	m.Remove(unsafeParseAddress("10.225.24.1"))
	assert.Equal(t, 1, m.Size())
}

func TestInsertPrefix(t *testing.T) {
	m := Map{}
	err := m.InsertPrefix(unsafeParsePrefix("10.224.24.0/24"), 3)
	assert.Nil(t, err)
	assert.Equal(t, 1, m.Size())

	data, ok := m.GetPrefix(unsafeParsePrefix("10.224.24.0/24"))
	assert.True(t, ok)
	assert.Equal(t, 3, data)

	data, ok = m.GetPrefix(unsafeParsePrefix("10.225.24.0/24"))
	assert.False(t, ok)
}

func TestInsertOrUpdatePrefix(t *testing.T) {
	m := Map{}
	m.InsertPrefix(unsafeParsePrefix("10.224.24.0/24"), nil)
	err := m.InsertOrUpdatePrefix(unsafeParsePrefix("10.224.24.0/24"), 3)
	assert.Nil(t, err)
	assert.Equal(t, 1, m.Size())

	data, ok := m.GetPrefix(unsafeParsePrefix("10.224.24.0/24"))
	assert.True(t, ok)
	assert.Equal(t, 3, data)

	data, ok = m.GetPrefix(unsafeParsePrefix("10.225.24.0/24"))
	assert.False(t, ok)
}

func TestRemovePrefix(t *testing.T) {
	m := Map{}
	err := m.InsertPrefix(unsafeParsePrefix("10.224.24.0/24"), 3)
	assert.Nil(t, err)
	assert.Equal(t, 1, m.Size())

	m.RemovePrefix(unsafeParsePrefix("10.224.24.0/24"))
	assert.Equal(t, 0, m.Size())
}

func TestRemovePrefixNotFound(t *testing.T) {
	m := Map{}
	err := m.InsertPrefix(unsafeParsePrefix("10.224.24.0/24"), 3)
	assert.Nil(t, err)
	assert.Equal(t, 1, m.Size())

	m.RemovePrefix(unsafeParsePrefix("10.225.24.0/24"))
	assert.Equal(t, 1, m.Size())
}

func TestMatchPrefixLongestPrefixMatch(t *testing.T) {
	m := Map{}
	m.InsertPrefix(unsafeParsePrefix("10.224.24.0/24"), 3)
	assert.Equal(t, 1, m.Size())
	m.InsertPrefix(unsafeParsePrefix("10.224.0.0/16"), 4)
	assert.Equal(t, 2, m.Size())

	matched, n, data := m.LongestMatchPrefix(unsafeParsePrefix("10.224.24.0/27"))
	assert.Equal(t, MatchContains, matched)
	assert.Equal(t, 3, data)
	assert.Equal(t, unsafeParsePrefix("10.224.24.0/24"), n)
}

func TestMatchPrefixNotFound(t *testing.T) {
	m := Map{}
	m.InsertPrefix(unsafeParsePrefix("10.224.24.0/24"), 3)
	assert.Equal(t, 1, m.Size())

	matched, _, _ := m.LongestMatchPrefix(unsafeParsePrefix("10.225.24.0/24"))
	assert.Equal(t, MatchNone, matched)
}

func TestExample1(t *testing.T) {
	m := Map{}
	m.InsertPrefix(unsafeParsePrefix("10.224.24.2/31"), true)
	m.InsertPrefix(unsafeParsePrefix("10.224.24.1/32"), true)
	m.InsertPrefix(unsafeParsePrefix("10.224.24.0/32"), true)

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
	m := Map{}
	m.InsertPrefix(unsafeParsePrefix("10.224.24.0/30"), true)
	m.InsertPrefix(unsafeParsePrefix("10.224.24.0/31"), false)
	m.InsertPrefix(unsafeParsePrefix("10.224.24.1/32"), true)
	m.InsertPrefix(unsafeParsePrefix("10.224.24.0/32"), false)

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
	m := Map{}
	m.InsertPrefix(unsafeParsePrefix("172.21.0.0/20"), nil)
	m.InsertPrefix(unsafeParsePrefix("192.68.27.0/25"), nil)
	m.InsertPrefix(unsafeParsePrefix("192.168.26.128/25"), nil)
	m.InsertPrefix(unsafeParsePrefix("10.224.24.0/32"), nil)
	m.InsertPrefix(unsafeParsePrefix("192.68.24.0/24"), nil)
	m.InsertPrefix(unsafeParsePrefix("172.16.0.0/12"), nil)
	m.InsertPrefix(unsafeParsePrefix("192.68.26.0/24"), nil)
	m.InsertPrefix(unsafeParsePrefix("10.224.24.0/30"), nil)
	m.InsertPrefix(unsafeParsePrefix("192.168.24.0/24"), nil)
	m.InsertPrefix(unsafeParsePrefix("192.168.25.0/24"), nil)
	m.InsertPrefix(unsafeParsePrefix("192.168.26.0/25"), nil)
	m.InsertPrefix(unsafeParsePrefix("192.68.25.0/24"), nil)
	m.InsertPrefix(unsafeParsePrefix("192.168.27.0/24"), nil)
	m.InsertPrefix(unsafeParsePrefix("172.20.128.0/19"), nil)
	m.InsertPrefix(unsafeParsePrefix("192.68.27.128/25"), nil)

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
	var trie Map
	assert.Equal(t, 0, trie.trie.NumNodes())

	key := Prefix{Address{0x0ae01800}, 24}
	err := trie.InsertPrefix(key, true)
	assert.Nil(t, err)
	assert.Equal(t, 1, trie.trie.NumNodes())
	assert.True(t, trie.trie.isValid())
}

func TestMapInsertOrUpdate(t *testing.T) {
	var trie Map
	assert.Equal(t, 0, trie.trie.NumNodes())

	key := Prefix{Address{0x0ae01800}, 24}
	err := trie.InsertOrUpdatePrefix(key, true)
	assert.Nil(t, err)
	assert.Equal(t, 1, trie.trie.NumNodes())
	match, matchedKey, value := trie.LongestMatchPrefix(key)
	assert.Equal(t, MatchExact, match)
	assert.Equal(t, key, matchedKey)
	assert.True(t, value.(bool))

	err = trie.InsertOrUpdatePrefix(key, false)
	assert.Nil(t, err)
	assert.Equal(t, 1, trie.trie.NumNodes())
	match, matchedKey, value = trie.LongestMatchPrefix(key)
	assert.Equal(t, MatchExact, match)
	assert.Equal(t, key, matchedKey)
	assert.False(t, value.(bool))
	assert.True(t, trie.trie.isValid())
}

func TestMapUpdate(t *testing.T) {
	var trie Map
	assert.Equal(t, 0, trie.trie.NumNodes())

	key := Prefix{Address{0x0ae01800}, 24}
	trie.InsertPrefix(key, false)

	err := trie.UpdatePrefix(key, true)
	assert.Nil(t, err)
	assert.Equal(t, 1, trie.trie.NumNodes())
	match, matchedKey, value := trie.LongestMatchPrefix(key)
	assert.Equal(t, MatchExact, match)
	assert.Equal(t, key, matchedKey)
	assert.True(t, value.(bool))

	err = trie.UpdatePrefix(key, false)
	assert.Nil(t, err)
	assert.Equal(t, 1, trie.trie.NumNodes())
	match, matchedKey, value = trie.LongestMatchPrefix(key)
	assert.Equal(t, MatchExact, match)
	assert.Equal(t, key, matchedKey)
	assert.False(t, value.(bool))
	assert.True(t, trie.trie.isValid())
}

func TestMapGetOrInsert(t *testing.T) {
	var trie Map
	assert.Equal(t, 0, trie.trie.NumNodes())

	key := Prefix{Address{0x0ae01800}, 24}
	value, err := trie.GetOrInsertPrefix(key, true)
	assert.Nil(t, err)
	assert.True(t, value.(bool))
	assert.Equal(t, 1, trie.trie.NumNodes())
	assert.True(t, trie.trie.isValid())
}

func TestMapMatch(t *testing.T) {
	var trie Map

	insertKey := Prefix{Address{0x0ae01800}, 24}
	trie.InsertPrefix(insertKey, true)

	t.Run("None", func(t *testing.T) {
		level, _, _ := trie.LongestMatchPrefix(Prefix{Address{0x0ae01000}, 24})
		assert.Equal(t, MatchNone, level)
		assert.True(t, trie.trie.isValid())
	})

	t.Run("Exact", func(t *testing.T) {
		level, key, value := trie.LongestMatchPrefix(Prefix{Address{0x0ae01800}, 24})
		assert.Equal(t, MatchExact, level)
		assert.Equal(t, insertKey, key)
		assert.True(t, value.(bool))
		assert.True(t, trie.trie.isValid())
	})

	t.Run("Contains", func(t *testing.T) {
		level, key, value := trie.LongestMatchPrefix(Prefix{Address{0x0ae01817}, 32})
		assert.Equal(t, MatchContains, level)
		assert.Equal(t, insertKey, key)
		assert.True(t, value.(bool))
		assert.True(t, trie.trie.isValid())
	})
}

func TestMapRemovePrefix(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var trie Map

		insertKey := Prefix{Address{0x0ae01800}, 24}
		trie.InsertPrefix(insertKey, true)

		key := Prefix{Address{0x0ae01800}, 24}
		err := trie.RemovePrefix(key)
		assert.Nil(t, err)
		assert.Equal(t, 0, trie.trie.NumNodes())
		assert.True(t, trie.trie.isValid())
	})

	t.Run("Not Found", func(t *testing.T) {
		var trie Map

		insertKey := Prefix{Address{0x0ae01800}, 24}
		trie.InsertPrefix(insertKey, true)

		key := Prefix{Address{0x0ae01000}, 24}
		err := trie.RemovePrefix(key)
		assert.NotNil(t, err)
		assert.Equal(t, 1, trie.trie.NumNodes())
		assert.True(t, trie.trie.isValid())
	})

	t.Run("Not Exact", func(t *testing.T) {
		var trie Map

		insertKey := Prefix{Address{0x0ae01800}, 24}
		trie.InsertPrefix(insertKey, true)

		key := Prefix{Address{0x0ae01817}, 32}
		err := trie.RemovePrefix(key)
		assert.NotNil(t, err)
		assert.Equal(t, 1, trie.trie.NumNodes())
		assert.True(t, trie.trie.isValid())
	})
}

func TestMapIterate(t *testing.T) {
	var trie Map

	insertKey := Prefix{Address{0x0ae01800}, 24}
	trie.InsertPrefix(insertKey, true)

	found := false
	trie.Iterate(func(key Prefix, value interface{}) bool {
		assert.Equal(t, insertKey, key)
		assert.True(t, value.(bool))
		found = true
		return true
	})
	assert.True(t, found)
	assert.True(t, trie.trie.isValid())
}

func TestMapIterateAggregates(t *testing.T) {
	var trie Map

	insertKey := Prefix{Address{0x0ae01800}, 24}
	trie.InsertPrefix(insertKey, true)

	secondKey := Prefix{Address{0x0ae01817}, 32}
	trie.InsertPrefix(secondKey, true)

	found := false
	trie.IterateAggregates(func(key Prefix, value interface{}) bool {
		assert.Equal(t, insertKey, key)
		assert.True(t, value.(bool))
		found = true
		return true
	})
	assert.True(t, found)
	assert.True(t, trie.trie.isValid())
}

func TestMapEqual(t *testing.T) {
	var a, b Map

	assert.True(t, a.trie.Equal(b.trie))
	assert.True(t, b.trie.Equal(a.trie))

	a.InsertPrefix(Prefix{Address{0x0ae01801}, 24}, true)
	assert.False(t, a.trie.Equal(b.trie))
	assert.False(t, b.trie.Equal(a.trie))

	b.InsertPrefix(Prefix{Address{0x0ae01800}, 24}, true)
	assert.False(t, a.trie.Equal(b.trie))
	assert.False(t, b.trie.Equal(a.trie))
}
