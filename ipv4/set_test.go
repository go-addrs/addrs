package ipv4

import (
	"math"
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetInsertPrefix(t *testing.T) {
	s := NewSet_()
	s.Insert(_p("10.0.0.0/24"))

	assert.True(t, s.Contains(_p("10.0.0.0/24")))
	assert.False(t, s.Contains(_p("10.0.0.0/16")))

	s.Insert(_p("10.0.0.0/16"))

	assert.True(t, s.Contains(_p("10.0.0.0/24")))
	assert.True(t, s.Contains(_p("10.0.0.0/16")))
}

func TestSetRemovePrefix(t *testing.T) {
	s := Set{}.Set_()
	s.Insert(_p("10.0.0.0/16"))

	assert.True(t, s.Contains(_p("10.0.0.0/24")))
	assert.True(t, s.Contains(_p("10.0.0.0/16")))

	s.Remove(_p("10.0.0.0/24"))

	assert.False(t, s.Contains(_p("10.0.0.0/24")))
	assert.False(t, s.Contains(_p("10.0.0.0/16")))
	assert.True(t, s.Contains(_p("10.0.1.0/24")))
}

func TestSetAsReferenceType(t *testing.T) {
	s := NewSet_()

	func(s Set_) {
		s.Insert(_p("10.0.0.0/24"))
	}(s)

	assert.True(t, s.Contains(_p("10.0.0.0/24")))
	assert.False(t, s.Contains(_p("10.0.0.0/16")))

	func(s Set_) {
		s.Insert(_p("10.0.0.0/16"))
	}(s)

	assert.True(t, s.Contains(_p("10.0.0.0/24")))
	assert.True(t, s.Contains(_p("10.0.0.0/16")))
}

func TestSetInsertSet(t *testing.T) {
	a, b := NewSet_(), NewSet_()
	a.Insert(_p("10.0.0.0/25"))
	b.Insert(_p("10.0.0.128/25"))

	a.Insert(b)
	assert.True(t, a.isValid())
	assert.True(t, a.Contains(_p("10.0.0.0/25")))
	assert.True(t, a.Contains(_p("10.0.0.128/25")))
	assert.True(t, a.Contains(_p("10.0.0.0/24")))
}

func TestSetRemoveSet(t *testing.T) {
	a, b := NewSet_(), NewSet_()
	a.Insert(_p("10.0.0.0/24"))
	b.Insert(_p("10.0.0.128/25"))

	a.Remove(b)
	assert.True(t, a.isValid())
	assert.True(t, a.Contains(_p("10.0.0.0/25")))
	assert.False(t, a.Contains(_p("10.0.0.128/25")))
	assert.False(t, a.Contains(_p("10.0.0.0/24")))
}

func TestSetConcurrentModification(t *testing.T) {
	set := NewSet_()

	wg := new(sync.WaitGroup)
	wg.Add(2)

	var panicked int
	wrap := func() {
		if r := recover(); r != nil {
			panicked++
		}
		wg.Done()
	}

	// Simulate two goroutines modifying at the same time using a channel to
	// freeze one in the middle and start the other.
	ch := make(chan bool)
	go func() {
		defer wrap()
		set.mutate(func() (bool, *setNode) {
			ch <- true
			return true, set.s.trie.Union(_a("10.0.0.1").Set().trie)
		})
	}()
	go func() {
		defer wrap()
		set.mutate(func() (bool, *setNode) {
			<-ch
			return true, set.s.trie.Union(_a("10.0.0.2").Set().trie)
		})
	}()
	wg.Wait()
	assert.Equal(t, 1, panicked)
}

func TestNilSet(t *testing.T) {
	var set Set_

	nonEmptySet := _p("203.0.113.0/24").Set().Set_()

	// On-offs
	assert.Equal(t, int64(0), set.NumAddresses())
	assert.Equal(t, int64(0), set.Set().NumAddresses())
	assert.False(t, set.Contains(_p("203.0.113.0/24")))

	// Equal
	assert.True(t, set.Equal(set))
	assert.True(t, set.Equal(NewSet_()))
	assert.True(t, NewSet_().Equal(set))
	assert.False(t, set.Equal(nonEmptySet))
	assert.False(t, nonEmptySet.Equal(set))

	// Union
	assert.True(t, set.Union(nonEmptySet).Equal(nonEmptySet.Set()))
	assert.True(t, nonEmptySet.Union(set).Equal(nonEmptySet.Set()))

	// Intersection
	assert.Equal(t, int64(0), set.Intersection(nonEmptySet).NumAddresses())
	assert.Equal(t, int64(0), nonEmptySet.Intersection(set).NumAddresses())

	// Difference
	assert.Equal(t, int64(0), set.Difference(nonEmptySet).NumAddresses())
	assert.Equal(t, int64(256), nonEmptySet.Difference(set).NumAddresses())

	// Walk
	assert.True(t, set.Set().WalkAddresses(func(Address) bool {
		panic("should not be called")
	}))
	assert.True(t, set.Set().WalkPrefixes(func(Prefix) bool {
		panic("should not be called")
	}))
	assert.True(t, set.Set().WalkRanges(func(Range) bool {
		panic("should not be called")
	}))

	t.Run("insert panics", func(t *testing.T) {
		var panicked bool
		func() {
			defer func() {
				if r := recover(); r != nil {
					panicked = true
				}
			}()
			set.Insert(nonEmptySet)
		}()
		assert.True(t, panicked)
	})
	t.Run("remove panics", func(t *testing.T) {
		var panicked bool
		func() {
			defer func() {
				if r := recover(); r != nil {
					panicked = true
				}
			}()
			set.Remove(nonEmptySet)
		}()
		assert.True(t, panicked)
	})
}

func TestSetContainsNil(t *testing.T) {
	assert.True(t, Set_{}.Contains(nil))
	assert.True(t, Set{}.Contains(nil))
}

func TestSetUnionNil(t *testing.T) {
	assert.Equal(t, int64(0), Set_{}.Union(nil).NumAddresses())
	assert.Equal(t, int64(0), Set{}.Union(nil).NumAddresses())
}

func TestSetIntesectionNil(t *testing.T) {
	assert.Equal(t, int64(0), Set_{}.Intersection(nil).NumAddresses())
	assert.Equal(t, int64(0), Set{}.Intersection(nil).NumAddresses())
}

func TestSetDifferenceNil(t *testing.T) {
	assert.Equal(t, int64(0), Set_{}.Difference(nil).NumAddresses())
	assert.Equal(t, int64(0), Set{}.Difference(nil).NumAddresses())
}

func TestSetInsertNil(t *testing.T) {
	s := NewSet_()
	s.Insert(nil)
	assert.Equal(t, int64(0), s.NumAddresses())
}

func TestSetRemoveNil(t *testing.T) {
	s := NewSet_()
	s.Remove(nil)
	assert.Equal(t, int64(0), s.NumAddresses())
}

func TestFixedSetContainsPrefix(t *testing.T) {
	s := Set{}.Build(func(s_ Set_) bool {
		s_.Insert(_p("10.0.0.0/16"))
		return true
	})
	s = s.Build(func(s_ Set_) bool {
		s_.Insert(_p("10.224.0.0/24"))
		return false
	})
	assert.True(t, s.Contains(_p("10.0.0.0/24")))
	assert.True(t, s.Contains(_p("10.0.30.0/27")))
	assert.True(t, s.Contains(_p("10.0.128.0/17")))
	assert.False(t, s.Contains(_p("10.224.0.0/24")))
	assert.False(t, s.Contains(_p("10.1.30.0/27")))
	assert.False(t, s.Contains(_p("10.0.128.0/15")))
}

func TestWalkRanges(t *testing.T) {
	tests := []struct {
		description string
		prefixes    []SetI
		ranges      []Range
		str         string
	}{
		{
			description: "empty",
			prefixes:    []SetI{},
			ranges:      []Range{},
			str:         "[]",
		}, {
			description: "simple prefix",
			prefixes: []SetI{
				_p("203.0.113.0/24"),
			},
			ranges: []Range{
				_p("203.0.113.0/24").Range(),
			},
			str: "[203.0.113.0/24]",
		}, {
			description: "adjacent prefixes",
			prefixes: []SetI{
				_p("203.0.113.0/25"),
				_p("203.0.113.128/26"),
			},
			ranges: []Range{
				_r(_a("203.0.113.0"), _a("203.0.113.191")),
			},
			str: "[203.0.113.0/25, 203.0.113.128/26]",
		}, {
			description: "disjoint prefixes",
			prefixes: []SetI{
				_p("203.0.113.0/25"),
				_p("203.0.113.192/26"),
			},
			ranges: []Range{
				_r(_a("203.0.113.0"), _a("203.0.113.127")),
				_r(_a("203.0.113.192"), _a("203.0.113.255")),
			},
			str: "[203.0.113.0/25, 203.0.113.192/26]",
		}, {
			// This is the reverse of the complicated test from range_test.go
			description: "complicated",
			prefixes: []SetI{
				_p("7.0.37.17/32"),
				_p("7.0.37.18/31"),
				_p("7.0.37.20/30"),
				_p("7.0.37.24/29"),
				_p("7.0.37.32/27"),
				_p("7.0.37.64/26"),
				_p("7.0.37.128/25"),
				_p("7.0.38.0/23"),
				_p("7.0.40.0/21"),
				_p("7.0.48.0/20"),
				_p("7.0.64.0/18"),
				_p("7.0.128.0/17"),
				_p("7.1.0.0/16"),
				_p("7.2.0.0/15"),
				_p("7.4.0.0/14"),
				_p("7.8.0.0/13"),
				_p("7.16.0.0/12"),
				_p("7.32.0.0/11"),
				_p("7.64.0.0/10"),
				_p("7.128.0.0/9"),
				_p("8.0.0.0/6"),
				_p("12.0.0.0/8"),
				_p("13.0.0.0/13"),
				_p("13.8.0.0/17"),
				_p("13.8.128.0/18"),
				_p("13.8.192.0/20"),
				_p("13.8.208.0/21"),
				_p("13.8.216.0/22"),
				_p("13.8.220.0/23"),
				_p("13.8.222.0/26"),
				_p("13.8.222.64/27"),
				_p("13.8.222.96/28"),
				_p("13.8.222.112/31"),
			},
			ranges: []Range{
				_r(_a("7.0.37.17"), _a("13.8.222.113")),
			},
			str: "[7.0.37.17/32, 7.0.37.18/31, 7.0.37.20/30, 7.0.37.24/29, 7.0.37.32/27, 7.0.37.64/26, 7.0.37.128/25, 7.0.38.0/23, 7.0.40.0/21, 7.0.48.0/20, 7.0.64.0/18, 7.0.128.0/17, 7.1.0.0/16, 7.2.0.0/15, 7.4.0.0/14, 7.8.0.0/13, 7.16.0.0/12, 7.32.0.0/11, 7.64.0.0/10, 7.128.0.0/9, 8.0.0.0/6, 12.0.0.0/8, 13.0.0.0/13, 13.8.0.0/17, 13.8.128.0/18, 13.8.192.0/20, 13.8.208.0/21, 13.8.216.0/22, 13.8.220.0/23, 13.8.222.0/26, 13.8.222.64/27, 13.8.222.96/28, 13.8.222.112/31]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			t.Run("finish", func(t *testing.T) {
				s := func() Set {
					s := NewSet_()
					for _, p := range tt.prefixes {
						s.Insert(p)
					}
					return s.Set()
				}()
				ranges := []Range{}
				finished := s.WalkRanges(func(r Range) bool {
					ranges = append(ranges, r)
					return true
				})
				assert.True(t, finished)
				require.Equal(t, len(tt.ranges), len(ranges))
				for i, r := range tt.ranges {
					assert.Equal(t, r, ranges[i])
				}
				assert.Equal(t, tt.str, s.String())
			})

			t.Run("don't finish", func(t *testing.T) {
				s := func() Set {
					s := NewSet_()
					for _, p := range tt.prefixes {
						s.Insert(p)
					}
					return s.Set()
				}()
				if s.NumAddresses() != 0 {
					ranges := []Range{}
					finished := s.WalkRanges(func(r Range) bool {
						ranges = append(ranges, r)
						return false
					})
					assert.False(t, finished)
					assert.Equal(t, 1, len(ranges))
					assert.Equal(t, tt.ranges[0], ranges[0])
				}
			})
		})
	}
}

