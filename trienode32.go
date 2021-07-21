package ipv4

import (
	"fmt"
	"math/bits"
)

type trieNode32 struct {
	Prefix
	Data     interface{}
	size     uint32
	h        uint16
	isActive bool
	children [2]*trieNode32
}

// bitsToBytes calculates the number of bytes (including possible
// least-significant partial) to hold the given number of bits.
func bitsToBytes(bits uint32) uint32 {
	return (bits + 7) / 8
}

func intMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func intMax(a, b int) int {
	if a < b {
		return b
	}
	return a
}

// contains32 is a helper which compares to see if the shorter prefix contains the
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
func contains32(shorter, longer Prefix) (matches, exact bool, common, child uint32) {
	pivotMask := uint32(0x80000000)

	// calculate `exact`, `common`, and `child` at the end with defer
	defer func() {
		if !matches {
			s, l := shorter.ui, longer.ui

			common = uint32(bits.LeadingZeros32(s ^ l))

			// Whether `longer` goes on the left (0) or right (1)
			if longer.ui < shorter.ui {
				child = 0
			} else {
				child = 1
			}
			return
		}

		common = shorter.length
		exact = shorter.length == longer.length
		if !exact {
			// Whether `longer` goes on the left (0) or right (1)
			if longer.ui&pivotMask == 0 {
				child = 0
			} else {
				child = 1
			}
		}
	}()

	// Prefix length of 0 matches everything!
	if shorter.length == 0 {
		matches = true
		return
	}

	mask := uint32(0xffffffff) << (32 - shorter.length)

	if shorter.ui&mask != longer.ui&mask {
		matches = false
		return
	}

	pivotMask >>= shorter.length

	matches = true
	return
}

// compare32 is a helper which compares two keys to find their relationship
//
// This function is not generally safe. It assumes non-nil pointers.
func compare32(a, b Prefix) (aMatch, bMatch, reversed bool, common, child uint32) {
	// Figure out which is the longer prefix and reverse them if b is shorter
	reversed = b.length < a.length
	if reversed {
		bMatch, aMatch, common, child = contains32(b, a)
	} else {
		aMatch, bMatch, common, child = contains32(a, b)
	}
	return
}

func (me *trieNode32) makeCopy() *trieNode32 {
	if me == nil {
		return nil
	}
	doppelganger := &trieNode32{}
	*doppelganger = *me
	return doppelganger
}

// Get is the public form of get(...)
func (me *trieNode32) GetOrInsert(searchKey Prefix, data interface{}) (newHead, result *trieNode32, err error) {
	newHead, result = me.getOrInsert(searchKey, data)
	return
}

func (me *trieNode32) setSize() {
	// me is not nil by design
	me.size = uint32(me.children[0].Size() + me.children[1].Size())
	me.h = 1 + uint16(uint16(intMax(me.children[0].height(), me.children[1].height())))
	if me.isActive {
		me.size++
	}
}

// getOrInsert returns the existing value if an exact match is found, otherwise, inserts the given default
func (me *trieNode32) getOrInsert(searchKey Prefix, data interface{}) (head, result *trieNode32) {
	defer func() {
		if result == nil {
			result = &trieNode32{Prefix: searchKey, Data: data}

			// The only error from insert is that the key already exists. But, that cannot happen by design.
			head, _ = me.insert(result, true, false)
		}
	}()

	if me == nil || searchKey.length < me.Prefix.length {
		return
	}

	matches, exact, _, child := contains32(me.Prefix, searchKey)
	if !matches {
		return
	}

	if !exact {
		var newChild *trieNode32
		newChild, result = me.children[child].getOrInsert(searchKey, data)

		head = me.makeCopy()
		head.children[child] = newChild
		head.setSize()
		return
	}

	if !me.isActive {
		return
	}

	return me, me
}

// Match is the public form of match(...)
func (me *trieNode32) Match(searchKey Prefix) *trieNode32 {
	return me.match(searchKey)
}

// match returns the existing entry with the longest prefix that fully contains
// the prefix given by the key argument or nil if none match.
//
// "contains" means that the first "length" bits in the entry's key are exactly
// the same as the same number of first bits in the given search key. This
// implies the search key is at least as long as any matching node's prefix.
//
// Some examples include the following ipv4 and ipv6 matches:
//     10.0.0.0/24 contains 10.0.0.0/24, 10.0.0.0/25, and 10.0.0.0/32
//     2001:cafe:beef::/64 contains 2001:cafe:beef::a/124
//
// "longest" means that if multiple existing entries in the trie match the one
// with the longest length will be returned. It is the most specific match.
func (me *trieNode32) match(searchKey Prefix) *trieNode32 {
	if me == nil {
		return nil
	}

	nodeKey := me.Prefix
	if searchKey.length < nodeKey.length {
		return nil
	}

	matches, exact, _, child := contains32(nodeKey, searchKey)
	if !matches {
		return nil
	}

	if !exact {
		if better := me.children[child].match(searchKey); better != nil {
			return better
		}
	}

	if !me.isActive {
		return nil
	}

	return me
}

