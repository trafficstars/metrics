package metrics

import (
	"testing"
)

func BenchmarkIncrementDecrement(b *testing.B) {
	metric := GaugeInt64(`test_metric`, nil)
	metric.SetGCEnabled(false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metric.Increment()
		metric.Decrement()
	}
}
