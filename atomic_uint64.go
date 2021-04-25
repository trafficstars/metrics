package metrics

import (
	"sync/atomic"
)

// AtomicUint64 is just a handy wrapper for uint64 with atomic primitives.
type AtomicUint64 uint64

// Get returns the current value
func (v *AtomicUint64) Get() uint64 {
	return atomic.LoadUint64((*uint64)(v))
}

// Set sets a new value
func (v *AtomicUint64) Set(n uint64) {
	atomic.StoreUint64((*uint64)(v), n)
}

// Add adds the value to the current one (operator "plus")
func (v *AtomicUint64) Add(a uint64) uint64 {
	return atomic.AddUint64((*uint64)(v), a)
}
