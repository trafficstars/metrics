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
	defaultIterateInterval = time.Minute
	gcUselessLimit         = 5
	maxPercentileValues    = 5
)

const (
	monitorState_Stopped  = 0
	monitorState_Started  = 1
	monitorState_Stopping = 2
)

var defaultFlowPercentiles = []float64{0.01, 0.1, 0.5, 0.9, 0.99}

var (
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
	defaultPercentiles       []float64
}

func SetLimit(newLimit uint) {
	// for future compatibility
}

func (r *Registry) SetDisabled(newIsDisabled bool) bool {
	newValue := uint32(0)
	if newIsDisabled {
		newValue = 1
	}

	return atomic.SwapUint32(&r.isDisabled, newValue) != 0
}

func SetDisabled(newIsDisabled bool) bool {
	return registry.SetDisabled(newIsDisabled)
}

func (r *Registry) IsDisabled() bool {
	return atomic.LoadUint32(&r.isDisabled) != 0
}

func IsDisabled() bool {
	return registry.IsDisabled()
}

func (r *Registry) GetDefaultIterateInterval() time.Duration {
	if r.metricsIterateIntervaler == nil {
		return defaultIterateInterval
	}
	return r.metricsIterateIntervaler.MetricsIterateInterval()
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
	if !MemoryReuseEnabled() {
		return
	}
	b.buf.Reset()
	keyGeneratorReusablesPool.Put(b)
}

func New() *Registry {
	r := &Registry{
		storage:            atomicmap.New(),
		defaultPercentiles: defaultFlowPercentiles,
	}
	r.SetDefaultGCEnabled(true)
	r.SetDefaultIsRan(true)
	r.SetSender(nil)
	return r
}

var registry = New()

func (r *Registry) get(metricType Type, key string, tags AnyTags) Metric {
	considerHiddenTags(tags)
	buf := generateStorageKey(metricType, key, tags)
	mI, _ := r.storage.GetByBytes(buf.buf.Bytes())
	buf.Release()
	if mI == nil {
		return nil
	}
	m := mI.(Metric)
	if m.IsRunning() {
		return m
	}

	m.lock()
	defer m.unlock()
	if !m.IsRunning() {
		m.run(r.GetDefaultIterateInterval())
		_ = r.set(m) // may be GC already cleanup this metric, so re-set it
	}
	return m
}

func (r *Registry) Get(metricType Type, key string, tags AnyTags) Metric {
	if r.IsDisabled() {
		return nil
	}
	return r.get(metricType, key, tags)
}

func (r *Registry) set(metric Metric) error {
	_ = r.storage.Set(metric.GetKey(), metric)
	return nil
}

func (r *Registry) Set(metric Metric) error {
	if v, _ := r.storage.GetByBytes(metric.GetKey()); v != nil {
		return ErrAlreadyExists
	}

	return r.set(metric)
}

