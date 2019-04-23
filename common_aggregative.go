package metrics

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

var (
	slicerInterval = time.Second
)

// ! Before read this file please read README.md !

type AggregativeStatistics interface {
	// GetPercentile returns the value for a given percentile (0.0 .. 1.0).
	// It returns nil if the percentile could not be calculated (it could be in case of using "flow" [instead of
	// "buffered"] aggregative metrics)
	//
	// If you need to calculate multiple percentiles then use GetPercentiles() to get better performance
	GetPercentile(percentile float64) *float64

	// GetPercentiles returns values for given percentiles (0.0 .. 1.0).
	// A value is nil if the percentile could not be calculated.
	GetPercentiles(percentile []float64) []*float64

	// Set forces all the values in the statistics to be equal to the passed values
	Set(staticValue float64)

	// ConsiderValue is analog of "Observe" of https://godoc.org/github.com/prometheus/client_golang/prometheus#Observer
	// It's used to merge the value to the statistics. For example if there were considered only values 1, 2 and 3 then
	// the average value will be 2.
	ConsiderValue(value float64)

	// Release is used for memory reuse (it's called when it's known that the statistics won't be used anymore)
	// This method is not supposed to be called from external code, it designed for internal uses only.
	Release()

	//
	MergeStatistics(AggregativeStatistics)
}

// SetSlicerInterval affects only new metrics (it doesn't affect already created one). You may use function `Reset()`
// to "update" configuration of all metrics.
func SetSlicerInterval(newSlicerInterval time.Duration) {
	slicerInterval = newSlicerInterval
}

// SetAggregationPeriods affects only new metrics (it doesn't affect already created on). You may use function
// `Reset()` to "update" configuration of all metrics.
//
// Every higher aggregation period should be a multiple of the lower one.
func SetAggregationPeriods(newAggregationPeriods []AggregationPeriod) {
	aggregationPeriods.Lock()
	aggregationPeriods.s = newAggregationPeriods
	aggregationPeriods.Unlock()
}

// AggregationPeriod is used to define aggregation periods (see "Slicing" in "README.md")
type AggregationPeriod struct {
	Interval uint64 // in slicerInterval-s
}

// String returns a string representation of the aggregation period
//
// It will return in a short format (like "5s", "1h") if the amount of seconds could be represented as exact value of
// days, hours or minutes, or if the amount of seconds is less than 60. Otherwise the format will be like `1h5m0s`.
func (period *AggregationPeriod) String() string {
	interval := time.Duration(period.Interval) * slicerInterval
	seconds := uint64(interval / time.Second)
	if seconds < 60 {
		return strconv.FormatUint(seconds, 10) + `s`
	}
	if seconds%(3600*24) == 0 {
		return strconv.FormatUint(seconds/(3600*24), 10) + `d`
	}
	if seconds%3600 == 0 {
		return strconv.FormatUint(seconds/3600, 10) + `h`
	}
	if seconds%60 == 0 {
		return strconv.FormatUint(seconds/60, 10) + `m`
	}
	return interval.String()
}

type aggregationPeriodsT struct {
	sync.RWMutex

	s []AggregationPeriod
}

var (
	// The default aggregation periods: 1s, 5s, 1m, 5m, 1h, 6h, 1d
	aggregationPeriods = aggregationPeriodsT{
		s: []AggregationPeriod{
			{5},
			{60},
			{300},
			{3600},
			{21600},
			{86400},
		},
	}
)

// GetBaseAggregationPeriod returns AggregationPeriod equals to the slicer's interval (see "Slicing" in README.md)
func GetBaseAggregationPeriod() *AggregationPeriod {
	return &AggregationPeriod{1}
}

// GetAggregationPeriods returns aggregations periods (see "Slicing" in README.md)
func GetAggregationPeriods() (r []AggregationPeriod) {
	aggregationPeriods.RLock()
	r = make([]AggregationPeriod, len(aggregationPeriods.s))
	copy(r, aggregationPeriods.s)
	aggregationPeriods.RUnlock()
	return
}

// AggregativeValue is a struct that contains all the values related to an aggregation period.
type AggregativeValue struct {
	sync.Mutex

	Count uint64
	Min   AtomicFloat64
	Avg   AtomicFloat64
	Max   AtomicFloat64

	AggregativeStatistics
}

