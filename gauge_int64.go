package metrics

type MetricGaugeInt64 struct {
	metricCommonInt64
}

func newMetricGaugeInt64(key string, tags AnyTags) *MetricGaugeInt64 {
	metric := metricGaugeInt64Pool.Get().(*MetricGaugeInt64)
	metric.init(key, tags)
	return metric
}

func (m *MetricGaugeInt64) init(key string, tags AnyTags) {
	m.metricCommonInt64.init(m, key, tags)
}

func GaugeInt64(key string, tags AnyTags) *MetricGaugeInt64 {
	if IsDisabled() {
		return (*MetricGaugeInt64)(nil)
	}

	m := metricsRegistry.Get(TypeGaugeInt64, key, tags)
	if m != nil {
		return m.(*MetricGaugeInt64)
	}

	return newMetricGaugeInt64(key, tags)
}

func (m *MetricGaugeInt64) GetType() Type {
	return TypeGaugeInt64
}

func (m *MetricGaugeInt64) Release() {
	metricGaugeInt64Pool.Put(m)
}
