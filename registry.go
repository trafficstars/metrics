package metrics

import (
	"bytes"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/xaionaro-go/atomicmap"
)

const (
	defaultIterateInterval = time.Minute
	gcUselessLimit         = 5
)

const (
	monitorState_Stopped  = 0
	monitorState_Started  = 1
	monitorState_Stopping = 2
)

var (
	registry    Registry
	defaultTags FastTags
)

type IterateIntervaler interface {
	MetricsIterateInterval() time.Duration
}

type Registry struct {
	storage                  atomicmap.Map
	isDisabled               uint32
	metricSender             *Sender
	metricsIterateIntervaler IterateIntervaler
	monitorState             uint32
	hiddenTags               *hiddenTagInternal
	defaultGCEnabled         uint32
	defaultIsRunned          uint32
}

func SetLimit(newLimit uint) {
	// for future compatibility
}

func (registry *Registry) SetDisabled(newIsDisabled bool) bool {
	newValue := uint32(0)
	if newIsDisabled {
		newValue = 1
	}

	return atomic.SwapUint32(&registry.isDisabled, newValue) != 0
}

func SetDisabled(newIsDisabled bool) bool {
	return registry.SetDisabled(newIsDisabled)
}

func (registry *Registry) IsDisabled() bool {
	return atomic.LoadUint32(&registry.isDisabled) != 0
}

func IsDisabled() bool {
	return registry.IsDisabled()
}

func (registry *Registry) GetDefaultIterateInterval() time.Duration {
	if registry.metricsIterateIntervaler == nil {
		return defaultIterateInterval
	}
	return registry.metricsIterateIntervaler.MetricsIterateInterval()
}

func GetDefaultIterateInterval() time.Duration {
	return registry.GetDefaultIterateInterval()
}

type keyGeneratorReusables struct {
	buf bytes.Buffer
}

var (
	keyGeneratorReusablesPool = sync.Pool{
		New: func() interface{} {
			return &keyGeneratorReusables{}
		},
	}
)

func newKeyGeneratorReusables() *keyGeneratorReusables {
	return keyGeneratorReusablesPool.Get().(*keyGeneratorReusables)
}

func (b *keyGeneratorReusables) Release() {
	if !memoryReuse {
		return
	}
	b.buf.Reset()
	keyGeneratorReusablesPool.Put(b)
}

func init() {
	SetDefaultGCEnabled(true)
	SetDefaultIsRunned(true)
	SetSender(nil)
}

func (registry *Registry) get(metricType Type, key string, tags AnyTags) Metric {
	considerHiddenTags(tags)
	buf := generateStorageKey(metricType, key, tags)
	rI, _ := registry.storage.GetByBytes(buf.buf.Bytes())
	buf.Release()
	if rI == nil {
		return nil
	}
	r := rI.(Metric)
	if !r.IsRunning() {
		r.Run(registry.GetDefaultIterateInterval())
		_ = registry.set(r) // may be GC already cleanup this metric, so re-set it
	}
	return r
}

func (registry *Registry) Get(metricType Type, key string, tags AnyTags) Metric {
	if registry.IsDisabled() {
		return nil
	}
	return registry.get(metricType, key, tags)
}

func (registry *Registry) set(metric Metric) error {
	_ = registry.storage.Set(metric.GetKey(), metric)
	return nil
}

func (registry *Registry) Set(metric Metric) error {
	if v, _ := registry.storage.GetByBytes(metric.GetKey()); v != nil {
		return ErrAlreadyExists
	}

	return registry.set(metric)
}

func (registry *Registry) list() *Metrics {
	result := newMetrics()
	for _, metricKey := range registry.storage.Keys() {
		metricI, _ := registry.storage.GetByBytes(metricKey.([]byte))
		if metricI == nil {
			continue
		}
		metric := metricI.(Metric)
		if !metric.IsRunning() {
			continue
		}
		*result = append(*result, metric)
	}
	return result
}

func (registry *Registry) listSorted() (result *Metrics) {
	list := registry.list()
	list.Sort()
	return list
}

func (registry *Registry) List() (result *Metrics) {
	return registry.list()
}

func (registry *Registry) remove(metric Metric) {
	metric.Stop()
	_ = registry.storage.Unset(metric.GetKey())
}

func (registry *Registry) GetSender() Sender {
	return *(*Sender)(atomic.LoadPointer((*unsafe.Pointer)((unsafe.Pointer)(&registry.metricSender))))
}

func (registry *Registry) SetSender(newMetricSender Sender) {
	atomic.StorePointer((*unsafe.Pointer)((unsafe.Pointer)(&registry.metricSender)), (unsafe.Pointer)(&newMetricSender))
}

