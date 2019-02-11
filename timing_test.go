package metrics

import (
	"math/rand"

	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	valuesAmount       = iterationsRequiredPerSecond
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

func TestWorkerTiming(t *testing.T) {
	worker := Timing(`test`, nil)
	worker.Run(5 * time.Second)
	worker.ConsiderValue(time.Nanosecond * 5000)
	worker.DoSlice()
	worker.ConsiderValue(time.Nanosecond * 6000)
	worker.ConsiderValue(time.Nanosecond * 7000)
	worker.DoSlice()
	worker.ConsiderValue(time.Nanosecond * 3000)
	worker.ConsiderValue(time.Nanosecond * 7000)
	worker.ConsiderValue(time.Nanosecond * 4000)
	worker.ConsiderValue(time.Nanosecond * 6000)
	worker.ConsiderValue(time.Nanosecond * 5000)
	worker.ConsiderValue(time.Nanosecond * 3000)
	worker.ConsiderValue(time.Nanosecond * 7000)
	worker.ConsiderValue(time.Nanosecond * 4000)
	worker.ConsiderValue(time.Nanosecond * 6000)
	worker.ConsiderValue(time.Nanosecond * 5000)
	worker.ConsiderValue(time.Nanosecond * 3000)
	worker.ConsiderValue(time.Nanosecond * 7000)
	worker.ConsiderValue(time.Nanosecond * 4000)
	worker.ConsiderValue(time.Nanosecond * 6000)
	worker.ConsiderValue(time.Nanosecond * 5000)
	worker.ConsiderValue(time.Nanosecond * 3000)
	worker.ConsiderValue(time.Nanosecond * 7000)
	worker.ConsiderValue(time.Nanosecond * 4000)
	worker.ConsiderValue(time.Nanosecond * 6000)
	worker.ConsiderValue(time.Nanosecond * 5000)
	worker.DoSlice()
	worker.ConsiderValue(time.Nanosecond * 500000)
	worker.Stop()

	values := worker.GetValuePointers()
	assert.Equal(t, uint64(500000), values.Last.Avg)
	assert.Equal(t, uint64(20), values.ByPeriod[0].Count)
	assert.Equal(t, uint64(3000), values.ByPeriod[0].Min.Get())
	assert.Equal(t, uint64(500), (values.ByPeriod[0].Avg.Get()+5)/10)
	assert.Equal(t, uint64(5), (*values.ByPeriod[0].AggregativeStatistics.GetPercentile(0.5)+500)/1000)
	assert.Equal(t, uint64(7), (*values.ByPeriod[0].AggregativeStatistics.GetPercentile(0.99)+999)/1000)
	assert.Equal(t, uint64(7000), values.ByPeriod[0].Max)
	assert.Equal(t, uint64(23), values.ByPeriod[1].Count)
	assert.Equal(t, uint64(3000), values.ByPeriod[1].Min)
	assert.Equal(t, *values.ByPeriod[3], *values.ByPeriod[2])
	assert.Equal(t, *values.ByPeriod[3], *values.ByPeriod[4])
	assert.Equal(t, *values.ByPeriod[3], *values.ByPeriod[5])
	assert.Equal(t, uint64(24), values.Total.Count)
	assert.Equal(t, uint64(3000), values.Total.Min)
	assert.Equal(t, uint64(2), values.Total.Avg/10000)
	assert.Equal(t, uint64(500000), values.Total.Max)
}
