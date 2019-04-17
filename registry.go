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

func (m *Registry) SetDisabled(newIsDisabled bool) bool {
	newValue := uint64(0)
	if newIsDisabled {
		newValue = 1
	}

	return atomic.SwapUint64(&m.isDisabled, newValue) != 0
}

func SetDisabled(newIsDisabled bool) bool {
	return registry.SetDisabled(newIsDisabled)
}

func (m *Registry) IsDisabled() bool {
	return atomic.LoadUint64(&m.isDisabled) != 0
}

func IsDisabled() bool {
	return registry.IsDisabled()
}

func (m *Registry) GetDefaultIterateInterval() time.Duration {
	if m.metricsIterateIntervaler == nil {
		return defaultIterateInterval
	}
	return m.metricsIterateIntervaler.MetricsIterateInterval()
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

func (m *Registry) get(metricType Type, key string, tags AnyTags) Metric {
	considerHiddenTags(tags)
	buf := generateStorageKey(metricType, key, tags)
	rI, _ := m.storage.GetByBytes(buf.buf.Bytes())
	buf.Release()
	if rI == nil {
		return nil
	}
	r := rI.(Metric)
	if !r.IsRunning() {
		/* Works unstable, commented for a while:
		if !r.IsRunning() {
			r.Run(metricsRegistry.GetDefaultIterateInterval())
		}
		m.set(r) // may be GC already cleanup this metric, so re-set it
		*/
	}
	return r
}

func (m *Registry) Get(metricType Type, key string, tags AnyTags) Metric {
	if m.IsDisabled() {
		return nil
	}
	return m.get(metricType, key, tags)
}

func (m *Registry) set(metric Metric) error {
	m.storage.Set(metric.(interface{ GetKey() []byte }).GetKey(), metric)
	return nil
}

func (m *Registry) Set(metric Metric) error {
	if v, _ := m.storage.GetByBytes(metric.(interface{ GetKey() []byte }).GetKey()); v != nil {
		return ErrAlreadyExists
	}

	return m.set(metric)
}

func (m *Registry) list() (result Metrics) {
	for _, metricKey := range m.storage.Keys() {
		metric, _ := m.storage.GetByUint64(metricKey.(uint64))
		if metric == nil {
			continue
		}
		result = append(result, metric.(Metric))
	}
	return
}

func (m *Registry) listSorted() (result Metrics) {
	list := m.list()
	list.Sort()
	return list
}

func (m *Registry) List() (result Metrics) {
	return m.listSorted()
	//return m.list()
}

func (m *Registry) remove(metric Metric) {
	metric.Stop()
	m.storage.Unset(metric.(interface{ GetKey() []byte }).GetKey())
}

func (m *Registry) GetSender() Sender {
	return m.metricSender
}

func (m *Registry) SetDefaultSender(newMetricSender Sender) {
	m.stopMonitor()
	m.metricSender = newMetricSender
	m.startMonitor()
}
func (m *Registry) GetDefaultSender() Sender {
	return m.metricSender
}

func (m *Registry) SetDefaultGCEnabled(newGCEnabledValue bool) {
	if newGCEnabledValue {
		atomic.StoreUint32(&m.defaultGCEnabled, 1)
	} else {
		atomic.StoreUint32(&m.defaultGCEnabled, 0)
	}
}

func SetDefaultGCEnabled(newValue bool) {
	registry.SetDefaultGCEnabled(newValue)
}

func (m *Registry) GetDefaultGCEnabled() bool {
	return atomic.LoadUint32(&m.defaultGCEnabled) != 0
}

func GetDefaultGCEnabled() bool {
	return registry.GetDefaultGCEnabled()
}

func (m *Registry) SetDefaultIsRunned(newIsRunnedValue bool) {
	if newIsRunnedValue {
		atomic.StoreUint32(&m.defaultIsRunned, 1)
	} else {
		atomic.StoreUint32(&m.defaultIsRunned, 0)
	}
}

func SetDefaultIsRunned(newValue bool) {
	registry.SetDefaultIsRunned(newValue)
}

func (m *Registry) GetDefaultIsRunned() bool {
	return atomic.LoadUint32(&m.defaultIsRunned) != 0
}

func GetDefaultIsRunned() bool {
	return registry.GetDefaultIsRunned()
}

func (m *Registry) getMonitorState() uint64 {
	return atomic.LoadUint64(&m.monitorState)
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

	panic("Shouldn't happend, ever")
}*/

func (m *Registry) startMonitor() {
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

func (m *Registry) stopMonitor() {
	return

	//m.setMonitorState(monitorState_Stopping)
}

func (m *Registry) GC() {
	for _, metricKey := range m.storage.Keys() {
		metricI, _ := m.storage.Get(metricKey)
		if metricI == nil {
			continue
		}
		metric := metricI.(Metric)
		if !metric.IsRunning() {
			m.remove(metric)
			metric.Release()
		}
	}
}

func GC() {
	registry.GC()
}

func (metricsRegistry *Registry) Register(metric Metric, key string, inTags AnyTags) error {
	tags := NewFastTags()
	for _, tag := range defaultTags {
		tags.Set(tag.Key, tag.StringValue)
	}
	if inTags != nil {
		inTags.Each(func(k string, v interface{}) bool {
			tags.Set(k, v)
			return true
		})
	}

	commons := metric.(interface{ GetCommons() *metricCommon }).GetCommons()
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

func generateStorageKey(metricType Type, key string, tags AnyTags) *keyGeneratorReusables {
	reusables := newKeyGeneratorReusables()
	reusables.buf.WriteString(key)

	if tags != nil {
		reusables.buf.WriteString(`,`)
		tags.WriteAsString(&reusables.buf)
	}

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

func (m *Registry) getHiddenTags() hiddenTagsInternal {
	result := atomic.LoadPointer((*unsafe.Pointer)((unsafe.Pointer)(&m.hiddenTags)))
	if result == nil {
		return nil
	}
	return *(*hiddenTagsInternal)(result)
}

func (m *Registry) SetHiddenTags(newRawHiddenTags HiddenTags) {
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
	atomic.StorePointer((*unsafe.Pointer)((unsafe.Pointer)(&m.hiddenTags)), (unsafe.Pointer)(&newHiddenTags))
}

func (m *Registry) IsHiddenTag(tagKey string, tagValue interface{}) bool {
	hiddenTags := m.getHiddenTags()
	return hiddenTags.IsHiddenTag(tagKey, tagValue)
}

func (m *Registry) Reset() {
	for _, metricKey := range m.storage.Keys() {
		metric, _ := m.storage.Get(metricKey)
		if metric == nil {
			continue
		}
		metric.(Metric).Stop()
		m.storage.Unset(metricKey)
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
