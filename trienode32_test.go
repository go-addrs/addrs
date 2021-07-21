package ipv4

import (
	"fmt"
	"reflect"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestActive32(t *testing.T) {
	var node *trieNode32
	assert.False(t, node.active())
	assert.False(t, (&trieNode32{}).active())
	assert.True(t, (&trieNode32{isActive: true}).active())
}

func TestStructSizes32(t *testing.T) {
	// This test has two purposes. The first is to remind any future
	// contributors to be mindful of the size and alignment of these structs
	// and how to measure it. The second is that I'm curious to see if this
	// breaks on any architectures. Like if the go compiler aligns things
	// differently on ARM or whatever. I don't think it will.

	// All the casting to `int` here is because testify didn't consider
	// `uintptr` as comparable and I want to use it for its verbose output on
	// failure. Even if uintptr were comparable, I would have had to cast the
	// constants to uintptr.

	key := Prefix{}
	keySize := int(unsafe.Sizeof(key))
	keyAlign := int(unsafe.Alignof(key))

	node := trieNode32{}
	nodeSize := int(unsafe.Sizeof(node))
	nodeAlign := int(unsafe.Alignof(node))

	// Why would this ever be more than 8?
	assert.LessOrEqual(t, keyAlign, 8)
	assert.LessOrEqual(t, nodeAlign, 8)

	assert.Equal(t,
		intMin(
			2*keyAlign,
			8,
		),
		keySize,
	)
	assert.Equal(t,
		intMin(
			48,
			keySize+6*nodeAlign,
		),
		nodeSize,
	)
}

func TestMatchNilKey32(t *testing.T) {
	var trie *trieNode32
	var key Prefix

	assert.Nil(t, trie.Match(key))
}

func TestInsertOrUpdateChangeValue32(t *testing.T) {
	var trie *trieNode32

	key := Prefix{}

	trie, err := trie.InsertOrUpdate(key, true)
	assert.Equal(t, 1, trie.NumNodes())
	assert.Nil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))

	trie, err = trie.InsertOrUpdate(key, false)
	assert.Equal(t, 1, trie.NumNodes())
	assert.Nil(t, err)
	assert.False(t, trie.Match(key).Data.(bool))
}

func TestInsertOrUpdateNewKey32(t *testing.T) {
	var trie *trieNode32

	key := Prefix{}

	trie, err := trie.InsertOrUpdate(key, true)
	assert.Equal(t, 1, trie.NumNodes())
	assert.Nil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))

	newKey := Prefix{Addr{0}, 1}
	trie, err = trie.InsertOrUpdate(newKey, false)
	assert.Equal(t, 2, trie.NumNodes())
	assert.Nil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))
	assert.False(t, trie.Match(newKey).Data.(bool))
}

func TestInsertOrUpdateNarrowerKey32(t *testing.T) {
	var trie *trieNode32

	key := Prefix{Addr{0}, 1}

	trie, err := trie.InsertOrUpdate(key, true)
	assert.Equal(t, 1, trie.NumNodes())
	assert.Nil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))

	newKey := Prefix{}
	trie, err = trie.InsertOrUpdate(newKey, false)
	assert.Equal(t, 2, trie.NumNodes())
	assert.Nil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))
	assert.False(t, trie.Match(newKey).Data.(bool))
}

func TestInsertOrUpdateDisjointKeys32(t *testing.T) {
	var trie *trieNode32

	key := Prefix{Addr{0}, 1}

	trie, err := trie.InsertOrUpdate(key, true)
	assert.Equal(t, 1, trie.NumNodes())
	assert.Nil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))

	newKey := Prefix{Addr{0x80000000}, 1}
	trie, err = trie.InsertOrUpdate(newKey, false)
	assert.Equal(t, 2, trie.NumNodes())
	assert.Nil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))
	assert.False(t, trie.Match(newKey).Data.(bool))
}

func TestInsertOrUpdateInactive32(t *testing.T) {
	var trie *trieNode32

	key := Prefix{Addr{0}, 1}

	trie, err := trie.InsertOrUpdate(key, true)
	assert.Equal(t, 1, trie.NumNodes())
	assert.Nil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))

	newKey := Prefix{Addr{0x80000000}, 1}
	trie, err = trie.InsertOrUpdate(newKey, false)
	assert.Equal(t, 2, trie.NumNodes())
	assert.Nil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))
	assert.False(t, trie.Match(newKey).Data.(bool))

	inactiveKey := Prefix{}
	trie, err = trie.InsertOrUpdate(inactiveKey, "value")
	assert.Equal(t, 3, trie.NumNodes())
	assert.Nil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))
	assert.False(t, trie.Match(newKey).Data.(bool))
	assert.Equal(t, "value", trie.Match(inactiveKey).Data.(string))
}

func TestUpdateChangeValue32(t *testing.T) {
	var trie *trieNode32

	key := Prefix{}

	trie, err := trie.Insert(key, true)
	assert.Equal(t, 1, trie.NumNodes())
	assert.Nil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))

	trie, err = trie.Update(key, false)
	assert.Equal(t, 1, trie.NumNodes())
	assert.Nil(t, err)
	assert.False(t, trie.Match(key).Data.(bool))
}

func TestUpdateNewKey32(t *testing.T) {
	var trie *trieNode32

	key := Prefix{}

	trie, err := trie.Insert(key, true)
	assert.Equal(t, 1, trie.NumNodes())
	assert.Nil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))

	newKey := Prefix{Addr{0}, 1}
	trie, err = trie.Update(newKey, false)
	assert.Equal(t, 1, trie.NumNodes())
	assert.NotNil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))
	assert.True(t, trie.Match(newKey).Data.(bool))
}

func TestUpdateNarrowerKey32(t *testing.T) {
	var trie *trieNode32

	key := Prefix{Addr{0}, 1}

	trie, err := trie.Insert(key, true)
	assert.Equal(t, 1, trie.NumNodes())
	assert.Nil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))

	newKey := Prefix{}
	trie, err = trie.Update(newKey, false)
	assert.Equal(t, 1, trie.NumNodes())
	assert.NotNil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))
	assert.Nil(t, trie.Match(newKey))
}

