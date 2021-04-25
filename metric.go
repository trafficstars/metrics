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
	GetKey() []byte
	GetType() Type
	GetName() string
	GetTags() *FastTags
	GetFloat64() float64
	IsRunning() bool
	Release()
	IsGCEnabled() bool
	SetGCEnabled(bool)
	GetTag(string) interface{}
	Registry() *Registry

	run(time.Duration)
	stop()
	lock()
	unlock()
}
