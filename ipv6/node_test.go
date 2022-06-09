package ipv6

import (
	"fmt"
	"reflect"
	"testing"
	"unsafe"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	assert.NotNil(t, key) // staticcheck has an issue without this
	keySize := int(unsafe.Sizeof(key))
	keyAlign := int(unsafe.Alignof(key))

	node := trieNode{}
	assert.NotNil(t, node) // staticcheck has an issue without this
	nodeSize := int(unsafe.Sizeof(node))
	nodeAlign := int(unsafe.Alignof(node))

	// Why would this ever be more than 8?
	assert.LessOrEqual(t, keyAlign, 8)
	assert.LessOrEqual(t, nodeAlign, 8)

	assert.Equal(t,
		intMin(
			5*keyAlign,
			24,
		),
		keySize,
	)
	assert.Equal(t,
		intMin(
			64,
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

	trie = trie.InsertOrUpdate(key, true, ieq)
	assert.True(t, trie.isValid())
	assert.True(t, trie.Match(key).Data.(bool))

	trie = trie.InsertOrUpdate(key, false, ieq)
	assert.True(t, trie.isValid())
	assert.False(t, trie.Match(key).Data.(bool))
}

func TestInsertOrUpdateNewKey(t *testing.T) {
	var trie *trieNode

	key := Prefix{}

	trie = trie.InsertOrUpdate(key, true, ieq)
	assert.True(t, trie.isValid())
	assert.True(t, trie.Match(key).Data.(bool))

	newKey := Prefix{Address{uint128{0, 0}}, 1}
	trie = trie.InsertOrUpdate(newKey, false, ieq)
	assert.True(t, trie.isValid())
	assert.True(t, trie.Match(key).Data.(bool))
	assert.False(t, trie.Match(newKey).Data.(bool))
}

func TestInsertOrUpdateNarrowerKey(t *testing.T) {
	var trie *trieNode

	key := Prefix{Address{uint128{0, 0}}, 1}

	trie = trie.InsertOrUpdate(key, true, ieq)
	assert.True(t, trie.isValid())
	assert.True(t, trie.Match(key).Data.(bool))

	newKey := Prefix{}
	trie = trie.InsertOrUpdate(newKey, false, ieq)
	assert.True(t, trie.isValid())
	assert.True(t, trie.Match(key).Data.(bool))
	assert.False(t, trie.Match(newKey).Data.(bool))
}

func TestInsertOrUpdateDisjointKeys(t *testing.T) {
	var trie *trieNode

	key := Prefix{Address{uint128{0, 0}}, 1}

	trie = trie.InsertOrUpdate(key, true, ieq)
	assert.True(t, trie.isValid())
	assert.True(t, trie.Match(key).Data.(bool))

	newKey := Prefix{Address{uint128{0x8000000000000000, 0}}, 1}
	trie = trie.InsertOrUpdate(newKey, false, ieq)
	assert.True(t, trie.isValid())
	assert.True(t, trie.Match(key).Data.(bool))
	assert.False(t, trie.Match(newKey).Data.(bool))
}

func TestInsertOrUpdateInactive(t *testing.T) {
	var trie *trieNode

	key := Prefix{Address{uint128{0, 0}}, 1}

	trie = trie.InsertOrUpdate(key, true, ieq)
	assert.True(t, trie.isValid())
	assert.True(t, trie.Match(key).Data.(bool))

	newKey := Prefix{Address{uint128{0x8000000000000000, 0}}, 1}
	trie = trie.InsertOrUpdate(newKey, false, ieq)
	assert.True(t, trie.isValid())
	assert.True(t, trie.Match(key).Data.(bool))
	assert.False(t, trie.Match(newKey).Data.(bool))

	inactiveKey := Prefix{}
	trie = trie.InsertOrUpdate(inactiveKey, "value", ieq)
	assert.True(t, trie.isValid())
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

	trie, err = trie.Update(key, false, ieq)
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

	newKey := Prefix{Address{uint128{0, 0}}, 1}
	trie, err = trie.Update(newKey, false, ieq)
	assert.True(t, trie.isValid())
	assert.NotNil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))
	assert.True(t, trie.Match(newKey).Data.(bool))
}

func TestUpdateNarrowerKey(t *testing.T) {
	var trie *trieNode

	key := Prefix{Address{uint128{0, 0}}, 1}

	trie, err := trie.Insert(key, true)
	assert.True(t, trie.isValid())
	assert.Nil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))

	newKey := Prefix{}
	trie, err = trie.Update(newKey, false, ieq)
	assert.True(t, trie.isValid())
	assert.NotNil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))
	assert.Nil(t, trie.Match(newKey))
}

func TestUpdateDisjointKeys(t *testing.T) {
	var trie *trieNode

	key := Prefix{Address{uint128{0, 0}}, 1}

	trie, err := trie.Insert(key, true)
	assert.True(t, trie.isValid())
	assert.Nil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))

	newKey := Prefix{Address{uint128{0x8000000000000000, 0}}, 1}
	trie, err = trie.Update(newKey, false, ieq)
	assert.True(t, trie.isValid())
	assert.NotNil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))
	assert.Nil(t, trie.Match(newKey))
}

func TestUpdateInactive(t *testing.T) {
	var trie *trieNode

	key := Prefix{Address{uint128{0, 0}}, 1}

	trie, err := trie.Insert(key, true)
	assert.True(t, trie.isValid())
	assert.Nil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))

	newKey := Prefix{Address{uint128{0x8000000000000000, 0}}, 1}
	trie, err = trie.Insert(newKey, false)
	assert.True(t, trie.isValid())
	assert.Nil(t, err)
	assert.True(t, trie.Match(key).Data.(bool))
	assert.False(t, trie.Match(newKey).Data.(bool))

	inactiveKey := Prefix{}
	trie, err = trie.Update(inactiveKey, "value", ieq)
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
		Address{uint128{0, 0}},
		0,
	}, nil)
	assert.Nil(t, err)
	assert.True(t, trie.active())
	assert.True(t, trie.isValid())
	assert.Equal(t, 1, trie.height())

	assert.Equal(t, trie, trie.Match(Prefix{
		_a("2001::"),
		0,
	}))
}

func TestGetOrInsertTrivial(t *testing.T) {
	var trie *trieNode
	assert.Equal(t, int64(0), trie.NumNodes())
	assert.True(t, trie.isValid())

	key := Prefix{Address{uint128{0, 0}}, 0}

	trie, node := trie.GetOrInsert(key, true)
	assert.True(t, trie.isValid())
	assert.Equal(t, trie, node)
	assert.True(t, node.Data.(bool))
}

