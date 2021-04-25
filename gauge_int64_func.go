package metrics

type MetricGaugeInt64Func struct {
	common
	fn func() int64
}

func (r *Registry) newMetricGaugeInt64Func(key string, tags AnyTags, fn func() int64) *MetricGaugeInt64Func {
	metric := metricGaugeInt64FuncPool.Get().(*MetricGaugeInt64Func)
	metric.init(r, key, tags, fn)
	return metric
}

func (m *MetricGaugeInt64Func) init(r *Registry, key string, tags AnyTags, fn func() int64) {
	m.fn = fn
	m.common.init(r, m, key, tags, func() bool { return m.wasUseless() })
}

func GaugeInt64Func(key string, tags AnyTags, fn func() int64) *MetricGaugeInt64Func {
	return registry.GaugeInt64Func(key, tags, fn)
}

func (r *Registry) GaugeInt64Func(key string, tags AnyTags, fn func() int64) *MetricGaugeInt64Func {
	if IsDisabled() {
		return (*MetricGaugeInt64Func)(nil)
	}

	m := r.Get(TypeGaugeInt64Func, key, tags)
	if m != nil {
		return m.(*MetricGaugeInt64Func)
	}

	return r.newMetricGaugeInt64Func(key, tags, fn)
}

func (m *MetricGaugeInt64Func) GetType() Type {
	return TypeGaugeInt64Func
}

func (m *MetricGaugeInt64Func) Get() int64 {
	if m == nil {
		return 0
	}
	return m.fn()
}

func (m *MetricGaugeInt64Func) GetFloat64() float64 {
	return float64(m.Get())
}

func (m *MetricGaugeInt64Func) Send(sender Sender) {
	if sender == nil {
		return
	}
	sender.SendUint64(m.parent, string(m.storageKey), uint64(m.Get()))
}

func (m *MetricGaugeInt64Func) wasUseless() bool {
	return m.Get() == 0
}