func TestFixedSetContainsSet(t *testing.T) {
	s := NewSet_()
	s.Insert(_p("10.0.0.0/16"))

	other := NewSet_()
	other.Insert(_p("10.0.0.0/24"))
	other.Insert(_p("10.0.30.0/27"))
	other.Insert(_p("10.0.128.0/17"))

	assert.True(t, s.Contains(other))

	other.Insert(_p("10.224.0.0/24"))
	other.Insert(_p("10.1.30.0/27"))
	other.Insert(_p("10.0.128.0/15"))

	assert.False(t, s.Contains(other))

}

var (
	Eights = _a("8.8.8.8")
	Nines  = _a("9.9.9.9")

	Ten24          = _p("10.0.0.0/24")
	TenOne24       = _p("10.0.1.0/24")
	TenTwo24       = _p("10.0.2.0/24")
	Ten24128       = _p("10.0.0.128/25")
	Ten24Router    = _a("10.0.0.1")
	Ten24Broadcast = _a("10.0.0.255")
)

func TestOldSetInit(t *testing.T) {
	set := Set{}

	assert.Equal(t, int64(0), set.NumAddresses())
	assert.True(t, set.isValid())
}

func TestOldSetContains(t *testing.T) {
	set := Set{}

	assert.Equal(t, int64(0), set.NumAddresses())
	assert.False(t, set.Contains(Eights))
	assert.False(t, set.Contains(Nines))
	assert.True(t, set.isValid())
}

