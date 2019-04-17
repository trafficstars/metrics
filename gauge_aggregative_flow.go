package metrics

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

	m := registry.Get(TypeGaugeAggregativeFlow, key, tags)
	if m != nil {
		return m.(*MetricGaugeAggregativeFlow)
	}

	return newMetricGaugeAggregativeFlow(key, tags)
}

func (m *MetricGaugeAggregativeFlow) ConsiderValue(v float64) {
	m.considerValue(v)
}

func (m *MetricGaugeAggregativeFlow) GetType() Type {
	return TypeGaugeAggregativeFlow
}
