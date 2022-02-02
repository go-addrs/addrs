package ipv4

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
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
			return true, set.s.trie.Union(NewFixedSet(_a("10.0.0.1")).trie)
		})
	}()
	go func() {
		defer wrap()
		set.mutate(func() (bool, *setNode) {
			<-ch
			return true, set.s.trie.Union(NewFixedSet(_a("10.0.0.2")).trie)
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
	assert.True(t, set.WalkAddresses(func(Address) bool {
		panic("should not be called")
	}))
	assert.True(t, set.WalkPrefixes(func(Prefix) bool {
		panic("should not be called")
	}))
	assert.True(t, set.WalkRanges(func(Range) bool {
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
