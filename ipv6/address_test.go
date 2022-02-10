package ipv6

import (
	"net"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func unsafeParseAddress(str string) Address {
	addr, err := ParseAddress(str)
	if err != nil {
		panic("only use this is happy cases")
	}
	return addr
}

func TestParseAddress(t *testing.T) {
	ip, err := ParseAddress("2001:0db8:85a3:0000:0000:8a2e:0370:7334")
	assert.Nil(t, err)
	assert.Equal(t, AddressFromUint64(0x20010db885a30000, 0x8a2e03707334), ip)
}

func TestAddressFromStdIP(t *testing.T) {
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
			description: "ipv4_4_bytes",
			ip:          net.ParseIP("10.224.24.1").To4(),
			isErr:       true,
		},
		{
			description: "ipv4_16_bytes",
			ip:          net.ParseIP("10.224.24.1"),
			expected:    AddressFromUint64(0x0, 0xffff0ae01801),
		},
		{
			description: "ipv6",
			ip:          net.ParseIP("2001:0db8:85a3:0000:0000:8a2e:0370:7334"),
			expected:    AddressFromUint64(0x20010db885a30000, 0x8a2e03707334),
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			ip, err := AddressFromStdIP(tt.ip)
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
	first, second := AddressFromUint64(0x20010db885a30000, 0x8a2e03707334), AddressFromUint64(0x20010db885a30000, 0x8a2e03707334)
	assert.Equal(t, first, second)
	assert.True(t, first.Equal(second))
	assert.True(t, first == second)
	assert.True(t, reflect.DeepEqual(first, second))

	third := AddressFromUint64(0x20010db885a30000, 0x8a2e03707434)
	assert.NotEqual(t, third, second)
	assert.False(t, third.Equal(first))
	assert.False(t, third == first)
	assert.False(t, reflect.DeepEqual(third, first))
}

func TestAddressLessThan(t *testing.T) {
	first, second, third := AddressFromUint64(0x20010db885a30000, 0x8a2e03707334), AddressFromUint64(0x20010db885a30000, 0x8a2e03707434), AddressFromUint64(0x20010db885a30000, 0x8a2e03707534)
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
	first, second := AddressFromUint64(0x20010db885a30000, 0x0), AddressFromUint64(0x20010db885a30000, 0x8a2e03707334)

	assert.Equal(t, MinAddress(first, second), first)
	assert.Equal(t, MinAddress(second, first), first)
}

func TestAddressMaxAddress(t *testing.T) {
	first, second := AddressFromUint64(0x20010db885a30000, 0x0), AddressFromUint64(0x20010db885a30000, 0x8a2e03707334)

	assert.Equal(t, MaxAddress(first, second), second)
	assert.Equal(t, MaxAddress(second, first), second)
}

func TestAddressFromBytes(t *testing.T) {
	ip := AddressFromUint64(0x20010db885a30000, 0x8a2e03707434)
	assert.Equal(t, AddressFromBytes([]byte{0x20, 0x1, 0xd, 0xb8, 0x85, 0xa3, 0x0, 0x0, 0x0, 0x0, 0x8a, 0x2e, 0x3, 0x70, 0x74, 0x34}), ip)
}

func TestAddressToBytes(t *testing.T) {
	ip := AddressFromUint64(0x20010db885a30000, 0x8a2e03707434)
	assert.Equal(t, ip.toBytes(), []byte{0x20, 0x1, 0xd, 0xb8, 0x85, 0xa3, 0x0, 0x0, 0x0, 0x0, 0x8a, 0x2e, 0x3, 0x70, 0x74, 0x34})
}

func TestAddressToString(t *testing.T) {
	ip := AddressFromUint64(0x20010db885a30000, 0x8a2e03707334)
	assert.Equal(t, ip.String(), "2001:db8:85a3::8a2e:370:7334")
}
