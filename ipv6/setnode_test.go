package ipv6

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func PrintTrieSet(trie *setNode) {
	PrintTrie((*trieNode)(trie))
}

func TestTrieNodeSet32Union(t *testing.T) {
	a := setNodeFromPrefix(_p("2001::224:24:0/128"))
	b := setNodeFromPrefix(_p("2001::224:24:8000/128"))
	tests := []struct {
		description string
		sets        []*setNode
		in, out     []Address
	}{
		{
			description: "two adjacent",
			sets: []*setNode{
				setNodeFromPrefix(_p("2001::224:24:0/112")),
				setNodeFromPrefix(_p("2001::224:24:8000/113")),
			},
			in: []Address{
				_a("2001::224:24:0"),
				_a("2001::224:24:ffff"),
				_a("2001::224:24:7fff"),
				_a("2001::224:24:8000"),
			},
			out: []Address{
				_a("2001::224:23:ffff"),
				_a("2001::224:25:0"),
			},
		},
		{
			description: "nil",
			sets: []*setNode{
				nil,
			},
			in: []Address{},
			out: []Address{
				_a("2001::224:23:117"),
				_a("2001:193::25:0"),
			},
		},
		{
			description: "not nil then nil",
			sets: []*setNode{
				setNodeFromPrefix(_p("2001::224:24:0/128")),
				nil,
			},
			in: []Address{
				_a("2001::224:24:0"),
			},
			out: []Address{
				_a("2001::224:23:ffff"),
				_a("2001::224:24:1"),
			},
		},
		{
			description: "same",
			sets: []*setNode{
				setNodeFromPrefix(_p("2001::224:24:0/128")),
				setNodeFromPrefix(_p("2001::224:24:0/128")),
			},
			in: []Address{
				_a("2001::224:24:0"),
			},
		},
		{
			description: "different then same",
			sets: []*setNode{
				setNodeFromPrefix(_p("2001::224:29:0/128")),
				setNodeFromPrefix(_p("2001::224:24:0/128")),
				//setNodeFromPrefix(_p("2001::224:24:0/128")),
			},
			in: []Address{
				_a("2001::224:24:0"),
				_a("2001::224:29:0"),
			},
		},
		{
			description: "duplicates",
			sets: []*setNode{
				setNodeFromPrefix(_p("2001::224:24:0/128")),
				setNodeFromPrefix(_p("2001::224:24:8000/128")),
				setNodeFromPrefix(_p("2001::224:24:0/112")),
			},
			in: []Address{
				_a("2001::224:24:0"),
				_a("2001::224:24:8000"),
				_a("2001::224:24:ffff"),
			},
			out: []Address{
				_a("2001::224:25:0"),
				_a("2001::224:28:0"),
			},
		},
		{
			description: "union of union",
			sets: []*setNode{
				setNodeFromPrefix(_p("2001::224:24:0/128")).Union(setNodeFromPrefix(_p("2001::224:24:8000/128"))),
				setNodeFromPrefix(_p("2001::224:24:0/128")).Union(setNodeFromPrefix(_p("2001::224:24:8000/128"))),
			},
			in: []Address{
				_a("2001::224:24:0"),
				_a("2001::224:24:8000"),
			},
			out: []Address{
				_a("2001::224:24:ffff"),
				_a("2001::224:25:0"),
				_a("2001::224:28:0"),
			},
		},
		{
			description: "reverse unions",
			sets: []*setNode{
				a.Union(b),
				b.Union(a),
			},
			in: []Address{
				_a("2001::224:24:0"),
				_a("2001::224:24:8000"),
			},
			out: []Address{
				_a("2001::224:24:ffff"),
				_a("2001::224:25:0"),
				_a("2001::224:28:0"),
			},
		},
		{
			description: "progressively super",
			sets: []*setNode{
				setNodeFromPrefix(_p("2001::224:24:0/128")),
				setNodeFromPrefix(_p("2001::224:24:0/127")),
				setNodeFromPrefix(_p("2001::224:24:0/126")),
				setNodeFromPrefix(_p("2001::224:24:0/125")),
				setNodeFromPrefix(_p("2001::224:24:0/124")),
				setNodeFromPrefix(_p("2001::224:24:0/123")),
				setNodeFromPrefix(_p("2001::224:24:0/122")),
				setNodeFromPrefix(_p("2001::224:24:0/121")),
				setNodeFromPrefix(_p("2001::224:24:0/120")),
				setNodeFromPrefix(_p("2001::224:24:0/119")),
				setNodeFromPrefix(_p("2001::224:24:0/118")),
				setNodeFromPrefix(_p("2001::224:24:0/117")),
				setNodeFromPrefix(_p("2001::224:24:0/116")),
				setNodeFromPrefix(_p("2001::224:24:0/115")),
				setNodeFromPrefix(_p("2001::224:24:0/114")),
				setNodeFromPrefix(_p("2001::224:24:0/113")),
				setNodeFromPrefix(_p("2001::224:24:0/112")),
				setNodeFromPrefix(_p("2001::224:24:0/111")),
				setNodeFromPrefix(_p("2001::224:24:0/110")),
				setNodeFromPrefix(_p("2001::224:24:0/109")),
				setNodeFromPrefix(_p("2001::224:24:0/108")),
			},
			in: []Address{
				_a("2001::224:20:0"),
				_a("2001::224:21:0"),
				_a("2001::224:22:0"),
				_a("2001::224:23:0"),
				_a("2001::224:24:0"),
				_a("2001::224:25:0"),
				_a("2001::224:26:0"),
				_a("2001::224:27:0"),
				_a("2001::224:28:0"),
				_a("2001::224:29:0"),
				_a("2001::224:2a:0"),
				_a("2001::224:2b:0"),
				_a("2001::224:2c:0"),
				_a("2001::224:2d:0"),
				_a("2001::224:2e:0"),
				_a("2001::224:2f:0"),
			},
			out: []Address{
				_a("2001::224:19:0"),
				_a("2001::224:30:0"),
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
			t.Run("backward", test(([]*setNode)(reverse(tt.sets))))
		})
	}
	t.Run("not active", func(t *testing.T) {
		var one, two *setNode
		one = one.Insert(_p("a8d2::100:0/113"))
		one = one.Insert(_p("2003::113:0/113"))

		two = two.Insert(_p("a8d2::100:8000/113"))
		two = two.Insert(_p("2003::113:8000/113"))

		result := one.Union(two)
		assert.False(t, result.IsEmpty())
		assert.NotNil(t, result.Match(_p("a8d2::100:0/112")))
		assert.NotNil(t, result.Match(_p("2003::113:0/112")))
		assert.Nil(t, result.Match(_p("2001::/16")))
	})
}

