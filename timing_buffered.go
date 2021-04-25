package metrics

import (
	"time"
)

type MetricTimingBuffered struct {
	commonAggregativeBuffered
}

func (r *Registry) newMetricTimingBuffered(key string, tags AnyTags) *MetricTimingBuffered {
	metric := metricTimingBufferedPool.Get().(*MetricTimingBuffered)
	metric.init(r, key, tags)
	return metric
}

func (m *MetricTimingBuffered) init(r *Registry, key string, tags AnyTags) {
	m.commonAggregativeBuffered.init(r, m, key, tags)
}

func TimingBuffered(key string, tags AnyTags) *MetricTimingBuffered {
	return registry.TimingBuffered(key, tags)
}

func (r *Registry) TimingBuffered(key string, tags AnyTags) *MetricTimingBuffered {
	if IsDisabled() {
		return (*MetricTimingBuffered)(nil)
	}

	m := r.Get(TypeTimingBuffered, key, tags)
	if m != nil {
		return m.(*MetricTimingBuffered)
	}

	return r.newMetricTimingBuffered(key, tags)
}

func (m *MetricTimingBuffered) ConsiderValue(v time.Duration) {
	m.considerValue(float64(v.Nanoseconds()))
}

func (m *MetricTimingBuffered) GetType() Type {
	return TypeTimingBuffered
}
