package metrics

import (
	"testing"
)

func TestGaugeInt64GC(t *testing.T) {
	r := New()
	defer r.Reset()
	r.SetDefaultIsRan(false)
	r.SetDefaultGCEnabled(false)
	testGC(t, func() {
		r.GaugeInt64(`test_gc`, nil)
	})
}

func BenchmarkIncrementDecrement(b *testing.B) {
	metric := GaugeInt64(`test_metric`, nil)
	metric.SetGCEnabled(false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metric.Increment()
		metric.Decrement()
	}
}
