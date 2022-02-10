package ipv6

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaskComparable(t *testing.T) {
	tests := []struct {
		description string
		a, b        Mask
		equal       bool
	}{
		{
			description: "equal",
			a:           Mask{ui: uint128{0xffffffffffffffff, 0xffff000000000000}},
			b:           Mask{ui: uint128{0xffffffffffffffff, 0xffff000000000000}},
			equal:       true,
		}, {
			description: "not equal",
			a:           Mask{ui: uint128{0xffffffffffffffff, 0x0000000000000000}},
			b:           Mask{ui: uint128{0xffffffffffffffff, 0xffffe00000000000}},
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

func TestMaskLength(t *testing.T) {
	assert.Equal(t, 0, Mask{uint128{0x0000000000000000, 0}}.Length())
	assert.Equal(t, 16, Mask{uint128{0xffff000000000000, 0}}.Length())
	assert.Equal(t, 27, Mask{uint128{0xffffffe000000000, 0}}.Length())
	assert.Equal(t, 32, Mask{uint128{0xffffffff00000000, 0}}.Length())
	assert.Equal(t, 64, Mask{uint128{0xffffffffffffffff, 0}}.Length())
	assert.Equal(t, 80, Mask{uint128{0xffffffffffffffff, 0xffff000000000000}}.Length())
	assert.Equal(t, 128, Mask{uint128{0xffffffffffffffff, 0xffffffffffffffff}}.Length())
}

func _m(length int) Mask {
	m, err := MaskFromLength(length)
	if err != nil {
		panic("only use this is happy cases")
	}
	return m
}

func TestMaskFromUint16Error(t *testing.T) {
	maskError := func(a, b, c, d, e, f, g, h uint16) error {
		_, err := MaskFromUint16(a, b, c, d, e, f, g, h)
		return err
	}
	assert.NotNil(t, maskError(0, 0xffff, 0, 0, 0, 0, 0, 0))
	assert.NotNil(t, maskError(0xffff, 0xffff, 0xffef, 0xffff, 0, 0, 0, 0))
	assert.NotNil(t, maskError(0, 0, 0, 0, 0, 0xffff, 0, 0))
	assert.NotNil(t, maskError(0xffff, 0xffff, 0xffff, 0xffff, 0xffff, 0xffff, 0xffef, 0))
}

func TestMaskFromUint64Error(t *testing.T) {
	maskError := func(m1 uint64, m2 uint64) error {
		_, err := MaskFromUint64(m1, m2)
		return err
	}

	assert.NotNil(t, maskError(0x1, 0))
	assert.NotNil(t, maskError(0xfffffeff00000000, 0))
	assert.NotNil(t, maskError(0xf8ffffff00000000, 0))
	assert.NotNil(t, maskError(0xfffffff100000000, 0))
}

func TestMaskFromUint16(t *testing.T) {
	assert.Equal(t, Mask{uint128{0x0000000000000000, 0}}, _m(0))
	assert.Equal(t, Mask{uint128{0xffff000000000000, 0}}, _m(16))
	assert.Equal(t, Mask{uint128{0xffffffe000000000, 0}}, _m(27))
	assert.Equal(t, Mask{uint128{0xffffffff00000000, 0}}, _m(32))
	assert.Equal(t, Mask{uint128{0xffffffffffffffff, 0}}, _m(64))
	assert.Equal(t, Mask{uint128{0xffffffffffffffff, 0xffff000000000000}}, _m(80))
	assert.Equal(t, Mask{uint128{0xffffffffffffffff, 0xffffffffffffffff}}, _m(128))
}

func TestMaskFromNetIPMask(t *testing.T) {
	convert := func(ones, bits int) Mask {
		stdMask := net.CIDRMask(ones, bits)
		mask, err := MaskFromNetIPMask(stdMask)
		assert.Nil(t, err)
		return mask
	}
	assert.Equal(t, Mask{uint128{0x0000000000000000, 0}}, convert(0, addressSize))
	assert.Equal(t, Mask{uint128{0xffff000000000000, 0}}, convert(16, addressSize))
	assert.Equal(t, Mask{uint128{0xffffffe000000000, 0}}, convert(27, addressSize))
	assert.Equal(t, Mask{uint128{0xffffffff00000000, 0}}, convert(32, addressSize))
	assert.Equal(t, Mask{uint128{0xffffffffffffffff, 0}}, convert(64, addressSize))
	assert.Equal(t, Mask{uint128{0xffffffffffffffff, 0xffff000000000000}}, convert(80, addressSize))
	assert.Equal(t, Mask{uint128{0xffffffffffffffff, 0xfffffffffe000000}}, convert(103, addressSize))
	assert.Equal(t, Mask{uint128{0xffffffffffffffff, 0xffffffffffffffff}}, convert(128, addressSize))

	runWithError := func(ones, bits int) {
		stdMask := net.CIDRMask(ones, bits)
		_, err := MaskFromNetIPMask(stdMask)
		assert.NotNil(t, err)
	}
	runWithError(64, 32)
	runWithError(16, 32)
	runWithError(129, 128)
}

func TestMaskToNetIPMask(t *testing.T) {
	assert.Equal(t, net.CIDRMask(25, addressSize), Mask{uint128{0xffffff8000000000, 0}}.ToNetIPMask())
	assert.Equal(t, net.CIDRMask(72, addressSize), Mask{uint128{0xffffffffffffffff, 0xff00000000000000}}.ToNetIPMask())
}

func TestAddressString(t *testing.T) {
	ips := []string{
		"::",
		"2001::",
		"ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
		"2001:db8:85a3::8a2e:370:7334",
	}

	for _, ip := range ips {
		t.Run(ip, func(t *testing.T) {
			assert.Equal(t, ip, _a(ip).String())
		})
	}
}

func TestMaskString(t *testing.T) {
	tests := []struct {
		length int
		str    string
	}{
		{
			length: 0,
			str:    "::",
		},
		{
			length: 37,
			str:    "ffff:ffff:f800::",
		},
		{
			length: 64,
			str:    "ffff:ffff:ffff:ffff::",
		},
		{
			length: 100,
			str:    "ffff:ffff:ffff:ffff:ffff:ffff:f000:0",
		},
		{
			length: 128,
			str:    "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
		},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			assert.Equal(t, tt.str, lengthToMask(tt.length).String())
		})
	}
}

func TestMaskAsMapKey(t *testing.T) {
	m := make(map[Mask]bool)

	m[_m(71)] = true

	assert.True(t, m[_m(71)])
}
