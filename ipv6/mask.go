package ipv6

import (
	"fmt"
	"math/bits"
	"net"
)

// Mask represents an IPv6 prefix mask. It has 0-128 leading 1s and then all
// remaining bits are 0s
type Mask Address

// MaxUint64 is the maximum integer that can be stored in a uint64, "all ones"
const MaxUint64 = ^uint64(0)

// CreateMask converts the given length (0-128) into a mask with that number of leading 1s
func CreateMask(length int) (Mask, error) {
	if length < 0 || SIZE < length {
		return Mask{}, fmt.Errorf("failed to create Mask where length %d isn't between 0 and 128", length)
	}

	return lengthToMask(length), nil
}

// MaskFromBytes returns the IPv6 address mask represented by `a:b:c:d`.
func MaskFromBytes(bytes []byte) (Mask, error) {
	m := Mask(AddressFromBytes(bytes))
	if !m.valid() {
		return Mask{}, fmt.Errorf("failed to create a valid mask from bytes: %v", bytes)
	}
	return m, nil
}

// MaskFromUint64 returns the IPv6 mask from its 128 bit unsigned representation
func MaskFromUint64(high uint64, low uint64) (Mask, error) {
	m := Mask{high, low}
	if !m.valid() {
		return Mask{}, fmt.Errorf("failed to create a valid mask from uint64: %x %x", high, low)
	}
	return m, nil
}

// MaskFromStdIPMask converts a net.IPMask to a Mask
func MaskFromStdIPMask(mask net.IPMask) (Mask, error) {
	ones, bits := mask.Size()
	if bits != SIZE {
		return Mask{}, fmt.Errorf("failed to convert IPMask with size != 128")
	}
	m, err := CreateMask(ones)
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
	return bits.LeadingZeros64(^uint64(me.high)) + bits.LeadingZeros64(^uint64(me.low))
}

// ToStdIPMask returns the net.IPMask representation of this Mask
func (me Mask) ToStdIPMask() net.IPMask {
	return net.CIDRMask(me.Length(), SIZE)
}

// String returns the net.IPMask representation of this Mask
func (me Mask) String() string {
	return Address(me).String()
}

// Uint64 returns the mask as a uint64
func (me Mask) Uint64() (uint64, uint64) {
	return me.high, me.low
}

func (me Mask) valid() bool {
	return me.Length() == bits.OnesCount64(me.high) && me.Length() == bits.OnesCount64(me.low)
}

func lengthToMask(length int) Mask {
	var highLength int
	var lowLength int
	if length > REPRESENTATIVE_SIZE {
		highLength = REPRESENTATIVE_SIZE
		lowLength = length % REPRESENTATIVE_SIZE
	} else {
		highLength = length % REPRESENTATIVE_SIZE
		lowLength = 0
	}
	return Mask{
		high: MaxUint64 << (REPRESENTATIVE_SIZE - highLength),
		low:  MaxUint64 << (REPRESENTATIVE_SIZE - lowLength),
	}
}