func TestUpdateDisjointKeys32(t *testing.T) {
	var trie *trieNode32

	key := Prefix{Addr{0}, 1}

	trie, err := trie.Insert(key, true)
	assert.Equal(t, 1, trie.NumNodes())
	assert.Nil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))

	newKey := Prefix{Addr{0x80000000}, 1}
	trie, err = trie.Update(newKey, false)
	assert.Equal(t, 1, trie.NumNodes())
	assert.NotNil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))
	assert.Nil(t, trie.Match(newKey))
}

func TestUpdateInactive32(t *testing.T) {
	var trie *trieNode32

	key := Prefix{Addr{0}, 1}

	trie, err := trie.Insert(key, true)
	assert.Equal(t, 1, trie.NumNodes())
	assert.Nil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))

	newKey := Prefix{Addr{0x80000000}, 1}
	trie, err = trie.Insert(newKey, false)
	assert.Equal(t, 2, trie.NumNodes())
	assert.Nil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))
	assert.False(t, trie.Match(newKey).Data.(bool))

	inactiveKey := Prefix{}
	trie, err = trie.Update(inactiveKey, "value")
	assert.Equal(t, 2, trie.NumNodes())
	assert.NotNil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))
	assert.False(t, trie.Match(newKey).Data.(bool))
	assert.Nil(t, trie.Match(inactiveKey))
}

func TestMatchNilTrie32(t *testing.T) {
	var trie *trieNode32

	key := Prefix{}
	assert.Nil(t, trie.Match(key))
}

func TestMatchZeroLength32(t *testing.T) {
	var trie *trieNode32

	trie, err := trie.Insert(Prefix{
		Addr{0},
		0,
	}, nil)
	assert.Nil(t, err)
	assert.True(t, trie.active())
	assert.Equal(t, 1, trie.NumNodes())
	assert.Equal(t, 1, trie.height())

	assert.Equal(t, trie, trie.Match(Prefix{
		unsafeParseAddr("10.0.0.0"),
		0,
	}))
}

func TestGetOrInsertTrivial32(t *testing.T) {
	var trie *trieNode32
	assert.Equal(t, 0, trie.NumNodes())

	key := Prefix{Addr{0}, 0}

	trie, node := trie.GetOrInsert(key, true)
	assert.Equal(t, 1, trie.NumNodes())
	assert.Equal(t, trie, node)
	assert.True(t, node.Data.(bool))
}

func TestGetOrInsertExists32(t *testing.T) {
	var trie *trieNode32

	key := Prefix{Addr{0}, 0}

	trie, err := trie.Insert(key, true)
	assert.Nil(t, err)
	assert.Equal(t, 1, trie.NumNodes())

	trie, node := trie.GetOrInsert(key, false)

	assert.Equal(t, 1, trie.NumNodes())
	assert.Equal(t, trie, node)
	assert.True(t, node.Data.(bool))
}

func TestGetOrInsertBroader32(t *testing.T) {
	var trie *trieNode32

	existingKey := Prefix{unsafeParseAddr("10.224.0.0"), 16}
	trie, err := trie.Insert(existingKey, true)
	assert.Nil(t, err)
	assert.Equal(t, 1, trie.NumNodes())

	broaderKey := Prefix{unsafeParseAddr("10.0.0.0"), 8}
	trie, node := trie.GetOrInsert(broaderKey, false)

	assert.Equal(t, 2, trie.NumNodes())
	assert.Equal(t, trie, node)
	assert.False(t, node.Data.(bool))

	assert.True(t, trie.Match(existingKey).Data.(bool))
	assert.False(t, trie.Match(broaderKey).Data.(bool))
}

func TestGetOrInsertNarrower32(t *testing.T) {
	var trie *trieNode32

	existingKey := Prefix{unsafeParseAddr("10.224.0.0"), 16}
	trie, err := trie.Insert(existingKey, true)
	assert.Nil(t, err)
	assert.Equal(t, 1, trie.NumNodes())

	narrowerKey := Prefix{unsafeParseAddr("10.224.24.00"), 24}
	trie, node := trie.GetOrInsert(narrowerKey, false)

	assert.Equal(t, 2, trie.NumNodes())
	assert.NotEqual(t, trie, node)
	assert.False(t, node.Data.(bool))

	assert.True(t, trie.Match(existingKey).Data.(bool))
	assert.False(t, trie.Match(narrowerKey).Data.(bool))
}

func TestGetOrInsertDisjoint32(t *testing.T) {
	var trie *trieNode32

	existingKey := Prefix{unsafeParseAddr("10.224.0.0"), 16}
	trie, err := trie.Insert(existingKey, true)
	assert.Nil(t, err)
	assert.Equal(t, 1, trie.NumNodes())

	disjointKey := Prefix{unsafeParseAddr("10.225.0.0"), 16}
	trie, node := trie.GetOrInsert(disjointKey, false)

	assert.Equal(t, 2, trie.NumNodes())
	assert.False(t, node.Data.(bool))

	assert.True(t, trie.Match(existingKey).Data.(bool))
	assert.False(t, trie.Match(disjointKey).Data.(bool))
}

func TestGetOrInsertInActive32(t *testing.T) {
	var trie *trieNode32

	trie, _ = trie.Insert(Prefix{unsafeParseAddr("10.224.0.0"), 16}, true)
	trie, _ = trie.Insert(Prefix{unsafeParseAddr("10.225.0.0"), 16}, true)
	assert.Equal(t, 2, trie.NumNodes())

	trie, node := trie.GetOrInsert(Prefix{unsafeParseAddr("10.224.0.0"), 15}, false)
	assert.Equal(t, 3, trie.NumNodes())
	assert.Equal(t, trie, node)
	assert.False(t, node.Data.(bool))
}

