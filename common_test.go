package metrics

import (
	"runtime/debug"
)

var (
	_ Metric = &MetricCount{}
	_ Metric = &MetricGaugeInt64{}
	_ Metric = &MetricGaugeInt64Func{}
	_ Metric = &MetricGaugeFloat64{}
	_ Metric = &MetricGaugeFloat64Func{}
	_ Metric = &MetricTimingBuffered{}
	_ Metric = &MetricTimingFlow{}
	_ Metric = &MetricTimingSimple{}
)

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