func TestOldSetInsert(t *testing.T) {
	s := NewSet_()

	s.Insert(Nines)
	assert.Equal(t, int64(1), s.NumAddresses())
	assert.True(t, s.Contains(Nines))
	assert.False(t, s.Contains(Eights))
	s.Insert(Eights)
	assert.Equal(t, int64(2), s.s.trie.NumNodes())
	assert.True(t, s.Contains(Eights))
	assert.True(t, s.isValid())
}

func TestOldSetInsertPrefixwork(t *testing.T) {
	s := NewSet_()

	s.Insert(Ten24)
	assert.Equal(t, int64(1), s.s.trie.NumNodes())
	assert.Equal(t, int64(256), s.NumAddresses())
	assert.True(t, s.Contains(Ten24))
	assert.True(t, s.Contains(Ten24128))
	assert.False(t, s.Contains(Nines))
	assert.False(t, s.Contains(Eights))
	assert.True(t, s.isValid())
}

func TestOldSetInsertSequential(t *testing.T) {
	s := NewSet_()

	s.Insert(_a("192.168.1.0"))
	assert.Equal(t, int64(1), s.s.trie.NumNodes())
	s.Insert(_a("192.168.1.1"))
	assert.Equal(t, int64(1), s.s.trie.NumNodes())
	s.Insert(_a("192.168.1.2"))
	assert.Equal(t, int64(2), s.s.trie.NumNodes())
	s.Insert(_a("192.168.1.3"))
	assert.Equal(t, int64(1), s.s.trie.NumNodes())
	assert.Equal(t, int64(4), s.NumAddresses())

	cidr := _p("192.168.1.0/30")
	assert.True(t, s.Contains(cidr))

	cidr = _p("192.168.1.4/31")
	s.Insert(cidr)
	assert.Equal(t, int64(2), s.s.trie.NumNodes())
	assert.True(t, s.Contains(cidr))

	cidr = _p("192.168.1.6/31")
	s.Insert(cidr)
	assert.Equal(t, int64(1), s.s.trie.NumNodes())
	assert.True(t, s.Contains(cidr))

	cidr = _p("192.168.1.6/31")
	s.Insert(cidr)
	assert.Equal(t, int64(1), s.s.trie.NumNodes())
	assert.True(t, s.Contains(cidr))

	cidr = _p("192.168.0.240/29")
	s.Insert(cidr)
	assert.Equal(t, int64(2), s.s.trie.NumNodes())
	assert.True(t, s.Contains(cidr))

	cidr = _p("192.168.0.248/29")
	s.Insert(cidr)
	assert.Equal(t, int64(2), s.s.trie.NumNodes())
	assert.True(t, s.Contains(cidr))
	assert.True(t, s.isValid())
}

