package ipv4

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetBuilderInsertPrefix(t *testing.T) {
	sb := SetBuilder{}
	sb.InsertPrefix(unsafeParsePrefix("10.0.0.0/24"))

	assert.True(t, sb.Set().ContainsPrefix(unsafeParsePrefix("10.0.0.0/24")))
	assert.False(t, sb.Set().ContainsPrefix(unsafeParsePrefix("10.0.0.0/16")))

	sb.InsertPrefix(unsafeParsePrefix("10.0.0.0/16"))

	assert.True(t, sb.Set().ContainsPrefix(unsafeParsePrefix("10.0.0.0/24")))
	assert.True(t, sb.Set().ContainsPrefix(unsafeParsePrefix("10.0.0.0/16")))
}

func TestSetBuilderRemovePrefix(t *testing.T) {
	sb := SetBuilder{}
	sb.InsertPrefix(unsafeParsePrefix("10.0.0.0/16"))

	assert.True(t, sb.Set().ContainsPrefix(unsafeParsePrefix("10.0.0.0/24")))
	assert.True(t, sb.Set().ContainsPrefix(unsafeParsePrefix("10.0.0.0/16")))

	sb.RemovePrefix(unsafeParsePrefix("10.0.0.0/24"))

	assert.False(t, sb.Set().ContainsPrefix(unsafeParsePrefix("10.0.0.0/24")))
	assert.False(t, sb.Set().ContainsPrefix(unsafeParsePrefix("10.0.0.0/16")))
	assert.True(t, sb.Set().ContainsPrefix(unsafeParsePrefix("10.0.1.0/24")))
}
