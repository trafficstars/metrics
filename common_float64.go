package metrics

import (
	"sync/atomic"
)

// commonFloat64
type commonFloat64 struct {
	common
	modifyCounter uint64
	valuePtr      AtomicFloat64Interface
}

func (m *commonFloat64) init(parent Metric, key string, tags AnyTags) {
	value := AtomicFloat64(0)
	m.valuePtr = &value
	m.common.init(parent, key, tags, func() bool { return m.wasUseless() })
}

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

func (m *commonFloat64) Get() float64 {
	if m == nil {
		return 0
	}
	if m.valuePtr == nil {
		return 0
	}
	return m.valuePtr.Get()
}

func (m *commonFloat64) GetFloat64() float64 {
	return m.Get()
}

func (m *commonFloat64) getModifyCounterDiffFlush() uint64 {
	if m == nil {
		return 0
	}
	return atomic.SwapUint64(&m.modifyCounter, 0)
}

func (w *commonFloat64) SetValuePointer(newValuePtr *float64) {
	if w == nil {
		return
	}
	w.valuePtr = &AtomicFloat64Ptr{Pointer: newValuePtr}
}

func (m *commonFloat64) Send(sender Sender) {
	if sender == nil {
		return
	}
	sender.SendFloat64(m.parent, string(m.storageKey), m.Get())
}

func (w *commonFloat64) wasUseless() bool {
	return w.getModifyCounterDiffFlush() == 0
}