// Size returns the number of entries in the trie
func (me *trieNode32) Size() int {
	if me == nil {
		return 0
	}
	return int(me.size)
}

// height returns the maximum height of the trie.
func (me *trieNode32) height() int {
	if me == nil {
		return 0
	}
	return int(me.h)
}

// Update updates the key / value only if the key already exists
func (me *trieNode32) Update(key Prefix, data interface{}) (newHead *trieNode32, err error) {
	return me.insert(&trieNode32{Prefix: key, Data: data}, false, true)
}

// InsertOrUpdate inserts the key / value if the key didn't previously exist.
// Otherwise, it updates the data.
func (me *trieNode32) InsertOrUpdate(key Prefix, data interface{}) (newHead *trieNode32, err error) {
	return me.insert(&trieNode32{Prefix: key, Data: data}, true, true)
}

// Insert is the public form of insert(...)
func (me *trieNode32) Insert(key Prefix, data interface{}) (newHead *trieNode32, err error) {
	return me.insert(&trieNode32{Prefix: key, Data: data}, true, false)
}

// insert adds a node into the trie and return the new root of the trie. It is
// important to note that the root of the trie can change. If the new node
// cannot be inserted, nil is returned.
func (me *trieNode32) insert(node *trieNode32, insert, update bool) (newHead *trieNode32, err error) {
	defer func() {
		if err == nil && newHead != nil {
			node.size = 1
			node.h = 1
			node.isActive = true
			newHead.setSize()
		}
	}()

	if me == nil {
		if !insert {
			return me, fmt.Errorf("the key doesn't exist to update")
		}
		return node, nil
	}

	// Test containership both ways
	trieContains, nodeContains, reversed, common, child := compare32(me.Prefix, node.Prefix)
	switch {
	case trieContains && nodeContains:
		// They have the same key
		if me.isActive && !update {
			return me, fmt.Errorf("a node with that key already exists")
		}
		if !me.isActive && !insert {
			return me, fmt.Errorf("the key doesn't exist to update")
		}
		node.children = me.children
		return node, nil

	case trieContains && !nodeContains:
		// Trie node's key contains the new node's key. Insert it recursively.
		newChild, err := me.children[child].insert(node, insert, update)
		newNode := me.makeCopy()
		if err == nil {
			newNode.children[child] = newChild
		}
		return newNode, err

	case !trieContains && nodeContains:
		if !insert {
			return me, fmt.Errorf("the key doesn't exist to update")
		}
		// New node's key contains the trie node's key. Insert new node as the parent of the trie.
		node.children[child] = me
		return node, nil

	default:
		if !insert {
			return me, fmt.Errorf("the key doesn't exist to update")
		}
		// Keys are disjoint. Create a new (inactive) parent node to join them side-by-side.
		var children [2]*trieNode32

		if (child == 1) != reversed { // (child == 1) XOR reversed
			children[0], children[1] = me, node
		} else {
			children[0], children[1] = node, me
		}

		// zero out the bits that are not in common
		bits := me.ui & ^(uint32(0xffffffff) >> common)

		return &trieNode32{
			Prefix: Prefix{
				Addr: Addr{
					ui: bits,
				},
				length: common,
			},
			children: children,
		}, nil
	}
}

// Delete is a public form of del(...) below
func (me *trieNode32) Delete(key Prefix) (newHead *trieNode32, err error) {
	return me.del(key)
}

// del removes a node into the trie given a key and returns the new root of
// the trie. It is important to note that the root of the trie can change.
func (me *trieNode32) del(key Prefix) (newHead *trieNode32, err error) {
	defer func() {
		if err == nil && newHead != nil {
			newHead.setSize()
		}
	}()

	if me == nil {
		return me, fmt.Errorf("cannot delete from a nil")
	}

	trieContains, nodeContains, _, _, child := compare32(me.Prefix, key)
	if !trieContains {
		return me, fmt.Errorf("key not found")
	}

	if !nodeContains {
		// Trie node's key contains the key. Delete recursively.
		newChild, err := me.children[child].del(key)
		if err != nil {
			return me, err
		}

		newNode := me.makeCopy()
		newNode.children[child] = newChild
		return newNode, nil
	}

	// The key matches this node exactly, delete this node
	if me.children[0] == nil {
		// At this point, it doesn't matter if it is nil or not
		return me.children[1], nil
	}
	if me.children[1] == nil {
		return me.children[0], nil
	}

	// The two children are disjoint so keep this inactive node.
	newNode := me.makeCopy()
	newNode.isActive = false
	return newNode, nil
}

// active returns whether a node represents an active prefix in the tree (true)
// or an intermediate node (false). It is safe to call on a nil pointer.
func (me *trieNode32) active() bool {
	if me == nil {
		return false
	}
	return me.isActive
}

