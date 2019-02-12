package metrics

import (
	"bytes"
	"sort"
)

type Metrics []Metric

func (s Metrics) Sort() {
	s.sortBuiltin()
}

func (s Metrics) sortBuiltin() {
	sort.Slice(s, func(i, j int) bool {
		if bytes.Compare(
			s[i].(interface{ GetKey() []byte }).GetKey(),
			s[j].(interface{ GetKey() []byte }).GetKey()) < 0 {
			return true
		}
		return false
	})
}
