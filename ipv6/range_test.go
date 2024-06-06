package ipv6

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func _r(first, last Address) Range {
	r, empty := RangeFromAddresses(first, last)
	if empty {
		panic("only use this in non-empty cases")
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
			a:           _r(_a("2001::"), _a("2001::1000:0")),
			b:           _r(_a("2001::"), _a("2001::1000:0")),
			equal:       true,
		}, {
			description: "first not equal",
			a:           _r(_a("2001::"), _a("2001::1000:0")),
			b:           _r(_a("2001::1"), _a("2001::1000:0")),
			equal:       false,
		}, {
			description: "last not equal",
			a:           _r(_a("2001::"), _a("2001::1000:0")),
			b:           _r(_a("2001::"), _a("2001::1:0")),
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

func TestRangeFromAddresses(t *testing.T) {
	rangeEmpty := func(first, last Address) bool {
		_, empty := RangeFromAddresses(first, last)
		return empty
	}

	assert.False(t, rangeEmpty(Address{uint128{100, 0}}, Address{uint128{200, 0}}))
	assert.False(t, rangeEmpty(Address{uint128{100, 0}}, Address{uint128{100, 0}}))
	assert.True(t, rangeEmpty(Address{uint128{200, 0}}, Address{uint128{100, 0}}))
	assert.True(t, rangeEmpty(Address{uint128{200, 0}}, Address{uint128{199, 0}}))
	assert.True(t, rangeEmpty(Address{uint128{0xffffffff, 0}}, Address{uint128{0, 0}}))
}

func TestRangeString(t *testing.T) {
	assert.Equal(t, "[2001:db8:85a3::8a2e:370:7334,2001:db8:9621::1234:c28:1]", _r(Address{ui: uint128{0x20010db885a30000, 0x8a2e03707334}}, Address{ui: uint128{0x20010db896210000, 0x12340c280001}}).String())
	assert.Equal(t, "[2001:db8:85a3:1234:abcd:8a2e::,2001:db8:85a3:1234:abcd:8a2e:ffff:ffff]", _p("2001:db8:85a3:1234:abcd:8a2e:1:1/96").Range().String())
}

func TestRangeFirstLast(t *testing.T) {
	tests := []struct {
		description string
		r           Range
		first, last Address
	}{
		{
			description: "unaligned",
			r:           _r(Address{ui: uint128{0x12345678, 0x0}}, Address{ui: uint128{0x23456789, 0x0}}),
			first:       Address{ui: uint128{0x12345678, 0x0}},
			last:        Address{ui: uint128{0x23456789, 0x0}},
		},
		{
			description: "prefix",
			r:           _p("2001::1/96").Range(),
			first:       _a("2001::"),
			last:        _a("2001::ffff:ffff"),
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
			a:           _p("::ffff:10.224.24.1/118").Range(),
			b:           _p("::ffff:10.224.26.1/120").Range(),
		},
		{
			description: "unaligned",
			a:           _r(Address{ui: uint128{low: 0xffff12345678}}, Address{ui: uint128{low: 0xffff23456789}}),
			b:           _p("::ffff:20.224.26.1/120").Range(),
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
			a:           _p("2001::1:0:0/112").Range(),
			b:           _p("2001::/112").Range(),
			result: []Range{
				_p("2001::1:0:0/112").Range(),
			},
			backwards: []Range{
				_p("2001::/112").Range(),
			},
		},
		{
			description: "overlap right",
			a:           Range{Address{uint128{0, 100}}, Address{uint128{0, 200}}},
			b:           Range{Address{uint128{0, 50}}, Address{uint128{0, 150}}},
			result: []Range{
				Range{Address{uint128{0, 151}}, Address{uint128{0, 200}}},
			},
			backwards: []Range{
				Range{Address{uint128{0, 50}}, Address{uint128{0, 99}}},
			},
		},
		{
			description: "larger same last",
			a:           _p("2001::ff:0/112").Range(),
			b:           _p("2001::fd:0/108").Range(),
			result:      []Range{},
			backwards: []Range{
				_r(_a("2001::f0:0"), _a("2001::fe:ffff")),
			},
		},
		{
			description: "overlap all",
			a:           Range{Address{uint128{0, 100}}, Address{uint128{0, 200}}},
			b:           Range{Address{uint128{0, 50}}, Address{uint128{0, 250}}},
			result:      []Range{},
			backwards: []Range{
				Range{Address{uint128{0, 50}}, Address{uint128{0, 99}}},
				Range{Address{uint128{0, 201}}, Address{uint128{0, 250}}},
			},
		},
		{
			description: "contained same first",
			a:           _p("2001::fc:0/110").Range(),
			b:           _p("2001::fc:0/112").Range(),
			result: []Range{
				_r(_a("2001::fd:0"), _a("2001::ff:ffff")),
			},
			backwards: []Range{},
		},
		{
			description: "same range",
			a:           _p("2001::fd:0/112").Range(),
			b:           _p("2001::fd:0/112").Range(),
			result:      []Range{},
			backwards:   []Range{},
		},
		{
			description: "wholly contained",
			a:           Range{Address{uint128{0, 100}}, Address{uint128{0, 200}}},
			b:           Range{Address{uint128{0, 110}}, Address{uint128{0, 190}}},
			result: []Range{
				Range{Address{uint128{0, 100}}, Address{uint128{0, 109}}},
				Range{Address{uint128{0, 191}}, Address{uint128{0, 200}}},
			},
			backwards: []Range{},
		},
		{
			description: "contained same last",
			a:           _p("2001::fb:0/108").Range(),
			b:           _p("2001::ff:0/112").Range(),
			result: []Range{
				_r(_a("2001::f0:0"), _a("2001::fe:ffff")),
			},
			backwards: []Range{},
		},
		{
			description: "overlap left",
			a:           Range{Address{uint128{0, 100}}, Address{uint128{0, 200}}},
			b:           Range{Address{uint128{0, 150}}, Address{uint128{0, 250}}},
			result: []Range{
				Range{Address{uint128{0, 100}}, Address{uint128{0, 149}}},
			},
			backwards: []Range{
				Range{Address{uint128{0, 201}}, Address{uint128{0, 250}}},
			},
		},
		{
			description: "first equals last",
			a:           Range{Address{uint128{0, 100}}, Address{uint128{0, 200}}},
			b:           Range{Address{uint128{0, 200}}, Address{uint128{0, 250}}},
			result: []Range{
				Range{Address{uint128{0, 100}}, Address{uint128{0, 199}}},
			},
			backwards: []Range{
				Range{Address{uint128{0, 201}}, Address{uint128{0, 250}}},
			},
		},
		{
			description: "first + 1 equals last",
			a:           Range{Address{uint128{0, 100}}, Address{uint128{0, 200}}},
			b:           Range{Address{uint128{0, 199}}, Address{uint128{0, 250}}},
			result: []Range{
				Range{Address{uint128{0, 100}}, Address{uint128{0, 198}}},
			},
			backwards: []Range{
				Range{Address{uint128{0, 201}}, Address{uint128{0, 250}}},
			},
		},
		{
			description: "first equals last + 1",
			a:           Range{Address{uint128{0, 100}}, Address{uint128{0, 200}}},
			b:           Range{Address{uint128{0, 201}}, Address{uint128{0, 250}}},
			result: []Range{
				Range{Address{uint128{0, 100}}, Address{uint128{0, 200}}},
			},
			backwards: []Range{
				Range{Address{uint128{0, 201}}, Address{uint128{0, 250}}},
			},
		},
		{
			description: "disjoint right",
			a:           _p("2001::fd:0/112").Range(),
			b:           _p("2001::f00:0/112").Range(),
			result: []Range{
				_p("2001::fd:0/112").Range(),
			},
			backwards: []Range{
				_p("2001::f00:0/112").Range(),
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
			t.Run("backwards", func(t *testing.T) { run(tt.b, tt.a, tt.backwards) })
		})
	}
}

func TestRangeSet(t *testing.T) {
	r := Range{_a("2001::0241:0122"), _a("2001::ad23:824")}

	// I calculated this manually from the above arbitrarily chosen range.
	golden := NewSet_()
	golden.Insert(_p("2001::241:122/127"))
	golden.Insert(_p("2001::241:124/126"))
	golden.Insert(_p("2001::241:128/125"))
	golden.Insert(_p("2001::241:130/124"))
	golden.Insert(_p("2001::241:140/122"))
	golden.Insert(_p("2001::241:180/121"))
	golden.Insert(_p("2001::241:200/119"))
	golden.Insert(_p("2001::241:400/118"))
	golden.Insert(_p("2001::241:800/117"))
	golden.Insert(_p("2001::241:1000/116"))
	golden.Insert(_p("2001::241:2000/115"))
	golden.Insert(_p("2001::241:4000/114"))
	golden.Insert(_p("2001::241:8000/113"))
	golden.Insert(_p("2001::242:0/111"))
	golden.Insert(_p("2001::244:0/110"))
	golden.Insert(_p("2001::248:0/109"))
	golden.Insert(_p("2001::250:0/108"))
	golden.Insert(_p("2001::260:0/107"))
	golden.Insert(_p("2001::280:0/105"))
	golden.Insert(_p("2001::300:0/104"))
	golden.Insert(_p("2001::400:0/102"))
	golden.Insert(_p("2001::800:0/101"))
	golden.Insert(_p("2001::1000:0/100"))
	golden.Insert(_p("2001::2000:0/99"))
	golden.Insert(_p("2001::4000:0/98"))
	golden.Insert(_p("2001::8000:0/99"))
	golden.Insert(_p("2001::a000:0/101"))
	golden.Insert(_p("2001::a800:0/102"))
	golden.Insert(_p("2001::ac00:0/104"))
	golden.Insert(_p("2001::ad00:0/107"))
	golden.Insert(_p("2001::ad20:0/111"))
	golden.Insert(_p("2001::ad22:0/112"))
	golden.Insert(_p("2001::ad23:0/117"))
	golden.Insert(_p("2001::ad23:800/123"))
	golden.Insert(_p("2001::ad23:820/126"))
	golden.Insert(_p("2001::ad23:824/128"))

	assert.True(t, golden.Set().Equal(r.Set()))
}

func TestRangePlus(t *testing.T) {
	tests := []struct {
		description string
		a, b        Range
		result      []Range
	}{
		{
			description: "disjoint",
			a:           _p("2001::1:0:0/112").Range(),
			b:           _p("2001::/112").Range(),
			result: []Range{
				_p("2001::/112").Range(),
				_p("2001::1:0:0/112").Range(),
			},
		}, {
			description: "adjacent",
			a:           _p("2001::1:0/112").Range(),
			b:           _p("2001::/112").Range(),
			result: []Range{
				_p("2001::/111").Range(),
			},
		}, {
			description: "containing prefix",
			a:           _p("2001::1:0/112").Range(),
			b:           _p("2001::/111").Range(),
			result: []Range{
				_p("2001::/111").Range(),
			},
		}, {
			description: "same",
			a:           _p("2001::/16").Range(),
			b:           _p("2001::/16").Range(),
			result: []Range{
				_p("2001::/16").Range(),
			},
		}, {
			description: "subset",
			a:           _r(_a("2001::f"), _a("2001::fe")),
			b:           _r(_a("2001::"), _a("2001::ff")),
			result: []Range{
				_r(_a("2001::"), _a("2001::ff")),
			},
		}, {
			description: "overlapping",
			a:           _r(_a("2001::"), _a("2001::fe")),
			b:           _r(_a("2001::f"), _a("2001::ff")),
			result: []Range{
				_r(_a("2001::"), _a("2001::ff")),
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

	m[_r(_a("2001::"), _a("2001::1000:0"))] = true

	assert.True(t, m[_r(_a("2001::"), _a("2001::1000:0"))])
}

func TestRangeNumPrefixes(t *testing.T) {
	tests := []struct {
		description string
		r           Range
		length      uint32
		count       uint64
		error       bool
	}{
		{
			description: "simple",
			r:           _r(_a("2001::"), _a("2002::")),
			length:      24,
			count:       256,
		}, {
			description: "bonus",
			r:           _r(_a("2001::"), _a("2001::1:0")),
			length:      128,
			count:       0x10001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			count, err := tt.r.NumPrefixes(tt.length)
			assert.Equal(t, tt.error, err != nil)
			assert.Equal(t, tt.count, count)
		})
	}
}
