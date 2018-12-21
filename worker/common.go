package metricworker

import (
	"time"
)

var (
	workersCount int64
)

type MetricType string

const (
	gcUselessLimit = 3600
)

const (
	MetricTypeIncrement = MetricType("increment")
	MetricTypeTiming    = MetricType("timing")
	MetricTypeGauge     = MetricType("gauge")
	MetricTypeUnique    = MetricType("unique")
	MetricTypeCount     = MetricType("count")
)

type MetricSender interface {
	Send(key string, value interface{}) error
}

type Worker interface {
	Get() int64
	GetKey() string
	GetType() MetricType
	Stop()
	Run(interval time.Duration)
	IsRunning() bool
	SetGCEnabled(bool)
}
