package metrics

import (
	"sync/atomic"
	"unsafe"
)

// commonFloat64 is an implementation of common routines through all non-aggregative int64 metrics
type commonInt64 struct {
	common
	valuePtr      *int64
	previousValue int64
}

func (m *commonInt64) init(r *Registry, parent Metric, key string, tags AnyTags) {
	m.valuePtr = &[]int64{0}[0]
	m.common.init(r, parent, key, tags, func() bool { return m.wasUseless() })
}

// Increment is an analog of Add(1). It just adds "1" to the internal value and returns the result.
func (m *commonInt64) Increment() int64 {
	return m.Add(1)
}

// Add adds (+) the value of "delta" to the internal value and returns the result
func (m *commonInt64) Add(delta int64) int64 {
	if m == nil {
		return 0
	}
	if m.valuePtr == nil {
		atomic.StorePointer((*unsafe.Pointer)((unsafe.Pointer)(&m.valuePtr)), (unsafe.Pointer)(&[]int64{0}[0]))
	}
	return atomic.AddInt64(m.valuePtr, delta)
}

// Set overwrites the internal value by the value of the argument "newValue"
func (m *commonInt64) Set(newValue int64) {
	if m == nil {
		return
	}
	if m.valuePtr == nil {
		atomic.StorePointer((*unsafe.Pointer)((unsafe.Pointer)(&m.valuePtr)), (unsafe.Pointer)(&[]int64{0}[0]))
	}
	atomic.StoreInt64(m.valuePtr, newValue)
}

// Get returns the current internal value
func (m *commonInt64) Get() int64 {
	if m == nil {
		return 0
	}
	if m.valuePtr == nil {
		atomic.StorePointer((*unsafe.Pointer)((unsafe.Pointer)(&m.valuePtr)), (unsafe.Pointer)(&[]int64{0}[0]))
	}
	return atomic.LoadInt64(m.valuePtr)
}

// GetFloat64 returns the current internal value as float64 (the same as `float64(Get())`)
func (m *commonInt64) GetFloat64() float64 {
	return float64(m.Get())
}

// getDifferenceFlush returns the difference of the current internal value and the previous internal the value
// (when was the last call of method "getDifferenceFlush")
func (m *commonInt64) getDifferenceFlush() int64 {
	if m == nil {
		return 0
	}
	newValue := m.Get()
	previousValue := atomic.SwapInt64(&m.previousValue, newValue)
	return newValue - previousValue
}

// SetValuePointer sets another pointer to be used to store the internal value of the metric
func (m *commonInt64) SetValuePointer(newValuePtr *int64) {
	if m == nil {
		return
	}
	m.valuePtr = newValuePtr
}

// Send initiates a sending of the internal value via the sender
func (m *commonInt64) Send(sender Sender) {
	if sender == nil {
		return
	}
	sender.SendUint64(m.parent, string(m.storageKey), uint64(m.Get()))
}

// wasUseless returns true if the metric's value didn't change since the last call of the method ("wasUseless")
func (m *commonInt64) wasUseless() bool {
	return m.getDifferenceFlush() == 0
}