func (registry *Registry) SetDefaultGCEnabled(newGCEnabledValue bool) {
	if newGCEnabledValue {
		atomic.StoreUint32(&registry.defaultGCEnabled, 1)
	} else {
		atomic.StoreUint32(&registry.defaultGCEnabled, 0)
	}
}

func SetDefaultGCEnabled(newValue bool) {
	registry.SetDefaultGCEnabled(newValue)
}

func (registry *Registry) GetDefaultGCEnabled() bool {
	return atomic.LoadUint32(&registry.defaultGCEnabled) != 0
}

func GetDefaultGCEnabled() bool {
	return registry.GetDefaultGCEnabled()
}

func (registry *Registry) SetDefaultIsRunned(newIsRunnedValue bool) {
	if newIsRunnedValue {
		atomic.StoreUint32(&registry.defaultIsRunned, 1)
	} else {
		atomic.StoreUint32(&registry.defaultIsRunned, 0)
	}
}

func SetDefaultIsRunned(newValue bool) {
	registry.SetDefaultIsRunned(newValue)
}

func (registry *Registry) GetDefaultIsRunned() bool {
	return atomic.LoadUint32(&registry.defaultIsRunned) != 0
}

func GetDefaultIsRunned() bool {
	return registry.GetDefaultIsRunned()
}

func (registry *Registry) getMonitorState() uint32 {
	return atomic.LoadUint32(&registry.monitorState)
}

/*func (m *MetricsRegistry) setMonitorState(newState uint64) uint64 { // returns the old state
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

	panic("Shouldn't happened, ever")
}*/

func (registry *Registry) setMonitorState(newState uint32) uint32 {
	return atomic.SwapUint32(&registry.monitorState, newState)
}

func (registry *Registry) startMonitor() {
	if v := registry.setMonitorState(monitorState_Started); v != monitorState_Stopped {
		registry.setMonitorState(v)
		return
	}

	var mem runtime.MemStats
	pausesLatency := int64(-1)

	go func() {
		interval := registry.GetDefaultIterateInterval()
		for {
			monitorState := registry.getMonitorState()
			if monitorState != monitorState_Started {
				break
			}

			runtime.ReadMemStats(&mem)

			if mem.NumGC != 0 {
				pausesLatency = int64(mem.PauseTotalNs / uint64(mem.NumGC))
			} else {
				pausesLatency = -1
			}

			time.Sleep(interval)
		}

		registry.setMonitorState(monitorState_Stopped)
	}()

	GaugeInt64Func(`runtime.memory.alloc.bytes`, Tags{`type`: `total`}, func() int64 { return int64(mem.Alloc) }).SetGCEnabled(false)
	GaugeInt64Func(`runtime.memory.alloc.bytes`, Tags{`type`: `heap`}, func() int64 { return int64(mem.HeapAlloc) }).SetGCEnabled(false)
	GaugeInt64Func(`runtime.memory.sys.bytes`, nil, func() int64 { return int64(mem.Sys) }).SetGCEnabled(false)
	GaugeInt64Func(`runtime.memory.heap_objects.count`, nil, func() int64 { return int64(mem.HeapObjects) }).SetGCEnabled(false)
	GaugeInt64Func(`runtime.gc.next.bytes`, nil, func() int64 { return int64(mem.NextGC) }).SetGCEnabled(false)
	GaugeInt64Func(`runtime.gc.pauses.ns`, nil, func() int64 { return int64(mem.PauseTotalNs) }).SetGCEnabled(false)
	GaugeInt64Func(`runtime.gc.pauses.count`, nil, func() int64 { return int64(mem.NumGC) }).SetGCEnabled(false)
	GaugeInt64Func(`runtime.gc.pauses.latency.ns`, nil, func() int64 { return pausesLatency }).SetGCEnabled(false)
	GaugeInt64Func(`runtime.cpu.fraction.e6`, nil, func() int64 { return int64(mem.GCCPUFraction * 1000000) }).SetGCEnabled(false)
}

func (registry *Registry) stopMonitor() {
	registry.setMonitorState(monitorState_Stopping)
}

func (registry *Registry) GC() {
	for _, metricKey := range registry.storage.Keys() {
		metricI, _ := registry.storage.Get(metricKey)
		if metricI == nil {
			continue
		}
		metric := metricI.(Metric)
		if !metric.IsRunning() {
			registry.remove(metric)
			metric.Release()
		}
	}
}

func GC() {
	registry.GC()
}

func (registry *Registry) Register(metric Metric, key string, inTags AnyTags) error {
	var tags *FastTags

	if inTags != nil {
		tags = newFastTags()
		inTags.Each(func(k string, v interface{}) bool {
			tags.Set(k, v)
			return true
		})
		tags.Sort()
	}

	commons := metric.(interface{ GetCommons() *common }).GetCommons()
	commons.tags = tags

	buf := generateStorageKey(metric.GetType(), key, tags)
	storageKey := buf.buf.Bytes()
	if cap(commons.storageKey) < len(storageKey) {
		commons.storageKey = make([]byte, len(storageKey))
	} else {
		commons.storageKey = commons.storageKey[:len(storageKey)]
	}
	copy(commons.storageKey, storageKey)
	buf.Release()

	return registry.Set(metric)
}

