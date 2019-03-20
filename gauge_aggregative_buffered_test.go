package metrics

import (
	"testing"
)

func TestMetricInterfaceOnGaugeAggregativeBuffered(t *testing.T) {
	m := newMetricGaugeAggregativeBuffered(``, nil)
	checkForInfiniteRecursion(m)
}