type dataContainer struct {
	valid bool
	data  interface{}
}

func dataEqual(a, b dataContainer) bool {
	if !(a.valid && b.valid) {
		return false
	}
	// If the data stored are EqualComparable, compare it using its method.
	// This is useful to allow mapping to a more complex type (e.g.
	// netaddr.IPSet)  that is not comparable by normal means.
	switch t := a.data.(type) {
	case EqualComparable:
		return t.EqualInterface(b.data)
	default:
		return a.data == b.data
	}
}

// EqualComparable is an interface used to compare data. If the datatype you
// store implements it, it can be used to aggregate prefixes.
type EqualComparable interface {
	EqualInterface(interface{}) bool
}

// aggregable returns if descendants can be aggregated into the current prefix,
// it considers the `isActive` attributes of all nodes under consideration and
// only aggregates where active nodes can be joined together in aggregation. It
// also only aggregates nodes whose data compares equal.
//
// returns true and the data used to compare with if they are aggregable, false
// otherwise (and data must be ignored).
func (me *trieNode32) aggregable(data dataContainer) (bool, dataContainer) {
	// Note that me != nil by design

	if me.isActive {
		return true, dataContainer{valid: true, data: me.Data}
	}

	// Thoughts on aggregation.
	//
	// If a parent node's data compares equal to that of descendent nodes, then
	// the descendent nodes should not be included in the aggregation. If there
	// is an intermediate descendent between two nodes that doesn't compare
	// equal, then all of them should be included. Another way to put this is
	// that each time a descendent doesn't compare equal to its direct ancestor
	// then it should be included in the aggregation. To accomplish this, each
	// parent passes its data to its children to make the comparison.
	//
	// Aggregation gets a little more complicated when peers are considered. If
	// a node's peer has the same length prefix and compare equal then they
	// should be aggregated together. However, it should be aware of their
	// joint direct ancestor and whether they should be aggrated into the
	// ancestor as discussed above.

	// NOTE that we know that BOTH children exist since me.isActive is false. If
	// less than one child existed, the tree would have been compacted to
	// eliminate this node (me).
	left, right := me.children[0], me.children[1]
	leftAggegable, leftData := left.aggregable(data)
	rightAggegable, rightData := right.aggregable(data)

	arePeers := (me.length+1) == left.length && left.length == right.length
	if arePeers && leftAggegable && rightAggegable && dataEqual(leftData, rightData) {
		return true, leftData
	}
	return false, dataContainer{}
}

// Callback32 should return true to indicate that iteration should continue or
// false to stop it immediately.
type Callback32 func(Prefix, interface{}) bool

// Iterate walks the entire tree and calls the given function for each active
// node. The order of visiting nodes is essentially lexigraphical:
// - disjoint prefixes are visited in lexigraphical order
// - shorter prefixes are visited immediately before longer prefixes that they contain
func (me *trieNode32) Iterate(callback Callback32) bool {
	if me == nil {
		return true
	}

	if me.isActive && callback != nil {
		if !callback(me.Prefix, me.Data) {
			return false
		}
	}
	for _, child := range me.children {
		if !child.Iterate(callback) {
			return false
		}
	}
	return true
}

// aggregate is the recursive implementation for Aggregate
// `data`:     the data value from nodes above to use for equal comparison. If
//             the current node is active and its data compares different to
//             this value then its key is not aggregable with containing
//             prefixes.
// `callback`: function to call with each key/data pair found.
func (me *trieNode32) aggregate(data dataContainer, callback Callback32) bool {
	if me == nil {
		return true
	}

	aggregable, d := me.aggregable(data)
	if aggregable && !dataEqual(data, d) {
		if callback != nil {
			if !callback(me.Prefix, d.data) {
				return false
			}
		}
		for _, child := range me.children {
			if !child.aggregate(d, callback) {
				return false
			}
		}
	} else {
		// Don't visit the current node but descend to children
		for _, child := range me.children {
			if !child.aggregate(data, callback) {
				return false
			}
		}
	}
	return true
}

// Aggregate is like iterate except that it has the capability of aggregating
// prefixes that are either adjacent to each other with the same prefix length
// or contained within another prefix with a shorter length.

// Aggregation visits the minimum set of prefix/data pairs needed to return the
// same data for any longest prefix match as would be returned by the the
// original trie, non-aggregated. This can be useful, for example, to minimize
// the number of prefixes needed to install into a router's datapath to
// guarantee that all of the next hops are correct.
//
// In general, routing protocols should not aggregate and then pass on the
// aggregates to neighbors as this will likely lead to poor comparisions by
// neighboring routers who receive routes aggregated differently from different
// peers.
//
// Prefixes are only considered aggregable if their data compare equal. This is
// useful for aggregating prefixes where the next hop is the same but not where
// they're different.
func (me *trieNode32) Aggregate(callback Callback32) bool {
	return me.aggregate(dataContainer{}, callback)
}