func List() *Metrics {
	return registry.List()
}

func Get(metricType Type, key string, tags AnyTags) Metric {
	return registry.Get(metricType, key, tags)
}

// copied from https://github.com/demdxx/sort-algorithms/blob/master/algorithms.go
func bubbleSort(data stringSlice) {
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
	hiddenTags := registry.getHiddenTags()
	if len(hiddenTags) == 0 {
		return
	}
	switch inTags := tags.(type) {
	case nil:
	case Tags:
		for k, v := range inTags {
			if hiddenTags.IsHiddenTag(k, v) {
				inTags.Set(k, hiddenTagValue)
			}
		}
	case *FastTags:
		for idx := range inTags.Slice {
			tag := inTags.Slice[idx]
			var s string
			if !tag.intValueIsSet {
				s = tag.StringValue
			}
			if hiddenTags.isHiddenTagByIntAndString(tag.Key, tag.intValue, s) {
				tag.intValue = 0
				tag.StringValue = hiddenTagValue
			}
		}
	default:
		inTags.Each(func(k string, v interface{}) bool {
			if hiddenTags.IsHiddenTag(k, v) {
				inTags.Set(k, hiddenTagValue)
			}
			return true
		})
	}
}

var (
	nilTags = (*FastTags)(nil)
)

func generateStorageKey(metricType Type, key string, tags AnyTags) *keyGeneratorReusables {
	reusables := newKeyGeneratorReusables()
	reusables.buf.WriteString(key)

	if tags == nil {
		tags = nilTags
	}
	if len(key) > 0 && tags.Len() > 0 {
		reusables.buf.WriteString(`,`)
	}

	tags.WriteAsString(&reusables.buf)

	if metricType > 0 {
		reusables.buf.WriteString("@")
		reusables.buf.WriteString(metricType.String())
	}

	return reusables
}

func init() {
	registry.storage = atomicmap.New()
}

// SetSender sets a handler responsible to send metric values to a metrics server (like StatsD)
func SetSender(newMetricSender Sender) {
	registry.SetSender(newMetricSender)
}

// GetSender returns the handler responsible to send metric values to a metrics server (like StatsD)
func GetSender() Sender {
	return registry.GetSender()
}

func SetMetricsIterateIntervaler(newMetricsIterateIntervaler IterateIntervaler) {
	registry.metricsIterateIntervaler = newMetricsIterateIntervaler
}

func SetDefaultTags(newDefaultAnyTags AnyTags) {
	defaultTags = *newDefaultAnyTags.ToFastTags()
}

func GetDefaultTags() *FastTags {
	return &defaultTags
}

func (registry *Registry) getHiddenTags() hiddenTagsInternal {
	result := atomic.LoadPointer((*unsafe.Pointer)((unsafe.Pointer)(&registry.hiddenTags)))
	if result == nil {
		return nil
	}
	return *(*hiddenTagsInternal)(result)
}

func (registry *Registry) SetHiddenTags(newRawHiddenTags HiddenTags) {
	var newHiddenTags hiddenTagsInternal
	if len(newRawHiddenTags) == 0 {
		newHiddenTags = nil
	} else {
		newHiddenTags = make(hiddenTagsInternal, 0, len(newRawHiddenTags))
		for _, rawHiddenTag := range newRawHiddenTags {
			newHiddenTags = append(newHiddenTags, *rawHiddenTag.toInternal())
		}
		newHiddenTags.Sort()
	}
	atomic.StorePointer((*unsafe.Pointer)((unsafe.Pointer)(&registry.hiddenTags)), (unsafe.Pointer)(&newHiddenTags))
}

func (registry *Registry) IsHiddenTag(tagKey string, tagValue interface{}) bool {
	hiddenTags := registry.getHiddenTags()
	return hiddenTags.IsHiddenTag(tagKey, tagValue)
}

func (registry *Registry) Reset() {
	for _, metricKey := range registry.storage.Keys() {
		metric, _ := registry.storage.Get(metricKey)
		if metric == nil {
			continue
		}
		metric.(Metric).Stop()
		_ = registry.storage.Unset(metricKey)
	}
}

func Reset() {
	registry.Reset()
}

func IsHiddenTag(tagKey string, tagValue interface{}) bool {
	return registry.IsHiddenTag(tagKey, tagValue)
}

func SetHiddenTags(newRawHiddenTags HiddenTags) {
	registry.SetHiddenTags(newRawHiddenTags)
}
