package metrics

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"runtime"
	_ "runtime/pprof"
	"strings"
	"sync/atomic"
	"testing"
	"time"
	_ "unsafe"

	"github.com/stretchr/testify/assert"
)

const (
	valuesAmount       = iterationsRequiredPerSecond * 1000
	permittedDeviation = 1 / (1 - 0.99) / iterationsRequiredPerSecond
)

func checkPercentile(t *testing.T, percentile float64) float64 {
	var values []float64

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
		assert.False(t, newV/oldV > (1+permittedDeviation) || oldV/newV > (1+permittedDeviation), fmt.Sprintf("Too different expected and result percentiles: %v %v", percentile, resultPercentile))
	}
}

func fillStats(t *testing.T, metric interface {
	ConsiderValue(time.Duration)
	DoSlice()
}) {
	waitForQueues := func(expectedCount uint64) uint64 {
		for {
			runtime.Gosched()
			if v := atomic.LoadUint64(&considerValueQueueCount); v >= expectedCount {
				return v
			}
		}
	}

	gosched := func() {
		for i := 0; i < 100; i++ {
			runtime.Gosched()
		}
	}

	gosched()

	queueCount := atomic.LoadUint64(&considerValueQueueCount)
	metric.ConsiderValue(time.Nanosecond * 5000)
	considerValueQueueChan <- swapConsiderValueQueue()
	queueCount = waitForQueues(queueCount + 1)

	metric.DoSlice()
	metric.ConsiderValue(time.Nanosecond * 6000)
	metric.ConsiderValue(time.Nanosecond * 7000)
	considerValueQueueChan <- swapConsiderValueQueue()
	queueCount = waitForQueues(queueCount + 1)

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
	considerValueQueueChan <- swapConsiderValueQueue()
	queueCount = waitForQueues(queueCount + 1)

	gosched()

	metric.DoSlice()
	metric.ConsiderValue(time.Nanosecond * 500000)
	considerValueQueueChan <- swapConsiderValueQueue()
	queueCount = waitForQueues(queueCount + 1)

	gosched()
}

func checkValues(t *testing.T, values *AggregativeValues) {
	assert.Equal(t, uint64(1), values.Last().Count.Get())
	assert.Equal(t, uint64(500000), uint64(values.Last().Avg.Get()))
	assert.Equal(t, uint64(500000), uint64(values.Last().Sum.Get()))
	assert.Equal(t, uint64(60), values.ByPeriod(0).Count.Get())
	assert.Equal(t, float64(12*(3000+4000+5000+6000+7000)), values.ByPeriod(0).Sum.Get(), values.ByPeriod(0))
	assert.Equal(t, uint64(3000), uint64(values.ByPeriod(0).Min.Get()))
	assert.Equal(t, uint64(500), uint64((values.ByPeriod(0).Avg.Get()+5)/10))
	assert.Equal(t, uint64(7), uint64((*values.ByPeriod(0).AggregativeStatistics.GetPercentile(0.99)+999)/1000))
	assert.Equal(t, uint64(3), uint64((*values.ByPeriod(1).AggregativeStatistics.GetPercentile(0.99)+999)/2000))
	assert.Equal(t, uint64(7000), uint64(values.ByPeriod(0).Max.Get()))
	assert.Equal(t, uint64(63), values.ByPeriod(1).Count.Get())
	assert.Equal(t, uint64(3000), uint64(values.ByPeriod(1).Min.Get()))
	assert.Equal(t, strings.Split(values.ByPeriod(3).String(), "sum")[0], strings.Split(values.ByPeriod(2).String(), "sum")[0])
	assert.Equal(t, strings.Split(values.ByPeriod(3).String(), "sum")[0], strings.Split(values.ByPeriod(4).String(), "sum")[0])
	assert.Equal(t, strings.Split(values.ByPeriod(3).String(), "sum")[0], strings.Split(values.ByPeriod(5).String(), "sum")[0])
	assert.Equal(t, uint64(64), values.Total().Count.Get())
	assert.Equal(t, uint64(3000), uint64(values.Total().Min.Get()))
	assert.Equal(t, uint64(1), uint64(values.Total().Avg.Get()/10000))
	assert.Equal(t, uint64(500000), uint64(values.Total().Max.Get()))
	assert.Equal(t, float64(500000+12*(3000+4000+5000+6000+7000)+7000+6000+5000), values.Total().Sum.Get())
}

func TestTimingBuffered(t *testing.T) {
	r := New()
	defer r.Reset()
	r.SetDefaultIsRan(false)
	r.SetDefaultGCEnabled(false)
	metric := r.TimingBuffered(`test`, nil)
	fillStats(t, metric)
	checkValues(t, metric.GetValuePointers())
}