func TestOldSetRemove(t *testing.T) {
	s := NewSet_()

	s.Insert(Ten24)
	assert.Equal(t, int64(1), s.s.trie.NumNodes())
	s.Remove(Ten24128)
	assert.Equal(t, int64(1), s.s.trie.NumNodes())
	assert.Equal(t, int64(128), s.NumAddresses())
	assert.False(t, s.Contains(Ten24))
	assert.False(t, s.Contains(Ten24128))
	cidr := _p("10.0.0.0/25")
	assert.True(t, s.Contains(cidr))

	s.Remove(Ten24Router)
	assert.Equal(t, int64(127), s.NumAddresses())
	assert.Equal(t, int64(7), s.s.trie.NumNodes())
	assert.True(t, s.isValid())
}

func TestOldSetRemovePrefixworkBroadcast(t *testing.T) {
	s := NewSet_()

	s.Insert(Ten24)
	assert.Equal(t, int64(1), s.s.trie.NumNodes())
	s.Remove(Ten24.Address())
	s.Remove(Ten24Broadcast)
	assert.Equal(t, int64(254), s.NumAddresses())
	assert.Equal(t, int64(14), s.s.trie.NumNodes())
	assert.False(t, s.Contains(Ten24))
	assert.False(t, s.Contains(Ten24128))
	assert.False(t, s.Contains(Ten24Broadcast))
	assert.False(t, s.Contains(Ten24.Address()))

	cidr := _p("10.0.0.128/26")
	assert.True(t, s.Contains(cidr))
	assert.True(t, s.Contains(Ten24Router))

	s.Remove(Ten24Router)
	assert.Equal(t, int64(253), s.NumAddresses())
	assert.Equal(t, int64(13), s.s.trie.NumNodes())
	assert.True(t, s.isValid())
}

func TestOldSetRemoveAll(t *testing.T) {
	s := NewSet_()

	s.Insert(Ten24)
	cidr1 := _p("192.168.0.0/25")
	s.Insert(cidr1)
	assert.Equal(t, int64(2), s.s.trie.NumNodes())

	cidr2 := _p("0.0.0.0/0")
	s.Remove(cidr2)
	assert.Equal(t, int64(0), s.s.trie.NumNodes())
	assert.False(t, s.Contains(Ten24))
	assert.False(t, s.Contains(Ten24128))
	assert.False(t, s.Contains(cidr1))
	assert.True(t, s.isValid())
}

func TestOldSet_RemoveTop(t *testing.T) {
	testSet := NewSet_()
	ip1 := _a("10.0.0.1")
	ip2 := _a("10.0.0.2")

	testSet.Insert(ip2) // top
	testSet.Insert(ip1) // inserted at left
	testSet.Remove(ip2) // remove top node

	assert.True(t, testSet.Contains(ip1))
	assert.False(t, testSet.Contains(ip2))
	assert.True(t, testSet.isValid())
}

func TestOldSetInsertOverlapping(t *testing.T) {
	s := NewSet_()

	s.Insert(Ten24128)
	assert.False(t, s.Contains(Ten24))
	assert.Equal(t, int64(1), s.s.trie.NumNodes())
	s.Insert(Ten24)
	assert.Equal(t, int64(1), s.s.trie.NumNodes())
	assert.Equal(t, int64(256), s.NumAddresses())
	assert.True(t, s.Contains(Ten24))
	assert.True(t, s.Contains(Ten24Router))
	assert.False(t, s.Contains(Eights))
	assert.False(t, s.Contains(Nines))
	assert.True(t, s.isValid())
}

