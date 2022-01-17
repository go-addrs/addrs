package ipv4

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetContainsPrefix(t *testing.T) {
	sb := SetBuilder{}
	sb.InsertPrefix(unsafeParsePrefix("10.0.0.0/16"))
	assert.True(t, sb.Set().ContainsPrefix(unsafeParsePrefix("10.0.0.0/24")))
	assert.True(t, sb.Set().ContainsPrefix(unsafeParsePrefix("10.0.30.0/27")))
	assert.True(t, sb.Set().ContainsPrefix(unsafeParsePrefix("10.0.128.0/17")))
	assert.False(t, sb.Set().ContainsPrefix(unsafeParsePrefix("10.224.0.0/24")))
	assert.False(t, sb.Set().ContainsPrefix(unsafeParsePrefix("10.1.30.0/27")))
	assert.False(t, sb.Set().ContainsPrefix(unsafeParsePrefix("10.0.128.0/15")))
}