func TestTimingFlow(t *testing.T) {
	r := New()
	defer r.Reset()
	r.SetDefaultIsRan(false)
	r.SetDefaultGCEnabled(false)
	metric := r.TimingFlow(`test`, nil)
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

type memProfileRecordKey struct {
	AllocBytes int64
	Stack      [32]uintptr
}

func heapDiff(old, new []runtime.MemProfileRecord) []runtime.MemProfileRecord {
	if len(old) == 0 || len(new) == 0 {
		return new
	}

	oldMap := make(map[memProfileRecordKey]struct{}, len(old))
	for _, p := range old {
		oldMap[memProfileRecordKey{
			AllocBytes: p.AllocBytes,
			Stack:      p.Stack0,
		}] = struct{}{}
	}

	var result []runtime.MemProfileRecord
	for _, p := range new {
		if _, ok := oldMap[memProfileRecordKey{
			AllocBytes: p.AllocBytes,
			Stack:      p.Stack0,
		}]; ok {
			continue
		}

		f := runtime.FuncForPC(p.Stack0[0])
		switch f.Name() {
		case "sync.(*Pool).pinSlow", "github.com/xaionaro-go/metrics.(*iterationHandler).Remove":
			continue
		}

		result = append(result, p)
	}

	return result
}

func formatHeapProfile(p []runtime.MemProfileRecord) string {
	var buf bytes.Buffer

	// copied from runtime/pprof/pprof.go:
	//
	// Copyright 2010 The Go Authors. All rights reserved.
	// Use of this source code is governed by a BSD-style
	// license that can be found in the LICENSE file.

	for i := range p {
		r := &p[i]
		fmt.Fprintf(&buf, "%d: %d [%d: %d] @",
			r.InUseObjects(), r.InUseBytes(),
			r.AllocObjects, r.AllocBytes)
		for _, pc := range r.Stack() {
			fmt.Fprintf(&buf, " %#x", pc)
		}
		fmt.Fprintf(&buf, "\n")
		formatStackRecord(&buf, r.Stack(), false)
	}
	return buf.String()
}

// formatStackRecord prints the function + source line information
// for a single stack trace.
//
// copied from runtime/pprof/pprof.go
//
// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
func formatStackRecord(w io.Writer, stk []uintptr, allFrames bool) {
	show := allFrames
	frames := runtime.CallersFrames(stk)
	for {
		frame, more := frames.Next()
		name := frame.Function
		if name == "" {
			show = true
			fmt.Fprintf(w, "#\t%#x\n", frame.PC)
		} else if name != "runtime.goexit" && (show || !strings.HasPrefix(name, "runtime.")) {
			// Hide runtime.goexit and any runtime functions at the beginning.
			// This is useful mainly for allocation traces.
			show = true
			fmt.Fprintf(w, "#\t%#x\t%s+%#x\t%s:%d\n", frame.PC, name, frame.PC-frame.Entry, frame.File, frame.Line)
		}
		if !more {
			break
		}
	}
	if !show {
		// We didn't print anything; do it again,
		// and this time include runtime functions.
		formatStackRecord(w, stk, true)
		return
	}
	fmt.Fprintf(w, "\n")
}

func testGC(t *testing.T, fn func()) {
	gc := func() {
		GC()
		for i := 0; i < 3; i++ {
			runtime.Gosched()
			runtime.GC()
		}
	}

	oldHeap, newHeap := make([]runtime.MemProfileRecord, 512), make([]runtime.MemProfileRecord, 512)

	SetMemoryReuseEnabled(false)
	Reset()

	fn()
	Reset()
	gc()

	keys := registry.storage.Keys()
	if len(keys) != 0 {
		t.Errorf(`len(keys) == %v\n`, keys)
	}
	gc()
	{
		n, _ := runtime.MemProfile(oldHeap, true)
		oldHeap = oldHeap[:n]
	}
	var memstats, cleanedMemstats runtime.MemStats
	goroutinesCount := runtime.NumGoroutine()
	gc()
	runtime.ReadMemStats(&memstats)
	fn()
	gc()
	keys = registry.storage.Keys()
	if len(keys) != 0 {
		t.Errorf(`len(keys) == %v\n`, keys)
	}
	cleanedGoroutinesCount := runtime.NumGoroutine()
	//assert.Equal(t, int64(0), iterationHandlers.routinesCount)
	assert.Equal(t, goroutinesCount, cleanedGoroutinesCount)

	{
		n, _ := runtime.MemProfile(newHeap, true)
		newHeap = newHeap[:n]
	}
	heapDiffResult := heapDiff(oldHeap, newHeap)
	assert.Zero(t, len(heapDiffResult), formatHeapProfile(heapDiffResult))

	for i := 0; i < 100000; i++ {
		fn()
	}
	GC()
	runtime.GC()
	runtime.ReadMemStats(&cleanedMemstats)

	if !assert.True(t, (int64(cleanedMemstats.HeapInuse)-int64(memstats.HeapInuse))/100000 < 1) {
		t.Error(cleanedMemstats.HeapInuse, int64(cleanedMemstats.HeapInuse)-int64(memstats.HeapInuse))
	}

	SetMemoryReuseEnabled(true)
}

func TestTimingBufferedGC(t *testing.T) {
	r := New()
	defer r.Reset()
	r.SetDefaultGCEnabled(false)
	r.SetDefaultIsRan(false)
	testGC(t, func() {
		metric := r.TimingBuffered(`test_gc`, nil)
		metric.Stop()
	})
}

func TestTimingFlowGC(t *testing.T) {
	r := New()
	defer r.Reset()
	r.SetDefaultGCEnabled(false)
	r.SetDefaultIsRan(false)
	testGC(t, func() {
		metric := TimingFlow(`test_gc`, nil)
		metric.Stop()
	})
}