func TestOldSetUnion(t *testing.T) {
	set1 := NewSet_()
	set2 := NewSet_()

	set1.Insert(Ten24)
	cidr := _p("192.168.0.248/29")
	set2.Insert(cidr)

	set := set1.Union(set2)
	assert.True(t, set.Contains(Ten24))
	assert.True(t, set.Contains(cidr))
	assert.True(t, set1.isValid())
	assert.True(t, set2.isValid())
}

func TestOldSetDifference(t *testing.T) {
	set1 := NewSet_()
	set2 := NewSet_()

	set1.Insert(Ten24)
	cidr := _p("192.168.0.248/29")
	set2.Insert(cidr)

	set := set1.Difference(set2)
	assert.True(t, set.Contains(Ten24))
	assert.False(t, set.Contains(cidr))
	assert.True(t, set1.isValid())
	assert.True(t, set2.isValid())
}

func TestOldIntersectionAinB1(t *testing.T) {
	case1 := []string{"10.0.16.0/20", "10.5.8.0/24", "10.23.224.0/23"}
	case2 := []string{"10.0.20.0/30", "10.5.8.0/29", "10.23.224.0/27"}
	output := []string{"10.23.224.0/27", "10.0.20.0/30", "10.5.8.0/29"}
	testIntersection(t, case1, case2, output)

}

func TestOldIntersectionAinB2(t *testing.T) {
	case1 := []string{"10.10.0.0/30", "10.5.8.0/29", "10.23.224.0/27"}
	case2 := []string{"10.10.0.0/20", "10.5.8.0/24", "10.23.224.0/23"}
	output := []string{"10.10.0.0/30", "10.5.8.0/29", "10.23.224.0/27"}
	testIntersection(t, case1, case2, output)
}

func TestOldIntersectionAinB3(t *testing.T) {
	case1 := []string{"10.0.5.0/24", "10.5.8.0/29", "10.23.224.0/27"}
	case2 := []string{"10.6.0.0/24", "10.9.9.0/29", "10.23.6.0/23"}
	output := []string{}
	testIntersection(t, case1, case2, output)
}

func TestOldIntersectionAinB4(t *testing.T) {
	case1 := []string{"10.23.6.0/24", "10.5.8.0/29", "10.23.224.0/27"}
	case2 := []string{"10.6.0.0/24", "10.9.9.0/29", "10.23.6.0/29"}
	output := []string{"10.23.6.0/29"}
	testIntersection(t, case1, case2, output)
}

func TestOldIntersectionAinB5(t *testing.T) {
	case1 := []string{"10.0.23.0/27", "10.0.20.0/27", "10.0.15.0/27"}
	case2 := []string{"10.0.23.0/24", "10.0.20.0/24", "10.0.15.0/24"}
	output := []string{"10.0.23.0/27", "10.0.20.0/27", "10.0.15.0/27"}
	testIntersection(t, case1, case2, output)
}

func TestOldIntersectionAinB6(t *testing.T) {
	case1 := []string{"10.0.23.0/24", "10.0.20.0/24", "10.0.15.0/24"}
	case2 := []string{"10.0.23.0/27", "10.0.20.0/27", "10.0.15.0/27"}
	output := []string{"10.0.15.0/27", "10.0.20.0/27", "10.0.23.0/27"}
	testIntersection(t, case1, case2, output)
}

func TestOldIntersectionAinB7(t *testing.T) {
	case1 := []string{"10.0.23.0/24", "10.0.20.0/24", "10.0.15.0/24"}
	case2 := []string{"10.0.14.0/27", "10.0.10.0/27", "10.0.8.0/27"}
	output := []string{}
	testIntersection(t, case1, case2, output)
}

func TestOldIntersectionAinB8(t *testing.T) {
	case1 := []string{"10.0.23.0/24", "10.0.20.0/24", "172.16.1.0/24"}
	case2 := []string{"10.0.14.0/27", "10.0.10.0/27", "172.16.1.0/28"}
	output := []string{"172.16.1.0/28"}
	testIntersection(t, case1, case2, output)
}

func TestOldIntersectionAinB9(t *testing.T) {
	case1 := []string{"10.5.8.0/29"}
	case2 := []string{"10.10.0.0/20", "10.5.8.0/24", "10.23.224.0/23"}
	output := []string{"10.5.8.0/29"}
	testIntersection(t, case1, case2, output)
}

func testIntersection(t *testing.T, input1 []string, input2 []string, output []string) {
	set1 := NewSet_()
	set2 := NewSet_()
	interSect := NewSet_()
	for i := 0; i < len(input1); i++ {
		cidr := _p(input1[i])
		set1.Insert(cidr)
	}
	for j := 0; j < len(input2); j++ {
		cidr := _p(input2[j])
		set2.Insert(cidr)
	}
	for k := 0; k < len(output); k++ {
		cidr := _p(output[k])
		interSect.Insert(cidr)
	}
	set := set1.Intersection(set2)
	assert.True(t, interSect.Set().Equal(set))
	assert.True(t, set1.isValid())
	assert.True(t, set2.isValid())
	assert.True(t, set.isValid())

}

