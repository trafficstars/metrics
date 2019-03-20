package metrics

import (
	"sync/atomic"
)

type metricCommonInt64 struct {
	metricCommon
	valuePtr      *int64
	previousValue int64
}

func (m *metricCommonInt64) init(parent Metric, key string, tags AnyTags) {
	m.valuePtr = &[]int64{0}[0]
	m.metricCommon.init(parent, key, tags, func() bool { return m.wasUseless() })
}

func (w *metricCommonInt64) Increment() int64 {
	if w == nil {
		return 0
	}
	return atomic.AddInt64(w.valuePtr, 1)
}

func (w *metricCommonInt64) Add(delta int64) int64 {
	return atomic.AddInt64(w.valuePtr, delta)
}

func (w *metricCommonInt64) Set(newValue int64) {
	if w == nil {
		return
	}
	atomic.StoreInt64(w.valuePtr, newValue)
}

func (w *metricCommonInt64) Get() int64 {
	if w == nil {
		return 0
	}
	return atomic.LoadInt64(w.valuePtr)
}

func (w *metricCommonInt64) GetFloat64() float64 {
	if w == nil {
		return 0
	}
	return float64(w.Get())
}

func (w *metricCommonInt64) getDifferenceFlush() int64 {
	if w == nil {
		return 0
	}
	newValue := w.Get()
	previousValue := atomic.SwapInt64(&w.previousValue, newValue)
	return newValue - previousValue
}

func (w *metricCommonInt64) SetValuePointer(newValuePtr *int64) {
	if w == nil {
		return
	}
	w.valuePtr = newValuePtr
}

func (m *metricCommonInt64) Send(sender Sender) {
	if sender == nil {
		return
	}
	sender.SendUint64(m.parent, string(m.storageKey), uint64(m.Get()))
}

func (w *metricCommonInt64) wasUseless() bool {
	return w.getDifferenceFlush() == 0
}
