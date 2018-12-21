package metricworker

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

const (
	timeHistoryDepth = 12 // see (*workerTiming).considerFilledValues(). It's the maximal constant.

	// If this value is too large then it will be required too many events
	// per second to calculate percentile correctly (Mid and Per99).
	// If this value is too small then the percentile will be calculated
	// not accurate.
	iterationsRequiredPerSecond = 20
)

type TimingValue struct {
	sync.RWMutex

	Count uint64
	Min   uint64
	Mid   uint64
	Avg   uint64
	Per99 uint64
	Max   uint64
}

type TimingValues struct {
	Last  *TimingValue
	S1    *TimingValue
	S5    *TimingValue
	M1    *TimingValue
	M5    *TimingValue
	H1    *TimingValue
	H6    *TimingValue
	D1    *TimingValue
	Total *TimingValue
}

type timingHistory struct {
	currentOffset uint8
	storage       [timeHistoryDepth]*TimingValue
}

type timingHistories struct {
	locker sync.Mutex

	S1 timingHistory
	S5 timingHistory
	M1 timingHistory
	M5 timingHistory
	H1 timingHistory
	H6 timingHistory
}

type workerTiming struct {
	sync.Mutex

	id         int64
	sender     MetricSender
	metricsKey string
	value      TimingValues
	histories  timingHistories
	//stopChan         chan bool
	stopSlicerChan   chan bool
	interval         time.Duration
	currentS1Data    *TimingValue
	tick             uint64
	slicerLoopFuncID uint64
	senderLoopFuncID uint64
	running          uint64
	isGCEnabled      uint64
	uselessCounter   uint64
}

func NewWorkerTiming(sender MetricSender, metricsKey string) *workerTiming {
	w := &workerTiming{}
	w.id = atomic.AddInt64(&workersCount, 1)
	w.sender = sender
	w.metricsKey = fmt.Sprintf("%s,worker_id=%d", metricsKey, w.id)
	//w.stopChan = make(chan bool)
	w.stopSlicerChan = make(chan bool)
	w.currentS1Data = &TimingValue{}
	w.value.Last = &TimingValue{}
	w.value.S1 = &TimingValue{}
	w.value.S5 = &TimingValue{}
	w.value.M1 = &TimingValue{}
	w.value.M5 = &TimingValue{}
	w.value.H1 = &TimingValue{}
	w.value.H6 = &TimingValue{}
	w.value.D1 = &TimingValue{}
	w.value.Total = &TimingValue{}
	return w
}

func (w *workerTiming) SetGCEnabled(enabled bool) {
	if enabled {
		atomic.StoreUint64(&w.isGCEnabled, 1)
	} else {
		atomic.StoreUint64(&w.isGCEnabled, 0)
	}
}

func (w *workerTiming) IsGCEnabled() bool {
	return atomic.LoadUint64(&w.isGCEnabled) > 0
}

func (w *workerTiming) IsRunning() bool {
	return atomic.LoadUint64(&w.running) > 0
}

func (w *workerTiming) GetType() MetricType {
	return MetricTypeTiming
}

// this is so-so correct only for big amount of events (> iterationsRequiredPerSecond)
func guessPercentile(curValue uint64, newValue, count uint64, perc float32) uint64 {
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
	updatedValue := uint64((float64(curValue)*inertness + float64(newValue)) / (inertness + 1))
	return updatedValue
}

func (v *TimingValue) RLockDo(fn func(*TimingValue)) {
	data := (*TimingValue)(atomic.LoadPointer((*unsafe.Pointer)((unsafe.Pointer)(&v))))
	if data == nil {
		return
	}
	v.RLock()
	defer v.RUnlock()
	fn(data)
}

func (v *TimingValue) LockDo(fn func(*TimingValue)) {
	data := (*TimingValue)(atomic.LoadPointer((*unsafe.Pointer)((unsafe.Pointer)(&v))))
	if data == nil {
		return
	}
	v.Lock()
	defer v.Unlock()
	fn(data)
}

