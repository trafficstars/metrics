package metrics

import (
	//"runtime"
	"testing"
)

func TestGaugeFloat64GC(t *testing.T) {
	r := New()
	defer r.Reset()
	r.SetDefaultIsRan(false)
	r.SetDefaultGCEnabled(false)
	testGC(t, func() {
		r.GaugeFloat64(`test_gc`, nil)
	})
}

func BenchmarkNewGaugeFloat64(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		Reset()
		//runtime.GC()
		b.StartTimer()
		GaugeFloat64(`test`, Tags{
			"i": i,
		})
	}
}

func TestMetricInterfaceOnGaugeFloat64(t *testing.T) {
	m := registry.newMetricGaugeFloat64(``, nil)
	checkForInfiniteRecursion(m)
}
