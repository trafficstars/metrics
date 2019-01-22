package metricworker

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

type GaugeFloatAggregativeValue struct {
	sync.RWMutex

	Count uint64
	Min   float64
	Mid   float64
	Avg   float64
	Per99 float64
	Max   float64
}

type GaugeFloatAggregativeValues struct {
	Last  *GaugeFloatAggregativeValue
	S1    *GaugeFloatAggregativeValue
	S5    *GaugeFloatAggregativeValue
	M1    *GaugeFloatAggregativeValue
	M5    *GaugeFloatAggregativeValue
	H1    *GaugeFloatAggregativeValue
	H6    *GaugeFloatAggregativeValue
	D1    *GaugeFloatAggregativeValue
	Total *GaugeFloatAggregativeValue
}

type floatAggregativeHistory struct {
	currentOffset uint8
	storage       [timeHistoryDepth]*GaugeFloatAggregativeValue
}

type floatAggregativeHistories struct {
	locker sync.Mutex

	S1 floatAggregativeHistory
	S5 floatAggregativeHistory
	M1 floatAggregativeHistory
	M5 floatAggregativeHistory
	H1 floatAggregativeHistory
	H6 floatAggregativeHistory
}

type workerGaugeFloatAggregative struct {
	sync.Mutex

	id         int64
	sender     MetricSender
	metricsKey string
	value      GaugeFloatAggregativeValues
	histories  floatAggregativeHistories
	//stopChan         chan bool
	stopSlicerChan   chan bool
	interval         time.Duration
	currentS1Data    *GaugeFloatAggregativeValue
	tick             uint64
	slicerLoopFuncID uint64
	senderLoopFuncID uint64
	running          uint64
	isGCEnabled      uint64
	uselessCounter   uint64
}

func NewWorkerGaugeFloatAggregative(sender MetricSender, metricsKey string) *workerGaugeFloatAggregative {
	w := &workerGaugeFloatAggregative{}
	w.id = atomic.AddInt64(&workersCount, 1)
	w.sender = sender
	w.metricsKey = fmt.Sprintf("%s,worker_id=%d", metricsKey, w.id)
	//w.stopChan = make(chan bool)
	w.stopSlicerChan = make(chan bool)
	w.currentS1Data = &GaugeFloatAggregativeValue{}
	w.value.Last = &GaugeFloatAggregativeValue{}
	w.value.S1 = &GaugeFloatAggregativeValue{}
	w.value.S5 = &GaugeFloatAggregativeValue{}
	w.value.M1 = &GaugeFloatAggregativeValue{}
	w.value.M5 = &GaugeFloatAggregativeValue{}
	w.value.H1 = &GaugeFloatAggregativeValue{}
	w.value.H6 = &GaugeFloatAggregativeValue{}
	w.value.D1 = &GaugeFloatAggregativeValue{}
	w.value.Total = &GaugeFloatAggregativeValue{}
	return w
}

func (w *workerGaugeFloatAggregative) SetGCEnabled(enabled bool) {
	if w == nil {
		return
	}
	if enabled {
		atomic.StoreUint64(&w.isGCEnabled, 1)
	} else {
		atomic.StoreUint64(&w.isGCEnabled, 0)
	}
}

func (w *workerGaugeFloatAggregative) IsGCEnabled() bool {
	if w == nil {
		return false
	}
	return atomic.LoadUint64(&w.isGCEnabled) > 0
}

func (w *workerGaugeFloatAggregative) IsRunning() bool {
	if w == nil {
		return false
	}
	return atomic.LoadUint64(&w.running) > 0
}

func (w *workerGaugeFloatAggregative) GetType() MetricType {
	return MetricTypeGauge
}

