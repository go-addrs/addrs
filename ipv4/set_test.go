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
