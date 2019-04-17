package metrics

import (
	"math/rand"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	valuesAmount       = iterationsRequiredPerSecond * 10
	permittedDeviation = 1 / (1 - 0.99) / iterationsRequiredPerSecond
)

func checkPercentile(t *testing.T, percentile float64) float64 {
	values := []float64{}

	for i := 0; i < valuesAmount; i++ {
		r := float64(rand.Intn(1000))
		values = append(values, r*r)
	}

	var result float64
	for idx, v := range values {
		result = guessPercentile(result, v, uint64(idx), percentile)
	}

	count := 0
	for _, v := range values {
		if v < result {
			count++
		}
	}

	return float64(count) / valuesAmount
}

func TestGuessPercentile(t *testing.T) {
	for _, percentile := range []float64{0.01, 0.1, 0.5, 0.9, 0.99} {
		resultPercentile := checkPercentile(t, percentile)
		oldV := percentile / (1 - percentile)
		newV := resultPercentile / (1 - resultPercentile)
		if newV/oldV > (1+permittedDeviation) || oldV/newV > (1+permittedDeviation) {
			t.Errorf("Too different expected and result percentiles: %v %v", percentile, resultPercentile)
		}
	}
}

func fillStats(metric interface {
	Run(time.Duration)
	ConsiderValue(time.Duration)
	DoSlice()
	Stop()
}) {
	metric.Run(5 * time.Second)
	metric.ConsiderValue(time.Nanosecond * 5000)
	metric.DoSlice()
	metric.ConsiderValue(time.Nanosecond * 6000)
	metric.ConsiderValue(time.Nanosecond * 7000)
	metric.DoSlice()
	metric.ConsiderValue(time.Nanosecond * 3000)
	metric.ConsiderValue(time.Nanosecond * 7000)
	metric.ConsiderValue(time.Nanosecond * 4000)
	metric.ConsiderValue(time.Nanosecond * 6000)
	metric.ConsiderValue(time.Nanosecond * 5000)
	metric.ConsiderValue(time.Nanosecond * 3000)
	metric.ConsiderValue(time.Nanosecond * 7000)
	metric.ConsiderValue(time.Nanosecond * 4000)
	metric.ConsiderValue(time.Nanosecond * 6000)
	metric.ConsiderValue(time.Nanosecond * 5000)
	metric.ConsiderValue(time.Nanosecond * 3000)
	metric.ConsiderValue(time.Nanosecond * 7000)
	metric.ConsiderValue(time.Nanosecond * 4000)
	metric.ConsiderValue(time.Nanosecond * 6000)
	metric.ConsiderValue(time.Nanosecond * 5000)
	metric.ConsiderValue(time.Nanosecond * 3000)
	metric.ConsiderValue(time.Nanosecond * 7000)
	metric.ConsiderValue(time.Nanosecond * 4000)
	metric.ConsiderValue(time.Nanosecond * 6000)
	metric.ConsiderValue(time.Nanosecond * 5000)
	metric.ConsiderValue(time.Nanosecond * 3000)
	metric.ConsiderValue(time.Nanosecond * 7000)
	metric.ConsiderValue(time.Nanosecond * 4000)
	metric.ConsiderValue(time.Nanosecond * 6000)
	metric.ConsiderValue(time.Nanosecond * 5000)
	metric.ConsiderValue(time.Nanosecond * 3000)
	metric.ConsiderValue(time.Nanosecond * 7000)
	metric.ConsiderValue(time.Nanosecond * 4000)
	metric.ConsiderValue(time.Nanosecond * 6000)
	metric.ConsiderValue(time.Nanosecond * 5000)
	metric.ConsiderValue(time.Nanosecond * 3000)
	metric.ConsiderValue(time.Nanosecond * 7000)
	metric.ConsiderValue(time.Nanosecond * 4000)
	metric.ConsiderValue(time.Nanosecond * 6000)
	metric.ConsiderValue(time.Nanosecond * 5000)
	metric.ConsiderValue(time.Nanosecond * 3000)
	metric.ConsiderValue(time.Nanosecond * 7000)
	metric.ConsiderValue(time.Nanosecond * 4000)
	metric.ConsiderValue(time.Nanosecond * 6000)
	metric.ConsiderValue(time.Nanosecond * 5000)
	metric.ConsiderValue(time.Nanosecond * 3000)
	metric.ConsiderValue(time.Nanosecond * 7000)
	metric.ConsiderValue(time.Nanosecond * 4000)
	metric.ConsiderValue(time.Nanosecond * 6000)
	metric.ConsiderValue(time.Nanosecond * 5000)
	metric.ConsiderValue(time.Nanosecond * 3000)
	metric.ConsiderValue(time.Nanosecond * 7000)
	metric.ConsiderValue(time.Nanosecond * 4000)
	metric.ConsiderValue(time.Nanosecond * 6000)
	metric.ConsiderValue(time.Nanosecond * 5000)
	metric.ConsiderValue(time.Nanosecond * 3000)
	metric.ConsiderValue(time.Nanosecond * 7000)
	metric.ConsiderValue(time.Nanosecond * 4000)
	metric.ConsiderValue(time.Nanosecond * 6000)
	metric.ConsiderValue(time.Nanosecond * 5000)
	metric.ConsiderValue(time.Nanosecond * 3000)
	metric.ConsiderValue(time.Nanosecond * 7000)
	metric.ConsiderValue(time.Nanosecond * 4000)
	metric.ConsiderValue(time.Nanosecond * 6000)
	metric.ConsiderValue(time.Nanosecond * 5000)
	metric.DoSlice()
	metric.ConsiderValue(time.Nanosecond * 500000)
	metric.Stop()
}

