package metrics

import (
	"time"
)

type MetricGaugeAggregative struct {
	metricCommonAggregativeFast
}

func newMetricGaugeAggregative(key string, tags AnyTags) *MetricGaugeAggregative {
	metric := metricGaugeAggregativePool.Get().(*MetricGaugeAggregative)
	metric.init(key, tags)
	return metric
}

func (m *MetricGaugeAggregative) init(key string, tags AnyTags) {
	m.metricCommonAggregativeFast.init(m, key, tags)
}

func GaugeAggregative(key string, tags AnyTags) *MetricGaugeAggregative {
	if IsDisabled() {
		return (*MetricGaugeAggregative)(nil)
	}

	m := metricsRegistry.Get(TypeGaugeAggregative, key, tags)
	if m != nil {
		return m.(*MetricGaugeAggregative)
	}

	return newMetricGaugeAggregative(key, tags)
}

func (m *MetricGaugeAggregative) ConsiderValue(v time.Duration) {
	m.considerValue(float64(v.Nanoseconds()))
}

func (m *MetricGaugeAggregative) GetType() Type {
	return TypeGaugeAggregative
}

func (m *MetricGaugeAggregative) Release() {
	*m = MetricGaugeAggregative{}
	metricGaugeAggregativePool.Put(m)
}
