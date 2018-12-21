package metrics

import (
	"github.com/trafficstars/fastmetrics/worker"
)

type MetricType = metricworker.MetricType

const (
	MetricTypeGauge  = metricworker.MetricTypeGauge
	MetricTypeCount  = metricworker.MetricTypeCount
	MetricTypeTiming = metricworker.MetricTypeTiming
)
