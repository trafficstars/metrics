package metrics

import (
	"bytes"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/xaionaro-go/atomicmap"

	"github.com/trafficstars/fastmetrics/worker"
)

const (
	defaultSendInterval       = 10 * time.Second
	gcMetricExpirationTimeout = time.Hour
)

const (
	monitorState_Stopped  = 0
	monitorState_Starting = 1
	monitorState_Started  = 2
	monitorState_Stopping = 3
)

var (
	metrics     Metrics
	defaultTags FastTags
)

type MetricsSendIntervaler interface {
	MetricsSendInterval() time.Duration
}

type Metrics struct {
	storage    atomicmap.Map
	isDisabled uint64
	//getterCache           atomicmap.Map
	metricSender          metricworker.MetricSender
	metricsSendIntervaler MetricsSendIntervaler
	monitorState          uint64
	hiddenTags            *[]string
}

func (m *Metrics) SetDisabled(newIsDisabled bool) bool {
	newValue := uint64(0)
	if newIsDisabled {
		newValue = 1
	}

	return atomic.SwapUint64(&m.isDisabled, newValue) != 0
}

func SetDisabled(newIsDisabled bool) bool {
	metrics.SetDisabled(newIsDisabled)
}

func (m *Metrics) IsDisabled() bool {
	return atomic.LoadUint64(&m.isDisabled) != 0
}

func IsDisabled() bool {
	return metrics.IsDisabled()
}

func (m *Metrics) GetSendInterval() time.Duration {
	if m.metricsSendIntervaler == nil {
		return defaultSendInterval
	}
	return m.metricsSendIntervaler.MetricsSendInterval()
}

type preallocatedStringerBuffer struct {
	locker  int32
	result  bytes.Buffer
	tagKeys sort.StringSlice
}

func (buf *preallocatedStringerBuffer) Lock() {
	for !atomic.CompareAndSwapInt32(&buf.locker, 0, 1) {
		runtime.Gosched()
	}
}

func (buf *preallocatedStringerBuffer) Unlock() {
	atomic.AddInt32(&buf.locker, -1)
}

type preallocatedGetterBuffer struct {
	sync.Mutex
	pc     []uintptr
	result bytes.Buffer
}

var (
	stringerBufs       [maxConcurrency]*preallocatedStringerBuffer
	stringerBufPointer uint64

/*	getterBufs       [maxConcurrency]*preallocatedGetterBuffer
	getterBufPointer uint64*/
)

func init() {
	for i := 0; i < maxConcurrency; i++ {
		stringerBufs[i] = &preallocatedStringerBuffer{}
		/*getterBufs[i] = &preallocatedGetterBuffer{
			pc: make([]uintptr, 8),
		}*/
	}
}

/*func (m *Metrics) getCallerCacheKey() *preallocatedGetterBuffer {
	curBufPointer := atomic.AddUint64(&getterBufPointer, 1)
	curBufPointer %= maxConcurrency
	buf := getterBufs[curBufPointer]
	buf.Lock()

	buf.result.Reset()
	runtime.Callers(0, buf.pc)
	for _, c := range buf.pc {
		i := uint64(c)
		buf.result.WriteByte(byte(i << 56))
		buf.result.WriteByte(byte(i << 48))
		buf.result.WriteByte(byte(i << 40))
		buf.result.WriteByte(byte(i << 32))
		buf.result.WriteByte(byte(i << 24))
		buf.result.WriteByte(byte(i << 16))
		buf.result.WriteByte(byte(i << 8))
		buf.result.WriteByte(byte(i))
	}
	return buf
}*/

func (m *Metrics) get(metricType MetricType, key string, tags AnyTags) *Metric {
	considerHiddenTags(tags)
	storageKeyBuf := generateStorageKey(metricType, key, tags)
	rI, _ := m.storage.GetByBytes(storageKeyBuf.result.Bytes())
	storageKeyBuf.Unlock()
	if rI == nil {
		return nil
	}
	r := rI.(*Metric)
	if !r.IsRunning() {
		if !r.IsRunning() {
			r.worker.Run(metrics.GetSendInterval())
		}
		m.set(r) // may be GC already cleanup this metric, so re-set it
	}
	return r
}

/*func (m *Metrics) getWithCache(metricType MetricType, key string, tags AnyTags) *Metric {
	buf := m.getCallerCacheKey()
	if metric, _ := m.getterCache.Get(buf.result.Bytes()); metric != nil {
		buf.Unlock()
		return metric.(*Metric)
	}
	if metric := m.get(metricType, key, tags); metric != nil {
		m.getterCache.Set(buf.result.Bytes(), metric)
		buf.Unlock()
		return metric
	}
	buf.Unlock()
	return nil
}*/

