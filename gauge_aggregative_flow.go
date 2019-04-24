package metrics

// MetricGaugeAggregativeFlow is an aggregative/summarizive metric (like "average", "percentile 99" and so on)..
// It's an analog of prometheus' "Summary" (see https://prometheus.io/docs/concepts/metric_types/#summary).
//
// MetricGaugeAggregativeFlow uses the "Flow" method to aggregate the statistics (see "Flow" in README.md)
type MetricGaugeAggregativeFlow struct {
	commonAggregativeFlow
}

func newMetricGaugeAggregativeFlow(key string, tags AnyTags) *MetricGaugeAggregativeFlow {
	metric := metricGaugeAggregativeFlowPool.Get().(*MetricGaugeAggregativeFlow)
	metric.init(key, tags)
	return metric
}

func (m *MetricGaugeAggregativeFlow) init(key string, tags AnyTags) {
	m.commonAggregativeFlow.init(m, key, tags)
}

// GaugeAggregativeFlow returns a metric of type "MetricGaugeAggregativeFlow".
//
// For the same key and tags it will return the same metric.
//
// If there's no such metric then it will create it, register it in the registry and return it.
// If there's already such metric then it will just return the metric.
//
// MetricGaugeAggregativeFlow is an aggregative/summarizive metric (like "average", "percentile 99" and so on)..
// It's an analog of prometheus' "Summary" (see https://prometheus.io/docs/concepts/metric_types/#summary).
//
// MetricGaugeAggregativeFlow uses the "Flow" method to aggregate the statistics (see "Flow" in README.md)
func GaugeAggregativeFlow(key string, tags AnyTags) *MetricGaugeAggregativeFlow {
	if IsDisabled() {
		return (*MetricGaugeAggregativeFlow)(nil)
	}

	m := registry.Get(TypeGaugeAggregativeFlow, key, tags)
	if m != nil {
		return m.(*MetricGaugeAggregativeFlow)
	}

	return newMetricGaugeAggregativeFlow(key, tags)
}

// ConsiderValue adds a value to the statistics, it's an analog of prometheus' "Observe"
// (see https://godoc.org/github.com/prometheus/client_golang/prometheus#Summary)
func (m *MetricGaugeAggregativeFlow) ConsiderValue(v float64) {
	m.considerValue(v)
}

// GetType always returns TypeGaugeAggregativeFlow (because of object type "MetricGaugeAggregativeFlow")
func (m *MetricGaugeAggregativeFlow) GetType() Type {
	return TypeGaugeAggregativeFlow
}