// this is so-so correct only for big amount of events (> iterationsRequiredPerSecond)
func guessPercentileFloat(curValue, newValue float64, count uint64, perc float32) float64 {
	inertness := float64(count) / iterationsRequiredPerSecond

	requireGreater := rand.Float32() > perc

	if newValue > curValue {
		if requireGreater {
			return curValue
		}
	} else {
		if !requireGreater {
			return curValue
		}
	}

	if requireGreater {
		inertness *= float64(perc)
	} else {
		inertness *= float64(1 - perc)
	}
	return (curValue*inertness + newValue) / (inertness + 1)
}

func (v *GaugeFloatAggregativeValue) RLockDo(fn func(*GaugeFloatAggregativeValue)) {
	if v == nil {
		return
	}
	data := (*GaugeFloatAggregativeValue)(atomic.LoadPointer((*unsafe.Pointer)((unsafe.Pointer)(&v))))
	if data == nil {
		return
	}
	v.RLock()
	defer v.RUnlock()
	fn(data)
}

func (v *GaugeFloatAggregativeValue) LockDo(fn func(*GaugeFloatAggregativeValue)) {
	if v == nil {
		return
	}
	data := (*GaugeFloatAggregativeValue)(atomic.LoadPointer((*unsafe.Pointer)((unsafe.Pointer)(&v))))
	if data == nil {
		return
	}
	v.Lock()
	defer v.Unlock()
	fn(data)
}

func (w *workerGaugeFloatAggregative) ConsiderValue(v float64) {
	if w == nil {
		return
	}
	appendData := func(data *GaugeFloatAggregativeValue) {
		if v < data.Min || data.Count == 0 {
			data.Min = v
		}

		if v > data.Max || data.Count == 0 {
			data.Max = v
		}

		data.Avg = (float64(data.Avg)*float64(data.Count) + float64(v)) / (float64(data.Count) + 1)
		if data.Count == 0 {
			data.Mid = v
			data.Per99 = v
		} else {
			data.Mid = guessPercentileFloat(data.Mid, v, data.Count, 0.5)
			data.Per99 = guessPercentileFloat(data.Per99, v, data.Count, 0.99)
		}

		data.Count++
	}

	curData := (*GaugeFloatAggregativeValue)(atomic.LoadPointer((*unsafe.Pointer)((unsafe.Pointer)(&w.currentS1Data))))
	curData.LockDo(appendData)

	totalData := (*GaugeFloatAggregativeValue)(atomic.LoadPointer((*unsafe.Pointer)((unsafe.Pointer)(&w.value.Total))))
	totalData.LockDo(appendData)

	lastValue := &GaugeFloatAggregativeValue{
		Count: 1,
		Min:   v,
		Mid:   v,
		Avg:   v,
		Per99: v,
		Max:   v,
	}
	atomic.StorePointer((*unsafe.Pointer)((unsafe.Pointer)(&w.value.Last)), (unsafe.Pointer)(lastValue))
}

func (w *workerGaugeFloatAggregative) GetFloat() float64 {
	if w == nil {
		return 0
	}
	return w.value.S1.Avg
}

func (w *workerGaugeFloatAggregative) Get() int64 {
	if w == nil {
		return 0
	}
	return int64(w.GetFloat())
}

func (w *workerGaugeFloatAggregative) GetValuePointers() *GaugeFloatAggregativeValues {
	if w == nil {
		return &GaugeFloatAggregativeValues{}
	}
	return &w.value
}

func (w *workerGaugeFloatAggregative) GetKey() string {
	if w == nil {
		return ``
	}
	return w.metricsKey
}

func (w *workerGaugeFloatAggregative) rotate(h *floatAggregativeHistory) {
	h.currentOffset++
	if h.currentOffset >= timeHistoryDepth {
		h.currentOffset = 0
	}
}

