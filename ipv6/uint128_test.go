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
