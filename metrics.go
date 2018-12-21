package metrics

import (
	"bytes"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"

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
	storage               atomicmap.Map
	metricSender          metricworker.MetricSender
	metricsSendIntervaler MetricsSendIntervaler
	monitorState          uint64
}

func (m *Metrics) GetSendInterval() time.Duration {
	if m.metricsSendIntervaler == nil {
		return defaultSendInterval
	}
	return m.metricsSendIntervaler.MetricsSendInterval()
}

func (m *Metrics) Get(metricType MetricType, key string, tags AnyTags) *Metric {
	storageKeyBuf := generateStorageKey(metricType, key, tags)
	rI, _ := m.storage.Get(storageKeyBuf.result.String())
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

func (m *Metrics) set(metric *Metric) error {
	m.storage.Set(string(metric.storageKey), metric)
	return nil
}

func (m *Metrics) Set(metric *Metric) error {
	if v, _ := m.storage.Get(string(metric.storageKey)); v != nil {
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
	m.storage.Unset(metric.storageKey)
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

type preallocatedBuffer struct {
	sync.Mutex
	tagKeys []string
	result  bytes.Buffer
}

var (
	bufs       [maxConcurrency]*preallocatedBuffer
	bufPointer uint64
)

func init() {
	for i := 0; i < maxConcurrency; i++ {
		bufs[i] = &preallocatedBuffer{}
	}
}

func generateStorageKey(metricType MetricType, key string, tags AnyTags) *preallocatedBuffer {
	// It's required to avoid memory allocations. So if we allocated a buffer once, we reuse it.
	// We have buffers (of amount maxConcurrency) to be able to process this function concurrently.

	curBufPointer := atomic.AddUint64(&bufPointer, 1)
	curBufPointer %= maxConcurrency
	buf := bufs[curBufPointer]
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
		sort.Strings(buf.tagKeys)

		for _, k := range buf.tagKeys {
			buf.result.WriteString(`,`)
			buf.result.WriteString(k)
			buf.result.WriteString(`=`)
			buf.result.WriteString(TagValueToString(inTags[k]))
		}
	case FastTags:
		for _, tag := range inTags {
			if defaultTags.IsSet(tag.Key) {
				continue
			}
			buf.result.WriteString(`,`)
			buf.result.WriteString(tag.Key)
			buf.result.WriteString(`=`)
			buf.result.Write(tag.Value)
		}
	default:
		panic("not implemented")
	}

	if len(metricType) > 0 {
		buf.result.WriteString("@")
		buf.result.WriteString(string(metricType))
	}

	return buf
}

func init() {
	metrics.storage = atomicmap.New()
}

func Init(newMetricSender metricworker.MetricSender, newMetricsSendIntervaler MetricsSendIntervaler, newDefaultAnyTags AnyTags) {
	metrics.SetSender(newMetricSender)
	metrics.metricsSendIntervaler = newMetricsSendIntervaler
	defaultTags = newDefaultAnyTags.ToFastTags()
}
