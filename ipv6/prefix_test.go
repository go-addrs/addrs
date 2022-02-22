package ipv6

import (
	"fmt"
	"net"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func _p(cidr string) Prefix {
	prefix, err := ParsePrefix(cidr)
	if err != nil {
		panic("only use this is happy cases")
	}
	return prefix
}

func unsafePrefixFromUint64(high, low uint64, length int) Prefix {
	mask, err := MaskFromLength(length)
	if err != nil {
		panic("only use this is happy cases")
	}
	return PrefixFromAddressMask(Address{uint128{high, low}}, mask)
}

func unsafeParseNet(prefix string) *net.IPNet {
	ipNet, err := parseNet(prefix)
	if err != nil {
		panic("only use this is happy cases")
	}
	return ipNet
}

func TestPrefixComparable(t *testing.T) {
	tests := []struct {
		description string
		a, b        Prefix
		equal       bool
	}{
		{
			description: "equal",
			a:           _p("2001::1/64"),
			b:           _p("2001::1/64"),
			equal:       true,
		}, {
			description: "lengths not equal",
			a:           _p("2001::1/64"),
			b:           _p("2001::1/65"),
			equal:       false,
		}, {
			description: "host bits not equal",
			a:           _p("2001::1/64"),
			b:           _p("2001::2/64"),
			equal:       false,
		}, {
			description: "prefix bits not equal",
			a:           _p("2001::1/64"),
			b:           _p("2002::1/64"),
			equal:       false,
		}, {
			description: "extremes",
			a:           _p("::/0"),
			b:           _p("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff/128"),
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

func TestParsePrefix(t *testing.T) {
	tests := []struct {
		description string
		cidr        string
		expected    Prefix
		isErr       bool
	}{
		{
			description: "ipv4",
			cidr:        "10.224.24.1/27",
			isErr:       true,
		},
		{
			description: "ipv6 dotted quad",
			cidr:        "::ffff:203.0.113.17/128",
			expected:    unsafePrefixFromUint64(0x0, 0xffffcb007111, 128),
		},
		{
			description: "ipv6 standard",
			cidr:        "2001::1/64",
			expected:    unsafePrefixFromUint64(0x2001000000000000, 0x1, 64),
		},
		{
			description: "bogus",
			cidr:        "bogus",
			isErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			net, err := ParsePrefix(tt.cidr)
			if tt.isErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.expected, net)
			}
		})
	}
}

func TestPrefixFromNetIPNet(t *testing.T) {
	tests := []struct {
		description string
		net         *net.IPNet
		expected    Prefix
		isErr       bool
	}{
		{
			description: "nil",
			net:         nil,
			isErr:       true,
		},
		{
			description: "ipv4",
			net:         unsafeParseNet("10.224.24.1/22"),
			isErr:       true,
		},
		{
			description: "ipv6",
			net:         unsafeParseNet("2001::/56"),
			expected:    unsafePrefixFromUint64(0x2001000000000000, 0x0, 56),
		},
		{
			description: "mixed up IPv4/6 IPNet",
			net: &net.IPNet{
				IP:   unsafeParseNet("2001::/56").IP,
				Mask: unsafeParseNet("10.0.0.0/16").Mask,
			},
			isErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			net, err := PrefixFromNetIPNet(tt.net)
			if tt.isErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.expected, net)
			}
		})
	}
}

func TestPrefixEqual(t *testing.T) {
	first, second := unsafePrefixFromUint64(0x20010db885a30000, 0x8a2e03707334, 56), unsafePrefixFromUint64(0x20010db885a30000, 0x8a2e03707334, 56)
	assert.Equal(t, first, second)
	assert.True(t, first == second)
	assert.True(t, reflect.DeepEqual(first, second))

	third := unsafePrefixFromUint64(0x30010db885a30000, 0x8a2e03707334, 56)
	assert.NotEqual(t, third, second)
	assert.False(t, third == first)
	assert.False(t, reflect.DeepEqual(third, first))
}

func TestPrefixLessThan(t *testing.T) {
	prefixes := []Prefix{
		unsafePrefixFromUint64(0x0, 0x0, 0),
		unsafePrefixFromUint64(0x0, 0x0, 32),
		unsafePrefixFromUint64(0x0, 0x0, 64),
		unsafePrefixFromUint64(0x0, 0x0, 96),
		unsafePrefixFromUint64(0x0, 0x0, 112),
		unsafePrefixFromUint64(0x0, 0x0, 127),
		unsafePrefixFromUint64(0x0, 0x0, 128),
		unsafePrefixFromUint64(0x20010db885a30000, 0x0, 32),
		unsafePrefixFromUint64(0x20010db885a3ffff, 0x0, 32),
		unsafePrefixFromUint64(0x20010db885a31701, 0x0, 56),
		unsafePrefixFromUint64(0x20010db885a31702, 0x0, 56),
		unsafePrefixFromUint64(0x20010db885a317ff, 0x0, 56),
		unsafePrefixFromUint64(0x20010db885a31701, 0x0, 64),
		unsafePrefixFromUint64(0x20010db885a31701, 0x8a2e03707334, 96),
		unsafePrefixFromUint64(0x20010db885a31701, 0x8a2e03707335, 96),
		unsafePrefixFromUint64(0x20010db885a31701, 0x8a2e037073ff, 96),
		unsafePrefixFromUint64(0x20010db885a31701, 0x8a2e03707334, 128),
		unsafePrefixFromUint64(0x20010db885a31701, 0x8a2e03707335, 128),
		unsafePrefixFromUint64(0x20010db885a31701, 0x8a2e03707434, 120),
	}

	for a := 0; a < len(prefixes); a++ {
		for b := a; b < len(prefixes); b++ {
			t.Run(fmt.Sprintf("%d < %d", a, b), func(t *testing.T) {
				if a == b {
					assert.False(t, prefixes[a].lessThan(prefixes[b]))
				} else {
					assert.True(t, prefixes[a].lessThan(prefixes[b]))
				}
				assert.False(t, prefixes[b].lessThan(prefixes[a]))
			})
		}
	}
}

