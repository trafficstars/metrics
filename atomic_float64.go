package metrics

import (
	"math"
	"sync/atomic"
	"unsafe"
)

// AtomicFloat64Interface is an interface of an atomic float64 implementation
// It's an abstraction over (*AtomicFloat64) and (*AtomicFloat64Ptr)
type AtomicFloat64Interface interface {
	// Get returns the current value
	Get() float64

	// Set sets a new value
	Set(float64)

	// Add adds the value to the current one (operator "plus")
	Add(float64) float64

	// GetFast is like Get but without atomicity
	GetFast() float64

	// SetFast is like Set but without atomicity
	SetFast(float64)

	// AddFast is like Add but without atomicity
	AddFast(float64) float64
}

// AtomicFloat64 is an implementation of atomic float64 using uint64 atomic instructions
// and `math.Float64frombits()`/`math.Float64bits()`
type AtomicFloat64 uint64

// Get returns the current value
func (f *AtomicFloat64) Get() float64 {
	return math.Float64frombits(atomic.LoadUint64((*uint64)(f)))
}

// Set sets a new value
func (f *AtomicFloat64) Set(n float64) {
	atomic.StoreUint64((*uint64)(f), math.Float64bits(n))
}

// Add adds the value to the current one (operator "plus")
func (f *AtomicFloat64) Add(a float64) float64 {
	for {
		// Get the old value
		o := atomic.LoadUint64((*uint64)(f))

		// Calculate the sum
		s := math.Float64frombits(o) + a

		// Get int64 representation of the sum to be able to use atomic operations
		n := math.Float64bits(s)

		// Swap the old value to the new one
		// If not successful then somebody changes the value while our calculations above
		// It means we need to recalculate the new value and try again (that's why it's in the loop)
		if atomic.CompareAndSwapUint64((*uint64)(f), o, n) {
			return s
		}
	}
}

// GetFast is like Get but without atomicity
func (f *AtomicFloat64) GetFast() float64 {
	return math.Float64frombits(*(*uint64)(f))
}

// SetFast is like Set but without atomicity
func (f *AtomicFloat64) SetFast(n float64) {
	*(*uint64)(f) = math.Float64bits(n)
}

// AddFast is like Add but without atomicity
func (f *AtomicFloat64) AddFast(n float64) float64 {
	s := math.Float64frombits(*(*uint64)(f)) + n
	*(*uint64)(f) = math.Float64bits(s)
	return s
}

// AtomicFloat64Ptr is like AtomicFloat64 but stores the value in a pointer "*float64".
type AtomicFloat64Ptr struct {
	Pointer *float64
}

// Get returns the current value
func (f *AtomicFloat64Ptr) Get() float64 {
	return math.Float64frombits(atomic.LoadUint64((*uint64)((unsafe.Pointer)(f.Pointer))))
}

// Set sets a new value
func (f *AtomicFloat64Ptr) Set(n float64) {
	atomic.StoreUint64((*uint64)((unsafe.Pointer)(f.Pointer)), math.Float64bits(n))
}

// Add adds the value to the current one (operator "plus")
func (f *AtomicFloat64Ptr) Add(n float64) float64 {
	for {
		// Get the old value
		a := atomic.LoadUint64((*uint64)((unsafe.Pointer)(f.Pointer)))

		// Calculate the sum
		s := math.Float64frombits(a) + n

		// Get int64 representation of the sum to be able to use atomic operations
		b := math.Float64bits(s)

		// Swap the old value to the new one
		// If not successful then somebody changes the value while our calculations above
		// It means we need to recalculate the new value and try again (that's why it's in the loop)
		if atomic.CompareAndSwapUint64((*uint64)((unsafe.Pointer)(f.Pointer)), a, b) {
			return s
		}
	}
}

// GetFast is like Get but without atomicity
func (f *AtomicFloat64Ptr) GetFast() float64 {
	return *f.Pointer
}

// SetFast is like Set but without atomicity
func (f *AtomicFloat64Ptr) SetFast(n float64) {
	*f.Pointer = n
}

// AddFast is like Add but without atomicity
func (f *AtomicFloat64Ptr) AddFast(n float64) float64 {
	*f.Pointer += n
	return *f.Pointer
}
