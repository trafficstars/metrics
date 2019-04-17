package metrics

type metricCommonAggregativeSimple struct {
	metricCommonAggregative
}

func (m *metricCommonAggregativeSimple) init(parent Metric, key string, tags AnyTags) {
	m.metricCommonAggregative.init(parent, key, tags)
}

func (m *metricCommonAggregativeSimple) NewAggregativeStatistics() AggregativeStatistics {
	return nil
}
