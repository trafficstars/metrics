package metrics

type MetricGaugeAggregativeBuffered struct {
	//metricCommonAggregativeFlow
	metricCommonAggregativeShortBuf
}

func newMetricGaugeAggregativeBuffered(key string, tags AnyTags) *MetricGaugeAggregativeBuffered {
	metric := metricGaugeAggregativeBufferedPool.Get().(*MetricGaugeAggregativeBuffered)
	metric.init(key, tags)
	return metric
}

func (m *MetricGaugeAggregativeBuffered) init(key string, tags AnyTags) {
	//m.metricCommonAggregativeFlow.init(m, key, tags)
	m.metricCommonAggregativeShortBuf.init(m, key, tags)
}

func GaugeAggregativeBuffered(key string, tags AnyTags) *MetricGaugeAggregativeBuffered {
	if IsDisabled() {
		return (*MetricGaugeAggregativeBuffered)(nil)
	}

	m := metricsRegistry.Get(TypeGaugeAggregativeBuffered, key, tags)
	if m != nil {
		return m.(*MetricGaugeAggregativeBuffered)
	}

	return newMetricGaugeAggregativeBuffered(key, tags)
}

func (m *MetricGaugeAggregativeBuffered) ConsiderValue(v float64) {
	m.considerValue(v)
}

func (m *MetricGaugeAggregativeBuffered) GetType() Type {
	return TypeGaugeAggregativeBuffered
}
