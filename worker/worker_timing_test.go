package metricworker

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

func checkPercentile(t *testing.T, percentile float32) float32 {
	values := []int{}

	for i := 0; i < valuesAmount; i++ {
		r := rand.Intn(1000)
		values = append(values, r*r)
	}

	var result uint64
	for idx, v := range values {
		result = guessPercentile(result, uint64(v), uint64(idx), percentile)
	}

	count := 0
	for _, v := range values {
		if v < int(result) {
			count++
		}
	}

	return float32(count) / valuesAmount
}

func TestGuessPercentile(t *testing.T) {
	for _, percentile := range []float32{0.5, 0.9, 0.99} {
		resultPercentile := checkPercentile(t, percentile)
		oldV := percentile / (1 - percentile)
		newV := resultPercentile / (1 - resultPercentile)
		if newV/oldV > (1+permittedDeviation) || oldV/newV > (1+permittedDeviation) {
			t.Errorf("Too different expected and result percentiles: %v %v", percentile, resultPercentile)
		}
	}
}

func TestWorkerTiming(t *testing.T) {
	worker := NewWorkerTiming(nil, `test`)
	worker.Run(5 * time.Second)
	worker.ConsiderValue(time.Nanosecond * 5000)
	worker.doSliceNow()
	worker.ConsiderValue(time.Nanosecond * 6000)
	worker.ConsiderValue(time.Nanosecond * 7000)
	worker.doSliceNow()
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
	worker.doSliceNow()
	worker.ConsiderValue(time.Nanosecond * 500000)
	worker.Stop()

	values := worker.GetValuePointers()
	assert.Equal(t, uint64(500000), values.Last.Avg)
	assert.Equal(t, uint64(20), values.S1.Count)
	assert.Equal(t, uint64(3000), values.S1.Min)
	assert.Equal(t, uint64(500), (values.S1.Avg+5)/10)
	assert.Equal(t, uint64(5), (values.S1.Mid+500)/1000)
	assert.Equal(t, uint64(7), (values.S1.Per99+999)/1000)
	assert.Equal(t, uint64(7000), values.S1.Max)
	assert.Equal(t, uint64(23), values.S5.Count)
	assert.Equal(t, uint64(3000), values.S5.Min)
	assert.Equal(t, *values.S5, *values.M1)
	assert.Equal(t, *values.S5, *values.H1)
	assert.Equal(t, *values.S5, *values.D1)
	assert.Equal(t, uint64(24), values.Total.Count)
	assert.Equal(t, uint64(3000), values.Total.Min)
	assert.Equal(t, uint64(2), values.Total.Avg/10000)
	assert.Equal(t, uint64(500000), values.Total.Max)
}
