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
)

const (
	defaultIterateInterval = 10 * time.Second
	gcUselessLimit         = 5
)

const (
	monitorState_Stopped  = 0
	monitorState_Starting = 1
	monitorState_Started  = 2
	monitorState_Stopping = 3
)

var (
	metricsRegistry MetricsRegistry
	defaultTags     FastTags
)

type MetricsIterateIntervaler interface {
	MetricsIterateInterval() time.Duration
}

type MetricsRegistry struct {
	storage    atomicmap.Map
	isDisabled uint64
	//getterCache           atomicmap.Map
	metricSender             Sender
	metricsIterateIntervaler MetricsIterateIntervaler
	monitorState             uint64
	hiddenTags               *[]string
	defaultGCEnabled         uint32
	defaultIsRunned          uint32
}

func (m *MetricsRegistry) SetDisabled(newIsDisabled bool) bool {
	newValue := uint64(0)
	if newIsDisabled {
		newValue = 1
	}

	return atomic.SwapUint64(&m.isDisabled, newValue) != 0
}

func SetDisabled(newIsDisabled bool) bool {
	return metricsRegistry.SetDisabled(newIsDisabled)
}

func (m *MetricsRegistry) IsDisabled() bool {
	return atomic.LoadUint64(&m.isDisabled) != 0
}

func IsDisabled() bool {
	return metricsRegistry.IsDisabled()
}

func (m *MetricsRegistry) GetDefaultIterateInterval() time.Duration {
	if m.metricsIterateIntervaler == nil {
		return defaultIterateInterval
	}
	return m.metricsIterateIntervaler.MetricsIterateInterval()
}

func GetDefaultIterateInterval() time.Duration {
	return metricsRegistry.GetDefaultIterateInterval()
}

type preallocatedStringerBuffer struct {
	locker  int32
	result  bytes.Buffer
	tagKeys sort.StringSlice
}

var (
	stringBufferPool = sync.Pool{
		New: func() interface{} {
			return &preallocatedStringerBuffer{}
		},
	}
)

func newStringBuffer() *preallocatedStringerBuffer {
	return stringBufferPool.Get().(*preallocatedStringerBuffer)
}

func (b *preallocatedStringerBuffer) Release() {
	b.result.Reset()
	stringBufferPool.Put(b)
}

/*type preallocatedGetterBuffer struct {
	sync.Mutex
	pc     []uintptr
	result bytes.Buffer
}*/

func init() {
	SetDefaultGCEnabled(true)
	SetDefaultIsRunned(true)
}

/*func (m *MetricsRegistry) getCallerCacheKey() *preallocatedGetterBuffer {
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

func (m *MetricsRegistry) get(metricType Type, key string, tags AnyTags) Metric {
	considerHiddenTags(tags)
	storageKeyBuf := generateStorageKey(metricType, key, tags)
	rI, _ := m.storage.GetByBytes(storageKeyBuf.result.Bytes())
	storageKeyBuf.Release()
	if rI == nil {
		return nil
	}
	r := rI.(Metric)
	if !r.IsRunning() {
		if !r.IsRunning() {
			r.Run(metricsRegistry.GetDefaultIterateInterval())
		}
		m.set(r) // may be GC already cleanup this metric, so re-set it
	}
	return r
}

/*func (m *MetricsRegistry) getWithCache(metricType Type, key string, tags AnyTags) Metric {
	buf := m.getCallerCacheKey()
	if metric, _ := m.getterCache.Get(buf.result.Bytes()); metric != nil {
		buf.Unlock()
		return metric.(Metric)
	}
	if metric := m.get(metricType, key, tags); metric != nil {
		m.getterCache.Set(buf.result.Bytes(), metric)
		buf.Unlock()
		return metric
	}
	buf.Unlock()
	return nil
}*/

func (m *MetricsRegistry) Get(metricType Type, key string, tags AnyTags) Metric {
	if m.IsDisabled() {
		return nil
	}
	return m.get(metricType, key, tags)
	//return m.getWithCache(metricType, key, tags)
}

func (m *MetricsRegistry) set(metric Metric) error {
	m.storage.Set(metric.(interface{ GetKey() []byte }).GetKey(), metric)
	return nil
}

