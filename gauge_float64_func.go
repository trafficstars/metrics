package metrics

// MetricGaugeFloat64Func is a gauge metric which uses a float64 value returned by a function.
//
// This metric is the same as MetricGaugeFloat64, but uses a function as a source of values.
type MetricGaugeFloat64Func struct {
	common
	fn func() float64
}

func (r *Registry) newMetricGaugeFloat64Func(key string, tags AnyTags, fn func() float64) *MetricGaugeFloat64Func {
	metric := metricGaugeFloat64FuncPool.Get().(*MetricGaugeFloat64Func)
	metric.init(r, key, tags, fn)
	return metric
}

// GaugeFloat64Func returns a metric of type "MetricGaugeFloat64Func".
//
// MetricGaugeFloat64Func is a gauge metric which uses a float64 value returned by the function "fn".
//
// This metric is the same as MetricGaugeFloat64, but uses the function "fn" as a source of values.
//
// Usually if somebody uses this metrics it requires to disable the GC: `metric.SetGCEnabled(false)`
func GaugeFloat64Func(key string, tags AnyTags, fn func() float64) *MetricGaugeFloat64Func {
	return registry.GaugeFloat64Func(key, tags, fn)
}

// GaugeFloat64Func returns a metric of type "MetricGaugeFloat64Func".
//
// MetricGaugeFloat64Func is a gauge metric which uses a float64 value returned by the function "fn".
//
// This metric is the same as MetricGaugeFloat64, but uses the function "fn" as a source of values.
//
// Usually if somebody uses this metrics it requires to disable the GC: `metric.SetGCEnabled(false)`
func (r *Registry) GaugeFloat64Func(key string, tags AnyTags, fn func() float64) *MetricGaugeFloat64Func {
	if r.IsDisabled() {
		return (*MetricGaugeFloat64Func)(nil)
	}

	m := r.Get(TypeGaugeFloat64Func, key, tags)
	if m != nil {
		return m.(*MetricGaugeFloat64Func)
	}

	return r.newMetricGaugeFloat64Func(key, tags, fn)
}

func (m *MetricGaugeFloat64Func) init(r *Registry, key string, tags AnyTags, fn func() float64) {
	m.fn = fn
	m.common.init(r, m, key, tags, func() bool { return m.wasUseless() })
}

func (m *MetricGaugeFloat64Func) GetType() Type {
	return TypeGaugeFloat64Func
}

func (m *MetricGaugeFloat64Func) Get() float64 {
	if m == nil {
		return 0
	}
	return m.fn()
}

func (m *MetricGaugeFloat64Func) GetFloat64() float64 {
	return m.Get()
}

func (m *MetricGaugeFloat64Func) Send(sender Sender) {
	if sender == nil {
		return
	}
	sender.SendFloat64(m.parent, string(m.storageKey), m.Get())
}

func (m *MetricGaugeFloat64Func) wasUseless() bool {
	return m.Get() == 0
}
