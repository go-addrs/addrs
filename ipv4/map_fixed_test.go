package ipv4

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixedMapMap(t *testing.T) {
	addrOne := _a("10.224.24.1")
	addrTwo := _a("10.224.24.2")
	addrThree := _a("10.224.24.3")

	m := NewMap()
	err := m.Insert(addrOne, nil)
	assert.Nil(t, err)

	im := m.FixedMap()
	err = m.Insert(addrTwo, nil)
	assert.Nil(t, err)

	m2 := im.Map()
	err = m2.Insert(addrThree, nil)
	assert.Nil(t, err)

	var found bool

	_, found = m.Get(addrOne)
	assert.True(t, found)
	_, found = m.Get(addrTwo)
	assert.True(t, found)
	_, found = m.Get(addrThree)
	assert.False(t, found)

	assert.Equal(t, int64(1), im.Size())
	_, found = im.Get(addrOne)
	assert.True(t, found)
	_, found = im.Get(addrTwo)
	assert.False(t, found)
	_, found = im.Get(addrThree)
	assert.False(t, found)

	assert.Equal(t, int64(2), m2.Size())
	_, found = m2.Get(addrOne)
	assert.True(t, found)
	_, found = m2.Get(addrTwo)
	assert.False(t, found)
	_, found = m2.Get(addrThree)
	assert.True(t, found)
}
