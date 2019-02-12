package metrics

import (
	"time"
)

type MetricTiming struct {
	//metricCommonAggregativeShortBuf
	metricCommonAggregativeFast
}

func newMetricTiming(key string, tags AnyTags) *MetricTiming {
	metric := metricTimingPool.Get().(*MetricTiming)
	metric.init(key, tags)
	return metric
}

func (m *MetricTiming) init(key string, tags AnyTags) {
	//m.metricCommonAggregativeShortBuf.init(m, key, tags)
	m.metricCommonAggregativeFast.init(m, key, tags)
}

func Timing(key string, tags AnyTags) *MetricTiming {
	if IsDisabled() {
		return (*MetricTiming)(nil)
	}

	m := metricsRegistry.Get(TypeTiming, key, tags)
	if m != nil {
		return m.(*MetricTiming)
	}

	return newMetricTiming(key, tags)
}

func (m *MetricTiming) ConsiderValue(v time.Duration) {
	m.considerValue(float64(v.Nanoseconds()))
}

func (m *MetricTiming) GetType() Type {
	return TypeTiming
}

func (m *MetricTiming) Release() {
	*m = MetricTiming{}
	metricTimingPool.Put(m)
}
