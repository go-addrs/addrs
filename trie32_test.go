package ipv4

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrieInsert(t *testing.T) {
	var trie trie32
	assert.Equal(t, 0, trie.NumNodes())

	key := Prefix{Addr{0x0ae01800}, 24}
	err := trie.Insert(key, true)
	assert.Nil(t, err)
	assert.Equal(t, 1, trie.NumNodes())
	assert.True(t, trie.isValid())
}

func TestTrieInsertOrUpdate(t *testing.T) {
	var trie trie32
	assert.Equal(t, 0, trie.NumNodes())

	key := Prefix{Addr{0x0ae01800}, 24}
	err := trie.InsertOrUpdate(key, true)
	assert.Nil(t, err)
	assert.Equal(t, 1, trie.NumNodes())
	match, matchedKey, value := trie.Match(key)
	assert.Equal(t, MatchExact, match)
	assert.Equal(t, key, matchedKey)
	assert.True(t, value.(bool))

	err = trie.InsertOrUpdate(key, false)
	assert.Nil(t, err)
	assert.Equal(t, 1, trie.NumNodes())
	match, matchedKey, value = trie.Match(key)
	assert.Equal(t, MatchExact, match)
	assert.Equal(t, key, matchedKey)
	assert.False(t, value.(bool))
	assert.True(t, trie.isValid())
}

func TestTrieUpdate(t *testing.T) {
	var trie trie32
	assert.Equal(t, 0, trie.NumNodes())

	key := Prefix{Addr{0x0ae01800}, 24}
	trie.Insert(key, false)

	err := trie.Update(key, true)
	assert.Nil(t, err)
	assert.Equal(t, 1, trie.NumNodes())
	match, matchedKey, value := trie.Match(key)
	assert.Equal(t, MatchExact, match)
	assert.Equal(t, key, matchedKey)
	assert.True(t, value.(bool))

	err = trie.Update(key, false)
	assert.Nil(t, err)
	assert.Equal(t, 1, trie.NumNodes())
	match, matchedKey, value = trie.Match(key)
	assert.Equal(t, MatchExact, match)
	assert.Equal(t, key, matchedKey)
	assert.False(t, value.(bool))
	assert.True(t, trie.isValid())
}

func TestTrieGetOrInsert(t *testing.T) {
	var trie trie32
	assert.Equal(t, 0, trie.NumNodes())

	key := Prefix{Addr{0x0ae01800}, 24}
	value, err := trie.GetOrInsert(key, true)
	assert.Nil(t, err)
	assert.True(t, value.(bool))
	assert.Equal(t, 1, trie.NumNodes())
	assert.True(t, trie.isValid())
}

func TestTrieMatch(t *testing.T) {
	var trie trie32

	insertKey := Prefix{Addr{0x0ae01800}, 24}
	trie.Insert(insertKey, true)

	t.Run("None", func(t *testing.T) {
		level, _, _ := trie.Match(Prefix{Addr{0x0ae01000}, 24})
		assert.Equal(t, MatchNone, level)
		assert.True(t, trie.isValid())
	})

	t.Run("Exact", func(t *testing.T) {
		level, key, value := trie.Match(Prefix{Addr{0x0ae01800}, 24})
		assert.Equal(t, MatchExact, level)
		assert.Equal(t, insertKey, key)
		assert.True(t, value.(bool))
		assert.True(t, trie.isValid())
	})

	t.Run("Contains", func(t *testing.T) {
		level, key, value := trie.Match(Prefix{Addr{0x0ae01817}, 32})
		assert.Equal(t, MatchContains, level)
		assert.Equal(t, insertKey, key)
		assert.True(t, value.(bool))
		assert.True(t, trie.isValid())
	})
}

func TestTrieDelete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var trie trie32

		insertKey := Prefix{Addr{0x0ae01800}, 24}
		trie.Insert(insertKey, true)

		key := Prefix{Addr{0x0ae01800}, 24}
		err := trie.Delete(key)
		assert.Nil(t, err)
		assert.Equal(t, 0, trie.NumNodes())
		assert.True(t, trie.isValid())
	})

	t.Run("Not Found", func(t *testing.T) {
		var trie trie32

		insertKey := Prefix{Addr{0x0ae01800}, 24}
		trie.Insert(insertKey, true)

		key := Prefix{Addr{0x0ae01000}, 24}
		err := trie.Delete(key)
		assert.NotNil(t, err)
		assert.Equal(t, 1, trie.NumNodes())
		assert.True(t, trie.isValid())
	})

	t.Run("Not Exact", func(t *testing.T) {
		var trie trie32

		insertKey := Prefix{Addr{0x0ae01800}, 24}
		trie.Insert(insertKey, true)

		key := Prefix{Addr{0x0ae01817}, 32}
		err := trie.Delete(key)
		assert.NotNil(t, err)
		assert.Equal(t, 1, trie.NumNodes())
		assert.True(t, trie.isValid())
	})
}

func TestTrieIterate(t *testing.T) {
	var trie trie32

	insertKey := Prefix{Addr{0x0ae01800}, 24}
	trie.Insert(insertKey, true)

	found := false
	trie.Iterate(func(key Prefix, value interface{}) bool {
		assert.Equal(t, insertKey, key)
		assert.True(t, value.(bool))
		found = true
		return true
	})
	assert.True(t, found)
	assert.True(t, trie.isValid())
}

func TestTrieAggregate(t *testing.T) {
	var trie trie32

	insertKey := Prefix{Addr{0x0ae01800}, 24}
	trie.Insert(insertKey, true)

	secondKey := Prefix{Addr{0x0ae01817}, 32}
	trie.Insert(secondKey, true)

	found := false
	trie.Aggregate(func(key Prefix, value interface{}) bool {
		assert.Equal(t, insertKey, key)
		assert.True(t, value.(bool))
		found = true
		return true
	})
	assert.True(t, found)
	assert.True(t, trie.isValid())
}

func TestTrieEqual(t *testing.T) {
	var a, b trie32

	assert.True(t, a.Equal(&b))
	assert.True(t, b.Equal(&a))

	a.Insert(Prefix{Addr{0x0ae01801}, 24}, true)
	assert.False(t, a.Equal(&b))
	assert.False(t, b.Equal(&a))

	b.Insert(Prefix{Addr{0x0ae01800}, 24}, true)
	assert.False(t, a.Equal(&b))
	assert.False(t, b.Equal(&a))
}
