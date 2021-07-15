package ipv4

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