func TestOldSetAllocateDeallocate(t *testing.T) {
	rand.Seed(29)

	s := NewSet_()

	bigNet := _p("15.1.0.0/16")
	s.Insert(bigNet)

	assert.Equal(t, int64(65536), s.NumAddresses())

	ips := make([]Address, 0, s.NumAddresses())
	s.Set().WalkAddresses(func(ip Address) bool {
		ips = append(ips, ip)
		return true
	})

	allocated := NewSet_()
	for i := 0; i != 16384; i++ {
		allocated.Insert(ips[rand.Intn(65536)])
	}
	assert.Equal(t, int64(14500), allocated.NumAddresses())
	allocated.Set().WalkAddresses(func(ip Address) bool {
		assert.True(t, s.Contains(ip))
		return true
	})

	available := s.Difference(allocated)
	assert.Equal(t, int64(51036), available.NumAddresses())
	available.Set().WalkAddresses(func(ip Address) bool {
		assert.True(t, s.Contains(ip))
		assert.False(t, allocated.Contains(ip))
		return true
	})
	assert.Equal(t, int64(51036), available.NumAddresses())
	assert.True(t, s.isValid())
	assert.True(t, allocated.isValid())
	assert.True(t, available.isValid())
}

func TestOldEqualTrivial(t *testing.T) {
	a := NewSet_()
	b := NewSet_()
	assert.True(t, a.Equal(b))

	a.Insert(_p("10.0.0.0/24"))
	assert.False(t, a.Equal(b))
	assert.False(t, b.Equal(a))
	assert.True(t, a.Equal(a))
	assert.True(t, b.Equal(b))
	assert.True(t, a.isValid())
	assert.True(t, b.isValid())
}

func TestOldEqualSingleNode(t *testing.T) {
	a := NewSet_()
	b := NewSet_()
	a.Insert(_p("10.0.0.0/24"))
	b.Insert(_p("10.0.0.0/24"))

	assert.True(t, a.Equal(b))
	assert.True(t, b.Equal(a))
	assert.True(t, a.isValid())
	assert.True(t, b.isValid())
}

func TestOldEqualAllIPv4(t *testing.T) {
	a := NewSet_()
	b := NewSet_()
	c := NewSet_()
	// Insert the entire IPv4 space into set a in one shot
	a.Insert(_p("0.0.0.0/0"))

	// Insert the entire IPv4 space piece by piece into b and c

	// This list was constructed starting with 0.0.0.0/32 and 0.0.0.1/32,
	// then adding 0.0.0.2/31, 0.0.0.4/30, ..., 128.0.0./1, and then
	// randomizing the list.
	bNets := []Prefix{
		_p("0.0.0.0/32"),
		_p("0.0.0.1/32"),
		_p("0.0.0.128/25"),
		_p("0.0.0.16/28"),
		_p("0.0.0.2/31"),
		_p("0.0.0.32/27"),
		_p("0.0.0.4/30"),
		_p("0.0.0.64/26"),
		_p("0.0.0.8/29"),
		_p("0.0.1.0/24"),
		_p("0.0.128.0/17"),
		_p("0.0.16.0/20"),
		_p("0.0.2.0/23"),
		_p("0.0.32.0/19"),
		_p("0.0.4.0/22"),
		_p("0.0.64.0/18"),
		_p("0.0.8.0/21"),
		_p("0.1.0.0/16"),
		_p("0.128.0.0/9"),
		_p("0.16.0.0/12"),
		_p("0.2.0.0/15"),
		_p("0.32.0.0/11"),
		_p("0.4.0.0/14"),
		_p("0.64.0.0/10"),
		_p("0.8.0.0/13"),
		_p("1.0.0.0/8"),
		_p("128.0.0.0/1"),
		_p("16.0.0.0/4"),
		_p("2.0.0.0/7"),
		_p("32.0.0.0/3"),
		_p("4.0.0.0/6"),
		_p("64.0.0.0/2"),
		_p("8.0.0.0/5"),
	}

	for _, n := range bNets {
		assert.False(t, a.Equal(b))
		assert.False(t, b.Equal(a))
		b.Insert(n)
		assert.False(t, b.Equal(c))
		assert.False(t, c.Equal(b))
	}

	// Constructed a different way
	cNets := []Prefix{
		_p("255.255.255.240/29"),
		_p("0.0.0.0/1"),
		_p("255.255.128.0/18"),
		_p("255.255.240.0/21"),
		_p("254.0.0.0/8"),
		_p("255.240.0.0/13"),
		_p("255.224.0.0/12"),
		_p("248.0.0.0/6"),
		_p("255.0.0.0/9"),
		_p("255.252.0.0/15"),
		_p("255.255.224.0/20"),
		_p("255.255.255.224/28"),
		_p("255.255.255.0/25"),
		_p("252.0.0.0/7"),
		_p("192.0.0.0/3"),
		_p("255.192.0.0/11"),
		_p("255.255.255.248/30"),
		_p("255.255.252.0/23"),
		_p("255.248.0.0/14"),
		_p("255.255.255.192/27"),
		_p("255.255.0.0/17"),
		_p("255.254.0.0/16"),
		_p("255.255.255.255/32"),
		_p("128.0.0.0/2"),
		_p("255.128.0.0/10"),
		_p("255.255.255.128/26"),
		_p("240.0.0.0/5"),
		_p("255.255.255.252/31"),
		_p("255.255.192.0/19"),
		_p("255.255.254.0/24"),
		_p("255.255.248.0/22"),
		_p("224.0.0.0/4"),
		_p("255.255.255.254/32"),
	}

	for _, n := range cNets {
		assert.False(t, c.Equal(a))
		assert.False(t, c.Equal(b))
		c.Insert(n)
		assert.True(t, a.isValid())
		assert.True(t, b.isValid())
		assert.True(t, c.isValid())
	}

	// At this point, all three should have the entire IPv4 space
	assert.True(t, a.Equal(b))
	assert.True(t, a.Equal(c))
	assert.True(t, b.Equal(a))
	assert.True(t, b.Equal(c))
	assert.True(t, c.Equal(a))
	assert.True(t, c.Equal(b))
}

