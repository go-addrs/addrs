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

func TestRangeMinus(t *testing.T) {
	tests := []struct {
		description string
		a, b        Range
		result      []Range
	}{
		{
			description: "disjoint left",
			a:           unsafeParsePrefix("10.224.24.0/22").Range(),
			b:           unsafeParsePrefix("10.224.0.0/22").Range(),
			result: []Range{
				unsafeParsePrefix("10.224.24.0/22").Range(),
			},
		},
		{
			description: "overlap right",
			a:           Range{Addr{100}, Addr{200}},
			b:           Range{Addr{50}, Addr{150}},
			result: []Range{
				Range{Addr{151}, Addr{200}},
			},
		},
		{
			description: "larger same last",
			a:           unsafeParsePrefix("10.224.27.0/24").Range(),
			b:           unsafeParsePrefix("10.224.24.0/22").Range(),
			result:      []Range{},
		},
		{
			description: "overlap all",
			a:           Range{Addr{100}, Addr{200}},
			b:           Range{Addr{50}, Addr{250}},
			result:      []Range{},
		},

		{
			description: "contained same first",
			a:           unsafeParsePrefix("10.224.24.0/22").Range(),
			b:           unsafeParsePrefix("10.224.24.0/24").Range(),
			result: []Range{
				validRange(t, unsafeParseAddr("10.224.25.0"), unsafeParseAddr("10.224.27.255")),
			},
		},
		{
			description: "same range",
			a:           unsafeParsePrefix("10.224.24.0/22").Range(),
			b:           unsafeParsePrefix("10.224.24.0/22").Range(),
			result:      []Range{},
		},
		{
			description: "larger same first",
			a:           unsafeParsePrefix("10.224.24.0/24").Range(),
			b:           unsafeParsePrefix("10.224.24.0/22").Range(),
			result:      []Range{},
		},

		{
			description: "wholly contained",
			a:           Range{Addr{100}, Addr{200}},
			b:           Range{Addr{110}, Addr{190}},
			result: []Range{
				Range{Addr{100}, Addr{109}},
				Range{Addr{191}, Addr{200}},
			},
		},
		{
			description: "contained same last",
			a:           unsafeParsePrefix("10.224.24.0/22").Range(),
			b:           unsafeParsePrefix("10.224.27.0/24").Range(),
			result: []Range{
				validRange(t, unsafeParseAddr("10.224.24.0"), unsafeParseAddr("10.224.26.255")),
			},
		},
		{
			description: "overlap left",
			a:           Range{Addr{100}, Addr{200}},
			b:           Range{Addr{150}, Addr{250}},
			result: []Range{
				Range{Addr{100}, Addr{149}},
			},
		},

		{
			description: "disjoint right",
			a:           unsafeParsePrefix("10.224.24.0/22").Range(),
			b:           unsafeParsePrefix("10.224.200.0/22").Range(),
			result: []Range{
				unsafeParsePrefix("10.224.24.0/22").Range(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			result := tt.a.Minus(tt.b)

			// A trick to compare the results as arrays
			var expected, actual [2]Range
			copy(expected[:], tt.result)
			copy(actual[:], result)
			assert.Equal(t, len(tt.result), len(result))
			assert.Equal(t, expected, actual)
		})
	}
}
