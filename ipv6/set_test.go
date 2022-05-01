package ipv6

import (
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
	s := NewSet_()
	s.Insert(_p("2001::/64"))
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