// newAggregativeValue returns an empty AggregativeValue (as a memory-reuse-away constructor).
func newAggregativeValue() *AggregativeValue {
	r := aggregativeValuePool.Get().(*AggregativeValue)
	return r
}

// set makes the value look like if there were only one event with the value passed as the argument
func (aggrV *AggregativeValue) set(v float64) {
	if aggrV == nil {
		return
	}
	atomic.StoreUint64(&aggrV.Count, 1)
	aggrV.Min.Set(v)
	aggrV.Avg.Set(v)
	aggrV.Max.Set(v)
	if aggrV.AggregativeStatistics != nil {
		aggrV.AggregativeStatistics.Set(v)
	}
}

// LockDo is just a wrapper around Lock()/Unlock(). It's quite handy to understand who caused a deadlock in
// stack traces.
func (aggrV *AggregativeValue) LockDo(fn func(*AggregativeValue)) {
	if aggrV == nil {
		return
	}
	aggrV.Lock()
	fn(aggrV)
	aggrV.Unlock()
}

// GetAvg just returns the average value
func (aggrV *AggregativeValue) GetAvg() float64 {
	if aggrV == nil {
		return 0
	}
	return aggrV.Avg.Get()
}

// AggregativeValues is a full collection of "AggregativeValue"-s (see "Slicing" in README.md)
type AggregativeValues struct {
	Last     *AggregativeValue
	Current  *AggregativeValue
	ByPeriod []*AggregativeValue
	Total    *AggregativeValue
}

// slicer returns an object that will call method DoSlice() of commonAggregative if method Iterate() was called.
//
// It's used to deduplicate code and reuse Iterators (see "Iterators" in README.md)
type commonAggregativeSlicer struct {
	metric   *commonAggregative
	interval time.Duration
}

func (slicer *commonAggregativeSlicer) Iterate() {
	slicer.metric.DoSlice()
}
func (slicer *commonAggregativeSlicer) GetInterval() time.Duration {
	return slicer.interval
}
func (slicer *commonAggregativeSlicer) IsRunning() bool {
	return slicer.metric.IsRunning()
}
func (slicer *commonAggregativeSlicer) EqualsTo(cmpI iterator) bool {
	cmp, ok := cmpI.(*commonAggregativeSlicer)
	if !ok {
		return false
	}
	return slicer == cmp
}

// commonAggregative is an implementation of common routines through all aggregative metrics
type commonAggregative struct {
	common

	aggregationPeriods []AggregationPeriod
	data               AggregativeValues
	currentSliceData   *AggregativeValue
	tick               uint64
	slicer             iterator

	histories histories
}

// newAggregativeStatistics returns an AggregativeStatistics (as a memory-reuse-aware constructor)
func (m *commonAggregative) newAggregativeStatistics() AggregativeStatistics {
	return m.parent.(interface{ NewAggregativeStatistics() AggregativeStatistics }).NewAggregativeStatistics()
}

func (m *commonAggregative) NewAggregativeValue() *AggregativeValue {
	v := newAggregativeValue()
	v.AggregativeStatistics = m.newAggregativeStatistics()
	return v
}

func (m *commonAggregative) init(parent Metric, key string, tags AnyTags) {
	m.parent = parent

	// See "Slicing" in README.md

	m.slicer = &commonAggregativeSlicer{
		metric:   m,
		interval: slicerInterval,
	}
	m.aggregationPeriods = GetAggregationPeriods()
	m.data.Last = m.NewAggregativeValue()
	m.data.Current = m.NewAggregativeValue()
	m.data.Total = m.NewAggregativeValue()

	m.histories.ByPeriod = make([]*history, 0, len(m.aggregationPeriods))
	previousPeriod := AggregationPeriod{1}
	for _, period := range m.aggregationPeriods {
		hist := &history{}
		if period.Interval%previousPeriod.Interval != 0 {
			// We support only an aggregation period that divides to the previous aggregation period
			// For example we support: 1s, 5s, 1m; but we doesn't support: 1s, 5s, 13s.
			//
			// It's caused by our algorithm of calculating statistics of higher aggregation periods using
			// history of statistics of lower aggregation periods. So a higher aggregation period statistics
			// is calculated from multiple lower aggregation period statistics

			// TODO: print error
			//panic(fmt.Errorf("period.Interval (%v) %% previousPeriod.Interval (%v) != 0 (%v)", period.Interval, previousPeriod.Interval, period.Interval%previousPeriod.Interval))
		}
		hist.storage = make([]*AggregativeValue, period.Interval/previousPeriod.Interval)

		m.histories.ByPeriod = append(m.histories.ByPeriod, hist)
		previousPeriod = period
	}

	// Allocate everything:

	m.data.ByPeriod = make([]*AggregativeValue, 0, len(m.aggregationPeriods)+1)
	v := m.NewAggregativeValue()
	m.data.ByPeriod = append(m.data.ByPeriod, v) // no aggregation, yet
	for range m.aggregationPeriods {
		v := m.NewAggregativeValue()
		m.data.ByPeriod = append(m.data.ByPeriod, v) // aggregated ones
	}

	// Init the downlaying structure
	m.common.init(parent, key, tags, func() bool { return atomic.LoadUint64(&m.data.ByPeriod[0].Count) == 0 })
}

