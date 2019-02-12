package metrics

import (
	"math/rand"
	"testing"
	"time"
	"sync/atomic"

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

func fillStats(metric *MetricTiming) {
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
	metric.DoSlice()
	metric.ConsiderValue(time.Nanosecond * 500000)
	metric.Stop()
}

func BenchmarkTiming(b *testing.B) {
	i := uint64(0)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			metric := Timing(`test`, Tags{
				"i": atomic.AddUint64(&i, 1),
			})

			fillStats(metric)
		}
	})
}

func TestTiming(t *testing.T) {
	metric := Timing(`test`, nil)
	fillStats(metric)

	values := metric.GetValuePointers()
	assert.Equal(t, uint64(500000), uint64(values.Last.Avg.Get()))
	assert.Equal(t, uint64(20), values.ByPeriod[0].Count)
	assert.Equal(t, uint64(3000), uint64(values.ByPeriod[0].Min.Get()))
	assert.Equal(t, uint64(500), uint64((values.ByPeriod[0].Avg.Get()+5)/10))
	assert.Equal(t, uint64(5), uint64((*values.ByPeriod[0].AggregativeStatistics.GetPercentile(0.5)+500)/1000))
	assert.Equal(t, uint64(7), uint64((*values.ByPeriod[0].AggregativeStatistics.GetPercentile(0.99)+999)/1000))
	assert.Equal(t, uint64(7000), uint64(values.ByPeriod[0].Max.Get()))
	assert.Equal(t, uint64(23), values.ByPeriod[1].Count)
	assert.Equal(t, uint64(3000), uint64(values.ByPeriod[1].Min.Get()))
	assert.Equal(t, *values.ByPeriod[3], *values.ByPeriod[2])
	assert.Equal(t, *values.ByPeriod[3], *values.ByPeriod[4])
	assert.Equal(t, *values.ByPeriod[3], *values.ByPeriod[5])
	assert.Equal(t, uint64(24), values.Total.Count)
	assert.Equal(t, uint64(3000), uint64(values.Total.Min.Get()))
	assert.Equal(t, uint64(2), uint64(values.Total.Avg.Get()/10000))
	assert.Equal(t, uint64(500000), uint64(values.Total.Max.Get()))
}
