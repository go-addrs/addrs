package ipv4

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func printTrieSet(trie *setNode) {
	printTrie((*trieNode)(trie))
}

func TestTrieNodeSet32Union(t *testing.T) {
	a := setNodeFromPrefix(unsafeParsePrefix("10.224.24.0/32"))
	b := setNodeFromPrefix(unsafeParsePrefix("10.224.24.128/32"))
	tests := []struct {
		description string
		sets        []*setNode
		in, out     []Address
	}{
		{
			description: "two adjacent",
			sets: []*setNode{
				setNodeFromPrefix(unsafeParsePrefix("10.224.24.0/25")),
				setNodeFromPrefix(unsafeParsePrefix("10.224.24.128/25")),
			},
			in: []Address{
				unsafeParseAddress("10.224.24.0"),
				unsafeParseAddress("10.224.24.255"),
				unsafeParseAddress("10.224.24.127"),
				unsafeParseAddress("10.224.24.128"),
			},
			out: []Address{
				unsafeParseAddress("10.224.23.255"),
				unsafeParseAddress("10.224.25.0"),
			},
		},
		{
			description: "nil",
			sets: []*setNode{
				nil,
			},
			in: []Address{},
			out: []Address{
				unsafeParseAddress("10.224.23.117"),
				unsafeParseAddress("200.193.25.0"),
			},
		},
		{
			description: "not nil then nil",
			sets: []*setNode{
				setNodeFromPrefix(unsafeParsePrefix("10.224.24.0/32")),
				nil,
			},
			in: []Address{
				unsafeParseAddress("10.224.24.0"),
			},
			out: []Address{
				unsafeParseAddress("10.224.23.255"),
				unsafeParseAddress("200.193.24.1"),
			},
		},
		{
			description: "same",
			sets: []*setNode{
				setNodeFromPrefix(unsafeParsePrefix("10.224.24.0/32")),
				setNodeFromPrefix(unsafeParsePrefix("10.224.24.0/32")),
			},
			in: []Address{
				unsafeParseAddress("10.224.24.0"),
			},
		},
		{
			description: "different then same",
			sets: []*setNode{
				setNodeFromPrefix(unsafeParsePrefix("10.224.29.0/32")),
				setNodeFromPrefix(unsafeParsePrefix("10.224.24.0/32")),
				setNodeFromPrefix(unsafeParsePrefix("10.224.24.0/32")),
			},
			in: []Address{
				unsafeParseAddress("10.224.24.0"),
				unsafeParseAddress("10.224.29.0"),
			},
		},
		{
			description: "duplicates",
			sets: []*setNode{
				setNodeFromPrefix(unsafeParsePrefix("10.224.24.0/32")),
				setNodeFromPrefix(unsafeParsePrefix("10.224.24.128/32")),
				setNodeFromPrefix(unsafeParsePrefix("10.224.24.0/24")),
			},
			in: []Address{
				unsafeParseAddress("10.224.24.0"),
				unsafeParseAddress("10.224.24.128"),
				unsafeParseAddress("10.224.24.255"),
			},
			out: []Address{
				unsafeParseAddress("10.224.25.0"),
				unsafeParseAddress("10.224.28.0"),
			},
		},
		{
			description: "union of union",
			sets: []*setNode{
				setNodeFromPrefix(unsafeParsePrefix("10.224.24.0/32")).Union(setNodeFromPrefix(unsafeParsePrefix("10.224.24.128/32"))),
				setNodeFromPrefix(unsafeParsePrefix("10.224.24.0/32")).Union(setNodeFromPrefix(unsafeParsePrefix("10.224.24.128/32"))),
			},
			in: []Address{
				unsafeParseAddress("10.224.24.0"),
				unsafeParseAddress("10.224.24.128"),
			},
			out: []Address{
				unsafeParseAddress("10.224.24.255"),
				unsafeParseAddress("10.224.25.0"),
				unsafeParseAddress("10.224.28.0"),
			},
		},
		{
			description: "reverse unions",
			sets: []*setNode{
				a.Union(b),
				b.Union(a),
			},
			in: []Address{
				unsafeParseAddress("10.224.24.0"),
				unsafeParseAddress("10.224.24.128"),
			},
			out: []Address{
				unsafeParseAddress("10.224.24.255"),
				unsafeParseAddress("10.224.25.0"),
				unsafeParseAddress("10.224.28.0"),
			},
		},
		{
			description: "progressively super",
			sets: []*setNode{
				setNodeFromPrefix(unsafeParsePrefix("10.224.24.0/32")),
				setNodeFromPrefix(unsafeParsePrefix("10.224.24.0/31")),
				setNodeFromPrefix(unsafeParsePrefix("10.224.24.0/30")),
				setNodeFromPrefix(unsafeParsePrefix("10.224.24.0/29")),
				setNodeFromPrefix(unsafeParsePrefix("10.224.24.0/28")),
				setNodeFromPrefix(unsafeParsePrefix("10.224.24.0/27")),
				setNodeFromPrefix(unsafeParsePrefix("10.224.24.0/26")),
				setNodeFromPrefix(unsafeParsePrefix("10.224.24.0/25")),
				setNodeFromPrefix(unsafeParsePrefix("10.224.24.0/24")),
				setNodeFromPrefix(unsafeParsePrefix("10.224.24.0/23")),
				setNodeFromPrefix(unsafeParsePrefix("10.224.24.0/22")),
				setNodeFromPrefix(unsafeParsePrefix("10.224.24.0/21")),
				setNodeFromPrefix(unsafeParsePrefix("10.224.24.0/20")),
			},
			in: []Address{
				unsafeParseAddress("10.224.16.0"),
				unsafeParseAddress("10.224.17.0"),
				unsafeParseAddress("10.224.18.0"),
				unsafeParseAddress("10.224.19.0"),
				unsafeParseAddress("10.224.20.0"),
				unsafeParseAddress("10.224.21.0"),
				unsafeParseAddress("10.224.22.0"),
				unsafeParseAddress("10.224.23.0"),
				unsafeParseAddress("10.224.24.0"),
				unsafeParseAddress("10.224.25.0"),
				unsafeParseAddress("10.224.26.0"),
				unsafeParseAddress("10.224.27.0"),
				unsafeParseAddress("10.224.28.0"),
				unsafeParseAddress("10.224.29.0"),
				unsafeParseAddress("10.224.30.0"),
				unsafeParseAddress("10.224.31.0"),
			},
			out: []Address{
				unsafeParseAddress("10.224.15.0"),
				unsafeParseAddress("10.224.32.0"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			test := func(sets []*setNode) func(*testing.T) {
				return func(t *testing.T) {
					var s *setNode
					for _, set := range sets {
						s = s.Union(set)
						assert.Equal(t, s, s.Union(s))
						assert.True(t, s.isValid())
					}
					for _, addr := range tt.in {
						t.Run(addr.String(), func(t *testing.T) {
							assert.NotNil(t, s.Match(addr.HostPrefix()))
						})
					}
					for _, addr := range tt.out {
						t.Run(addr.String(), func(t *testing.T) {
							assert.Nil(t, s.Match(addr.HostPrefix()))
						})
					}
				}
			}
			t.Run("forward", test(tt.sets))
			t.Run("backward", test(([]*setNode)(reverse(tt.sets))))
		})
	}
	t.Run("not active", func(t *testing.T) {
		var one, two *setNode
		one = one.Insert(unsafeParsePrefix("198.51.100.0/25"))
		one = one.Insert(unsafeParsePrefix("203.0.113.0/25"))
		printTrieSet(one)

		two = two.Insert(unsafeParsePrefix("198.51.100.128/25"))
		two = two.Insert(unsafeParsePrefix("203.0.113.128/25"))
		printTrieSet(two)

		result := one.Union(two)
		assert.Equal(t, int64(512), result.Size())
		assert.NotNil(t, result.Match(unsafeParsePrefix("198.51.100.0/24")))
		assert.NotNil(t, result.Match(unsafeParsePrefix("203.0.113.0/24")))
		printTrieSet(result)
		assert.Nil(t, result.Match(unsafeParsePrefix("192.0.0.0/4")))
	})
}

func TestInsertOverlappingSet(t *testing.T) {
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
		{
			desc: "0 and 8",
			a:    Prefix{Address{0}, 0},
			b:    Prefix{unsafeParseAddress("10.0.0.0"), 8},
			c:    Prefix{unsafeParseAddress("10.0.0.0"), 8},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			// This test inserts the three given nodes in the order given and
			// checks that they are found in the resulting trie
			subTest := func(first, second, third Prefix) func(t *testing.T) {
				return func(t *testing.T) {
					var trie *setNode

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
func reverse(s []*setNode) []*setNode {
	a := make([]*setNode, len(s))
	copy(a, s)

	for i := len(a)/2 - 1; i >= 0; i-- {
		opp := len(a) - 1 - i
		a[i], a[opp] = a[opp], a[i]
	}

	return a
}

func TestSetRemove(t *testing.T) {
	t.Run("test remove from nil", func(t *testing.T) {
		var trie *setNode

		assert.Equal(t, int64(0), trie.Size())

		trie = trie.Remove(unsafeParsePrefix("203.0.113.0/24"))
		assert.Equal(t, int64(0), trie.Size())
	})
	t.Run("test remove all", func(t *testing.T) {
		var trie *setNode

		prefix := unsafeParsePrefix("203.0.113.0/24")
		trie = trie.Insert(prefix)
		assert.Equal(t, int64(256), trie.Size())

		trie = trie.Remove(prefix)
		assert.Equal(t, int64(0), trie.Size())
	})
	t.Run("test remove one address", func(t *testing.T) {
		var trie *setNode

		prefix := unsafeParsePrefix("203.0.113.0/24")
		trie = trie.Insert(prefix)
		assert.Equal(t, int64(256), trie.Size())

		trie = trie.Remove(unsafeParsePrefix("203.0.113.233/32"))
		assert.Equal(t, int64(255), trie.Size())
	})
	t.Run("test remove half", func(t *testing.T) {
		var trie *setNode

		prefix := unsafeParsePrefix("203.0.113.0/24")
		trie = trie.Insert(prefix)
		assert.Equal(t, int64(256), trie.Size())

		trie = trie.Remove(unsafeParsePrefix("203.0.113.128/25"))
		assert.Equal(t, int64(128), trie.Size())
	})
	t.Run("test remove more", func(t *testing.T) {
		var trie *setNode

		prefix := unsafeParsePrefix("203.0.113.0/24")
		trie = trie.Insert(prefix)
		assert.Equal(t, int64(256), trie.Size())

		trie = trie.Remove(unsafeParsePrefix("203.0.112.0/23"))
		assert.Equal(t, int64(0), trie.Size())
	})
}

func TestSetIntersect(t *testing.T) {
	t.Run("with nil", func(t *testing.T) {
		var one, two, three *setNode

		three = three.Insert(unsafeParsePrefix("203.0.113.0/24"))

		assert.Equal(t, int64(0), one.Size())
		assert.Equal(t, int64(0), two.Size())
		assert.Equal(t, int64(256), three.Size())

		assert.Equal(t, int64(0), one.Intersect(two).Size())
		assert.Equal(t, int64(0), one.Intersect(three).Size())
		assert.Equal(t, int64(0), three.Intersect(one).Size())
	})
	t.Run("disjoint", func(t *testing.T) {
		var one, two *setNode

		one = one.Insert(unsafeParsePrefix("203.0.113.0/27"))
		two = two.Insert(unsafeParsePrefix("203.0.113.128/25"))
		assert.Equal(t, int64(0), one.Intersect(two).Size())
		assert.Equal(t, int64(0), two.Intersect(one).Size())
	})
	t.Run("subset", func(t *testing.T) {
		var one, two *setNode

		one = one.Insert(unsafeParsePrefix("203.0.113.0/24"))
		two = two.Insert(unsafeParsePrefix("203.0.113.128/25"))
		result := one.Intersect(two)
		assert.Equal(t, int64(128), result.Size())
		assert.Equal(t, int64(128), two.Intersect(one).Size())
		assert.Nil(t, result.Match(unsafeParsePrefix("203.0.113.117/32")))
		assert.NotNil(t, result.Match(unsafeParsePrefix("203.0.113.217/32")))
	})
	t.Run("recursive", func(t *testing.T) {
		var one, two *setNode
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

func TestSetDifference(t *testing.T) {
	t.Run("with nil", func(t *testing.T) {
		var one, two, three *setNode

		three = three.Insert(unsafeParsePrefix("203.0.113.0/24"))

		assert.Equal(t, int64(0), one.Size())
		assert.Equal(t, int64(0), two.Size())
		assert.Equal(t, int64(256), three.Size())

		assert.Equal(t, int64(0), one.Difference(two).Size())
		assert.Equal(t, int64(0), one.Difference(three).Size())
		assert.Equal(t, int64(256), three.Difference(one).Size())
	})
	t.Run("disjoint", func(t *testing.T) {
		var one, two *setNode

		one = one.Insert(unsafeParsePrefix("203.0.113.0/27"))
		two = two.Insert(unsafeParsePrefix("203.0.113.128/25"))

		result := one.Difference(two)
		assert.Equal(t, int64(32), result.Size())
		assert.NotNil(t, result.Match(unsafeParsePrefix("203.0.113.0/27")))
		assert.Nil(t, result.Match(unsafeParsePrefix("203.0.113.128/25")))

		result = two.Difference(one)
		assert.Equal(t, int64(128), result.Size())
		assert.Nil(t, result.Match(unsafeParsePrefix("203.0.113.0/27")))
		assert.NotNil(t, result.Match(unsafeParsePrefix("203.0.113.128/25")))
	})
	t.Run("subset", func(t *testing.T) {
		var one, two *setNode

		one = one.Insert(unsafeParsePrefix("203.0.113.0/24"))
		two = two.Insert(unsafeParsePrefix("203.0.113.128/25"))

		result := one.Difference(two)
		assert.Equal(t, int64(128), result.Size())
		assert.NotNil(t, result.Match(unsafeParsePrefix("203.0.113.117/32")))
		assert.Nil(t, result.Match(unsafeParsePrefix("203.0.113.217/32")))

		assert.Equal(t, int64(128), two.Difference(one).Size())
	})
	t.Run("recursive", func(t *testing.T) {
		var one, two *setNode
		one = one.Insert(unsafeParsePrefix("198.51.100.0/24"))
		one = one.Insert(unsafeParsePrefix("203.0.113.0/24"))
		two = two.Insert(unsafeParsePrefix("203.0.113.128/25"))

		result := one.Difference(two)
		assert.Equal(t, int64(384), result.Size())
		assert.NotNil(t, result.Match(unsafeParsePrefix("198.51.100.0/24")))
		assert.NotNil(t, result.Match(unsafeParsePrefix("203.0.113.0/25")))
		assert.Nil(t, result.Match(unsafeParsePrefix("203.0.113.128/25")))

		result = two.Difference(one)
		assert.Equal(t, int64(128), result.Size())
		assert.Nil(t, result.Match(unsafeParsePrefix("198.51.100.0/24")))
		assert.Nil(t, result.Match(unsafeParsePrefix("203.0.113.0/25")))
		assert.NotNil(t, result.Match(unsafeParsePrefix("203.0.113.128/25")))
	})
	t.Run("no difference", func(t *testing.T) {
		var one, two *setNode
		one = one.Insert(unsafeParsePrefix("198.51.100.0/24"))
		one = one.Insert(unsafeParsePrefix("203.0.113.0/24"))

		two = two.Insert(unsafeParsePrefix("192.0.2.0/24"))

		result := one.Difference(two)
		assert.Equal(t, int64(512), result.Size())
		assert.NotNil(t, result.Match(unsafeParsePrefix("198.51.100.0/24")))
		assert.NotNil(t, result.Match(unsafeParsePrefix("203.0.113.0/24")))
		assert.Nil(t, result.Match(unsafeParsePrefix("192.0.2.0/24")))
	})
	t.Run("not active", func(t *testing.T) {
		var one, two *setNode
		one = one.Insert(unsafeParsePrefix("192.0.0.0/4"))

		two = two.Insert(unsafeParsePrefix("198.51.100.0/24"))
		two = two.Insert(unsafeParsePrefix("203.0.113.0/24"))

		result := one.Difference(two)
		assert.Equal(t, int64(268434944), result.Size())
		assert.Nil(t, result.Match(unsafeParsePrefix("198.51.100.0/24")))
		assert.Nil(t, result.Match(unsafeParsePrefix("203.0.113.0/24")))
		assert.NotNil(t, result.Match(unsafeParsePrefix("197.51.100.128/25")))
		assert.NotNil(t, result.Match(unsafeParsePrefix("204.0.113.128/25")))
	})
}
