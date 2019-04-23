package metrics

// MetricCount is the type of a "Count" metric.
//
// Count metric is an analog of prometheus' "Counter",
// see: https://godoc.org/github.com/prometheus/client_golang/prometheus#Counter
type MetricCount struct {
	commonInt64
}

func newMetricCount(key string, tags AnyTags) *MetricCount {
	metric := metricCountPool.Get().(*MetricCount)
	metric.init(key, tags)
	return metric
}

func (m *MetricCount) init(key string, tags AnyTags) {
	m.commonInt64.init(m, key, tags)
}

// Count returns a metric of type "MetricCount".
//
// For the same key and tags it will return the same metric.
//
// If there's no such metric then it will create it, register in the registry and return it.
// If there's already such metric then if will just return it.
func Count(key string, tags AnyTags) *MetricCount {
	if IsDisabled() {
		return (*MetricCount)(nil)
	}

	m := registry.Get(TypeCount, key, tags)
	if m != nil {
		return m.(*MetricCount)
	}

	return newMetricCount(key, tags)
}

// GetType always returns "TypeCount" (because of type "MetricCount")
func (m *MetricCount) GetType() Type {
	return TypeCount
}
