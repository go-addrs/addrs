package ipv6

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUint128FromBytes(t *testing.T) {
	tests := []struct {
		description string
		bytes       []byte
		expected    uint128
		isErr       bool
	}{
		{
			description: "nil",
			bytes:       nil,
			isErr:       true,
		},
		{
			description: "slice length not equal to 16",
			bytes:       []byte{0x20, 0x1, 0xd, 0xb8},
			isErr:       true,
		},
		{
			description: "valid",
			bytes:       []byte{0x20, 0x1, 0xd, 0xb8, 0x85, 0xa3, 0x0, 0x0, 0x0, 0x0, 0x8a, 0x2e, 0x3, 0x70, 0x74, 0x34},
			expected:    uint128{0x20010db885a30000, 0x8a2e03707434},
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			net, err := uint128FromBytes(tt.bytes)
			if tt.isErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.expected, net)
			}
		})
	}
}

func TestUint128ToBytes(t *testing.T) {
	uint128Bytes := []byte{0x20, 0x1, 0xd, 0xb8, 0x85, 0xa3, 0x0, 0x0, 0x0, 0x0, 0x8a, 0x2e, 0x3, 0x70, 0x74, 0x34}
	assert.Equal(t, uint128{0x20010db885a30000, 0x8a2e03707434}.toBytes(), uint128Bytes)
}

func TestLeadingZero(t *testing.T) {
	assert.Equal(t, 128, uint128{0x0, 0x0}.leadingZeros())
	assert.Equal(t, 100, uint128{0x0, 0x000000000F000000}.leadingZeros())
	assert.Equal(t, 84, uint128{0x0, 0x000e0000000000}.leadingZeros())
	assert.Equal(t, 64, uint128{0x0, 0xF000000000000000}.leadingZeros())
	assert.Equal(t, 36, uint128{0x000000000FFFFFFF, 0x0}.leadingZeros())
	assert.Equal(t, 0, uint128{0xF000000000000000, 0x0}.leadingZeros())
}

func TestOnesCount(t *testing.T) {
	assert.Equal(t, 0, uint128{0x0, 0x0}.onesCount())
	assert.Equal(t, 19, uint128{0x0, 0x8a2e03707434}.onesCount())
	assert.Equal(t, 35, uint128{0x20010db885a30000, 0x8a2e03707434}.onesCount())
	assert.Equal(t, 61, uint128{0x20010FFFFa30000, 0x8a2e0FFFFFFFF}.onesCount())
	assert.Equal(t, 80, uint128{0x20010db8FFFFFFF, 0x8FFFFF707FFFF}.onesCount())
	assert.Equal(t, 128, uint128{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF}.onesCount())
}

func TestCompare(t *testing.T) {
	tests := []struct {
		description string
		num         uint128
		expected    int
		comparedTo  uint128
	}{
		{
			description: "equal",
			num:         uint128{0x20010db885a30000, 0x8a2e03707434},
			comparedTo:  uint128{0x20010db885a30000, 0x8a2e03707434},
			expected:    0,
		},
		{
			description: "high less than",
			num:         uint128{0x20010db885a30000, 0x8a2e03707434},
			comparedTo:  uint128{0x20010db885a30001, 0x8a2e03707434},
			expected:    -1,
		},
		{
			description: "low less than",
			num:         uint128{0x20010db885a30000, 0x8a2e03707434},
			comparedTo:  uint128{0x20010db885a30000, 0x8a2e03707435},
			expected:    -1,
		},
		{
			description: "greater than",
			num:         uint128{0x20010db885a30000, 0x8a2e03707435},
			comparedTo:  uint128{0x20010db885a30000, 0x8a2e03707434},
			expected:    1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.num.compare(tt.comparedTo))
		})
	}
}

func TestComplement(t *testing.T) {
	assert.Equal(t, uint128{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF}, uint128{0x0, 0x0}.complement())
	assert.Equal(t, uint128{0xDFFEF2477A5CFFFF, 0xFFFF75D1FC8F8BCB}, uint128{0x20010db885a30000, 0x8a2e03707434}.complement())
	assert.Equal(t, uint128{0xFF15D400005CF286, 0xFFF75D1F00000000}, uint128{0x0ea2bFFFFa30d79, 0x8a2e0FFFFFFFF}.complement())
	assert.Equal(t, uint128{0xFFFFFFFFF0000000, 0xFFFFFFFFFFFFFFFF}, uint128{0x000000000FFFFFFF, 0x0}.complement())
	assert.Equal(t, uint128{0x0, 0x0}, uint128{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF}.complement())
}

func TestAnd(t *testing.T) {
	tests := []struct {
		description string
		num         uint128
		andWith     uint128
		expected    uint128
	}{
		{
			description: "same",
			num:         uint128{0x20010db885a30000, 0x00008a2e03707434},
			andWith:     uint128{0x20010db885a30000, 0x00008a2e03707434},
			expected:    uint128{0x20010db885a30000, 0x00008a2e03707434},
		},
		{
			description: "different",
			num:         uint128{0x20010db885a30000, 0x00008a2e03707434},
			andWith:     uint128{0x27810ec88f12c000, 0x11008a5fa37d1934},
			expected:    uint128{0x20010c8885020000, 0x00008a0e03701034},
		},
		{
			description: "extreme",
			num:         uint128{0x0, 0x0},
			andWith:     uint128{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF},
			expected:    uint128{0x0, 0x0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.num.and(tt.andWith))
		})
	}
}