func (w *workerGaugeFloatAggregative) calculateValue(h *floatAggregativeHistory, depth int) (r *GaugeFloatAggregativeValue) {
	offset := h.currentOffset
	if h.storage[offset] == nil {
		return
	}

	r = &GaugeFloatAggregativeValue{}

	for depth > 0 {
		e := h.storage[offset]
		if e == nil {
			break
		}
		depth--
		offset--
		if offset == ^uint8(0) { // analog of "offset == -1", but for unsigned integer
			offset = timeHistoryDepth - 1
		}

		if (e.Min < r.Min || (r.Count == 0 && e.Count != 0)) && e.Min != 0 { // TODO: should work correctly without "e.Min != 0" but it doesn't: min value is always zero
			r.Min = e.Min
		}
		if e.Max > r.Max || (r.Count == 0 && e.Count != 0) {
			r.Max = e.Max
		}

		r.Count += e.Count

		r.Mid += e.Mid * float64(e.Count)
		r.Avg += e.Avg * float64(e.Count)
		r.Per99 += e.Per99 * float64(e.Count)
	}

	if r.Count != 0 {
		r.Mid /= float64(r.Count) // it seems to be incorrent, but I don't see other fast way to calculate it, yet
		r.Avg /= float64(r.Count)
		r.Per99 /= float64(r.Count) // it seems to be incorrent, but I don't see other fast way to calculate it, yet
	}

	return
}

func (w *workerGaugeFloatAggregative) considerFilledValue(filledValue *GaugeFloatAggregativeValue) {
	w.histories.locker.Lock()
	defer w.histories.locker.Unlock()

	tick := atomic.AddUint64(&w.tick, 1)

	updateLastHistoryRecord := func(h *floatAggregativeHistory, newValue *GaugeFloatAggregativeValue) {
		h.storage[h.currentOffset] = newValue
	}

	atomic.StorePointer((*unsafe.Pointer)((unsafe.Pointer)(&w.value.S1)), (unsafe.Pointer)(filledValue))
	w.rotate(&w.histories.S1)
	updateLastHistoryRecord(&w.histories.S1, filledValue)

	newValueS5 := w.calculateValue(&w.histories.S1, 5)
	atomic.StorePointer((*unsafe.Pointer)((unsafe.Pointer)(&w.value.S5)), (unsafe.Pointer)(newValueS5))
	if tick%5 == 0 {
		w.rotate(&w.histories.S5)
	}
	updateLastHistoryRecord(&w.histories.S5, newValueS5)

	newValueM1 := w.calculateValue(&w.histories.S5, 12)
	atomic.StorePointer((*unsafe.Pointer)((unsafe.Pointer)(&w.value.M1)), (unsafe.Pointer)(newValueM1))
	if tick%60 == 0 {
		w.rotate(&w.histories.M1)
	}
	updateLastHistoryRecord(&w.histories.M1, newValueM1)

	newValueM5 := w.calculateValue(&w.histories.M1, 5)
	atomic.StorePointer((*unsafe.Pointer)((unsafe.Pointer)(&w.value.M5)), (unsafe.Pointer)(newValueM5))
	if tick%300 == 0 {
		w.rotate(&w.histories.M5)
	}
	updateLastHistoryRecord(&w.histories.M5, newValueM5)

	newValueH1 := w.calculateValue(&w.histories.M5, 12)
	atomic.StorePointer((*unsafe.Pointer)((unsafe.Pointer)(&w.value.H1)), (unsafe.Pointer)(newValueH1))
	if tick%3600 == 0 {
		w.rotate(&w.histories.H1)
	}
	updateLastHistoryRecord(&w.histories.H1, newValueH1)

	newValueH6 := w.calculateValue(&w.histories.H1, 6)
	atomic.StorePointer((*unsafe.Pointer)((unsafe.Pointer)(&w.value.H6)), (unsafe.Pointer)(newValueH6))
	if tick%3600*6 == 0 {
		w.rotate(&w.histories.H6)
	}
	updateLastHistoryRecord(&w.histories.H6, newValueH6)

	newValueD1 := w.calculateValue(&w.histories.H6, 4)
	atomic.StorePointer((*unsafe.Pointer)((unsafe.Pointer)(&w.value.D1)), (unsafe.Pointer)(newValueD1))
}

