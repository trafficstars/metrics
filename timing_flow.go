package metrics

import (
	"time"
)

type MetricTimingFlow struct {
	metricCommonAggregativeFlow
}

func newMetricTimingFlow(key string, tags AnyTags) *MetricTimingFlow {
	metric := metricTimingFlowPool.Get().(*MetricTimingFlow)
	metric.init(key, tags)
	return metric
}

func (m *MetricTimingFlow) init(key string, tags AnyTags) {
	m.metricCommonAggregativeFlow.init(m, key, tags)
}

func TimingFlow(key string, tags AnyTags) *MetricTimingFlow {
	if IsDisabled() {
		return (*MetricTimingFlow)(nil)
	}

	m := metricsRegistry.Get(TypeTimingFlow, key, tags)
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
