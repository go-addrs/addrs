package ipv4

import (
	"net"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func unsafeParseAddr(str string) Addr {
	addr, _ := ParseAddr(str)
	return addr
}

func TestParseAddr(t *testing.T) {
	ip, err := ParseAddr("10.224.24.1")
	assert.Nil(t, err)
	assert.Equal(t, AddrFromUint32(0x0ae01801), ip)
}

func TestAddrFromStdIP(t *testing.T) {
	tests := []struct {
		description string
		ip          net.IP
		expected    Addr
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
			expected:    AddrFromUint32(0x0ae01801),
		},
		{
			description: "ipv6",
			ip:          net.ParseIP("2001::"),
			isErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			ip, err := AddrFromStdIP(tt.ip)
			if tt.isErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.expected, ip)
			}
		})
	}
}

func TestAddrEquality(t *testing.T) {
	first, second := AddrFromUint32(0x0ae01801), AddrFromUint32(0x0ae01801)
	assert.Equal(t, first, second)
	assert.True(t, first.Equal(second))
	assert.True(t, first == second)
	assert.True(t, reflect.DeepEqual(first, second))

	third := AddrFromUint32(0x0ae01701)
	assert.NotEqual(t, third, second)
	assert.False(t, third.Equal(first))
	assert.False(t, third == first)
	assert.False(t, reflect.DeepEqual(third, first))
}

func TestAddrLessThan(t *testing.T) {
	first, second, third := AddrFromUint32(0x0ae01701), AddrFromUint32(0x0ae01801), AddrFromUint32(0x0ae01901)
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

func TestAddrMinAddr(t *testing.T) {
	first, second := AddrFromUint32(0x0ae01701), AddrFromUint32(0x0ae01801)

	assert.Equal(t, MinAddr(first, second), first)
	assert.Equal(t, MinAddr(second, first), first)
}

func TestAddrMaxAddr(t *testing.T) {
	first, second := AddrFromUint32(0x0ae01701), AddrFromUint32(0x0ae01801)

	assert.Equal(t, MaxAddr(first, second), second)
	assert.Equal(t, MaxAddr(second, first), second)
}
