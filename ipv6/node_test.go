package ipv6

import (
	"fmt"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

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
			var mismatch uint128
			mismatch = uint128{0xff00000000000000, 0}.and(uint128{0x8000000000000000, 0}.rightShift(int(tt.nodeLength - 1)).complement())

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
