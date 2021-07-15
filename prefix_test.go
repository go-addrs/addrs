package ipv4

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func unsafeParsePrefix(cidr string) Prefix {
	prefix, _ := ParsePrefix(cidr)
	return prefix
}

func unsafePrefixFromUint32(ip, length uint32) Prefix {
	prefix, _ := PrefixFromUint32(ip, length)
	return prefix
}

func unsafeParseNet(prefix string) *net.IPNet {
	ipNet, _ := parseNet(prefix)
	return ipNet
}

func TestParsePrefix(t *testing.T) {
	tests := []struct {
		description string
		cidr        string
		expected    Prefix
		isErr       bool
	}{
		{
			description: "success",
			cidr:        "10.224.24.1/27",
			expected:    unsafePrefixFromUint32(0x0ae01801, 27),
		},
		{
			description: "ipv6",
			cidr:        "2001::1/64",
			isErr:       true,
		},
		{
			description: "bogus",
			cidr:        "bogus",
			isErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			net, err := ParsePrefix(tt.cidr)
			if tt.isErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.expected, net)
			}
		})
	}
}

func TestPrefixFromStdIPNet(t *testing.T) {
	tests := []struct {
		description string
		net         *net.IPNet
		expected    Prefix
		isErr       bool
	}{
		{
			description: "nil",
			net:         nil,
			isErr:       true,
		},
		{
			description: "ipv4",
			net:         unsafeParseNet("10.224.24.1/22"),
			expected:    unsafePrefixFromUint32(0x0ae01801, 22),
		},
		{
			description: "ipv6",
			net:         unsafeParseNet("2001::/56"),
			isErr:       true,
		},
		{
			description: "mixed up IPv4/6 IPNet",
			net: &net.IPNet{
				IP:   unsafeParseNet("2001::/56").IP,
				Mask: unsafeParseNet("10.0.0.0/16").Mask,
			},
			isErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			net, err := PrefixFromStdIPNet(tt.net)
			if tt.isErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.expected, net)
			}
		})
	}
}

func TestPrefixFromBytes(t *testing.T) {
	prefix, err := PrefixFromBytes(10, 224, 24, 1, 22)
	assert.Equal(t, unsafePrefixFromUint32(0x0ae01801, 22), prefix)
	assert.Nil(t, err)

	prefix, err = PrefixFromBytes(10, 224, 24, 1, 33)
	assert.Equal(t, unsafePrefixFromUint32(0x0ae01801, 33), prefix)
	assert.NotNil(t, err)
}

func TestPrefixFromUint32(t *testing.T) {
	prefix, err := PrefixFromUint32(0x0ae01801, 22)
	assert.Equal(t, unsafeParsePrefix("10.224.24.1/22"), prefix)
	assert.Nil(t, err)

	prefix, err = PrefixFromUint32(0x0ae01801, 33)
	assert.NotNil(t, err)
}