func (m *MetricsRegistry) Set(metric Metric) error {
	if v, _ := m.storage.GetByBytes(metric.(interface{ GetKey() []byte }).GetKey()); v != nil {
		return ErrAlreadyExists
	}

	return m.set(metric)
}

func (m *MetricsRegistry) list() (result Metrics) {
	for _, metricKey := range m.storage.Keys() {
		metric, _ := m.storage.GetByBytes(metricKey.([]byte))
		if metric == nil {
			continue
		}
		result = append(result, metric.(Metric))
	}
	return
}

func (m *MetricsRegistry) listSorted() (result Metrics) {
	list := m.list()
	list.Sort()
	return list
}

func (m *MetricsRegistry) List() (result Metrics) {
	return m.listSorted()
	//return m.list()
}

func (m *MetricsRegistry) remove(metric Metric) {
	metric.Stop()
	m.storage.LockUnset(metric.(interface{ GetKey() []byte }).GetKey())
}

func (m *MetricsRegistry) GetSender() Sender {
	return m.metricSender
}

func (m *MetricsRegistry) SetDefaultSender(newMetricSender Sender) {
	m.stopMonitor()
	m.metricSender = newMetricSender
	m.startMonitor()
}
func (m *MetricsRegistry) GetDefaultSender() Sender {
	return m.metricSender
}

func (m *MetricsRegistry) SetDefaultGCEnabled(newGCEnabledValue bool) {
	if newGCEnabledValue {
		atomic.StoreUint32(&m.defaultGCEnabled, 1)
	} else {
		atomic.StoreUint32(&m.defaultGCEnabled, 0)
	}
}

func SetDefaultGCEnabled(newValue bool) {
	metricsRegistry.SetDefaultGCEnabled(newValue)
}

func (m *MetricsRegistry) GetDefaultGCEnabled() bool {
	return atomic.LoadUint32(&m.defaultGCEnabled) != 0
}

func GetDefaultGCEnabled() bool {
	return metricsRegistry.GetDefaultGCEnabled()
}

func (m *MetricsRegistry) SetDefaultIsRunned(newIsRunnedValue bool) {
	if newIsRunnedValue {
		atomic.StoreUint32(&m.defaultIsRunned, 1)
	} else {
		atomic.StoreUint32(&m.defaultIsRunned, 0)
	}
}

func SetDefaultIsRunned(newValue bool) {
	metricsRegistry.SetDefaultIsRunned(newValue)
}

func (m *MetricsRegistry) GetDefaultIsRunned() bool {
	return atomic.LoadUint32(&m.defaultIsRunned) != 0
}

func GetDefaultIsRunned() bool {
	return metricsRegistry.GetDefaultIsRunned()
}

func (m *MetricsRegistry) getMonitorState() uint64 {
	return atomic.LoadUint64(&m.monitorState)
}

