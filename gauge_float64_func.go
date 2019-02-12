package metrics

type MetricGaugeFloat64Func struct {
	metricCommon
	fn func() float64
}

func newMetricGaugeFloat64Func(key string, tags AnyTags, fn func() float64) *MetricGaugeFloat64Func {
	metric := metricGaugeFloat64FuncPool.Get().(*MetricGaugeFloat64Func)
	metric.init(key, tags, fn)
	return metric
}

func GaugeFloat64Func(key string, tags AnyTags, fn func() float64) *MetricGaugeFloat64Func {
	if IsDisabled() {
		return (*MetricGaugeFloat64Func)(nil)
	}

	m := metricsRegistry.Get(TypeGaugeFloat64Func, key, tags)
	if m != nil {
		return m.(*MetricGaugeFloat64Func)
	}

	return newMetricGaugeFloat64Func(key, tags, fn)
}

func (m *MetricGaugeFloat64Func) init(key string, tags AnyTags, fn func() float64) {
	m.fn = fn
	m.metricCommon.init(m, key, tags, func() bool { return m.wasUseless() })
}

func (m *MetricGaugeFloat64Func) GetType() Type {
	return TypeGaugeFloat64Func
}

func (m *MetricGaugeFloat64Func) Get() float64 {
	return m.fn()
}

func (m *MetricGaugeFloat64Func) wasUseless() bool {
	return m.Get() == 0
}