func (r *Registry) list() *Metrics {
	result := newMetrics()
	for _, metricKey := range r.storage.Keys() {
		metricI, _ := r.storage.GetByBytes(metricKey.([]byte))
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

func (r *Registry) listSorted() (result *Metrics) {
	list := r.list()
	list.Sort()
	return list
}

func (r *Registry) List() (result *Metrics) {
	return r.list()
}

func (r *Registry) remove(metric Metric) {
	metric.stop()
	err := r.storage.Unset(metric.GetKey())
	if err != nil {
		panic(err)
	}
}

func (r *Registry) GetSender() Sender {
	return *(*Sender)(atomic.LoadPointer((*unsafe.Pointer)((unsafe.Pointer)(&r.metricSender))))
}

func (r *Registry) SetSender(newMetricSender Sender) {
	atomic.StorePointer((*unsafe.Pointer)((unsafe.Pointer)(&r.metricSender)), (unsafe.Pointer)(&newMetricSender))
}

func (r *Registry) SetDefaultGCEnabled(newGCEnabledValue bool) {
	if newGCEnabledValue {
		atomic.StoreUint32(&r.defaultGCEnabled, 1)
	} else {
		atomic.StoreUint32(&r.defaultGCEnabled, 0)
	}
}

func SetDefaultGCEnabled(newValue bool) {
	registry.SetDefaultGCEnabled(newValue)
}

func (r *Registry) GetDefaultGCEnabled() bool {
	return atomic.LoadUint32(&r.defaultGCEnabled) != 0
}

func GetDefaultGCEnabled() bool {
	return registry.GetDefaultGCEnabled()
}

func (r *Registry) SetDefaultIsRan(newIsRanValue bool) {
	if newIsRanValue {
		atomic.StoreUint32(&r.defaultIsRunned, 1)
	} else {
		atomic.StoreUint32(&r.defaultIsRunned, 0)
	}
}

func SetDefaultIsRan(newValue bool) {
	registry.SetDefaultIsRan(newValue)
}

func (r *Registry) GetDefaultIsRan() bool {
	return atomic.LoadUint32(&r.defaultIsRunned) != 0
}

func GetDefaultIsRunned() bool {
	return registry.GetDefaultIsRan()
}

func (r *Registry) getMonitorState() uint32 {
	return atomic.LoadUint32(&r.monitorState)
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

func (r *Registry) setMonitorState(newState uint32) uint32 {
	return atomic.SwapUint32(&r.monitorState, newState)
}

func (r *Registry) startMonitor() {
	if v := r.setMonitorState(monitorState_Started); v != monitorState_Stopped {
		r.setMonitorState(v)
		return
	}

	var mem runtime.MemStats
	pausesLatency := int64(-1)

	go func() {
		interval := r.GetDefaultIterateInterval()
		for {
			monitorState := r.getMonitorState()
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

		r.setMonitorState(monitorState_Stopped)
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

func (r *Registry) stopMonitor() {
	r.setMonitorState(monitorState_Stopping)
}

func (r *Registry) GC() {
	for _, metricKey := range r.storage.Keys() {
		metricI, _ := r.storage.Get(metricKey)
		if metricI == nil {
			continue
		}
		metric := metricI.(Metric)
		if metric.IsRunning() || !metric.IsGCEnabled() {
			continue
		}

		metric.lock()
		if !metric.IsRunning() {
			r.remove(metric)
			metric.Release()
		}
		metric.unlock()
	}
}

func GC() {
	registry.GC()
}

func (r *Registry) Register(metric Metric, key string, inTags AnyTags) error {
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

	return r.Set(metric)
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

func (r *Registry) getHiddenTags() hiddenTagsInternal {
	result := atomic.LoadPointer((*unsafe.Pointer)((unsafe.Pointer)(&r.hiddenTags)))
	if result == nil {
		return nil
	}
	return *(*hiddenTagsInternal)(result)
}

func (r *Registry) SetHiddenTags(newRawHiddenTags HiddenTags) {
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
	atomic.StorePointer((*unsafe.Pointer)((unsafe.Pointer)(&r.hiddenTags)), (unsafe.Pointer)(&newHiddenTags))
}

func (r *Registry) IsHiddenTag(tagKey string, tagValue interface{}) bool {
	hiddenTags := r.getHiddenTags()
	return hiddenTags.IsHiddenTag(tagKey, tagValue)
}

func (r *Registry) Reset() {
	for _, metricKey := range r.storage.Keys() {
		metricI, _ := r.storage.Get(metricKey)
		if metricI == nil {
			continue
		}
		metric := metricI.(Metric)
		metric.lock()
		metric.stop()
		_ = r.storage.Unset(metricKey)
		metric.unlock()
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

func SetDefaultPercentiles(p []float64) {
	registry.SetDefaultPercentiles(p)
}

func (r *Registry) SetDefaultPercentiles(p []float64) {
	sort.Float64s(p)
	r.defaultPercentiles = p
}
