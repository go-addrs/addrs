package ipv6

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContains(t *testing.T) {
	tests := []struct {
		desc           string
		a, b           Prefix
		matches, exact bool
		common         uint32
		child          int
	}{
		{
			desc:    "trivial",
			a:       Prefix{Address{uint128{}}, 0},
			b:       Prefix{Address{uint128{}}, 0},
			matches: true,
			exact:   true,
			common:  0,
		},
		{
			desc:    "exact",
			a:       Prefix{_a("2001::"), 32},
			b:       Prefix{_a("2001::"), 32},
			matches: true,
			exact:   true,
			common:  32,
		},
		{
			desc:    "exact partial",
			a:       Prefix{_a("0a00:1f00::"), 39},
			b:       Prefix{_a("0a00:1f00:00F0::"), 39},
			matches: true,
			exact:   true,
			common:  39,
		},
		{
			desc:    "empty prefix match",
			a:       Prefix{Address{uint128{}}, 0},
			b:       Prefix{_a("2001:10::"), 32},
			matches: true,
			exact:   false,
			common:  0,
			child:   0,
		},
		{
			desc:    "empty prefix match backwards",
			a:       Prefix{Address{uint128{}}, 0},
			b:       Prefix{_a("F030:10::"), 32},
			matches: true,
			exact:   false,
			common:  0,
			child:   1,
		},
		{
			desc:    "matches",
			a:       Prefix{_a("2001::"), 16},
			b:       Prefix{_a("2001:10::"), 32},
			matches: true,
			exact:   false,
			common:  16,
			child:   0,
		},
		{
			desc:    "matches partial",
			a:       Prefix{_a("2001:2000::"), 17},
			b:       Prefix{_a("2001:2190::"), 32},
			matches: true,
			exact:   false,
			common:  17,
			child:   0,
		},
		{
			desc:    "matches backwards",
			a:       Prefix{_a("A0::"), 16},
			b:       Prefix{_a("A0:c800::"), 32},
			matches: true,
			exact:   false,
			common:  16,
			child:   1,
		},
		{
			desc:    "matches backwards partial",
			a:       Prefix{_a("10:f000::"), 17},
			b:       Prefix{_a("10:c800::"), 32},
			matches: true,
			exact:   false,
			common:  17,
			child:   1,
		},
		{
			desc:    "disjoint",
			a:       Prefix{Address{uint128{}}, 1},
			b:       Prefix{_a("8000::"), 1},
			matches: false,
			common:  0,
			child:   1,
		},
		{
			desc:    "disjoint longer",
			a:       Prefix{_a("::"), 65},
			b:       Prefix{_a("::8000:0:0:0"), 65},
			matches: false,
			common:  64,
			child:   1,
		},
		{
			desc:    "disjoint longer partial",
			a:       Prefix{_a("::"), 65},
			b:       Prefix{_a("0:0:0:1::"), 65},
			matches: false,
			common:  63,
			child:   1,
		},
		{
			desc:    "disjoint backwards",
			a:       Prefix{_a("8000::"), 1},
			b:       Prefix{Address{uint128{}}, 1},
			matches: false,
			common:  0,
			child:   0,
		},
		{
			desc:    "disjoint backwards longer",
			a:       Prefix{_a("::8000:0:0:0"), 71},
			b:       Prefix{_a("::"), 71},
			matches: false,
			common:  64,
			child:   0,
		},
		{
			desc:    "disjoint backwards longer partial",
			a:       Prefix{_a("0:0:0:1::"), 71},
			b:       Prefix{_a("::"), 71},
			matches: false,
			common:  63,
			child:   0,
		},
		{
			desc:    "disjoint with common",
			a:       Prefix{_a("A0::"), 32},
			b:       Prefix{_a("A0:A0::"), 32},
			matches: false,
			common:  24,
			child:   1,
		},
		{
			desc:    "disjoint with more disjoint bytes",
			a:       Prefix{_a("0:0:ffff:ffff:ffff:ffff:0:0"), 112},
			b:       Prefix{_a("8000::"), 112},
			matches: false,
			common:  0,
			child:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			matches, exact, common, child := contains(tt.a, tt.b)
			assert.Equal(t, tt.matches, matches)
			assert.Equal(t, tt.exact, exact)
			assert.Equal(t, tt.common, common)
			assert.Equal(t, tt.child, child)
		})
	}
}
