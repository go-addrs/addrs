//go:build go1.18
// +build go1.18

package ipv4

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixedTableT(t *testing.T) {
	addrOne := _a("10.224.24.1")
	addrTwo := _a("10.224.24.2")
	addrThree := _a("10.224.24.3")

	m := NewTable[int]()
	succeeded := m.Insert(addrOne, 1)
	assert.True(t, succeeded)

	im := m.FixedTable()
	succeeded = m.Insert(addrTwo, 2)
	assert.True(t, succeeded)

	m2 := im.Table()
	succeeded = m2.Insert(addrThree, 3)
	assert.True(t, succeeded)

	var found bool
	var value int

	value, found = m.Get(addrOne)
	assert.True(t, found)
	assert.Equal(t, 1, value)
	value, found = m.Get(addrTwo)
	assert.True(t, found)
	assert.Equal(t, 2, value)
	_, found = m.Get(addrThree)
	assert.False(t, found)

	assert.Equal(t, int64(1), im.Size())
	value, found = im.Get(addrOne)
	assert.True(t, found)
	assert.Equal(t, 1, value)
	_, found = im.Get(addrTwo)
	assert.False(t, found)
	_, found = im.Get(addrThree)
	assert.False(t, found)

	assert.Equal(t, int64(2), m2.Size())
	value, found = m2.Get(addrOne)
	assert.True(t, found)
	assert.Equal(t, 1, value)
	_, found = m2.Get(addrTwo)
	assert.False(t, found)
	value, found = m2.Get(addrThree)
	assert.True(t, found)
	assert.Equal(t, 3, value)
}
