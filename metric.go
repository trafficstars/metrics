package metrics

import (
	"time"
)

type Metric interface {
	Iterate()
	GetInterval() time.Duration
	Run(time.Duration)
	Stop()
	Send(Sender)
	GetType() Type
	GetName() string
	GetTags() Tags
	GetFloat64() float64
	IsRunning() bool
	Release()
	SetGCEnabled(bool)
	GetTag(string) interface{}
}
