package ipv6

import (
	"sync/atomic"
	"unsafe"
)

func swapSetNodePtr(ptr **setNode, old, new *setNode) bool {
	return atomic.CompareAndSwapPointer(
		(*unsafe.Pointer)(
			unsafe.Pointer(ptr),
		),
		unsafe.Pointer(old),
		unsafe.Pointer(new),
	)
}
