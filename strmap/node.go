package strmap

import (
	"fmt"
	"math/bits"
)

// trieNode is a node in the binary patricia trie where keys are bit strings.
// For active (user-visible) nodes, Key is the user's string and keyBits equals
// len(Key)*8. Inactive intermediate nodes may have keyBits that is not a
// multiple of 8; their Key holds the common bit-prefix bytes with the final
// byte's trailing bits zeroed out.
type trieNode struct {
	Key      string
	keyBits  int
	Data     interface{}
	size     uint32
	h        uint16
	isActive bool
	children [2]*trieNode
}

func intMax(a, b int) int {
	if a < b {
		return b
	}
	return a
}

// commonBitPrefixKey returns the key bytes for an intermediate node whose
// bit-prefix length is numBits. The trailing bits of the last byte are zeroed.
func commonBitPrefixKey(key string, numBits int) string {
	byteLen := (numBits + 7) / 8
	b := []byte(key[:byteLen])
	if numBits%8 != 0 {
		b[byteLen-1] &= byte(0xff) << (8 - numBits%8)
	}
	return string(b)
}

// containsStrBits checks if shorter (shorterBits significant bits) is a
// bit-prefix of longer (longerBits significant bits).
//
// `matches`: true iff shorter is a bit-prefix of longer
// `exact`:   true iff shorter == longer (implies matches)
// `common`:  number of leading bits that shorter and longer share
// `child`:   0 or 1, the routing direction for longer at bit position common
func containsStrBits(shorter string, shorterBits int, longer string, longerBits int) (matches, exact bool, common int, child int) {
	// Compare full bytes of shorter.
	fullBytes := shorterBits / 8
	byteCommon := 0
	for byteCommon < fullBytes && byteCommon < len(longer) && shorter[byteCommon] == longer[byteCommon] {
		byteCommon++
	}

	if byteCommon < fullBytes {
		// Bytes diverge before we finish shorter's full bytes.
		common = byteCommon * 8
		if byteCommon < len(longer) {
			xor := shorter[byteCommon] ^ longer[byteCommon]
			common += bits.LeadingZeros8(xor)
		}
		bytePos := common / 8
		bitOff := uint(common % 8)
		if bytePos < len(longer) && longer[bytePos]&(0x80>>bitOff) != 0 {
			child = 1
		}
		return // matches = false
	}

	common = fullBytes * 8

	// Check partial last byte of shorter, if any.
	partBits := shorterBits % 8
	if partBits > 0 {
		b := fullBytes
		if b >= len(longer) {
			return // longer too short; matches = false
		}
		mask := byte(0xff) << (8 - partBits)
		if shorter[b]&mask != longer[b]&mask {
			xor := (shorter[b] ^ longer[b]) & mask
			common += bits.LeadingZeros8(xor)
			bytePos := common / 8
			bitOff := uint(common % 8)
			if bytePos < len(longer) && longer[bytePos]&(0x80>>bitOff) != 0 {
				child = 1
			}
			return // matches = false
		}
		common += partBits
	}

	matches = true
	exact = shorterBits == longerBits

	if !exact {
		bytePos := common / 8
		bitOff := uint(common % 8)
		var pivotByte byte
		if bytePos < len(longer) {
			pivotByte = longer[bytePos]
		}
		if pivotByte&(0x80>>bitOff) != 0 {
			child = 1
		}
	}
	return
}

const (
	compareSame        int = iota
	compareContains        // Second key is contained by (has more bits than) the first
	compareIsContained     // Second key contains (has fewer bits than) the first
	compareDisjoint
)

