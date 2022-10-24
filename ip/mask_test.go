package ip

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/addrs.v1/ipv4"
	"gopkg.in/addrs.v1/ipv6"
)

func TestMaskFromNetIPMask(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		m, err := MaskFromNetIPMask(net.IPMask{})
		assert.NotNil(t, err)
		assert.Nil(t, m)
	})
	t.Run("bad length", func(t *testing.T) {
		m, err := MaskFromNetIPMask(net.CIDRMask(24, 31))
		assert.NotNil(t, err)
		assert.Nil(t, m)
	})
	t.Run("negative length", func(t *testing.T) {
		m, err := MaskFromNetIPMask(net.CIDRMask(24, -1))
		assert.NotNil(t, err)
		assert.Nil(t, m)
	})
	t.Run("negative bits", func(t *testing.T) {
		m, err := MaskFromNetIPMask(net.CIDRMask(-1, 32))
		assert.NotNil(t, err)
		assert.Nil(t, m)
	})
	t.Run("v4", func(t *testing.T) {
		m, err := MaskFromNetIPMask(net.CIDRMask(24, 32))
		assert.Nil(t, err)
		assert.IsType(t, ipv4.Mask{}, m)
		assert.Equal(t, 24, m.Length())
		assert.Equal(t, "255.255.255.0", m.String())
	})
	t.Run("v6", func(t *testing.T) {
		m, err := MaskFromNetIPMask(net.CIDRMask(64, 128))
		assert.Nil(t, err)
		assert.IsType(t, ipv6.Mask{}, m)
		assert.Equal(t, 64, m.Length())
		assert.Equal(t, "ffff:ffff:ffff:ffff::", m.String())
	})
}