func TestOr(t *testing.T) {
	tests := []struct {
		description string
		num         uint128
		orWith      uint128
		expected    uint128
	}{
		{
			description: "same",
			num:         uint128{0x20010db885a30000, 0x00008a2e03707434},
			orWith:      uint128{0x20010db885a30000, 0x00008a2e03707434},
			expected:    uint128{0x20010db885a30000, 0x00008a2e03707434},
		},
		{
			description: "different",
			num:         uint128{0x20010db885a30000, 0x00008a2e03707434},
			orWith:      uint128{0x27810ec88f12c000, 0x11008a5fa37d1934},
			expected:    uint128{0x27810ff88fb3c000, 0x11008a7fa37d7d34},
		},
		{
			description: "extreme",
			num:         uint128{0x0, 0x0},
			orWith:      uint128{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF},
			expected:    uint128{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF},
		},
	}
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.num.or(tt.orWith))
		})
	}
}

func TestLeftShift(t *testing.T) {
	assert.Equal(t, uint128{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF}, uint128{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF}.leftShift(0))
	assert.Equal(t, uint128{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFF00000000}, uint128{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF}.leftShift(32))
	assert.Equal(t, uint128{0xFFFFFFFFFFFFFFFF, 0xFFFF800000000000}, uint128{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF}.leftShift(47))
	assert.Equal(t, uint128{0xFFFFFFFFFFFFFFFF, 0x0}, uint128{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF}.leftShift(64))
	assert.Equal(t, uint128{0xFFFFFFFFFFFF8000, 0x0}, uint128{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF}.leftShift(79))
	assert.Equal(t, uint128{0xFFFFFFF000000000, 0x0}, uint128{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF}.leftShift(100))
	assert.Equal(t, uint128{0x0, 0x0}, uint128{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF}.leftShift(128))
	assert.Equal(t, uint128{0x0, 0x0}, uint128{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF}.leftShift(129))
}

func TestRightShift(t *testing.T) {
	assert.Equal(t, uint128{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF}, uint128{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF}.rightShift(0))
	assert.Equal(t, uint128{0x00000000FFFFFFFF, 0xFFFFFFFFFFFFFFFF}, uint128{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF}.rightShift(32))
	assert.Equal(t, uint128{0x000000000001FFFF, 0xFFFFFFFFFFFFFFFF}, uint128{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF}.rightShift(47))
	assert.Equal(t, uint128{0x0, 0xFFFFFFFFFFFFFFFF}, uint128{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF}.rightShift(64))
	assert.Equal(t, uint128{0x0, 0x0001FFFFFFFFFFFF}, uint128{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF}.rightShift(79))
	assert.Equal(t, uint128{0x0, 0x000000000FFFFFFF}, uint128{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF}.rightShift(100))
	assert.Equal(t, uint128{0x0, 0x0}, uint128{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF}.rightShift(128))
	assert.Equal(t, uint128{0x0, 0x0}, uint128{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF}.rightShift(129))
}

func TestAddUint64(t *testing.T) {
	assert.Equal(t, uint128{0x20010db885a30001, 0x00008a2e03707433}, uint128{0x20010db885a30000, 0x00008a2e03707434}.addUint64(0xFFFFFFFFFFFFFFFF))
	assert.Equal(t, uint128{0x20010db885a30000, 0x00008a2f03707433}, uint128{0x20010db885a30000, 0x00008a2e03707434}.addUint64(0x00000000FFFFFFFF))
	assert.Equal(t, uint128{0x20010db885a30000, 0x00008a2e03707435}, uint128{0x20010db885a30000, 0x00008a2e03707434}.addUint64(1))
	assert.Equal(t, uint128{0x20010db885a30000, 0x00008a2e03707434}, uint128{0x20010db885a30000, 0x00008a2e03707434}.addUint64(0))
}

func TestSubtractUint64(t *testing.T) {
	assert.Equal(t, uint128{0x20010db885a2ffff, 0x00008a2e03707435}, uint128{0x20010db885a30000, 0x00008a2e03707434}.subtractUint64(0xFFFFFFFFFFFFFFFF))
	assert.Equal(t, uint128{0x20010db885a30000, 0x00008a2d03707435}, uint128{0x20010db885a30000, 0x00008a2e03707434}.subtractUint64(0x00000000FFFFFFFF))
	assert.Equal(t, uint128{0x20010db885a30000, 0x00008a2e03707433}, uint128{0x20010db885a30000, 0x00008a2e03707434}.subtractUint64(1))
	assert.Equal(t, uint128{0x20010db885a30000, 0x00008a2e03707434}, uint128{0x20010db885a30000, 0x00008a2e03707434}.subtractUint64(0))
}
