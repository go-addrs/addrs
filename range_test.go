package ipv4

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func validRange(t *testing.T, first, last Address) Range {
	r, err := NewRange(first, last)
	assert.Nil(t, err)
	return r
}

func TestNewRange(t *testing.T) {
	rangeError := func(first, last Address) error {
		_, err := NewRange(first, last)
		return err
	}

	assert.Nil(t, rangeError(Address{100}, Address{200}))
	assert.Nil(t, rangeError(Address{100}, Address{100}))
	assert.NotNil(t, rangeError(Address{200}, Address{100}))
	assert.NotNil(t, rangeError(Address{200}, Address{199}))
	assert.NotNil(t, rangeError(Address{0xffffffff}, Address{0}))
}

func TestRangeString(t *testing.T) {
	assert.Equal(t, "[18.52.86.120,35.69.103.137]", validRange(t, Address{ui: 0x12345678}, Address{ui: 0x23456789}).String())
	assert.Equal(t, "[10.224.24.0,10.224.24.255]", unsafeParsePrefix("10.224.24.1/24").Range().String())
}

func TestRangeSize(t *testing.T) {
	assert.Equal(t, 0x100, unsafeParsePrefix("10.224.24.1/24").Range().Size())
	assert.Equal(t, 0x20000, unsafeParsePrefix("10.224.24.1/15").Range().Size())
	assert.Equal(t, 0x11111112, validRange(t, Address{ui: 0x12345678}, Address{ui: 0x23456789}).Size())
}

func TestRangeFirstLast(t *testing.T) {
	tests := []struct {
		description string
		r           Range
		first, last Address
	}{
		{
			description: "unaligned",
			r:           validRange(t, Address{ui: 0x12345678}, Address{ui: 0x23456789}),
			first:       Address{ui: 0x12345678},
			last:        Address{ui: 0x23456789},
		},
		{
			description: "prefix",
			r:           unsafeParsePrefix("10.224.24.1/24").Range(),
			first:       unsafeParseAddress("10.224.24.0"),
			last:        unsafeParseAddress("10.224.24.255"),
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
			a:           validRange(t, Address{ui: 0x12345678}, Address{ui: 0x23456789}),
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
			a:           Range{Address{100}, Address{200}},
			b:           Range{Address{50}, Address{150}},
			result: []Range{
				Range{Address{151}, Address{200}},
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
			a:           Range{Address{100}, Address{200}},
			b:           Range{Address{50}, Address{250}},
			result:      []Range{},
		},

		{
			description: "contained same first",
			a:           unsafeParsePrefix("10.224.24.0/22").Range(),
			b:           unsafeParsePrefix("10.224.24.0/24").Range(),
			result: []Range{
				validRange(t, unsafeParseAddress("10.224.25.0"), unsafeParseAddress("10.224.27.255")),
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
			a:           Range{Address{100}, Address{200}},
			b:           Range{Address{110}, Address{190}},
			result: []Range{
				Range{Address{100}, Address{109}},
				Range{Address{191}, Address{200}},
			},
		},
		{
			description: "contained same last",
			a:           unsafeParsePrefix("10.224.24.0/22").Range(),
			b:           unsafeParsePrefix("10.224.27.0/24").Range(),
			result: []Range{
				validRange(t, unsafeParseAddress("10.224.24.0"), unsafeParseAddress("10.224.26.255")),
			},
		},
		{
			description: "overlap left",
			a:           Range{Address{100}, Address{200}},
			b:           Range{Address{150}, Address{250}},
			result: []Range{
				Range{Address{100}, Address{149}},
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

func TestRangeSet(t *testing.T) {
	r := Range{unsafeParseAddress("7.0.37.17"), unsafeParseAddress("13.8.222.113")}

	// I calculated this manually from the above arbitrarily chosen range.
	golden := SetBuilder{}
	golden.InsertPrefix(unsafeParsePrefix("7.0.37.17/32"))
	golden.InsertPrefix(unsafeParsePrefix("7.0.37.18/31"))
	golden.InsertPrefix(unsafeParsePrefix("7.0.37.20/30"))
	golden.InsertPrefix(unsafeParsePrefix("7.0.37.24/29"))
	golden.InsertPrefix(unsafeParsePrefix("7.0.37.32/27"))
	golden.InsertPrefix(unsafeParsePrefix("7.0.37.64/26"))
	golden.InsertPrefix(unsafeParsePrefix("7.0.37.128/25"))
	golden.InsertPrefix(unsafeParsePrefix("7.0.38.0/23"))
	golden.InsertPrefix(unsafeParsePrefix("7.0.40.0/21"))
	golden.InsertPrefix(unsafeParsePrefix("7.0.48.0/20"))
	golden.InsertPrefix(unsafeParsePrefix("7.0.64.0/18"))
	golden.InsertPrefix(unsafeParsePrefix("7.0.128.0/17"))
	golden.InsertPrefix(unsafeParsePrefix("7.1.0.0/16"))
	golden.InsertPrefix(unsafeParsePrefix("7.2.0.0/15"))
	golden.InsertPrefix(unsafeParsePrefix("7.4.0.0/14"))
	golden.InsertPrefix(unsafeParsePrefix("7.8.0.0/13"))
	golden.InsertPrefix(unsafeParsePrefix("7.16.0.0/12"))
	golden.InsertPrefix(unsafeParsePrefix("7.32.0.0/11"))
	golden.InsertPrefix(unsafeParsePrefix("7.64.0.0/10"))
	golden.InsertPrefix(unsafeParsePrefix("7.128.0.0/9"))
	golden.InsertPrefix(unsafeParsePrefix("8.0.0.0/6"))
	golden.InsertPrefix(unsafeParsePrefix("12.0.0.0/8"))
	golden.InsertPrefix(unsafeParsePrefix("13.0.0.0/13"))
	golden.InsertPrefix(unsafeParsePrefix("13.8.0.0/17"))
	golden.InsertPrefix(unsafeParsePrefix("13.8.128.0/18"))
	golden.InsertPrefix(unsafeParsePrefix("13.8.192.0/20"))
	golden.InsertPrefix(unsafeParsePrefix("13.8.208.0/21"))
	golden.InsertPrefix(unsafeParsePrefix("13.8.216.0/22"))
	golden.InsertPrefix(unsafeParsePrefix("13.8.220.0/23"))
	golden.InsertPrefix(unsafeParsePrefix("13.8.222.0/26"))
	golden.InsertPrefix(unsafeParsePrefix("13.8.222.64/27"))
	golden.InsertPrefix(unsafeParsePrefix("13.8.222.96/28"))
	golden.InsertPrefix(unsafeParsePrefix("13.8.222.112/31"))

	assert.True(t, golden.Set().Equal(r.Set()))
}
