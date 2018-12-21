package metricworker

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type workerGaugeFunc struct {
	sync.Mutex

	id         int64
	sender     MetricSender
	metricsKey string
	fn         func() int64
	//stopChan         chan bool
	senderLoopFuncID uint64
	interval         time.Duration
	running          uint64
	isGCEnabled      uint64
	uselessCounter   uint64
}

func NewWorkerGaugeFunc(sender MetricSender, metricsKey string, fn func() int64) *workerGaugeFunc {
	w := &workerGaugeFunc{}
	w.id = atomic.AddInt64(&workersCount, 1)
	w.sender = sender
	w.metricsKey = fmt.Sprintf("%s,worker_id=%d", metricsKey, w.id)
	w.fn = fn
	//w.stopChan = make(chan bool)
	return w
}

func (w *workerGaugeFunc) SetGCEnabled(enabled bool) {
	if enabled {
		atomic.StoreUint64(&w.isGCEnabled, 1)
	} else {
		atomic.StoreUint64(&w.isGCEnabled, 0)
	}
}

func (w *workerGaugeFunc) IsGCEnabled() bool {
	return atomic.LoadUint64(&w.isGCEnabled) > 0
}

func (w *workerGaugeFunc) IsRunning() bool {
	return atomic.LoadUint64(&w.running) > 0
}

func (w *workerGaugeFunc) GetType() MetricType {
	return MetricTypeGauge
}

func (w *workerGaugeFunc) Get() int64 {
	return w.fn()
}

func (w *workerGaugeFunc) GetKey() string {
	return w.metricsKey
}

func (w *workerGaugeFunc) doSend() {
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

func (w *workerGaugeFunc) Run(interval time.Duration) {
	w.Lock()
	defer w.Unlock()
	if w.IsRunning() {
		return
	}
	w.senderLoopFuncID = appendToSenderLoop(interval, func() {
		w.doSend()
	})
	atomic.StoreUint64(&w.uselessCounter, 0)
	w.interval = interval
	atomic.StoreUint64(&w.running, 1)
	return
}

func (w *workerGaugeFunc) Stop() {
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
