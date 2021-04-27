package metrics

// MetricGaugeAggregativeSimple is an aggregative/summarizive metric (like "average", "percentile 99" and so on)..
// It's an analog of prometheus' "Summary" (see https://prometheus.io/docs/concepts/metric_types/#summary).
//
// MetricGaugeAggregativeSimple uses the "Simple" method to aggregate the statistics (see "Simple" in README.md)
type MetricGaugeAggregativeSimple struct {
	commonAggregativeSimple
}

func (r *Registry) newMetricGaugeAggregativeSimple(key string, tags AnyTags) *MetricGaugeAggregativeSimple {
	metric := metricGaugeAggregativeSimplePool.Get().(*MetricGaugeAggregativeSimple)
	metric.init(r, key, tags)
	return metric
}

func (m *MetricGaugeAggregativeSimple) init(r *Registry, key string, tags AnyTags) {
	m.commonAggregativeSimple.init(r, m, key, tags)
}

// GaugeAggregativeSimple returns a metric of type "MetricGaugeAggregativeSimple".
//
// For the same key and tags it will return the same metric.
//
// If there's no such metric then it will create it, register it in the registry and return it.
// If there's already such metric then it will just return the metric.
//
// MetricGaugeAggregativeSimple is an aggregative/summarizive metric (like "average", "percentile 99" and so on)..
// It's an analog of prometheus' "Summary" (see https://prometheus.io/docs/concepts/metric_types/#summary).
//
// MetricGaugeAggregativeSimple uses the "Simple" method to aggregate the statistics (see "Simple" in README.md)
func GaugeAggregativeSimple(key string, tags AnyTags) *MetricGaugeAggregativeSimple {
	return registry.GaugeAggregativeSimple(key, tags)
}

func (r *Registry) GaugeAggregativeSimple(key string, tags AnyTags) *MetricGaugeAggregativeSimple {
	if IsDisabled() {
		return (*MetricGaugeAggregativeSimple)(nil)
	}

	m := r.Get(TypeGaugeAggregativeSimple, key, tags)
	if m != nil {
		return m.(*MetricGaugeAggregativeSimple)
	}

	return r.newMetricGaugeAggregativeSimple(key, tags)
}

// ConsiderValue adds a value to the statistics, it's an analog of prometheus' "Observe"
// (see https://godoc.org/github.com/prometheus/client_golang/prometheus#Summary)
func (m *MetricGaugeAggregativeSimple) ConsiderValue(v float64) {
	m.considerValue(v)
}

// GetType always returns TypeGaugeAggregativeSimple (because of object type "MetricGaugeAggregativeSimple")
func (m *MetricGaugeAggregativeSimple) GetType() Type {
	return TypeGaugeAggregativeSimple
}