func (w *workerTiming) ConsiderValue(d time.Duration) {
	v := uint64(d.Nanoseconds())

	appendData := func(data *TimingValue) {
		if v < data.Min || data.Count == 0 {
			data.Min = v
		}

		if v > data.Max || data.Count == 0 {
			data.Max = v
		}

		data.Avg = uint64((float64(data.Avg)*float64(data.Count) + float64(v)) / (float64(data.Count) + 1))
		if data.Count == 0 {
			data.Mid = v
			data.Per99 = v
		} else {
			data.Mid = guessPercentile(data.Mid, v, data.Count, 0.5)
			data.Per99 = guessPercentile(data.Per99, v, data.Count, 0.99)
		}

		data.Count++
	}

	curData := (*TimingValue)(atomic.LoadPointer((*unsafe.Pointer)((unsafe.Pointer)(&w.currentS1Data))))
	curData.LockDo(appendData)

	totalData := (*TimingValue)(atomic.LoadPointer((*unsafe.Pointer)((unsafe.Pointer)(&w.value.Total))))
	totalData.LockDo(appendData)

	lastValue := &TimingValue{
		Count: 1,
		Min:   v,
		Mid:   v,
		Avg:   v,
		Per99: v,
		Max:   v,
	}
	atomic.StorePointer((*unsafe.Pointer)((unsafe.Pointer)(&w.value.Last)), (unsafe.Pointer)(lastValue))
}

func (w *workerTiming) Get() int64 {
	return int64(atomic.LoadUint64(&w.value.S1.Avg))
}

func (w *workerTiming) GetValuePointers() *TimingValues {
	return &w.value
}

func (w *workerTiming) GetKey() string {
	return w.metricsKey
}

func (w *workerTiming) rotate(h *timingHistory) {
	h.currentOffset++
	if h.currentOffset >= timeHistoryDepth {
		h.currentOffset = 0
	}
}

func (w *workerTiming) calculateValue(h *timingHistory, depth int) (r *TimingValue) {
	offset := h.currentOffset
	if h.storage[offset] == nil {
		return
	}

	r = &TimingValue{}

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

		r.Mid += e.Mid * e.Count
		r.Avg += e.Avg * e.Count
		r.Per99 += e.Per99 * e.Count
	}

	if r.Count != 0 {
		r.Mid /= r.Count // it seems to be incorrent, but I don't see other fast way to calculate it, yet
		r.Avg /= r.Count
		r.Per99 /= r.Count // it seems to be incorrent, but I don't see other fast way to calculate it, yet
	}

	return
}

func (w *workerTiming) considerFilledValue(filledValue *TimingValue) {
	w.histories.locker.Lock()
	defer w.histories.locker.Unlock()

	tick := atomic.AddUint64(&w.tick, 1)

	updateLastHistoryRecord := func(h *timingHistory, newValue *TimingValue) {
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

func (w *workerTiming) doSliceNow() {
	filledValue := (*TimingValue)(atomic.SwapPointer((*unsafe.Pointer)((unsafe.Pointer)(&w.currentS1Data)), (unsafe.Pointer)(&TimingValue{})))
	w.considerFilledValue(filledValue)
}

/*
func (w *workerTiming) slicerLoop(interval time.Duration) {
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

func (w *workerTiming) getValueForInterval(interval time.Duration) *TimingValue {
	// TODO: remove this
	//
	// It's unobvious for user of this lib that the interval affects on value
	// that we send to the StatsD
	switch {
	case interval <= time.Second:
		return (*TimingValue)(atomic.LoadPointer((*unsafe.Pointer)((unsafe.Pointer)(&w.value.S1))))
	case interval <= 5*time.Second:
		return (*TimingValue)(atomic.LoadPointer((*unsafe.Pointer)((unsafe.Pointer)(&w.value.S5))))
	case interval <= time.Minute:
		return (*TimingValue)(atomic.LoadPointer((*unsafe.Pointer)((unsafe.Pointer)(&w.value.M1))))
	case interval <= 5*time.Minute:
		return (*TimingValue)(atomic.LoadPointer((*unsafe.Pointer)((unsafe.Pointer)(&w.value.M5))))
	case interval <= time.Hour:
		return (*TimingValue)(atomic.LoadPointer((*unsafe.Pointer)((unsafe.Pointer)(&w.value.H1))))
	case interval <= 6*time.Hour:
		return (*TimingValue)(atomic.LoadPointer((*unsafe.Pointer)((unsafe.Pointer)(&w.value.H6))))
	case interval <= 24*time.Hour:
		return (*TimingValue)(atomic.LoadPointer((*unsafe.Pointer)((unsafe.Pointer)(&w.value.D1))))
	}

	return (*TimingValue)(atomic.LoadPointer((*unsafe.Pointer)((unsafe.Pointer)(&w.value.Total))))
}

func (w *workerTiming) doSend(interval time.Duration) {
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
		w.metricsKey: int(atomic.LoadUint64(&v.Avg)),
	}
	w.sender.Send(string(MetricTypeTiming), dataMap) // TODO: process the returned error somehow
}

func (w *workerTiming) Run(interval time.Duration) {
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

func (w *workerTiming) Stop() {
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
