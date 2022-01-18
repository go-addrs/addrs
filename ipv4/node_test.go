package ipv4

import (
	"fmt"
	"reflect"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestActive(t *testing.T) {
	var node *trieNode
	assert.False(t, node.active())
	assert.False(t, (&trieNode{}).active())
	assert.True(t, (&trieNode{isActive: true}).active())
}

func TestStructSizes(t *testing.T) {
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

	node := trieNode{}
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

func TestMatchNilKey(t *testing.T) {
	var trie *trieNode
	var key Prefix

	assert.Nil(t, trie.Match(key))
}

func TestInsertOrUpdateChangeValue(t *testing.T) {
	var trie *trieNode

	key := Prefix{}

	trie, err := trie.InsertOrUpdate(key, true)
	assert.True(t, trie.isValid())
	assert.Nil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))

	trie, err = trie.InsertOrUpdate(key, false)
	assert.True(t, trie.isValid())
	assert.Nil(t, err)
	assert.False(t, trie.Match(key).Data.(bool))
}

func TestInsertOrUpdateNewKey(t *testing.T) {
	var trie *trieNode

	key := Prefix{}

	trie, err := trie.InsertOrUpdate(key, true)
	assert.True(t, trie.isValid())
	assert.Nil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))

	newKey := Prefix{Address{0}, 1}
	trie, err = trie.InsertOrUpdate(newKey, false)
	assert.True(t, trie.isValid())
	assert.Nil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))
	assert.False(t, trie.Match(newKey).Data.(bool))
}

func TestInsertOrUpdateNarrowerKey(t *testing.T) {
	var trie *trieNode

	key := Prefix{Address{0}, 1}

	trie, err := trie.InsertOrUpdate(key, true)
	assert.True(t, trie.isValid())
	assert.Nil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))

	newKey := Prefix{}
	trie, err = trie.InsertOrUpdate(newKey, false)
	assert.True(t, trie.isValid())
	assert.Nil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))
	assert.False(t, trie.Match(newKey).Data.(bool))
}

func TestInsertOrUpdateDisjointKeys(t *testing.T) {
	var trie *trieNode

	key := Prefix{Address{0}, 1}

	trie, err := trie.InsertOrUpdate(key, true)
	assert.True(t, trie.isValid())
	assert.Nil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))

	newKey := Prefix{Address{0x80000000}, 1}
	trie, err = trie.InsertOrUpdate(newKey, false)
	assert.True(t, trie.isValid())
	assert.Nil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))
	assert.False(t, trie.Match(newKey).Data.(bool))
}

func TestInsertOrUpdateInactive(t *testing.T) {
	var trie *trieNode

	key := Prefix{Address{0}, 1}

	trie, err := trie.InsertOrUpdate(key, true)
	assert.True(t, trie.isValid())
	assert.Nil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))

	newKey := Prefix{Address{0x80000000}, 1}
	trie, err = trie.InsertOrUpdate(newKey, false)
	assert.True(t, trie.isValid())
	assert.Nil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))
	assert.False(t, trie.Match(newKey).Data.(bool))

	inactiveKey := Prefix{}
	trie, err = trie.InsertOrUpdate(inactiveKey, "value")
	assert.True(t, trie.isValid())
	assert.Nil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))
	assert.False(t, trie.Match(newKey).Data.(bool))
	assert.Equal(t, "value", trie.Match(inactiveKey).Data.(string))
}

func TestUpdateChangeValue(t *testing.T) {
	var trie *trieNode

	key := Prefix{}

	trie, err := trie.Insert(key, true)
	assert.True(t, trie.isValid())
	assert.Nil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))

	trie, err = trie.Update(key, false)
	assert.True(t, trie.isValid())
	assert.Nil(t, err)
	assert.False(t, trie.Match(key).Data.(bool))
}

func TestUpdateNewKey(t *testing.T) {
	var trie *trieNode

	key := Prefix{}

	trie, err := trie.Insert(key, true)
	assert.True(t, trie.isValid())
	assert.Nil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))

	newKey := Prefix{Address{0}, 1}
	trie, err = trie.Update(newKey, false)
	assert.True(t, trie.isValid())
	assert.NotNil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))
	assert.True(t, trie.Match(newKey).Data.(bool))
}

func TestUpdateNarrowerKey(t *testing.T) {
	var trie *trieNode

	key := Prefix{Address{0}, 1}

	trie, err := trie.Insert(key, true)
	assert.True(t, trie.isValid())
	assert.Nil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))

	newKey := Prefix{}
	trie, err = trie.Update(newKey, false)
	assert.True(t, trie.isValid())
	assert.NotNil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))
	assert.Nil(t, trie.Match(newKey))
}

func TestUpdateDisjointKeys(t *testing.T) {
	var trie *trieNode

	key := Prefix{Address{0}, 1}

	trie, err := trie.Insert(key, true)
	assert.True(t, trie.isValid())
	assert.Nil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))

	newKey := Prefix{Address{0x80000000}, 1}
	trie, err = trie.Update(newKey, false)
	assert.True(t, trie.isValid())
	assert.NotNil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))
	assert.Nil(t, trie.Match(newKey))
}

func TestUpdateInactive(t *testing.T) {
	var trie *trieNode

	key := Prefix{Address{0}, 1}

	trie, err := trie.Insert(key, true)
	assert.True(t, trie.isValid())
	assert.Nil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))

	newKey := Prefix{Address{0x80000000}, 1}
	trie, err = trie.Insert(newKey, false)
	assert.True(t, trie.isValid())
	assert.Nil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))
	assert.False(t, trie.Match(newKey).Data.(bool))

	inactiveKey := Prefix{}
	trie, err = trie.Update(inactiveKey, "value")
	assert.True(t, trie.isValid())
	assert.NotNil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))
	assert.False(t, trie.Match(newKey).Data.(bool))
	assert.Nil(t, trie.Match(inactiveKey))
}

