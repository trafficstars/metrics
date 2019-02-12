package metrics

type MetricGaugeFloat64 struct {
	metricCommonFloat64
}

func newMetricGaugeFloat64(key string, tags AnyTags) *MetricGaugeFloat64 {
	metric := metricGaugeFloat64Pool.Get().(*MetricGaugeFloat64)
	metric.init(key, tags)
	return metric
}

func (m *MetricGaugeFloat64) init(key string, tags AnyTags) {
	m.metricCommonFloat64.init(m, key, tags)
}

func GaugeFloat64(key string, tags AnyTags) *MetricGaugeFloat64 {
	if IsDisabled() {
		return (*MetricGaugeFloat64)(nil)
	}

	m := metricsRegistry.Get(TypeGaugeFloat64, key, tags)
	if m != nil {
		return m.(*MetricGaugeFloat64)
	}

	return newMetricGaugeFloat64(key, tags)
}

func (m *MetricGaugeFloat64) GetType() Type {
	return TypeGaugeFloat64
}

func (m *MetricGaugeFloat64) Release() {
	*m = MetricGaugeFloat64{}
	metricGaugeFloat64Pool.Put(m)
}
