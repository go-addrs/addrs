/*
This file takes influence from uint128: https://github.com/davidminor/uint128

Copyright (c) 2014 David Minor

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/

package ipv6

import "math/bits"

type uint128 struct {
	high, low uint64
}

// Uint128FromUint64 returns a uint128 from its two 64 bit unsigned representations
func Uint128FromUint64(high uint64, low uint64) uint128 {
	return uint128{high, low}
}

// Uint128FromBytes returns the the uint128 converted from a slice of 16 bytes
func Uint128FromBytes(s []byte) uint128 {
	return uint128{
		high: uint64(s[0])<<56 | uint64(s[1])<<48 | uint64(s[2])<<40 | uint64(s[3])<<32 | uint64(s[4])<<24 | uint64(s[5])<<16 | uint64(s[6])<<8 | uint64(s[7]),
		low:  uint64(s[8])<<56 | uint64(s[9])<<48 | uint64(s[10])<<40 | uint64(s[11])<<32 | uint64(s[12])<<24 | uint64(s[13])<<16 | uint64(s[14])<<8 | uint64(s[15]),
	}
}

// MaxUint128 is the maximum integer that can be stored in a uint128, "all ones"
func MaxUint128() uint128 {
	return uint128{^uint64(0), ^uint64(0)}
}

// OnesCount128 returns the number of one bits ("population count") in x.
func OnesCount128(x uint128) int {
	return bits.OnesCount64(x.high) + bits.OnesCount64(x.low)
}

// LeadingZeros128 returns the number of leading zero bits in x; the result is 128 for x == 0.
func LeadingZeros128(x uint128) int {
	leadingZeros := bits.LeadingZeros64(x.high)
	if leadingZeros == 64 {
		leadingZeros += bits.LeadingZeros64(x.low)
	}
	return leadingZeros
}

// ToBytes returns a slice representation of 16 bytes of the uint128
func (me uint128) ToBytes() []byte {
	bytes := []byte{
		byte(0xff & (me.high >> 56)),
		byte(0xff & (me.high >> 48)),
		byte(0xff & (me.high >> 40)),
		byte(0xff & (me.high >> 32)),
		byte(0xff & (me.high >> 24)),
		byte(0xff & (me.high >> 16)),
		byte(0xff & (me.high >> 8)),
		byte(0xff & me.high),
		byte(0xff & (me.low >> 56)),
		byte(0xff & (me.low >> 48)),
		byte(0xff & (me.low >> 40)),
		byte(0xff & (me.low >> 32)),
		byte(0xff & (me.low >> 24)),
		byte(0xff & (me.low >> 16)),
		byte(0xff & (me.low >> 8)),
		byte(0xff & me.low),
	}
	return bytes
}

// Uint64 returns the address as two uint64
func (me uint128) Uint64() (uint64, uint64) {
	return me.high, me.low
}

// Equal reports whether this uint128 is the same as other
func (me uint128) Equal(other uint128) bool {
	return me == other
}

// Compare returns comparison of two uint128s and returns:
//  O if equal
// -1 if me is less than other
//  1 if me is greater than other
func (me uint128) Compare(other uint128) int {
	if me == other {
		return 0
	} else if me.high < other.high || (me.high == other.high && me.low < other.low) {
		return -1
	} else {
		return 1
	}
}

// IsZero returns whether uint128 equals zero
func (me uint128) IsZero() bool {
	return me.low == 0 && me.high == 0
}

// SubtractUint64 returns difference of y (uint64) from x (uint128)
func SubtractUint64(x uint128, y uint64) uint128 {
	low := x.low - y
	high := x.high
	if x.low < low {
		high--
	}
	return uint128{high, low}
}

// SubtractUint64 returns difference of y (uint128) from x (uint128)
func SubtractUint128(x uint128, y uint128) uint128 {
	low, borrow := bits.Sub64(x.low, y.low, 0)
	high, borrow := bits.Sub64(x.high, y.high, borrow)
	if borrow != 0 {
		panic("underflow")
	}
	return uint128{high, low}
}

// AddUint64 returns sum of x (uint128) with y (uint64)
func AddUint64(x uint128, y uint64) uint128 {
	low := x.low + y
	high := x.high
	if x.low > low {
		high++
	}
	return uint128{high, low}
}

// AddUint128 returns sum of x (uint128) with y (uint128)
func AddUint128(x uint128, y uint128) uint128 {
	low, borrow := bits.Add64(x.low, y.low, 0)
	high, borrow := bits.Add64(x.high, y.high, borrow)
	if borrow != 0 {
		panic("overflow")
	}
	return uint128{high, low}
}

// And returns a bitwise AND of x with y
func And(x uint128, y uint128) uint128 {
	high := x.high & y.high
	low := x.low & y.low
	return uint128{high, low}
}

// Xor returns a bitwise XOR of x with y
func Xor(x uint128, y uint128) uint128 {
	high := x.high ^ y.high
	low := x.low ^ y.low
	return uint128{high, low}
}

// Or returns a bitwise OR of x with y
func Or(x uint128, y uint128) uint128 {
	high := x.high | y.high
	low := x.low | y.low
	return uint128{high, low}
}

// Complement returns the bitwise complement of x
func Complement(x uint128) uint128 {
	high := ^x.high
	low := ^x.low
	return uint128{high, low}
}

// LeftShift returns a the bitwise shift left of x by bits
func LeftShift(x uint128, bits int) uint128 {
	high := x.high
	low := x.low
	if bits >= 128 {
		high = 0
		low = 0
	} else if bits >= 64 {
		high = x.low << (bits - 64)
		low = 0
	} else {
		high <<= bits
		high |= low >> (64 - bits)
		low <<= bits
	}
	return uint128{high, low}
}

// RightShift returns a the bitwise shift right of x by bits
func RightShift(x uint128, bits int) uint128 {
	high := x.high
	low := x.low
	if bits >= 128 {
		high = 0
		low = 0
	} else if bits >= 64 {
		low = high >> (bits - 64)
		high = 0
	} else {
		low >>= bits
		low |= high << (64 - bits)
		high >>= bits
	}
	return uint128{high, low}
}