// compareStr compares the keys of two trie nodes to find their relationship.
func compareStr(a, b *trieNode) (result int, reversed bool, common int, child int) {
	var aMatch, bMatch bool
	reversed = b.keyBits < a.keyBits
	if reversed {
		bMatch, aMatch, common, child = containsStrBits(b.Key, b.keyBits, a.Key, a.keyBits)
	} else {
		aMatch, bMatch, common, child = containsStrBits(a.Key, a.keyBits, b.Key, b.keyBits)
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
	mutated := doppelganger.mutate(mutator)
	if *mutated == *me {
		return me
	}
	return mutated
}

type comparator func(a, b interface{}) bool

// Equal returns true if all entries are the same in both tries.
func (me *trieNode) Equal(other *trieNode, eq comparator) bool {
	switch {
	case me == other:
		return true
	case me == nil:
		return false
	case other == nil:
		return false
	case me.isActive != other.isActive:
		return false
	case me.keyBits != other.keyBits:
		return false
	case me.Key != other.Key:
		return false
	case me.isActive && !eq(me.Data, other.Data):
		return false
	case !me.children[0].Equal(other.children[0], eq):
		return false
	case !me.children[1].Equal(other.children[1], eq):
		return false
	default:
		return true
	}
}

// GetOrInsert returns the existing node for searchKey if found; otherwise
// inserts the given default value and returns that node.
func (me *trieNode) GetOrInsert(searchKey string, data interface{}) (head, result *trieNode) {
	searchBits := len(searchKey) * 8
	defer func() {
		if result == nil {
			result = &trieNode{Key: searchKey, keyBits: searchBits, Data: data}
			var err error
			head, err = me.insert(result, insertOpts{insert: true})
			if err != nil {
				panic(fmt.Errorf("this error shouldn't happen: %w", err))
			}
		}
	}()

	if me == nil || searchBits < me.keyBits {
		return
	}

	matches, exact, _, child := containsStrBits(me.Key, me.keyBits, searchKey, searchBits)
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

// Match returns the active node with the longest key that is a bit-prefix of
// searchKey, or nil if none match.
func (me *trieNode) Match(searchKey string) *trieNode {
	if me == nil {
		return nil
	}

	searchBits := len(searchKey) * 8
	if searchBits < me.keyBits {
		return nil
	}

	matches, exact, _, child := containsStrBits(me.Key, me.keyBits, searchKey, searchBits)
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

// NumNodes returns the number of active entries in the trie.
func (me *trieNode) NumNodes() int64 {
	if me == nil {
		return 0
	}
	return int64(me.size)
}

func (me *trieNode) height() int {
	if me == nil {
		return 0
	}
	return int(me.h)
}

// isValid reports whether the trie structure is internally consistent.
func (me *trieNode) isValid() bool {
	return me.isValidBits(0)
}

func (me *trieNode) isValidBits(minBits int) bool {
	if me == nil {
		return true
	}
	left, right := me.children[0], me.children[1]
	size := me.size
	if me.isActive {
		size--
	} else {
		if left == nil || right == nil {
			return false
		}
	}
	if size != uint32(left.NumNodes()+right.NumNodes()) {
		return false
	}
	if me.h != 1+uint16(intMax(left.height(), right.height())) {
		return false
	}
	if me.keyBits < minBits {
		return false
	}
	return left.isValidBits(me.keyBits+1) && right.isValidBits(me.keyBits+1)
}

// Update updates the value for key only if it already exists.
func (me *trieNode) Update(key string, data interface{}, eq comparator) (newHead *trieNode, err error) {
	return me.insert(&trieNode{Key: key, keyBits: len(key) * 8, Data: data}, insertOpts{update: true, eq: eq})
}

// InsertOrUpdate inserts the key/value if absent, or updates it if present.
func (me *trieNode) InsertOrUpdate(key string, data interface{}, eq comparator) *trieNode {
	newHead, err := me.insert(&trieNode{Key: key, keyBits: len(key) * 8, Data: data}, insertOpts{insert: true, update: true, eq: eq})
	if err != nil {
		panic(fmt.Errorf("this error shouldn't happen: %w", err))
	}
	return newHead
}

// Insert inserts a new key/value, failing if the key already exists.
func (me *trieNode) Insert(key string, data interface{}) (newHead *trieNode, err error) {
	return me.insert(&trieNode{Key: key, keyBits: len(key) * 8, Data: data}, insertOpts{insert: true})
}

type insertOpts struct {
	insert, update bool
	eq             comparator
}

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

	result, reversed, common, child := compareStr(me, node)
	switch result {
	case compareSame:
		if me.isActive && !opts.update {
			return me, fmt.Errorf("a node with that key already exists")
		}
		if !me.isActive && !opts.insert {
			return me, fmt.Errorf("the key doesn't exist to update")
		}
		return node.mutate(func(n *trieNode) {
			if me.isActive && opts.eq != nil && opts.eq(me.Data, node.Data) {
				node.Data = me.Data
			}
			n.children = me.children
			n.isActive = true
		}), nil

	case compareContains:
		newChild, err := me.children[child].insert(node, opts)
		if err != nil {
			return me, err
		}
		return me.copyMutate(func(n *trieNode) {
			n.children[child] = newChild
		}), nil

	case compareIsContained:
		if !opts.insert {
			return me, fmt.Errorf("the key doesn't exist to update")
		}
		node = node.mutate(func(n *trieNode) {
			n.children[child] = me
			n.isActive = true
		})
		return node, nil

	case compareDisjoint:
		var newChild *trieNode
		newChild, err = newChild.insert(node, opts)
		if err != nil {
			return me, err
		}

		var children [2]*trieNode
		if (child == 1) != reversed {
			children[0], children[1] = me, newChild
		} else {
			children[0], children[1] = newChild, me
		}

		newNode := &trieNode{
			Key:      commonBitPrefixKey(me.Key, common),
			keyBits:  common,
			children: children,
		}
		newNode.mutate(func(*trieNode) {})
		return newNode, nil
	}
	panic("unreachable code")
}

func reverseChild(child int) int {
	return (child + 1) % 2
}

// Delete removes the node with the given key and returns the new root.
func (me *trieNode) Delete(key string) (newHead *trieNode, err error) {
	if me == nil {
		return me, fmt.Errorf("cannot delete from a nil")
	}

	result, _, _, child := compareStr(me, &trieNode{Key: key, keyBits: len(key) * 8})
	switch result {
	case compareSame:
		if me.children[0] == nil {
			return me.children[1], nil
		}
		if me.children[1] == nil {
			return me.children[0], nil
		}
		return me.copyMutate(func(n *trieNode) {
			n.isActive = false
			n.Data = nil
		}), nil

	case compareContains:
		newChild, err := me.children[child].Delete(key)
		if err != nil {
			return me, err
		}
		if newChild == nil && !me.isActive {
			return me.children[reverseChild(child)], nil
		}
		return me.copyMutate(func(n *trieNode) {
			n.children[child] = newChild
		}), nil

	case compareIsContained:
		return me, fmt.Errorf("key not found")

	case compareDisjoint:
		return me, fmt.Errorf("key not found")
	}
	panic("unreachable code")
}

func (me *trieNode) active() bool {
	if me == nil {
		return false
	}
	return me.isActive
}

// Walk visits all active nodes in lexicographical order.
func (me *trieNode) Walk(callback func(string, interface{}) bool) bool {
	if callback == nil {
		callback = func(string, interface{}) bool { return true }
	}

	var empty *trieNode
	handler := trieDiffHandler{
		Added: func(n *trieNode) bool {
			return callback(n.Key, n.Data)
		},
	}
	return empty.Diff(me, handler, func(a, b interface{}) bool { return false })
}

type trieDiffHandler struct {
	Removed  func(left *trieNode) bool
	Added    func(right *trieNode) bool
	Modified func(left, right *trieNode) bool
	Same     func(common *trieNode) bool
}

func (left *trieNode) diff(right *trieNode, handler trieDiffHandler) bool {
	if left == right && handler.Same == nil {
		return true
	}

	var result, child int
	switch {
	case left != nil && right != nil:
		result, _, _, child = compareStr(left, right)
	case left != nil:
		result = compareContains
	case right != nil:
		result = compareIsContained
	default:
		return true
	}

	switch result {
	case compareSame:
		if !handler.Modified(left, right) {
			return false
		}
	case compareContains:
		if !handler.Removed(left) {
			return false
		}
	case compareIsContained:
		if !handler.Added(right) {
			return false
		}
	}

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
		if child == 0 {
			newLeft[1] = left
			newRight[0] = right
		} else {
			newLeft[0] = left
			newRight[1] = right
		}
	}

	if !newLeft[0].diff(newRight[0], handler) {
		return false
	}
	return newLeft[1].diff(newRight[1], handler)
}

// Diff compares two tries, calling handler callbacks for removed, added,
// modified, and unchanged entries.
func (left *trieNode) Diff(right *trieNode, handler trieDiffHandler, eq comparator) bool {
	noop := func(*trieNode) bool { return true }
	common := noop

	if handler.Removed == nil {
		handler.Removed = noop
	}
	if handler.Added == nil {
		handler.Added = noop
	}
	if handler.Modified == nil {
		handler.Modified = func(l, r *trieNode) bool { return true }
	}
	if handler.Same != nil {
		common = handler.Same
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
				if !eq(left.Data, right.Data) {
					return handler.Modified(left, right)
				}
				return common(left)
			case left.isActive:
				return handler.Removed(left)
			case right.isActive:
				return handler.Added(right)
			}
			return true
		},
		Same: handler.Same,
	})
}

// Map runs mapper over every active node and returns the new trie root.
func (me *trieNode) Map(mapper func(string, interface{}) interface{}, eq comparator) *trieNode {
	return me.copyMutate(func(n *trieNode) {
		n.Data = mapper(me.Key, me.Data)
		if eq(me.Data, n.Data) {
			n.Data = me.Data
		}
		n.children = [2]*trieNode{
			me.children[0].Map(mapper, eq),
			me.children[1].Map(mapper, eq),
		}
	})
}