func TestNoMatchTooBroad32(t *testing.T) {
	var trie *trieNode32

	trie, err := trie.Insert(Prefix{
		unsafeParseAddr("10.0.0.0"),
		24,
	}, nil)
	assert.Nil(t, err)
	assert.True(t, trie.active())
	assert.Equal(t, 1, trie.NumNodes())
	assert.Equal(t, 1, trie.height())

	assert.Nil(t, trie.Match(Prefix{
		unsafeParseAddr("10.0.0.0"),
		23,
	}))
}

func TestNoMatchPrefixMisMatch32(t *testing.T) {
	tests := []struct {
		desc         string
		nodeAddr     Addr
		nodeLength   uint32
		searchAddr   Addr
		searchLength uint32
	}{
		{
			desc:         "full bytes, mismatch in last byte",
			nodeAddr:     unsafeParseAddr("10.0.0.0"),
			nodeLength:   24,
			searchAddr:   unsafeParseAddr("10.0.1.0"),
			searchLength: 32,
		},
		{
			desc:         "full bytes, mismatch in earlier byte",
			nodeAddr:     unsafeParseAddr("10.0.0.0"),
			nodeLength:   24,
			searchAddr:   unsafeParseAddr("10.1.0.0"),
			searchLength: 32,
		},
		{
			desc:         "full bytes, mismatch in first byte",
			nodeAddr:     unsafeParseAddr("10.0.0.0"),
			nodeLength:   24,
			searchAddr:   unsafeParseAddr("11.0.0.0"),
			searchLength: 32,
		},
		{
			desc:         "mismatch in partial byte",
			nodeAddr:     unsafeParseAddr("10.0.0.0"),
			nodeLength:   27,
			searchAddr:   unsafeParseAddr("10.0.0.128"),
			searchLength: 32,
		},
		{
			desc:         "only one partial byte",
			nodeAddr:     Addr{},
			nodeLength:   7,
			searchAddr:   unsafeParseAddr("2.0.0.0"),
			searchLength: 8,
		},
		{
			desc:         "only one full byte",
			nodeAddr:     Addr{},
			nodeLength:   8,
			searchAddr:   unsafeParseAddr("10.0.0.0"),
			searchLength: 16,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			var trie *trieNode32

			trie, err := trie.Insert(Prefix{
				tt.nodeAddr,
				tt.nodeLength,
			}, nil)
			assert.Nil(t, err)
			assert.True(t, trie.active())
			assert.Equal(t, 1, trie.NumNodes())
			assert.Equal(t, 1, trie.height())

			assert.Nil(t, trie.Match(Prefix{
				tt.searchAddr,
				tt.searchLength,
			}))
		})
	}
}

func TestMatchSimplePrefixMatch32(t *testing.T) {
	tests := []struct {
		desc       string
		nodeAddr   Addr
		nodeLength uint32
	}{
		{
			desc:       "full bytes, mismatch in last byte",
			nodeAddr:   unsafeParseAddr("10.0.0.0"),
			nodeLength: 24,
		},
		{
			desc:       "full bytes, mismatch in earlier byte",
			nodeAddr:   unsafeParseAddr("10.0.0.0"),
			nodeLength: 24,
		},
		{
			desc:       "full bytes, mismatch in first byte",
			nodeAddr:   unsafeParseAddr("10.0.0.0"),
			nodeLength: 24,
		},
		{
			desc:       "mismatch in partial byte",
			nodeAddr:   unsafeParseAddr("10.0.0.0"),
			nodeLength: 27,
		},
		{
			desc:       "only one full byte",
			nodeAddr:   Addr{},
			nodeLength: 8,
		},
		{
			desc:       "partial byte",
			nodeAddr:   Addr{0xfe000000},
			nodeLength: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			var trie *trieNode32

			key := Prefix{
				tt.nodeAddr,
				tt.nodeLength,
			}
			trie, err := trie.Insert(key, nil)
			assert.Nil(t, err)
			assert.Equal(t, 1, trie.NumNodes())
			assert.Equal(t, 1, trie.height())

			assert := assert.New(t)
			assert.Equal(trie, trie.Match(key))
		})
	}
}

func TestMatchPartialByteMatches32(t *testing.T) {
	tests := []struct {
		nodeAddr   uint32
		nodeLength uint32
	}{
		{
			nodeAddr:   0x80000000,
			nodeLength: 1,
		},
		{
			nodeAddr:   0xc0000000,
			nodeLength: 2,
		},
		{
			nodeAddr:   0xe0000000,
			nodeLength: 3,
		},
		{
			nodeAddr:   0xf0000000,
			nodeLength: 4,
		},
		{
			nodeAddr:   0xf8000000,
			nodeLength: 5,
		},
		{
			nodeAddr:   0xfc000000,
			nodeLength: 6,
		},
		{
			nodeAddr:   0xfe000000,
			nodeLength: 7,
		},
		{
			nodeAddr:   0xff000000,
			nodeLength: 8,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d", tt.nodeLength), func(t *testing.T) {
			var trie *trieNode32

			key := Prefix{
				Addr{tt.nodeAddr},
				tt.nodeLength,
			}
			trie, err := trie.Insert(key, nil)
			assert.Nil(t, err)
			assert.True(t, trie.active())
			assert.Equal(t, 1, trie.NumNodes())
			assert.Equal(t, 1, trie.height())

			assert := assert.New(t)
			assert.Equal(trie, trie.Match(Prefix{
				// Always use 0xff to ensure that extraneous bits in the data are ignored
				Addr{0xff000000},
				tt.nodeLength,
			}))

			// byte with 0 in the last bit to match based on nodeLength
			var mismatch uint32
			mismatch = 0xff000000 & ^(0x80000000 >> (tt.nodeLength - 1))

			assert.Nil(trie.Match(Prefix{
				// Always use a byte with a 0 is the last matched bit
				Addr{mismatch},
				tt.nodeLength,
			}))
		})
	}
}

