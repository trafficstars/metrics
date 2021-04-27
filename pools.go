package metrics

import (
	"bytes"
	"sync"
	"sync/atomic"
)

var (
	memoryReuse = uint64(1)
)

// MemoryReuseEnabled returns if memory reuse is enabled.
func MemoryReuseEnabled() bool {
	return atomic.LoadUint64(&memoryReuse) != 0
}

// SetMemoryReuseEnabled defines if memory reuse will be enabled (default -- enabled).
func SetMemoryReuseEnabled(isEnabled bool) {
	if isEnabled {
		atomic.StoreUint64(&memoryReuse, 1)
	} else {
		atomic.StoreUint64(&memoryReuse, 0)
	}
}

type bytesBuffer struct {
	bytes.Buffer
}

type stringSlice []string

func (p stringSlice) Len() int           { return len(p) }
func (p stringSlice) Less(i, j int) bool { return p[i] < p[j] }
func (p stringSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

var (
	bytesBufferPool = &sync.Pool{
		New: func() interface{} {
			return &bytesBuffer{}
		},
	}
	stringSlicePool = &sync.Pool{
		New: func() interface{} {
			return &stringSlice{}
		},
	}
	metricCountPool = &sync.Pool{
		New: func() interface{} {
			return &MetricCount{}
		},
	}
	metricGaugeAggregativeBufferedPool = &sync.Pool{
		New: func() interface{} {
			return &MetricGaugeAggregativeBuffered{}
		},
	}
	metricGaugeAggregativeFlowPool = &sync.Pool{
		New: func() interface{} {
			return &MetricGaugeAggregativeFlow{}
		},
	}
	metricGaugeAggregativeSimplePool = &sync.Pool{
		New: func() interface{} {
			return &MetricGaugeAggregativeSimple{}
		},
	}
	metricGaugeFloat64Pool = &sync.Pool{
		New: func() interface{} {
			return &MetricGaugeFloat64{}
		},
	}
	metricGaugeFloat64FuncPool = &sync.Pool{
		New: func() interface{} {
			return &MetricGaugeFloat64Func{}
		},
	}
	metricGaugeInt64Pool = &sync.Pool{
		New: func() interface{} {
			return &MetricGaugeInt64{}
		},
	}
	metricGaugeInt64FuncPool = &sync.Pool{
		New: func() interface{} {
			return &MetricGaugeInt64Func{}
		},
	}
	metricTimingBufferedPool = &sync.Pool{
		New: func() interface{} {
			return &MetricTimingBuffered{}
		},
	}
	metricTimingFlowPool = &sync.Pool{
		New: func() interface{} {
			return &MetricTimingFlow{}
		},
	}
	metricTimingSimplePool = &sync.Pool{
		New: func() interface{} {
			return &MetricTimingSimple{}
		},
	}
	aggregativeValuePool = &sync.Pool{
		New: func() interface{} {
			return &AggregativeValue{}
		},
	}
	aggregativeStatisticsFlowPool = &sync.Pool{
		New: func() interface{} {
			return &aggregativeStatisticsFlow{}
		},
	}
	aggregativeBufferPool = &sync.Pool{
		New: func() interface{} {
			buf := &aggregativeBuffer{}
			if uint(cap(buf.data)) < bufferSize {
				buf.data = make(aggregativeBufferItems, bufferSize)
			} else if uint(len(buf.data)) != bufferSize {
				buf.data = buf.data[:bufferSize]
			}
			return buf
		},
	}
	aggregativeStatisticsBufferedPool = &sync.Pool{
		New: func() interface{} {
			buf := &aggregativeStatisticsBuffered{}
			if uint(cap(buf.data)) < bufferSize {
				buf.data = make(aggregativeBufferItems, bufferSize)
			} else if uint(len(buf.data)) != bufferSize {
				buf.data = buf.data[:bufferSize]
			}
			return buf
		},
	}
	iterationHandlerPool = &sync.Pool{
		New: func() interface{} {
			iterationHandler := &iterationHandler{
				stopChan: make(chan struct{}),
			}
			return iterationHandler
		},
	}
	metricsPool = &sync.Pool{
		New: func() interface{} {
			return &Metrics{}
		},
	}
)

func (s *Metrics) Release() {
	if !MemoryReuseEnabled() {
		return
	}
	*s = (*s)[:0]
	metricsPool.Put(s)
}

func newBytesBuffer() *bytesBuffer {
	return bytesBufferPool.Get().(*bytesBuffer)
}

func (buf *bytesBuffer) Release() {
	if !MemoryReuseEnabled() {
		return
	}
	buf.Reset()
	bytesBufferPool.Put(buf)
}

func newStringSlice() *stringSlice {
	return stringSlicePool.Get().(*stringSlice)
}

func (s *stringSlice) Release() {
	if !MemoryReuseEnabled() {
		return
	}
	*s = (*s)[:0]
	stringSlicePool.Put(s)
}

// Release should be called when the buffer won't be used anymore (to put into into the pool of free buffers) to
// reduce pressure on GC.
func (s *aggregativeStatisticsBuffered) Release() {
	if !MemoryReuseEnabled() {
		return
	}
	s.filledSize = 0
	s.tickID = 0
	aggregativeStatisticsBufferedPool.Put(s)
}

func newAggregativeStatisticsBuffered(defaultPercentiles []float64) *aggregativeStatisticsBuffered {
	stats := aggregativeStatisticsBufferedPool.Get().(*aggregativeStatisticsBuffered)
	stats.defaultPercentiles = defaultPercentiles
	return stats
}

// Release should be called when the buffer won't be used anymore (to put into into the pool of free buffers) to
// reduce pressure on GC.
func (b *aggregativeBuffer) Release() {
	if !MemoryReuseEnabled() {
		return
	}
	b.filledSize = 0
	aggregativeBufferPool.Put(b)
}

func newAggregativeBuffer() *aggregativeBuffer {
	return aggregativeBufferPool.Get().(*aggregativeBuffer)
}

func (s *aggregativeStatisticsFlow) Release() {
	if !MemoryReuseEnabled() {
		return
	}
	s.Set(0)
	s.tickID = 0
	aggregativeStatisticsFlowPool.Put(s)
}

func newAggregativeStatisticsFlow() *aggregativeStatisticsFlow {
	return aggregativeStatisticsFlowPool.Get().(*aggregativeStatisticsFlow)
}

// Release is an opposite to NewAggregativeValue and it saves the variable to a pool to a prevent memory allocation in future.
// It's not necessary to call this method when you finished to work with an AggregativeValue, but recommended to (for better performance).
func (v *AggregativeValue) Release() {
	if !MemoryReuseEnabled() {
		return
	}
	if v == nil {
		return
	}

	v.Count = 0
	v.Min.SetFast(0)
	v.Avg.SetFast(0)
	v.Max.SetFast(0)

	if v.AggregativeStatistics != nil {
		v.AggregativeStatistics.Release()
		v.AggregativeStatistics = nil
	}
	aggregativeValuePool.Put(v)
}

func (m *MetricGaugeFloat64Func) Release() {
	if !MemoryReuseEnabled() {
		return
	}
	atomic.StoreUint64(&m.running, 0)
	*m = MetricGaugeFloat64Func{}
	metricGaugeFloat64FuncPool.Put(m)
}

func (m *MetricCount) Release() {
	if !MemoryReuseEnabled() {
		return
	}
	atomic.StoreUint64(&m.running, 0)
	*m = MetricCount{}
	metricCountPool.Put(m)
}

func (m *MetricGaugeAggregativeFlow) Release() {
	if !MemoryReuseEnabled() {
		return
	}
	atomic.StoreUint64(&m.running, 0)
	//m.reset()
	metricGaugeAggregativeFlowPool.Put(m)
}

func (m *MetricGaugeAggregativeBuffered) Release() {
	if !MemoryReuseEnabled() {
		return
	}
	atomic.StoreUint64(&m.running, 0)
	//m.reset()
	metricGaugeAggregativeBufferedPool.Put(m)
}

func (m *MetricGaugeAggregativeSimple) Release() {
	if !MemoryReuseEnabled() {
		return
	}
	atomic.StoreUint64(&m.running, 0)
	//m.reset()
	metricGaugeAggregativeSimplePool.Put(m)
}

func (m *MetricGaugeInt64) Release() {
	if !MemoryReuseEnabled() {
		return
	}
	atomic.StoreUint64(&m.running, 0)
	//m.reset()
	metricGaugeInt64Pool.Put(m)
}

func (m *MetricGaugeInt64Func) Release() {
	if !MemoryReuseEnabled() {
		return
	}
	atomic.StoreUint64(&m.running, 0)
	//m.reset()
	metricGaugeInt64FuncPool.Put(m)
}

func (m *MetricTimingFlow) Release() {
	if !MemoryReuseEnabled() {
		return
	}
	atomic.StoreUint64(&m.running, 0)
	//m.reset()
	metricTimingFlowPool.Put(m)
}

func (m *MetricTimingBuffered) Release() {
	if !MemoryReuseEnabled() {
		return
	}
	atomic.StoreUint64(&m.running, 0)
	//m.reset()
	metricTimingBufferedPool.Put(m)
}

func (m *MetricTimingSimple) Release() {
	if !MemoryReuseEnabled() {
		return
	}
	atomic.StoreUint64(&m.running, 0)
	//m.reset()
	metricTimingSimplePool.Put(m)
}

func (m *MetricGaugeFloat64) Release() {
	if !MemoryReuseEnabled() {
		return
	}
	atomic.StoreUint64(&m.running, 0)
	//m.reset()
	metricGaugeFloat64Pool.Put(m)
}

func newIterationHandler() *iterationHandler {
	return iterationHandlerPool.Get().(*iterationHandler)
}

func (m *iterationHandler) Release() {
	if !MemoryReuseEnabled() {
		return
	}
	iterationHandlerPool.Put(m)
}
