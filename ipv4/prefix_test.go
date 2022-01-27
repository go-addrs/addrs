package ipv4

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

func unsafePrefixFromUint32(ip uint32, length int) Prefix {
	mask, err := MaskFromLength(length)
	if err != nil {
		panic("only use this is happy cases")
	}
	return PrefixFromAddressMask(Address{ip}, mask)
}

func unsafeParseNet(prefix string) *net.IPNet {
	ipNet, err := parseNet(prefix)
	if err != nil {
		panic("only use this is happy cases")
	}
	return ipNet
}

func TestParsePrefix(t *testing.T) {
	tests := []struct {
		description string
		cidr        string
		expected    Prefix
		isErr       bool
	}{
		{
			description: "success",
			cidr:        "10.224.24.1/27",
			expected:    unsafePrefixFromUint32(0x0ae01801, 27),
		},
		{
			description: "ipv6",
			cidr:        "2001::1/64",
			isErr:       true,
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
			expected:    unsafePrefixFromUint32(0x0ae01801, 22),
		},
		{
			description: "ipv6",
			net:         unsafeParseNet("2001::/56"),
			isErr:       true,
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
	first, second := unsafePrefixFromUint32(0x0ae01801, 24), unsafePrefixFromUint32(0x0ae01801, 24)
	assert.Equal(t, first, second)
	assert.True(t, first.Equal(second))
	assert.True(t, first == second)
	assert.True(t, reflect.DeepEqual(first, second))

	third := unsafePrefixFromUint32(0x0ae01701, 24)
	assert.NotEqual(t, third, second)
	assert.False(t, third.Equal(first))
	assert.False(t, third == first)
	assert.False(t, reflect.DeepEqual(third, first))
}

func TestPrefixLessThan(t *testing.T) {
	prefixes := []Prefix{
		unsafePrefixFromUint32(0x0, 0),
		unsafePrefixFromUint32(0x0, 16),
		unsafePrefixFromUint32(0x0, 31),
		unsafePrefixFromUint32(0x0, 32),
		unsafePrefixFromUint32(0x0ae01701, 16),
		unsafePrefixFromUint32(0x0ae0ffff, 16),
		unsafePrefixFromUint32(0x0ae01701, 24),
		unsafePrefixFromUint32(0x0ae01702, 24),
		unsafePrefixFromUint32(0x0ae017ff, 24),
		unsafePrefixFromUint32(0x0ae01701, 32),
		unsafePrefixFromUint32(0x0ae01702, 32),
		unsafePrefixFromUint32(0x0ae017ff, 32),
		unsafePrefixFromUint32(0x0ae01801, 24),
	}

	for a := 0; a < len(prefixes); a++ {
		for b := a; b < len(prefixes); b++ {
			t.Run(fmt.Sprintf("%d < %d", a, b), func(t *testing.T) {
				if a == b {
					assert.False(t, prefixes[a].LessThan(prefixes[b]))
				} else {
					assert.True(t, prefixes[a].LessThan(prefixes[b]))
				}
				assert.False(t, prefixes[b].LessThan(prefixes[a]))
			})
		}
	}
}

func TestNetworkHostBroadcast(t *testing.T) {
	tests := []struct {
		description              string
		prefix                   Prefix
		network, host, broadcast Prefix
	}{
		{
			description: "0",
			prefix:      _p("10.224.24.1/0"),
			network:     _p("0.0.0.0/0"),
			host:        _p("10.224.24.1/0"),
			broadcast:   _p("255.255.255.255/0"),
		},
		{
			description: "8",
			prefix:      _p("10.224.24.1/8"),
			network:     _p("10.0.0.0/8"),
			host:        _p("0.224.24.1/8"),
			broadcast:   _p("10.255.255.255/8"),
		},
		{
			description: "22",
			prefix:      _p("10.224.24.1/22"),
			network:     _p("10.224.24.0/22"),
			host:        _p("0.0.0.1/22"),
			broadcast:   _p("10.224.27.255/22"),
		},
		{
			description: "32",
			prefix:      _p("10.224.24.1/32"),
			network:     _p("10.224.24.1/32"),
			host:        _p("0.0.0.0/32"),
			broadcast:   _p("10.224.24.1/32"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			assert.Equal(t, tt.network, tt.prefix.Network())
			assert.Equal(t, tt.host, tt.prefix.Host())
			assert.Equal(t, tt.broadcast, tt.prefix.Broadcast())
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
			container:   _p("0.0.0.0/0"),
			containee:   _p("1.2.3.4/32"),
		},
		{
			description: "same host",
			container:   _p("1.2.3.4/32"),
			containee:   _p("1.2.3.4/32"),
		},
		{
			description: "same host route",
			container:   _p("1.2.3.4/32"),
			containee:   _p("1.2.3.4/32"),
		},
		{
			description: "same prefix",
			container:   _p("192.168.20.0/24"),
			containee:   _p("192.168.20.0/24"),
		},
		{
			description: "contained smaller",
			container:   _p("192.168.0.0/16"),
			containee:   _p("192.168.20.0/24"),
		},
		{
			description: "ignore host part",
			container:   _p("1.2.3.4/24"),
			containee:   _p("1.2.3.5/32"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			assert.True(t, tt.container.Contains(tt.containee))
			if tt.container.Equal(tt.containee) {
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
			container:   _p("0.0.0.0/0"),
			containees: []Address{
				_a("1.2.3.4"),
				_a("192.168.4.2"),
			},
		},
		{
			description: "host route",
			container:   _p("1.2.3.4/32"),
			containees: []Address{
				_a("1.2.3.4"),
			},
			not: []Address{
				_a("1.2.3.5"),
				_a("1.2.3.3"),
			},
		},
		{
			description: "same prefix",
			container:   _p("192.168.20.0/24"),
			containees: []Address{
				_a("192.168.20.0"),
			},
		},
		{
			description: "contained smaller",
			container:   _p("192.168.0.0/16"),
			containees: []Address{
				_a("192.168.20.0"),
			},
		},
		{
			description: "ignore host part",
			container:   _p("1.2.3.4/24"),
			containees: []Address{
				_a("1.2.3.5"),
				_a("1.2.3.245"),
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

func TestPrefixSize(t *testing.T) {
	tests := []struct {
		description string
		prefix      Prefix
		expected    int64
	}{
		{
			description: "all",
			prefix:      _p("0.0.0.0/0"),
			expected:    int64(0x100000000),
		},
		{
			description: "private",
			prefix:      _p("172.16.0.0/12"),
			expected:    0x00100000,
		},
		{
			description: "host",
			prefix:      _p("172.16.244.117/32"),
			expected:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.prefix.Size())
		})
	}
}

func TestPrefixToNetIPNet(t *testing.T) {
	assert.Equal(t, "10.224.24.1/24", _p("10.224.24.1/24").ToNetIPNet().String())
}

func TestPrefixString(t *testing.T) {
	cidrs := []string{
		"0.0.0.0/0",
		"10.224.24.117/25",
		"1.2.3.4/32",
	}

	for _, cidr := range cidrs {
		t.Run(cidr, func(t *testing.T) {
			assert.Equal(t, cidr, _p(cidr).String())
		})
	}
}

func TestPrefixUint32(t *testing.T) {
	address, mask := _p("10.224.24.1/24").Uint32()
	assert.Equal(t, uint32(0x0ae01801), address)
	assert.Equal(t, uint32(0xffffff00), mask)
}

func TestPrefixFromAddressMask(t *testing.T) {
	address := Address{ui: 0x0ae01801}
	mask, _ := MaskFromLength(24)
	assert.Equal(t, Prefix{Address: address, length: 24}, PrefixFromAddressMask(address, mask))
}

func TestPrefixHalves(t *testing.T) {
	tests := []struct {
		prefix Prefix
		a, b   Prefix
	}{
		{
			prefix: _p("0.0.0.0/0"),
			a:      _p("0.0.0.0/1"),
			b:      _p("128.0.0.0/1"),
		},
		{
			prefix: _p("10.224.0.0/16"),
			a:      _p("10.224.0.0/17"),
			b:      _p("10.224.128.0/17"),
		},
		{
			prefix: _p("10.224.24.1/24"),
			a:      _p("10.224.24.0/25"),
			b:      _p("10.224.24.128/25"),
		},
		{
			prefix: _p("10.224.24.117/31"),
			a:      _p("10.224.24.116/32"),
			b:      _p("10.224.24.117/32"),
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

func TestIterateAddress(t *testing.T) {
	tests := []struct {
		prefix      Prefix
		first, last Address
	}{
		{
			prefix: _p("10.224.0.0/24"),
			first:  _a("10.224.0.0"),
			last:   _a("10.224.0.99"),
		},
		{
			prefix: _p("203.0.113.116/31"),
			first:  _a("203.0.113.116"),
			last:   _a("203.0.113.117"),
		},
		{
			prefix: _p("100.64.0.1/32"),
			first:  _a("100.64.0.1"),
			last:   _a("100.64.0.1"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.prefix.String(), func(t *testing.T) {
			count := 0
			ips := []Address{}
			tt.prefix.Iterate(func(ip Address) bool {
				count++
				ips = append(ips, ip)
				if count == 100 {
					return false
				}
				return true
			})
			assert.Equal(t, tt.first, ips[0])
			assert.Equal(t, tt.last, ips[len(ips)-1])
		})
	}
}

func TestPrefixSet(t *testing.T) {
	tests := []struct {
		prefix  Prefix
		in, out Address
	}{
		{
			prefix: _p("0.0.0.0/1"),
			in:     _a("127.0.0.1"),
			out:    _a("192.168.0.3"),
		},
		{
			prefix: _p("10.224.0.0/16"),
			in:     _a("10.224.0.123"),
			out:    _a("10.225.128.123"),
		},
		{
			prefix: _p("10.224.24.1/24"),
			in:     _a("10.224.24.0"),
			out:    _a("10.224.25.128"),
		},
		{
			prefix: _p("10.224.24.117/31"),
			in:     _a("10.224.24.116"),
			out:    _a("10.224.24.118"),
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
