package metrics

import (
	"bytes"
	"sync"
)

var (
	memoryReuse = true
)

// SetMemoryReuseEnabled defines if memory reuse will be enabled (default -- enabled).
func SetMemoryReuseEnabled(isEnabled bool) {
	memoryReuse = isEnabled
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
			s := &aggregativeStatisticsFlow{}
			s.Per1.Pointer = &[]float64{0}[0]
			s.Per10.Pointer = &[]float64{0}[0]
			s.Per50.Pointer = &[]float64{0}[0]
			s.Per90.Pointer = &[]float64{0}[0]
			s.Per99.Pointer = &[]float64{0}[0]
			return s
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
	aggregativeStatisticsShortBufPool = &sync.Pool{
		New: func() interface{} {
			buf := &aggregativeStatisticsShortBuf{}
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
)

func newBytesBuffer() *bytesBuffer {
	return bytesBufferPool.Get().(*bytesBuffer)
}

func (buf *bytesBuffer) Release() {
	if !memoryReuse {
		return
	}
	buf.Reset()
	bytesBufferPool.Put(buf)
}

func newStringSlice() *stringSlice {
	return stringSlicePool.Get().(*stringSlice)
}

func (s *stringSlice) Release() {
	if !memoryReuse {
		return
	}
	*s = (*s)[:0]
	stringSlicePool.Put(s)
}

// Release should be called when the buffer won't be used anymore (to put into into the pool of free buffers) to
// reduce pressure on GC.
func (s *aggregativeStatisticsShortBuf) Release() {
	if !memoryReuse {
		return
	}
	s.filledSize = 0
	s.tickID = 0
	aggregativeStatisticsShortBufPool.Put(s)
}

func newAggregativeStatisticsShortBuf() *aggregativeStatisticsShortBuf {
	return aggregativeStatisticsShortBufPool.Get().(*aggregativeStatisticsShortBuf)
}

// Release should be called when the buffer won't be used anymore (to put into into the pool of free buffers) to
// reduce pressure on GC.
func (b *aggregativeBuffer) Release() {
	if !memoryReuse {
		return
	}
	b.filledSize = 0
	aggregativeBufferPool.Put(b)
}

func newAggregativeBuffer() *aggregativeBuffer {
	return aggregativeBufferPool.Get().(*aggregativeBuffer)
}

func (s *aggregativeStatisticsFlow) Release() {
	if !memoryReuse {
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
	if !memoryReuse {
		return
	}
	if v == nil {
		return
	}
	if v.AggregativeStatistics != nil {
		v.AggregativeStatistics.Release()
		v.AggregativeStatistics = nil
	}
	aggregativeValuePool.Put(v)
}

func (m *MetricGaugeFloat64Func) Release() {
	if !memoryReuse {
		return
	}
	*m = MetricGaugeFloat64Func{}
	metricGaugeFloat64FuncPool.Put(m)
}

func (m *MetricCount) Release() {
	if !memoryReuse {
		return
	}
	*m = MetricCount{}
	metricCountPool.Put(m)
}

func (m *MetricGaugeAggregativeFlow) Release() {
	if !memoryReuse {
		return
	}
	*m = MetricGaugeAggregativeFlow{}
	metricGaugeAggregativeFlowPool.Put(m)
}

func (m *MetricGaugeAggregativeBuffered) Release() {
	if !memoryReuse {
		return
	}
	*m = MetricGaugeAggregativeBuffered{}
	metricGaugeAggregativeBufferedPool.Put(m)
}

func (m *MetricGaugeAggregativeSimple) Release() {
	if !memoryReuse {
		return
	}
	*m = MetricGaugeAggregativeSimple{}
	metricGaugeAggregativeSimplePool.Put(m)
}

func (m *MetricGaugeInt64) Release() {
	if !memoryReuse {
		return
	}
	*m = MetricGaugeInt64{}
	metricGaugeInt64Pool.Put(m)
}

func (m *MetricGaugeInt64Func) Release() {
	if !memoryReuse {
		return
	}
	*m = MetricGaugeInt64Func{}
	metricGaugeInt64FuncPool.Put(m)
}

func (m *MetricTimingFlow) Release() {
	if !memoryReuse {
		return
	}
	*m = MetricTimingFlow{}
	metricTimingFlowPool.Put(m)
}

func (m *MetricTimingBuffered) Release() {
	if !memoryReuse {
		return
	}
	*m = MetricTimingBuffered{}
	metricTimingBufferedPool.Put(m)
}

func (m *MetricTimingSimple) Release() {
	if !memoryReuse {
		return
	}
	*m = MetricTimingSimple{}
	metricTimingSimplePool.Put(m)
}

func (m *MetricGaugeFloat64) Release() {
	if !memoryReuse {
		return
	}
	*m = MetricGaugeFloat64{}
	metricGaugeFloat64Pool.Put(m)
}

func newIterationHandler() *iterationHandler {
	return iterationHandlerPool.Get().(*iterationHandler)
}

func (m *iterationHandler) Release() {
	if !memoryReuse {
		return
	}
	iterationHandlerPool.Put(m)
}
