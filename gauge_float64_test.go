package metrics

import (
	//"runtime"
	"testing"
)

func TestGaugeFloat64GC(t *testing.T) {
	testGC(t, func() {
		metric := GaugeFloat64(`test_gc`, nil)
		metric.Stop()
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