func TestGetOrInsertExists(t *testing.T) {
	var trie *trieNode

	key := Prefix{Address{uint128{0, 0}}, 0}

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

	existingKey := Prefix{_a("2001:224::"), 32}
	trie, err := trie.Insert(existingKey, true)
	assert.Nil(t, err)
	assert.True(t, trie.isValid())

	broaderKey := Prefix{_a("2001::"), 16}
	trie, node := trie.GetOrInsert(broaderKey, false)

	assert.True(t, trie.isValid())
	assert.Equal(t, trie, node)
	assert.False(t, node.Data.(bool))

	assert.True(t, trie.Match(existingKey).Data.(bool))
	assert.False(t, trie.Match(broaderKey).Data.(bool))
}

func TestGetOrInsertNarrower(t *testing.T) {
	var trie *trieNode

	existingKey := Prefix{_a("2001:224::"), 32}
	trie, err := trie.Insert(existingKey, true)
	assert.Nil(t, err)
	assert.True(t, trie.isValid())

	narrowerKey := Prefix{_a("2001:224:24::"), 96}
	trie, node := trie.GetOrInsert(narrowerKey, false)

	assert.True(t, trie.isValid())
	assert.NotEqual(t, trie, node)
	assert.False(t, node.Data.(bool))

	assert.True(t, trie.Match(existingKey).Data.(bool))
	assert.False(t, trie.Match(narrowerKey).Data.(bool))
}

func TestGetOrInsertDisjoint(t *testing.T) {
	var trie *trieNode

	existingKey := Prefix{_a("2001:224::"), 32}
	trie, err := trie.Insert(existingKey, true)
	assert.Nil(t, err)
	assert.True(t, trie.isValid())

	disjointKey := Prefix{_a("2001:225::"), 32}
	trie, node := trie.GetOrInsert(disjointKey, false)

	assert.True(t, trie.isValid())
	assert.False(t, node.Data.(bool))

	assert.True(t, trie.Match(existingKey).Data.(bool))
	assert.False(t, trie.Match(disjointKey).Data.(bool))
}

func TestGetOrInsertInActive(t *testing.T) {
	var trie *trieNode

	trie, _ = trie.Insert(Prefix{_a("2001:224::"), 32}, true)
	trie, _ = trie.Insert(Prefix{_a("2001:225::"), 32}, true)
	assert.True(t, trie.isValid())

	trie, node := trie.GetOrInsert(Prefix{_a("2001:224::"), 31}, false)
	assert.True(t, trie.isValid())
	assert.Equal(t, trie, node)
	assert.False(t, node.Data.(bool))
}

