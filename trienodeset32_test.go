package ipv4

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrieNodeHalves(t *testing.T) {
	set := trieNodeSet32FromPrefix(unsafeParsePrefix("0.0.0.0/0"))
	a, b := set.halves()
	assert.Equal(t, trieNodeSet32FromPrefix(unsafeParsePrefix("0.0.0.0/1")), a)
	assert.Equal(t, trieNodeSet32FromPrefix(unsafeParsePrefix("128.0.0.0/1")), b)
}
