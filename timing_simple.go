package metrics

import (
	"time"
)

type MetricTimingSimple struct {
	metricCommonAggregativeSimple
}

func newMetricTimingSimple(key string, tags AnyTags) *MetricTimingSimple {
	metric := metricTimingSimplePool.Get().(*MetricTimingSimple)
	metric.init(key, tags)
	return metric
}

func (m *MetricTimingSimple) init(key string, tags AnyTags) {
	m.metricCommonAggregativeSimple.init(m, key, tags)
}

func TimingSimple(key string, tags AnyTags) *MetricTimingSimple {
	if IsDisabled() {
		return (*MetricTimingSimple)(nil)
	}

	m := registry.Get(TypeTimingSimple, key, tags)
	if m != nil {
		return m.(*MetricTimingSimple)
	}

	return newMetricTimingSimple(key, tags)
}

func (m *MetricTimingSimple) ConsiderValue(v time.Duration) {
	m.considerValue(float64(v.Nanoseconds()))
}

func (m *MetricTimingSimple) GetType() Type {
	return TypeTimingSimple
}
