package ipv4

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func printTrieSet32(trie *trieNodeSet32) {
	printTrie32((*trieNode32)(trie))
}

func TestTrieNodeSet32Halves(t *testing.T) {
	set := trieNodeSet32FromPrefix(unsafeParsePrefix("0.0.0.0/0"))
	a, b := set.halves()
	assert.Equal(t, trieNodeSet32FromPrefix(unsafeParsePrefix("0.0.0.0/1")), a)
	assert.Equal(t, trieNodeSet32FromPrefix(unsafeParsePrefix("128.0.0.0/1")), b)
}

func TestTrieNodeSet32Union(t *testing.T) {
	a := trieNodeSet32FromPrefix(unsafeParsePrefix("10.224.24.0/32"))
	b := trieNodeSet32FromPrefix(unsafeParsePrefix("10.224.24.128/32"))
	tests := []struct {
		description string
		sets        []*trieNodeSet32
		in, out     []Addr
	}{
		{
			description: "two adjacent",
			sets: []*trieNodeSet32{
				trieNodeSet32FromPrefix(unsafeParsePrefix("10.224.24.0/25")),
				trieNodeSet32FromPrefix(unsafeParsePrefix("10.224.24.128/25")),
			},
			in: []Addr{
				unsafeParseAddr("10.224.24.0"),
				unsafeParseAddr("10.224.24.255"),
				unsafeParseAddr("10.224.24.127"),
				unsafeParseAddr("10.224.24.128"),
			},
			out: []Addr{
				unsafeParseAddr("10.224.23.255"),
				unsafeParseAddr("10.224.25.0"),
			},
		},
		{
			description: "nil",
			sets: []*trieNodeSet32{
				nil,
			},
			in: []Addr{},
			out: []Addr{
				unsafeParseAddr("10.224.23.117"),
				unsafeParseAddr("200.193.25.0"),
			},
		},
		{
			description: "not nil then nil",
			sets: []*trieNodeSet32{
				trieNodeSet32FromPrefix(unsafeParsePrefix("10.224.24.0/32")),
				nil,
			},
			in: []Addr{
				unsafeParseAddr("10.224.24.0"),
			},
			out: []Addr{
				unsafeParseAddr("10.224.23.255"),
				unsafeParseAddr("200.193.24.1"),
			},
		},
		{
			description: "same",
			sets: []*trieNodeSet32{
				trieNodeSet32FromPrefix(unsafeParsePrefix("10.224.24.0/32")),
				trieNodeSet32FromPrefix(unsafeParsePrefix("10.224.24.0/32")),
			},
			in: []Addr{
				unsafeParseAddr("10.224.24.0"),
			},
		},
		{
			description: "different then same",
			sets: []*trieNodeSet32{
				trieNodeSet32FromPrefix(unsafeParsePrefix("10.224.29.0/32")),
				trieNodeSet32FromPrefix(unsafeParsePrefix("10.224.24.0/32")),
				trieNodeSet32FromPrefix(unsafeParsePrefix("10.224.24.0/32")),
			},
			in: []Addr{
				unsafeParseAddr("10.224.24.0"),
				unsafeParseAddr("10.224.29.0"),
			},
		},
		{
			description: "duplicates",
			sets: []*trieNodeSet32{
				trieNodeSet32FromPrefix(unsafeParsePrefix("10.224.24.0/32")),
				trieNodeSet32FromPrefix(unsafeParsePrefix("10.224.24.128/32")),
				trieNodeSet32FromPrefix(unsafeParsePrefix("10.224.24.0/24")),
			},
			in: []Addr{
				unsafeParseAddr("10.224.24.0"),
				unsafeParseAddr("10.224.24.128"),
				unsafeParseAddr("10.224.24.255"),
			},
			out: []Addr{
				unsafeParseAddr("10.224.25.0"),
				unsafeParseAddr("10.224.28.0"),
			},
		},
		{
			description: "union of union",
			sets: []*trieNodeSet32{
				trieNodeSet32FromPrefix(unsafeParsePrefix("10.224.24.0/32")).Union(trieNodeSet32FromPrefix(unsafeParsePrefix("10.224.24.128/32"))),
				trieNodeSet32FromPrefix(unsafeParsePrefix("10.224.24.0/32")).Union(trieNodeSet32FromPrefix(unsafeParsePrefix("10.224.24.128/32"))),
			},
			in: []Addr{
				unsafeParseAddr("10.224.24.0"),
				unsafeParseAddr("10.224.24.128"),
			},
			out: []Addr{
				unsafeParseAddr("10.224.24.255"),
				unsafeParseAddr("10.224.25.0"),
				unsafeParseAddr("10.224.28.0"),
			},
		},
		{
			description: "reverse unions",
			sets: []*trieNodeSet32{
				a.Union(b),
				b.Union(a),
			},
			in: []Addr{
				unsafeParseAddr("10.224.24.0"),
				unsafeParseAddr("10.224.24.128"),
			},
			out: []Addr{
				unsafeParseAddr("10.224.24.255"),
				unsafeParseAddr("10.224.25.0"),
				unsafeParseAddr("10.224.28.0"),
			},
		},
		{
			description: "progressively super",
			sets: []*trieNodeSet32{
				trieNodeSet32FromPrefix(unsafeParsePrefix("10.224.24.0/32")),
				trieNodeSet32FromPrefix(unsafeParsePrefix("10.224.24.0/31")),
				trieNodeSet32FromPrefix(unsafeParsePrefix("10.224.24.0/30")),
				trieNodeSet32FromPrefix(unsafeParsePrefix("10.224.24.0/29")),
				trieNodeSet32FromPrefix(unsafeParsePrefix("10.224.24.0/28")),
				trieNodeSet32FromPrefix(unsafeParsePrefix("10.224.24.0/27")),
				trieNodeSet32FromPrefix(unsafeParsePrefix("10.224.24.0/26")),
				trieNodeSet32FromPrefix(unsafeParsePrefix("10.224.24.0/25")),
				trieNodeSet32FromPrefix(unsafeParsePrefix("10.224.24.0/24")),
				trieNodeSet32FromPrefix(unsafeParsePrefix("10.224.24.0/23")),
				trieNodeSet32FromPrefix(unsafeParsePrefix("10.224.24.0/22")),
				trieNodeSet32FromPrefix(unsafeParsePrefix("10.224.24.0/21")),
				trieNodeSet32FromPrefix(unsafeParsePrefix("10.224.24.0/20")),
			},
			in: []Addr{
				unsafeParseAddr("10.224.16.0"),
				unsafeParseAddr("10.224.17.0"),
				unsafeParseAddr("10.224.18.0"),
				unsafeParseAddr("10.224.19.0"),
				unsafeParseAddr("10.224.20.0"),
				unsafeParseAddr("10.224.21.0"),
				unsafeParseAddr("10.224.22.0"),
				unsafeParseAddr("10.224.23.0"),
				unsafeParseAddr("10.224.24.0"),
				unsafeParseAddr("10.224.25.0"),
				unsafeParseAddr("10.224.26.0"),
				unsafeParseAddr("10.224.27.0"),
				unsafeParseAddr("10.224.28.0"),
				unsafeParseAddr("10.224.29.0"),
				unsafeParseAddr("10.224.30.0"),
				unsafeParseAddr("10.224.31.0"),
			},
			out: []Addr{
				unsafeParseAddr("10.224.15.0"),
				unsafeParseAddr("10.224.32.0"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			test := func(sets []*trieNodeSet32) func(*testing.T) {
				return func(t *testing.T) {
					var s *trieNodeSet32
					for _, set := range sets {
						s = s.Union(set)
						assert.Equal(t, s, s.Union(s))
						assert.True(t, s.isValid())
					}
					for _, addr := range tt.in {
						t.Run(addr.String(), func(t *testing.T) {
							assert.NotNil(t, s.Match(addr.Prefix()))
						})
					}
					for _, addr := range tt.out {
						t.Run(addr.String(), func(t *testing.T) {
							assert.Nil(t, s.Match(addr.Prefix()))
						})
					}
				}
			}
			t.Run("forward", test(tt.sets))
			t.Run("backward", test(([]*trieNodeSet32)(reverse(tt.sets))))
		})
	}
}

func TestInsertOverlappingSet32(t *testing.T) {
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
		{
			desc: "0 and 8",
			a:    Prefix{Addr{0}, 0},
			b:    Prefix{unsafeParseAddr("10.0.0.0"), 8},
			c:    Prefix{unsafeParseAddr("10.0.0.0"), 8},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			// This test inserts the three given nodes in the order given and
			// checks that they are found in the resulting trie
			subTest := func(first, second, third Prefix) func(t *testing.T) {
				return func(t *testing.T) {
					var trie *trieNodeSet32

					trie = trie.Insert(first)
					assert.NotNil(t, trie.Match(first))
					assert.True(t, trie.isValid())
					assert.Equal(t, 1, trie.height())
					assert.Equal(t, 1, trie.NumNodes())

					trie = trie.Insert(second)
					assert.NotNil(t, trie.Match(second))
					assert.True(t, trie.isValid())
					assert.Equal(t, 1, trie.height())
					assert.Equal(t, 1, trie.NumNodes())

					trie = trie.Insert(third)
					assert.NotNil(t, trie.Match(third))
					assert.True(t, trie.isValid())
					assert.Equal(t, 1, trie.height())
					assert.Equal(t, 1, trie.NumNodes())
				}
			}
			t.Run("forward", subTest(tt.a, tt.b, tt.c))
			t.Run("backward", subTest(tt.c, tt.b, tt.a))
		})
	}
}

