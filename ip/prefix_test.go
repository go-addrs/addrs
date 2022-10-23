package ip

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/addrs.v1/ipv4"
	"gopkg.in/addrs.v1/ipv6"
)

func TestPrefixFromAddressMask(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		_, err := PrefixFromAddressMask(nil, nil)
		assert.NotNil(t, err)
	})
	t.Run("nil address", func(t *testing.T) {
		var a Address
		m, _ := ipv6.MaskFromLength(24)
		_, err := PrefixFromAddressMask(a, m)
		assert.NotNil(t, err)
	})
	t.Run("nil mask", func(t *testing.T) {
		a, _ := ipv4.AddressFromString("203.0.113.17")
		var m Mask
		_, err := PrefixFromAddressMask(a, m)
		assert.NotNil(t, err)
	})
	t.Run("nil nil", func(t *testing.T) {
		var a Address
		var m Mask
		_, err := PrefixFromAddressMask(a, m)
		assert.NotNil(t, err)
	})
	t.Run("v4/6", func(t *testing.T) {
		a, _ := ipv4.AddressFromString("203.0.113.17")
		m, _ := ipv6.MaskFromLength(24)
		_, err := PrefixFromAddressMask(a, m)
		assert.NotNil(t, err)
	})
	t.Run("v6/4", func(t *testing.T) {
		a, _ := ipv6.AddressFromString("2001:db8::1")
		m, _ := ipv4.MaskFromLength(24)
		_, err := PrefixFromAddressMask(a, m)
		assert.NotNil(t, err)
	})
	t.Run("v4", func(t *testing.T) {
		a, _ := ipv4.AddressFromString("203.0.113.17")
		m, _ := ipv4.MaskFromLength(24)
		p, err := PrefixFromAddressMask(a, m)
		assert.Nil(t, err)
		assert.IsType(t, ipv4.Prefix{}, p)
		assert.Equal(t, "203.0.113.17/24", p.String())
	})
	t.Run("v6", func(t *testing.T) {
		a, _ := ipv6.AddressFromString("2001:db8::1")
		m, _ := ipv6.MaskFromLength(64)
		p, err := PrefixFromAddressMask(a, m)
		assert.Nil(t, err)
		assert.IsType(t, ipv6.Prefix{}, p)
		assert.Equal(t, "2001:db8::1/64", p.String())
	})
}

func TestParsePrefix(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		_, err := PrefixFromString("bogus")
		assert.NotNil(t, err)
	})
	t.Run("v4", func(t *testing.T) {
		p, err := PrefixFromString("203.0.113.17/24")
		assert.Nil(t, err)
		assert.IsType(t, ipv4.Prefix{}, p)
		assert.Equal(t, "203.0.113.17/24", p.String())
	})
	t.Run("v6", func(t *testing.T) {
		p, err := PrefixFromString("2001:db8::1/64")
		assert.Nil(t, err)
		assert.IsType(t, ipv6.Prefix{}, p)
		assert.Equal(t, "2001:db8::1/64", p.String())
	})
}

func TestPrefixFromNetIP(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		_, err := PrefixFromNetIPNet(&net.IPNet{})
		assert.NotNil(t, err)
	})
	t.Run("v4", func(t *testing.T) {
		netP, _ := ipv4.PrefixFromString("203.0.113.29/24")
		p, err := PrefixFromNetIPNet(netP.ToNetIPNet())
		assert.Nil(t, err)
		assert.IsType(t, ipv4.Prefix{}, p)
		assert.Equal(t, "203.0.113.29/24", p.String())
	})
	t.Run("v6", func(t *testing.T) {
		netP, _ := ipv6.PrefixFromString("2001:db8::1/64")
		p, err := PrefixFromNetIPNet(netP.ToNetIPNet())
		assert.Nil(t, err)
		assert.IsType(t, ipv6.Prefix{}, p)
		assert.Equal(t, "2001:db8::1/64", p.String())
	})
}

func TestAddressFromPrefix(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		a := AddressFromPrefix(nil)
		assert.Nil(t, a)
	})
	t.Run("v4", func(t *testing.T) {
		p, _ := PrefixFromString("1.2.3.4/24")
		a := AddressFromPrefix(p)
		assert.IsType(t, ipv4.Address{}, a)
		assert.Equal(t, "1.2.3.4", a.String())
	})
	t.Run("v6", func(t *testing.T) {
		p, _ := PrefixFromString("2001:db8::1/64")
		a := AddressFromPrefix(p)
		assert.IsType(t, ipv6.Address{}, a)
		assert.Equal(t, "2001:db8::1", a.String())
	})
}

func TestMaskFromPrefix(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		m := MaskFromPrefix(nil)
		assert.Nil(t, m)
	})
	t.Run("v4", func(t *testing.T) {
		p, _ := PrefixFromString("1.2.3.4/24")
		m := MaskFromPrefix(p)
		assert.IsType(t, ipv4.Mask{}, m)
		assert.Equal(t, "255.255.255.0", m.String())
	})
	t.Run("v6", func(t *testing.T) {
		p, _ := PrefixFromString("2001:db8::1/64")
		m := MaskFromPrefix(p)
		assert.IsType(t, ipv6.Mask{}, m)
		assert.Equal(t, "ffff:ffff:ffff:ffff::", m.String())
	})
}
