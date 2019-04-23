package metrics

type commonAggregativeSimple struct {
	commonAggregative
}

func (m *commonAggregativeSimple) init(parent Metric, key string, tags AnyTags) {
	m.commonAggregative.init(parent, key, tags)
}

// NewAggregativeStatistics returns nil
//
// "Simple" doesn't calculate percentile values, so it doesn't have specific aggregative statistics, so "nil"
//
// See "Simple" in README.md
func (m *commonAggregativeSimple) NewAggregativeStatistics() AggregativeStatistics {
	return nil
}