func (m *Metrics) Get(metricType MetricType, key string, tags AnyTags) *Metric {
	if m.IsDisabled() {
		return nil
	}
	return m.get(metricType, key, tags)
	//return m.getWithCache(metricType, key, tags)
}

func (m *Metrics) set(metric *Metric) error {
	m.storage.Set(metric.storageKey, metric)
	return nil
}

func (m *Metrics) Set(metric *Metric) error {
	if v, _ := m.storage.Get(metric.storageKey); v != nil {
		return ErrAlreadyExists
	}

	return m.set(metric)
}

func (m *Metrics) list() (result []*Metric) {
	for _, metricKey := range m.storage.Keys() {
		metric, _ := m.storage.Get(metricKey)
		if metric == nil {
			continue
		}
		result = append(result, metric.(*Metric))
	}
	return
}

func (m *Metrics) listSorted() (result []*Metric) {
	list := m.list()
	sort.Slice(list, func(i, j int) bool {
		if string(list[i].storageKey) < string(list[j].storageKey) {
			return true
		}
		return false
	})
	return list
}

func (m *Metrics) List() (result []*Metric) {
	return m.listSorted()
	//return m.list()
}

func (m *Metrics) remove(metric *Metric) {
	metric.Stop()
	m.storage.LockUnset(metric.storageKey)
}

func (m *Metrics) GetSender() metricworker.MetricSender {
	return m.metricSender
}

func (m *Metrics) SetSender(newMetricSender metricworker.MetricSender) {
	m.stopMonitor()
	m.metricSender = newMetricSender
	m.startMonitor()
}

func (m *Metrics) getMonitorState() uint64 {
	return atomic.LoadUint64(&m.monitorState)
}

func (m *Metrics) setMonitorState(newState uint64) uint64 { // returns the old state
	// FSM state switcher
	if atomic.LoadUint64(&m.monitorState) == newState {
		return newState
	}
	switch newState {
	case monitorState_Starting:
		for {
			if atomic.LoadUint64(&m.monitorState) == monitorState_Starting {
				return monitorState_Starting
			}
			if atomic.LoadUint64(&m.monitorState) == monitorState_Started {
				return monitorState_Started
			}
			if atomic.CompareAndSwapUint64(&m.monitorState, monitorState_Stopped, monitorState_Starting) {
				return monitorState_Stopped
			}
		}
	case monitorState_Started:
		for {
			if atomic.LoadUint64(&m.monitorState) == monitorState_Started {
				return monitorState_Started
			}
			if atomic.CompareAndSwapUint64(&m.monitorState, monitorState_Starting, monitorState_Started) {
				return monitorState_Starting
			}
		}
	case monitorState_Stopping:
		for {
			if atomic.LoadUint64(&m.monitorState) == monitorState_Stopping {
				return monitorState_Stopping
			}
			if atomic.LoadUint64(&m.monitorState) == monitorState_Stopped {
				return monitorState_Stopped
			}
			if atomic.CompareAndSwapUint64(&m.monitorState, monitorState_Started, monitorState_Stopping) {
				return monitorState_Started
			}
		}
	case monitorState_Stopped:
		for {
			if atomic.LoadUint64(&m.monitorState) == monitorState_Stopped {
				return monitorState_Stopped
			}
			if atomic.CompareAndSwapUint64(&m.monitorState, monitorState_Stopping, monitorState_Stopped) {
				return monitorState_Stopping
			}
		}
	}

	panic("Shouldn't happend, ever")
	return 0
}