func TestNoMatchTooBroad(t *testing.T) {
	var trie *trieNode

	trie, err := trie.Insert(Prefix{
		_a("10::"),
		24,
	}, nil)
	assert.Nil(t, err)
	assert.True(t, trie.active())
	assert.True(t, trie.isValid())
	assert.Equal(t, 1, trie.height())

	assert.Nil(t, trie.Match(Prefix{
		_a("10::"),
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
			nodeAddress:   _a("10::"),
			nodeLength:    112,
			searchAddress: _a("10::1:0"),
			searchLength:  128,
		},
		{
			desc:          "full bytes, mismatch in earlier byte",
			nodeAddress:   _a("10::"),
			nodeLength:    112,
			searchAddress: _a("10:1::"),
			searchLength:  128,
		},
		{
			desc:          "full bytes, mismatch in first byte",
			nodeAddress:   _a("10::"),
			nodeLength:    112,
			searchAddress: _a("11::"),
			searchLength:  128,
		},
		{
			desc:          "mismatch in partial byte",
			nodeAddress:   _a("10::"),
			nodeLength:    120,
			searchAddress: _a("10::8000"),
			searchLength:  128,
		},
		{
			desc:          "only one partial byte",
			nodeAddress:   Address{},
			nodeLength:    15,
			searchAddress: _a("2::"),
			searchLength:  16,
		},
		{
			desc:          "only one full byte",
			nodeAddress:   Address{},
			nodeLength:    16,
			searchAddress: _a("10::"),
			searchLength:  32,
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
			nodeAddress: _a("10::"),
			nodeLength:  112,
		},
		{
			desc:        "full bytes, mismatch in earlier byte",
			nodeAddress: _a("10::"),
			nodeLength:  112,
		},
		{
			desc:        "full bytes, mismatch in first byte",
			nodeAddress: _a("10::"),
			nodeLength:  112,
		},
		{
			desc:        "mismatch in partial byte",
			nodeAddress: _a("10::"),
			nodeLength:  120,
		},
		{
			desc:        "only one full byte",
			nodeAddress: Address{},
			nodeLength:  16,
		},
		{
			desc:        "partial byte",
			nodeAddress: Address{uint128{0xfe00000000000000, 0}},
			nodeLength:  15,
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
		nodeAddress uint128
		nodeLength  uint32
	}{
		{
			nodeAddress: uint128{0x8000000000000000, 0},
			nodeLength:  1,
		},
		{
			nodeAddress: uint128{0xc000000000000000, 0},
			nodeLength:  2,
		},
		{
			nodeAddress: uint128{0xe000000000000000, 0},
			nodeLength:  3,
		},
		{
			nodeAddress: uint128{0xf000000000000000, 0},
			nodeLength:  4,
		},
		{
			nodeAddress: uint128{0xf800000000000000, 0},
			nodeLength:  5,
		},
		{
			nodeAddress: uint128{0xfc00000000000000, 0},
			nodeLength:  6,
		},
		{
			nodeAddress: uint128{0xfe00000000000000, 0},
			nodeLength:  7,
		},
		{
			nodeAddress: uint128{0xff00000000000000, 0},
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
				Address{uint128{0xff00000000000000, 0}},
				tt.nodeLength,
			}))

			// byte with 0 in the last bit to match based on nodeLength
			var mismatch uint128 = uint128{0xff00000000000000, 0}.and(uint128{0x8000000000000000, 0}.rightShift(int(tt.nodeLength - 1)).complement())

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
			desc: "32 and 48",
			a:    Prefix{_a("10:200::"), 32},
			b:    Prefix{_a("10:200:20::"), 48},
			c:    Prefix{_a("10:200:20::"), 128},
		},
		{
			desc: "34 and 60",
			a:    Prefix{_a("10:200::"), 34},
			b:    Prefix{Address{uint128{0x0ac800e000000000, 0}}, 60},
			c:    Prefix{Address{uint128{0x0ac800f800000000, 0}}, 127},
		},
		{
			desc: "0 and 32",
			a:    Prefix{Address{uint128{0, 0}}, 0},
			b:    Prefix{_a("10::"), 16},
			c:    Prefix{_a("10:10::"), 32},
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
			a:     Prefix{Address{uint128{0, 0}}, 1},
			b:     Prefix{_a("8000::"), 1},
			super: Prefix{Address{uint128{0, 0}}, 0},
		},
		{
			desc:  "thirty-third bit",
			a:     Prefix{_a("A000:2240::"), 33},
			b:     Prefix{_a("A000:2240:8000::"), 33},
			super: Prefix{_a("A000:2240::"), 32},
		},
		{
			desc:  "partial b bit",
			a:     Prefix{_a("A000:2240::"), 47},
			b:     Prefix{_a("A000:2240:8::"), 47},
			super: Prefix{_a("A000:2240::"), 44},
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
				Prefix{Address{uint128{0, 0}}, 0},
				Prefix{Address{uint128{0xff00000000000000, 0}}, 8},
				Prefix{Address{uint128{0xfe00000000000000, 0}}, 8},
				Prefix{Address{uint128{0xffff000000000000, 0}}, 16},
				Prefix{Address{uint128{0xfffe000000000000, 0}}, 16},
				Prefix{Address{uint128{0xffff000000000000, 0}}, 17},
				Prefix{Address{uint128{0xfffe800000000000, 0}}, 17},
				Prefix{Address{uint128{0xfffe800000000000, 0}}, 18},
				Prefix{Address{uint128{0xffffb00000000000, 0}}, 18},
				Prefix{Address{uint128{0xfffebf0000000000, 0}}, 24},
				Prefix{Address{uint128{0xffffbe0000000000, 0}}, 24},
				Prefix{Address{uint128{0xfffffffffffffebf, 0}}, 64},
				Prefix{Address{uint128{0xffffffffffffffbe, 0}}, 64},
				Prefix{Address{uint128{0xffffffffffffffff, 0xffffffff00000000}}, 96},
				Prefix{Address{uint128{0xffffffffffffffff, 0xfffffffe00000000}}, 96},
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
		common         uint32
		child          int
	}{
		{
			desc:    "trivial",
			a:       Prefix{Address{uint128{}}, 0},
			b:       Prefix{Address{uint128{}}, 0},
			matches: true,
			exact:   true,
			common:  0,
		},
		{
			desc:    "exact",
			a:       Prefix{_a("2001::"), 32},
			b:       Prefix{_a("2001::"), 32},
			matches: true,
			exact:   true,
			common:  32,
		},
		{
			desc:    "exact partial",
			a:       Prefix{_a("0a00:1f00::"), 39},
			b:       Prefix{_a("0a00:1f00:00F0::"), 39},
			matches: true,
			exact:   true,
			common:  39,
		},
		{
			desc:    "empty prefix match",
			a:       Prefix{Address{uint128{}}, 0},
			b:       Prefix{_a("2001:10::"), 32},
			matches: true,
			exact:   false,
			common:  0,
			child:   0,
		},
		{
			desc:    "empty prefix match backwards",
			a:       Prefix{Address{uint128{}}, 0},
			b:       Prefix{_a("F030:10::"), 32},
			matches: true,
			exact:   false,
			common:  0,
			child:   1,
		},
		{
			desc:    "matches",
			a:       Prefix{_a("2001::"), 16},
			b:       Prefix{_a("2001:10::"), 32},
			matches: true,
			exact:   false,
			common:  16,
			child:   0,
		},
		{
			desc:    "matches partial",
			a:       Prefix{_a("2001:2000::"), 17},
			b:       Prefix{_a("2001:2190::"), 32},
			matches: true,
			exact:   false,
			common:  17,
			child:   0,
		},
		{
			desc:    "matches backwards",
			a:       Prefix{_a("A0::"), 16},
			b:       Prefix{_a("A0:c800::"), 32},
			matches: true,
			exact:   false,
			common:  16,
			child:   1,
		},
		{
			desc:    "matches backwards partial",
			a:       Prefix{_a("10:f000::"), 17},
			b:       Prefix{_a("10:c800::"), 32},
			matches: true,
			exact:   false,
			common:  17,
			child:   1,
		},
		{
			desc:    "disjoint",
			a:       Prefix{Address{uint128{}}, 1},
			b:       Prefix{_a("8000::"), 1},
			matches: false,
			common:  0,
			child:   1,
		},
		{
			desc:    "disjoint longer",
			a:       Prefix{_a("::"), 65},
			b:       Prefix{_a("::8000:0:0:0"), 65},
			matches: false,
			common:  64,
			child:   1,
		},
		{
			desc:    "disjoint longer partial",
			a:       Prefix{_a("::"), 65},
			b:       Prefix{_a("0:0:0:1::"), 65},
			matches: false,
			common:  63,
			child:   1,
		},
		{
			desc:    "disjoint backwards",
			a:       Prefix{_a("8000::"), 1},
			b:       Prefix{Address{uint128{}}, 1},
			matches: false,
			common:  0,
			child:   0,
		},
		{
			desc:    "disjoint backwards longer",
			a:       Prefix{_a("::8000:0:0:0"), 71},
			b:       Prefix{_a("::"), 71},
			matches: false,
			common:  64,
			child:   0,
		},
		{
			desc:    "disjoint backwards longer partial",
			a:       Prefix{_a("0:0:0:1::"), 71},
			b:       Prefix{_a("::"), 71},
			matches: false,
			common:  63,
			child:   0,
		},
		{
			desc:    "disjoint with common",
			a:       Prefix{_a("A0::"), 32},
			b:       Prefix{_a("A0:A0::"), 32},
			matches: false,
			common:  24,
			child:   1,
		},
		{
			desc:    "disjoint with more disjoint bytes",
			a:       Prefix{_a("0:0:ffff:ffff:ffff:ffff:0:0"), 112},
			b:       Prefix{_a("8000::"), 112},
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
		_a("1720:16:200::"),
		128,
	}
	trie, err := trie.Insert(key, nil)
	assert.Nil(t, err)
	trie, err = trie.Delete(key)
	assert.Nil(t, err)
	assert.Nil(t, trie)
}

func TestDeleteLeftChild(t *testing.T) {
	var trie *trieNode

	key := Prefix{
		_a("1720:16:200::"),
		48,
	}
	trie, err := trie.Insert(key, nil)
	assert.Nil(t, err)
	childKey := Prefix{
		_a("1720:16:200::"),
		49,
	}
	trie, err = trie.Insert(childKey, nil)
	assert.Nil(t, err)
	trie, err = trie.Delete(key)
	assert.Nil(t, err)
	assert.NotNil(t, trie)

	assert.Nil(t, trie.Match(key))
	assert.NotNil(t, trie.Match(childKey))
}

func TestDeleteRightChild(t *testing.T) {
	var trie *trieNode

	key := Prefix{
		_a("1720:16:200::"),
		48,
	}
	trie, err := trie.Insert(key, nil)
	assert.Nil(t, err)
	childKey := Prefix{
		_a("1720:16:200::8000"),
		49,
	}
	trie, err = trie.Insert(childKey, nil)
	assert.Nil(t, err)
	trie, err = trie.Delete(key)
	assert.Nil(t, err)
	assert.NotNil(t, trie)

	assert.Nil(t, trie.Match(key))
	assert.NotNil(t, trie.Match(childKey))
}

func TestDeleteBothChildren(t *testing.T) {
	var trie *trieNode

	key := Prefix{
		_a("1720:16:200::"),
		48,
	}
	trie, err := trie.Insert(key, nil)
	assert.Nil(t, err)
	leftChild := Prefix{
		_a("1720:16:200::"),
		49,
	}
	trie, err = trie.Insert(leftChild, nil)
	assert.Nil(t, err)
	rightChild := Prefix{
		_a("1720:16:200:8000::"),
		49,
	}
	trie, err = trie.Insert(rightChild, nil)
	assert.Nil(t, err)
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
		_a("1720:16:200::"),
		48,
	}
	trie, err := trie.Insert(key, nil)
	assert.Nil(t, err)
	childKey := Prefix{
		_a("1720:16:200::"),
		49,
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
		_a("1720:16:200::"),
		48,
	}
	trie, err := trie.Insert(key, nil)
	assert.Nil(t, err)
	childKey := Prefix{
		_a("1720:16:200::"),
		49,
	}
	trie, err = trie.Insert(childKey, nil)
	assert.Nil(t, err)
	trie, err = trie.Delete(childKey)
	assert.Nil(t, err)
	assert.NotNil(t, trie)

	assert.NotNil(t, trie.Match(key))
	match := trie.Match(childKey)
	assert.NotEqual(t, childKey, match.Prefix)
}