func TestInsertOverlappingSet(t *testing.T) {
	tests := []struct {
		desc    string
		a, b, c Prefix
	}{
		{
			desc: "64 and 112",
			a:    Prefix{_a("2001:0:0:200::"), 64},
			b:    Prefix{_a("2001:0:0:200:20::"), 112},
			c:    Prefix{_a("2001:0:0:200:20::"), 128},
		},
		{
			desc: "65 and 115",
			a:    Prefix{_a("2001:0:0:200::"), 65},
			b:    Prefix{_a("2001:0:0:200::E000"), 115},
			c:    Prefix{_a("2001:0:0:200::f800"), 127},
		},
		{
			desc: "0 and 16",
			a:    Prefix{_a("::"), 0},
			b:    Prefix{_a("2001::"), 16},
			c:    Prefix{_a("2001:10::"), 32},
		},
		{
			desc: "0 and 16",
			a:    Prefix{_a("::"), 0},
			b:    Prefix{_a("2001::"), 16},
			c:    Prefix{_a("2001::"), 16},
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
					assert.Equal(t, int64(1), trie.NumNodes())

					trie = trie.Insert(second)
					assert.NotNil(t, trie.Match(second))
					assert.True(t, trie.isValid())
					assert.Equal(t, 1, trie.height())
					assert.Equal(t, int64(1), trie.NumNodes())

					trie = trie.Insert(third)
					assert.NotNil(t, trie.Match(third))
					assert.True(t, trie.isValid())
					assert.Equal(t, 1, trie.height())
					assert.Equal(t, int64(1), trie.NumNodes())
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

		assert.True(t, trie.IsEmpty())

		trie = trie.Remove(_p("2001::1234:0/112"))
		assert.True(t, trie.IsEmpty())
	})
	t.Run("test remove all", func(t *testing.T) {
		var trie *setNode

		prefix := _p("2001::1234:0/112")
		trie = trie.Insert(prefix)
		assert.False(t, trie.IsEmpty())

		trie = trie.Remove(prefix)
		assert.True(t, trie.IsEmpty())
	})
	t.Run("test remove one address", func(t *testing.T) {
		var trie *setNode

		prefix := _p("2001::1234:0/112")
		trie = trie.Insert(prefix)
		assert.False(t, trie.IsEmpty())

		trie = trie.Remove(_p("2001::1234:233/128"))
		assert.False(t, trie.IsEmpty())
	})
	t.Run("test remove half", func(t *testing.T) {
		var trie *setNode

		prefix := _p("2001::1234:0/112")
		trie = trie.Insert(prefix)
		assert.False(t, trie.IsEmpty())

		trie = trie.Remove(_p("2001::1234:8000/113"))
		assert.False(t, trie.IsEmpty())
	})
	t.Run("test remove more", func(t *testing.T) {
		var trie *setNode

		prefix := _p("2001::1234:0/112")
		trie = trie.Insert(prefix)
		assert.False(t, trie.IsEmpty())

		trie = trie.Remove(_p("2001::1233:0/111"))
		assert.False(t, trie.IsEmpty())
	})
}

