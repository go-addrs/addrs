package ipv4

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
			a:           Mask{ui: 0xff000000},
			b:           Mask{ui: 0xff000000},
			equal:       true,
		}, {
			description: "not equal",
			a:           Mask{ui: 0xff000000},
			b:           Mask{ui: 0xffffff00},
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
	assert.Equal(t, 0, Mask{ui: 0x00000000}.Length())
	assert.Equal(t, 16, Mask{ui: 0xffff0000}.Length())
	assert.Equal(t, 27, Mask{ui: 0xffffffe0}.Length())
	assert.Equal(t, 32, Mask{ui: 0xffffffff}.Length())
}

func _m(length int) Mask {
	m, err := MaskFromLength(length)
	if err != nil {
		panic("only use this in happy cases")
	}
	return m
}

func TestMaskFromBytesError(t *testing.T) {
	maskError := func(a, b, c, d byte) error {
		_, err := MaskFromBytes(a, b, c, d)
		return err
	}

	assert.NotNil(t, maskError(0, 1, 0, 0))
	assert.NotNil(t, maskError(0xff, 0xff, 0xfe, 0xff))
}

func TestMaskFromUint32Error(t *testing.T) {
	maskError := func(m uint32) error {
		_, err := MaskFromUint32(m)
		return err
	}

	assert.NotNil(t, maskError(0x1))
	assert.NotNil(t, maskError(0xfffffeff))
	assert.NotNil(t, maskError(0xf8ffffff))
	assert.NotNil(t, maskError(0xfffffff1))
}

func TestMaskFromBytes(t *testing.T) {
	assert.Equal(t, Mask{ui: 0x00000000}, _m(0))
	assert.Equal(t, Mask{ui: 0xffff0000}, _m(16))
	assert.Equal(t, Mask{ui: 0xffffffe0}, _m(27))
	assert.Equal(t, Mask{ui: 0xffffffff}, _m(32))
}

func TestMaskFromNetIPMask(t *testing.T) {
	convert := func(ones, bits int) Mask {
		stdMask := net.CIDRMask(ones, bits)
		mask, err := MaskFromNetIPMask(stdMask)
		assert.Nil(t, err)
		return mask
	}
	assert.Equal(t, Mask{ui: 0x00000000}, convert(0, addressSize))
	assert.Equal(t, Mask{ui: 0xffff0000}, convert(16, addressSize))
	assert.Equal(t, Mask{ui: 0xffffffe0}, convert(27, addressSize))
	assert.Equal(t, Mask{ui: 0xffffffff}, convert(32, addressSize))

	runWithError := func(ones, bits int) {
		stdMask := net.CIDRMask(ones, bits)
		_, err := MaskFromNetIPMask(stdMask)
		assert.NotNil(t, err)
	}
	runWithError(64, 128)
	runWithError(16, 128)
	runWithError(33, 32)
}

func TestMaskToNetIPMask(t *testing.T) {
	assert.Equal(t, net.CIDRMask(25, addressSize), Mask{ui: 0xffffff80}.ToNetIPMask())
}

func TestAddressString(t *testing.T) {
	ips := []string{
		"0.0.0.0",
		"10.224.24.117",
		"255.255.255.255",
		"1.2.3.4",
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
			str:    "0.0.0.0",
		},
		{
			length: 32,
			str:    "255.255.255.255",
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

	m[_m(27)] = true

	assert.True(t, m[_m(27)])
}