func TestInsertOverlapping32(t *testing.T) {
	tests := []struct {
		desc    string
		a, b, c Prefix
	}{
		{
			desc: "16 and 24",
			a:    Prefix{unsafeParseAddr("10.200.0.0"), 16},
			b:    Prefix{unsafeParseAddr("10.200.20.0"), 24},
			c:    Prefix{unsafeParseAddr("10.200.20.0"), 32},
		},
		{
			desc: "17 and 27",
			a:    Prefix{unsafeParseAddr("10.200.0.0"), 17},
			b:    Prefix{Addr{0x0ac800e0}, 27},
			c:    Prefix{Addr{0x0ac800f8}, 31},
		},
		{
			desc: "0 and 8",
			a:    Prefix{Addr{0}, 0},
			b:    Prefix{unsafeParseAddr("10.0.0.0"), 8},
			c:    Prefix{unsafeParseAddr("10.10.0.0"), 16},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			// This test inserts the three given nodes in the order given and
			// checks that they are found in the resulting trie
			subTest := func(first, second, third Prefix) func(t *testing.T) {
				return func(t *testing.T) {
					var trie *trieNode32

					trie, err := trie.Insert(first, nil)
					assert.Nil(t, err)
					assert.NotNil(t, trie.Match(first))
					assert.Equal(t, 1, trie.NumNodes())
					assert.Equal(t, 1, trie.height())

					trie, err = trie.Insert(second, nil)
					assert.Nil(t, err)
					assert.NotNil(t, trie.Match(second))
					assert.Equal(t, 2, trie.NumNodes())
					assert.Equal(t, 2, trie.height())

					trie, err = trie.Insert(third, nil)
					assert.Nil(t, err)
					assert.NotNil(t, trie.Match(third))
					assert.Equal(t, 3, trie.NumNodes())
					assert.Equal(t, 3, trie.height())
				}
			}
			t.Run("forward", subTest(tt.a, tt.b, tt.c))
			t.Run("backward", subTest(tt.c, tt.b, tt.a))

			// This sub-test tests that a node cannot be inserted twice
			insertDuplicate := func(key Prefix) func(t *testing.T) {
				return func(t *testing.T) {
					var trie *trieNode32

					trie, err := trie.Insert(key, nil)
					assert.Nil(t, err)
					assert.True(t, trie.active())
					assert.NotNil(t, trie)
					assert.Equal(t, 1, trie.NumNodes())
					assert.Equal(t, 1, trie.height())

					dup := key
					newTrie, err := trie.Insert(dup, nil)
					assert.NotNil(t, err)
					assert.Equal(t, trie, newTrie)
					assert.Equal(t, 1, trie.NumNodes())
					assert.Equal(t, 1, trie.height())
				}
			}
			t.Run("duplicate a", insertDuplicate(tt.a))
			t.Run("duplicate b", insertDuplicate(tt.b))
		})
	}
}

func TestInsertDisjoint32(t *testing.T) {
	tests := []struct {
		desc        string
		a, b, super Prefix
	}{
		{
			desc:  "first bit",
			a:     Prefix{Addr{0}, 1},
			b:     Prefix{unsafeParseAddr("128.0.0.0"), 1},
			super: Prefix{Addr{0}, 0},
		},
		{
			desc:  "seventeenth bit",
			a:     Prefix{unsafeParseAddr("10.224.0.0"), 17},
			b:     Prefix{unsafeParseAddr("10.224.128.0"), 17},
			super: Prefix{unsafeParseAddr("10.224.0.0"), 16},
		},
		{
			desc:  "partial b bit",
			a:     Prefix{unsafeParseAddr("10.224.0.0"), 23},
			b:     Prefix{unsafeParseAddr("10.224.8.0"), 23},
			super: Prefix{unsafeParseAddr("10.224.0.0"), 20},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			subTest := func(first, second Prefix) func(t *testing.T) {
				// This test inserts the two given nodes in the order given and
				// checks that they are both found in the resulting trie
				return func(t *testing.T) {
					var trie *trieNode32

					trie, err := trie.Insert(first, nil)
					assert.Nil(t, err)
					assert.Equal(t, trie.Prefix, first)
					assert.Equal(t, 1, trie.NumNodes())
					assert.Equal(t, 1, trie.height())

					trie, err = trie.Insert(second, nil)
					assert.Nil(t, err)
					assert.NotNil(t, trie.Match(second))
					assert.Equal(t, 2, trie.NumNodes())
					assert.Equal(t, 2, trie.height())

					assert.Nil(t, trie.Match(tt.super))

					// The following are testing a bit more of the internals
					// than I normally do.
					assert.False(t, trie.active())
					assert.Equal(t, tt.super, trie.Prefix)

					// insert an active node the same as `super` to turn it active
					trie, err = trie.Insert(tt.super, nil)
					assert.Nil(t, err)
					assert.NotNil(t, trie.Match(tt.super))
					assert.Equal(t, 3, trie.NumNodes())
					assert.Equal(t, 2, trie.height())
				}
			}
			t.Run("forward", subTest(tt.a, tt.b))
			t.Run("backward", subTest(tt.b, tt.a))
		})
	}
}

func TestInsertMoreComplex32(t *testing.T) {
	tests := []struct {
		desc string
		keys []Prefix
	}{
		{
			desc: "mix disjoint and overlapping",
			keys: []Prefix{
				Prefix{Addr{0}, 0},
				Prefix{Addr{0xff000000}, 8},
				Prefix{Addr{0xfe000000}, 8},
				Prefix{Addr{0xffff0000}, 16},
				Prefix{Addr{0xfffe0000}, 16},
				Prefix{Addr{0xffff0000}, 17},
				Prefix{Addr{0xfffe8000}, 17},
				Prefix{Addr{0xfffe8000}, 18},
				Prefix{Addr{0xffffb000}, 18},
				Prefix{Addr{0xfffebf00}, 24},
				Prefix{Addr{0xffffbe00}, 24},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			t.Run("forward", func(t *testing.T) {
				var trie *trieNode32

				for _, key := range tt.keys {
					var err error
					trie, err = trie.Insert(key, nil)
					assert.Nil(t, err)
					assert.NotNil(t, trie.Match(key))
				}
			})
			t.Run("backward", func(t *testing.T) {
				var trie *trieNode32

				for i := len(tt.keys); i != 0; i-- {
					var err error
					key := tt.keys[i-1]

					trie, err = trie.Insert(key, nil)
					assert.Nil(t, err)
					assert.NotNil(t, trie.Match(key))
				}
			})
		})
	}
}

