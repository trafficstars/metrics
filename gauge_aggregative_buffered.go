package metrics

// MetricGaugeAggregativeBuffered is an aggregative/summarizive metric (like "average", "percentile 99" and so on).
// It's an analog of prometheus' "Summary" (see https://prometheus.io/docs/concepts/metric_types/#summary).
//
// MetricGaugeAggregativeBuffered uses the "Buffered" method to aggregate the statistics
// (see "Buffered" in README.md)
type MetricGaugeAggregativeBuffered struct {
	commonAggregativeBuffered
}

func newMetricGaugeAggregativeBuffered(key string, tags AnyTags) *MetricGaugeAggregativeBuffered {
	metric := metricGaugeAggregativeBufferedPool.Get().(*MetricGaugeAggregativeBuffered)
	metric.init(key, tags)
	return metric
}

func (m *MetricGaugeAggregativeBuffered) init(key string, tags AnyTags) {
	m.commonAggregativeBuffered.init(m, key, tags)
}

// GaugeAggregativeBuffered returns a metric of type "MetricGaugeAggregativeBuffered".
//
// For the same key and tags it will return the same metric.
//
// If there's no such metric then it will create it, register it in the registry and return it.
// If there's already such metric then it will just return the metric.
//
// MetricGaugeAggregativeBuffered is an aggregative/summarizive metric (like "average", "percentile 99" and so on).
// It's an analog of prometheus' "Summary" (see https://prometheus.io/docs/concepts/metric_types/#summary).
//
// MetricGaugeAggregativeBuffered uses the "Buffered" method to aggregate the statistics
// (see "Buffered" in README.md)
func GaugeAggregativeBuffered(key string, tags AnyTags) *MetricGaugeAggregativeBuffered {
	if IsDisabled() {
		return (*MetricGaugeAggregativeBuffered)(nil)
	}

	m := registry.Get(TypeGaugeAggregativeBuffered, key, tags)
	if m != nil {
		return m.(*MetricGaugeAggregativeBuffered)
	}

	return newMetricGaugeAggregativeBuffered(key, tags)
}

// ConsiderValue adds a value to the statistics, it's an analog of prometheus' "Observe"
// (see https://godoc.org/github.com/prometheus/client_golang/prometheus#Summary)
func (m *MetricGaugeAggregativeBuffered) ConsiderValue(v float64) {
	m.considerValue(v)
}

// GetType always returns TypeGaugeAggregativeBuffered (because of object type "MetricGaugeAggregativeBuffered")
func (m *MetricGaugeAggregativeBuffered) GetType() Type {
	return TypeGaugeAggregativeBuffered
}