func TestSetNumPrefixes(t *testing.T) {
	tests := []struct {
		description string
		prefixes    []SetI
		length      uint32
		count       uint64
		error       bool
	}{
		{
			description: "empty set",
			prefixes:    []SetI{},
			length:      32,
			count:       0,
		}, {
			description: "single prefix",
			prefixes:    []SetI{_p("203.0.113.0/8")},
			length:      16,
			count:       256,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			s := Set{}.Build(func(s Set_) bool {
				for _, prefix := range tt.prefixes {
					s.Insert(prefix)
				}
				return true
			})

			count, err := s.NumPrefixes(tt.length)
			assert.Equal(t, tt.error, err != nil)
			if !tt.error {
				assert.Equal(t, tt.count, count)
			}
		})
	}
}

func TestSetNumPrefixesStairs(t *testing.T) {
	tests := []struct {
		description string
		length      uint32
		count       uint64
	}{
		{length: 8, count: 0x00001},
		{length: 9, count: 0x00003},
		{length: 10, count: 0x00007},
		{length: 11, count: 0x0000f},
		{length: 12, count: 0x0001f},
		{length: 13, count: 0x0003f},
		{length: 14, count: 0x0007f},
		{length: 15, count: 0x000ff},
		{length: 16, count: 0x001ff},
		{length: 17, count: 0x003ff},
		{length: 18, count: 0x007ff},
		{length: 19, count: 0x00fff},
		{length: 20, count: 0x01fff},
		{length: 21, count: 0x03fff},
		{length: 22, count: 0x07fff},
		{length: 23, count: 0x0fffe},
		{length: 24, count: 0x1fffc},
	}

	s := Set{}.Build(func(s Set_) bool {
		s.Insert(_p("10.0.0.0/8"))
		s.Insert(_p("11.0.0.0/9"))
		s.Insert(_p("11.128.0.0/10"))
		s.Insert(_p("11.192.0.0/11"))
		s.Insert(_p("11.224.0.0/12"))
		s.Insert(_p("11.240.0.0/13"))
		s.Insert(_p("11.248.0.0/14"))
		s.Insert(_p("11.252.0.0/15"))
		s.Insert(_p("11.254.0.0/16"))
		s.Insert(_p("11.255.0.0/17"))
		s.Insert(_p("11.255.128.0/18"))
		s.Insert(_p("11.255.192.0/19"))
		s.Insert(_p("11.255.224.0/20"))
		s.Insert(_p("11.255.240.0/21"))
		s.Insert(_p("11.255.248.0/22"))
		return true
	})

	for _, tt := range tests {
		t.Run(strconv.FormatUint(uint64(tt.length), 10), func(t *testing.T) {
			count, err := s.NumPrefixes(tt.length)
			assert.Nil(t, err)
			assert.Equal(t, tt.count, count)
		})
	}
}

