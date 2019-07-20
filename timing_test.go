package metrics

import (
	"math/rand"
	"runtime"
	"strings"
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
		result = guessPercentileValue(result, v, uint64(idx), percentile)
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

func fillStats(t *testing.T, metric interface {
	Run(time.Duration)
	ConsiderValue(time.Duration)
	DoSlice()
	Stop()
	GetValuePointers() *AggregativeValues
}) {
	metric.Run(5 * time.Second)
	metric.ConsiderValue(time.Nanosecond * 5000)
	submitConsiderValueQueue(swapConsiderValueQueue())
	waitUntilAllSubmittedConsiderValueQueuesProcessed()

	assert.Equal(t, uint64(1), metric.GetValuePointers().Total.Count)
	metric.DoSlice()
	metric.ConsiderValue(time.Nanosecond * 6000)
	metric.ConsiderValue(time.Nanosecond * 7000)
	submitConsiderValueQueue(swapConsiderValueQueue())
	waitUntilAllSubmittedConsiderValueQueuesProcessed()

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
	submitConsiderValueQueue(swapConsiderValueQueue())
	waitUntilAllSubmittedConsiderValueQueuesProcessed()

	metric.DoSlice()
	metric.ConsiderValue(time.Nanosecond * 500000)
	submitConsiderValueQueue(swapConsiderValueQueue())
	waitUntilAllSubmittedConsiderValueQueuesProcessed()

	metric.Stop()
}

func checkValues(t *testing.T, values *AggregativeValues) {
	assert.Equal(t, uint64(500000), uint64(values.Last.Avg.Get()))
	assert.Equal(t, uint64(500000), uint64(values.Last.Sum.Get()))
	assert.Equal(t, uint64(60), values.ByPeriod[0].Count)
	assert.Equal(t, float64(12*(3000+4000+5000+6000+7000)), values.ByPeriod[0].Sum.Get())
	assert.Equal(t, uint64(3000), uint64(values.ByPeriod[0].Min.Get()))
	assert.Equal(t, uint64(500), uint64((values.ByPeriod[0].Avg.Get()+5)/10))
	assert.Equal(t, uint64(2), uint64((*values.ByPeriod[0].AggregativeStatistics.GetPercentile(0.5)+500)/2500))
	assert.Equal(t, uint64(7), uint64((*values.ByPeriod[0].AggregativeStatistics.GetPercentile(0.99)+999)/1000))
	assert.Equal(t, uint64(3), uint64((*values.ByPeriod[1].AggregativeStatistics.GetPercentile(0.99)+999)/2000))
	assert.Equal(t, uint64(7000), uint64(values.ByPeriod[0].Max.Get()))
	assert.Equal(t, uint64(63), values.ByPeriod[1].Count)
	assert.Equal(t, uint64(3000), uint64(values.ByPeriod[1].Min.Get()))
	assert.Equal(t, strings.Split(values.ByPeriod[3].String(), "sum")[0], strings.Split(values.ByPeriod[2].String(), "sum")[0])
	assert.Equal(t, strings.Split(values.ByPeriod[3].String(), "sum")[0], strings.Split(values.ByPeriod[4].String(), "sum")[0])
	assert.Equal(t, strings.Split(values.ByPeriod[3].String(), "sum")[0], strings.Split(values.ByPeriod[5].String(), "sum")[0])
	assert.Equal(t, uint64(64), values.Total.Count)
	assert.Equal(t, uint64(3000), uint64(values.Total.Min.Get()))
	assert.Equal(t, uint64(1), uint64(values.Total.Avg.Get()/10000))
	assert.Equal(t, uint64(500000), uint64(values.Total.Max.Get()))
	assert.Equal(t, float64(500000+12*(3000+4000+5000+6000+7000)+7000+6000+5000), values.Total.Sum.Get())
}

func TestTimingBuffered(t *testing.T) {
	metric := TimingBuffered(`test`, nil)
	fillStats(t, metric)
	checkValues(t, metric.GetValuePointers())
}

func TestTimingFlow(t *testing.T) {
	metric := TimingFlow(`test`, nil)
	fillStats(t, metric)
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
	assert.True(t, cleanedMemstats.HeapInuse <= memstats.HeapInuse)

	for i := 0; i < 100000; i++ {
		fn()
	}
	GC()
	//assert.Equal(t, int64(0), iterationHandlers.routinesCount)
	runtime.GC()
	runtime.ReadMemStats(&cleanedMemstats)

	if !assert.True(t, (int64(cleanedMemstats.HeapInuse)-int64(memstats.HeapInuse))/100000 < 1) {
		t.Error(cleanedMemstats.HeapInuse, int64(cleanedMemstats.HeapInuse)-int64(memstats.HeapInuse))
	}

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