func (w *workerGaugeFloatAggregative) doSliceNow() {
	filledValue := (*GaugeFloatAggregativeValue)(atomic.SwapPointer((*unsafe.Pointer)((unsafe.Pointer)(&w.currentS1Data)), (unsafe.Pointer)(&GaugeFloatAggregativeValue{})))
	w.considerFilledValue(filledValue)
}

/*
func (w *workerGaugeFloatAggregative) slicerLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for w.IsRunning() {
		select {
		case <-w.stopSlicerChan:
			ticker.Stop()
			return
		case <-ticker.C:
		}
		w.doSliceNow()
	}
}
*/

func (w *workerGaugeFloatAggregative) getValueForInterval(interval time.Duration) *GaugeFloatAggregativeValue {
	// TODO: remove this
	//
	// It's unobvious for user of this lib that the interval affects on value
	// that we send to the StatsD
	switch {
	case interval <= time.Second:
		return (*GaugeFloatAggregativeValue)(atomic.LoadPointer((*unsafe.Pointer)((unsafe.Pointer)(&w.value.S1))))
	case interval <= 5*time.Second:
		return (*GaugeFloatAggregativeValue)(atomic.LoadPointer((*unsafe.Pointer)((unsafe.Pointer)(&w.value.S5))))
	case interval <= time.Minute:
		return (*GaugeFloatAggregativeValue)(atomic.LoadPointer((*unsafe.Pointer)((unsafe.Pointer)(&w.value.M1))))
	case interval <= 5*time.Minute:
		return (*GaugeFloatAggregativeValue)(atomic.LoadPointer((*unsafe.Pointer)((unsafe.Pointer)(&w.value.M5))))
	case interval <= time.Hour:
		return (*GaugeFloatAggregativeValue)(atomic.LoadPointer((*unsafe.Pointer)((unsafe.Pointer)(&w.value.H1))))
	case interval <= 6*time.Hour:
		return (*GaugeFloatAggregativeValue)(atomic.LoadPointer((*unsafe.Pointer)((unsafe.Pointer)(&w.value.H6))))
	case interval <= 24*time.Hour:
		return (*GaugeFloatAggregativeValue)(atomic.LoadPointer((*unsafe.Pointer)((unsafe.Pointer)(&w.value.D1))))
	}

	return (*GaugeFloatAggregativeValue)(atomic.LoadPointer((*unsafe.Pointer)((unsafe.Pointer)(&w.value.Total))))
}

func (w *workerGaugeFloatAggregative) doSend(interval time.Duration) {
	v := w.getValueForInterval(interval)
	if v == nil {
		return
	}
	if w.IsGCEnabled() {
		if atomic.LoadUint64(&v.Count) == 0 {
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
		w.metricsKey: int(v.Avg),
	}
	w.sender.Send(string(MetricTypeGauge), dataMap) // TODO: process the returned error somehow
}

func (w *workerGaugeFloatAggregative) Run(interval time.Duration) {
	if w == nil {
		return
	}
	w.Lock()
	if w.IsRunning() {
		w.Unlock()
		return
	}
	w.interval = interval
	atomic.StoreUint64(&w.uselessCounter, 0)
	atomic.StoreUint64(&w.running, 1)

	w.slicerLoopFuncID = appendToSenderLoop(time.Second, func() {
		w.doSliceNow()
	})

	w.senderLoopFuncID = appendToSenderLoop(interval, func() {
		time.Sleep(time.Millisecond * 500) // The half of interval of slicerLoop()
		w.doSend(interval)
	})
	w.Unlock()
	return
}

func (w *workerGaugeFloatAggregative) Stop() {
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
	removeFromSenderLoop(time.Second, w.slicerLoopFuncID)
	w.interval = time.Duration(0)
	atomic.StoreUint64(&w.running, 0)
}
