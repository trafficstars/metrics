package metrics

import (
	"testing"
)

func TestMetricInterfaceOnGaugeAggregativeFlow(t *testing.T) {
	m := newMetricGaugeAggregativeFlow(``, nil)
	checkForInfiniteRecursion(m)
}