// GetAggregationPeriods returns aggregation periods of the metric (see "Slicing" in README.md)
func (m *commonAggregative) GetAggregationPeriods() (r []AggregationPeriod) {
	m.Lock()
	r = make([]AggregationPeriod, len(m.aggregationPeriods))
	copy(r, aggregationPeriods.s)
	m.Unlock()
	return
}

// considerValue is an analog of method `Observe` of prometheus' metrics.
func (m *commonAggregative) considerValue(v float64) {
	if m == nil {
		return
	}

	appendData := func(data *AggregativeValue) {
		count := data.Count

		// It's already locked so we use `*Fast` methods

		if count == 0 || v < data.Min.GetFast() {
			data.Min.SetFast(v)
		}

		if count == 0 || v > data.Max.GetFast() {
			data.Max.SetFast(v)
		}

		data.Avg.SetFast((data.Avg.GetFast()*float64(count) + v) / (float64(count) + 1))
		if data.AggregativeStatistics != nil {
			data.AggregativeStatistics.ConsiderValue(v)
		}
		data.Count++
	}

	(*AggregativeValue)(atomic.LoadPointer((*unsafe.Pointer)((unsafe.Pointer)(&m.data.Current)))).LockDo(appendData)
	(*AggregativeValue)(atomic.LoadPointer((*unsafe.Pointer)((unsafe.Pointer)(&m.data.Total)))).LockDo(appendData)
	(*AggregativeValue)(atomic.LoadPointer((*unsafe.Pointer)((unsafe.Pointer)(&m.data.Last)))).set(v)

}

// GetValuePointers returns the pointer to the collection of aggregative values (min, max, ... for every aggregation
// period)
func (w *commonAggregative) GetValuePointers() *AggregativeValues {
	if w == nil {
		return &AggregativeValues{}
	}
	return &w.data
}

// String returns a JSON string representing values (min, max, count, ...) of an aggregative value
func (v *AggregativeValue) String() string {
	if v.AggregativeStatistics == nil {
		return fmt.Sprintf(`{"count":%d,"min":%g,"avg":%g,"max":%g}`,
			atomic.LoadUint64(&v.Count),
			v.Min.Get(),
			v.Avg.Get(),
			v.Max.Get(),
		)
	}
	percentiles := v.AggregativeStatistics.GetPercentiles([]float64{0.01, 0.1, 0.5, 0.9, 0.99})
	return fmt.Sprintf(`{"count":%d,"min":%g,"per1":%g,"per10":%g,"per50":%g,"avg":%g,"per90":%g,"per99":%g,"max":%g}`,
		atomic.LoadUint64(&v.Count),
		v.Min.Get(),
		*percentiles[0],
		*percentiles[1],
		*percentiles[2],
		v.Avg.Get(),
		*percentiles[3],
		*percentiles[4],
		v.Max.Get(),
	)
}

