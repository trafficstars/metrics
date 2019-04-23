package metrics

import (
	"bytes"
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
	registry    Registry
	defaultTags FastTags
)

type IterateIntervaler interface {
	MetricsIterateInterval() time.Duration
}

type Registry struct {
	storage                  atomicmap.Map
	isDisabled               uint64
	metricSender             Sender
	metricsIterateIntervaler IterateIntervaler
	monitorState             uint64
	hiddenTags               *hiddenTagInternal
	defaultGCEnabled         uint32
	defaultIsRunned          uint32
}

func SetLimit(newLimit uint) {
	// for future compatibility
}

func (registry *Registry) SetDisabled(newIsDisabled bool) bool {
	newValue := uint64(0)
	if newIsDisabled {
		newValue = 1
	}

	return atomic.SwapUint64(&registry.isDisabled, newValue) != 0
}

func SetDisabled(newIsDisabled bool) bool {
	return registry.SetDisabled(newIsDisabled)
}

func (registry *Registry) IsDisabled() bool {
	return atomic.LoadUint64(&registry.isDisabled) != 0
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
		return nil
		/* Works unstable, commented for a while:
		if !r.IsRunning() {
			r.Run(metricsRegistry.GetDefaultIterateInterval())
		}
		m.set(r) // may be GC already cleanup this metric, so re-set it
		*/
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
	registry.storage.Set(metric.(interface{ GetKey() []byte }).GetKey(), metric)
	return nil
}

func (registry *Registry) Set(metric Metric) error {
	if v, _ := registry.storage.GetByBytes(metric.(interface{ GetKey() []byte }).GetKey()); v != nil {
		return ErrAlreadyExists
	}

	return registry.set(metric)
}

func (registry *Registry) list() (result Metrics) {
	for _, metricKey := range registry.storage.Keys() {
		metric, _ := registry.storage.GetByBytes(metricKey.([]byte))
		if metric == nil {
			continue
		}
		result = append(result, metric.(Metric))
	}
	return
}

func (registry *Registry) listSorted() (result Metrics) {
	list := registry.list()
	list.Sort()
	return list
}

func (registry *Registry) List() (result Metrics) {
	return registry.listSorted()
	//return registry.list()
}

func (registry *Registry) remove(metric Metric) {
	metric.Stop()
	registry.storage.Unset(metric.(interface{ GetKey() []byte }).GetKey())
}

func (registry *Registry) GetSender() Sender {
	return registry.metricSender
}

func (registry *Registry) SetDefaultSender(newMetricSender Sender) {
	registry.stopMonitor()
	registry.metricSender = newMetricSender
	registry.startMonitor()
}
func (registry *Registry) GetDefaultSender() Sender {
	return registry.metricSender
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

func (registry *Registry) getMonitorState() uint64 {
	return atomic.LoadUint64(&registry.monitorState)
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

func (registry *Registry) startMonitor() {
	return
	/*
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
	*/
}

func (registry *Registry) stopMonitor() {
	return

	//m.setMonitorState(monitorState_Stopping)
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
		tags = NewFastTags()
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

func List() []Metric {
	return registry.List()
}

func Get(metricType Type, key string, tags AnyTags) Metric {
	return registry.Get(metricType, key, tags)
}

// copied from https://github.com/demdxx/sort-algorithms/blob/master/algorithms.go
func BubbleSort(data stringSlice) {
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
		for idx := range *inTags {
			tag := (*inTags)[idx]
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

	if len(key) > 0 {
		reusables.buf.WriteString(`,`)
	}
	if tags == nil {
		tags = nilTags
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

func SetDefaultSender(newMetricSender Sender) {
	registry.SetDefaultSender(newMetricSender)
}

func GetDefaultSender() Sender {
	return registry.GetDefaultSender()
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
		registry.storage.Unset(metricKey)
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
