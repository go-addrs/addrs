package ipv4

import (
	"fmt"
	"math/bits"
	"net"
)

// Mask represents a prefix mask. It has any number of leading 1s  and then the
// remaining bits are 0s up to the number of bits in an address. It can be all
// zeroes or all ones.
// The zero value of a Mask is "/0"
type Mask struct {
	ui uint32
}

const maxUint32 = ^uint32(0)

// MaskFromLength converts the given length into a mask with that number of leading 1s
func MaskFromLength(length int) (Mask, error) {
	if length < 0 || addressSize < length {
		return Mask{}, fmt.Errorf("failed to create Mask where length %d isn't between 0 and 32", length)
	}

	return lengthToMask(length), nil
}

// MaskFromBytes returns the mask represented by the given bytes ordered from
// highest to lowest significance
func MaskFromBytes(a, b, c, d byte) (Mask, error) {
	m := Mask{AddressFromBytes(a, b, c, d).ui}
	if !m.valid() {
		return Mask{}, fmt.Errorf("failed to create a valid mask from bytes: %d, %d, %d, %d", a, b, c, d)
	}
	return m, nil
}

// MaskFromUint32 returns the mask from its unsigned integer
// representation.
func MaskFromUint32(ui uint32) (Mask, error) {
	m := Mask{ui}
	if !m.valid() {
		return Mask{}, fmt.Errorf("failed to create a valid mask from uint32: %x", ui)
	}
	return m, nil
}

// MaskFromNetIPMask converts a net.IPMask to a Mask
func MaskFromNetIPMask(mask net.IPMask) (Mask, error) {
	ones, bits := mask.Size()
	if bits != addressSize {
		return Mask{}, fmt.Errorf("failed to convert IPMask with incorrect size")
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
	return bits.LeadingZeros32(^me.ui)
}

// ToNetIPMask returns the net.IPMask representation of this Mask
func (me Mask) ToNetIPMask() net.IPMask {
	return net.CIDRMask(me.Length(), addressSize)
}

// String returns the net.IPMask representation of this Mask
func (me Mask) String() string {
	return Address{me.ui}.String()
}

// Uint32 returns the mask as a uint32
func (me Mask) Uint32() uint32 {
	return me.ui
}

func (me Mask) valid() bool {
	return me.Length() == bits.OnesCount32(me.ui)
}

func lengthToMask(length int) Mask {
	return Mask{
		ui: maxUint32 << (addressSize - length),
	}
}
