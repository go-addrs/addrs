package ipv4

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func validRange(t *testing.T, first, last Addr) Range {
	r, err := NewRange(first, last)
	assert.Nil(t, err)
	return r
}

func TestNewRange(t *testing.T) {
	rangeError := func(first, last Addr) error {
		_, err := NewRange(first, last)
		return err
	}

	assert.Nil(t, rangeError(Addr{100}, Addr{200}))
	assert.Nil(t, rangeError(Addr{100}, Addr{100}))
	assert.NotNil(t, rangeError(Addr{200}, Addr{100}))
	assert.NotNil(t, rangeError(Addr{200}, Addr{199}))
	assert.NotNil(t, rangeError(Addr{0xffffffff}, Addr{0}))
}

func TestRangeString(t *testing.T) {
	assert.Equal(t, "[18.52.86.120,35.69.103.137]", validRange(t, Addr{ui: 0x12345678}, Addr{ui: 0x23456789}).String())
	assert.Equal(t, "[10.224.24.0,10.224.24.255]", unsafeParsePrefix("10.224.24.1/24").Range().String())
}

func TestRangeSize(t *testing.T) {
	assert.Equal(t, 0x100, unsafeParsePrefix("10.224.24.1/24").Range().Size())
	assert.Equal(t, 0x20000, unsafeParsePrefix("10.224.24.1/15").Range().Size())
	assert.Equal(t, 0x11111112, validRange(t, Addr{ui: 0x12345678}, Addr{ui: 0x23456789}).Size())
}

func TestRangeFirstLast(t *testing.T) {
	tests := []struct {
		description string
		r           Range
		first, last Addr
	}{
		{
			description: "unaligned",
			r:           validRange(t, Addr{ui: 0x12345678}, Addr{ui: 0x23456789}),
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
			assert.Equal(t, tt.first, tt.r.First())
			assert.Equal(t, tt.last, tt.r.Last())
		})
	}
}

func TestRangeContains(t *testing.T) {
	tests := []struct {
		description string
		a, b        Range
	}{
		{
			description: "larger",
			a:           unsafeParsePrefix("10.224.24.1/22").Range(),
			b:           unsafeParsePrefix("10.224.26.1/24").Range(),
		},
		{
			description: "unaligned",
			a:           validRange(t, Addr{ui: 0x12345678}, Addr{ui: 0x23456789}),
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