func checkValues(t *testing.T, values *AggregativeValues) {
	assert.Equal(t, uint64(500000), uint64(values.Last.Avg.Get()))
	assert.Equal(t, uint64(60), values.ByPeriod[0].Count)
	assert.Equal(t, uint64(3000), uint64(values.ByPeriod[0].Min.Get()))
	assert.Equal(t, uint64(500), uint64((values.ByPeriod[0].Avg.Get()+5)/10))
	assert.Equal(t, uint64(2), uint64((*values.ByPeriod[0].AggregativeStatistics.GetPercentile(0.5)+500)/2500))
	assert.Equal(t, uint64(7), uint64((*values.ByPeriod[0].AggregativeStatistics.GetPercentile(0.99)+999)/1000))
	assert.Equal(t, uint64(3), uint64((*values.ByPeriod[1].AggregativeStatistics.GetPercentile(0.99)+999)/2000))
	assert.Equal(t, uint64(7000), uint64(values.ByPeriod[0].Max.Get()))
	assert.Equal(t, uint64(63), values.ByPeriod[1].Count)
	assert.Equal(t, uint64(3000), uint64(values.ByPeriod[1].Min.Get()))
	assert.Equal(t, values.ByPeriod[3].String(), values.ByPeriod[2].String())
	assert.Equal(t, values.ByPeriod[3].String(), values.ByPeriod[4].String())
	assert.Equal(t, values.ByPeriod[3].String(), values.ByPeriod[5].String())
	assert.Equal(t, uint64(64), values.Total.Count)
	assert.Equal(t, uint64(3000), uint64(values.Total.Min.Get()))
	assert.Equal(t, uint64(1), uint64(values.Total.Avg.Get()/10000))
	assert.Equal(t, uint64(500000), uint64(values.Total.Max.Get()))
}

func TestTimingBufferd(t *testing.T) {
	metric := TimingBuffered(`test`, nil)
	fillStats(metric)
	checkValues(t, metric.GetValuePointers())
}

func TestTimingFlow(t *testing.T) {
	metric := TimingFlow(`test`, nil)
	fillStats(metric)
	checkValues(t, metric.GetValuePointers())
}

func BenchmarkNewTimingBuffered(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		Reset()
		//runtime.GC()
		b.StartTimer()
		TimingBuffered(`test`, Tags{
			"i": i,
		})
	}
}

func BenchmarkNewTimingFlow(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		Reset()
		//runtime.GC()
		b.StartTimer()
		TimingFlow(`test`, Tags{
			"i": i,
		})
	}
}

func testGC(t *testing.T, fn func()) {
	return
	memoryReuse = false
	Reset()
	keys := registry.storage.Keys()
	if len(keys) != 0 {
		t.Errorf(`len(keys) == %v\n`, keys)
	}
	GC()
	runtime.GC()
	/*if iterationHandlers.routinesCount > 0 {
		t.Errorf(`iterationHandlers.routinesCount == %v\n`, iterationHandlers.routinesCount)
		t.Errorf(`iterationHandlers.m.Keys() == %v`, iterationHandlers.m.Keys())
	}*/
	var memstats, cleanedMemstats runtime.MemStats
	goroutinesCount := runtime.NumGoroutine()
	runtime.GC()
	runtime.ReadMemStats(&memstats)
	fn()
	GC()
	runtime.GC()
	runtime.ReadMemStats(&cleanedMemstats)
	cleanedGoroutinesCount := runtime.NumGoroutine()
	//assert.Equal(t, int64(0), iterationHandlers.routinesCount)
	assert.Equal(t, goroutinesCount, cleanedGoroutinesCount)
	//assert.Equal(t, memstats.HeapInuse, cleanedMemstats.HeapInuse)

	for i := 0; i < 400000; i++ {
		fn()
	}
	GC()
	//assert.Equal(t, int64(0), iterationHandlers.routinesCount)
	runtime.GC()
	runtime.ReadMemStats(&cleanedMemstats)

	assert.Equal(t, uint64(0), (cleanedMemstats.HeapInuse-memstats.HeapInuse)/400000)

	memoryReuse = true
}

func TestTimingBufferedGC(t *testing.T) {
	testGC(t, func() {
		metric := TimingBuffered(`test_gc`, nil)
		metric.Stop()
	})
}

func TestTimingFlowGC(t *testing.T) {
	testGC(t, func() {
		metric := TimingFlow(`test_gc`, nil)
		metric.Stop()
	})
}
