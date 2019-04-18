package metrics

type MetricGaugeAggregativeSimple struct {
	commonAggregativeSimple
}

func newMetricGaugeAggregativeSimple(key string, tags AnyTags) *MetricGaugeAggregativeSimple {
	metric := metricGaugeAggregativeSimplePool.Get().(*MetricGaugeAggregativeSimple)
	metric.init(key, tags)
	return metric
}

func (m *MetricGaugeAggregativeSimple) init(key string, tags AnyTags) {
	m.commonAggregativeSimple.init(m, key, tags)
}

func GaugeAggregativeSimple(key string, tags AnyTags) *MetricGaugeAggregativeSimple {
	if IsDisabled() {
		return (*MetricGaugeAggregativeSimple)(nil)
	}

	m := registry.Get(TypeGaugeAggregativeSimple, key, tags)
	if m != nil {
		return m.(*MetricGaugeAggregativeSimple)
	}

	return newMetricGaugeAggregativeSimple(key, tags)
}

func (m *MetricGaugeAggregativeSimple) ConsiderValue(v float64) {
	m.considerValue(v)
}

func (m *MetricGaugeAggregativeSimple) GetType() Type {
	return TypeGaugeAggregativeSimple
}
