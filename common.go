package metrics

import (
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Sender is a sender to be used to periodically send metric values (for example to StatsD)
// On high loaded systems we recommend to use prometheus and a status page with all exported metrics instead of sending
// metrics to somewhere.
type Sender interface {
	// SendInt64 is used to send signed integer values
	SendInt64(metric Metric, key string, value int64) error

	// SendUint64 is used to send unsigned integer values
	SendUint64(metric Metric, key string, value uint64) error

	// SendFloat64 is used to send float values
	SendFloat64(metric Metric, key string, value float64) error
}

// common is an implementation of base routines of a metric, it's inherited by other implementations
type common struct {
	registryItem // any metric could be saved into the registry, so include "registryItem"

	sync.Mutex
	running uint64

	// interval in the interval to be used to:
	//  * recheck if the metric value was changed.
	//  * send the value through the "sender" (see description of "Sender").
	//
	// If the value wasn't changed not then the value of uselessCounter is increased.
	// If the value of uselessCounter reaches a threshold (see gcUselessLimit) then
	// the method "Stop" will be called and the metric will be removed from the registry by registry's GC.
	interval time.Duration

	isGCEnabled    uint64
	uselessCounter uint64

	// parent is a pointer to the object of the final implementation of a metric (for example *GaugeFloat64)
	parent Metric

	// getWasUseless is a function to check if the metric have changed since the last call.
	// It depends on specific final implementation so it's passed-throughed from the parent
	// It's used to determine if the metric could be removed by GC (if haven't changed then it's useless then
	// it could be removed; see description of "interval").
	getWasUseless func() bool
}

func (m *common) init(parent Metric, key string, tags AnyTags, getWasUseless func() bool) {
	m.parent = parent
	m.SetGCEnabled(GetDefaultGCEnabled())

	registry.Register(parent, key, tags)

	m.getWasUseless = getWasUseless
	m.registryItem.init(parent, key)

	if GetDefaultIsRunned() {
		parent.Run(GetDefaultIterateInterval())
	}
}

// IsRunning returns if the metric is run()'ed and not Stop()'ed.
func (m *common) IsRunning() bool {
	if m == nil {
		return false
	}
	return atomic.LoadUint64(&m.running) > 0
}

// GetInterval return the iteration interval (between sending or GC checks)
func (m *common) GetInterval() time.Duration {
	return m.interval
}

// SetGCEnabled sets if this metric could be stopped and removed from the metrics registry if the value do not change for a long time
func (m *common) SetGCEnabled(enabled bool) {
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
func (m *common) IsGCEnabled() bool {
	if m == nil {
		return false
	}
	return atomic.LoadUint64(&m.isGCEnabled) > 0
}

func (m *common) uselessCounterIncrement() {
	if atomic.AddUint64(&m.uselessCounter, 1) <= gcUselessLimit {
		return
	}
	if !m.IsRunning() {
		return
	}
	go m.parent.Stop()
}

func (m *common) uselessCounterReset() {
	atomic.StoreUint64(&m.uselessCounter, 0)
}

func (m *common) doIterateGC() {
	if !m.IsGCEnabled() {
		return
	}

	if m.getWasUseless == nil {
		return
	}

	if m.getWasUseless() {
		m.uselessCounterIncrement()
		return
	}
	m.uselessCounterReset()
}

func (m *common) doIterateSender() {
	sender := registry.GetSender()
	if sender == nil {
		return
	}

	m.Send(sender)
}

// Iterate runs routines supposed to be runned once per selected interval.
// This routines are sending the metric value via sender (see `SetSender`) and GC (to remove the metric if it is not
// used for a long time).
func (m *common) Iterate() {
	defer recoverPanic()
	m.doIterateGC()
	m.doIterateSender()
}

func (m *common) run(interval time.Duration) {
	if m.IsRunning() {
		return
	}
	m.interval = interval
	iterationHandlers.Add(m)
	atomic.StoreUint64(&m.uselessCounter, 0)
	atomic.StoreUint64(&m.running, 1)
	return
}

// Run starts the metric. We did not check if it is safe to call this method from external code. Not recommended to use, yet.
// Metrics starts automatically after it's creation, so there's no need to call this method, usually.
func (m *common) Run(interval time.Duration) {
	if m == nil {
		return
	}
	m.Lock()
	m.run(interval)
	m.Unlock()
}

func (m *common) stop() {
	if !m.IsRunning() {
		return
	}
	iterationHandlers.Remove(m)
	m.interval = time.Duration(0)
	atomic.StoreUint64(&m.running, 0)
}

// Stop ends any activity on this metric, except Garbage collector that will remove this metric from the metrics registry.
func (m *common) Stop() {
	if m == nil {
		return
	}
	m.Lock()
	m.stop()
	m.Unlock()
}

// MarshalJSON returns JSON representation of a metric for external monitoring systems
func (m *common) MarshalJSON() ([]byte, error) {
	nameJSON, _ := json.Marshal(m.name)
	descriptionJSON, _ := json.Marshal(m.description)
	tagsJSON, _ := json.Marshal(m.tags.String())
	typeJSON, _ := json.Marshal(m.GetType().String())
	value := m.GetFloat64()

	metricJSON := fmt.Sprintf(`{"name":%s,"tags":%s,"value":%v,"description":%s,"type":%s}`,
		string(nameJSON),
		string(tagsJSON),
		value,
		string(descriptionJSON),
		string(typeJSON),
	)
	return []byte(metricJSON), nil
}

// GetCommons returns the *common of a metric (it supposed to be used for internal routines only).
// The "*common" is a structure that is common through all types of metrics (with GC info, registry info and so on).
func (m *common) GetCommons() *common {
	return m
}

// Placeholders
// TODO: remove this hacks :(

// Send initiates a sending of the metric value through the sender (see "SetSender")
func (m *common) Send(sender Sender) {
	m.parent.Send(sender)
}

// GetType returns type of the metric
func (m *common) GetType() Type {
	return m.parent.GetType()
}

// GetFloat64 returns current value of the metric
func (m *common) GetFloat64() float64 {
	return m.parent.GetFloat64()
}

// EqualsTo checks if it's the same metric passed as the argument
func (m *common) EqualsTo(cmpI iterator) bool {
	cmp, ok := cmpI.(interface{ GetCommons() *common })
	if !ok {
		return false
	}
	return m == cmp.GetCommons()
}
