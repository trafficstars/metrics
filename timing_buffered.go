package metrics

import (
	"time"
)

type MetricTimingBuffered struct {
	commonAggregativeBuffered
}

func newMetricTimingBuffered(key string, tags AnyTags) *MetricTimingBuffered {
	metric := metricTimingBufferedPool.Get().(*MetricTimingBuffered)
	metric.init(key, tags)
	return metric
}

func (m *MetricTimingBuffered) init(key string, tags AnyTags) {
	m.commonAggregativeBuffered.init(m, key, tags)
}

func TimingBuffered(key string, tags AnyTags) *MetricTimingBuffered {
	if IsDisabled() {
		return (*MetricTimingBuffered)(nil)
	}

	m := registry.Get(TypeTimingBuffered, key, tags)
	if m != nil {
		return m.(*MetricTimingBuffered)
	}

	return newMetricTimingBuffered(key, tags)
}

func (m *MetricTimingBuffered) ConsiderValue(v time.Duration) {
	m.considerValue(float64(v.Nanoseconds()))
}

func (m *MetricTimingBuffered) GetType() Type {
	return TypeTimingBuffered
}
