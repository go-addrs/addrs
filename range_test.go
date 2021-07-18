package ipv4

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRangeString(t *testing.T) {
	assert.Equal(t, "[18.52.86.120,35.69.103.137]", NewRange(Addr{ui: 0x12345678}, Addr{ui: 0x23456789}).String())
	assert.Equal(t, "[10.224.24.0,10.224.24.255]", unsafeParsePrefix("10.224.24.1/24").Range().String())
}

func TestRangeSize(t *testing.T) {
	assert.Equal(t, 0, EmptyRange().Size())
	assert.Equal(t, 0x100, unsafeParsePrefix("10.224.24.1/24").Range().Size())
	assert.Equal(t, 0x20000, unsafeParsePrefix("10.224.24.1/15").Range().Size())
	assert.Equal(t, 0x11111112, NewRange(Addr{ui: 0x12345678}, Addr{ui: 0x23456789}).Size())
}

func TestRangeFirstLast(t *testing.T) {
	tests := []struct {
		description string
		r           Range
		first, last Addr
		empty       bool
	}{
		{
			description: "empty",
			r:           EmptyRange(),
			empty:       true,
		},
		{
			description: "unaligned",
			r:           NewRange(Addr{ui: 0x12345678}, Addr{ui: 0x23456789}),
			first:       Addr{ui: 0x12345678},
			last:        Addr{ui: 0x23456789},
		},
		{
			description: "prefix",
			r:           unsafeParsePrefix("10.224.24.1/24").Range(),
			first:       unsafeParseAddr("10.224.24.0"),
			last:        unsafeParseAddr("10.224.24.255"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			firstExists, first := tt.r.First()
			lastExists, last := tt.r.Last()
			assert.NotEqual(t, tt.empty, firstExists)
			assert.NotEqual(t, tt.empty, lastExists)
			if !tt.empty {
				assert.True(t, firstExists)
				assert.True(t, lastExists)
				assert.Equal(t, tt.first, first)
				assert.Equal(t, tt.last, last)
			} else {
				assert.False(t, firstExists)
				assert.False(t, lastExists)
			}
		})
	}
}

func TestRangeContains(t *testing.T) {
	tests := []struct {
		description string
		a, b        Range
	}{
		{
			description: "empty,empty",
			a:           EmptyRange(),
			b:           EmptyRange(),
		},
		{
			description: "empty",
			a:           unsafeParsePrefix("10.224.24.1/24").Range(),
			b:           EmptyRange(),
		},
		{
			description: "larger",
			a:           unsafeParsePrefix("10.224.24.1/22").Range(),
			b:           unsafeParsePrefix("10.224.26.1/24").Range(),
		},
		{
			description: "unaligned",
			a:           NewRange(Addr{ui: 0x12345678}, Addr{ui: 0x23456789}),
			b:           unsafeParsePrefix("20.224.26.1/24").Range(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			assert.True(t, tt.a.Contains(tt.b))
			// If they're equal then containership goes the other way too.
			assert.Equal(t, tt.a == tt.b, tt.b.Contains(tt.a))
		})
	}
}