func TestContains32(t *testing.T) {
	tests := []struct {
		desc           string
		a, b           Prefix
		matches, exact bool
		common, child  uint32
	}{
		{
			desc:    "trivial",
			a:       Prefix{Addr{0}, 0},
			b:       Prefix{Addr{0}, 0},
			matches: true,
			exact:   true,
			common:  0,
		},
		{
			desc:    "exact",
			a:       Prefix{unsafeParseAddr("10.0.0.0"), 16},
			b:       Prefix{unsafeParseAddr("10.0.0.0"), 16},
			matches: true,
			exact:   true,
			common:  16,
		},
		{
			desc:    "exact partial",
			a:       Prefix{unsafeParseAddr("10.0.0.0"), 19},
			b:       Prefix{Addr{0x0a001f00}, 19},
			matches: true,
			exact:   true,
			common:  19,
		},
		{
			desc:    "empty prefix match",
			a:       Prefix{Addr{0}, 0},
			b:       Prefix{unsafeParseAddr("10.10.0.0"), 16},
			matches: true,
			exact:   false,
			common:  0,
			child:   0,
		},
		{
			desc:    "empty prefix match backwards",
			a:       Prefix{Addr{0}, 0},
			b:       Prefix{unsafeParseAddr("130.10.0.0"), 16},
			matches: true,
			exact:   false,
			common:  0,
			child:   1,
		},
		{
			desc:    "matches",
			a:       Prefix{unsafeParseAddr("10.0.0.0"), 8},
			b:       Prefix{unsafeParseAddr("10.10.0.0"), 16},
			matches: true,
			exact:   false,
			common:  8,
			child:   0,
		},
		{
			desc:    "matches partial",
			a:       Prefix{unsafeParseAddr("10.200.0.0"), 9},
			b:       Prefix{unsafeParseAddr("10.129.0.0"), 16},
			matches: true,
			exact:   false,
			common:  9,
			child:   0,
		},
		{
			desc:    "matches backwards",
			a:       Prefix{unsafeParseAddr("10.0.0.0"), 8},
			b:       Prefix{unsafeParseAddr("10.200.0.0"), 16},
			matches: true,
			exact:   false,
			common:  8,
			child:   1,
		},
		{
			desc:    "matches backwards partial",
			a:       Prefix{unsafeParseAddr("10.240.0.0"), 9},
			b:       Prefix{unsafeParseAddr("10.200.0.0"), 16},
			matches: true,
			exact:   false,
			common:  9,
			child:   1,
		},
		{
			desc:    "disjoint",
			a:       Prefix{Addr{0}, 1},
			b:       Prefix{unsafeParseAddr("128.0.0.0"), 1},
			matches: false,
			common:  0,
			child:   1,
		},
		{
			desc:    "disjoint longer",
			a:       Prefix{unsafeParseAddr("0.0.0.0"), 17},
			b:       Prefix{unsafeParseAddr("0.0.128.0"), 17},
			matches: false,
			common:  16,
			child:   1,
		},
		{
			desc:    "disjoint longer partial",
			a:       Prefix{unsafeParseAddr("0.0.0.0"), 17},
			b:       Prefix{unsafeParseAddr("0.1.0.0"), 17},
			matches: false,
			common:  15,
			child:   1,
		},
		{
			desc:    "disjoint backwards",
			a:       Prefix{unsafeParseAddr("128.0.0.0"), 1},
			b:       Prefix{Addr{0}, 1},
			matches: false,
			common:  0,
			child:   0,
		},
		{
			desc:    "disjoint backwards longer",
			a:       Prefix{unsafeParseAddr("0.0.128.0"), 19},
			b:       Prefix{unsafeParseAddr("0.0.0.0"), 19},
			matches: false,
			common:  16,
			child:   0,
		},
		{
			desc:    "disjoint backwards longer partial",
			a:       Prefix{unsafeParseAddr("0.1.0.0"), 19},
			b:       Prefix{unsafeParseAddr("0.0.0.0"), 19},
			matches: false,
			common:  15,
			child:   0,
		},
		{
			desc:    "disjoint with common",
			a:       Prefix{unsafeParseAddr("10.0.0.0"), 16},
			b:       Prefix{unsafeParseAddr("10.10.0.0"), 16},
			matches: false,
			common:  12,
			child:   1,
		},
		{
			desc:    "disjoint with more disjoint bytes",
			a:       Prefix{unsafeParseAddr("0.255.255.0"), 24},
			b:       Prefix{unsafeParseAddr("128.0.0.0"), 24},
			matches: false,
			common:  0,
			child:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			matches, exact, common, child := contains32(tt.a, tt.b)
			assert.Equal(t, tt.matches, matches)
			assert.Equal(t, tt.exact, exact)
			assert.Equal(t, tt.common, common)
			assert.Equal(t, tt.child, child)

			// Opportunistically test the compare function
			t.Run("compare forward", func(t *testing.T) {
				_, _, reversed, _, _ := compare32(tt.a, tt.b)
				assert.False(t, reversed)
			})
			t.Run("compare reversed", func(t *testing.T) {
				_, _, reversed, _, _ := compare32(tt.b, tt.a)
				assert.Equal(t, tt.a.length != tt.b.length, reversed)
			})
		})
	}
}

