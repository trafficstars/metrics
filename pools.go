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
			return &AggregationPeriod{}
		},
	}
	aggregativeStatisticsFastPool = &sync.Pool{
		New: func() interface{} {
			return &AggregativeStatisticsFast{}
		},
	}
)
