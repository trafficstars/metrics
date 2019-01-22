package metricworker

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type workerGaugeFloat struct {
	sync.Mutex

	id         int64
	sender     MetricSender
	metricsKey string
	valuePtr   *AtomicFloat64
	//stopChan         chan bool
	senderLoopFuncID uint64
	interval         time.Duration
	running          uint64
	isGCEnabled      uint64
	uselessCounter   uint64
}

func NewWorkerGaugeFloat(sender MetricSender, metricsKey string) *workerGaugeFloat {
	w := &workerGaugeFloat{}
	w.id = atomic.AddInt64(&workersCount, 1)
	w.sender = sender
	w.metricsKey = fmt.Sprintf("%s,worker_id=%d", metricsKey, w.id)
	w.valuePtr = &[]AtomicFloat64{AtomicFloat64(0)}[0]
	w.valuePtr.Set(0)
	//w.stopChan = make(chan bool)
	return w
}

func (w *workerGaugeFloat) SetGCEnabled(enabled bool) {
	if w == nil {
		return
	}
	if enabled {
		atomic.StoreUint64(&w.isGCEnabled, 1)
	} else {
		atomic.StoreUint64(&w.isGCEnabled, 0)
	}
}

func (w *workerGaugeFloat) IsGCEnabled() bool {
	if w == nil {
		return false
	}
	return atomic.LoadUint64(&w.isGCEnabled) > 0
}

func (w *workerGaugeFloat) IsRunning() bool {
	if w == nil {
		return false
	}
	return atomic.LoadUint64(&w.running) > 0
}

func (w *workerGaugeFloat) GetType() MetricType {
	return MetricTypeGauge
}

func (w *workerGaugeFloat) Set(newValue float64) {
	if w == nil {
		return
	}
	w.valuePtr.Set(newValue)
}

func (w *workerGaugeFloat) GetFloat() float64 {
	if w == nil {
		return 0
	}
	return w.valuePtr.Get()
}

func (w *workerGaugeFloat) Get() int64 {
	if w == nil {
		return 0
	}
	return int64(w.GetFloat())
}

func (w *workerGaugeFloat) GetKey() string {
	if w == nil {
		return ``
	}
	return w.metricsKey
}

func (w *workerGaugeFloat) SetValuePointer(newValuePtr *AtomicFloat64) {
	if w == nil {
		return
	}
	w.valuePtr = newValuePtr
}

func (w *workerGaugeFloat) doSend() {
	value := w.Get()
	if w.IsGCEnabled() {
		if value == 0 {
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
		w.metricsKey: int(value),
	}
	w.sender.Send(string(MetricTypeGauge), dataMap) // TODO: process the returned error somehow
}

func (w *workerGaugeFloat) Run(interval time.Duration) {
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

func (w *workerGaugeFloat) Stop() {
	if w == nil {
		return
	}
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
