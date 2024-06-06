package ipv6

import (
	"fmt"
	"net"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func _p(cidr string) Prefix {
	prefix, err := PrefixFromString(cidr)
	if err != nil {
		panic("only use this in happy cases")
	}
	return prefix
}

func unsafePrefixFromUint64(high, low uint64, length int) Prefix {
	mask, err := MaskFromLength(length)
	if err != nil {
		panic("only use this in happy cases")
	}
	return PrefixFromAddressMask(Address{uint128{high, low}}, mask)
}

func unsafeParseNet(prefix string) *net.IPNet {
	ipNet, err := parseNet(prefix)
	if err != nil {
		panic("only use this in happy cases")
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

func TestPrefixFromString(t *testing.T) {
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
			net, err := PrefixFromString(tt.cidr)
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

func TestPrefixContainsPrefix(t *testing.T) {
	tests := []struct {
		description          string
		container, containee Prefix
	}{
		{
			description: "all",
			container:   _p("::/0"),
			containee:   _p("1:2:3::4/128"),
		},
		{
			description: "same host",
			container:   _p("1:2:3::4/128"),
			containee:   _p("1:2:3::4/128"),
		},
		{
			description: "same host route",
			container:   _p("1:2:3::4/128"),
			containee:   _p("1:2:3::4/128"),
		},
		{
			description: "same prefix",
			container:   _p("C000:1680::20/112"),
			containee:   _p("C000:1680::20/112"),
		},
		{
			description: "contained smaller",
			container:   _p("C000:1680::/32"),
			containee:   _p("C000:1680:20::/48"),
		},
		{
			description: "ignore host part",
			container:   _p("1:2:3::4/112"),
			containee:   _p("1:2:3::5/128"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			assert.True(t, tt.container.Contains(tt.containee))
			if tt.container == tt.containee {
				assert.True(t, tt.containee.Contains(tt.container))
			} else {
				assert.False(t, tt.containee.Contains(tt.container))
			}
		})
	}
}

func TestPrefixContainsAddress(t *testing.T) {
	tests := []struct {
		description     string
		container       Prefix
		containees, not []Address
	}{
		{
			description: "all",
			container:   _p("::/0"),
			containees: []Address{
				_a("2001:db8:3::4"),
				_a("2001:db8:192:168::4:2"),
			},
		},
		{
			description: "host route",
			container:   _p("2001:db8::1:2:3:4/128"),
			containees: []Address{
				_a("2001:db8::1:2:3:4"),
			},
			not: []Address{
				_a("2001:db8::1:2:3:5"),
				_a("2001:db8::1:2:3:3"),
			},
		},
		{
			description: "same prefix",
			container:   _p("2001:db8:192:168::/64"),
			containees: []Address{
				_a("2001:db8:192:168::1234"),
			},
		},
		{
			description: "contained smaller",
			container:   _p("2001:db8:192:168::/64"),
			containees: []Address{
				_a("2001:db8:192:168:20::"),
			},
		},
		{
			description: "ignore host part",
			container:   _p("2001:db8:1:2::3:4/64"),
			containees: []Address{
				_a("2001:db8:1:2::3:5"),
				_a("2001:db8:1:2:3::245"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			for i, containee := range tt.containees {
				t.Run(fmt.Sprintf("contains %d", i), func(t *testing.T) {
					assert.True(t, tt.container.Contains(containee))
				})
			}
			for i, notContainee := range tt.not {
				t.Run(fmt.Sprintf("doesn't contain %d", i), func(t *testing.T) {
					assert.False(t, tt.container.Contains(notContainee))
				})
			}
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

func TestPrefixSet(t *testing.T) {
	tests := []struct {
		prefix  Prefix
		in, out Address
	}{
		{
			prefix: _p("::/1"),
			in:     _a("2001::1"),
			out:    _a("8000::"),
		},
		{
			prefix: _p("2001:db8::/32"),
			in:     _a("2001:db8::"),
			out:    _a("2001:ef01::ab60:43"),
		},
		{
			prefix: _p("2001:db8:85a3:8a2e::/64"),
			in:     _a("2001:db8:85a3:8a2e::"),
			out:    _a("3abc:db8:90a3::231:1"),
		},
		{
			prefix: _p("2001:db8:85a3::8a2e:370:7334/127"),
			in:     _a("2001:db8:85a3::8a2e:370:7334"),
			out:    _a("2001:db8:85a3::8a2e:370:7336"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.prefix.String(), func(t *testing.T) {
			set := tt.prefix.Set()
			assert.True(t, set.Contains(tt.in))
			assert.False(t, set.Contains(tt.out))
		})
	}
}

func TestPrefixAsMapKey(t *testing.T) {
	m := make(map[Prefix]bool)

	m[_p("2001::/56")] = true

	assert.True(t, m[_p("2001::/56")])
}

func TestPrefixNumPrefixes(t *testing.T) {
	tests := []struct {
		description string
		prefix      Prefix
		length      uint32
		count       uint64
		error       bool
	}{
		{
			description: "overflow",
			length:      64,
			error:       true,
		}, {
			description: "bad length",
			length:      129,
			error:       true,
		}, {
			description: "same size",
			prefix:      _p("2001:db8::/64"),
			length:      64,
			count:       1,
		}, {
			description: "too big",
			prefix:      _p("2001:db8::/64"),
			length:      32,
		}, {
			description: "32",
			prefix:      _p("2001:db8::/32"),
			length:      64,
			count:       0x100000000,
		}, {
			description: "56",
			prefix:      _p("2001:db8::/56"),
			length:      64,
			count:       256,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			count, err := tt.prefix.NumPrefixes(tt.length)
			assert.Equal(t, tt.error, err != nil)
			assert.Equal(t, tt.count, count)
		})
	}
}