// https://stackoverflow.com/a/61218109
func reverse(s []*trieNodeSet32) []*trieNodeSet32 {
	a := make([]*trieNodeSet32, len(s))
	copy(a, s)

	for i := len(a)/2 - 1; i >= 0; i-- {
		opp := len(a) - 1 - i
		a[i], a[opp] = a[opp], a[i]
	}

	return a
}

func TestSetRemove(t *testing.T) {
	t.Run("test remove from nil", func(t *testing.T) {
		var trie *trieNodeSet32

		assert.Equal(t, int64(0), trie.Size())

		trie = trie.Remove(unsafeParsePrefix("203.0.113.0/24"))
		assert.Equal(t, int64(0), trie.Size())
	})
	t.Run("test remove all", func(t *testing.T) {
		var trie *trieNodeSet32

		prefix := unsafeParsePrefix("203.0.113.0/24")
		trie = trie.Insert(prefix)
		assert.Equal(t, int64(256), trie.Size())

		trie = trie.Remove(prefix)
		assert.Equal(t, int64(0), trie.Size())
	})
	t.Run("test remove one address", func(t *testing.T) {
		var trie *trieNodeSet32

		prefix := unsafeParsePrefix("203.0.113.0/24")
		trie = trie.Insert(prefix)
		assert.Equal(t, int64(256), trie.Size())

		trie = trie.Remove(unsafeParsePrefix("203.0.113.233/32"))
		assert.Equal(t, int64(255), trie.Size())
	})
	t.Run("test remove half", func(t *testing.T) {
		var trie *trieNodeSet32

		prefix := unsafeParsePrefix("203.0.113.0/24")
		trie = trie.Insert(prefix)
		assert.Equal(t, int64(256), trie.Size())

		trie = trie.Remove(unsafeParsePrefix("203.0.113.128/25"))
		assert.Equal(t, int64(128), trie.Size())
	})
	t.Run("test remove more", func(t *testing.T) {
		var trie *trieNodeSet32

		prefix := unsafeParsePrefix("203.0.113.0/24")
		trie = trie.Insert(prefix)
		assert.Equal(t, int64(256), trie.Size())

		trie = trie.Remove(unsafeParsePrefix("203.0.112.0/23"))
		assert.Equal(t, int64(0), trie.Size())
	})
}

