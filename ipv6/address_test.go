package ipv6

import (
	"net"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func _a(str string) Address {
	addr, err := AddressFromString(str)
	if err != nil {
		panic("only use this in happy cases")
	}
	return addr
}

func TestAddressComparable(t *testing.T) {
	tests := []struct {
		description string
		a, b        Address
		equal       bool
	}{
		{
			description: "equal",
			a:           _a("2001::"),
			b:           _a("2001::"),
			equal:       true,
		}, {
			description: "not equal",
			a:           _a("2001:db8:85a3::8a2e:370:7334"),
			b:           _a("2001:db8:85a3::8a2e:370:7335"),
			equal:       false,
		}, {
			description: "extremes",
			a:           _a("::"),
			b:           _a("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff"),
			equal:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			assert.Equal(t, tt.equal, tt.a == tt.b)
			assert.Equal(t, !tt.equal, tt.a != tt.b)
			assert.Equal(t, tt.equal, tt.a.Prefix() == tt.b.Prefix())
			assert.Equal(t, !tt.equal, tt.a.Prefix() != tt.b.Prefix())
		})
	}
}

func TestAddressSize(t *testing.T) {
	assert.Equal(t, 128, Address{}.NumBits())
}

func TestAddressFromString(t *testing.T) {
	ip, err := AddressFromString("2001:0db8:85a3:0000:0000:8a2e:0370:7334")
	assert.Nil(t, err)
	assert.Equal(t, AddressFromUint64(0x20010db885a30000, 0x8a2e03707334), ip)
}

func TestAddressFromNetIP(t *testing.T) {
	tests := []struct {
		description string
		ip          net.IP
		expected    Address
		isErr       bool
	}{
		{
			description: "nil",
			ip:          nil,
			isErr:       true,
		},
		{
			description: "ipv4",
			ip:          net.ParseIP("10.224.24.1").To4(),
			isErr:       true,
		},
		{
			description: "ipv6 dotted quad",
			ip:          net.ParseIP("::ffff:203.0.113.17"),
			expected:    AddressFromUint64(0x0, 0xffffcb007111),
		},
		{
			description: "ipv6 standard",
			ip:          net.ParseIP("2001:0db8:85a3:0000:0000:8a2e:0370:7334"),
			expected:    AddressFromUint64(0x20010db885a30000, 0x8a2e03707334),
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			ip, err := AddressFromNetIP(tt.ip)
			if tt.isErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.expected, ip)
				assert.True(t, tt.ip.Equal(ip.ToNetIP()))
				assert.Equal(t, 16, len(ip.ToNetIP()))
			}
		})
	}
}

func TestAddressEquality(t *testing.T) {
	first, second := AddressFromUint64(0x20010db885a30000, 0x8a2e03707334), AddressFromUint64(0x20010db885a30000, 0x8a2e03707334)
	assert.Equal(t, first, second)
	assert.True(t, first == second)
	assert.True(t, reflect.DeepEqual(first, second))

	third := AddressFromUint64(0x20010db885a30000, 0x8a2e03707434)
	assert.NotEqual(t, third, second)
	assert.False(t, third == first)
	assert.False(t, reflect.DeepEqual(third, first))
}

func TestAddressLessThan(t *testing.T) {
	first, second, third := AddressFromUint64(0x20010db885a30000, 0x8a2e03707334), AddressFromUint64(0x20010db885a30000, 0x8a2e03707434), AddressFromUint64(0x20010db885a30000, 0x8a2e03707534)
	assert.True(t, first.lessThan(second))
	assert.True(t, second.lessThan(third))
	assert.True(t, first.lessThan(third))

	assert.False(t, second.lessThan(first))
	assert.False(t, third.lessThan(second))
	assert.False(t, third.lessThan(first))

	assert.False(t, first.lessThan(first))
	assert.False(t, second.lessThan(second))
	assert.False(t, third.lessThan(third))
}

func TestAddressMinAddress(t *testing.T) {
	first, second := AddressFromUint64(0x20010db885a30000, 0x0), AddressFromUint64(0x20010db885a30000, 0x8a2e03707334)

	assert.Equal(t, minAddress(first, second), first)
	assert.Equal(t, minAddress(second, first), first)
}

func TestAddressMaxAddress(t *testing.T) {
	first, second := AddressFromUint64(0x20010db885a30000, 0x0), AddressFromUint64(0x20010db885a30000, 0x8a2e03707334)

	assert.Equal(t, maxAddress(first, second), second)
	assert.Equal(t, maxAddress(second, first), second)
}

func TestAddressAsMapKey(t *testing.T) {
	m := make(map[Address]bool)

	m[_a("2001:db8:85a3::8a2e:370:7335")] = true

	assert.True(t, m[_a("2001:db8:85a3::8a2e:370:7335")])
}

func TestAddressToString(t *testing.T) {
	ip := AddressFromUint64(0x20010db885a30000, 0x8a2e03707334)
	assert.Equal(t, ip.String(), "2001:db8:85a3::8a2e:370:7334")
}

func TestAddressFromUint16(t *testing.T) {
	ip := AddressFromUint16(0x2001, 0xdb8, 0x85a3, 0xabcd, 0, 0, 0, 0x1)
	assert.Equal(t, ip.String(), "2001:db8:85a3:abcd::1")
}
