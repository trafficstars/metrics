package metrics

type MetricGaugeInt64Func struct {
	metricCommon
	fn func() int64
}

func newMetricGaugeInt64Func(key string, tags AnyTags, fn func() int64) *MetricGaugeInt64Func {
	metric := metricGaugeInt64FuncPool.Get().(*MetricGaugeInt64Func)
	metric.init(key, tags, fn)
	return metric
}

func (m *MetricGaugeInt64Func) init(key string, tags AnyTags, fn func() int64) {
	m.fn = fn
	m.metricCommon.init(m, key, tags, func() bool { return m.wasUseless() })
}

func GaugeInt64Func(key string, tags AnyTags, fn func() int64) *MetricGaugeInt64Func {
	if IsDisabled() {
		return (*MetricGaugeInt64Func)(nil)
	}

	m := metricsRegistry.Get(TypeGaugeInt64Func, key, tags)
	if m != nil {
		return m.(*MetricGaugeInt64Func)
	}

	return newMetricGaugeInt64Func(key, tags, fn)
}

func (m *MetricGaugeInt64Func) GetType() Type {
	return TypeGaugeInt64Func
}

func (m *MetricGaugeInt64Func) Release() {
	metricGaugeInt64FuncPool.Put(m)
}

func (m *MetricGaugeInt64Func) Get() int64 {
	return m.fn()
}

func (m *MetricGaugeInt64Func) wasUseless() bool {
	return m.Get() == 0
}
