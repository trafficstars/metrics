package metrics

import (
	"time"
)

type MetricGaugeAggregativeFlow struct {
	metricCommonAggregativeFlow
}

func newMetricGaugeAggregativeFlow(key string, tags AnyTags) *MetricGaugeAggregativeFlow {
	metric := metricGaugeAggregativeFlowPool.Get().(*MetricGaugeAggregativeFlow)
	metric.init(key, tags)
	return metric
}

func (m *MetricGaugeAggregativeFlow) init(key string, tags AnyTags) {
	m.metricCommonAggregativeFlow.init(m, key, tags)
}

func GaugeAggregativeFlow(key string, tags AnyTags) *MetricGaugeAggregativeFlow {
	if IsDisabled() {
		return (*MetricGaugeAggregativeFlow)(nil)
	}

	m := metricsRegistry.Get(TypeGaugeAggregativeFlow, key, tags)
	if m != nil {
		return m.(*MetricGaugeAggregativeFlow)
	}

	return newMetricGaugeAggregativeFlow(key, tags)
}

func (m *MetricGaugeAggregativeFlow) ConsiderValue(v time.Duration) {
	m.considerValue(float64(v.Nanoseconds()))
}

func (m *MetricGaugeAggregativeFlow) GetType() Type {
	return TypeGaugeAggregativeFlow
}

func (m *MetricGaugeAggregativeFlow) Release() {
	*m = MetricGaugeAggregativeFlow{}
	metricGaugeAggregativeFlowPool.Put(m)
}
