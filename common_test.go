package metrics

import (
	"testing"
)

func TestTypes(t *testing.T) {
	var metric Metric

	metric = &MetricCount{}
	metric = &MetricGaugeInt64{}
	metric = &MetricGaugeInt64Func{}
	metric = &MetricGaugeFloat64{}
	metric = &MetricGaugeFloat64Func{}
	metric = &MetricTiming{}

	_ = metric
}
