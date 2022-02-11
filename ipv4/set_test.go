package ipv4

import (
	"math/rand"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetInsertPrefix(t *testing.T) {
	s := NewSet()
	s.Insert(_p("10.0.0.0/24"))

	assert.True(t, s.Contains(_p("10.0.0.0/24")))
	assert.False(t, s.Contains(_p("10.0.0.0/16")))

	s.Insert(_p("10.0.0.0/16"))

	assert.True(t, s.Contains(_p("10.0.0.0/24")))
	assert.True(t, s.Contains(_p("10.0.0.0/16")))
}

func TestSetRemovePrefix(t *testing.T) {
	s := FixedSet{}.Set()
	s.Insert(_p("10.0.0.0/16"))

	assert.True(t, s.Contains(_p("10.0.0.0/24")))
	assert.True(t, s.Contains(_p("10.0.0.0/16")))

	s.Remove(_p("10.0.0.0/24"))

	assert.False(t, s.Contains(_p("10.0.0.0/24")))
	assert.False(t, s.Contains(_p("10.0.0.0/16")))
	assert.True(t, s.Contains(_p("10.0.1.0/24")))
}

func TestSetAsReferenceType(t *testing.T) {
	s := NewSet()

	func(s Set) {
		s.Insert(_p("10.0.0.0/24"))
	}(s)

	assert.True(t, s.Contains(_p("10.0.0.0/24")))
	assert.False(t, s.Contains(_p("10.0.0.0/16")))

	func(s Set) {
		s.Insert(_p("10.0.0.0/16"))
	}(s)

	assert.True(t, s.Contains(_p("10.0.0.0/24")))
	assert.True(t, s.Contains(_p("10.0.0.0/16")))
}

func TestSetInsertSet(t *testing.T) {
	a, b := NewSet(), NewSet()
	a.Insert(_p("10.0.0.0/25"))
	b.Insert(_p("10.0.0.128/25"))

	a.Insert(b)
	assert.True(t, a.isValid())
	assert.True(t, a.Contains(_p("10.0.0.0/25")))
	assert.True(t, a.Contains(_p("10.0.0.128/25")))
	assert.True(t, a.Contains(_p("10.0.0.0/24")))
}

func TestSetRemoveSet(t *testing.T) {
	a, b := NewSet(), NewSet()
	a.Insert(_p("10.0.0.0/24"))
	b.Insert(_p("10.0.0.128/25"))

	a.Remove(b)
	assert.True(t, a.isValid())
	assert.True(t, a.Contains(_p("10.0.0.0/25")))
	assert.False(t, a.Contains(_p("10.0.0.128/25")))
	assert.False(t, a.Contains(_p("10.0.0.0/24")))
}

func TestSetConcurrentModification(t *testing.T) {
	set := NewSet()

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
			return true, set.s.trie.Union(_a("10.0.0.1").FixedSet().trie)
		})
	}()
	go func() {
		defer wrap()
		set.mutate(func() (bool, *setNode) {
			<-ch
			return true, set.s.trie.Union(_a("10.0.0.2").FixedSet().trie)
		})
	}()
	wg.Wait()
	assert.Equal(t, 1, panicked)
}