func TestSetIntersect(t *testing.T) {
	t.Run("with nil", func(t *testing.T) {
		var one, two, three *setNode

		three = three.Insert(_p("2001::1234:0/112"))

		assert.True(t, one.IsEmpty())
		assert.True(t, two.IsEmpty())
		assert.False(t, three.IsEmpty())

		assert.True(t, one.Intersect(two).IsEmpty())
		assert.True(t, one.Intersect(three).IsEmpty())
		assert.True(t, three.Intersect(one).IsEmpty())
	})
	t.Run("disjoint", func(t *testing.T) {
		var one, two *setNode

		one = one.Insert(_p("2001::1234:0/115"))
		two = two.Insert(_p("2001::1234:8000/113"))
		assert.True(t, one.Intersect(two).IsEmpty())
		assert.True(t, two.Intersect(one).IsEmpty())
	})
	t.Run("subset", func(t *testing.T) {
		var one, two *setNode

		one = one.Insert(_p("2001::1234:0/112"))
		two = two.Insert(_p("2001::1234:8000/113"))
		result := one.Intersect(two)
		assert.False(t, result.IsEmpty())
		assert.False(t, two.Intersect(one).IsEmpty())
		assert.Nil(t, result.Match(_p("2001::1234:7100/128")))
		assert.NotNil(t, result.Match(_p("2001::1234:8100/128")))
	})
	t.Run("recursive", func(t *testing.T) {
		var one, two *setNode
		one = one.Insert(_p("198:51::100:0/112"))
		one = one.Insert(_p("2001::1234:0/112"))
		two = two.Insert(_p("2001::1234:8000/113"))

		result := one.Intersect(two)
		assert.False(t, result.IsEmpty())
		assert.Nil(t, result.Match(_p("198:51::100:0/112")))
		assert.Nil(t, result.Match(_p("2001::1234:0/113")))
		assert.NotNil(t, result.Match(_p("2001::1234:8000/113")))

		result = two.Intersect(one)
		assert.False(t, result.IsEmpty())
		assert.Nil(t, result.Match(_p("198:51::100:0/112")))
		assert.Nil(t, result.Match(_p("2001::1234:0/113")))
		assert.NotNil(t, result.Match(_p("2001::1234:8000/113")))
	})
}