// MarshalJSON is a JSON marshalizer for an aggregative metric to be exported as JSON (for example
// using https://godoc.org/github.com/trafficstars/statuspage)
func (metric *commonAggregative) MarshalJSON() ([]byte, error) {
	var jsonValues []string

	considerValue := func(label string, data *AggregativeValue) {
		if data.Count == 0 {
			return
		}
		jsonValues = append(jsonValues, fmt.Sprintf(`"%v":%v`,
			label,
			data.String(),
		))
	}

	values := metric.data

	considerValue(`last`, values.Last)
	for idx, values := range metric.data.ByPeriod {
		considerValue(metric.aggregationPeriods[idx].String(), values)
	}
	considerValue(`total`, values.Total)

	nameJSON, _ := json.Marshal(metric.name)
	descriptionJSON, _ := json.Marshal(metric.description)
	tagsJSON, _ := json.Marshal(metric.tags.String())
	typeJSON, _ := json.Marshal(string(metric.GetType()))

	valueJSON := `{` + strings.Join(jsonValues, `,`) + `}`

	metricJSON := fmt.Sprintf(`{"name":%s,"tags":%s,"value":%s,"description":%s,"type":%s}`,
		string(nameJSON),
		tagsJSON,
		valueJSON,
		string(descriptionJSON),
		string(typeJSON),
	)
	return []byte(metricJSON), nil
}

// Send is a function to send the metric values through a Sender (see "Sender" in common.go)
func (m *commonAggregative) Send(sender Sender) {
	if sender == nil {
		return
	}

	considerValue := func(label string, data *AggregativeValue) {
		baseKey := string(m.storageKey) + `_` + label + `_`

		sender.SendUint64(m.parent, baseKey+`count`, atomic.LoadUint64(&data.Count))
		sender.SendFloat64(m.parent, baseKey+`min`, data.Min.Get())
		sender.SendFloat64(m.parent, baseKey+`avg`, data.Avg.Get())
		sender.SendFloat64(m.parent, baseKey+`max`, data.Max.Get())
		if data.AggregativeStatistics == nil {
			return
		}
		percentiles := data.AggregativeStatistics.GetPercentiles([]float64{0.01, 0.1, 0.5, 0.9, 0.99})
		sender.SendFloat64(m.parent, baseKey+`per1`, *percentiles[0])
		sender.SendFloat64(m.parent, baseKey+`per10`, *percentiles[1])
		sender.SendFloat64(m.parent, baseKey+`per50`, *percentiles[2])
		sender.SendFloat64(m.parent, baseKey+`per90`, *percentiles[3])
		sender.SendFloat64(m.parent, baseKey+`per99`, *percentiles[4])
	}

	values := m.data

	considerValue(`last`, values.Last)
	for idx, values := range m.data.ByPeriod {
		considerValue(m.aggregationPeriods[idx].String(), values)
	}
	considerValue(`total`, values.Total)
}

// Run starts the metric. We did not check if it is safe to call this method from external code.
// Not recommended to use it, yet (only for internal uses).
// Metrics starts automatically after it's creation, so there's no need to call this method, usually.
func (m *commonAggregative) Run(interval time.Duration) {
	if m == nil {
		return
	}
	if m.IsRunning() {
		return
	}
	m.Lock()
	m.run(interval)

	// We need not only to send the data to somewhere, but also to aggregate statistics. Our aggregation quant of time is one second, so it's required to aggregate the statistics once per second. So we create an object that will do that on method Iterate() and pass it to the `iterators`.
	iterationHandlers.Add(m.slicer)

	m.Unlock()
	return
}

// Stop is a function to stop the metric. It will be cleaned up by GC.
// Not recommended to use it, yet (only for internal uses).
// Metrics stops automatically if an counter of uselessness reaches a threshold (see "Garbage Collection" in README.md).
func (m *commonAggregative) Stop() {
	if m == nil {
		return
	}
	if !m.IsRunning() {
		return
	}
	m.Lock()
	m.stop()

	iterationHandlers.Remove(m.slicer)

	m.Unlock()
	return
}

// history is a structure that stores previous aggregative values for an aggregation period
// it's used to calculate statistics for higher aggregation periods (see "Slicing" in README.md).
//
// Actually "history" is just a cyclic buffer of "*AggregativeValue".
type history struct {
	currentOffset uint32
	storage       []*AggregativeValue
}

// histories is just a collection of "history"-ies for every aggregation period.
type histories struct {
	sync.Mutex

	ByPeriod []*history
}

