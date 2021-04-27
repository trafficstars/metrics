package metrics

import (
	"testing"
)

func TestMetricInterfaceOnGaugeInt64Func(t *testing.T) {
	m := registry.newMetricGaugeInt64Func(``, nil, func() int64 { return 0 })
	checkForInfiniteRecursion(m)
}