func TestNetworkHost(t *testing.T) {
	tests := []struct {
		description                     string
		prefix                          Prefix
		network, host, prefixUpperLimit Prefix
	}{
		{
			description:      "0",
			prefix:           _p("2001::1/0"),
			network:          _p("::/0"),
			host:             _p("2001::1/0"),
			prefixUpperLimit: _p("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff/0"),
		},
		{
			description:      "32",
			prefix:           _p("2001::1/32"),
			network:          _p("2001::0/32"),
			host:             _p("::1/32"),
			prefixUpperLimit: _p("2001::ffff:ffff:ffff:ffff:ffff:ffff/32"),
		},
		{
			description:      "64",
			prefix:           _p("2001::1/64"),
			network:          _p("2001::0/64"),
			host:             _p("::1/64"),
			prefixUpperLimit: _p("2001::ffff:ffff:ffff:ffff/64"),
		},
		{
			description:      "96",
			prefix:           _p("2001::1/96"),
			network:          _p("2001::0/96"),
			host:             _p("::1/96"),
			prefixUpperLimit: _p("2001::ffff:ffff/96"),
		},
		{
			description:      "128",
			prefix:           _p("2001::1/128"),
			network:          _p("2001::1/128"),
			host:             _p("::/128"),
			prefixUpperLimit: _p("2001::1/128"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			assert.Equal(t, tt.network, tt.prefix.Network())
			assert.Equal(t, tt.host, tt.prefix.Host())
			assert.Equal(t, tt.prefixUpperLimit, tt.prefix.prefixUpperLimit())
		})
	}
}

func TestPrefixToNetIPNet(t *testing.T) {
	assert.Equal(t, "2001:db8:85a3::8a2e:370:7334/64", _p("2001:db8:85a3::8a2e:370:7334/64").ToNetIPNet().String())
}

func TestPrefixString(t *testing.T) {
	cidrs := []string{
		"::/0",
		"2001::/32",
		"ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff/128",
		"2001:db8:85a3::8a2e:370:7334/64",
	}

	for _, cidr := range cidrs {
		t.Run(cidr, func(t *testing.T) {
			assert.Equal(t, cidr, _p(cidr).String())
		})
	}
}

func TestPrefixUint64(t *testing.T) {
	addressHigh, addressLow, maskHigh, maskLow := _p("2001:db8:85a3::8a2e:370:7334/80").Uint64()
	assert.Equal(t, uint64(0x20010db885a30000), addressHigh)
	assert.Equal(t, uint64(0x8a2e03707334), addressLow)
	assert.Equal(t, uint64(0xffffffffffffffff), maskHigh)
	assert.Equal(t, uint64(0xffff000000000000), maskLow)
}

func TestPrefixFromAddressMask(t *testing.T) {
	address := Address{ui: uint128{0x20010db885a30000, 0x8a2e03707334}}
	mask, _ := MaskFromLength(80)
	assert.Equal(t, Prefix{addr: address, length: 80}, PrefixFromAddressMask(address, mask))
}

func TestPrefixHalves(t *testing.T) {
	tests := []struct {
		prefix Prefix
		a, b   Prefix
	}{
		{
			prefix: _p("::/0"),
			a:      _p("::/1"),
			b:      _p("8000:0000::/1"),
		},
		{
			prefix: _p("2001:db8::/32"),
			a:      _p("2001:db8:0000::/33"),
			b:      _p("2001:db8:8000::/33"),
		},
		{
			prefix: _p("2001:db8:85a3:8a2e::/64"),
			a:      _p("2001:db8:85a3:8a2e::/65"),
			b:      _p("2001:db8:85a3:8a2e:8000::/65"),
		},
		{
			prefix: _p("2001:db8:85a3::8a2e:370:7334/127"),
			a:      _p("2001:db8:85a3::8a2e:370:7334/128"),
			b:      _p("2001:db8:85a3::8a2e:370:7335/128"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.prefix.String(), func(t *testing.T) {
			a, b := tt.prefix.Halves()
			assert.Equal(t, tt.a, a)
			assert.Equal(t, tt.b, b)
		})
	}
}

func TestPrefixAsMapKey(t *testing.T) {
	m := make(map[Prefix]bool)

	m[_p("2001::/56")] = true

	assert.True(t, m[_p("2001::/56")])
}
