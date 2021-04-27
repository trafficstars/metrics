package metrics

import (
	"time"
)

type MetricTimingSimple struct {
	commonAggregativeSimple
}

func (r *Registry) newMetricTimingSimple(key string, tags AnyTags) *MetricTimingSimple {
	metric := metricTimingSimplePool.Get().(*MetricTimingSimple)
	metric.init(r, key, tags)
	return metric
}

func (m *MetricTimingSimple) init(r *Registry, key string, tags AnyTags) {
	m.commonAggregativeSimple.init(r, m, key, tags)
}

func TimingSimple(key string, tags AnyTags) *MetricTimingSimple {
	return registry.TimingSimple(key, tags)
}

func (r *Registry) TimingSimple(key string, tags AnyTags) *MetricTimingSimple {
	if IsDisabled() {
		return (*MetricTimingSimple)(nil)
	}

	m := r.Get(TypeTimingSimple, key, tags)
	if m != nil {
		return m.(*MetricTimingSimple)
	}

	return r.newMetricTimingSimple(key, tags)
}

func (m *MetricTimingSimple) ConsiderValue(v time.Duration) {
	m.considerValue(float64(v.Nanoseconds()))
}

func (m *MetricTimingSimple) GetType() Type {
	return TypeTimingSimple
}
