package ipv4

import (
	"fmt"
	"math/bits"
	"net"
)

// Mask represents an IPv4 prefix mask. It has 0-32 leading 1s and then all
// remaining bits are 0s
// The zero value of a Mask is "/0"
type Mask struct {
	ui uint32
}

const maxUint32 = ^uint32(0)

// MaskFromLength converts the given length (0-32) into a mask with that number of leading 1s
func MaskFromLength(length int) (Mask, error) {
	if length < 0 || addressSize < length {
		return Mask{}, fmt.Errorf("failed to create Mask where length %d isn't between 0 and 32", length)
	}

	return lengthToMask(length), nil
}

// MaskFromBytes returns the IPv4 address mask represented by `a.b.c.d`.
func MaskFromBytes(a, b, c, d byte) (Mask, error) {
	m := Mask{AddressFromBytes(a, b, c, d).ui}
	if !m.valid() {
		return Mask{}, fmt.Errorf("failed to create a valid mask from bytes: %d, %d, %d, %d", a, b, c, d)
	}
	return m, nil
}

// MaskFromUint32 returns the IPv4 mask from its 32 bit unsigned
// representation. The string of bits can only have one transition from 1s to
// 0s. For example: 11111111111111111110000000000000. It can be all 1s or all
// 0s.
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
		return Mask{}, fmt.Errorf("failed to convert IPMask with size != 32")
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
