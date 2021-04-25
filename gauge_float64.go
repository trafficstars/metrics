package metrics

// MetricGaugeFloat64 is just a gauge metric which stores the value as float64.
// It's an analog of "Gauge" metric of prometheus, see: https://prometheus.io/docs/concepts/metric_types/#gauge
type MetricGaugeFloat64 struct {
	commonFloat64
}

func (r *Registry) newMetricGaugeFloat64(key string, tags AnyTags) *MetricGaugeFloat64 {
	metric := metricGaugeFloat64Pool.Get().(*MetricGaugeFloat64)
	metric.init(r, key, tags)
	return metric
}

func (m *MetricGaugeFloat64) init(r *Registry, key string, tags AnyTags) {
	m.commonFloat64.init(r, m, key, tags)
}

// GaugeFloat64 returns a metric of type "MetricGaugeFloat64".
//
// For the same key and tags it will return the same metric.
//
// If there's no such metric then it will create it, register it in the registry and return it.
// If there's already such metric then it will just return the metric.
//
// MetricGaugeFloat64 is just a gauge metric which stores the value as float64.
// It's an analog of "Gauge" metric of prometheus, see: https://prometheus.io/docs/concepts/metric_types/#gauge
func GaugeFloat64(key string, tags AnyTags) *MetricGaugeFloat64 {
	return registry.GaugeFloat64(key, tags)
}

// GaugeFloat64 returns a metric of type "MetricGaugeFloat64".
//
// For the same key and tags it will return the same metric.
//
// If there's no such metric then it will create it, register it in the registry and return it.
// If there's already such metric then it will just return the metric.
//
// MetricGaugeFloat64 is just a gauge metric which stores the value as float64.
// It's an analog of "Gauge" metric of prometheus, see: https://prometheus.io/docs/concepts/metric_types/#gauge
func (r *Registry) GaugeFloat64(key string, tags AnyTags) *MetricGaugeFloat64 {
	if IsDisabled() {
		return (*MetricGaugeFloat64)(nil)
	}

	m := r.Get(TypeGaugeFloat64, key, tags)
	if m != nil {
		return m.(*MetricGaugeFloat64)
	}

	return r.newMetricGaugeFloat64(key, tags)
}

// GetType always returns TypeGaugeFloat64 (because of object type "MetricGaugeFloat64")
func (m *MetricGaugeFloat64) GetType() Type {
	return TypeGaugeFloat64
}
