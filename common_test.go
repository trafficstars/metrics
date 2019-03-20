package metrics

import (
	"runtime/debug"
	"testing"
)

func TestTypes(t *testing.T) {
	var metric Metric

	metric = &MetricCount{}
	metric = &MetricGaugeInt64{}
	metric = &MetricGaugeInt64Func{}
	metric = &MetricGaugeFloat64{}
	metric = &MetricGaugeFloat64Func{}
	metric = &MetricTimingBuffered{}
	metric = &MetricTimingFlow{}

	_ = metric
}

func checkForInfiniteRecursion(m Metric) {
	// it will panic if there's an infinite recursion

	debug.SetMaxStack(50000)

	m.Iterate()
	m.GetInterval()
	m.Run(0)
	m.Send(nil)
	m.GetType()
	m.GetName()
	m.GetTags()
	m.GetFloat64()
	m.IsRunning()
	m.Release()
	m.SetGCEnabled(true)
	m.GetTag(``)
	m.Stop()
}
