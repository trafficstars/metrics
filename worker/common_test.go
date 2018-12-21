package metricworker

import (
	"testing"
)

func TestTypes(t *testing.T) {
	var worker Worker

	worker = &workerCount{}
	worker = &workerGauge{}
	worker = &workerGaugeFunc{}
	worker = &workerTiming{}

	_ = worker
}
