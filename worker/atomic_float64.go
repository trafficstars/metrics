package metricworker

import (
	"math"
	"sync/atomic"
)

type AtomicFloat64 uint64

func (f *AtomicFloat64) Get() float64 {
	return math.Float64frombits(atomic.LoadUint64((*uint64)(f)))
}

func (f *AtomicFloat64) Set(n float64) {
	atomic.StoreUint64((*uint64)(f), math.Float64bits(n))
}
