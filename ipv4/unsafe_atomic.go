package ipv4

import (
	"sync/atomic"
	"unsafe"
)

func swapTrieNodePtr(ptr **trieNode, old, new *trieNode) bool {
	return atomic.CompareAndSwapPointer(
		(*unsafe.Pointer)(
			unsafe.Pointer(ptr),
		),
		unsafe.Pointer(old),
		unsafe.Pointer(new),
	)
}

func swapSetNodePtr(ptr **setNode, old, new *setNode) bool {
	return atomic.CompareAndSwapPointer(
		(*unsafe.Pointer)(
			unsafe.Pointer(ptr),
		),
		unsafe.Pointer(old),
		unsafe.Pointer(new),
	)
}
