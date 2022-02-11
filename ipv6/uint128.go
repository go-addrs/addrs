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

// Uint128FromBytes returns the the uint128 converted from a slice of 16 bytes
func Uint128FromBytes(s []byte) uint128 {
	return uint128{
		high: uint64(s[0])<<56 |
			uint64(s[1])<<48 |
			uint64(s[2])<<40 |
			uint64(s[3])<<32 |
			uint64(s[4])<<24 |
			uint64(s[5])<<16 |
			uint64(s[6])<<8 |
			uint64(s[7]),
		low: uint64(s[8])<<56 |
			uint64(s[9])<<48 |
			uint64(s[10])<<40 |
			uint64(s[11])<<32 |
			uint64(s[12])<<24 |
			uint64(s[13])<<16 |
			uint64(s[14])<<8 |
			uint64(s[15]),
	}
}

// OnesCount128 returns the number of one bits ("population count") in x.
func (me uint128) OnesCount() int {
	return bits.OnesCount64(me.high) + bits.OnesCount64(me.low)
}

// LeadingZeros128 returns the number of leading zero bits in x; the result is 128 for x == 0.
func (me uint128) LeadingZeros() int {
	leadingZeros := bits.LeadingZeros64(me.high)
	if leadingZeros == 64 {
		leadingZeros += bits.LeadingZeros64(me.low)
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

// Uint64 returns the uint128 as two uint64
func (me uint128) Uint64() (uint64, uint64) {
	return me.high, me.low
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

// SubtractUint64 returns the difference of uint128 with x (uint64)
func (me uint128) SubtractUint64(x uint64) uint128 {
	low := me.low - x
	high := me.high
	if me.low < low {
		high--
	}
	return uint128{high, low}
}

// SubtractUint128 returns the difference of uint128 with x (uint128)
func (me uint128) SubtractUint128(x uint128) uint128 {
	low := me.low - x.low
	high := me.high - x.high
	if me.low < low {
		high--
	}
	return uint128{high, low}
}

// AddUint64 returns sum of uint128 with x (uint64)
func (me uint128) AddUint64(x uint64) uint128 {
	low := me.low + x
	high := me.high
	if me.low > low {
		high++
	}
	return uint128{high, low}
}

// AddUint128 returns sum of uint128 with x (uint128)
func (me uint128) AddUint128(x uint128) uint128 {
	low := me.low + x.low
	high := me.high + x.high
	if me.low > low {
		high++
	}
	return uint128{high, low}
}

// And returns a bitwise AND with x
func (me uint128) And(x uint128) uint128 {
	return uint128{me.high & x.high, me.low & x.low}
}

// Xor returns a bitwise XOR with x
func (me uint128) Xor(x uint128) uint128 {
	return uint128{me.high ^ x.high, me.low ^ x.low}
}

// Or returns a bitwise OR with x
func (me uint128) Or(x uint128) uint128 {
	return uint128{me.high | x.high, me.low | x.low}
}

// Complement returns the bitwise complement
func (me uint128) Complement() uint128 {
	return uint128{^me.high, ^me.low}
}

// LeftShift returns the bitwise shift left by bits
func (me uint128) LeftShift(bits int) uint128 {
	high := me.high
	low := me.low
	if bits >= 128 {
		high = 0
		low = 0
	} else if bits >= 64 {
		high = low << (bits - 64)
		low = 0
	} else {
		high <<= bits
		high |= low >> (64 - bits)
		low <<= bits
	}
	return uint128{high, low}
}

// RightShift returns the bitwise shift right by bits
func (me uint128) RightShift(bits int) uint128 {
	high := me.high
	low := me.low
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
