package ip

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/addrs.v1/ipv4"
	"gopkg.in/addrs.v1/ipv6"
)

func TestParseAddress(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		_, err := AddressFromString("bogus")
		assert.NotNil(t, err)
	})
	t.Run("v4", func(t *testing.T) {
		a, err := AddressFromString("203.0.113.17")
		assert.Nil(t, err)
		assert.IsType(t, ipv4.Address{}, a)
		assert.Equal(t, "203.0.113.17", a.String())
	})
	t.Run("v6", func(t *testing.T) {
		a, err := AddressFromString("2001:db8::1")
		assert.Nil(t, err)
		assert.IsType(t, ipv6.Address{}, a)
		assert.Equal(t, "2001:db8::1", a.String())
	})
}

func TestAddressFromNetIP(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		_, err := AddressFromNetIP(net.IP{})
		assert.NotNil(t, err)
	})
	t.Run("v4", func(t *testing.T) {
		a, err := AddressFromNetIP(net.IPv4(203, 0, 113, 29).To4())
		assert.Nil(t, err)
		assert.IsType(t, ipv4.Address{}, a)
		assert.Equal(t, "203.0.113.29", a.String())
	})
	t.Run("v6", func(t *testing.T) {
		a, err := AddressFromNetIP(net.ParseIP("2001:db8::1"))
		assert.Nil(t, err)
		assert.IsType(t, ipv6.Address{}, a)
		assert.Equal(t, "2001:db8::1", a.String())
	})
}

func TestPrefixFromAddress(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		p := PrefixFromAddress(nil)
		assert.Nil(t, p)
	})
	t.Run("v4", func(t *testing.T) {
		a, _ := AddressFromString("1.2.3.4")
		p := PrefixFromAddress(a)
		assert.IsType(t, ipv4.Prefix{}, p)
		assert.Equal(t, "1.2.3.4/32", p.String())
	})
	t.Run("v6", func(t *testing.T) {
		a, _ := AddressFromString("2001:db8::1")
		p := PrefixFromAddress(a)
		assert.IsType(t, ipv6.Prefix{}, p)
		assert.Equal(t, "2001:db8::1/128", p.String())
	})
}
