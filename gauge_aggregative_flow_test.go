package metrics

import (
	"testing"
)

func TestMetricInterfaceOnGaugeAggregativeFlow(t *testing.T) {
	m := registry.newMetricGaugeAggregativeFlow(``, nil)
	checkForInfiniteRecursion(m)
}