func TestMatchNilTrie(t *testing.T) {
	var trie *trieNode

	key := Prefix{}
	assert.Nil(t, trie.Match(key))
}

func TestMatchZeroLength(t *testing.T) {
	var trie *trieNode

	trie, err := trie.Insert(Prefix{
		Address{0},
		0,
	}, nil)
	assert.Nil(t, err)
	assert.True(t, trie.active())
	assert.True(t, trie.isValid())
	assert.Equal(t, 1, trie.height())

	assert.Equal(t, trie, trie.Match(Prefix{
		unsafeParseAddress("10.0.0.0"),
		0,
	}))
}

func TestGetOrInsertTrivial(t *testing.T) {
	var trie *trieNode
	assert.Equal(t, int64(0), trie.NumNodes())
	assert.True(t, trie.isValid())

	key := Prefix{Address{0}, 0}

	trie, node := trie.GetOrInsert(key, true)
	assert.True(t, trie.isValid())
	assert.Equal(t, trie, node)
	assert.True(t, node.Data.(bool))
}

func TestGetOrInsertExists(t *testing.T) {
	var trie *trieNode

	key := Prefix{Address{0}, 0}

	trie, err := trie.Insert(key, true)
	assert.Nil(t, err)
	assert.True(t, trie.isValid())

	trie, node := trie.GetOrInsert(key, false)

	assert.True(t, trie.isValid())
	assert.Equal(t, trie, node)
	assert.True(t, node.Data.(bool))
}

func TestGetOrInsertBroader(t *testing.T) {
	var trie *trieNode

	existingKey := Prefix{unsafeParseAddress("10.224.0.0"), 16}
	trie, err := trie.Insert(existingKey, true)
	assert.Nil(t, err)
	assert.True(t, trie.isValid())

	broaderKey := Prefix{unsafeParseAddress("10.0.0.0"), 8}
	trie, node := trie.GetOrInsert(broaderKey, false)

	assert.True(t, trie.isValid())
	assert.Equal(t, trie, node)
	assert.False(t, node.Data.(bool))

	assert.True(t, trie.Match(existingKey).Data.(bool))
	assert.False(t, trie.Match(broaderKey).Data.(bool))
}

func TestGetOrInsertNarrower(t *testing.T) {
	var trie *trieNode

	existingKey := Prefix{unsafeParseAddress("10.224.0.0"), 16}
	trie, err := trie.Insert(existingKey, true)
	assert.Nil(t, err)
	assert.True(t, trie.isValid())

	narrowerKey := Prefix{unsafeParseAddress("10.224.24.0"), 24}
	trie, node := trie.GetOrInsert(narrowerKey, false)

	assert.True(t, trie.isValid())
	assert.NotEqual(t, trie, node)
	assert.False(t, node.Data.(bool))

	assert.True(t, trie.Match(existingKey).Data.(bool))
	assert.False(t, trie.Match(narrowerKey).Data.(bool))
}

func TestGetOrInsertDisjoint(t *testing.T) {
	var trie *trieNode

	existingKey := Prefix{unsafeParseAddress("10.224.0.0"), 16}
	trie, err := trie.Insert(existingKey, true)
	assert.Nil(t, err)
	assert.True(t, trie.isValid())

	disjointKey := Prefix{unsafeParseAddress("10.225.0.0"), 16}
	trie, node := trie.GetOrInsert(disjointKey, false)

	assert.True(t, trie.isValid())
	assert.False(t, node.Data.(bool))

	assert.True(t, trie.Match(existingKey).Data.(bool))
	assert.False(t, trie.Match(disjointKey).Data.(bool))
}

func TestGetOrInsertInActive(t *testing.T) {
	var trie *trieNode

	trie, _ = trie.Insert(Prefix{unsafeParseAddress("10.224.0.0"), 16}, true)
	trie, _ = trie.Insert(Prefix{unsafeParseAddress("10.225.0.0"), 16}, true)
	assert.True(t, trie.isValid())

	trie, node := trie.GetOrInsert(Prefix{unsafeParseAddress("10.224.0.0"), 15}, false)
	assert.True(t, trie.isValid())
	assert.Equal(t, trie, node)
	assert.False(t, node.Data.(bool))
}

func TestNoMatchTooBroad(t *testing.T) {
	var trie *trieNode

	trie, err := trie.Insert(Prefix{
		unsafeParseAddress("10.0.0.0"),
		24,
	}, nil)
	assert.Nil(t, err)
	assert.True(t, trie.active())
	assert.True(t, trie.isValid())
	assert.Equal(t, 1, trie.height())

	assert.Nil(t, trie.Match(Prefix{
		unsafeParseAddress("10.0.0.0"),
		23,
	}))
}

func TestNoMatchPrefixMisMatch(t *testing.T) {
	tests := []struct {
		desc          string
		nodeAddress   Address
		nodeLength    uint32
		searchAddress Address
		searchLength  uint32
	}{
		{
			desc:          "full bytes, mismatch in last byte",
			nodeAddress:   unsafeParseAddress("10.0.0.0"),
			nodeLength:    24,
			searchAddress: unsafeParseAddress("10.0.1.0"),
			searchLength:  32,
		},
		{
			desc:          "full bytes, mismatch in earlier byte",
			nodeAddress:   unsafeParseAddress("10.0.0.0"),
			nodeLength:    24,
			searchAddress: unsafeParseAddress("10.1.0.0"),
			searchLength:  32,
		},
		{
			desc:          "full bytes, mismatch in first byte",
			nodeAddress:   unsafeParseAddress("10.0.0.0"),
			nodeLength:    24,
			searchAddress: unsafeParseAddress("11.0.0.0"),
			searchLength:  32,
		},
		{
			desc:          "mismatch in partial byte",
			nodeAddress:   unsafeParseAddress("10.0.0.0"),
			nodeLength:    27,
			searchAddress: unsafeParseAddress("10.0.0.128"),
			searchLength:  32,
		},
		{
			desc:          "only one partial byte",
			nodeAddress:   Address{},
			nodeLength:    7,
			searchAddress: unsafeParseAddress("2.0.0.0"),
			searchLength:  8,
		},
		{
			desc:          "only one full byte",
			nodeAddress:   Address{},
			nodeLength:    8,
			searchAddress: unsafeParseAddress("10.0.0.0"),
			searchLength:  16,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			var trie *trieNode

			trie, err := trie.Insert(Prefix{
				tt.nodeAddress,
				tt.nodeLength,
			}, nil)
			assert.Nil(t, err)
			assert.True(t, trie.active())
			assert.True(t, trie.isValid())
			assert.Equal(t, 1, trie.height())

			assert.Nil(t, trie.Match(Prefix{
				tt.searchAddress,
				tt.searchLength,
			}))
		})
	}
}

