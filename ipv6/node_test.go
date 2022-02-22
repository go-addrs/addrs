package ipv6

import (
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
