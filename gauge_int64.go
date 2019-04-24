package metrics

// MetricGaugeInt64 is just a gauge metric which stores the value as int64.
// It's an analog of "Gauge" metric of prometheus, see: https://prometheus.io/docs/concepts/metric_types/#gauge
type MetricGaugeInt64 struct {
	commonInt64
}

func newMetricGaugeInt64(key string, tags AnyTags) *MetricGaugeInt64 {
	metric := metricGaugeInt64Pool.Get().(*MetricGaugeInt64)
	metric.init(key, tags)
	return metric
}

func (m *MetricGaugeInt64) init(key string, tags AnyTags) {
	m.commonInt64.init(m, key, tags)
}

// GaugeInt64 returns a metric of type "MetricGaugeFloat64".
//
// For the same key and tags it will return the same metric.
//
// If there's no such metric then it will create it, register it in the registry and return it.
// If there's already such metric then it will just return the metric.
//
// MetricGaugeInt64 is just a gauge metric which stores the value as int64.
// It's an analog of "Gauge" metric of prometheus, see: https://prometheus.io/docs/concepts/metric_types/#gauge
func GaugeInt64(key string, tags AnyTags) *MetricGaugeInt64 {
	if IsDisabled() {
		return (*MetricGaugeInt64)(nil)
	}

	m := registry.Get(TypeGaugeInt64, key, tags)
	if m != nil {
		return m.(*MetricGaugeInt64)
	}

	return newMetricGaugeInt64(key, tags)
}

// GetType always returns TypeGaugeInt64 (because of object type "MetricGaugeInt64")
func (m *MetricGaugeInt64) GetType() Type {
	return TypeGaugeInt64
}

// Decrement is an analog of Add(-1). It just subtracts "1" from the internal value and returns the result.
func (m *MetricGaugeInt64) Decrement() int64 {
	return m.Add(-1)
}
