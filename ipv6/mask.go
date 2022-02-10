package ipv6

import (
	"fmt"
	"net"
)

// Mask represents an IPv6 prefix mask. It has 0-128 leading 1s and then all
// remaining bits are 0s
type Mask struct {
	ui uint128
}

var maxUint128 = uint128{^uint64(0), ^uint64(0)}

// MaskFromLength converts the given length into a mask with that number of leading 1s
func MaskFromLength(length int) (Mask, error) {
	if length < 0 || addressSize < length {
		return Mask{}, fmt.Errorf("failed to create Mask where length %d isn't between 0 and 128", length)
	}

	return lengthToMask(length), nil
}

// MaskFromUint16 returns the mask represented by the given uint16s ordered from
// highest to lowest significance
func MaskFromUint16(a, b, c, d, e, f, g, h uint16) (Mask, error) {
	m := Mask{AddressFromUint16(a, b, c, d, e, f, g, h).ui}
	if !m.valid() {
		return Mask{}, fmt.Errorf("failed to create a valid mask from uint16s: %x, %x, %x, %x, %x, %x, %x, %x", a, b, c, d, e, f, g, h)
	}
	return m, nil
}

// MaskFromUint64 returns the mask from its two uint64 unsigned integer
// representation.
func MaskFromUint64(high uint64, low uint64) (Mask, error) {
	m := Mask{uint128{high, low}}
	if !m.valid() {
		return Mask{}, fmt.Errorf("failed to create a valid mask from uint64: %x, %x", high, low)
	}
	return m, nil
}

// MaskFromNetIPMask converts a net.IPMask to a Mask
func MaskFromNetIPMask(mask net.IPMask) (Mask, error) {
	ones, bits := mask.Size()
	if bits != addressSize {
		return Mask{}, fmt.Errorf("failed to convert IPMask with size != 128")
	}
	m, err := MaskFromLength(ones)
	if err != nil {
		return Mask{}, err
	}
	if !m.valid() {
		return Mask{}, fmt.Errorf("failed to create a valid mask from net.IPMask: %v", mask)
	}
	return m, nil
}

// Length returns the number of leading 1s in the mask
func (me Mask) Length() int {
	return me.ui.complement().leadingZeros()
}

// ToNetIPMask returns the net.IPMask representation of this Mask
func (me Mask) ToNetIPMask() net.IPMask {
	return net.CIDRMask(me.Length(), addressSize)
}

// String returns the net.IPMask representation of this Mask
func (me Mask) String() string {
	return Address{me.ui}.String()
}

// Uint64 returns the mask as two uint64s
func (me Mask) Uint64() (uint64, uint64) {
	return me.ui.uint64()
}

func (me Mask) valid() bool {
	return me.Length() == me.ui.onesCount()
}

func lengthToMask(length int) Mask {
	return Mask{
		ui: maxUint128.leftShift(addressSize - length),
	}
}