func TestFindAvailablePrefix(t *testing.T) {
	tests := []struct {
		description string
		space       []SetI
		reserved    []SetI
		length      uint32
		expected    Prefix
		err         bool
		change      int
	}{
		{
			description: "empty",
			space: []SetI{
				_p("10.0.0.0/8"),
			},
			length: 24,
			change: 1,
		}, {
			description: "find adjacent",
			space: []SetI{
				_p("10.0.0.0/8"),
			},
			reserved: []SetI{
				_p("10.224.123.0/24"),
			},
			length:   24,
			expected: _p("10.224.122.0/24"),
		}, {
			description: "many fewer prefixes",
			space: []SetI{
				_p("10.0.0.0/16"),
			},
			reserved: []SetI{
				_p("10.0.1.0/24"),
				_p("10.0.2.0/23"),
				_p("10.0.4.0/22"),
				_p("10.0.8.0/21"),
				_p("10.0.16.0/20"),
				_p("10.0.32.0/19"),
				_p("10.0.64.0/18"),
				_p("10.0.128.0/17"),
			},
			length: 24,
			change: -7,
		}, {
			description: "toobig",
			space: []SetI{
				_p("10.0.0.0/8"),
			},
			reserved: []SetI{
				_p("10.128.0.0/9"),
				_p("10.64.0.0/10"),
				_p("10.32.0.0/11"),
				_p("10.16.0.0/12"),
			},
			length: 11,
			err:    true,
		}, {
			description: "full",
			space: []SetI{
				_p("10.0.0.0/8"),
			},
			length: 7,
			err:    true,
		}, {
			description: "random disjoint example",
			space: []SetI{
				_p("10.0.0.0/22"),
				_p("192.168.0.0/21"),
				_p("172.16.0.0/20"),
			},
			reserved: []SetI{
				_p("192.168.0.0/21"),
				_p("172.16.0.0/21"),
				_p("172.16.8.0/22"),
				_p("10.0.0.0/22"),
				_p("172.16.12.0/24"),
				_p("172.16.14.0/24"),
				_p("172.16.15.0/24"),
			},
			length:   24,
			expected: _p("172.16.13.0/24"),
			change:   1,
		}, {
			description: "too fragmented",
			space: []SetI{
				_p("10.0.0.0/24"),
			},
			reserved: []SetI{
				_p("10.0.0.0/27"),
				_p("10.0.0.64/27"),
				_p("10.0.0.128/27"),
				_p("10.0.0.192/27"),
			},
			length: 25,
			err:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			// This is the full usable IP space
			space := Set{}.Build(func(s_ Set_) bool {
				for _, p := range tt.space {
					s_.Insert(p)
				}
				return true
			})
			// This is the part of the usable space which has already been allocated
			reserved := Set{}.Build(func(s_ Set_) bool {
				for _, p := range tt.reserved {
					s_.Insert(p)
				}
				return true
			})

			// Call the method under test to find the best allocation to avoid fragmentation.
			prefix, err := space.FindAvailablePrefix(reserved, tt.length)

			assert.Equal(t, tt.err, err != nil)
			if err != nil {
				return
			}

			assert.Equal(t, int64(0), reserved.Intersection(prefix).NumAddresses())

			// Not all test cases care which prefix is returned but in some
			// cases, there is only one right answer and so we might check it.
			// This isn't strictly necessary but was handy with the first few.
			if tt.expected.length != 0 {
				assert.Equal(t, tt.expected.String(), prefix.String())
			}

			// What really matters is that fragmentation in the IP space is
			// always avoided as much as possible. The `change` field in each
			// test indicates what should happen to IP space fragmentation.
			// This test framework measures fragmentation as the change in the
			// minimal number of prefixes required to span the reserved set.
			before := countPrefixes(reserved)
			after := countPrefixes(reserved.Build(func(s_ Set_) bool {
				s_.Insert(prefix)
				return true
			}))

			diff := after - before
			assert.LessOrEqual(t, diff, 1)
			assert.LessOrEqual(t, diff, tt.change)
		})
	}

	t.Run("randomized", func(t *testing.T) {
		// Start with a space and an empty reserved set.
		// This test will attempt to fragment the space by pulling out
		space := _p("10.128.0.0/12").Set()
		available := space.NumAddresses()

		reserved := NewSet_()

		rand.Seed(29)
		for available > 0 {
			// This is the most we can pull out. Assuming we avoid
			// fragmentation, it should be the largest power of two that is
			// less than or equal to the number of available addresses.
			maxExponent := log2(available)

			// Finding the maximum prefix here, proves we are avoiding fragmentation
			maxPrefix, err := space.FindAvailablePrefix(reserved, 32-maxExponent)
			assert.Nil(t, err)
			assert.Equal(t, pow2(maxExponent), maxPrefix.NumAddresses())
			assert.Equal(t, int64(0), reserved.Intersection(maxPrefix).NumAddresses())

			// Pull out a random sized prefix up to the maximum size to attempt to further fragment the space.
			randomSize := (rand.Uint32()%maxExponent + 1)
			if randomSize > 12 {
				randomSize = 12
			}

			randomSizePrefix, err := space.FindAvailablePrefix(reserved, 32-randomSize)
			assert.Nil(t, err)
			assert.Equal(t, pow2(randomSize), randomSizePrefix.NumAddresses())
			assert.Equal(t, int64(0), reserved.Intersection(randomSizePrefix).NumAddresses())

			// Reserve only the random sized one
			reserved.Insert(randomSizePrefix)
			available -= randomSizePrefix.NumAddresses()
			assert.Equal(t, available, space.NumAddresses()-reserved.NumAddresses())
		}
	})
}

func pow2(x uint32) int64 {
	return int64(math.Pow(2, float64(x)))
}

func log2(available_addresses int64) uint32 {
	return uint32(math.Log2(float64(available_addresses)))
}

func countPrefixes(s Set) int {
	var numPrefixes int
	s.WalkPrefixes(func(_ Prefix) bool {
		numPrefixes += 1
		return true
	})
	return numPrefixes
}