func TestMatchSimplePrefixMatch(t *testing.T) {
	tests := []struct {
		desc        string
		nodeAddress Address
		nodeLength  uint32
	}{
		{
			desc:        "full bytes, mismatch in last byte",
			nodeAddress: unsafeParseAddress("10.0.0.0"),
			nodeLength:  24,
		},
		{
			desc:        "full bytes, mismatch in earlier byte",
			nodeAddress: unsafeParseAddress("10.0.0.0"),
			nodeLength:  24,
		},
		{
			desc:        "full bytes, mismatch in first byte",
			nodeAddress: unsafeParseAddress("10.0.0.0"),
			nodeLength:  24,
		},
		{
			desc:        "mismatch in partial byte",
			nodeAddress: unsafeParseAddress("10.0.0.0"),
			nodeLength:  27,
		},
		{
			desc:        "only one full byte",
			nodeAddress: Address{},
			nodeLength:  8,
		},
		{
			desc:        "partial byte",
			nodeAddress: Address{0xfe000000},
			nodeLength:  7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			var trie *trieNode

			key := Prefix{
				tt.nodeAddress,
				tt.nodeLength,
			}
			trie, err := trie.Insert(key, nil)
			assert.Nil(t, err)
			assert.True(t, trie.isValid())
			assert.Equal(t, 1, trie.height())

			assert := assert.New(t)
			assert.Equal(trie, trie.Match(key))
		})
	}
}

func TestMatchPartialByteMatches(t *testing.T) {
	tests := []struct {
		nodeAddress uint32
		nodeLength  uint32
	}{
		{
			nodeAddress: 0x80000000,
			nodeLength:  1,
		},
		{
			nodeAddress: 0xc0000000,
			nodeLength:  2,
		},
		{
			nodeAddress: 0xe0000000,
			nodeLength:  3,
		},
		{
			nodeAddress: 0xf0000000,
			nodeLength:  4,
		},
		{
			nodeAddress: 0xf8000000,
			nodeLength:  5,
		},
		{
			nodeAddress: 0xfc000000,
			nodeLength:  6,
		},
		{
			nodeAddress: 0xfe000000,
			nodeLength:  7,
		},
		{
			nodeAddress: 0xff000000,
			nodeLength:  8,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d", tt.nodeLength), func(t *testing.T) {
			var trie *trieNode

			key := Prefix{
				Address{tt.nodeAddress},
				tt.nodeLength,
			}
			trie, err := trie.Insert(key, nil)
			assert.Nil(t, err)
			assert.True(t, trie.active())
			assert.True(t, trie.isValid())
			assert.Equal(t, 1, trie.height())

			assert := assert.New(t)
			assert.Equal(trie, trie.Match(Prefix{
				// Always use 0xff to ensure that extraneous bits in the data are ignored
				Address{0xff000000},
				tt.nodeLength,
			}))

			// byte with 0 in the last bit to match based on nodeLength
			var mismatch uint32
			mismatch = 0xff000000 & ^(0x80000000 >> (tt.nodeLength - 1))

			assert.Nil(trie.Match(Prefix{
				// Always use a byte with a 0 is the last matched bit
				Address{mismatch},
				tt.nodeLength,
			}))
		})
	}
}

func TestInsertOverlapping(t *testing.T) {
	tests := []struct {
		desc    string
		a, b, c Prefix
	}{
		{
			desc: "16 and 24",
			a:    Prefix{unsafeParseAddress("10.200.0.0"), 16},
			b:    Prefix{unsafeParseAddress("10.200.20.0"), 24},
			c:    Prefix{unsafeParseAddress("10.200.20.0"), 32},
		},
		{
			desc: "17 and 27",
			a:    Prefix{unsafeParseAddress("10.200.0.0"), 17},
			b:    Prefix{Address{0x0ac800e0}, 27},
			c:    Prefix{Address{0x0ac800f8}, 31},
		},
		{
			desc: "0 and 8",
			a:    Prefix{Address{0}, 0},
			b:    Prefix{unsafeParseAddress("10.0.0.0"), 8},
			c:    Prefix{unsafeParseAddress("10.10.0.0"), 16},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			// This test inserts the three given nodes in the order given and
			// checks that they are found in the resulting trie
			subTest := func(first, second, third Prefix) func(t *testing.T) {
				return func(t *testing.T) {
					var trie *trieNode

					trie, err := trie.Insert(first, nil)
					assert.Nil(t, err)
					assert.NotNil(t, trie.Match(first))
					assert.True(t, trie.isValid())
					assert.Equal(t, 1, trie.height())

					trie, err = trie.Insert(second, nil)
					assert.Nil(t, err)
					assert.NotNil(t, trie.Match(second))
					assert.True(t, trie.isValid())
					assert.Equal(t, 2, trie.height())

					trie, err = trie.Insert(third, nil)
					assert.Nil(t, err)
					assert.NotNil(t, trie.Match(third))
					assert.True(t, trie.isValid())
					assert.Equal(t, 3, trie.height())
				}
			}
			t.Run("forward", subTest(tt.a, tt.b, tt.c))
			t.Run("backward", subTest(tt.c, tt.b, tt.a))

			// This sub-test tests that a node cannot be inserted twice
			insertDuplicate := func(key Prefix) func(t *testing.T) {
				return func(t *testing.T) {
					var trie *trieNode

					trie, err := trie.Insert(key, nil)
					assert.Nil(t, err)
					assert.True(t, trie.active())
					assert.NotNil(t, trie)
					assert.True(t, trie.isValid())
					assert.Equal(t, 1, trie.height())

					dup := key
					newTrie, err := trie.Insert(dup, nil)
					assert.NotNil(t, err)
					assert.Equal(t, trie, newTrie)
					assert.True(t, trie.isValid())
					assert.Equal(t, 1, trie.height())
				}
			}
			t.Run("duplicate a", insertDuplicate(tt.a))
			t.Run("duplicate b", insertDuplicate(tt.b))
		})
	}
}