func (m *Metrics) startMonitor() {
	return

	switch m.setMonitorState(monitorState_Starting) {
	case monitorState_Starting:
		// setMonitorState doesn't change from monitorStatus_Runningto monitorStatus_Starting
		// so there's no need to fix the "state" here.
		return // already starting
	case monitorState_Started:
		return // already started
	}

	var mem runtime.MemStats
	pauses_latency := int64(-1)

	go func() {
		interval := m.GetSendInterval()
		for m.getMonitorState() == monitorState_Started {
			runtime.ReadMemStats(&mem)

			if mem.NumGC != 0 {
				pauses_latency = int64(mem.PauseTotalNs / uint64(mem.NumGC))
			} else {
				pauses_latency = -1
			}

			time.Sleep(interval)
		}
		m.setMonitorState(monitorState_Stopped)
	}()

	CreateOrGetWorkerGaugeFunc(`runtime.memory.alloc.bytes`, Tags{`type`: `total`}, func() int64 { return int64(mem.Alloc) }).SetGCEnabled(false)
	CreateOrGetWorkerGaugeFunc(`runtime.memory.alloc.bytes`, Tags{`type`: `heap`}, func() int64 { return int64(mem.HeapAlloc) }).SetGCEnabled(false)
	CreateOrGetWorkerGaugeFunc(`runtime.memory.sys.bytes`, nil, func() int64 { return int64(mem.Sys) }).SetGCEnabled(false)
	CreateOrGetWorkerGaugeFunc(`runtime.memory.heap_objects.count`, nil, func() int64 { return int64(mem.HeapObjects) }).SetGCEnabled(false)
	CreateOrGetWorkerGaugeFunc(`runtime.gc.next.bytes`, nil, func() int64 { return int64(mem.NextGC) }).SetGCEnabled(false)
	CreateOrGetWorkerGaugeFunc(`runtime.gc.pauses.ns`, nil, func() int64 { return int64(mem.PauseTotalNs) }).SetGCEnabled(false)
	CreateOrGetWorkerGaugeFunc(`runtime.gc.pauses.count`, nil, func() int64 { return int64(mem.NumGC) }).SetGCEnabled(false)
	CreateOrGetWorkerGaugeFunc(`runtime.gc.pauses.latency.ns`, nil, func() int64 { return pauses_latency }).SetGCEnabled(false)
	CreateOrGetWorkerGaugeFunc(`runtime.cpu.fraction.e6`, nil, func() int64 { return int64(mem.GCCPUFraction * 1000000) }).SetGCEnabled(false)
}

func (m *Metrics) stopMonitor() {
	return

	m.setMonitorState(monitorState_Stopping)
}

func (m *Metrics) GC() {
	for _, metricKey := range m.storage.Keys() {
		metricI, _ := m.storage.Get(metricKey)
		if metricI == nil {
			continue
		}
		metric := metricI.(*Metric)
		if !metric.IsRunning() {
			m.remove(metric)
		}
	}
}

func GC() {
	metrics.GC()
}

func register(name string, worker Worker, description string, inTags AnyTags) error {
	tags := Tags{}
	if inTags != nil {
		inTags.Each(func(k string, v interface{}) bool {
			tags[k] = v
			return true
		})
	}
	for _, tag := range defaultTags {
		tags[tag.Key] = tag.Value
	}

	metric := &Metric{
		worker:      worker,
		name:        name,
		tags:        tags,
		description: description,
	}

	metric.considerHiddenTags()
	storageKeyBuf := metric.generateStorageKey()
	storageKey := storageKeyBuf.result.Bytes()
	metric.storageKey = make([]byte, len(storageKey))
	copy(metric.storageKey, storageKey)
	storageKeyBuf.Unlock()
	return metrics.Set(metric)
}

func runAndRegister(key string, worker Worker, tags AnyTags) error {
	worker.Run(metrics.GetSendInterval())
	return register(key, worker, "", tags)
}

func List() []*Metric {
	return metrics.List()
}

func Get(metricType MetricType, key string, tags AnyTags) *Metric {
	return metrics.Get(metricType, key, tags)
}

// copied from https://github.com/demdxx/sort-algorithms/blob/master/algorithms.go
func BubbleSort(data sort.StringSlice) {
	n := data.Len() - 1
	b := false
	for i := 0; i < n; i++ {
		for j := 0; j < n-i; j++ {
			if data.Less(j+1, j) {
				data.Swap(j+1, j)
				b = true
			}
		}
		if !b {
			break
		}
		b = false
	}
}

func considerHiddenTags(tags AnyTags) {
	switch inTags := tags.(type) {
	case nil:
	case Tags:
		for k, _ := range inTags {
			if IsHiddenTag(k) {
				inTags.Set(k, hiddenTagValue)
			}
		}
	case *FastTags:
		for _, tag := range *inTags {
			if IsHiddenTag(tag.Key) {
				tag.Value = hiddenTagValue
			}
		}
	default:
		inTags.Each(func(k string, v interface{}) bool {
			if IsHiddenTag(k) {
				inTags.Set(k, hiddenTagValue)
			}
			return true
		})
	}
}