func TestSetIntersect(t *testing.T) {
	t.Run("with nil", func(t *testing.T) {
		var one, two, three *trieNodeSet32

		three = three.Insert(unsafeParsePrefix("203.0.113.0/24"))

		assert.Equal(t, int64(0), one.Size())
		assert.Equal(t, int64(0), two.Size())
		assert.Equal(t, int64(256), three.Size())

		assert.Equal(t, int64(0), one.Intersect(two).Size())
		assert.Equal(t, int64(0), one.Intersect(three).Size())
		assert.Equal(t, int64(0), three.Intersect(one).Size())
	})
	t.Run("disjoint", func(t *testing.T) {
		var one, two *trieNodeSet32

		one = one.Insert(unsafeParsePrefix("203.0.113.0/27"))
		two = two.Insert(unsafeParsePrefix("203.0.113.128/25"))
		assert.Equal(t, int64(0), one.Intersect(two).Size())
		assert.Equal(t, int64(0), two.Intersect(one).Size())
	})
	t.Run("subset", func(t *testing.T) {
		var one, two *trieNodeSet32

		one = one.Insert(unsafeParsePrefix("203.0.113.0/24"))
		two = two.Insert(unsafeParsePrefix("203.0.113.128/25"))
		result := one.Intersect(two)
		assert.Equal(t, int64(128), result.Size())
		assert.Equal(t, int64(128), two.Intersect(one).Size())
		assert.Nil(t, result.Match(unsafeParsePrefix("203.0.113.117/32")))
		assert.NotNil(t, result.Match(unsafeParsePrefix("203.0.113.217/32")))
	})
	t.Run("recursive", func(t *testing.T) {
		var one, two *trieNodeSet32
		one = one.Insert(unsafeParsePrefix("198.51.100.0/24"))
		one = one.Insert(unsafeParsePrefix("203.0.113.0/24"))
		two = two.Insert(unsafeParsePrefix("203.0.113.128/25"))

		result := one.Intersect(two)
		assert.Equal(t, int64(128), result.Size())
		assert.Nil(t, result.Match(unsafeParsePrefix("198.51.100.0/24")))
		assert.Nil(t, result.Match(unsafeParsePrefix("203.0.113.0/25")))
		assert.NotNil(t, result.Match(unsafeParsePrefix("203.0.113.128/25")))

		result = two.Intersect(one)
		assert.Equal(t, int64(128), result.Size())
		assert.Nil(t, result.Match(unsafeParsePrefix("198.51.100.0/24")))
		assert.Nil(t, result.Match(unsafeParsePrefix("203.0.113.0/25")))
		assert.NotNil(t, result.Match(unsafeParsePrefix("203.0.113.128/25")))
	})
}