func TestBitsToBytes32(t *testing.T) {
	tests := []struct {
		bits, bytes uint32
	}{
		{bits: 0, bytes: 0},
		{bits: 1, bytes: 1},
		{bits: 8, bytes: 1},
		{bits: 9, bytes: 2},
		{bits: 16, bytes: 2},
		{bits: 17, bytes: 3},
		{bits: 24, bytes: 3},
		{bits: 25, bytes: 4},
		{bits: 32, bytes: 4},
		{bits: 33, bytes: 5},
		{bits: 40, bytes: 5},
		{bits: 41, bytes: 6},
		{bits: 48, bytes: 6},
		{bits: 49, bytes: 7},
		{bits: 64, bytes: 8},
		{bits: 65, bytes: 9},
		{bits: 128, bytes: 16},
		{bits: 129, bytes: 17},
		{bits: 256, bytes: 32},
		{bits: 257, bytes: 33},
		{bits: 512, bytes: 64},
		{bits: 513, bytes: 65},
		{bits: 1024, bytes: 128},
		{bits: 1025, bytes: 129},
		{bits: 4096, bytes: 512},
		{bits: 4097, bytes: 513},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d", tt.bits), func(t *testing.T) {
			assert.Equal(t, tt.bytes, bitsToBytes(tt.bits))
		})
	}
}

func TestDeleteFromNilTree32(t *testing.T) {
	var trie *trieNode32

	key := Prefix{}
	trie, err := trie.Delete(key)
	assert.Nil(t, trie)
	assert.NotNil(t, err)
}

func TestDeleteSimple32(t *testing.T) {
	var trie *trieNode32

	key := Prefix{
		unsafeParseAddr("172.16.200.0"),
		24,
	}
	trie, err := trie.Insert(key, nil)
	trie, err = trie.Delete(key)
	assert.Nil(t, err)
	assert.Nil(t, trie)
}

func TestDeleteLeftChild32(t *testing.T) {
	var trie *trieNode32

	key := Prefix{
		unsafeParseAddr("172.16.200.0"),
		24,
	}
	trie, err := trie.Insert(key, nil)
	childKey := Prefix{
		unsafeParseAddr("172.16.200.0"),
		25,
	}
	trie, err = trie.Insert(childKey, nil)
	trie, err = trie.Delete(key)
	assert.Nil(t, err)
	assert.NotNil(t, trie)

	assert.Nil(t, trie.Match(key))
	assert.NotNil(t, trie.Match(childKey))
}

func TestDeleteRightChild32(t *testing.T) {
	var trie *trieNode32

	key := Prefix{
		unsafeParseAddr("172.16.200.0"),
		24,
	}
	trie, err := trie.Insert(key, nil)
	childKey := Prefix{
		unsafeParseAddr("172.16.200.128"),
		25,
	}
	trie, err = trie.Insert(childKey, nil)
	trie, err = trie.Delete(key)
	assert.Nil(t, err)
	assert.NotNil(t, trie)

	assert.Nil(t, trie.Match(key))
	assert.NotNil(t, trie.Match(childKey))
}

func TestDeleteBothChildren32(t *testing.T) {
	var trie *trieNode32

	key := Prefix{
		unsafeParseAddr("172.16.200.0"),
		24,
	}
	trie, err := trie.Insert(key, nil)
	leftChild := Prefix{
		unsafeParseAddr("172.16.200.0"),
		25,
	}
	trie, err = trie.Insert(leftChild, nil)
	rightChild := Prefix{
		unsafeParseAddr("172.16.200.128"),
		25,
	}
	trie, err = trie.Insert(rightChild, nil)
	trie, err = trie.Delete(key)
	assert.Nil(t, err)
	assert.NotNil(t, trie)

	assert.Nil(t, trie.Match(key))
	assert.NotNil(t, trie.Match(leftChild))
	assert.NotNil(t, trie.Match(rightChild))
}

func TestDeleteRecursiveNil32(t *testing.T) {
	var trie *trieNode32

	key := Prefix{
		unsafeParseAddr("172.16.200.0"),
		24,
	}
	trie, err := trie.Insert(key, nil)
	childKey := Prefix{
		unsafeParseAddr("172.16.200.0"),
		25,
	}
	trie, err = trie.Delete(childKey)
	assert.NotNil(t, err)
	assert.NotNil(t, trie)

	assert.NotNil(t, trie.Match(key))
	match := trie.Match(childKey)
	assert.NotEqual(t, childKey, match.Prefix)
	// assert.Nil(t, trie.Get(childKey))
}

func TestDeleteRecursiveLeftChild32(t *testing.T) {
	// NOTE: There's no specific test for other child combinations because I
	// didn't feel it added much. It uses already well-tested code paths.
	var trie *trieNode32

	key := Prefix{
		unsafeParseAddr("172.16.200.0"),
		24,
	}
	trie, err := trie.Insert(key, nil)
	childKey := Prefix{
		unsafeParseAddr("172.16.200.0"),
		25,
	}
	trie, err = trie.Insert(childKey, nil)
	trie, err = trie.Delete(childKey)
	assert.Nil(t, err)
	assert.NotNil(t, trie)

	assert.NotNil(t, trie.Match(key))
	match := trie.Match(childKey)
	assert.NotEqual(t, childKey, match.Prefix)
	// assert.Nil(t, trie.Get(childKey))
}

func TestDeleteKeyTooBroad32(t *testing.T) {
	var trie *trieNode32

	key := Prefix{
		unsafeParseAddr("172.16.200.0"),
		25,
	}
	trie, err := trie.Insert(key, nil)

	broadKey := Prefix{
		unsafeParseAddr("172.16.200.0"),
		24,
	}
	trie, err = trie.Delete(broadKey)
	assert.NotNil(t, err)
	assert.NotNil(t, trie)

	assert.NotNil(t, trie.Match(key))
	assert.Nil(t, trie.Match(broadKey))
}

func TestDeleteKeyDisjoint32(t *testing.T) {
	var trie *trieNode32

	key := Prefix{
		unsafeParseAddr("172.16.200.0"),
		25,
	}
	trie, err := trie.Insert(key, nil)

	disjointKey := Prefix{
		unsafeParseAddr("172.16.200.128"),
		25,
	}
	trie, err = trie.Delete(disjointKey)
	assert.NotNil(t, err)
	assert.NotNil(t, trie)

	assert.NotNil(t, trie.Match(key))
	assert.Nil(t, trie.Match(disjointKey))
}

