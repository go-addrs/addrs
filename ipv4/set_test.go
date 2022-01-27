package ipv4

import (
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
	s := ImmutableSet{}.Set()
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

	a.RemoveSet(b)
	assert.True(t, a.isValid())
	assert.True(t, a.Contains(_p("10.0.0.0/25")))
	assert.False(t, a.Contains(_p("10.0.0.128/25")))
	assert.False(t, a.Contains(_p("10.0.0.0/24")))
}