// rotateHistory just shifts the pointer ("currentOffset") of the next element in the "history" to be filled.
// A "history" is a cyclic buffer, so it's just a rotation of a cyclic buffer.
func rotateHistory(h *history) {
	h.currentOffset++
	if h.currentOffset >= uint32(len(h.storage)) {
		h.currentOffset = 0
	}
}

// calculateValue just merges the statistics of all elements in the history and returns the result.
func (m *commonAggregative) calculateValue(h *history) (r *AggregativeValue) {
	depth := len(h.storage)
	offset := h.currentOffset
	if h.storage[offset] == nil {
		return
	}

	r = m.NewAggregativeValue()

	for depth > 0 {
		e := h.storage[offset]
		if e == nil {
			break
		}
		depth--
		offset--
		if offset == ^uint32(0) { // analog of "offset == -1", but for unsigned integer
			offset = uint32(len(h.storage) - 1)
		}

		r.MergeData(e)
	}
	return
}

// MergeData merges/joins the statistics of the argument.
func (r *AggregativeValue) MergeData(e *AggregativeValue) {
	if (e.Min < r.Min || (r.Count == 0 && e.Count != 0)) && e.Min != 0 {
		// TODO: should work correctly without "e.Min != 0" but it doesn't: min value is always zero
		r.Min = e.Min
	}
	if e.Max > r.Max || (r.Count == 0 && e.Count != 0) {
		r.Max = e.Max
	}

	addCount := e.Count
	addValue := e.Avg.GetFast()
	oldCount := r.Count
	oldValue := r.Avg.GetFast()
	if oldCount+addCount == 0 {
		r.Avg.SetFast(0)
	} else {
		r.Avg.SetFast((oldValue*float64(oldCount) + addValue*float64(addCount)) / float64(oldCount+addCount))
	}
	r.Count += addCount
	if e.AggregativeStatistics != nil {
		r.AggregativeStatistics.MergeStatistics(e.AggregativeStatistics)
	}
}

// considerFilledValue is a stage of the slicing process (see "Slicing" in README.md).
// When the slicer swaps the values, the previously "Current" value should be placed to the ByPeriod[0] of both:
// actual aggregative value (the actual value of the metric) and histories (to calculate values of higher
// aggregation periods).
func (m *commonAggregative) considerFilledValue(filledValue *AggregativeValue) {
	m.histories.Lock()
	defer m.histories.Unlock()

	tick := atomic.AddUint64(&m.tick, 1)

	updateLastHistoryRecord := func(h *history, newValue *AggregativeValue) {
		if h.storage[h.currentOffset] != nil {
			h.storage[h.currentOffset].Release()
		}
		h.storage[h.currentOffset] = newValue
	}

	// Store as the actual value
	atomic.StorePointer((*unsafe.Pointer)((unsafe.Pointer)(&m.data.ByPeriod[0])), (unsafe.Pointer)(filledValue))

	// Store to the history
	rotateHistory(m.histories.ByPeriod[0])
	updateLastHistoryRecord(m.histories.ByPeriod[0], filledValue)

	// Recalculate the actual values of higher aggregation periods
	for lIdx, aggregationPeriod := range m.aggregationPeriods {
		idx := lIdx + 1
		newValue := m.calculateValue(m.histories.ByPeriod[idx-1])
		oldValue := (*AggregativeValue)(atomic.SwapPointer((*unsafe.Pointer)((unsafe.Pointer)(&m.data.ByPeriod[idx])), (unsafe.Pointer)(newValue)))
		if idx < len(m.histories.ByPeriod) {
			if tick%aggregationPeriod.Interval == 0 {
				rotateHistory(m.histories.ByPeriod[idx])
			}
			updateLastHistoryRecord(m.histories.ByPeriod[idx], newValue)
		} else {
			oldValue.Release()
		}
	}
}

// DoSlice does the slicing (see "slicing" in README.md)
func (m *commonAggregative) DoSlice() {
	nextValue := m.NewAggregativeValue()
	filledValue := (*AggregativeValue)(atomic.SwapPointer((*unsafe.Pointer)((unsafe.Pointer)(&m.data.Current)), (unsafe.Pointer)(nextValue)))
	m.considerFilledValue(filledValue)
}

// GetFloat64 is required to be implemented by any metrics, so for aggregative metrics we use the last value.
func (m *commonAggregative) GetFloat64() float64 {
	return m.data.Last.GetAvg()
}
