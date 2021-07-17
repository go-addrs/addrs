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

func TestAddrMin(t *testing.T) {
	first, second := AddrFromUint32(0x0ae01701), AddrFromUint32(0x0ae01801)

	assert.Equal(t, Min(first, second), first)
	assert.Equal(t, Min(second, first), first)
}

func TestAddrMax(t *testing.T) {
	first, second := AddrFromUint32(0x0ae01701), AddrFromUint32(0x0ae01801)

	assert.Equal(t, Max(first, second), second)
	assert.Equal(t, Max(second, first), second)
}

func TestDefaultMask(t *testing.T) {
	ip, _ := ParseAddr("192.0.2.1")
	assert.Equal(t, Mask{ui: 0xffffff00}, ip.DefaultMask())
}

func TestMaskLength(t *testing.T) {
	assert.Equal(t, 0, Mask{ui: 0x00000000}.Length())
	assert.Equal(t, 16, Mask{ui: 0xffff0000}.Length())
	assert.Equal(t, 27, Mask{ui: 0xffffffe0}.Length())
	assert.Equal(t, 32, Mask{ui: 0xffffffff}.Length())
}

func TestMaskFromBytes(t *testing.T) {
	assert.Equal(t, Mask{ui: 0x00000000}, MaskFromBytes(0x00, 0x00, 0x00, 0x00))
	assert.Equal(t, Mask{ui: 0xffff0000}, MaskFromBytes(0xff, 0xff, 0x00, 0x00))
	assert.Equal(t, Mask{ui: 0xffffffe0}, MaskFromBytes(0xff, 0xff, 0xff, 0xe0))
	assert.Equal(t, Mask{ui: 0xffffffff}, MaskFromBytes(0xff, 0xff, 0xff, 0xff))
}

func TestMaskFromStdIPMask(t *testing.T) {
	convert := func(ones, bits int) Mask {
		stdMask := net.CIDRMask(ones, bits)
		mask, err := MaskFromStdIPMask(stdMask)
		assert.Nil(t, err)
		return mask
	}
	assert.Equal(t, Mask{ui: 0x00000000}, convert(0, SIZE))
	assert.Equal(t, Mask{ui: 0xffff0000}, convert(16, SIZE))
	assert.Equal(t, Mask{ui: 0xffffffe0}, convert(27, SIZE))
	assert.Equal(t, Mask{ui: 0xffffffff}, convert(32, SIZE))

	runWithError := func(ones, bits int) {
		stdMask := net.CIDRMask(ones, bits)
		_, err := MaskFromStdIPMask(stdMask)
		assert.NotNil(t, err)
	}
	runWithError(64, 128)
	runWithError(16, 128)
	runWithError(33, 32)
}

func TestMaskToStdIPMask(t *testing.T) {
	assert.Equal(t, net.CIDRMask(25, SIZE), Mask{ui: 0xffffff80}.ToStdIPMask())
}
