package metrics

import (
	"time"
)

type MetricTimingFlow struct {
	commonAggregativeFlow
}

func newMetricTimingFlow(key string, tags AnyTags) *MetricTimingFlow {
	metric := metricTimingFlowPool.Get().(*MetricTimingFlow)
	metric.init(key, tags)
	return metric
}

func (m *MetricTimingFlow) init(key string, tags AnyTags) {
	m.commonAggregativeFlow.init(m, key, tags)
}

func TimingFlow(key string, tags AnyTags) *MetricTimingFlow {
	if IsDisabled() {
		return (*MetricTimingFlow)(nil)
	}

	m := registry.Get(TypeTimingFlow, key, tags)
	if m != nil {
		return m.(*MetricTimingFlow)
	}

	return newMetricTimingFlow(key, tags)
}

func (m *MetricTimingFlow) ConsiderValue(v time.Duration) {
	m.considerValue(float64(v.Nanoseconds()))
}

func (m *MetricTimingFlow) GetType() Type {
	return TypeTimingFlow
}