func TestSetDifference(t *testing.T) {
	t.Run("with nil", func(t *testing.T) {
		var one, two, three *setNode

		three = three.Insert(_p("2001::1234:0/112"))

		assert.True(t, one.IsEmpty())
		assert.True(t, two.IsEmpty())
		assert.False(t, three.IsEmpty())

		assert.True(t, one.Difference(two).IsEmpty())
		assert.True(t, one.Difference(three).IsEmpty())
		assert.False(t, three.Difference(one).IsEmpty())
	})
	t.Run("disjoint", func(t *testing.T) {
		var one, two *setNode

		one = one.Insert(_p("2001::1234:0/115"))
		two = two.Insert(_p("2001::1234:8000/113"))

		result := one.Difference(two)
		assert.False(t, result.IsEmpty())
		assert.NotNil(t, result.Match(_p("2001::1234:0/115")))
		assert.Nil(t, result.Match(_p("2001::1234:8000/113")))

		result = two.Difference(one)
		assert.False(t, result.IsEmpty())
		assert.Nil(t, result.Match(_p("2001::1234:0/115")))
		assert.NotNil(t, result.Match(_p("2001::1234:8000/113")))
	})
	t.Run("subset", func(t *testing.T) {
		var one, two *setNode

		one = one.Insert(_p("2001::1234:0/112"))
		two = two.Insert(_p("2001::1234:8000/113"))

		result := one.Difference(two)
		assert.False(t, result.IsEmpty())
		assert.NotNil(t, result.Match(_p("2001::1234:7100/128")))
		assert.Nil(t, result.Match(_p("2001::1234:8100/128")))

		assert.True(t, two.Difference(one).IsEmpty())
	})
	t.Run("recursive", func(t *testing.T) {
		var one, two *setNode
		one = one.Insert(_p("198:51::100:0/112"))
		one = one.Insert(_p("2001::1234:0/112"))
		two = two.Insert(_p("2001::1234:8000/113"))

		result := one.Difference(two)
		assert.False(t, result.IsEmpty())
		assert.NotNil(t, result.Match(_p("198:51::100:0/112")))
		assert.NotNil(t, result.Match(_p("2001::1234:0/113")))
		assert.Nil(t, result.Match(_p("2001::1234:8000/113")))

		result = two.Difference(one)
		assert.True(t, result.IsEmpty())
		assert.Nil(t, result.Match(_p("198:51::100:0/112")))
		assert.Nil(t, result.Match(_p("2001::1234:0/113")))
		assert.Nil(t, result.Match(_p("2001::1234:8000/113")))
	})
	t.Run("no difference", func(t *testing.T) {
		var one, two *setNode
		one = one.Insert(_p("198:51::100:0/112"))
		one = one.Insert(_p("2001::1234:0/112"))

		two = two.Insert(_p("192::2:0/112"))

		result := one.Difference(two)
		assert.False(t, result.IsEmpty())
		assert.NotNil(t, result.Match(_p("198:51::100:0/112")))
		assert.NotNil(t, result.Match(_p("2001::1234:0/112")))
		assert.Nil(t, result.Match(_p("192::2:0/112")))
	})
	t.Run("not active", func(t *testing.T) {
		var one, two *setNode
		one = one.Insert(_p("2192::/8"))

		two = two.Insert(_p("2198:51::100:0/112"))
		two = two.Insert(_p("2001::1234:0/112"))

		result := one.Difference(two)
		fmt.Printf("%v\n", result)
		assert.False(t, result.IsEmpty())
		assert.Nil(t, result.Match(_p("2198:51::100:0/112")))
		assert.Nil(t, result.Match(_p("2001::1234:0/112")))
		assert.NotNil(t, result.Match(_p("2197:51::100:8000/113")))
		assert.NotNil(t, result.Match(_p("2100::1234:8000/113")))
	})
}
