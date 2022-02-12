package ipv4

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func _r(first, last Address) Range {
	r, empty := NewRange(first, last)
	if empty {
		panic("only use this is non-empty cases")
	}
	return r
}

func TestRangeComparable(t *testing.T) {
	tests := []struct {
		description string
		a, b        Range
		equal       bool
	}{
		{
			description: "equal",
			a:           _r(_a("10.0.0.0"), _a("10.1.0.0")),
			b:           _r(_a("10.0.0.0"), _a("10.1.0.0")),
			equal:       true,
		}, {
			description: "first not equal",
			a:           _r(_a("10.0.0.0"), _a("10.1.0.0")),
			b:           _r(_a("10.0.0.1"), _a("10.1.0.0")),
			equal:       false,
		}, {
			description: "last not equal",
			a:           _r(_a("10.0.0.0"), _a("10.0.1.0")),
			b:           _r(_a("10.0.0.0"), _a("10.1.0.0")),
			equal:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			assert.Equal(t, tt.equal, tt.a == tt.b)
			assert.Equal(t, !tt.equal, tt.a != tt.b)
		})
	}
}

func TestNewRange(t *testing.T) {
	rangeEmpty := func(first, last Address) bool {
		_, empty := NewRange(first, last)
		return empty
	}

	assert.False(t, rangeEmpty(Address{100}, Address{200}))
	assert.False(t, rangeEmpty(Address{100}, Address{100}))
	assert.True(t, rangeEmpty(Address{200}, Address{100}))
	assert.True(t, rangeEmpty(Address{200}, Address{199}))
	assert.True(t, rangeEmpty(Address{0xffffffff}, Address{0}))
}

func TestRangeString(t *testing.T) {
	assert.Equal(t, "[18.52.86.120,35.69.103.137]", _r(Address{ui: 0x12345678}, Address{ui: 0x23456789}).String())
	assert.Equal(t, "[10.224.24.0,10.224.24.255]", _p("10.224.24.1/24").Range().String())
}

func TestRangeSize(t *testing.T) {
	assert.Equal(t, int64(0x100), _p("10.224.24.1/24").Range().Size())
	assert.Equal(t, int64(0x20000), _p("10.224.24.1/15").Range().Size())
	assert.Equal(t, int64(0x11111112), _r(Address{ui: 0x12345678}, Address{ui: 0x23456789}).Size())
}

