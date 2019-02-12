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
	aggregativeValuePool = &sync.Pool{
		New: func() interface{} {
			return &AggregativeValue{}
		},
	}
	aggregativeStatisticsFastPool = &sync.Pool{
		New: func() interface{} {
			s := &AggregativeStatisticsFast{}
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

func (s *AggregativeStatisticsFast) Release() {
	s.Set(0)
	s.tickID = 0
	aggregativeStatisticsFastPool.Put(s)
}

func newAggregativeStatisticsFast() *AggregativeStatisticsFast {
	return aggregativeStatisticsFastPool.Get().(*AggregativeStatisticsFast)
}
