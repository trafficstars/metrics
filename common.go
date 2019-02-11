package metrics

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	metricsCount uint64
)

type Sender interface {
	SendInt64(metric Metric, key string, value int64) error
	SendUint64(metric Metric, key string, value uint64) error
	SendFloat64(metric Metric, key string, value float64) error
}

type metricCommon struct {
	metricRegistryItem

	sync.Mutex
	interval       time.Duration
	running        uint64
	isGCEnabled    uint64
	uselessCounter uint64

	sender      Sender
	isSenderSet uint64

	parent        Metric
	getWasUseless func() bool
}

func (metric *metricCommon) init(parent Metric, key string, tags AnyTags, getWasUseless func() bool) {
	metric.parent = parent
	metric.SetSender(GetDefaultSender())
	metric.SetGCEnabled(GetDefaultGCEnabled())

	metricsRegistry.Register(parent, key, tags)
	if GetDefaultIsRunned() {
		parent.Run(GetDefaultIterateInterval())
	}

	metric.getWasUseless = getWasUseless
	metric.metricRegistryItem.init(parent)
}

func (m *metricCommon) getIsSenderSet() bool {
	return atomic.LoadUint64(&m.isSenderSet) != 0
}

// SetSender sets the sender to be used to periodically send metric values (for example to StatsD)
// On high loaded systems we recommend to use prometheus and a status page with all exported metrics instead of sending metrics to somewhere.
func (m *metricCommon) SetSender(sender Sender) {
	if m == nil {
		return
	}
	m.Lock()
	m.sender = sender
	if m.sender == nil {
		atomic.StoreUint64(&m.isSenderSet, 0)
	} else {
		atomic.StoreUint64(&m.isSenderSet, 1)
	}
	m.Unlock()
}

// GetSender returns the sender (see SetSender)
func (m *metricCommon) GetSender() Sender {
	if m == nil {
		return nil
	}
	if !m.getIsSenderSet() {
		return nil
	}

	m.Lock()
	sender := m.sender
	m.Unlock()
	return sender
}

// IsRunning returnes if the metric is run()'ed and not Stop()'ed.
func (m *metricCommon) IsRunning() bool {
	if m == nil {
		return false
	}
	return atomic.LoadUint64(&m.running) > 0
}

// GetInterval return the iteration interval (between sending or GC checks)
func (m *metricCommon) GetInterval() time.Duration {
	return m.interval
}

// SetGCEnabled sets if this metric could be stopped and removed from the metrics registry if the value do not change for a long time
func (m *metricCommon) SetGCEnabled(enabled bool) {
	if m == nil {
		return
	}
	if enabled {
		atomic.StoreUint64(&m.isGCEnabled, 1)
	} else {
		atomic.StoreUint64(&m.isGCEnabled, 0)
	}
}

// IsGCEnabled returns if the GC enabled for this metric (see method `SetGCEnabled`)
func (m *metricCommon) IsGCEnabled() bool {
	if m == nil {
		return false
	}
	return atomic.LoadUint64(&m.isGCEnabled) > 0
}

func (m *metricCommon) uselessCounterIncrement() {
	if atomic.AddUint64(&m.uselessCounter, 1) <= gcUselessLimit {
		return
	}
	if !m.IsRunning() {
		return
	}
	go m.parent.Stop()
}

func (m *metricCommon) uselessCounterReset() {
	atomic.StoreUint64(&m.uselessCounter, 0)
}

func (m *metricCommon) doIterateGC() {
	if !m.IsGCEnabled() {
		return
	}

	if m.getWasUseless() {
		m.uselessCounterIncrement()
		return
	}
	m.uselessCounterReset()
}

func (m *metricCommon) doIterateSender() {
	sender := m.GetSender()
	if sender == nil {
		return
	}

	m.Send(sender)
}

// Iterate runs routines supposed to be runned once per selected interval.
// This routines are sending the metric value via sender (see `SetSender`) and GC (to remove the metric if it is not used for a long time).
func (m *metricCommon) Iterate() {
	m.doIterateGC()
	m.doIterateSender()
}

func (m *metricCommon) run(interval time.Duration) {
	if m.IsRunning() {
		return
	}
	iterationHandlers.Add(m)
	m.interval = interval
	atomic.StoreUint64(&m.uselessCounter, 0)
	atomic.StoreUint64(&m.running, 1)
	return
}

// Run starts the metric. We did not check if it is safe to call this method from external code. Not recommended to use, yet.
// Metrics starts automatically after it's creation, so there's no need to call this method, usually.
func (m *metricCommon) Run(interval time.Duration) {
	if m == nil {
		return
	}
	m.Lock()
	m.run(interval)
	m.Unlock()
}

func (m *metricCommon) stop() {
	if !m.IsRunning() {
		return
	}
	iterationHandlers.Remove(m)
	m.interval = time.Duration(0)
	atomic.StoreUint64(&m.running, 0)
}

// Stop ends any activity on this metric, except Garbage collector that will remove this metric from the metrics registry.
func (m *metricCommon) Stop() {
	if m == nil {
		return
	}
	m.Lock()
	m.stop()
	m.Unlock()
}

func (metric *metricCommon) MarshalJSON() ([]byte, error) {
	nameJSON, _ := json.Marshal(metric.name)
	descriptionJSON, _ := json.Marshal(metric.description)
	tagsJSON, _ := json.Marshal(string(metric.storageKey[:strings.IndexByte(string(metric.storageKey), '@')]))
	typeJSON, _ := json.Marshal(string(metric.GetType()))
	value := metric.GetFloat64()

	metricJSON := fmt.Sprintf(`{"name":%s,"tags":%s,"value":%v,"description":%s,"type":%s}`,
		string(nameJSON),
		string(tagsJSON),
		value,
		string(descriptionJSON),
		string(typeJSON),
	)
	return []byte(metricJSON), nil
}

// Placeholders
// TODO: remove this hacks :(

func (m *metricCommon) Send(sender Sender) {
	m.parent.Send(sender)
}

func (m *metricCommon) GetType() Type {
	return m.parent.GetType()
}

func (m *metricCommon) GetFloat64() float64 {
	return m.parent.GetFloat64()
}
