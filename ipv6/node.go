package ipv6

// contains is a helper which compares to see if the shorter prefix contains the
// longer.
//
// This function is not generally safe. It assumes non-nil pointers and that
// smaller.length < larger.length.
//
// `matches`: is true if the shorter key is a prefix of the longer key.
// `exact`: is true if the two keys are exactly the same (implies `matches`)
// `common`: is always the number of bits that the two keys have in common
// `child`: tells whether the first non-common bit in `longer` is a 0 or 1. It
//          is only valid if either `matches` or `exact` is false. The
//          following table describes how to interpret results.

// | matches | exact | child | note
// |---------|-------|-------|-------
// | false   | NA    | 0     | the two are disjoint and `longer` compares less than `shorter`
// | false   | NA    | 1     | the two are disjoint and `longer` compares greater than `shorter`
// | true    | false | 0     | `longer` belongs in `shorter`'s `children[0]`
// | true    | false | 1     | `longer` belongs in `shorter`'s `children[1]`
// | true    | true  | NA    | `shorter` and `longer` are the same key
func contains(shorter, longer Prefix) (matches, exact bool, common uint32, child int) {
	mask := uint128{0xffffffffffffffff, 0xffffffffffffffff}.leftShift(int(128 - shorter.length))

	matches = shorter.addr.ui.and(mask) == longer.addr.ui.and(mask)
	if matches {
		exact = shorter.length == longer.length
		common = shorter.length
	} else {
		common = uint32(shorter.addr.ui.xor(longer.addr.ui).leadingZeros())
	}
	if !exact {
		// Whether `longer` goes on the left (0) or right (1)
		pivotMask := uint128{0x8000000000000000, 0}.rightShift(int(common))
		if (longer.addr.ui.and(pivotMask) != uint128{}) {
			child = 1
		}
	}
	return
}

const (
	compareSame        int = iota
	compareContains        // Second key is a subset of the first
	compareIsContained     // Second key is a superset of the first
	compareDisjoint
)

// compare is a helper which compares two keys to find their relationship
func compare(a, b Prefix) (result int, reversed bool, common uint32, child int) {
	var aMatch, bMatch bool
	// Figure out which is the longer prefix and reverse them if b is shorter
	reversed = b.length < a.length
	if reversed {
		bMatch, aMatch, common, child = contains(b, a)
	} else {
		aMatch, bMatch, common, child = contains(a, b)
	}
	switch {
	case aMatch && bMatch:
		result = compareSame
	case aMatch && !bMatch:
		result = compareContains
	case !aMatch && bMatch:
		result = compareIsContained
	case !aMatch && !bMatch:
		result = compareDisjoint
	}
	return
}
