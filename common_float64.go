package metrics

import (
	"sync/atomic"
)

// commonFloat64 is an implementation of common routines through all non-aggregative float64 metrics
type commonFloat64 struct {
	common
	modifyCounter uint64
	valuePtr      AtomicFloat64Interface
}

func (m *commonFloat64) init(r *Registry, parent Metric, key string, tags AnyTags) {
	value := AtomicFloat64(0)
	m.valuePtr = &value
	m.common.init(r, parent, key, tags, func() bool { return m.wasUseless() })
}

// Add adds (+) the value of "delta" to the internal value and returns the result
func (m *commonFloat64) Add(delta float64) float64 {
	if m == nil {
		return 0
	}
	if m.valuePtr == nil {
		return 0
	}
	r := m.valuePtr.Add(delta)
	atomic.AddUint64(&m.modifyCounter, 1)
	return r
}

// Set overwrites the internal value by the value of the argument "newValue"
func (m *commonFloat64) Set(newValue float64) {
	if m == nil {
		return
	}
	if m.valuePtr == nil {
		return
	}
	m.valuePtr.Set(newValue)
	atomic.AddUint64(&m.modifyCounter, 1)
}

// Get returns the current internal value
func (m *commonFloat64) Get() float64 {
	if m == nil {
		return 0
	}
	if m.valuePtr == nil {
		return 0
	}
	return m.valuePtr.Get()
}

// GetFloat64 returns the current internal value
//
// (the same as `Get` for float64 metrics)
func (m *commonFloat64) GetFloat64() float64 {
	return m.Get()
}

// getModifyCounterDiffFlush returns the count of modifications collected since the last call
// of this method
func (m *commonFloat64) getModifyCounterDiffFlush() uint64 {
	if m == nil {
		return 0
	}
	return atomic.SwapUint64(&m.modifyCounter, 0)
}

// SetValuePointer sets another pointer to be used to store the internal value of the metric
func (w *commonFloat64) SetValuePointer(newValuePtr *float64) {
	if w == nil {
		return
	}
	w.valuePtr = &AtomicFloat64Ptr{Pointer: newValuePtr}
}

// Send initiates a sending of the internal value via the sender
func (m *commonFloat64) Send(sender Sender) {
	if sender == nil {
		return
	}
	sender.SendFloat64(m.parent, string(m.storageKey), m.Get())
}

// wasUseless returns true if the metric's value didn't change since the last call of the method ("wasUseless")
func (w *commonFloat64) wasUseless() bool {
	return w.getModifyCounterDiffFlush() == 0
}
