package ipv6

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
	s.Insert(_p("2001::/112"))

	assert.True(t, s.Contains(_p("2001::/112")))
	assert.False(t, s.Contains(_p("2001::/104")))

	s.Insert(_p("2001::/104"))

	assert.True(t, s.Contains(_p("2001::/112")))
	assert.True(t, s.Contains(_p("2001::/104")))
}

func TestSetRemovePrefix(t *testing.T) {
	s := Set{}.Set_()
	s.Insert(_p("2001::/104"))

	assert.True(t, s.Contains(_p("2001::/112")))
	assert.True(t, s.Contains(_p("2001::/104")))

	s.Remove(_p("2001::/112"))

	assert.False(t, s.Contains(_p("2001::/112")))
	assert.False(t, s.Contains(_p("2001::/104")))
	assert.True(t, s.Contains(_p("2001::1:0/112")))
}

func TestSetAsReferenceType(t *testing.T) {
	s := NewSet_()

	func(s Set_) {
		s.Insert(_p("2001::/112"))
	}(s)

	assert.True(t, s.Contains(_p("2001::/112")))
	assert.False(t, s.Contains(_p("2001::/104")))

	func(s Set_) {
		s.Insert(_p("2001::/104"))
	}(s)

	assert.True(t, s.Contains(_p("2001::/112")))
	assert.True(t, s.Contains(_p("2001::/104")))
}

func TestSetInsertSet(t *testing.T) {
	a, b := NewSet_(), NewSet_()
	a.Insert(_p("2001::/112"))
	b.Insert(_p("2001::8000/113"))

	a.Insert(b)
	assert.True(t, a.isValid())
	assert.True(t, a.Contains(_p("2001::/113")))
	assert.True(t, a.Contains(_p("2001::8000/113")))
	assert.True(t, a.Contains(_p("2001::/112")))
}

func TestSetRemoveSet(t *testing.T) {
	a, b := NewSet_(), NewSet_()
	a.Insert(_p("2001::/112"))
	b.Insert(_p("2001::8000/113"))

	a.Remove(b)
	assert.True(t, a.isValid())
	assert.True(t, a.Contains(_p("2001::/113")))
	assert.False(t, a.Contains(_p("2001::8000/113")))
	assert.False(t, a.Contains(_p("2001::/112")))
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
			return true, set.s.trie.Union(_a("2001::1").Set().trie)
		})
	}()
	go func() {
		defer wrap()
		set.mutate(func() (bool, *setNode) {
			<-ch
			return true, set.s.trie.Union(_a("2001::2").Set().trie)
		})
	}()
	wg.Wait()
	assert.Equal(t, 1, panicked)
}

func TestNilSet(t *testing.T) {
	var set Set_

	nonEmptySet := _p("2001::123:0/112").Set().Set_()

	// On-offs
	assert.True(t, set.IsEmpty())
	assert.True(t, set.Set().IsEmpty())
	assert.False(t, set.Contains(_p("2001::123:0/112")))

	// Equal
	assert.True(t, set.Equal(set))
	assert.True(t, set.Equal(NewSet_()))
	assert.True(t, NewSet_().Equal(set))
	assert.False(t, set.Equal(nonEmptySet))
	assert.False(t, nonEmptySet.Equal(set))

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
	assert.True(t, set.Intersection(nonEmptySet).IsEmpty())
	assert.True(t, nonEmptySet.Intersection(set).IsEmpty())

	// Difference
	assert.True(t, set.Difference(nonEmptySet).IsEmpty())
	assert.False(t, nonEmptySet.Difference(set).IsEmpty())

	// Walk
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
	assert.True(t, Set_{}.Union(nil).IsEmpty())
	assert.True(t, Set{}.Union(nil).IsEmpty())
}

func TestSetIntesectionNil(t *testing.T) {
	assert.True(t, Set_{}.Intersection(nil).IsEmpty())
	assert.True(t, Set{}.Intersection(nil).IsEmpty())
}

func TestSetDifferenceNil(t *testing.T) {
	assert.True(t, Set_{}.Difference(nil).IsEmpty())
	assert.True(t, Set{}.Difference(nil).IsEmpty())
}

func TestSetInsertNil(t *testing.T) {
	s := NewSet_()
	s.Insert(nil)
	assert.True(t, s.IsEmpty())
}

func TestSetRemoveNil(t *testing.T) {
	s := NewSet_()
	s.Remove(nil)
	assert.True(t, s.IsEmpty())
}

