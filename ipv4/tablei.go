package ipv4

import (
	"sync/atomic"
	"unsafe"
)

// TableI is a structure that tables IP prefixes to values. For example, you can
// insert the following values and they will all exist as distinct prefix/value
// pairs in the table.
//
// 10.0.0.0/16 -> 1
// 10.0.0.0/24 -> 1
// 10.0.0.0/32 -> 2
//
// The table supports looking up values based on a longest prefix match and also
// supports efficient aggregation of prefix/value pairs based on equality of
// values. See the README.md file for a more detailed discussion..
type TableI struct {
	// This is an abuse of FixedTableI because it uses its package privileges
	// to turn it into a mutable one. This could be refactored to be cleaner
	// without changing the interface.

	// Be careful not to take an FixedTableI from outside the package and turn
	// it into a mutable one. That would break the contract.
	m *FixedTableI
}

// NewTableI returns a new fully-initialized TableI
func NewTableI() TableI {
	return TableI{
		m: &FixedTableI{},
	}
}

// Match indicates how closely the given key matches the search result
type Match int

const (
	// MatchNone indicates that no match was found
	MatchNone Match = iota
	// MatchContains indicates that a match was found that contains the search key but isn't exact
	MatchContains
	// MatchExact indicates that a match with the same prefix
	MatchExact
)

// Size returns the number of exact prefixes stored in the table
func (me TableI) Size() int64 {
	if me.m == nil {
		return 0
	}
	return me.m.Size()
}

// mutate should be called by any method that modifies the table in any way
func (me TableI) mutate(mutator func() (ok bool, node *trieNode)) {
	oldNode := me.m.trie
	ok, newNode := mutator()
	if ok && oldNode != newNode {
		swapped := atomic.CompareAndSwapPointer(
			(*unsafe.Pointer)(
				unsafe.Pointer(&me.m.trie),
			),
			unsafe.Pointer(oldNode),
			unsafe.Pointer(newNode),
		)
		if !swapped {
			panic("concurrent modification of Table detected")
		}
	}
}

// Insert inserts the given prefix with the given value into the table.
// If an entry with the same prefix already exists, it will not overwrite it
// and return false.
func (me TableI) Insert(p PrefixI, value interface{}) (succeeded bool) {
	if me.m == nil {
		panic("cannot modify an unitialized Set")
	}
	var err error
	me.mutate(func() (bool, *trieNode) {
		var newHead *trieNode
		newHead, err = me.m.trie.Insert(p.Prefix(), value)
		if err != nil {
			return false, nil
		}
		return true, newHead
	})
	return err == nil
}

// Update inserts the given prefix with the given value into the table. If the
// prefix already existed, it updates the associated value in place and return
// true. Otherwise, it returns false.
func (me TableI) Update(p PrefixI, value interface{}) (succeeded bool) {
	if me.m == nil {
		panic("cannot modify an unitialized Set")
	}
	var err error
	me.mutate(func() (bool, *trieNode) {
		var newHead *trieNode
		newHead, err = me.m.trie.Update(p.Prefix(), value)
		if err != nil {
			return false, nil
		}
		return true, newHead
	})
	return err == nil
}

// InsertOrUpdate inserts the given prefix with the given value into the table.
// If the prefix already existed, it updates the associated value in place.
func (me TableI) InsertOrUpdate(p PrefixI, value interface{}) {
	if me.m == nil {
		panic("cannot modify an unitialized Set")
	}
	me.mutate(func() (bool, *trieNode) {
		return true, me.m.trie.InsertOrUpdate(p.Prefix(), value)
	})
}

// Get returns the value in the table associated with the given network prefix
// with an exact match: both the IP and the prefix length must match. If an
// exact match is not found, found is false and value is nil and should be
// ignored.
func (me TableI) Get(prefix PrefixI) (interface{}, bool) {
	if me.m == nil {
		return nil, false
	}
	return me.m.Get(prefix)
}

// GetOrInsert returns the value associated with the given prefix if it already
// exists. If it does not exist, it inserts it with the given value and returns
// that.
func (me TableI) GetOrInsert(p PrefixI, value interface{}) interface{} {
	if me.m == nil {
		panic("cannot modify an unitialized Set")
	}
	var node *trieNode
	me.mutate(func() (bool, *trieNode) {
		var newHead *trieNode
		newHead, node = me.m.trie.GetOrInsert(p.Prefix(), value)
		return true, newHead
	})
	return node.Data
}

// LongestMatch returns the value in the table associated with the given network
// prefix using a longest prefix match. If a match is found, it returns a
// Prefix representing the longest prefix matched. If a match is *not* found,
// matched is MatchNone and the other fields should be ignored
func (me TableI) LongestMatch(searchPrefix PrefixI) (value interface{}, matched Match, prefix Prefix) {
	if me.m == nil {
		return nil, MatchNone, Prefix{}
	}
	return me.m.LongestMatch(searchPrefix)
}

// Remove removes the given prefix from the table with its associated value and
// returns true if it was found. Only a prefix with an exact match will be
// removed. If no entry with the given prefix exists, it will do nothing and
// return false.
func (me TableI) Remove(p PrefixI) (succeeded bool) {
	if me.m == nil {
		panic("cannot modify an unitialized Set")
	}
	var err error
	me.mutate(func() (bool, *trieNode) {
		var newHead *trieNode
		newHead, err = me.m.trie.Delete(p.Prefix())
		return true, newHead
	})
	return err == nil
}

// Walk invokes the given callback function for each prefix/value pair in
// the table in lexigraphical order.
//
// It returns false if iteration was stopped due to a callback return false or
// true if it iterated all items.
func (me TableI) Walk(callback TableICallback) bool {
	if me.m == nil {
		return true
	}
	return me.m.Walk(callback)
}

// WalkAggregates invokes then given callback function for each prefix/value
// pair in the table, aggregated by value, in lexigraphical order.
//
// If two prefixes table to the same value, one contains the other, and there is
// no intermediate prefix between the two with a different value then only the
// broader prefix will be visited with the value.
//
// 1. The values stored must be comparable to be aggregable. Prefixes get
//    aggregated only where their values compare equal.
// 2. The set of prefix/value pairs visited is the minimal set such that any
//    longest prefix match against the aggregated set will always return the
//    same value as the same match against the non-aggregated set.
// 3. The aggregated and non-aggregated sets of prefixes may be disjoint.
//
// It returns false if iteration was stopped due to a callback return false or
// true if it iterated all items.
func (me TableI) WalkAggregates(callback TableICallback) bool {
	if me.m == nil {
		return true
	}
	return me.m.WalkAggregates(callback)
}

// FixedTable returns an immutable snapshot of this TableI. Due to the COW
// nature of the underlying datastructure, it is very cheap to create these --
// effectively a pointer copy.
func (me TableI) FixedTable() FixedTableI {
	if me.m == nil {
		return FixedTableI{}
	}
	return FixedTableI{
		trie: me.m.trie,
	}
}
