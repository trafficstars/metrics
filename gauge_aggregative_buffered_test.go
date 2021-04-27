package metrics

import (
	"testing"
)

func TestMetricInterfaceOnGaugeAggregativeBuffered(t *testing.T) {
	m := registry.newMetricGaugeAggregativeBuffered(``, nil)
	checkForInfiniteRecursion(m)
}
