package ipv4

import (
	"net"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func _a(str string) Address {
	addr, err := ParseAddress(str)
	if err != nil {
		panic("only use this is happy cases")
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
			a:           _a("10.0.0.1"),
			b:           _a("10.0.0.1"),
			equal:       true,
		}, {
			description: "not equal",
			a:           _a("10.0.0.1"),
			b:           _a("10.0.0.2"),
			equal:       false,
		}, {
			description: "extremes",
			a:           _a("0.0.0.0"),
			b:           _a("255.255.255.255"),
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
	assert.Equal(t, 32, Address{}.Size())
}

func TestParseAddress(t *testing.T) {
	ip, err := ParseAddress("10.224.24.1")
	assert.Nil(t, err)
	assert.Equal(t, AddressFromUint32(0x0ae01801), ip)
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
			ip:          net.ParseIP("10.224.24.1"),
			expected:    AddressFromUint32(0x0ae01801),
		},
		{
			description: "ipv6",
			ip:          net.ParseIP("2001::"),
			isErr:       true,
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
			}
		})
	}
}

func TestAddressEquality(t *testing.T) {
	first, second := AddressFromUint32(0x0ae01801), AddressFromUint32(0x0ae01801)
	assert.Equal(t, first, second)
	assert.True(t, first.Equal(second))
	assert.True(t, first == second)
	assert.True(t, reflect.DeepEqual(first, second))

	third := AddressFromUint32(0x0ae01701)
	assert.NotEqual(t, third, second)
	assert.False(t, third.Equal(first))
	assert.False(t, third == first)
	assert.False(t, reflect.DeepEqual(third, first))
}

func TestAddressLessThan(t *testing.T) {
	first, second, third := AddressFromUint32(0x0ae01701), AddressFromUint32(0x0ae01801), AddressFromUint32(0x0ae01901)
	assert.True(t, first.LessThan(second))
	assert.True(t, second.LessThan(third))
	assert.True(t, first.LessThan(third))

	assert.False(t, second.LessThan(first))
	assert.False(t, third.LessThan(second))
	assert.False(t, third.LessThan(first))

	assert.False(t, first.LessThan(first))
	assert.False(t, second.LessThan(second))
	assert.False(t, third.LessThan(third))
}

func TestAddressMinAddress(t *testing.T) {
	first, second := AddressFromUint32(0x0ae01701), AddressFromUint32(0x0ae01801)

	assert.Equal(t, MinAddress(first, second), first)
	assert.Equal(t, MinAddress(second, first), first)
}

func TestAddressMaxAddress(t *testing.T) {
	first, second := AddressFromUint32(0x0ae01701), AddressFromUint32(0x0ae01801)

	assert.Equal(t, MaxAddress(first, second), second)
	assert.Equal(t, MaxAddress(second, first), second)
}