func TestInsertDisjoint(t *testing.T) {
	tests := []struct {
		desc        string
		a, b, super Prefix
	}{
		{
			desc:  "first bit",
			a:     Prefix{Address{0}, 1},
			b:     Prefix{unsafeParseAddress("128.0.0.0"), 1},
			super: Prefix{Address{0}, 0},
		},
		{
			desc:  "seventeenth bit",
			a:     Prefix{unsafeParseAddress("10.224.0.0"), 17},
			b:     Prefix{unsafeParseAddress("10.224.128.0"), 17},
			super: Prefix{unsafeParseAddress("10.224.0.0"), 16},
		},
		{
			desc:  "partial b bit",
			a:     Prefix{unsafeParseAddress("10.224.0.0"), 23},
			b:     Prefix{unsafeParseAddress("10.224.8.0"), 23},
			super: Prefix{unsafeParseAddress("10.224.0.0"), 20},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			subTest := func(first, second Prefix) func(t *testing.T) {
				// This test inserts the two given nodes in the order given and
				// checks that they are both found in the resulting trie
				return func(t *testing.T) {
					var trie *trieNode

					trie, err := trie.Insert(first, nil)
					assert.Nil(t, err)
					assert.Equal(t, trie.Prefix, first)
					assert.True(t, trie.isValid())
					assert.Equal(t, 1, trie.height())

					trie, err = trie.Insert(second, nil)
					assert.Nil(t, err)
					assert.NotNil(t, trie.Match(second))
					assert.True(t, trie.isValid())
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
					assert.True(t, trie.isValid())
					assert.Equal(t, 2, trie.height())
				}
			}
			t.Run("forward", subTest(tt.a, tt.b))
			t.Run("backward", subTest(tt.b, tt.a))
		})
	}
}