func TestNilSet(t *testing.T) {
	var set Set

	nonEmptySet := _p("203.0.113.0/24").FixedSet().Set()

	// On-offs
	assert.Equal(t, int64(0), set.Size())
	assert.Equal(t, int64(0), set.FixedSet().Size())
	assert.False(t, set.Contains(_p("203.0.113.0/24")))

	// Equal
	assert.True(t, set.Equal(set))
	assert.True(t, set.Equal(NewSet()))
	assert.True(t, NewSet().Equal(set))
	assert.False(t, set.Equal(nonEmptySet))
	assert.False(t, nonEmptySet.Equal(set))

	// EqualInterface
	assert.True(t, set.EqualInterface(set))
	assert.True(t, set.EqualInterface(NewSet()))
	assert.True(t, NewSet().EqualInterface(set))
	assert.False(t, set.EqualInterface(_p("203.0.113.0/24")))
	assert.False(t, _p("203.0.113.0/24").FixedSet().Set().EqualInterface(set))

	// Union
	assert.True(t, set.Union(nonEmptySet).Equal(nonEmptySet))
	assert.True(t, nonEmptySet.Union(set).Equal(nonEmptySet))

	// Intersection
	assert.Equal(t, int64(0), set.Intersection(nonEmptySet).Size())
	assert.Equal(t, int64(0), nonEmptySet.Intersection(set).Size())

	// Difference
	assert.Equal(t, int64(0), set.Difference(nonEmptySet).Size())
	assert.Equal(t, int64(256), nonEmptySet.Difference(set).Size())

	// Walk
	assert.True(t, set.FixedSet().WalkAddresses(func(Address) bool {
		panic("should not be called")
	}))
	assert.True(t, set.FixedSet().WalkPrefixes(func(Prefix) bool {
		panic("should not be called")
	}))
	assert.True(t, set.FixedSet().WalkRanges(func(Range) bool {
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

func TestSetEqualNil(t *testing.T) {
	assert.True(t, Set{}.Equal(nil))
	assert.True(t, FixedSet{}.Equal(nil))
}

func TestSetContainsNil(t *testing.T) {
	assert.True(t, Set{}.Contains(nil))
	assert.True(t, FixedSet{}.Contains(nil))
}

func TestSetUnionNil(t *testing.T) {
	assert.Equal(t, int64(0), Set{}.Union(nil).Size())
	assert.Equal(t, int64(0), FixedSet{}.Union(nil).Size())
}

func TestSetIntesectionNil(t *testing.T) {
	assert.Equal(t, int64(0), Set{}.Intersection(nil).Size())
	assert.Equal(t, int64(0), FixedSet{}.Intersection(nil).Size())
}

func TestSetDifferenceNil(t *testing.T) {
	assert.Equal(t, int64(0), Set{}.Difference(nil).Size())
	assert.Equal(t, int64(0), FixedSet{}.Difference(nil).Size())
}

func TestSetInsertNil(t *testing.T) {
	s := NewSet()
	s.Insert(nil)
	assert.Equal(t, int64(0), s.Size())
}

func TestSetRemoveNil(t *testing.T) {
	s := NewSet()
	s.Remove(nil)
	assert.Equal(t, int64(0), s.Size())
}

func TestFixedSetContainsPrefix(t *testing.T) {
	s := NewSet()
	s.Insert(_p("10.0.0.0/16"))
	assert.True(t, s.FixedSet().Contains(_p("10.0.0.0/24")))
	assert.True(t, s.FixedSet().Contains(_p("10.0.30.0/27")))
	assert.True(t, s.FixedSet().Contains(_p("10.0.128.0/17")))
	assert.False(t, s.FixedSet().Contains(_p("10.224.0.0/24")))
	assert.False(t, s.FixedSet().Contains(_p("10.1.30.0/27")))
	assert.False(t, s.FixedSet().Contains(_p("10.0.128.0/15")))
}

func TestWalkRanges(t *testing.T) {
	tests := []struct {
		description string
		prefixes    []SetI
		ranges      []Range
	}{
		{
			description: "empty",
			prefixes:    []SetI{},
			ranges:      []Range{},
		}, {
			description: "simple prefix",
			prefixes: []SetI{
				_p("203.0.113.0/24"),
			},
			ranges: []Range{
				_p("203.0.113.0/24").Range(),
			},
		}, {
			description: "adjacent prefixes",
			prefixes: []SetI{
				_p("203.0.113.0/25"),
				_p("203.0.113.128/26"),
			},
			ranges: []Range{
				_r(_a("203.0.113.0"), _a("203.0.113.191")),
			},
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
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			t.Run("finish", func(t *testing.T) {
				s := func() FixedSet {
					s := NewSet()
					for _, p := range tt.prefixes {
						s.Insert(p)
					}
					return s.FixedSet()
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
			})

			t.Run("don't finish", func(t *testing.T) {
				s := func() FixedSet {
					s := NewSet()
					for _, p := range tt.prefixes {
						s.Insert(p)
					}
					return s.FixedSet()
				}()
				if s.Size() != 0 {
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
	s := NewSet()
	s.Insert(_p("10.0.0.0/16"))

	other := NewSet()
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
	set := FixedSet{}

	assert.Equal(t, int64(0), set.Size())
	assert.True(t, set.isValid())
}

func TestOldSetContains(t *testing.T) {
	set := FixedSet{}

	assert.Equal(t, int64(0), set.Size())
	assert.False(t, set.Contains(Eights))
	assert.False(t, set.Contains(Nines))
	assert.True(t, set.isValid())
}

func TestOldSetInsert(t *testing.T) {
	s := NewSet()

	s.Insert(Nines)
	assert.Equal(t, int64(1), s.Size())
	assert.True(t, s.Contains(Nines))
	assert.False(t, s.Contains(Eights))
	s.Insert(Eights)
	assert.Equal(t, int64(2), s.s.trie.NumNodes())
	assert.True(t, s.Contains(Eights))
	assert.True(t, s.isValid())
}

func TestOldSetInsertPrefixwork(t *testing.T) {
	s := NewSet()

	s.Insert(Ten24)
	assert.Equal(t, int64(1), s.s.trie.NumNodes())
	assert.Equal(t, int64(256), s.Size())
	assert.True(t, s.Contains(Ten24))
	assert.True(t, s.Contains(Ten24128))
	assert.False(t, s.Contains(Nines))
	assert.False(t, s.Contains(Eights))
	assert.True(t, s.isValid())
}

func TestOldSetInsertSequential(t *testing.T) {
	s := NewSet()

	s.Insert(_a("192.168.1.0"))
	assert.Equal(t, int64(1), s.s.trie.NumNodes())
	s.Insert(_a("192.168.1.1"))
	assert.Equal(t, int64(1), s.s.trie.NumNodes())
	s.Insert(_a("192.168.1.2"))
	assert.Equal(t, int64(2), s.s.trie.NumNodes())
	s.Insert(_a("192.168.1.3"))
	assert.Equal(t, int64(1), s.s.trie.NumNodes())
	assert.Equal(t, int64(4), s.Size())

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
	s := NewSet()

	s.Insert(Ten24)
	assert.Equal(t, int64(1), s.s.trie.NumNodes())
	s.Remove(Ten24128)
	assert.Equal(t, int64(1), s.s.trie.NumNodes())
	assert.Equal(t, int64(128), s.Size())
	assert.False(t, s.Contains(Ten24))
	assert.False(t, s.Contains(Ten24128))
	cidr := _p("10.0.0.0/25")
	assert.True(t, s.Contains(cidr))

	s.Remove(Ten24Router)
	assert.Equal(t, int64(127), s.Size())
	assert.Equal(t, int64(7), s.s.trie.NumNodes())
	assert.True(t, s.isValid())
}

func TestOldSetRemovePrefixworkBroadcast(t *testing.T) {
	s := NewSet()

	s.Insert(Ten24)
	assert.Equal(t, int64(1), s.s.trie.NumNodes())
	s.Remove(Ten24.Address())
	s.Remove(Ten24Broadcast)
	assert.Equal(t, int64(254), s.Size())
	assert.Equal(t, int64(14), s.s.trie.NumNodes())
	assert.False(t, s.Contains(Ten24))
	assert.False(t, s.Contains(Ten24128))
	assert.False(t, s.Contains(Ten24Broadcast))
	assert.False(t, s.Contains(Ten24.Address()))

	cidr := _p("10.0.0.128/26")
	assert.True(t, s.Contains(cidr))
	assert.True(t, s.Contains(Ten24Router))

	s.Remove(Ten24Router)
	assert.Equal(t, int64(253), s.Size())
	assert.Equal(t, int64(13), s.s.trie.NumNodes())
	assert.True(t, s.isValid())
}

func TestOldSetRemoveAll(t *testing.T) {
	s := NewSet()

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
	testSet := NewSet()
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
	s := NewSet()

	s.Insert(Ten24128)
	assert.False(t, s.Contains(Ten24))
	assert.Equal(t, int64(1), s.s.trie.NumNodes())
	s.Insert(Ten24)
	assert.Equal(t, int64(1), s.s.trie.NumNodes())
	assert.Equal(t, int64(256), s.Size())
	assert.True(t, s.Contains(Ten24))
	assert.True(t, s.Contains(Ten24Router))
	assert.False(t, s.Contains(Eights))
	assert.False(t, s.Contains(Nines))
	assert.True(t, s.isValid())
}

func TestOldSetUnion(t *testing.T) {
	set1 := NewSet()
	set2 := NewSet()

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
	set1 := NewSet()
	set2 := NewSet()

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
	set1 := NewSet()
	set2 := NewSet()
	interSect := NewSet()
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
	assert.True(t, interSect.EqualInterface(set))
	assert.True(t, set1.isValid())
	assert.True(t, set2.isValid())
	assert.True(t, set.isValid())

}

func TestOldSetAllocateDeallocate(t *testing.T) {
	rand.Seed(29)

	s := NewSet()

	bigNet := _p("15.1.0.0/16")
	s.Insert(bigNet)

	assert.Equal(t, int64(65536), s.Size())

	ips := make([]Address, 0, s.Size())
	s.FixedSet().WalkAddresses(func(ip Address) bool {
		ips = append(ips, ip)
		return true
	})

	allocated := NewSet()
	for i := 0; i != 16384; i++ {
		allocated.Insert(ips[rand.Intn(65536)])
	}
	assert.Equal(t, int64(14500), allocated.Size())
	allocated.FixedSet().WalkAddresses(func(ip Address) bool {
		assert.True(t, s.Contains(ip))
		return true
	})

	available := s.Difference(allocated)
	assert.Equal(t, int64(51036), available.Size())
	available.FixedSet().WalkAddresses(func(ip Address) bool {
		assert.True(t, s.Contains(ip))
		assert.False(t, allocated.Contains(ip))
		return true
	})
	assert.Equal(t, int64(51036), available.Size())
	assert.True(t, s.isValid())
	assert.True(t, allocated.isValid())
	assert.True(t, available.isValid())
}

func TestOldEqualTrivial(t *testing.T) {
	a := NewSet()
	b := NewSet()
	assert.True(t, a.EqualInterface(b))

	a.Insert(_p("10.0.0.0/24"))
	assert.False(t, a.EqualInterface(b))
	assert.False(t, b.EqualInterface(a))
	assert.True(t, a.EqualInterface(a))
	assert.True(t, b.EqualInterface(b))
	assert.True(t, a.isValid())
	assert.True(t, b.isValid())
}

func TestOldEqualSingleNode(t *testing.T) {
	a := NewSet()
	b := NewSet()
	a.Insert(_p("10.0.0.0/24"))
	b.Insert(_p("10.0.0.0/24"))

	assert.True(t, a.EqualInterface(b))
	assert.True(t, b.EqualInterface(a))
	assert.True(t, a.isValid())
	assert.True(t, b.isValid())
}

func TestOldEqualAllIPv4(t *testing.T) {
	a := NewSet()
	b := NewSet()
	c := NewSet()
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
		assert.False(t, a.EqualInterface(b))
		assert.False(t, b.EqualInterface(a))
		b.Insert(n)
		assert.False(t, b.EqualInterface(c))
		assert.False(t, c.EqualInterface(b))
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
		assert.False(t, c.EqualInterface(a))
		assert.False(t, c.EqualInterface(b))
		c.Insert(n)
		assert.True(t, a.isValid())
		assert.True(t, b.isValid())
		assert.True(t, c.isValid())
	}

	// At this point, all three should have the entire IPv4 space
	assert.True(t, a.EqualInterface(b))
	assert.True(t, a.EqualInterface(c))
	assert.True(t, b.EqualInterface(a))
	assert.True(t, b.EqualInterface(c))
	assert.True(t, c.EqualInterface(a))
	assert.True(t, c.EqualInterface(b))
}
