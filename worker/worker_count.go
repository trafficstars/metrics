package metricworker

import (
	"sync"
	"sync/atomic"
	"time"
)

type workerCount struct {
	sync.Mutex

	id            int64
	sender        MetricSender
	metricsKey    string
	valuePtr      *uint64
	previousValue uint64
	//stopChan         chan bool
	senderLoopFuncID uint64
	interval         time.Duration
	running          uint64
	isGCEnabled      uint64
	uselessCounter   uint64
}

func NewWorkerCount(sender MetricSender, metricsKey string) *workerCount {
	w := &workerCount{}
	w.id = atomic.AddInt64(&workersCount, 1)
	w.sender = sender
	w.metricsKey = metricsKey
	w.valuePtr = &[]uint64{0}[0]
	//w.stopChan = make(chan bool)
	return w
}

func (w *workerCount) GetType() MetricType {
	return MetricTypeCount
}

func (w *workerCount) IsRunning() bool {
	if w == nil {
		return false
	}
	return atomic.LoadUint64(&w.running) > 0
}

func (w *workerCount) SetGCEnabled(enabled bool) {
	if w == nil {
		return
	}
	if enabled {
		atomic.StoreUint64(&w.isGCEnabled, 1)
	} else {
		atomic.StoreUint64(&w.isGCEnabled, 0)
	}
}

func (w *workerCount) IsGCEnabled() bool {
	if w == nil {
		return false
	}
	return atomic.LoadUint64(&w.isGCEnabled) > 0
}

func (w *workerCount) Increment() uint64 {
	if w == nil {
		return 0
	}
	return atomic.AddUint64(w.valuePtr, 1)
}

func (w *workerCount) Add(delta uint64) uint64 {
	return atomic.AddUint64(w.valuePtr, delta)
}

func (w *workerCount) Set(newValue uint64) {
	if w == nil {
		return
	}
	atomic.StoreUint64(w.valuePtr, newValue)
}

func (w *workerCount) Get() int64 {
	if w == nil {
		return 0
	}
	return int64(w.GetValue())
}

func (w *workerCount) GetDifferenceFlush() uint64 {
	if w == nil {
		return 0
	}
	newValue := w.GetValue()
	previousValue := atomic.SwapUint64(&w.previousValue, newValue)
	return newValue - previousValue
}

func (w *workerCount) GetValue() uint64 {
	if w == nil {
		return 0
	}
	return atomic.LoadUint64(w.valuePtr)
}

func (w *workerCount) GetKey() string {
	if w == nil {
		return ``
	}
	return w.metricsKey
}

func (w *workerCount) SetValuePointer(newValuePtr *uint64) {
	if w == nil {
		return
	}
	w.valuePtr = newValuePtr
}

func (w *workerCount) doSend() {
	diff := w.GetDifferenceFlush()
	if w.IsGCEnabled() {
		if diff == 0 {
			if atomic.AddUint64(&w.uselessCounter, 1) > gcUselessLimit {
				if w.IsRunning() {
					go w.Stop()
				}
			}
			return
		} else {
			atomic.StoreUint64(&w.uselessCounter, 0)
		}
	}
	if w.sender == nil {
		return
	}
	dataMap := map[string]int{
		w.metricsKey: int(diff),
	}
	w.sender.Send(string(MetricTypeCount), dataMap) // TODO: process the returned error somehow
}

func (w *workerCount) Run(interval time.Duration) {
	if w == nil {
		return
	}
	w.Lock()
	defer w.Unlock()
	if w.IsRunning() {
		return
	}
	w.senderLoopFuncID = appendToSenderLoop(interval, func() {
		w.doSend()
	})
	w.interval = interval
	atomic.StoreUint64(&w.uselessCounter, 0)
	atomic.StoreUint64(&w.running, 1)
	return
}

func (w *workerCount) Stop() {
	if !w.IsRunning() {
		return
	}
	w.Lock()
	defer w.Unlock()
	//w.stopChan <- true
	removeFromSenderLoop(w.interval, w.senderLoopFuncID)
	w.interval = time.Duration(0)
	atomic.StoreUint64(&w.running, 0)
}