func TestSuccessivelyBetter32(t *testing.T) {
	var trie *trieNode32

	keys := []Prefix{
		Prefix{unsafeParseAddr("10.224.24.0"), 0},
		Prefix{unsafeParseAddr("10.224.24.0"), 1},
		Prefix{unsafeParseAddr("10.224.24.0"), 8},
		Prefix{unsafeParseAddr("10.224.24.0"), 12},
		Prefix{unsafeParseAddr("10.224.24.0"), 16},
		Prefix{unsafeParseAddr("10.224.24.0"), 18},
		Prefix{unsafeParseAddr("10.224.24.0"), 20},
		Prefix{unsafeParseAddr("10.224.24.0"), 21},
		Prefix{unsafeParseAddr("10.224.24.0"), 22},
		Prefix{unsafeParseAddr("10.224.24.0"), 24},
		Prefix{unsafeParseAddr("10.224.24.0"), 27},
		Prefix{unsafeParseAddr("10.224.24.0"), 30},
		Prefix{unsafeParseAddr("10.224.24.0"), 32},
	}

	// Add successively more specific keys to the trie and assert that exact
	// matches are returned when appropriate and non-exact, but longest matches
	// are returned for the rest.
	for index, key := range keys {
		var err error
		trie, err = trie.Insert(key, nil)
		assert.Nil(t, err)
		assert.Equal(t, index+1, trie.NumNodes())
		assert.Equal(t, index+1, trie.height())

		for i, searchKey := range keys {
			node := trie.Match(searchKey)
			assert.NotNil(t, node)
			if i <= index {
				assert.Equal(t, searchKey, node.Prefix)
			} else {
				assert.Equal(t, keys[index], node.Prefix)
			}
		}
	}
	// Delete the nodes in the same order they were added and check that the
	// broader keys are no longer found in the trie as they're deleted but
	// the more specific ones are still found.
	for index, key := range keys {
		var err error
		trie, err = trie.Delete(key)
		assert.Nil(t, err)
		assert.Equal(t, len(keys)-index-1, trie.NumNodes())
		assert.Equal(t, len(keys)-index-1, trie.height())

		for i, searchKey := range keys {
			node := trie.Match(searchKey)
			if i <= index {
				assert.Nil(t, node)
			} else {
				assert.Equal(t, node.Prefix, searchKey)
			}
		}
	}
}

func TestIterate32(t *testing.T) {
	keys := []Prefix{
		Prefix{unsafeParseAddr("172.21.0.0"), 20},
		Prefix{unsafeParseAddr("192.68.27.0"), 25},
		Prefix{unsafeParseAddr("192.168.26.128"), 25},
		Prefix{unsafeParseAddr("10.224.24.0"), 32},
		Prefix{unsafeParseAddr("192.68.24.0"), 24},
		Prefix{unsafeParseAddr("172.16.0.0"), 12},
		Prefix{unsafeParseAddr("192.68.26.0"), 24},
		Prefix{unsafeParseAddr("10.224.24.0"), 30},
		Prefix{unsafeParseAddr("192.168.24.0"), 24},
		Prefix{unsafeParseAddr("192.168.25.0"), 24},
		Prefix{unsafeParseAddr("192.168.26.0"), 25},
		Prefix{unsafeParseAddr("192.68.25.0"), 24},
		Prefix{unsafeParseAddr("192.168.27.0"), 24},
		Prefix{unsafeParseAddr("172.20.128.0"), 19},
		Prefix{unsafeParseAddr("192.68.27.128"), 25},
	}

	golden := []Prefix{
		Prefix{unsafeParseAddr("10.224.24.0"), 30},
		Prefix{unsafeParseAddr("10.224.24.0"), 32},
		Prefix{unsafeParseAddr("172.16.0.0"), 12},
		Prefix{unsafeParseAddr("172.20.128.0"), 19},
		Prefix{unsafeParseAddr("172.21.0.0"), 20},
		Prefix{unsafeParseAddr("192.68.24.0"), 24},
		Prefix{unsafeParseAddr("192.68.25.0"), 24},
		Prefix{unsafeParseAddr("192.68.26.0"), 24},
		Prefix{unsafeParseAddr("192.68.27.0"), 25},
		Prefix{unsafeParseAddr("192.68.27.128"), 25},
		Prefix{unsafeParseAddr("192.168.24.0"), 24},
		Prefix{unsafeParseAddr("192.168.25.0"), 24},
		Prefix{unsafeParseAddr("192.168.26.0"), 25},
		Prefix{unsafeParseAddr("192.168.26.128"), 25},
		Prefix{unsafeParseAddr("192.168.27.0"), 24},
	}

	var trie *trieNode32
	check := func(t *testing.T) {
		result := []Prefix{}
		trie.Iterate(func(key Prefix, _ interface{}) bool {
			result = append(result, key)
			return true
		})
		assert.Equal(t, golden, result)

		iterations := 0
		trie.Iterate(func(key Prefix, _ interface{}) bool {
			iterations++
			return false
		})
		assert.Equal(t, 1, iterations)

		// Just ensure that iterating with a nil callback doesn't crash
		trie.Iterate(nil)
	}

	t.Run("normal insert", func(t *testing.T) {
		trie = nil
		for _, key := range keys {
			trie, _ = trie.Insert(key, nil)
		}
		check(t)
	})
	t.Run("get or insert", func(t *testing.T) {
		trie = nil
		for _, key := range keys {
			trie, _ = trie.GetOrInsert(key, nil)
		}
		check(t)
	})
}

type pair32 struct {
	key  Prefix
	data interface{}
}

func printTrie32(trie *trieNode32) {
	var recurse func(trie *trieNode32, level int)

	recurse = func(trie *trieNode32, level int) {
		if trie == nil {
			return
		}
		for i := 0; i < level; i++ {
			fmt.Printf("    ")
		}
		fmt.Printf("%+v\n", trie)
		recurse(trie.children[0], level+1)
		recurse(trie.children[1], level+1)
	}

	recurse(trie, 0)
}

