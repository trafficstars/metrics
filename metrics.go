package metrics

import (
	"sort"
)

type Metrics []Metric

func (s Metrics) Sort() {
	s.sortBuiltin()
}

func (s Metrics) sortBuiltin() {
	sort.Slice(s, func(i, j int) bool {
		return s[i].(interface{ GetKey() uint64 }).GetKey() < s[j].(interface{ GetKey() uint64 }).GetKey()
	})
}
