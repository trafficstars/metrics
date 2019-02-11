package metrics

import (
	"math"
	"sync/atomic"
	"unsafe"
)

type AtomicFloat64Interface interface {
	Get() float64
	Set(float64)
	Add(float64) float64

	// GetFast is like Get but without atomicity
	GetFast() float64

	// SetFast is like Set but without atomicity
	SetFast(float64)

	// AddFast is like Add but without atomicity
	AddFast(float64) float64
}

type AtomicFloat64 uint64

func (f *AtomicFloat64) Get() float64 {
	return math.Float64frombits(atomic.LoadUint64((*uint64)(f)))
}

func (f *AtomicFloat64) Set(n float64) {
	atomic.StoreUint64((*uint64)(f), math.Float64bits(n))
}

func (f *AtomicFloat64) Add(n float64) float64 {
	for {
		a := atomic.LoadUint64((*uint64)(f))
		s := math.Float64frombits(a) + n
		b := math.Float64bits(s)
		if atomic.CompareAndSwapUint64((*uint64)(f), a, b) {
			return s
		}
	}
}

func (f *AtomicFloat64) GetFast() float64 {
	return math.Float64frombits(*(*uint64)(f))
}

func (f *AtomicFloat64) SetFast(n float64) {
	*(*uint64)(f) = math.Float64bits(n)
}

func (f *AtomicFloat64) AddFast(n float64) float64 {
	s := math.Float64frombits(*(*uint64)(f)) + n
	*(*uint64)(f) = math.Float64bits(s)
	return s
}

type AtomicFloat64Ptr struct {
	Pointer *float64
}

func (f *AtomicFloat64Ptr) Get() float64 {
	return math.Float64frombits(atomic.LoadUint64((*uint64)((unsafe.Pointer)(f.Pointer))))
}

func (f *AtomicFloat64Ptr) Set(n float64) {
	atomic.StoreUint64((*uint64)((unsafe.Pointer)(f.Pointer)), math.Float64bits(n))
}

func (f *AtomicFloat64Ptr) Add(n float64) float64 {
	for {
		a := atomic.LoadUint64((*uint64)((unsafe.Pointer)(f.Pointer)))
		s := math.Float64frombits(a) + n
		b := math.Float64bits(s)
		if atomic.CompareAndSwapUint64((*uint64)((unsafe.Pointer)(f.Pointer)), a, b) {
			return s
		}
	}
}

func (f *AtomicFloat64Ptr) GetFast() float64 {
	return *f.Pointer
}

func (f *AtomicFloat64Ptr) SetFast(n float64) {
	*f.Pointer = n
}

func (f *AtomicFloat64Ptr) AddFast(n float64) float64 {
	*f.Pointer += n
	return *f.Pointer
}
