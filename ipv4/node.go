package ipv4

import (
	"fmt"
	"math/bits"
)

type trieNode struct {
	Prefix   Prefix
	Data     interface{}
	size     uint32
	h        uint16
	isActive bool
	children [2]*trieNode
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
func contains(shorter, longer Prefix) (matches, exact bool, common uint32, child int) {
	mask := uint32(0xffffffff) << (32 - shorter.length)

	matches = shorter.addr.ui&mask == longer.addr.ui&mask
	if matches {
		exact = shorter.length == longer.length
		common = shorter.length
	} else {
		common = uint32(bits.LeadingZeros32(shorter.addr.ui ^ longer.addr.ui))
	}
	if !exact {
		// Whether `longer` goes on the left (0) or right (1)
		pivotMask := uint32(0x80000000) >> common
		if longer.addr.ui&pivotMask != 0 {
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

func (me *trieNode) mutate(mutator func(*trieNode)) *trieNode {
	if me == nil {
		return nil
	}

	mutator(me)

	numNodes := me.children[0].NumNodes() + me.children[1].NumNodes()
	height := 1 + intMax(me.children[0].height(), me.children[1].height())

	me.size = uint32(numNodes)
	me.h = uint16(height)
	if me.isActive {
		me.size++
	}
	return me
}

func (me *trieNode) copyMutate(mutator func(*trieNode)) *trieNode {
	if me == nil {
		return nil
	}
	doppelganger := &trieNode{}
	*doppelganger = *me
	return doppelganger.mutate(mutator)
}

// Equal returns true if all of the entries are the same in the two data structures
func (me *trieNode) Equal(other *trieNode) bool {
	switch {
	case me == other:
		return true

	case me == nil:
		return false
	case other == nil:
		return false
	case me.isActive != other.isActive:
		return false
	case !me.Prefix.Equal(other.Prefix):
		return false
	case me.isActive && !dataEqual(me.Data, other.Data):
		return false
	case !me.children[0].Equal(other.children[0]):
		return false
	case !me.children[1].Equal(other.children[1]):
		return false

	default:
		return true
	}
}

// GetOrInsert returns the existing value if an exact match is found, otherwise, inserts the given default
func (me *trieNode) GetOrInsert(searchKey Prefix, data interface{}) (head, result *trieNode) {
	defer func() {
		if result == nil {
			result = &trieNode{Prefix: searchKey, Data: data}

			var err error
			head, err = me.insert(result, insertOpts{insert: true})
			if err != nil {
				// when getting *or* inserting, we design around the errors that could come from insert
				panic(fmt.Errorf("this error shouldn't happen: %w", err))
			}
		}
	}()

	if me == nil || searchKey.length < me.Prefix.length {
		return
	}

	matches, exact, _, child := contains(me.Prefix, searchKey)
	if !matches {
		return
	}

	if !exact {
		var newChild *trieNode
		newChild, result = me.children[child].GetOrInsert(searchKey, data)

		head = me.copyMutate(func(n *trieNode) {
			n.children[child] = newChild
		})
		return
	}

	if !me.isActive {
		return
	}

	return me, me
}

// Match returns the existing entry with the longest prefix that fully contains
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
func (me *trieNode) Match(searchKey Prefix) *trieNode {
	if me == nil {
		return nil
	}

	nodeKey := me.Prefix
	if searchKey.length < nodeKey.length {
		return nil
	}

	matches, exact, _, child := contains(nodeKey, searchKey)
	if !matches {
		return nil
	}

	if !exact {
		if better := me.children[child].Match(searchKey); better != nil {
			return better
		}
	}

	if !me.isActive {
		return nil
	}

	return me
}

// Size returns the number of addresses that could match this node
// Note that this may have to search all nodes recursively to find the answer.
// The implementation can be changed to store the size in each node at the cost
// of adding about 8 more bytes to the size of each node.
func (me *trieNode) Size() int64 {
	if me == nil {
		return 0
	}
	if me.isActive {
		return int64(^me.Prefix.Mask().Uint32()) + 1
	}
	return me.children[0].Size() + me.children[1].Size()
}

// NumNodes returns the number of entries in the trie
func (me *trieNode) NumNodes() int64 {
	if me == nil {
		return 0
	}
	return int64(me.size)
}

// height returns the maximum height of the trie.
func (me *trieNode) height() int {
	if me == nil {
		return 0
	}
	return int(me.h)
}

// isValid returns true if the tree is valid
// this method is only for unit tests to check the integrity of the structure
func (me *trieNode) isValid() bool {
	return me.isValidLen(0)
}

func (me *trieNode) isValidLen(minLen uint32) bool {
	if me == nil {
		return true
	}
	left, right := me.children[0], me.children[1]
	size := me.size
	if me.isActive {
		size--
	} else {
		if left == nil || right == nil {
			// Any child node should have been pulled up since this node isn't active
			return false
		}
	}
	if size != uint32(left.NumNodes()+right.NumNodes()) {
		return false
	}
	if me.h != 1+uint16(uint16(intMax(left.height(), right.height()))) {
		return false
	}
	if me.Prefix.length < minLen {
		return false
	}
	return left.isValidLen(me.Prefix.length+1) && right.isValidLen(me.Prefix.length+1)
}

// Update updates the key / value only if the key already exists
func (me *trieNode) Update(key Prefix, data interface{}) (newHead *trieNode, err error) {
	return me.insert(&trieNode{Prefix: key, Data: data}, insertOpts{update: true})
}

// InsertOrUpdate inserts the key / value if the key didn't previously exist.
// Otherwise, it updates the data.
func (me *trieNode) InsertOrUpdate(key Prefix, data interface{}) (newHead *trieNode) {
	var err error
	newHead, err = me.insert(&trieNode{Prefix: key, Data: data}, insertOpts{insert: true, update: true})
	if err != nil {
		// when inserting *or* updating, we design around the errors that could come from insert
		panic(fmt.Errorf("this error shouldn't happen: %w", err))
	}
	return newHead
}

// Insert is the public form of insert(...)
func (me *trieNode) Insert(key Prefix, data interface{}) (newHead *trieNode, err error) {
	return me.insert(&trieNode{Prefix: key, Data: data}, insertOpts{insert: true})
}

type insertOpts struct {
	insert, update, flatten bool
}

// flatten assumes that `me` is a new node. It should not be called that had
// already existed as a node in the trie because it does not make a copy.
func (me *trieNode) flatten() {
	if me.isActive {
		// If the current node is active, then anything referenced by the
		// children is redundant, they can be removed.
		me.children = [2]*trieNode{}
		return
	}
	left, right := me.children[0], me.children[1]
	if left == nil || right == nil {
		panic("this should never happen; it means that the structure is not optimized")
	}
	if left.Prefix.length != right.Prefix.length {
		// If the childen have different size prefixes, then we cannot combine
		// them. Do nothing.
		return
	}
	if left.Prefix.length != me.Prefix.length+1 {
		// If the children aren't exactly half the current node's prefix then
		// we cannot combine them. Do nothing.
		return
	}
	if !left.isActive || !right.isActive {
		// If the children aren't both active, it means they are sparse and
		// cannot be combined. Do nothing.
		return
	}
	me.children = [2]*trieNode{}
	me.isActive = true
}

// insert adds a node into the trie and return the new root of the trie. It is
// important to note that the root of the trie can change. If the new node
// cannot be inserted, nil is returned.
func (me *trieNode) insert(node *trieNode, opts insertOpts) (newHead *trieNode, err error) {
	if me == nil {
		if !opts.insert {
			return me, fmt.Errorf("the key doesn't exist to update")
		}
		node = node.mutate(func(n *trieNode) {
			n.isActive = true
		})
		return node, nil
	}

	// Test containership both ways
	result, reversed, common, child := compare(me.Prefix, node.Prefix)
	switch result {
	case compareSame:
		// They have the same key
		if me.isActive && !opts.update {
			return me, fmt.Errorf("a node with that key already exists")
		}
		if !me.isActive && !opts.insert {
			return me, fmt.Errorf("the key doesn't exist to update")
		}
		if opts.flatten {
			// avoid copy-on-write when it will be flattened resulting in no effective change
			return me, nil
		}
		node = node.mutate(func(n *trieNode) {
			n.children = me.children
			n.isActive = true
			if opts.flatten {
				n.flatten()
			}
		})
		return node, nil

	case compareContains:
		// Trie node's key contains the new node's key. Insert it recursively.
		if opts.flatten && me.isActive {
			// avoid copy-on-write when it will be flattened resulting in no effective change
			return me, nil
		}
		newChild, err := me.children[child].insert(node, opts)
		if err != nil {
			return me, err
		}
		newNode := me.copyMutate(func(n *trieNode) {
			n.children[child] = newChild
			if opts.flatten {
				n.flatten()
			}
		})
		return newNode, nil

	case compareIsContained:
		// New node's key contains the trie node's key. Insert new node as the parent of the trie.
		if !opts.insert {
			return me, fmt.Errorf("the key doesn't exist to update")
		}
		node = node.mutate(func(n *trieNode) {
			n.children[child] = me
			n.isActive = true
			if opts.flatten {
				n.flatten()
			}
		})
		return node, nil

	case compareDisjoint:
		// Keys are disjoint. Create a new (inactive) parent node to join them side-by-side.
		var newChild *trieNode
		newChild, err := newChild.insert(node, opts)
		if err != nil {
			return me, err
		}

		var children [2]*trieNode

		if (child == 1) != reversed { // (child == 1) XOR reversed
			children[0], children[1] = me, newChild
		} else {
			children[0], children[1] = newChild, me
		}

		newNode := &trieNode{
			Prefix: Prefix{
				addr: Address{
					ui: me.Prefix.addr.ui & ^(uint32(0xffffffff) >> common), // zero out bits not in common
				},
				length: common,
			},
			children: children,
		}
		newNode.mutate(func(n *trieNode) {
			if opts.flatten {
				n.flatten()
			}
		})
		return newNode, nil
	}
	panic("unreachable code")
}

type deleteOpts struct {
	flatten bool
}

// Delete removes a node from the trie given a key and returns the new root of
// the trie. It is important to note that the root of the trie can change.
func (me *trieNode) Delete(key Prefix) (newHead *trieNode, err error) {
	return me.del(key, deleteOpts{})
}

func reverseChild(child int) int {
	return (child + 1) % 2
}

func (me *trieNode) del(key Prefix, opts deleteOpts) (newHead *trieNode, err error) {
	if me == nil {
		if opts.flatten {
			return nil, nil
		}
		return me, fmt.Errorf("cannot delete from a nil")
	}

	result, _, _, child := compare(me.Prefix, key)
	switch result {
	case compareSame:
		if opts.flatten {
			return nil, nil
		}
		// Delete this node
		if me.children[0] == nil {
			// At this point, it doesn't matter if it is nil or not
			return me.children[1], nil
		}
		if me.children[1] == nil {
			return me.children[0], nil
		}

		// The two children are disjoint so keep this inactive node.
		newNode := me.copyMutate(func(n *trieNode) {
			n.isActive = false
		})
		return newNode, nil

	case compareContains:
		if me.isActive && opts.flatten {
			// TODO This is for sets. Break it out of here.
			// split this prefix into ranges and insert them
			super, sub := me.Prefix.Range(), key.Range()
			remainingRanges := super.Minus(sub)
			var s *setNode
			if len(remainingRanges) == 1 {
				s = setNodeFromRange(remainingRanges[0])
			} else {
				a := setNodeFromRange(remainingRanges[0])
				b := setNodeFromRange(remainingRanges[1])
				s = a.Union(b)
			}
			return (*trieNode)(s), nil
		}
		// Delete recursively.
		newChild, err := me.children[child].del(key, opts)
		if err != nil {
			return me, err
		}

		if newChild == nil && !me.isActive {
			// Promote the other child up
			return me.children[reverseChild(child)], nil
		}
		newNode := me.copyMutate(func(n *trieNode) {
			n.children[child] = newChild
		})
		return newNode, nil

	case compareIsContained:
		if opts.flatten {
			return nil, nil
		}
		return me, fmt.Errorf("key not found")

	case compareDisjoint:
		return me, fmt.Errorf("key not found")
	}
	panic("unreachable code")
}

// active returns whether a node represents an active prefix in the tree (true)
// or an intermediate node (false). It is safe to call on a nil pointer.
func (me *trieNode) active() bool {
	if me == nil {
		return false
	}
	return me.isActive
}

func dataEqual(a, b interface{}) bool {
	// If the data stored are EqualComparable, compare it using its method.
	// This is useful to allow mapping to a more complex type (e.g. Set) that
	// is not comparable by normal means.
	switch t := a.(type) {
	case EqualComparable:
		return t.EqualInterface(b)
	default:
		return a == b
	}
}

// EqualComparable is an interface used to compare data. If the datatype you
// store implements it, it can be used to aggregate prefixes.
type EqualComparable interface {
	EqualInterface(interface{}) bool
}

// trieCallback defines the signature of a function you can pass to Walk or
// Aggregate to handle each key / value pair found while iterating.
//
// Each invocation of your callback should return true if iteration should
// continue (as long as another key / value pair exists) or false to stop
// iterating and return immediately (meaning your callback will not be called
// again).
type trieCallback func(Prefix, interface{}) bool

// Walk walks the entire tree and calls the given function for each active
// node. The order of visiting nodes is essentially lexigraphical:
// - disjoint prefixes are visited in lexigraphical order
// - shorter prefixes are visited immediately before longer prefixes that they contain
//
// It returns false if iteration was stopped due to a callback return false or
// true if it iterated all items.
func (me *trieNode) Walk(callback trieCallback) bool {
	if callback == nil {
		callback = func(Prefix, interface{}) bool {
			return true
		}
	}

	var empty *trieNode
	handler := trieDiffHandler{
		Added: func(n *trieNode) bool {
			return callback(n.Prefix, n.Data)
		},
	}
	return empty.Diff(me, handler)
}

type trieDiffHandler struct {
	Removed  func(left *trieNode) bool
	Added    func(right *trieNode) bool
	Modified func(left, right *trieNode) bool
}

func (left *trieNode) diff(right *trieNode, handler trieDiffHandler) bool {
	if left == right {
		return true
	}

	// Compare the two nodes.
	// If one of them is nil, we treat it as if it is contained by the non-nil one.
	// In that case, `child` doesn't matter so we leave it initialized at zero.
	// If both are nil, there is nothing to do.
	var result, child int
	switch {
	case left != nil && right != nil:
		result, _, _, child = compare(left.Prefix, right.Prefix)

	case left != nil:
		result = compareContains

	case right != nil:
		result = compareIsContained

	default:
		return true
	}

	// Based on the comparison above, determine where to descend to child nodes
	// before recursing.
	//
	// The side that doesn't descend is included in the pair of nodes as either
	// the first (0) or second (1) element based on the comparison (child) with
	// the other side being an empty set (nil by default).
	//
	// If the two sides are disjoint, neither one descends and the two sides
	// are split apart to compare each independently with an empty set.

	var newLeft, newRight [2]*trieNode
	switch result {
	case compareSame:
		newLeft = left.children
		newRight = right.children
	case compareIsContained:
		newLeft[child] = left
		newRight = right.children
	case compareContains:
		newLeft = left.children
		newRight[child] = right
	case compareDisjoint:
		// Divide and conquer. Compare each with an empty set. Order based on
		// the comparison.
		if child == 0 {
			newLeft[1] = left
			newRight[0] = right
		} else {
			newLeft[0] = left
			newRight[1] = right
		}
	}

	// Call handlers. If the nodes are disjoint, nothing is called yet.
	switch result {
	case compareSame:
		// They have the same key
		if !handler.Modified(left, right) {
			return false
		}

	case compareContains:
		// Left node's key contains the right node's key
		if !handler.Removed(left) {
			return false
		}

	case compareIsContained:
		// Right node's key contains the left node's key
		if !handler.Added(right) {
			return false
		}
	}

	// Recurse into children
	if !newLeft[0].diff(newRight[0], handler) {
		return false
	}
	if !newLeft[1].diff(newRight[1], handler) {
		return false
	}
	return true
}

// Diff compares the two tries to find entries that are removed, added, or
// changed between the two. It calls the appropriate callback
func (left *trieNode) Diff(right *trieNode, handler trieDiffHandler) bool {
	noop := func(*trieNode) bool {
		return true
	}

	// Ensure I don't have to check for nil everywhere.
	if handler.Removed == nil {
		handler.Removed = noop
	}
	if handler.Added == nil {
		handler.Added = noop
	}
	if handler.Modified == nil {
		handler.Modified = func(l, r *trieNode) bool {
			return true
		}
	}

	return left.diff(right, trieDiffHandler{
		Removed: func(left *trieNode) bool {
			if left.isActive {
				return handler.Removed(left)
			}
			return true
		},
		Added: func(right *trieNode) bool {
			if right.isActive {
				return handler.Added(right)
			}
			return true
		},
		Modified: func(left, right *trieNode) bool {
			switch {
			case left.isActive && right.isActive:
				if !dataEqual(left.Data, right.Data) {
					return handler.Modified(left, right)
				}
			case left.isActive:
				return handler.Removed(left)
			case right.isActive:
				return handler.Added(right)
			}
			return true
		},
	})
}

type umbrella struct {
	Data interface{}
}

func (me *trieNode) aggregate(parentUmbrella *umbrella) (result *trieNode) {
	if me == nil {
		return me
	}

	isActive := me.isActive
	data := me.Data
	children := me.children
	createReturnValue := func() *trieNode {
		if isActive == me.isActive && children == me.children && dataEqual(data, me.Data) {
			return me
		}
		return me.copyMutate(func(n *trieNode) {
			n.isActive = isActive
			n.Data = data
			n.children = children
		})
	}

	u := parentUmbrella
	if isActive {
		if parentUmbrella == nil || !dataEqual(parentUmbrella.Data, data) {
			u = &umbrella{data}
		} else {
			isActive = false
			data = nil
		}
	}
	children = [2]*trieNode{
		children[0].aggregate(u),
		children[1].aggregate(u),
	}

	childrenAggregate := func(a, b *trieNode) bool {
		if a.active() && b.active() {
			if a.Prefix.Length() == (me.Prefix.Length() + 1) {
				if a.Prefix.Length() == b.Prefix.Length() {
					if dataEqual(a.Data, b.Data) {
						return true
					}
				}
			}
		}
		return false
	}(children[0], children[1])

	if childrenAggregate {
		isActive = true
		data = children[0].Data
		children = [2]*trieNode{}
	}

	if isActive {
		return createReturnValue()
	}

	if children[0] == nil {
		return children[1]
	}

	if children[1] == nil {
		return children[0]
	}

	return createReturnValue()
}

func (me *trieNode) Aggregate() *trieNode {
	return me.aggregate(nil)
}