func TestDeleteRecursiveLeftChild128Promote(t *testing.T) {
	// NOTE: There's no specific test for other child combinations because I
	// didn't feel it added much. It uses already well-tested code paths.
	var trie *trieNode

	key := Prefix{
		_a("1720:16:200:8000::"),
		49,
	}
	trie, err := trie.Insert(key, nil)
	assert.Nil(t, err)
	childKey := Prefix{
		_a("1720:16:200::"),
		49,
	}
	trie, err = trie.Insert(childKey, nil)
	assert.Nil(t, err)
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
		_a("1720:16:200::"),
		49,
	}
	trie, err := trie.Insert(key, nil)
	assert.Nil(t, err)

	broadKey := Prefix{
		_a("1720:16:200::"),
		48,
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
		_a("1720:16:200::"),
		49,
	}
	trie, err := trie.Insert(key, nil)
	assert.Nil(t, err)

	disjointKey := Prefix{
		_a("1720:16:200:8000::"),
		49,
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
		Prefix{_a("2001:224::0d24"), 0},
		Prefix{_a("2001:224::0d24"), 1},
		Prefix{_a("2001:224::0d24"), 8},
		Prefix{_a("2001:224::0d24"), 16},
		Prefix{_a("2001:224::0d24"), 24},
		Prefix{_a("2001:224::0d24"), 32},
		Prefix{_a("2001:224::0d24"), 40},
		Prefix{_a("2001:224::0d24"), 48},
		Prefix{_a("2001:224::0d24"), 56},
		Prefix{_a("2001:224::0d24"), 64},
		Prefix{_a("2001:224::0d24"), 72},
		Prefix{_a("2001:224::0d24"), 80},
		Prefix{_a("2001:224::0d24"), 88},
		Prefix{_a("2001:224::0d24"), 96},
		Prefix{_a("2001:224::0d24"), 104},
		Prefix{_a("2001:224::0d24"), 112},
		Prefix{_a("2001:224::0d24"), 120},
		Prefix{_a("2001:224::0d24"), 128},
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

func TestWalk(t *testing.T) {
	keys := []Prefix{
		Prefix{_a("1272:21::"), 40},
		Prefix{_a("3920:6c8:27::"), 49},
		Prefix{_a("3920:16c8:26::8000"), 49},
		Prefix{_a("2001:2d24:24::0"), 128},
		Prefix{_a("3920:6c8:24::"), 48},
		Prefix{_a("1272:16::"), 18},
		Prefix{_a("3920:6c8:26::"), 48},
		Prefix{_a("2001:2d24:24::0"), 124},
		Prefix{_a("3920:16c8:24::"), 48},
		Prefix{_a("3920:16c8:25::"), 48},
		Prefix{_a("3920:16c8:26::"), 49},
		Prefix{_a("3920:6c8:25::"), 48},
		Prefix{_a("3920:16c8:27::"), 48},
		Prefix{_a("1272:20:8000::"), 39},
		Prefix{_a("3920:6c8:25::8000"), 49},
	}

	golden := []Prefix{
		Prefix{_a("1272:16::"), 18},
		Prefix{_a("1272:20:8000::"), 39},
		Prefix{_a("1272:21::"), 40},
		Prefix{_a("2001:2d24:24::"), 124},
		Prefix{_a("2001:2d24:24::"), 128},
		Prefix{_a("3920:6c8:24::"), 48},
		Prefix{_a("3920:6c8:25::"), 48},
		Prefix{_a("3920:6c8:25::8000"), 49},
		Prefix{_a("3920:6c8:26::"), 48},
		Prefix{_a("3920:6c8:27::"), 49},
		Prefix{_a("3920:16c8:24::"), 48},
		Prefix{_a("3920:16c8:25::"), 48},
		Prefix{_a("3920:16c8:26::8000"), 49},
		Prefix{_a("3920:16c8:27::"), 48},
	}

	var trie *trieNode
	check := func(t *testing.T) {
		result := []Prefix{}
		trie.Walk(func(key Prefix, _ interface{}) bool {
			result = append(result, key)
			return true
		})
		assert.Equal(t, golden, result)

		iterations := 0
		trie.Walk(func(key Prefix, _ interface{}) bool {
			iterations++
			return false
		})
		assert.Equal(t, 1, iterations)

		// Just ensure that iterating with a nil callback doesn't crash
		trie.Walk(nil)
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

type pair128 struct {
	key  Prefix
	data interface{}
}

func PrintTrie(trie *trieNode) {
	if trie == nil {
		fmt.Println("<nil>")
		return
	}
	var recurse func(trie *trieNode, level int)

	recurse = func(trie *trieNode, level int) {
		if trie == nil {
			return
		}
		for i := 0; i < level; i++ {
			fmt.Printf("   ")
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
		pairs  []pair128
		golden []pair128
	}{
		{
			desc:   "nothing",
			pairs:  []pair128{},
			golden: []pair128{},
		},
		{
			desc: "simple aggregation",
			pairs: []pair128{
				pair128{key: Prefix{_a("2001:2d24:24::2"), 127}},
				pair128{key: Prefix{_a("2001:2d24:24::1"), 128}},
				pair128{key: Prefix{_a("2001:2d24:24::0"), 128}},
			},
			golden: []pair128{
				pair128{key: Prefix{_a("2001:2d24:24::0"), 126}},
			},
		},
		{
			desc: "same as iterate",
			pairs: []pair128{
				pair128{key: Prefix{_a("1272:21::"), 40}},
				pair128{key: Prefix{_a("3920:6c8:27::"), 49}},
				pair128{key: Prefix{_a("3920:16c8:26::8000"), 49}},
				pair128{key: Prefix{_a("2001:2d24:24::0"), 128}},
				pair128{key: Prefix{_a("3920:6c8:24::"), 48}},
				pair128{key: Prefix{_a("1272:16::"), 18}},
				pair128{key: Prefix{_a("3920:6c8:26::"), 48}},
				pair128{key: Prefix{_a("2001:2d24:24::0"), 124}},
				pair128{key: Prefix{_a("3920:16c8:24::"), 48}},
				pair128{key: Prefix{_a("3920:16c8:25::"), 48}},
				pair128{key: Prefix{_a("3920:16c8:26::"), 49}},
				pair128{key: Prefix{_a("3920:6c8:25::"), 48}},
				pair128{key: Prefix{_a("3920:16c8:27::"), 48}},
				pair128{key: Prefix{_a("1272:20:8000::"), 39}},
				pair128{key: Prefix{_a("3920:6c8:25::8000"), 49}},
			},
			golden: []pair128{
				pair128{key: Prefix{_a("1272:16::"), 18}},
				pair128{key: Prefix{_a("2001:2d24:24::"), 124}},
				pair128{key: Prefix{_a("3920:6c8:24::"), 47}},
				pair128{key: Prefix{_a("3920:6c8:26::"), 48}},
				pair128{key: Prefix{_a("3920:6c8:27::"), 49}},
				pair128{key: Prefix{_a("3920:16c8:24::"), 47}},
				pair128{key: Prefix{_a("3920:16c8:26::8000"), 49}},
				pair128{key: Prefix{_a("3920:16c8:27::"), 48}},
			},
		},
		{
			desc: "mixed umbrellas",
			pairs: []pair128{
				pair128{key: Prefix{_a("2001:2d24:24::0"), 126}, data: true},
				pair128{key: Prefix{_a("2001:2d24:24::0"), 127}, data: false},
				pair128{key: Prefix{_a("2001:2d24:24::1"), 128}, data: true},
				pair128{key: Prefix{_a("2001:2d24:24::0"), 128}, data: false},
			},
			golden: []pair128{
				pair128{key: Prefix{_a("2001:2d24:24::0"), 126}, data: true},
				pair128{key: Prefix{_a("2001:2d24:24::0"), 127}, data: false},
				pair128{key: Prefix{_a("2001:2d24:24::1"), 128}, data: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			var trie *trieNode
			check := func(t *testing.T) {
				expectedIterations := 0
				result := []pair128{}
				trie.Aggregate(ieq).Walk(
					func(key Prefix, data interface{}) bool {
						result = append(result, pair128{key: key, data: data})
						expectedIterations = 1
						return true
					},
				)
				assert.Equal(t, tt.golden, result)

				iterations := 0
				trie.Aggregate(ieq).Walk(
					func(key Prefix, data interface{}) bool {
						result = append(result, pair128{key: key, data: data})
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

type thing struct {
	// Begin with a type (slice) that is not thing with standard ==
	data []string
}

func (me *thing) IEqual(other interface{}) bool {
	return reflect.DeepEqual(me, other)
}

// Like the TestAggregate above but using a type that is thing through the
// equalComparable interface.
func TestAggregateEqualComparable(t *testing.T) {
	NextHop1 := &thing{data: []string{"2001:2d24:24::1"}}
	NextHop2 := &thing{data: []string{"2001:2d24:24::111"}}
	tests := []struct {
		desc   string
		pairs  []pair128
		golden []pair128
	}{
		{
			desc: "mixed umbrellas",
			pairs: []pair128{
				pair128{key: Prefix{_a("2001:2d24:24::0"), 126}, data: NextHop1},
				pair128{key: Prefix{_a("2001:2d24:24::0"), 127}, data: NextHop2},
				pair128{key: Prefix{_a("2001:2d24:24::1"), 128}, data: NextHop1},
				pair128{key: Prefix{_a("2001:2d24:24::0"), 128}, data: NextHop2},
			},
			golden: []pair128{
				pair128{key: Prefix{_a("2001:2d24:24::0"), 126}, data: NextHop1},
				pair128{key: Prefix{_a("2001:2d24:24::0"), 127}, data: NextHop2},
				pair128{key: Prefix{_a("2001:2d24:24::1"), 128}, data: NextHop1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			var trie *trieNode
			for _, p := range tt.pairs {
				trie, _ = trie.Insert(p.key, p.data)
			}

			result := []pair128{}
			trie.Aggregate(ieq).Walk(
				func(key Prefix, data interface{}) bool {
					result = append(result, pair128{key: key, data: data})
					return true
				},
			)
			assert.Equal(t, tt.golden, result)
		})
	}
}

// Like the TestAggregate above but using a type that is comparable through the
// equalComparable interface.
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
			a:     &trieNode{isActive: false, Prefix: Prefix{Address{uint128{}}, 128}},
			b:     &trieNode{isActive: false, Prefix: Prefix{Address{uint128{0, 1}}, 128}},
			equal: false,
		},
		{
			desc:  "inactive different prefix lengths",
			a:     &trieNode{isActive: false, Prefix: Prefix{length: 128}},
			b:     &trieNode{isActive: false, Prefix: Prefix{length: 127}},
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
			assert.Equal(t, tt.equal, tt.a.Equal(tt.b, ieq))
			assert.Equal(t, tt.equal, tt.b.Equal(tt.a, ieq))
		})
	}
}

func TestFlatten(t *testing.T) {
	t.Run("active node needs no children", func(t *testing.T) {
		prefix := _p("1:2:3::/48")
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
		assert.True(t, n.isActive)
		assert.Nil(t, n.children[0])
		assert.Nil(t, n.children[1])
	})
	t.Run("children of unequal size", func(t *testing.T) {
		prefix := _p("1:2:3::/48")
		left := _p("1:2:3::/52")
		right := _p("1:2:3:8000::/49")
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
		prefix := _p("1:2:3::/48")
		left := _p("1:2:3::/52")
		right := _p("1:2:3:8000::/52")
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
		prefix := _p("1:2:3::/48")
		left := _p("1:2:3::/49")
		right := _p("1:2:3:8000::/49")
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
		prefix := _p("1:2:3::/48")
		left := _p("1:2:3::/49")
		right := _p("1:2:3:8000::/49")
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
		prefix := _p("1:2:3::/48")
		left := _p("1:2:3::/49")
		right := _p("1:2:3:8000::/49")
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
		assert.Nil(t, n.children[0])
		assert.Nil(t, n.children[1])
	})
}

type actionType int

const (
	actionTypeRemove actionType = iota
	actionTypeAdd
	actionTypeChange
	actionTypeSame
)

type diffAction struct {
	t    actionType
	pair pair128
	val  interface{}
}

func TestDiff(t *testing.T) {
	tests := []struct {
		desc        string
		left, right []pair128
		actions     []diffAction
		aggregated  []diffAction
	}{
		{
			desc: "empty",
		}, {
			desc: "right_empty",
			left: []pair128{
				pair128{key: Prefix{_a("2003::1b13:0"), 112}},
			},
			actions: []diffAction{
				diffAction{actionTypeRemove, pair128{key: Prefix{_a("2003::1b13:0"), 112}}, nil},
			},
		}, {
			desc: "right_empty_with_subprefixes",
			left: []pair128{
				pair128{key: Prefix{_a("2003::1b13:0"), 112}},
				pair128{key: Prefix{_a("2003::1b13:8000"), 113}},
			},
			actions: []diffAction{
				diffAction{actionTypeRemove, pair128{key: Prefix{_a("2003::1b13:0"), 112}}, nil},
				diffAction{actionTypeRemove, pair128{key: Prefix{_a("2003::1b13:8000"), 113}}, nil},
			},
			aggregated: []diffAction{
				diffAction{actionTypeRemove, pair128{key: Prefix{_a("2003::1b13:0"), 112}}, nil},
			},
		}, {
			desc: "right_empty_with_subprefix_on_left",
			left: []pair128{
				pair128{key: Prefix{_a("2003::1b13:0"), 112}},
				pair128{key: Prefix{_a("2003::1b13:4000"), 114}},
			},
			actions: []diffAction{
				diffAction{actionTypeRemove, pair128{key: Prefix{_a("2003::1b13:0"), 112}}, nil},
				diffAction{actionTypeRemove, pair128{key: Prefix{_a("2003::1b13:4000"), 114}}, nil},
			},
			aggregated: []diffAction{
				diffAction{actionTypeRemove, pair128{key: Prefix{_a("2003::1b13:0"), 112}}, nil},
			},
		}, {
			desc: "right_empty_with_subprefix_on_right",
			left: []pair128{
				pair128{key: Prefix{_a("2003::1b13:0"), 112}},
				pair128{key: Prefix{_a("2003::1b13:C000"), 114}},
			},
			actions: []diffAction{
				diffAction{actionTypeRemove, pair128{key: Prefix{_a("2003::1b13:0"), 112}}, nil},
				diffAction{actionTypeRemove, pair128{key: Prefix{_a("2003::1b13:C000"), 114}}, nil},
			},
			aggregated: []diffAction{
				diffAction{actionTypeRemove, pair128{key: Prefix{_a("2003::1b13:0"), 112}}, nil},
			},
		}, {
			desc: "no_diff",
			left: []pair128{
				pair128{key: Prefix{_a("2003::1b13:0"), 112}},
			},
			right: []pair128{
				pair128{key: Prefix{_a("2003::1b13:0"), 112}},
			},
			actions: []diffAction{
				diffAction{actionTypeSame, pair128{key: Prefix{_a("2003::1b13:0"), 112}}, nil},
			},
		}, {
			desc: "different_data",
			left: []pair128{
				pair128{key: Prefix{_a("2003::1b13:0"), 112}, data: 2},
			},
			right: []pair128{
				pair128{key: Prefix{_a("2003::1b13:0"), 112}, data: 3},
			},
			actions: []diffAction{
				diffAction{actionTypeChange, pair128{key: Prefix{_a("2003::1b13:0"), 112}, data: 2}, 3},
			},
		}, {
			desc: "right_side_aggregable",
			left: []pair128{
				pair128{key: Prefix{_a("2003::1b13:0"), 112}, data: 2},
			},
			right: []pair128{
				pair128{key: Prefix{_a("2003::1b13:0"), 113}, data: 3},
				pair128{key: Prefix{_a("2003::1b13:8000"), 113}, data: 3},
			},
			actions: []diffAction{
				diffAction{actionTypeRemove, pair128{key: Prefix{_a("2003::1b13:0"), 112}, data: 2}, nil},
				diffAction{actionTypeAdd, pair128{key: Prefix{_a("2003::1b13:0"), 113}, data: 3}, nil},
				diffAction{actionTypeAdd, pair128{key: Prefix{_a("2003::1b13:8000"), 113}, data: 3}, nil},
			},
			aggregated: []diffAction{
				diffAction{actionTypeChange, pair128{key: Prefix{_a("2003::1b13:0"), 112}, data: 2}, 3},
			},
		}, {
			desc: "both_sides_aggregable",
			left: []pair128{
				pair128{key: Prefix{_a("2003::1b13:0"), 113}, data: 2},
				pair128{key: Prefix{_a("2003::1b13:8000"), 113}, data: 2},
			},
			right: []pair128{
				pair128{key: Prefix{_a("2003::1b13:0"), 113}, data: 3},
				pair128{key: Prefix{_a("2003::1b13:8000"), 113}, data: 3},
			},
			actions: []diffAction{
				diffAction{actionTypeChange, pair128{key: Prefix{_a("2003::1b13:0"), 113}, data: 2}, 3},
				diffAction{actionTypeChange, pair128{key: Prefix{_a("2003::1b13:8000"), 113}, data: 2}, 3},
			},
			aggregated: []diffAction{
				diffAction{actionTypeChange, pair128{key: Prefix{_a("2003::1b13:0"), 112}, data: 2}, 3},
			},
		}, {
			desc: "disjoint",
			left: []pair128{
				pair128{key: Prefix{_a("2003::1b13:0"), 113}},
			},
			right: []pair128{
				pair128{key: Prefix{_a("2003::1b13:8000"), 113}},
			},
			actions: []diffAction{
				diffAction{actionTypeRemove, pair128{key: Prefix{_a("2003::1b13:0"), 113}}, nil},
				diffAction{actionTypeAdd, pair128{key: Prefix{_a("2003::1b13:8000"), 113}}, nil},
			},
		}, {
			desc: "contained_right",
			left: []pair128{
				pair128{key: Prefix{_a("2003::1b13:0"), 112}},
			},
			right: []pair128{
				pair128{key: Prefix{_a("2003::1b13:8000"), 113}},
			},
			actions: []diffAction{
				diffAction{actionTypeRemove, pair128{key: Prefix{_a("2003::1b13:0"), 112}}, nil},
				diffAction{actionTypeAdd, pair128{key: Prefix{_a("2003::1b13:8000"), 113}}, nil},
			},
		}, {
			desc: "contained_right_subprefix",
			left: []pair128{
				pair128{key: Prefix{_a("2003::1b13:0"), 113}},
				pair128{key: Prefix{_a("2003::1b13:8000"), 113}},
			},
			right: []pair128{
				pair128{key: Prefix{_a("2003::1b13:8000"), 113}},
			},
			actions: []diffAction{
				diffAction{actionTypeRemove, pair128{key: Prefix{_a("2003::1b13:0"), 113}}, nil},
				diffAction{actionTypeSame, pair128{key: Prefix{_a("2003::1b13:8000"), 113}}, nil},
			},
			aggregated: []diffAction{
				diffAction{actionTypeRemove, pair128{key: Prefix{_a("2003::1b13:0"), 112}}, nil},
				diffAction{actionTypeAdd, pair128{key: Prefix{_a("2003::1b13:8000"), 113}}, nil},
			},
		}, {
			desc: "contained_left",
			left: []pair128{
				pair128{key: Prefix{_a("2003::1b13:0"), 112}},
			},
			right: []pair128{
				pair128{key: Prefix{_a("2003::1b13:0"), 113}},
			},
			actions: []diffAction{
				diffAction{actionTypeRemove, pair128{key: Prefix{_a("2003::1b13:0"), 112}}, nil},
				diffAction{actionTypeAdd, pair128{key: Prefix{_a("2003::1b13:0"), 113}}, nil},
			},
		}, {
			desc: "contained_left_subprefix",
			left: []pair128{
				pair128{key: Prefix{_a("2003::1b13:0"), 113}},
				pair128{key: Prefix{_a("2003::1b13:8000"), 113}},
			},
			right: []pair128{
				pair128{key: Prefix{_a("2003::1b13:0"), 113}},
			},
			actions: []diffAction{
				diffAction{actionTypeSame, pair128{key: Prefix{_a("2003::1b13:0"), 113}}, nil},
				diffAction{actionTypeRemove, pair128{key: Prefix{_a("2003::1b13:8000"), 113}}, nil},
			},
			aggregated: []diffAction{
				diffAction{actionTypeRemove, pair128{key: Prefix{_a("2003::1b13:0"), 112}}, nil},
				diffAction{actionTypeAdd, pair128{key: Prefix{_a("2003::1b13:0"), 113}}, nil},
			},
		}, {
			desc: "aggregated same",
			left: []pair128{
				pair128{key: Prefix{_a("2003::1b13:0"), 113}},
				pair128{key: Prefix{_a("2003::1b13:8000"), 113}},
			},
			right: []pair128{
				pair128{key: Prefix{_a("2003::1b13:0"), 112}},
			},
			actions: []diffAction{
				diffAction{actionTypeAdd, pair128{key: Prefix{_a("2003::1b13:0"), 112}}, nil},
				diffAction{actionTypeRemove, pair128{key: Prefix{_a("2003::1b13:0"), 113}}, nil},
				diffAction{actionTypeRemove, pair128{key: Prefix{_a("2003::1b13:8000"), 113}}, nil},
			},
			aggregated: []diffAction{
				diffAction{actionTypeSame, pair128{key: Prefix{_a("2003::1b13:0"), 112}}, nil},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			left, right := func() (left, right *trieNode) {
				fill := func(pairs []pair128) (trie *trieNode) {
					var err error
					for _, p := range pairs {
						trie, err = trie.Insert(p.key, p.data)
						require.Nil(t, err)
					}
					return
				}
				return fill(tt.left), fill(tt.right)
			}()

			var actions []diffAction
			getHandler := func(ret bool) trieDiffHandler {
				actions = nil
				return trieDiffHandler{
					Removed: func(left *trieNode) bool {
						require.True(t, left.isActive)
						actions = append(actions, diffAction{actionTypeRemove, pair128{key: left.Prefix, data: left.Data}, nil})
						return ret
					},
					Added: func(right *trieNode) bool {
						require.True(t, right.isActive)
						actions = append(actions, diffAction{actionTypeAdd, pair128{key: right.Prefix, data: right.Data}, nil})
						return ret
					},
					Modified: func(left, right *trieNode) bool {
						require.True(t, left.isActive)
						require.True(t, right.isActive)
						actions = append(actions, diffAction{actionTypeChange, pair128{key: left.Prefix, data: left.Data}, right.Data})
						return ret
					},
					Same: func(common *trieNode) bool {
						require.True(t, common.isActive)
						actions = append(actions, diffAction{actionTypeSame, pair128{key: common.Prefix, data: common.Data}, nil})
						return ret
					},
				}
			}

			aggregatedExpected := tt.aggregated
			if aggregatedExpected == nil {
				aggregatedExpected = tt.actions
			}

			t.Run("forward", func(t *testing.T) {
				t.Run("normal", func(t *testing.T) {
					left.Diff(right, getHandler(true), ieq)
					assert.True(t,
						reflect.DeepEqual(tt.actions, actions),
						cmp.Diff(tt.actions, actions, cmp.Exporter(func(reflect.Type) bool { return true })),
					)
					if len(tt.actions) >= 1 {
						// Run the same thing but return false to stop iteration
						t.Run("stop", func(t *testing.T) {
							left.Diff(right, getHandler(false), ieq)
							assert.True(t, reflect.DeepEqual(tt.actions[:1], actions))
						})
					}
				})

				t.Run("aggregated", func(t *testing.T) {
					left.Aggregate(ieq).Diff(right.Aggregate(ieq), getHandler(true), ieq)
					assert.True(t,
						reflect.DeepEqual(aggregatedExpected, actions),
						cmp.Diff(aggregatedExpected, actions, cmp.Exporter(func(reflect.Type) bool { return true })),
					)
				})
			})

			t.Run("backward", func(t *testing.T) {
				t.Run("normal", func(t *testing.T) {
					right.Diff(left, getHandler(true), ieq)

					var expected []diffAction
					for _, action := range tt.actions {
						var t actionType
						pair := action.pair
						val := action.val
						switch action.t {
						case actionTypeRemove:
							t = actionTypeAdd
						case actionTypeAdd:
							t = actionTypeRemove
						default:
							t = action.t
							pair.data, val = val, pair.data
						}
						expected = append(expected, diffAction{t, pair, val})
					}
					assert.True(t,
						reflect.DeepEqual(expected, actions),
						cmp.Diff(expected, actions, cmp.Exporter(func(reflect.Type) bool { return true })),
					)
				})

				t.Run("aggregated", func(t *testing.T) {
					right.Aggregate(ieq).Diff(left.Aggregate(ieq), getHandler(true), ieq)

					var expected []diffAction
					for _, action := range aggregatedExpected {
						var t actionType
						pair := action.pair
						val := action.val
						switch action.t {
						case actionTypeRemove:
							t = actionTypeAdd
						case actionTypeAdd:
							t = actionTypeRemove
						default:
							t = action.t
							pair.data, val = val, pair.data
						}
						expected = append(expected, diffAction{t, pair, val})
					}
					assert.True(t,
						reflect.DeepEqual(expected, actions),
						cmp.Diff(expected, actions, cmp.Exporter(func(reflect.Type) bool { return true })),
					)
				})
			})
		})
	}
}

func TestMap(t *testing.T) {
	tests := []struct {
		desc     string
		original []Prefix
	}{
		{
			desc: "empty",
		}, {
			desc: "single_entry",
			original: []Prefix{
				Prefix{_a("2001:db8:203:113::"), 56},
			},
		}, {
			desc: "bunch of entries",
			original: []Prefix{
				Prefix{_a("2001:db8:203:113::"), 56},
				Prefix{_a("2001:db8:203:113::"), 64},
				Prefix{_a("2001:db8:192:2::"), 59},
				Prefix{_a("2001:db8:198:51::"), 56},
				Prefix{_a("2001:db8:198:51::"), 57},
				Prefix{_a("2001:db8:198:51::"), 58},
				Prefix{_a("2001:db8:198:51::"), 59},
				Prefix{_a("2001:db8:198:51::"), 60},
				Prefix{_a("2001:db8:198:51::"), 61},
				Prefix{_a("2001:db8:198:51::"), 62},
				Prefix{_a("2001:db8:198:51::"), 63},
				Prefix{_a("2001:db8:198:51::"), 64},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			original, expected := func() (left, right *trieNode) {
				fill := func(prefixes []Prefix, value interface{}) (trie *trieNode) {
					var err error
					for _, p := range prefixes {
						trie, err = trie.Insert(p, value)
						require.Nil(t, err)
					}
					return
				}
				return fill(tt.original, false), fill(tt.original, true)
			}()

			result := original.Map(func(Prefix, interface{}) interface{} {
				return true
			}, ieq)
			assert.Equal(t, original.NumNodes(), result.NumNodes())
			expected.Diff(result, trieDiffHandler{
				Removed: func(left *trieNode) bool {
					assert.Fail(t, fmt.Sprintf("found a removed node: %+v, %+v", left.Prefix, left.Data))
					return true
				},
				Added: func(right *trieNode) bool {
					assert.Fail(t, fmt.Sprintf("found an added node: %+v, %+v", right.Prefix, right.Data))
					return true
				},
				Modified: func(left, right *trieNode) bool {
					assert.Fail(t, fmt.Sprintf("found a changed node: %+v: %+v -> %+v", left.Prefix, left.Data, right.Data))
					return true
				},
			}, ieq)
			result = original.Map(func(_ Prefix, value interface{}) interface{} {
				return value
			}, ieq)
			assert.True(t, original == result)
		})
	}
}

func TestNewAggregate(t *testing.T) {
	tests := []struct {
		desc              string
		table, aggregated []pair128
	}{
		{
			desc:       "empty",
			table:      []pair128{},
			aggregated: []pair128{},
		}, {
			desc: "trivial",
			table: []pair128{
				pair128{key: Prefix{_a("2003::1b13:0"), 113}},
			},
			aggregated: []pair128{
				pair128{key: Prefix{_a("2003::1b13:0"), 113}},
			},
		}, {
			desc: "sub_prefix_left",
			table: []pair128{
				pair128{key: Prefix{_a("2003::1b13:0"), 112}},
				pair128{key: Prefix{_a("2003::1b13:0"), 113}},
			},
			aggregated: []pair128{
				pair128{key: Prefix{_a("2003::1b13:0"), 112}},
			},
		}, {
			desc: "different_data",
			table: []pair128{
				pair128{key: Prefix{_a("2003::1b13:0"), 112}},
				pair128{key: Prefix{_a("2003::1b13:0"), 113}, data: true},
			},
			aggregated: []pair128{
				pair128{key: Prefix{_a("2003::1b13:0"), 112}},
				pair128{key: Prefix{_a("2003::1b13:0"), 113}, data: true},
			},
		}, {
			desc: "disjoint",
			table: []pair128{
				pair128{key: Prefix{_a("2003::1b13:0"), 114}},
				pair128{key: Prefix{_a("2003::1b13:8000"), 114}},
			},
			aggregated: []pair128{
				pair128{key: Prefix{_a("2003::1b13:0"), 114}},
				pair128{key: Prefix{_a("2003::1b13:8000"), 114}},
			},
		}, {
			desc: "subprefix_different_data",
			table: []pair128{
				pair128{key: Prefix{_a("2003::1b13:0"), 112}},
				pair128{key: Prefix{_a("2003::1b13:0"), 114}, data: true},
				pair128{key: Prefix{_a("2003::1b13:4000"), 114}},
			},
			aggregated: []pair128{
				pair128{key: Prefix{_a("2003::1b13:0"), 112}},
				pair128{key: Prefix{_a("2003::1b13:0"), 114}, data: true},
			},
		}, {
			desc: "aggregated_children",
			table: []pair128{
				pair128{key: Prefix{_a("2003::1b13:0"), 113}},
				pair128{key: Prefix{_a("2003::1b13:8000"), 113}},
			},
			aggregated: []pair128{
				pair128{key: Prefix{_a("2003::1b13:0"), 112}},
			},
		}, {
			desc: "adjacent_different_data",
			table: []pair128{
				pair128{key: Prefix{_a("2003::1b13:0"), 113}},
				pair128{key: Prefix{_a("2003::1b13:8000"), 113}, data: true},
			},
			aggregated: []pair128{
				pair128{key: Prefix{_a("2003::1b13:0"), 113}},
				pair128{key: Prefix{_a("2003::1b13:8000"), 113}, data: true},
			},
		}, {
			desc: "adjacent_different_lengths",
			table: []pair128{
				pair128{key: Prefix{_a("2003::1b13:0"), 113}},
				pair128{key: Prefix{_a("2003::1b13:8000"), 114}},
			},
			aggregated: []pair128{
				pair128{key: Prefix{_a("2003::1b13:0"), 113}},
				pair128{key: Prefix{_a("2003::1b13:8000"), 114}},
			},
		}, {
			desc: "aggregated_children_have_precedence",
			table: []pair128{
				pair128{key: Prefix{_a("2003::1b13:0"), 112}},
				pair128{key: Prefix{_a("2003::1b13:0"), 113}, data: true},
				pair128{key: Prefix{_a("2003::1b13:8000"), 113}, data: true},
			},
			aggregated: []pair128{
				pair128{key: Prefix{_a("2003::1b13:0"), 112}, data: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			table, aggregated := func() (a, b *trieNode) {
				fill := func(pairs []pair128) (trie *trieNode) {
					var err error
					for _, p := range pairs {
						trie, err = trie.Insert(p.key, p.data)
						require.Nil(t, err)
					}
					return
				}
				return fill(tt.table), fill(tt.aggregated)
			}()

			assert.True(t, table.Aggregate(ieq).Equal(aggregated, ieq))
		})
	}
}
