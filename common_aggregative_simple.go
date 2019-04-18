package metrics

type commonAggregativeSimple struct {
	commonAggregative
}

func (m *commonAggregativeSimple) init(parent Metric, key string, tags AnyTags) {
	m.commonAggregative.init(parent, key, tags)
}

func (m *commonAggregativeSimple) NewAggregativeStatistics() AggregativeStatistics {
	return nil
}