func TestAggregate32(t *testing.T) {
	tests := []struct {
		desc   string
		pairs  []pair32
		golden []pair32
	}{
		{
			desc:   "nothing",
			pairs:  []pair32{},
			golden: []pair32{},
		},
		{
			desc: "simple aggregation",
			pairs: []pair32{
				pair32{key: Prefix{unsafeParseAddr("10.224.24.2"), 31}},
				pair32{key: Prefix{unsafeParseAddr("10.224.24.1"), 32}},
				pair32{key: Prefix{unsafeParseAddr("10.224.24.0"), 32}},
			},
			golden: []pair32{
				pair32{key: Prefix{unsafeParseAddr("10.224.24.0"), 30}},
			},
		},
		{
			desc: "same as iterate",
			pairs: []pair32{
				pair32{key: Prefix{unsafeParseAddr("172.21.0.0"), 20}},
				pair32{key: Prefix{unsafeParseAddr("192.68.27.0"), 25}},
				pair32{key: Prefix{unsafeParseAddr("192.168.26.128"), 25}},
				pair32{key: Prefix{unsafeParseAddr("10.224.24.0"), 32}},
				pair32{key: Prefix{unsafeParseAddr("192.68.24.0"), 24}},
				pair32{key: Prefix{unsafeParseAddr("172.16.0.0"), 12}},
				pair32{key: Prefix{unsafeParseAddr("192.68.26.0"), 24}},
				pair32{key: Prefix{unsafeParseAddr("10.224.24.0"), 30}},
				pair32{key: Prefix{unsafeParseAddr("192.168.24.0"), 24}},
				pair32{key: Prefix{unsafeParseAddr("192.168.25.0"), 24}},
				pair32{key: Prefix{unsafeParseAddr("192.168.26.0"), 25}},
				pair32{key: Prefix{unsafeParseAddr("192.68.25.0"), 24}},
				pair32{key: Prefix{unsafeParseAddr("192.168.27.0"), 24}},
				pair32{key: Prefix{unsafeParseAddr("172.20.128.0"), 19}},
				pair32{key: Prefix{unsafeParseAddr("192.68.27.128"), 25}},
			},
			golden: []pair32{
				pair32{key: Prefix{unsafeParseAddr("10.224.24.0"), 30}},
				pair32{key: Prefix{unsafeParseAddr("172.16.0.0"), 12}},
				pair32{key: Prefix{unsafeParseAddr("192.68.24.0"), 22}},
				pair32{key: Prefix{unsafeParseAddr("192.168.24.0"), 22}},
			},
		},
		{
			desc: "mixed umbrellas",
			pairs: []pair32{
				pair32{key: Prefix{unsafeParseAddr("10.224.24.0"), 30}, data: true},
				pair32{key: Prefix{unsafeParseAddr("10.224.24.0"), 31}, data: false},
				pair32{key: Prefix{unsafeParseAddr("10.224.24.1"), 32}, data: true},
				pair32{key: Prefix{unsafeParseAddr("10.224.24.0"), 32}, data: false},
			},
			golden: []pair32{
				pair32{key: Prefix{unsafeParseAddr("10.224.24.0"), 30}, data: true},
				pair32{key: Prefix{unsafeParseAddr("10.224.24.0"), 31}, data: false},
				pair32{key: Prefix{unsafeParseAddr("10.224.24.1"), 32}, data: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			var trie *trieNode32
			check := func(t *testing.T) {
				expectedIterations := 0
				result := []pair32{}
				trie.Aggregate(
					func(key Prefix, data interface{}) bool {
						result = append(result, pair32{key: key, data: data})
						expectedIterations = 1
						return true
					},
				)
				assert.Equal(t, tt.golden, result)

				iterations := 0
				trie.Aggregate(
					func(key Prefix, data interface{}) bool {
						result = append(result, pair32{key: key, data: data})
						iterations++
						return false
					},
				)
				assert.Equal(t, expectedIterations, iterations)
			}

			t.Run("normal insert", func(t *testing.T) {
				for _, p := range tt.pairs {
					trie, _ = trie.Insert(p.key, p.data)
				}
				check(t)
			})
			t.Run("get or insert", func(t *testing.T) {
				for _, p := range tt.pairs {
					trie, _ = trie.GetOrInsert(p.key, p.data)
				}
				check(t)
			})
		})
	}
}

type comparable struct {
	// Begin with a type (slice) that is not comparable with standard ==
	data []string
}

func (me *comparable) EqualInterface(other interface{}) bool {
	return reflect.DeepEqual(me, other)
}

// Like the TestAggregate above but using a type that is comparable through the
// EqualComparable interface.
func TestAggregateEqualComparable32(t *testing.T) {
	NextHop1 := &comparable{data: []string{"10.224.24.1"}}
	NextHop2 := &comparable{data: []string{"10.224.24.111"}}
	tests := []struct {
		desc   string
		pairs  []pair32
		golden []pair32
	}{
		{
			desc: "mixed umbrellas",
			pairs: []pair32{
				pair32{key: Prefix{unsafeParseAddr("10.224.24.0"), 30}, data: NextHop1},
				pair32{key: Prefix{unsafeParseAddr("10.224.24.0"), 31}, data: NextHop2},
				pair32{key: Prefix{unsafeParseAddr("10.224.24.1"), 32}, data: NextHop1},
				pair32{key: Prefix{unsafeParseAddr("10.224.24.0"), 32}, data: NextHop2},
			},
			golden: []pair32{
				pair32{key: Prefix{unsafeParseAddr("10.224.24.0"), 30}, data: NextHop1},
				pair32{key: Prefix{unsafeParseAddr("10.224.24.0"), 31}, data: NextHop2},
				pair32{key: Prefix{unsafeParseAddr("10.224.24.1"), 32}, data: NextHop1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			var trie *trieNode32
			for _, p := range tt.pairs {
				trie, _ = trie.Insert(p.key, p.data)
			}

			result := []pair32{}
			trie.Aggregate(
				func(key Prefix, data interface{}) bool {
					result = append(result, pair32{key: key, data: data})
					return true
				},
			)
			assert.Equal(t, tt.golden, result)
		})
	}
}
