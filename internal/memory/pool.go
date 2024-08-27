package memory

import (
	"sync"

	"github.com/apache/arrow/go/v17/arrow/memory"
)

var memPool sync.Pool

func init() {
	memPool = sync.Pool{
		New: func() interface{} {
			// This creates a new GoAllocator when the pool is empty
			return memory.NewGoAllocator()
		},
	}
}

// getAllocator retrieves an allocator from the pool
func getAllocator() memory.Allocator {
	// Get an allocator from the pool, or create a new one if the pool is empty
	return memPool.Get().(memory.Allocator)
}

// putAllocator returns an allocator back to the pool
func putAllocator(alloc memory.Allocator) {
	// Reset or clean up the allocator if necessary before putting it back
	memPool.Put(alloc)
}

// GetAllocator is a public function to retrieve an allocator from the pool
func GetAllocator() memory.Allocator {
	return getAllocator()
}

// PutAllocator is a public function to return an allocator back to the pool
func PutAllocator(alloc memory.Allocator) {
	putAllocator(alloc)
}

// NewGoAllocator creates a new Go allocator without using the pool
func NewGoAllocator() memory.Allocator {
	return memory.NewGoAllocator()
}

// NewAllocator returns the default allocator, which in this case is also a GoAllocator
func NewAllocator() memory.Allocator {
	return memory.DefaultAllocator
}