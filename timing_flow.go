package metrics

import (
	"time"
)

type MetricTimingFlow struct {
	commonAggregativeFlow
}

func (r *Registry) newMetricTimingFlow(key string, tags AnyTags) *MetricTimingFlow {
	metric := metricTimingFlowPool.Get().(*MetricTimingFlow)
	metric.init(r, key, tags)
	return metric
}

func (m *MetricTimingFlow) init(r *Registry, key string, tags AnyTags) {
	m.commonAggregativeFlow.init(r, m, key, tags)
}

func TimingFlow(key string, tags AnyTags) *MetricTimingFlow {
	return registry.TimingFlow(key, tags)
}

func (r *Registry) TimingFlow(key string, tags AnyTags) *MetricTimingFlow {
	if IsDisabled() {
		return (*MetricTimingFlow)(nil)
	}

	m := r.Get(TypeTimingFlow, key, tags)
	if m != nil {
		return m.(*MetricTimingFlow)
	}

	return r.newMetricTimingFlow(key, tags)
}

func (m *MetricTimingFlow) ConsiderValue(v time.Duration) {
	m.considerValue(float64(v.Nanoseconds()))
}

func (m *MetricTimingFlow) GetType() Type {
	return TypeTimingFlow
}
