package metrics

import (
	"sync"
)

var (
	metricCountPool = &sync.Pool{
		New: func() interface{} {
			return &MetricCount{}
		},
	}
	metricGaugeAggregativePool = &sync.Pool{
		New: func() interface{} {
			return &MetricGaugeAggregative{}
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
	metricTimingPool = &sync.Pool{
		New: func() interface{} {
			return &MetricTiming{}
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
	s.filledSize = 0
	s.tickID = 0
	aggregativeStatisticsShortBufPool.Put(s)
}

func newAggregativeStatisticsShortBuf() *AggregativeStatisticsShortBuf {
	return aggregativeStatisticsShortBufPool.Get().(*AggregativeStatisticsShortBuf)
}

func (b *aggregativeBuffer) Release() {
	b.filledSize = 0
	aggregativeBufferPool.Put(b)
}

func newAggregativeBuffer() *aggregativeBuffer {
	return aggregativeBufferPool.Get().(*aggregativeBuffer)
}

func (s *AggregativeStatisticsFlow) Release() {
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
	if v == nil {
		return
	}
	if v.AggregativeStatistics != nil {
		v.AggregativeStatistics.Release()
		v.AggregativeStatistics = nil
	}
	aggregativeValuePool.Put(v)
}