func TestRangeFirstLast(t *testing.T) {
	tests := []struct {
		description string
		r           Range
		first, last Address
	}{
		{
			description: "unaligned",
			r:           _r(Address{ui: 0x12345678}, Address{ui: 0x23456789}),
			first:       Address{ui: 0x12345678},
			last:        Address{ui: 0x23456789},
		},
		{
			description: "prefix",
			r:           _p("10.224.24.1/24").Range(),
			first:       _a("10.224.24.0"),
			last:        _a("10.224.24.255"),
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
			a:           _p("10.224.24.1/22").Range(),
			b:           _p("10.224.26.1/24").Range(),
		},
		{
			description: "unaligned",
			a:           _r(Address{ui: 0x12345678}, Address{ui: 0x23456789}),
			b:           _p("20.224.26.1/24").Range(),
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
		backwards   []Range
	}{
		{
			description: "disjoint left",
			a:           _p("10.224.24.0/22").Range(),
			b:           _p("10.224.0.0/22").Range(),
			result: []Range{
				_p("10.224.24.0/22").Range(),
			},
			backwards: []Range{
				_p("10.224.0.0/22").Range(),
			},
		},
		{
			description: "overlap right",
			a:           Range{Address{100}, Address{200}},
			b:           Range{Address{50}, Address{150}},
			result: []Range{
				Range{Address{151}, Address{200}},
			},
			backwards: []Range{
				Range{Address{50}, Address{99}},
			},
		},
		{
			description: "larger same last",
			a:           _p("10.224.27.0/24").Range(),
			b:           _p("10.224.24.0/22").Range(),
			result:      []Range{},
			backwards: []Range{
				_r(_a("10.224.24.0"), _a("10.224.26.255")),
			},
		},
		{
			description: "overlap all",
			a:           Range{Address{100}, Address{200}},
			b:           Range{Address{50}, Address{250}},
			result:      []Range{},
			backwards: []Range{
				Range{Address{50}, Address{99}},
				Range{Address{201}, Address{250}},
			},
		},

		{
			description: "contained same first",
			a:           _p("10.224.24.0/22").Range(),
			b:           _p("10.224.24.0/24").Range(),
			result: []Range{
				_r(_a("10.224.25.0"), _a("10.224.27.255")),
			},
			backwards: []Range{},
		},
		{
			description: "same range",
			a:           _p("10.224.24.0/22").Range(),
			b:           _p("10.224.24.0/22").Range(),
			result:      []Range{},
			backwards:   []Range{},
		},
		{
			description: "wholly contained",
			a:           Range{Address{100}, Address{200}},
			b:           Range{Address{110}, Address{190}},
			result: []Range{
				Range{Address{100}, Address{109}},
				Range{Address{191}, Address{200}},
			},
			backwards: []Range{},
		},
		{
			description: "contained same last",
			a:           _p("10.224.24.0/22").Range(),
			b:           _p("10.224.27.0/24").Range(),
			result: []Range{
				_r(_a("10.224.24.0"), _a("10.224.26.255")),
			},
			backwards: []Range{},
		},
		{
			description: "overlap left",
			a:           Range{Address{100}, Address{200}},
			b:           Range{Address{150}, Address{250}},
			result: []Range{
				Range{Address{100}, Address{149}},
			},
			backwards: []Range{
				Range{Address{201}, Address{250}},
			},
		},
		{
			description: "first equals last",
			a:           Range{Address{100}, Address{200}},
			b:           Range{Address{200}, Address{250}},
			result: []Range{
				Range{Address{100}, Address{199}},
			},
			backwards: []Range{
				Range{Address{201}, Address{250}},
			},
		},
		{
			description: "first + 1 equals last",
			a:           Range{Address{100}, Address{200}},
			b:           Range{Address{199}, Address{250}},
			result: []Range{
				Range{Address{100}, Address{198}},
			},
			backwards: []Range{
				Range{Address{201}, Address{250}},
			},
		},
		{
			description: "first equals last + 1",
			a:           Range{Address{100}, Address{200}},
			b:           Range{Address{201}, Address{250}},
			result: []Range{
				Range{Address{100}, Address{200}},
			},
			backwards: []Range{
				Range{Address{201}, Address{250}},
			},
		},
		{
			description: "disjoint right",
			a:           _p("10.224.24.0/22").Range(),
			b:           _p("10.224.200.0/22").Range(),
			result: []Range{
				_p("10.224.24.0/22").Range(),
			},
			backwards: []Range{
				_p("10.224.200.0/22").Range(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			run := func(a, b Range, r []Range) {
				result := a.Minus(b)

				// A trick to compare the results as arrays
				var expected, actual [2]Range
				copy(expected[:], r)
				copy(actual[:], result)
				assert.Equal(t, len(r), len(result))
				assert.Equal(t, expected, actual)
			}
			t.Run("forward", func(t *testing.T) { run(tt.a, tt.b, tt.result) })
			t.Run("forward", func(t *testing.T) { run(tt.b, tt.a, tt.backwards) })
		})
	}
}

func TestRangeSet(t *testing.T) {
	r := Range{_a("7.0.37.17"), _a("13.8.222.113")}

	// I calculated this manually from the above arbitrarily chosen range.
	golden := NewSet_()
	golden.Insert(_p("7.0.37.17/32"))
	golden.Insert(_p("7.0.37.18/31"))
	golden.Insert(_p("7.0.37.20/30"))
	golden.Insert(_p("7.0.37.24/29"))
	golden.Insert(_p("7.0.37.32/27"))
	golden.Insert(_p("7.0.37.64/26"))
	golden.Insert(_p("7.0.37.128/25"))
	golden.Insert(_p("7.0.38.0/23"))
	golden.Insert(_p("7.0.40.0/21"))
	golden.Insert(_p("7.0.48.0/20"))
	golden.Insert(_p("7.0.64.0/18"))
	golden.Insert(_p("7.0.128.0/17"))
	golden.Insert(_p("7.1.0.0/16"))
	golden.Insert(_p("7.2.0.0/15"))
	golden.Insert(_p("7.4.0.0/14"))
	golden.Insert(_p("7.8.0.0/13"))
	golden.Insert(_p("7.16.0.0/12"))
	golden.Insert(_p("7.32.0.0/11"))
	golden.Insert(_p("7.64.0.0/10"))
	golden.Insert(_p("7.128.0.0/9"))
	golden.Insert(_p("8.0.0.0/6"))
	golden.Insert(_p("12.0.0.0/8"))
	golden.Insert(_p("13.0.0.0/13"))
	golden.Insert(_p("13.8.0.0/17"))
	golden.Insert(_p("13.8.128.0/18"))
	golden.Insert(_p("13.8.192.0/20"))
	golden.Insert(_p("13.8.208.0/21"))
	golden.Insert(_p("13.8.216.0/22"))
	golden.Insert(_p("13.8.220.0/23"))
	golden.Insert(_p("13.8.222.0/26"))
	golden.Insert(_p("13.8.222.64/27"))
	golden.Insert(_p("13.8.222.96/28"))
	golden.Insert(_p("13.8.222.112/31"))

	assert.True(t, golden.Equal(r.Set()))
}

func TestRangePlus(t *testing.T) {
	tests := []struct {
		description string
		a, b        Range
		result      []Range
	}{
		{
			description: "disjoint",
			a:           _p("10.224.24.0/22").Range(),
			b:           _p("10.224.0.0/22").Range(),
			result: []Range{
				_p("10.224.0.0/22").Range(),
				_p("10.224.24.0/22").Range(),
			},
		}, {
			description: "adjacent",
			a:           _p("10.224.4.0/22").Range(),
			b:           _p("10.224.0.0/22").Range(),
			result: []Range{
				_p("10.224.0.0/21").Range(),
			},
		}, {
			description: "containing prefix",
			a:           _p("10.224.4.0/22").Range(),
			b:           _p("10.224.0.0/21").Range(),
			result: []Range{
				_p("10.224.0.0/21").Range(),
			},
		}, {
			description: "same",
			a:           _p("10.224.0.0/21").Range(),
			b:           _p("10.224.0.0/21").Range(),
			result: []Range{
				_p("10.224.0.0/21").Range(),
			},
		}, {
			description: "subset",
			a:           _r(_a("10.224.0.99"), _a("10.224.0.247")),
			b:           _r(_a("10.224.0.90"), _a("10.224.0.248")),
			result: []Range{
				_r(_a("10.224.0.90"), _a("10.224.0.248")),
			},
		}, {
			description: "overlapping",
			a:           _r(_a("10.224.0.90"), _a("10.224.0.247")),
			b:           _r(_a("10.224.0.99"), _a("10.224.0.248")),
			result: []Range{
				_r(_a("10.224.0.90"), _a("10.224.0.248")),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			run := func(first, second Range) {
				result := first.Plus(second)

				// A trick to compare the results as arrays
				var expected, actual [2]Range
				copy(expected[:], tt.result)
				copy(actual[:], result)
				assert.Equal(t, len(tt.result), len(result))
				assert.Equal(t, expected, actual)
			}
			t.Run("forward", func(t *testing.T) { run(tt.a, tt.b) })
			t.Run("backward", func(t *testing.T) { run(tt.b, tt.a) })
		})
	}
}

func TestRangeAsMapKey(t *testing.T) {
	m := make(map[Range]bool)

	m[_r(_a("203.0.113.0"), _a("203.0.113.127"))] = true

	assert.True(t, m[_r(_a("203.0.113.0"), _a("203.0.113.127"))])
}
