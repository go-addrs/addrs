package ipv4

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetInsertPrefix(t *testing.T) {
	s := NewSet()
	s.InsertPrefix(unsafeParsePrefix("10.0.0.0/24"))

	assert.True(t, s.ContainsPrefix(unsafeParsePrefix("10.0.0.0/24")))
	assert.False(t, s.ContainsPrefix(unsafeParsePrefix("10.0.0.0/16")))

	s.InsertPrefix(unsafeParsePrefix("10.0.0.0/16"))

	assert.True(t, s.ContainsPrefix(unsafeParsePrefix("10.0.0.0/24")))
	assert.True(t, s.ContainsPrefix(unsafeParsePrefix("10.0.0.0/16")))
}

func TestSetRemovePrefix(t *testing.T) {
	s := ImmutableSet{}.Set()
	s.InsertPrefix(unsafeParsePrefix("10.0.0.0/16"))

	assert.True(t, s.ContainsPrefix(unsafeParsePrefix("10.0.0.0/24")))
	assert.True(t, s.ContainsPrefix(unsafeParsePrefix("10.0.0.0/16")))

	s.RemovePrefix(unsafeParsePrefix("10.0.0.0/24"))

	assert.False(t, s.ContainsPrefix(unsafeParsePrefix("10.0.0.0/24")))
	assert.False(t, s.ContainsPrefix(unsafeParsePrefix("10.0.0.0/16")))
	assert.True(t, s.ContainsPrefix(unsafeParsePrefix("10.0.1.0/24")))
}

func TestSetAsReferenceType(t *testing.T) {
	s := NewSet()

	func(s Set) {
		s.InsertPrefix(unsafeParsePrefix("10.0.0.0/24"))
	}(s)

	assert.True(t, s.ContainsPrefix(unsafeParsePrefix("10.0.0.0/24")))
	assert.False(t, s.ContainsPrefix(unsafeParsePrefix("10.0.0.0/16")))

	func(s Set) {
		s.InsertPrefix(unsafeParsePrefix("10.0.0.0/16"))
	}(s)

	assert.True(t, s.ContainsPrefix(unsafeParsePrefix("10.0.0.0/24")))
	assert.True(t, s.ContainsPrefix(unsafeParsePrefix("10.0.0.0/16")))
}

func TestSetInsertSet(t *testing.T) {
	a, b := NewSet(), NewSet()
	a.InsertPrefix(unsafeParsePrefix("10.0.0.0/25"))
	b.InsertPrefix(unsafeParsePrefix("10.0.0.128/25"))

	a.InsertSet(b)
	assert.True(t, a.isValid())
	assert.True(t, a.ContainsPrefix(unsafeParsePrefix("10.0.0.0/25")))
	assert.True(t, a.ContainsPrefix(unsafeParsePrefix("10.0.0.128/25")))
	assert.True(t, a.ContainsPrefix(unsafeParsePrefix("10.0.0.0/24")))
}

func TestSetRemoveSet(t *testing.T) {
	a, b := NewSet(), NewSet()
	a.InsertPrefix(unsafeParsePrefix("10.0.0.0/24"))
	b.InsertPrefix(unsafeParsePrefix("10.0.0.128/25"))

	a.RemoveSet(b)
	assert.True(t, a.isValid())
	assert.True(t, a.ContainsPrefix(unsafeParsePrefix("10.0.0.0/25")))
	assert.False(t, a.ContainsPrefix(unsafeParsePrefix("10.0.0.128/25")))
	assert.False(t, a.ContainsPrefix(unsafeParsePrefix("10.0.0.0/24")))
}