func TestInsertMoreComplex(t *testing.T) {
	tests := []struct {
		desc string
		keys []Prefix
	}{
		{
			desc: "mix disjoint and overlapping",
			keys: []Prefix{
				Prefix{Address{0}, 0},
				Prefix{Address{0xff000000}, 8},
				Prefix{Address{0xfe000000}, 8},
				Prefix{Address{0xffff0000}, 16},
				Prefix{Address{0xfffe0000}, 16},
				Prefix{Address{0xffff0000}, 17},
				Prefix{Address{0xfffe8000}, 17},
				Prefix{Address{0xfffe8000}, 18},
				Prefix{Address{0xffffb000}, 18},
				Prefix{Address{0xfffebf00}, 24},
				Prefix{Address{0xffffbe00}, 24},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			t.Run("forward", func(t *testing.T) {
				var trie *trieNode

				for _, key := range tt.keys {
					var err error
					trie, err = trie.Insert(key, nil)
					assert.Nil(t, err)
					assert.NotNil(t, trie.Match(key))
				}
			})
			t.Run("backward", func(t *testing.T) {
				var trie *trieNode

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

func TestContains(t *testing.T) {
	tests := []struct {
		desc           string
		a, b           Prefix
		matches, exact bool
		common, child  uint32
	}{
		{
			desc:    "trivial",
			a:       Prefix{Address{0}, 0},
			b:       Prefix{Address{0}, 0},
			matches: true,
			exact:   true,
			common:  0,
		},
		{
			desc:    "exact",
			a:       Prefix{unsafeParseAddress("10.0.0.0"), 16},
			b:       Prefix{unsafeParseAddress("10.0.0.0"), 16},
			matches: true,
			exact:   true,
			common:  16,
		},
		{
			desc:    "exact partial",
			a:       Prefix{unsafeParseAddress("10.0.0.0"), 19},
			b:       Prefix{Address{0x0a001f00}, 19},
			matches: true,
			exact:   true,
			common:  19,
		},
		{
			desc:    "empty prefix match",
			a:       Prefix{Address{0}, 0},
			b:       Prefix{unsafeParseAddress("10.10.0.0"), 16},
			matches: true,
			exact:   false,
			common:  0,
			child:   0,
		},
		{
			desc:    "empty prefix match backwards",
			a:       Prefix{Address{0}, 0},
			b:       Prefix{unsafeParseAddress("130.10.0.0"), 16},
			matches: true,
			exact:   false,
			common:  0,
			child:   1,
		},
		{
			desc:    "matches",
			a:       Prefix{unsafeParseAddress("10.0.0.0"), 8},
			b:       Prefix{unsafeParseAddress("10.10.0.0"), 16},
			matches: true,
			exact:   false,
			common:  8,
			child:   0,
		},
		{
			desc:    "matches partial",
			a:       Prefix{unsafeParseAddress("10.200.0.0"), 9},
			b:       Prefix{unsafeParseAddress("10.129.0.0"), 16},
			matches: true,
			exact:   false,
			common:  9,
			child:   0,
		},
		{
			desc:    "matches backwards",
			a:       Prefix{unsafeParseAddress("10.0.0.0"), 8},
			b:       Prefix{unsafeParseAddress("10.200.0.0"), 16},
			matches: true,
			exact:   false,
			common:  8,
			child:   1,
		},
		{
			desc:    "matches backwards partial",
			a:       Prefix{unsafeParseAddress("10.240.0.0"), 9},
			b:       Prefix{unsafeParseAddress("10.200.0.0"), 16},
			matches: true,
			exact:   false,
			common:  9,
			child:   1,
		},
		{
			desc:    "disjoint",
			a:       Prefix{Address{0}, 1},
			b:       Prefix{unsafeParseAddress("128.0.0.0"), 1},
			matches: false,
			common:  0,
			child:   1,
		},
		{
			desc:    "disjoint longer",
			a:       Prefix{unsafeParseAddress("0.0.0.0"), 17},
			b:       Prefix{unsafeParseAddress("0.0.128.0"), 17},
			matches: false,
			common:  16,
			child:   1,
		},
		{
			desc:    "disjoint longer partial",
			a:       Prefix{unsafeParseAddress("0.0.0.0"), 17},
			b:       Prefix{unsafeParseAddress("0.1.0.0"), 17},
			matches: false,
			common:  15,
			child:   1,
		},
		{
			desc:    "disjoint backwards",
			a:       Prefix{unsafeParseAddress("128.0.0.0"), 1},
			b:       Prefix{Address{0}, 1},
			matches: false,
			common:  0,
			child:   0,
		},
		{
			desc:    "disjoint backwards longer",
			a:       Prefix{unsafeParseAddress("0.0.128.0"), 19},
			b:       Prefix{unsafeParseAddress("0.0.0.0"), 19},
			matches: false,
			common:  16,
			child:   0,
		},
		{
			desc:    "disjoint backwards longer partial",
			a:       Prefix{unsafeParseAddress("0.1.0.0"), 19},
			b:       Prefix{unsafeParseAddress("0.0.0.0"), 19},
			matches: false,
			common:  15,
			child:   0,
		},
		{
			desc:    "disjoint with common",
			a:       Prefix{unsafeParseAddress("10.0.0.0"), 16},
			b:       Prefix{unsafeParseAddress("10.10.0.0"), 16},
			matches: false,
			common:  12,
			child:   1,
		},
		{
			desc:    "disjoint with more disjoint bytes",
			a:       Prefix{unsafeParseAddress("0.255.255.0"), 24},
			b:       Prefix{unsafeParseAddress("128.0.0.0"), 24},
			matches: false,
			common:  0,
			child:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			matches, exact, common, child := contains(tt.a, tt.b)
			assert.Equal(t, tt.matches, matches)
			assert.Equal(t, tt.exact, exact)
			assert.Equal(t, tt.common, common)
			assert.Equal(t, tt.child, child)

			// Opportunistically test the compare function
			t.Run("compare forward", func(t *testing.T) {
				_, reversed, _, _ := compare(tt.a, tt.b)
				assert.False(t, reversed)
			})
			t.Run("compare reversed", func(t *testing.T) {
				_, reversed, _, _ := compare(tt.b, tt.a)
				assert.Equal(t, tt.a.length != tt.b.length, reversed)
			})
		})
	}
}

func TestBitsToBytes(t *testing.T) {
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

func TestDeleteFromNilTree(t *testing.T) {
	var trie *trieNode

	key := Prefix{}
	trie, err := trie.Delete(key)
	assert.Nil(t, trie)
	assert.NotNil(t, err)
}

func TestDeleteSimple(t *testing.T) {
	var trie *trieNode

	key := Prefix{
		unsafeParseAddress("172.16.200.0"),
		24,
	}
	trie, err := trie.Insert(key, nil)
	trie, err = trie.Delete(key)
	assert.Nil(t, err)
	assert.Nil(t, trie)
}

func TestDeleteLeftChild(t *testing.T) {
	var trie *trieNode

	key := Prefix{
		unsafeParseAddress("172.16.200.0"),
		24,
	}
	trie, err := trie.Insert(key, nil)
	childKey := Prefix{
		unsafeParseAddress("172.16.200.0"),
		25,
	}
	trie, err = trie.Insert(childKey, nil)
	trie, err = trie.Delete(key)
	assert.Nil(t, err)
	assert.NotNil(t, trie)

	assert.Nil(t, trie.Match(key))
	assert.NotNil(t, trie.Match(childKey))
}

func TestDeleteRightChild(t *testing.T) {
	var trie *trieNode

	key := Prefix{
		unsafeParseAddress("172.16.200.0"),
		24,
	}
	trie, err := trie.Insert(key, nil)
	childKey := Prefix{
		unsafeParseAddress("172.16.200.128"),
		25,
	}
	trie, err = trie.Insert(childKey, nil)
	trie, err = trie.Delete(key)
	assert.Nil(t, err)
	assert.NotNil(t, trie)

	assert.Nil(t, trie.Match(key))
	assert.NotNil(t, trie.Match(childKey))
}

func TestDeleteBothChildren(t *testing.T) {
	var trie *trieNode

	key := Prefix{
		unsafeParseAddress("172.16.200.0"),
		24,
	}
	trie, err := trie.Insert(key, nil)
	leftChild := Prefix{
		unsafeParseAddress("172.16.200.0"),
		25,
	}
	trie, err = trie.Insert(leftChild, nil)
	rightChild := Prefix{
		unsafeParseAddress("172.16.200.128"),
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

func TestDeleteRecursiveNil(t *testing.T) {
	var trie *trieNode

	key := Prefix{
		unsafeParseAddress("172.16.200.0"),
		24,
	}
	trie, err := trie.Insert(key, nil)
	childKey := Prefix{
		unsafeParseAddress("172.16.200.0"),
		25,
	}
	trie, err = trie.Delete(childKey)
	assert.NotNil(t, err)
	assert.NotNil(t, trie)

	assert.NotNil(t, trie.Match(key))
	match := trie.Match(childKey)
	assert.NotEqual(t, childKey, match.Prefix)
}

func TestDeleteRecursiveLeftChild(t *testing.T) {
	// NOTE: There's no specific test for other child combinations because I
	// didn't feel it added much. It uses already well-tested code paths.
	var trie *trieNode

	key := Prefix{
		unsafeParseAddress("172.16.200.0"),
		24,
	}
	trie, err := trie.Insert(key, nil)
	childKey := Prefix{
		unsafeParseAddress("172.16.200.0"),
		25,
	}
	trie, err = trie.Insert(childKey, nil)
	trie, err = trie.Delete(childKey)
	assert.Nil(t, err)
	assert.NotNil(t, trie)

	assert.NotNil(t, trie.Match(key))
	match := trie.Match(childKey)
	assert.NotEqual(t, childKey, match.Prefix)
}

func TestDeleteRecursiveLeftChild32Promote(t *testing.T) {
	// NOTE: There's no specific test for other child combinations because I
	// didn't feel it added much. It uses already well-tested code paths.
	var trie *trieNode

	key := Prefix{
		unsafeParseAddress("172.16.200.128"),
		25,
	}
	trie, err := trie.Insert(key, nil)
	childKey := Prefix{
		unsafeParseAddress("172.16.200.0"),
		25,
	}
	trie, err = trie.Insert(childKey, nil)
	trie, err = trie.Delete(childKey)
	assert.Nil(t, err)
	assert.NotNil(t, trie)

	assert.NotNil(t, trie.Match(key))
	match := trie.Match(childKey)
	assert.Nil(t, match)
	assert.Equal(t, 1, trie.height())
	assert.Equal(t, int64(1), trie.NumNodes())
}

func TestDeleteKeyTooBroad(t *testing.T) {
	var trie *trieNode

	key := Prefix{
		unsafeParseAddress("172.16.200.0"),
		25,
	}
	trie, err := trie.Insert(key, nil)

	broadKey := Prefix{
		unsafeParseAddress("172.16.200.0"),
		24,
	}
	trie, err = trie.Delete(broadKey)
	assert.NotNil(t, err)
	assert.NotNil(t, trie)

	assert.NotNil(t, trie.Match(key))
	assert.Nil(t, trie.Match(broadKey))
}

func TestDeleteKeyDisjoint(t *testing.T) {
	var trie *trieNode

	key := Prefix{
		unsafeParseAddress("172.16.200.0"),
		25,
	}
	trie, err := trie.Insert(key, nil)

	disjointKey := Prefix{
		unsafeParseAddress("172.16.200.128"),
		25,
	}
	trie, err = trie.Delete(disjointKey)
	assert.NotNil(t, err)
	assert.NotNil(t, trie)

	assert.NotNil(t, trie.Match(key))
	assert.Nil(t, trie.Match(disjointKey))
}

func TestSuccessivelyBetter(t *testing.T) {
	var trie *trieNode

	keys := []Prefix{
		Prefix{unsafeParseAddress("10.224.24.0"), 0},
		Prefix{unsafeParseAddress("10.224.24.0"), 1},
		Prefix{unsafeParseAddress("10.224.24.0"), 8},
		Prefix{unsafeParseAddress("10.224.24.0"), 12},
		Prefix{unsafeParseAddress("10.224.24.0"), 16},
		Prefix{unsafeParseAddress("10.224.24.0"), 18},
		Prefix{unsafeParseAddress("10.224.24.0"), 20},
		Prefix{unsafeParseAddress("10.224.24.0"), 21},
		Prefix{unsafeParseAddress("10.224.24.0"), 22},
		Prefix{unsafeParseAddress("10.224.24.0"), 24},
		Prefix{unsafeParseAddress("10.224.24.0"), 27},
		Prefix{unsafeParseAddress("10.224.24.0"), 30},
		Prefix{unsafeParseAddress("10.224.24.0"), 32},
	}

	// Add successively more specific keys to the trie and assert that exact
	// matches are returned when appropriate and non-exact, but longest matches
	// are returned for the rest.
	for index, key := range keys {
		var err error
		trie, err = trie.Insert(key, nil)
		assert.Nil(t, err)
		assert.Equal(t, int64(index+1), trie.NumNodes())
		assert.True(t, trie.isValid())
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
		assert.Equal(t, int64(len(keys)-index-1), trie.NumNodes())
		assert.True(t, trie.isValid())
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

func TestIterate(t *testing.T) {
	keys := []Prefix{
		Prefix{unsafeParseAddress("172.21.0.0"), 20},
		Prefix{unsafeParseAddress("192.68.27.0"), 25},
		Prefix{unsafeParseAddress("192.168.26.128"), 25},
		Prefix{unsafeParseAddress("10.224.24.0"), 32},
		Prefix{unsafeParseAddress("192.68.24.0"), 24},
		Prefix{unsafeParseAddress("172.16.0.0"), 12},
		Prefix{unsafeParseAddress("192.68.26.0"), 24},
		Prefix{unsafeParseAddress("10.224.24.0"), 30},
		Prefix{unsafeParseAddress("192.168.24.0"), 24},
		Prefix{unsafeParseAddress("192.168.25.0"), 24},
		Prefix{unsafeParseAddress("192.168.26.0"), 25},
		Prefix{unsafeParseAddress("192.68.25.0"), 24},
		Prefix{unsafeParseAddress("192.168.27.0"), 24},
		Prefix{unsafeParseAddress("172.20.128.0"), 19},
		Prefix{unsafeParseAddress("192.68.27.128"), 25},
	}

	golden := []Prefix{
		Prefix{unsafeParseAddress("10.224.24.0"), 30},
		Prefix{unsafeParseAddress("10.224.24.0"), 32},
		Prefix{unsafeParseAddress("172.16.0.0"), 12},
		Prefix{unsafeParseAddress("172.20.128.0"), 19},
		Prefix{unsafeParseAddress("172.21.0.0"), 20},
		Prefix{unsafeParseAddress("192.68.24.0"), 24},
		Prefix{unsafeParseAddress("192.68.25.0"), 24},
		Prefix{unsafeParseAddress("192.68.26.0"), 24},
		Prefix{unsafeParseAddress("192.68.27.0"), 25},
		Prefix{unsafeParseAddress("192.68.27.128"), 25},
		Prefix{unsafeParseAddress("192.168.24.0"), 24},
		Prefix{unsafeParseAddress("192.168.25.0"), 24},
		Prefix{unsafeParseAddress("192.168.26.0"), 25},
		Prefix{unsafeParseAddress("192.168.26.128"), 25},
		Prefix{unsafeParseAddress("192.168.27.0"), 24},
	}

	var trie *trieNode
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

func printTrie(trie *trieNode) {
	var recurse func(trie *trieNode, level int)

	recurse = func(trie *trieNode, level int) {
		if trie == nil {
			return
		}
		for i := 0; i < level; i++ {
			fmt.Printf(" ")
		}
		fmt.Printf("%+v, %v, %d\n", trie, trie.isActive, trie.size)
		recurse(trie.children[0], level+1)
		recurse(trie.children[1], level+1)
	}

	recurse(trie, 0)
}

func TestAggregate(t *testing.T) {
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
				pair32{key: Prefix{unsafeParseAddress("10.224.24.2"), 31}},
				pair32{key: Prefix{unsafeParseAddress("10.224.24.1"), 32}},
				pair32{key: Prefix{unsafeParseAddress("10.224.24.0"), 32}},
			},
			golden: []pair32{
				pair32{key: Prefix{unsafeParseAddress("10.224.24.0"), 30}},
			},
		},
		{
			desc: "same as iterate",
			pairs: []pair32{
				pair32{key: Prefix{unsafeParseAddress("172.21.0.0"), 20}},
				pair32{key: Prefix{unsafeParseAddress("192.68.27.0"), 25}},
				pair32{key: Prefix{unsafeParseAddress("192.168.26.128"), 25}},
				pair32{key: Prefix{unsafeParseAddress("10.224.24.0"), 32}},
				pair32{key: Prefix{unsafeParseAddress("192.68.24.0"), 24}},
				pair32{key: Prefix{unsafeParseAddress("172.16.0.0"), 12}},
				pair32{key: Prefix{unsafeParseAddress("192.68.26.0"), 24}},
				pair32{key: Prefix{unsafeParseAddress("10.224.24.0"), 30}},
				pair32{key: Prefix{unsafeParseAddress("192.168.24.0"), 24}},
				pair32{key: Prefix{unsafeParseAddress("192.168.25.0"), 24}},
				pair32{key: Prefix{unsafeParseAddress("192.168.26.0"), 25}},
				pair32{key: Prefix{unsafeParseAddress("192.68.25.0"), 24}},
				pair32{key: Prefix{unsafeParseAddress("192.168.27.0"), 24}},
				pair32{key: Prefix{unsafeParseAddress("172.20.128.0"), 19}},
				pair32{key: Prefix{unsafeParseAddress("192.68.27.128"), 25}},
			},
			golden: []pair32{
				pair32{key: Prefix{unsafeParseAddress("10.224.24.0"), 30}},
				pair32{key: Prefix{unsafeParseAddress("172.16.0.0"), 12}},
				pair32{key: Prefix{unsafeParseAddress("192.68.24.0"), 22}},
				pair32{key: Prefix{unsafeParseAddress("192.168.24.0"), 22}},
			},
		},
		{
			desc: "mixed umbrellas",
			pairs: []pair32{
				pair32{key: Prefix{unsafeParseAddress("10.224.24.0"), 30}, data: true},
				pair32{key: Prefix{unsafeParseAddress("10.224.24.0"), 31}, data: false},
				pair32{key: Prefix{unsafeParseAddress("10.224.24.1"), 32}, data: true},
				pair32{key: Prefix{unsafeParseAddress("10.224.24.0"), 32}, data: false},
			},
			golden: []pair32{
				pair32{key: Prefix{unsafeParseAddress("10.224.24.0"), 30}, data: true},
				pair32{key: Prefix{unsafeParseAddress("10.224.24.0"), 31}, data: false},
				pair32{key: Prefix{unsafeParseAddress("10.224.24.1"), 32}, data: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			var trie *trieNode
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
func TestAggregateEqualComparable(t *testing.T) {
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
				pair32{key: Prefix{unsafeParseAddress("10.224.24.0"), 30}, data: NextHop1},
				pair32{key: Prefix{unsafeParseAddress("10.224.24.0"), 31}, data: NextHop2},
				pair32{key: Prefix{unsafeParseAddress("10.224.24.1"), 32}, data: NextHop1},
				pair32{key: Prefix{unsafeParseAddress("10.224.24.0"), 32}, data: NextHop2},
			},
			golden: []pair32{
				pair32{key: Prefix{unsafeParseAddress("10.224.24.0"), 30}, data: NextHop1},
				pair32{key: Prefix{unsafeParseAddress("10.224.24.0"), 31}, data: NextHop2},
				pair32{key: Prefix{unsafeParseAddress("10.224.24.1"), 32}, data: NextHop1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			var trie *trieNode
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

// Like the TestAggregate above but using a type that is comparable through the
// EqualComparable interface.
func TestTrieNodeEqual(t *testing.T) {
	node := &trieNode{}
	tests := []struct {
		desc  string
		a, b  *trieNode
		equal bool
	}{
		{
			desc:  "nil",
			equal: true,
		},
		{
			desc:  "one nil",
			a:     &trieNode{},
			equal: false,
		},
		{
			desc:  "two simple ones",
			a:     &trieNode{},
			b:     &trieNode{},
			equal: true,
		},
		{
			desc:  "one active",
			a:     &trieNode{},
			b:     &trieNode{isActive: true},
			equal: false,
		},
		{
			desc:  "two active",
			a:     &trieNode{isActive: true},
			b:     &trieNode{isActive: true},
			equal: true,
		},
		{
			desc:  "different data",
			a:     &trieNode{isActive: true, Data: true},
			b:     &trieNode{isActive: true, Data: false},
			equal: false,
		},
		{
			desc:  "same node",
			a:     node,
			b:     node,
			equal: true,
		},
		{
			desc:  "inactive different prefixes",
			a:     &trieNode{isActive: false, Prefix: Prefix{Address{0}, 32}},
			b:     &trieNode{isActive: false, Prefix: Prefix{Address{1}, 32}},
			equal: false,
		},
		{
			desc:  "inactive different prefix lengths",
			a:     &trieNode{isActive: false, Prefix: Prefix{length: 32}},
			b:     &trieNode{isActive: false, Prefix: Prefix{length: 31}},
			equal: false,
		},
		{
			desc:  "child 1 different",
			a:     &trieNode{},
			b:     &trieNode{children: [2]*trieNode{&trieNode{}, nil}},
			equal: false,
		},
		{
			desc:  "child 2 different",
			a:     &trieNode{},
			b:     &trieNode{children: [2]*trieNode{nil, &trieNode{}}},
			equal: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			assert.Equal(t, tt.equal, tt.a.Equal(tt.b))
			assert.Equal(t, tt.equal, tt.b.Equal(tt.a))
		})
	}
}

func TestFlatten(t *testing.T) {
	t.Run("active node needs no children", func(t *testing.T) {
		prefix := unsafeParsePrefix("1.2.3.0/24")
		n := trieNode{
			Prefix:   prefix,
			isActive: true,
			children: [2]*trieNode{
				&trieNode{},
				&trieNode{},
			},
		}
		n.flatten()
		assert.Equal(t, prefix, n.Prefix)
		assert.Equal(t, uint32(1), n.size)
		assert.Equal(t, uint16(1), n.h)
		assert.True(t, n.isActive)
		assert.Nil(t, n.children[0])
		assert.Nil(t, n.children[1])
	})
	t.Run("children of unequal size", func(t *testing.T) {
		prefix := unsafeParsePrefix("1.2.3.0/24")
		left := unsafeParsePrefix("1.2.3.0/26")
		right := unsafeParsePrefix("1.2.3.128/25")
		n := trieNode{
			Prefix: prefix,
			children: [2]*trieNode{
				&trieNode{Prefix: left},
				&trieNode{Prefix: right},
			},
		}
		n.flatten()
		assert.Equal(t, prefix, n.Prefix)
		assert.False(t, n.isActive)
		assert.Equal(t, left, n.children[0].Prefix)
		assert.Equal(t, right, n.children[1].Prefix)
	})
	t.Run("children smaller than half", func(t *testing.T) {
		prefix := unsafeParsePrefix("1.2.3.0/24")
		left := unsafeParsePrefix("1.2.3.0/26")
		right := unsafeParsePrefix("1.2.3.128/26")
		n := trieNode{
			Prefix: prefix,
			children: [2]*trieNode{
				&trieNode{Prefix: left},
				&trieNode{Prefix: right},
			},
		}
		n.flatten()
		assert.Equal(t, prefix, n.Prefix)
		assert.False(t, n.isActive)
		assert.Equal(t, left, n.children[0].Prefix)
		assert.Equal(t, right, n.children[1].Prefix)
	})

	t.Run("children not both active (left)", func(t *testing.T) {
		prefix := unsafeParsePrefix("1.2.3.0/24")
		left := unsafeParsePrefix("1.2.3.0/25")
		right := unsafeParsePrefix("1.2.3.128/25")
		n := trieNode{
			Prefix: prefix,
			children: [2]*trieNode{
				&trieNode{
					Prefix:   left,
					isActive: true,
				},
				&trieNode{
					Prefix: right,
				},
			},
		}
		n.flatten()
		assert.Equal(t, prefix, n.Prefix)
		assert.False(t, n.isActive)
		assert.Equal(t, left, n.children[0].Prefix)
		assert.Equal(t, right, n.children[1].Prefix)
	})

	t.Run("children not both active (right)", func(t *testing.T) {
		prefix := unsafeParsePrefix("1.2.3.0/24")
		left := unsafeParsePrefix("1.2.3.0/25")
		right := unsafeParsePrefix("1.2.3.128/25")
		n := trieNode{
			Prefix: prefix,
			children: [2]*trieNode{
				&trieNode{
					Prefix: left,
				},
				&trieNode{
					Prefix:   right,
					isActive: true,
				},
			},
		}
		n.flatten()
		assert.Equal(t, prefix, n.Prefix)
		assert.False(t, n.isActive)
		assert.Equal(t, left, n.children[0].Prefix)
		assert.Equal(t, right, n.children[1].Prefix)
	})
	t.Run("children both active", func(t *testing.T) {
		prefix := unsafeParsePrefix("1.2.3.0/24")
		left := unsafeParsePrefix("1.2.3.0/25")
		right := unsafeParsePrefix("1.2.3.128/25")
		n := trieNode{
			Prefix: prefix,
			children: [2]*trieNode{
				&trieNode{
					Prefix:   left,
					isActive: true,
				},
				&trieNode{
					Prefix:   right,
					isActive: true,
				},
			},
		}
		n.flatten()
		assert.Equal(t, prefix, n.Prefix)
		assert.True(t, n.isActive)
		assert.Equal(t, uint32(1), n.size)
		assert.Equal(t, uint16(1), n.h)
		assert.Nil(t, n.children[0])
		assert.Nil(t, n.children[1])
	})
}