func generateStorageKey(metricType MetricType, key string, tags AnyTags) *preallocatedStringerBuffer {
	// It's required to avoid memory allocations. So if we allocated a buffer once, we reuse it.
	// We have buffers (of amount maxConcurrency) to be able to process this function concurrently.

	curBufPointer := atomic.AddUint64(&stringerBufPointer, 1)
	curBufPointer %= maxConcurrency
	buf := stringerBufs[curBufPointer]
	buf.Lock()

	buf.result.Reset()

	buf.result.WriteString(key)

	for _, tag := range defaultTags {
		buf.result.WriteString(`,`)
		buf.result.WriteString(tag.Key)
		buf.result.WriteString(`=`)
		buf.result.Write(tag.Value)
	}

	switch inTags := tags.(type) {
	case nil:
	case Tags:
		buf.tagKeys = buf.tagKeys[:0]
		for k, _ := range inTags {
			if defaultTags.IsSet(k) {
				continue
			}
			buf.tagKeys = append(buf.tagKeys, k)
		}
		if len(buf.tagKeys) > 0 {
			if len(buf.tagKeys) > 24 {
				sort.Strings(buf.tagKeys) // It requires to wrap the slice into an interface, so it has a memory allocation
			} else {
				BubbleSort(buf.tagKeys)
			}
		}

		for _, k := range buf.tagKeys {
			buf.result.WriteString(`,`)
			buf.result.WriteString(k)
			buf.result.WriteString(`=`)
			buf.result.WriteString(TagValueToString(inTags[k]))
		}
	case *FastTags:
		for _, tag := range *inTags {
			if defaultTags.IsSet(tag.Key) {
				continue
			}
			buf.result.WriteString(`,`)
			buf.result.WriteString(tag.Key)
			buf.result.WriteString(`=`)
			buf.result.Write(tag.Value)
		}
	default:
		buf.tagKeys = buf.tagKeys[:0]
		inTags.Each(func(k string, v interface{}) bool {
			if defaultTags.IsSet(k) {
				return true
			}
			buf.tagKeys = append(buf.tagKeys, k)
			return true
		})
		if len(buf.tagKeys) > 0 {
			if len(buf.tagKeys) > 24 {
				sort.Strings(buf.tagKeys) // It requires to wrap the slice into an interface, so it has a memory allocation
			} else {
				BubbleSort(buf.tagKeys)
			}
		}

		for _, k := range buf.tagKeys {
			buf.result.WriteString(`,`)
			buf.result.WriteString(k)
			buf.result.WriteString(`=`)
			buf.result.WriteString(TagValueToString(inTags.Get(k)))
		}
	}

	if len(metricType) > 0 {
		buf.result.WriteString("@")
		buf.result.WriteString(string(metricType))
	}

	return buf
}

func init() {
	metrics.storage = atomicmap.New()
	//metrics.getterCache = atomicmap.New()
}

func Init(newMetricSender metricworker.MetricSender, newMetricsSendIntervaler MetricsSendIntervaler, newDefaultAnyTags AnyTags) {
	metrics.SetSender(newMetricSender)
	metrics.metricsSendIntervaler = newMetricsSendIntervaler
	defaultTags = *newDefaultAnyTags.ToFastTags()
}

func (m *Metrics) GetHiddenTags() []string {
	result := atomic.LoadPointer((*unsafe.Pointer)((unsafe.Pointer)(&m.hiddenTags)))
	if result == nil {
		return nil
	}
	return *(*[]string)(result)
}

func (m *Metrics) SetHiddenTags(newHiddenTags []string) {
	atomic.StorePointer((*unsafe.Pointer)((unsafe.Pointer)(&m.hiddenTags)), (unsafe.Pointer)(&newHiddenTags))
}

func (m *Metrics) IsHiddenTag(tagKey string) bool {
	hiddenTags := m.GetHiddenTags()
	l := len(hiddenTags)
	idx := sort.Search(l, func(i int) bool {
		return hiddenTags[i] >= tagKey
	})

	if idx < 0 || idx >= l {
		return false
	}

	if hiddenTags[idx] != tagKey {
		return false
	}

	return true
}

func IsHiddenTag(tagKey string) bool {
	return metrics.IsHiddenTag(tagKey)
}

func GetHiddenTags() []string {
	return metrics.GetHiddenTags()
}

func SetHiddenTags(hiddenTags []string) {
	if len(hiddenTags) == 0 {
		hiddenTags = nil
	}
	sort.Strings(hiddenTags) // It's required to use binary search in IsHiddenTag()
	metrics.SetHiddenTags(hiddenTags)
}