func (m *MetricsRegistry) setMonitorState(newState uint64) uint64 { // returns the old state
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

func (m *MetricsRegistry) startMonitor() {
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
		interval := m.GetDefaultIterateInterval()
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

	GaugeInt64Func(`runtime.memory.alloc.bytes`, Tags{`type`: `total`}, func() int64 { return int64(mem.Alloc) }).SetGCEnabled(false)
	GaugeInt64Func(`runtime.memory.alloc.bytes`, Tags{`type`: `heap`}, func() int64 { return int64(mem.HeapAlloc) }).SetGCEnabled(false)
	GaugeInt64Func(`runtime.memory.sys.bytes`, nil, func() int64 { return int64(mem.Sys) }).SetGCEnabled(false)
	GaugeInt64Func(`runtime.memory.heap_objects.count`, nil, func() int64 { return int64(mem.HeapObjects) }).SetGCEnabled(false)
	GaugeInt64Func(`runtime.gc.next.bytes`, nil, func() int64 { return int64(mem.NextGC) }).SetGCEnabled(false)
	GaugeInt64Func(`runtime.gc.pauses.ns`, nil, func() int64 { return int64(mem.PauseTotalNs) }).SetGCEnabled(false)
	GaugeInt64Func(`runtime.gc.pauses.count`, nil, func() int64 { return int64(mem.NumGC) }).SetGCEnabled(false)
	GaugeInt64Func(`runtime.gc.pauses.latency.ns`, nil, func() int64 { return pauses_latency }).SetGCEnabled(false)
	GaugeInt64Func(`runtime.cpu.fraction.e6`, nil, func() int64 { return int64(mem.GCCPUFraction * 1000000) }).SetGCEnabled(false)
}

func (m *MetricsRegistry) stopMonitor() {
	return

	m.setMonitorState(monitorState_Stopping)
}

func (m *MetricsRegistry) GC() {
	for _, metricKey := range m.storage.Keys() {
		metricI, _ := m.storage.Get(metricKey)
		if metricI == nil {
			continue
		}
		metric := metricI.(Metric)
		if !metric.IsRunning() {
			m.remove(metric)
		}
		metric.Release()
	}
}

func GC() {
	metricsRegistry.GC()
}

func (metricsRegistry *MetricsRegistry) Register(metric Metric, key string, inTags AnyTags) error {
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

	commons := metric.(interface{ GetCommons() *metricCommon }).GetCommons()
	commons.tags = tags

	keyBuf := generateStorageKey(metric.GetType(), key, tags)
	storageKey := keyBuf.result.String()
	keyBuf.Release()

	commons.storageKey = make([]byte, len(storageKey))
	copy(commons.storageKey, storageKey)
	return metricsRegistry.Set(metric)
}

func List() []Metric {
	return metricsRegistry.List()
}

func Get(metricType Type, key string, tags AnyTags) Metric {
	return metricsRegistry.Get(metricType, key, tags)
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

func generateStorageKey(metricType Type, key string, tags AnyTags) *preallocatedStringerBuffer {
	// It's required to avoid memory allocations. So if we allocated a buffer once, we reuse it.

	buf := newStringBuffer()
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

	if metricType > 0 {
		buf.result.WriteString("@")
		buf.result.WriteString(metricType.String())
	}

	return buf
}

func init() {
	metricsRegistry.storage = atomicmap.New()
	//metricsRegistry.getterCache = atomicmap.New()
}

func SetDefaultSender(newMetricSender Sender) {
	metricsRegistry.SetDefaultSender(newMetricSender)
}

func GetDefaultSender() Sender {
	return metricsRegistry.GetDefaultSender()
}

func SetIntervaler(newMetricsIterateIntervaler MetricsIterateIntervaler) {
	metricsRegistry.metricsIterateIntervaler = newMetricsIterateIntervaler
}

func SetDefaultTags(newDefaultAnyTags AnyTags) {
	defaultTags = *newDefaultAnyTags.ToFastTags()
}

func (m *MetricsRegistry) GetHiddenTags() []string {
	result := atomic.LoadPointer((*unsafe.Pointer)((unsafe.Pointer)(&m.hiddenTags)))
	if result == nil {
		return nil
	}
	return *(*[]string)(result)
}

func (m *MetricsRegistry) SetHiddenTags(newHiddenTags []string) {
	atomic.StorePointer((*unsafe.Pointer)((unsafe.Pointer)(&m.hiddenTags)), (unsafe.Pointer)(&newHiddenTags))
}

func (m *MetricsRegistry) IsHiddenTag(tagKey string) bool {
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

func (m *MetricsRegistry) Reset() {
	for _, metricKey := range m.storage.Keys() {
		metric, _ := m.storage.Get(metricKey)
		if metric == nil {
			continue
		}
		metric.(Metric).Stop()
		m.storage.LockUnset(metricKey)
	}
}

func Reset() {
	metricsRegistry.Reset()
}

func IsHiddenTag(tagKey string) bool {
	return metricsRegistry.IsHiddenTag(tagKey)
}

func GetHiddenTags() []string {
	return metricsRegistry.GetHiddenTags()
}

func SetHiddenTags(hiddenTags []string) {
	if len(hiddenTags) == 0 {
		hiddenTags = nil
	}
	sort.Strings(hiddenTags) // It's required to use binary search in IsHiddenTag()
	metricsRegistry.SetHiddenTags(hiddenTags)
}
