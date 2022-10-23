package ip

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/addrs.v1/ipv4"
	"gopkg.in/addrs.v1/ipv6"
)

func TestRangeFromAddresses(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		_, _, err := RangeFromAddresses(nil, nil)
		assert.NotNil(t, err)
	})
	t.Run("nil address", func(t *testing.T) {
		var a1 Address
		a2, _ := ipv6.AddressFromString("2001:db8::1")
		_, _, err := RangeFromAddresses(a1, a2)
		assert.NotNil(t, err)
	})
	t.Run("nil mask", func(t *testing.T) {
		a1, _ := ipv4.AddressFromString("203.0.113.17")
		var a2 Address
		_, _, err := RangeFromAddresses(a1, a2)
		assert.NotNil(t, err)
	})
	t.Run("nil nil", func(t *testing.T) {
		var a1, a2 Address
		_, _, err := RangeFromAddresses(a1, a2)
		assert.NotNil(t, err)
	})
	t.Run("v4/6", func(t *testing.T) {
		a1, _ := ipv4.AddressFromString("203.0.113.17")
		a2, _ := ipv6.AddressFromString("2001:db8::1")
		_, _, err := RangeFromAddresses(a1, a2)
		assert.NotNil(t, err)
	})
	t.Run("v6/4", func(t *testing.T) {
		a1, _ := ipv6.AddressFromString("2001:db8::1")
		a2, _ := ipv4.AddressFromString("203.0.113.17")
		_, _, err := RangeFromAddresses(a1, a2)
		assert.NotNil(t, err)
	})
	t.Run("v4", func(t *testing.T) {
		a1, _ := ipv4.AddressFromString("203.0.113.17")
		a2, _ := ipv4.AddressFromString("203.0.113.27")
		r, empty, err := RangeFromAddresses(a1, a2)
		assert.Nil(t, err)
		assert.False(t, empty)
		assert.IsType(t, ipv4.Range{}, r)
		assert.Equal(t, "[203.0.113.17,203.0.113.27]", r.String())
	})
	t.Run("v6", func(t *testing.T) {
		a1, _ := ipv6.AddressFromString("2001:db8::1")
		a2, _ := ipv6.AddressFromString("2001:db8::101")
		r, empty, err := RangeFromAddresses(a1, a2)
		assert.Nil(t, err)
		assert.False(t, empty)
		assert.IsType(t, ipv6.Range{}, r)
		assert.Equal(t, "[2001:db8::1,2001:db8::101]", r.String())
	})
	t.Run("v4 empty", func(t *testing.T) {
		a1, _ := ipv4.AddressFromString("203.0.113.27")
		a2, _ := ipv4.AddressFromString("203.0.113.17")
		_, empty, err := RangeFromAddresses(a1, a2)
		assert.Nil(t, err)
		assert.True(t, empty)
	})
	t.Run("v6 empty", func(t *testing.T) {
		a1, _ := ipv6.AddressFromString("2001:db8::101")
		a2, _ := ipv6.AddressFromString("2001:db8::1")
		_, empty, err := RangeFromAddresses(a1, a2)
		assert.Nil(t, err)
		assert.True(t, empty)
	})
}

func TestFirstFromRange(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		a := FirstFromRange(nil)
		assert.Nil(t, a)
	})
	t.Run("v4", func(t *testing.T) {
		p, _ := PrefixFromString("1.2.3.0/24")
		a := FirstFromRange(RangeFromPrefix(p))
		assert.IsType(t, ipv4.Address{}, a)
		assert.Equal(t, "1.2.3.0", a.String())
	})
	t.Run("v6", func(t *testing.T) {
		p, _ := PrefixFromString("2001:db8::/64")
		a := FirstFromRange(RangeFromPrefix(p))
		assert.IsType(t, ipv6.Address{}, a)
		assert.Equal(t, "2001:db8::", a.String())
	})
}

func TestLastFromRange(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		a := LastFromRange(nil)
		assert.Nil(t, a)
	})
	t.Run("v4", func(t *testing.T) {
		p, _ := PrefixFromString("1.2.3.0/24")
		a := LastFromRange(RangeFromPrefix(p))
		assert.IsType(t, ipv4.Address{}, a)
		assert.Equal(t, "1.2.3.255", a.String())
	})
	t.Run("v6", func(t *testing.T) {
		p, _ := PrefixFromString("2001:db8::/64")
		a := LastFromRange(RangeFromPrefix(p))
		assert.IsType(t, ipv6.Address{}, a)
		assert.Equal(t, "2001:db8::ffff:ffff:ffff:ffff", a.String())
	})
}
