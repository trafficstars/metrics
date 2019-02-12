package metrics

import (
	"sync"
)

var (
	memoryReuse = true
)

var (
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
	aggregativeValuePool = &sync.Pool{
		New: func() interface{} {
			return &AggregativeValue{}
		},
	}
	aggregativeStatisticsFlowPool = &sync.Pool{
		New: func() interface{} {
			s := &AggregativeStatisticsFlow{}
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
			return &aggregativeBuffer{}
		},
	}
	aggregativeStatisticsShortBufPool = &sync.Pool{
		New: func() interface{} {
			return &AggregativeStatisticsShortBuf{}
		},
	}
)

func (s *AggregativeStatisticsShortBuf) Release() {
	if !memoryReuse {
		return
	}
	s.filledSize = 0
	s.tickID = 0
	aggregativeStatisticsShortBufPool.Put(s)
}

func newAggregativeStatisticsShortBuf() *AggregativeStatisticsShortBuf {
	return aggregativeStatisticsShortBufPool.Get().(*AggregativeStatisticsShortBuf)
}

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

func (s *AggregativeStatisticsFlow) Release() {
	if !memoryReuse {
		return
	}
	s.Set(0)
	s.tickID = 0
	aggregativeStatisticsFlowPool.Put(s)
}

func newAggregativeStatisticsFlow() *AggregativeStatisticsFlow {
	return aggregativeStatisticsFlowPool.Get().(*AggregativeStatisticsFlow)
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

func (m *MetricGaugeFloat64) Release() {
	if !memoryReuse {
		return
	}
	*m = MetricGaugeFloat64{}
	metricGaugeFloat64Pool.Put(m)
}