func TestFixedSetContainsPrefix(t *testing.T) {
	s := Set{}.Build(func(s_ Set_) bool {
		s_.Insert(_p("2001::/64"))
		return true
	})
	s = s.Build(func(s_ Set_) bool {
		s_.Insert(_p("2001:ffff::/64"))
		return false
	})
	assert.True(t, s.Set().Contains(_p("2001::/112")))
	assert.True(t, s.Set().Contains(_p("2001::1234:0/115")))
	assert.True(t, s.Set().Contains(_p("2001:0:0:0:8000:0:0:0/65")))
	assert.False(t, s.Set().Contains(_p("2001:1234::/112")))
	assert.False(t, s.Set().Contains(_p("2001:1::1234:0/115")))
	assert.False(t, s.Set().Contains(_p("2001:0:0:0:8000:0:0:0/63")))
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
				_p("2001::123:0/112"),
			},
			ranges: []Range{
				_p("2001::123:0/112").Range(),
			},
		}, {
			description: "adjacent prefixes",
			prefixes: []SetI{
				_p("2001::123:0/113"),
				_p("2001::123:8000/114"),
			},
			ranges: []Range{
				_r(_a("2001::123:0"), _a("2001::123:bfff")),
			},
		}, {
			description: "disjoint prefixes",
			prefixes: []SetI{
				_p("2001::123:0/113"),
				_p("2001::123:c000/114"),
			},
			ranges: []Range{
				_r(_a("2001::123:0"), _a("2001::123:7fff")),
				_r(_a("2001::123:c000"), _a("2001::123:ffff")),
			},
		}, {
			// This is the reverse of the complicated test from range_test.go
			description: "complicated",
			prefixes: []SetI{
				_p("2001::241:122/127"),
				_p("2001::241:124/126"),
				_p("2001::241:128/125"),
				_p("2001::241:130/124"),
				_p("2001::241:140/122"),
				_p("2001::241:180/121"),
				_p("2001::241:200/119"),
				_p("2001::241:400/118"),
				_p("2001::241:800/117"),
				_p("2001::241:1000/116"),
				_p("2001::241:2000/115"),
				_p("2001::241:4000/114"),
				_p("2001::241:8000/113"),
				_p("2001::242:0/111"),
				_p("2001::244:0/110"),
				_p("2001::248:0/109"),
				_p("2001::250:0/108"),
				_p("2001::260:0/107"),
				_p("2001::280:0/105"),
				_p("2001::300:0/104"),
				_p("2001::400:0/102"),
				_p("2001::800:0/101"),
				_p("2001::1000:0/100"),
				_p("2001::2000:0/99"),
				_p("2001::4000:0/98"),
				_p("2001::8000:0/99"),
				_p("2001::a000:0/101"),
				_p("2001::a800:0/102"),
				_p("2001::ac00:0/104"),
				_p("2001::ad00:0/107"),
				_p("2001::ad20:0/111"),
				_p("2001::ad22:0/112"),
				_p("2001::ad23:0/117"),
				_p("2001::ad23:800/123"),
				_p("2001::ad23:820/126"),
				_p("2001::ad23:824/128"),
			},
			ranges: []Range{
				_r(_a("2001::0241:0122"), _a("2001::ad23:824")),
			},
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
			})

			t.Run("don't finish", func(t *testing.T) {
				s := func() Set {
					s := NewSet_()
					for _, p := range tt.prefixes {
						s.Insert(p)
					}
					return s.Set()
				}()
				if !s.IsEmpty() {
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
	s.Insert(_p("2001::/64"))

	other := NewSet_()
	other.Insert(_p("2001::/112"))
	other.Insert(_p("2001::1234:0/115"))
	other.Insert(_p("2001:0:0:0:8000:0:0:0/65"))

	assert.True(t, s.Contains(other))

	other.Insert(_p("2001:1234::/112"))
	other.Insert(_p("2001:1::1234:0/115"))
	other.Insert(_p("2001:0:0:0:8000:0:0:0/63"))

	assert.False(t, s.Contains(other))

}

var (
	Eights = _a("8:8:8:8:8:8:8:8")
	Nines  = _a("9:9:9:9:9:9:9:9")

	Ten112          = _p("10::/112")
	TenOne112       = _p("10::1:0/112")
	TenTwo112       = _p("10::2:0/112")
	Ten112x8000     = _p("10::8000/113")
	Ten112Router    = _a("10::1")
	Ten112Broadcast = _a("10::ffff")
)

func TestOldSetInit(t *testing.T) {
	set := Set{}

	assert.True(t, set.IsEmpty())
	assert.True(t, set.isValid())
}

func TestOldSetContains(t *testing.T) {
	set := Set{}

	assert.True(t, set.IsEmpty())
	assert.False(t, set.Contains(Eights))
	assert.False(t, set.Contains(Nines))
	assert.True(t, set.isValid())
}

func TestOldSetInsert(t *testing.T) {
	s := NewSet_()

	s.Insert(Nines)
	assert.False(t, s.IsEmpty())
	assert.True(t, s.Contains(Nines))
	assert.False(t, s.Contains(Eights))
	s.Insert(Eights)
	assert.Equal(t, int64(2), s.s.trie.NumNodes())
	assert.True(t, s.Contains(Eights))
	assert.True(t, s.isValid())
}

func TestOldSetInsertPrefixwork(t *testing.T) {
	s := NewSet_()

	s.Insert(Ten112)
	assert.Equal(t, int64(1), s.s.trie.NumNodes())
	assert.False(t, s.IsEmpty())
	assert.True(t, s.Contains(Ten112))
	assert.True(t, s.Contains(Ten112x8000))
	assert.False(t, s.Contains(Nines))
	assert.False(t, s.Contains(Eights))
	assert.True(t, s.isValid())
}

func TestOldSetInsertSequential(t *testing.T) {
	s := NewSet_()

	s.Insert(_a("2001::"))
	assert.Equal(t, int64(1), s.s.trie.NumNodes())
	s.Insert(_a("2001::1"))
	assert.Equal(t, int64(1), s.s.trie.NumNodes())
	s.Insert(_a("2001::2"))
	assert.Equal(t, int64(2), s.s.trie.NumNodes())
	s.Insert(_a("2001::3"))
	assert.Equal(t, int64(1), s.s.trie.NumNodes())
	assert.False(t, s.IsEmpty())

	cidr := _p("2001::/126")
	assert.True(t, s.Contains(cidr))

	cidr = _p("2001::4/127")
	s.Insert(cidr)
	assert.Equal(t, int64(2), s.s.trie.NumNodes())
	assert.True(t, s.Contains(cidr))

	cidr = _p("2001::6/127")
	s.Insert(cidr)
	assert.Equal(t, int64(1), s.s.trie.NumNodes())
	assert.True(t, s.Contains(cidr))

	cidr = _p("2001::6/127")
	s.Insert(cidr)
	assert.Equal(t, int64(1), s.s.trie.NumNodes())
	assert.True(t, s.Contains(cidr))

	cidr = _p("2001::240/125")
	s.Insert(cidr)
	assert.Equal(t, int64(2), s.s.trie.NumNodes())
	assert.True(t, s.Contains(cidr))

	cidr = _p("2001::248/125")
	s.Insert(cidr)
	assert.Equal(t, int64(2), s.s.trie.NumNodes())
	assert.True(t, s.Contains(cidr))
	assert.True(t, s.isValid())
}

func TestOldSetRemove(t *testing.T) {
	s := NewSet_()

	s.Insert(Ten112)
	assert.Equal(t, int64(1), s.s.trie.NumNodes())
	s.Remove(Ten112x8000)
	assert.Equal(t, int64(1), s.s.trie.NumNodes())
	assert.False(t, s.IsEmpty())
	assert.False(t, s.Contains(Ten112))
	assert.False(t, s.Contains(Ten112x8000))
	cidr := _p("10::/113")
	assert.True(t, s.Contains(cidr))

	s.Remove(Ten112Router)
	assert.False(t, s.IsEmpty())
	assert.Equal(t, int64(15), s.s.trie.NumNodes())
	assert.True(t, s.isValid())
}

func TestOldSetRemovePrefixworkBroadcast(t *testing.T) {
	s := NewSet_()

	s.Insert(Ten112)
	assert.Equal(t, int64(1), s.s.trie.NumNodes())
	s.Remove(Ten112.Address())
	s.Remove(Ten112Broadcast)
	assert.False(t, s.IsEmpty())
	assert.Equal(t, int64(30), s.s.trie.NumNodes())
	assert.False(t, s.Contains(Ten112))
	assert.False(t, s.Contains(Ten112x8000))
	assert.False(t, s.Contains(Ten112Broadcast))
	assert.False(t, s.Contains(Ten112.Address()))

	cidr := _p("10::8000/114")
	assert.True(t, s.Contains(cidr))
	assert.True(t, s.Contains(Ten112Router))

	s.Remove(Ten112Router)
	assert.False(t, s.IsEmpty())
	assert.Equal(t, int64(29), s.s.trie.NumNodes())
	assert.True(t, s.isValid())
}

func TestOldSetRemoveAll(t *testing.T) {
	s := NewSet_()

	s.Insert(Ten112)
	cidr1 := _p("2001::/113")
	s.Insert(cidr1)
	assert.Equal(t, int64(2), s.s.trie.NumNodes())

	cidr2 := _p("::/0")
	s.Remove(cidr2)
	assert.Equal(t, int64(0), s.s.trie.NumNodes())
	assert.False(t, s.Contains(Ten112))
	assert.False(t, s.Contains(Ten112x8000))
	assert.False(t, s.Contains(cidr1))
	assert.True(t, s.isValid())
}

func TestOldSet_RemoveTop(t *testing.T) {
	testSet := NewSet_()
	ip1 := _a("10::1")
	ip2 := _a("10::2")

	testSet.Insert(ip2) // top
	testSet.Insert(ip1) // inserted at left
	testSet.Remove(ip2) // remove top node

	assert.True(t, testSet.Contains(ip1))
	assert.False(t, testSet.Contains(ip2))
	assert.True(t, testSet.isValid())
}

func TestOldSetInsertOverlapping(t *testing.T) {
	s := NewSet_()

	s.Insert(Ten112x8000)
	assert.False(t, s.Contains(Ten112))
	assert.Equal(t, int64(1), s.s.trie.NumNodes())
	s.Insert(Ten112)
	assert.Equal(t, int64(1), s.s.trie.NumNodes())
	assert.False(t, s.IsEmpty())
	assert.True(t, s.Contains(Ten112))
	assert.True(t, s.Contains(Ten112Router))
	assert.False(t, s.Contains(Eights))
	assert.False(t, s.Contains(Nines))
	assert.True(t, s.isValid())
}

func TestOldSetUnion(t *testing.T) {
	set1 := NewSet_()
	set2 := NewSet_()

	set1.Insert(Ten112)
	cidr := _p("2001::248/125")
	set2.Insert(cidr)

	set := set1.Union(set2)
	assert.True(t, set.Contains(Ten112))
	assert.True(t, set.Contains(cidr))
	assert.True(t, set1.isValid())
	assert.True(t, set2.isValid())
}

func TestOldSetDifference(t *testing.T) {
	set1 := NewSet_()
	set2 := NewSet_()

	set1.Insert(Ten112)
	cidr := _p("2001::248/125")
	set2.Insert(cidr)

	set := set1.Difference(set2)
	assert.True(t, set.Contains(Ten112))
	assert.False(t, set.Contains(cidr))
	assert.True(t, set1.isValid())
	assert.True(t, set2.isValid())
}

func TestOldIntersectionAinB1(t *testing.T) {
	case1 := []string{"10::16:0/104", "10::5:8:0/112", "10::23:224:0/110"}
	case2 := []string{"10::20:0/124", "10::5:8:0/122", "10::23:224:0/118"}
	output := []string{"10::23:224:0/118", "10::20:0/124", "10::5:8:0/122"}
	testIntersection(t, case1, case2, output)

}

func TestOldIntersectionAinB2(t *testing.T) {
	case1 := []string{"10::10:0:0/124", "10::5:8:0/122", "10::23:224:0/118"}
	case2 := []string{"10::10:0:0/104", "10::5:8:0/112", "10::23:224:0/110"}
	output := []string{"10::10:0:0/124", "10::5:8:0/122", "10::23:224:0/118"}
	testIntersection(t, case1, case2, output)
}

func TestOldIntersectionAinB3(t *testing.T) {
	case1 := []string{"10::5:0/112", "10::5:8:0/122", "10::23:224:0/118"}
	case2 := []string{"10::6:0:0/112", "10::9:9:0/122", "10::23:6:0/110"}
	output := []string{}
	testIntersection(t, case1, case2, output)
}

func TestOldIntersectionAinB4(t *testing.T) {
	case1 := []string{"10::23:6:0/112", "10::5:8:0/122", "10::23:224:0/118"}
	case2 := []string{"10::6:0:0/112", "10::9:9:0/122", "10::23:6:0/122"}
	output := []string{"10::23:6:0/122"}
	testIntersection(t, case1, case2, output)
}

func TestOldIntersectionAinB5(t *testing.T) {
	case1 := []string{"10::23:0/118", "10::20:0/118", "10::15:0/118"}
	case2 := []string{"10::23:0/112", "10::20:0/112", "10::15:0/112"}
	output := []string{"10::23:0/118", "10::20:0/118", "10::15:0/118"}
	testIntersection(t, case1, case2, output)
}

func TestOldIntersectionAinB6(t *testing.T) {
	case1 := []string{"10::23:0/112", "10::20:0/112", "10::15:0/112"}
	case2 := []string{"10::23:0/118", "10::20:0/118", "10::15:0/118"}
	output := []string{"10::15:0/118", "10::20:0/118", "10::23:0/118"}
	testIntersection(t, case1, case2, output)
}

func TestOldIntersectionAinB7(t *testing.T) {
	case1 := []string{"10::23:0/112", "10::20:0/112", "10::15:0/112"}
	case2 := []string{"10::14:0/118", "10::10:0/118", "10::8:0/118"}
	output := []string{}
	testIntersection(t, case1, case2, output)
}

func TestOldIntersectionAinB8(t *testing.T) {
	case1 := []string{"10::23:0/112", "10::20:0/112", "172::16:1:0/112"}
	case2 := []string{"10::14:0/118", "10::10:0/118", "172::16:1:0/120"}
	output := []string{"172::16:1:0/120"}
	testIntersection(t, case1, case2, output)
}

func TestOldIntersectionAinB9(t *testing.T) {
	case1 := []string{"10::5:8:0/122"}
	case2 := []string{"10::10:0:0/104", "10::5:8:0/112", "10::23:224:0/110"}
	output := []string{"10::5:8:0/122"}
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

func TestOldEqualTrivial(t *testing.T) {
	a := NewSet_()
	b := NewSet_()
	assert.True(t, a.Equal(b))

	a.Insert(_p("10::/112"))
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
	a.Insert(_p("10::/112"))
	b.Insert(_p("10::/112"))

	assert.True(t, a.Equal(b))
	assert.True(t, b.Equal(a))
	assert.True(t, a.isValid())
	assert.True(t, b.isValid())
}

func TestOldEqualAllIPv6(t *testing.T) {
	a := NewSet_()
	b := NewSet_()
	c := NewSet_()
	// Insert the entire IPv6 space into set a in one shot
	a.Insert(_p("::/0"))

	// Insert the entire IPv6 space piece by piece into b and c

	// This list was constructed starting with ::/128 and ::1/128,
	// then adding ::2/127, ::4/126, ..., 128::/1
	bNets := []Prefix{_p("::/128")}
	mask := uint128{0x8000000000000000, 0}
	for length := 129; length > 0; length-- {
		prefix := Prefix{Address{ui: mask.rightShift(length - 1)}, uint32(length)}
		bNets = append(bNets, prefix)
	}

	for _, n := range bNets {
		assert.False(t, a.Equal(b))
		assert.False(t, b.Equal(a))
		b.Insert(n)
		assert.False(t, b.Equal(c))
		assert.False(t, c.Equal(b))
	}

	// Constructed a different way
	cNets := []Prefix{_p("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff/128")}
	mask = uint128{0xffffffffffffffff, 0xffffffffffffffff}
	for length := 128; length > 0; length-- {
		prefix := Prefix{Address{ui: mask.leftShift(128 - (length - 1))}, uint32(length)}
		cNets = append(cNets, prefix)
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
			prefixes:    []SetI{_p("2001:db8::/56")},
			length:      64,
			count:       256,
		}, {
			description: "overflow",
			prefixes:    []SetI{_p("2001:db8::/56")},
			length:      128,
			error:       true,
		}, {
			description: "overflow with multiple valid prefixes",
			prefixes: []SetI{
				_p("2001:db8:0:0::/65"),
				_p("2001:db8:0:1::/65"),
			},
			length: 128,
			error:  true,
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
		{length: 48, count: 0x00001},
		{length: 49, count: 0x00003},
		{length: 50, count: 0x00007},
		{length: 51, count: 0x0000f},
		{length: 52, count: 0x0001f},
		{length: 53, count: 0x0003f},
		{length: 54, count: 0x0007f},
		{length: 55, count: 0x000ff},
		{length: 56, count: 0x001ff},
		{length: 57, count: 0x003ff},
		{length: 58, count: 0x007ff},
		{length: 59, count: 0x00fff},
		{length: 60, count: 0x01fff},
		{length: 61, count: 0x03fff},
		{length: 62, count: 0x07fff},
		{length: 63, count: 0x0ffff},
		{length: 64, count: 0x1ffff},
		{length: 65, count: 0x3fffe},
		{length: 66, count: 0x7fffc},
	}

	s := Set{}.Build(func(s Set_) bool {
		s.Insert(_p("2001:db8:0:0000::/48"))
		s.Insert(_p("2001:db8:1:0000::/49"))
		s.Insert(_p("2001:db8:1:8000::/50"))
		s.Insert(_p("2001:db8:1:c000::/51"))
		s.Insert(_p("2001:db8:1:e000::/52"))
		s.Insert(_p("2001:db8:1:f000::/53"))
		s.Insert(_p("2001:db8:1:f800::/54"))
		s.Insert(_p("2001:db8:1:fc00::/55"))
		s.Insert(_p("2001:db8:1:fe00::/56"))
		s.Insert(_p("2001:db8:1:ff00::/57"))
		s.Insert(_p("2001:db8:1:ff80::/58"))
		s.Insert(_p("2001:db8:1:ffc0::/59"))
		s.Insert(_p("2001:db8:1:ffe0::/60"))
		s.Insert(_p("2001:db8:1:fff0::/61"))
		s.Insert(_p("2001:db8:1:fff8::/62"))
		s.Insert(_p("2001:db8:1:fffc::/63"))
		s.Insert(_p("2001:db8:1:fffe::/64"))
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

func TestFindPrefixWithLength(t *testing.T) {
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
				_p("::ffff:10.0.0.0/104"),
			},
			length: 120,
			change: 1,
		}, {
			description: "find adjacent",
			space: []SetI{
				_p("::ffff:10.0.0.0/104"),
			},
			reserved: []SetI{
				_p("::ffff:10.224.123.0/120"),
			},
			length:   120,
			expected: _p("::ffff:10.224.122.0/120"),
		}, {
			description: "many fewer prefixes",
			space: []SetI{
				_p("::ffff:10.0.0.0/112"),
			},
			reserved: []SetI{
				_p("::ffff:10.0.1.0/120"),
				_p("::ffff:10.0.2.0/119"),
				_p("::ffff:10.0.4.0/118"),
				_p("::ffff:10.0.8.0/117"),
				_p("::ffff:10.0.16.0/116"),
				_p("::ffff:10.0.32.0/115"),
				_p("::ffff:10.0.64.0/114"),
				_p("::ffff:10.0.128.0/113"),
			},
			length: 120,
			change: -7,
		}, {
			description: "toobig",
			space: []SetI{
				_p("::ffff:10.0.0.0/104"),
			},
			reserved: []SetI{
				_p("::ffff:10.128.0.0/105"),
				_p("::ffff:10.64.0.0/106"),
				_p("::ffff:10.32.0.0/107"),
				_p("::ffff:10.16.0.0/108"),
			},
			length: 107,
			err:    true,
		}, {
			description: "full",
			space: []SetI{
				_p("::ffff:10.0.0.0/104"),
			},
			length: 103,
			err:    true,
		}, {
			description: "random disjoint example",
			space: []SetI{
				_p("::ffff:10.0.0.0/118"),
				_p("::ffff:192.168.0.0/117"),
				_p("::ffff:172.16.0.0/116"),
			},
			reserved: []SetI{
				_p("::ffff:192.168.0.0/117"),
				_p("::ffff:172.16.0.0/117"),
				_p("::ffff:172.16.8.0/118"),
				_p("::ffff:10.0.0.0/118"),
				_p("::ffff:172.16.12.0/120"),
				_p("::ffff:172.16.14.0/120"),
				_p("::ffff:172.16.15.0/120"),
			},
			length:   120,
			expected: _p("::ffff:172.16.13.0/120"),
			change:   1,
		}, {
			description: "too fragmented",
			space: []SetI{
				_p("::ffff:10.0.0.0/120"),
			},
			reserved: []SetI{
				_p("::ffff:10.0.0.0/123"),
				_p("::ffff:10.0.0.64/123"),
				_p("::ffff:10.0.0.128/123"),
				_p("::ffff:10.0.0.192/123"),
			},
			length: 121,
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
			prefix, err := space.FindPrefixWithLength(reserved, tt.length)

			assert.Equal(t, tt.err, err != nil)
			if err != nil {
				return
			}

			assert.True(t, reserved.Intersection(prefix).IsEmpty())

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
		space := _p("::ffff:10.128.0.0/100").Set()
		available, _ := space.NumPrefixes(128)

		reserved := NewSet_()

		rand.Seed(29)
		for available > 0 {
			// This is the most we can pull out. Assuming we avoid
			// fragmentation, it should be the largest power of two that is
			// less than or equal to the number of available addresses.
			maxExponent := log2(available)

			// Finding the maximum prefix here, proves we are avoiding fragmentation
			maxPrefix, err := space.FindPrefixWithLength(reserved, 128-maxExponent)
			assert.Nil(t, err)
			maxPrefixes, _ := maxPrefix.NumPrefixes(128)
			assert.Equal(t, pow2(maxExponent), maxPrefixes)
			assert.True(t, reserved.Intersection(maxPrefix).IsEmpty())

			// Pull out a random sized prefix up to the maximum size to attempt to further fragment the space.
			randomSize := (rand.Uint32()%maxExponent + 1)
			if randomSize > 12 {
				randomSize = 12
			}

			randomSizePrefix, err := space.FindPrefixWithLength(reserved, 128-randomSize)
			assert.Nil(t, err)
			randomSizePrefixes, _ := randomSizePrefix.NumPrefixes(128)
			assert.Equal(t, pow2(randomSize), randomSizePrefixes)
			assert.True(t, reserved.Intersection(randomSizePrefix).IsEmpty())

			// Reserve only the random sized one
			reserved.Insert(randomSizePrefix)
			available -= randomSizePrefixes
			spacePrefixes, _ := space.NumPrefixes(128)
			reservedPrefixes, _ := reserved.Set().NumPrefixes(128)
			assert.Equal(t, available, spacePrefixes-reservedPrefixes)
		}
	})

	t.Run("randomized 64", func(t *testing.T) {
		// Essentially the same as the previous test but hits the upper 64 of the address range
		// Start with a space and an empty reserved set.
		// This test will attempt to fragment the space by pulling out
		space := _p("2001:db8::/48").Set()
		available, _ := space.NumPrefixes(64)

		reserved := NewSet_()

		rand.Seed(17)
		for available > 0 {
			// This is the most we can pull out. Assuming we avoid
			// fragmentation, it should be the largest power of two that is
			// less than or equal to the number of available addresses.
			maxExponent := log2(available)

			// Finding the maximum prefix here, proves we are avoiding fragmentation
			maxPrefix, err := space.FindPrefixWithLength(reserved, 64-maxExponent)
			assert.Nil(t, err)
			maxPrefixes, _ := maxPrefix.NumPrefixes(64)
			assert.Equal(t, pow2(maxExponent), maxPrefixes)
			assert.True(t, reserved.Intersection(maxPrefix).IsEmpty())

			// Pull out a random sized prefix up to the maximum size to attempt to further fragment the space.
			randomSize := (rand.Uint32()%maxExponent + 1)
			if randomSize > 12 {
				randomSize = 12
			}

			randomSizePrefix, err := space.FindPrefixWithLength(reserved, 64-randomSize)
			assert.Nil(t, err)
			randomSizePrefixes, _ := randomSizePrefix.NumPrefixes(64)
			assert.Equal(t, pow2(randomSize), randomSizePrefixes)
			assert.True(t, reserved.Intersection(randomSizePrefix).IsEmpty())

			// Reserve only the random sized one
			reserved.Insert(randomSizePrefix)
			available -= randomSizePrefixes
			spacePrefixes, _ := space.NumPrefixes(64)
			reservedPrefixes, _ := reserved.Set().NumPrefixes(64)
			assert.Equal(t, available, spacePrefixes-reservedPrefixes)
		}
	})
}

func pow2(x uint32) uint64 {
	return uint64(math.Pow(2, float64(x)))
}

func log2(i uint64) uint32 {
	return uint32(math.Log2(float64(i)))
}

func countPrefixes(s Set) int {
	var numPrefixes int
	s.WalkPrefixes(func(_ Prefix) bool {
		numPrefixes += 1
		return true
	})
	return numPrefixes
}
