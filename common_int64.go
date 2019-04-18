package metrics

import (
	"sync/atomic"
	"unsafe"
)

type commonInt64 struct {
	common
	valuePtr      *int64
	previousValue int64
}

func (m *commonInt64) init(parent Metric, key string, tags AnyTags) {
	m.valuePtr = &[]int64{0}[0]
	m.common.init(parent, key, tags, func() bool { return m.wasUseless() })
}

func (m *commonInt64) Increment() int64 {
	if m == nil {
		return 0
	}
	if m.valuePtr == nil {
		atomic.StorePointer((*unsafe.Pointer)((unsafe.Pointer)(&m.valuePtr)), (unsafe.Pointer)(&[]int64{0}[0]))
	}
	return atomic.AddInt64(m.valuePtr, 1)
}

func (m *commonInt64) Add(delta int64) int64 {
	if m == nil {
		return 0
	}
	if m.valuePtr == nil {
		atomic.StorePointer((*unsafe.Pointer)((unsafe.Pointer)(&m.valuePtr)), (unsafe.Pointer)(&[]int64{0}[0]))
	}
	return atomic.AddInt64(m.valuePtr, delta)
}

func (m *commonInt64) Set(newValue int64) {
	if m == nil {
		return
	}
	if m.valuePtr == nil {
		atomic.StorePointer((*unsafe.Pointer)((unsafe.Pointer)(&m.valuePtr)), (unsafe.Pointer)(&[]int64{0}[0]))
	}
	atomic.StoreInt64(m.valuePtr, newValue)
}

func (m *commonInt64) Get() int64 {
	if m == nil {
		return 0
	}
	if m.valuePtr == nil {
		atomic.StorePointer((*unsafe.Pointer)((unsafe.Pointer)(&m.valuePtr)), (unsafe.Pointer)(&[]int64{0}[0]))
	}
	return atomic.LoadInt64(m.valuePtr)
}

func (m *commonInt64) GetFloat64() float64 {
	if m == nil {
		return 0
	}
	return float64(m.Get())
}

func (m *commonInt64) getDifferenceFlush() int64 {
	if m == nil {
		return 0
	}
	newValue := m.Get()
	previousValue := atomic.SwapInt64(&m.previousValue, newValue)
	return newValue - previousValue
}

func (m *commonInt64) SetValuePointer(newValuePtr *int64) {
	if m == nil {
		return
	}
	m.valuePtr = newValuePtr
}

func (m *commonInt64) Send(sender Sender) {
	if sender == nil {
		return
	}
	sender.SendUint64(m.parent, string(m.storageKey), uint64(m.Get()))
}

func (m *commonInt64) wasUseless() bool {
	return m.getDifferenceFlush() == 0
}
