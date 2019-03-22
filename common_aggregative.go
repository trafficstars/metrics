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

type AggregativeStatistics interface {
	// GetPercentile returns the value for a given percentile (0.0 .. 1.0).
	// It returns nil if the percentile could not be calculated (it could be in case of using "fast" [instead of "shortBuf"] aggregative metrics)
	//
	// If you need to calculate multiple percentiles then use GetPercentiles() to get better performance
	GetPercentile(percentile float64) *float64

	// GetPercentiles returns values for given percentiles (0.0 .. 1.0).
	// A value is nil if the percentile could not be calculated.
	GetPercentiles(percentile []float64) []*float64

	Set(staticValue float64)

	ConsiderValue(value float64)

	Release()

	MergeStatistics(AggregativeStatistics)
}

// SetSlicerInterval affects only new metrics (it doesn't affect already created one). You may use function `Reset()` to "update" configuration of all metrics.
func SetSlicerInterval(newSlicerInterval time.Duration) {
	slicerInterval = newSlicerInterval
}

// SetAggregationPeriods affects only new metrics (it doesn't affect already created on). You may use function `Reset()` to "update" configuration of all metrics.
func SetAggregationPeriods(newAggregationPeriods []AggregationPeriod) {
	aggregationPeriods.Lock()
	aggregationPeriods.s = newAggregationPeriods
	aggregationPeriods.Unlock()
}

type AggregationPeriod struct {
	Interval uint64 // in slicerInterval-s
}

func (period *AggregationPeriod) String() string {
	if period.Interval < 60 {
		return strconv.FormatUint(period.Interval, 10) + `s`
	}
	if period.Interval%(3600*24) == 0 {
		return strconv.FormatUint(period.Interval/(3600*24), 10) + `d`
	}
	if period.Interval%3600 == 0 {
		return strconv.FormatUint(period.Interval/3600, 10) + `h`
	}
	if period.Interval%60 == 0 {
		return strconv.FormatUint(period.Interval/60, 10) + `m`
	}
	return (time.Duration(period.Interval) * slicerInterval).String()
}

type aggregationPeriodsT struct {
	sync.RWMutex

	s []AggregationPeriod
}

var (
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

func GetBaseAggregationPeriod() *AggregationPeriod {
	return &AggregationPeriod{1}
}

func GetAggregationPeriods() (r []AggregationPeriod) {
	aggregationPeriods.RLock()
	r = make([]AggregationPeriod, len(aggregationPeriods.s))
	copy(r, aggregationPeriods.s)
	aggregationPeriods.RUnlock()
	return
}

type AggregativeValue struct {
	sync.Mutex

	Count uint64
	Min   AtomicFloat64
	Avg   AtomicFloat64
	Max   AtomicFloat64

	AggregativeStatistics
}

func NewAggregativeValue() *AggregativeValue {
	r := aggregativeValuePool.Get().(*AggregativeValue)
	r.Count = 0
	r.Min.Set(0)
	r.Avg.Set(0)
	r.Max.Set(0)
	r.AggregativeStatistics = nil
	return r
}

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
func (aggrV *AggregativeValue) LockDo(fn func(*AggregativeValue)) {
	if aggrV == nil {
		return
	}
	aggrV.Lock()
	fn(aggrV)
	aggrV.Unlock()
}

func (aggrV *AggregativeValue) GetAvg() float64 {
	if aggrV == nil {
		return 0
	}
	return aggrV.Avg.Get()
}

type AggregativeValues struct {
	Last     *AggregativeValue
	Current  *AggregativeValue
	ByPeriod []*AggregativeValue
	Total    *AggregativeValue
}

// slicer returns an object that will call method DoSlice() of metricCommonAggregative if method Iterate() was called.
type metricCommonAggregativeSlicer struct {
	metric   *metricCommonAggregative
	interval time.Duration
}

func (slicer *metricCommonAggregativeSlicer) Iterate() {
	slicer.metric.DoSlice()
}
func (slicer *metricCommonAggregativeSlicer) GetInterval() time.Duration {
	return slicer.interval
}
func (slicer *metricCommonAggregativeSlicer) IsRunning() bool {
	return slicer.metric.IsRunning()
}

func (m *metricCommonAggregativeSlicer) EqualsTo(cmpI iterator) bool {
	cmp, ok := cmpI.(*metricCommonAggregativeSlicer)
	if !ok {
		return false
	}
	return m == cmp
}

type metricCommonAggregative struct {
	metricCommon

	aggregationPeriods []AggregationPeriod
	data               AggregativeValues
	currentSliceData   *AggregativeValue
	tick               uint64
	slicer             iterator

	histories histories
}

func (m *metricCommonAggregative) init(parent Metric, key string, tags AnyTags) {
	m.slicer = &metricCommonAggregativeSlicer{
		metric:   m,
		interval: slicerInterval,
	}
	m.aggregationPeriods = GetAggregationPeriods()
	m.data.Last = NewAggregativeValue()
	m.data.Current = NewAggregativeValue()
	m.data.Total = NewAggregativeValue()

	m.histories.ByPeriod = make([]*history, 0, len(m.aggregationPeriods))
	previousPeriod := AggregationPeriod{1}
	for _, period := range m.aggregationPeriods {
		hist := &history{}
		if period.Interval%previousPeriod.Interval != 0 {
			// TODO: print error
			//panic(fmt.Errorf("period.Interval (%v) %% previousPeriod.Interval (%v) != 0 (%v)", period.Interval, previousPeriod.Interval, period.Interval%previousPeriod.Interval))
		}
		hist.storage = make([]*AggregativeValue, period.Interval/previousPeriod.Interval)

		m.histories.ByPeriod = append(m.histories.ByPeriod, hist)
		previousPeriod = period
	}

	m.parent = parent

	m.data.ByPeriod = make([]*AggregativeValue, 0, len(m.aggregationPeriods)+1)
	v := NewAggregativeValue()
	v.AggregativeStatistics = m.newAggregativeStatistics()
	m.data.ByPeriod = append(m.data.ByPeriod, v) // no aggregation, yet
	for range m.aggregationPeriods {
		v := NewAggregativeValue()
		v.AggregativeStatistics = m.newAggregativeStatistics()
		m.data.ByPeriod = append(m.data.ByPeriod, v) // aggregated ones
	}

	m.metricCommon.init(parent, key, tags, func() bool { return atomic.LoadUint64(&m.data.ByPeriod[0].Count) == 0 })
}

func (m *metricCommonAggregative) GetAggregationPeriods() (r []AggregationPeriod) {
	m.Lock()
	r = make([]AggregationPeriod, len(m.aggregationPeriods))
	copy(r, aggregationPeriods.s)
	m.Unlock()
	return
}

func (m *metricCommonAggregative) considerValue(v float64) {
	if m == nil {
		return
	}

	appendData := func(data *AggregativeValue) {
		count := data.Count
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

func (w *metricCommonAggregative) GetValuePointers() *AggregativeValues {
	if w == nil {
		return &AggregativeValues{}
	}
	return &w.data
}

func (v *AggregativeValue) String() string {
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

func (metric *metricCommonAggregative) MarshalJSON() ([]byte, error) {
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
	tagsJSON, _ := json.Marshal(string(metric.storageKey[:strings.IndexByte(string(metric.storageKey), '@')]))
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

func (m *metricCommonAggregative) Send(sender Sender) {
	if sender == nil {
		return
	}

	considerValue := func(label string, data *AggregativeValue) {
		baseKey := string(m.storageKey) + `_` + label + `_`

		percentiles := data.AggregativeStatistics.GetPercentiles([]float64{0.01, 0.1, 0.5, 0.9, 0.99})
		sender.SendUint64(m.parent, baseKey+`count`, atomic.LoadUint64(&data.Count))
		sender.SendFloat64(m.parent, baseKey+`min`, data.Min.Get())
		sender.SendFloat64(m.parent, baseKey+`per1`, *percentiles[0])
		sender.SendFloat64(m.parent, baseKey+`per10`, *percentiles[1])
		sender.SendFloat64(m.parent, baseKey+`per50`, *percentiles[2])
		sender.SendFloat64(m.parent, baseKey+`avg`, data.Avg.Get())
		sender.SendFloat64(m.parent, baseKey+`per90`, *percentiles[3])
		sender.SendFloat64(m.parent, baseKey+`per99`, *percentiles[4])
		sender.SendFloat64(m.parent, baseKey+`max`, data.Max.Get())
	}

	values := m.data

	considerValue(`last`, values.Last)
	for idx, values := range m.data.ByPeriod {
		considerValue(m.aggregationPeriods[idx].String(), values)
	}
	considerValue(`total`, values.Total)
}

// Run starts the metric. We did not check if it is safe to call this method from external code. Not recommended to use, yet.
// Metrics starts automatically after it's creation, so there's no need to call this method, usually.
func (m *metricCommonAggregative) Run(interval time.Duration) {
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

func (m *metricCommonAggregative) Stop() {
	if m == nil {
		return
	}
	if !m.IsRunning() {
		return
	}
	return
	m.Lock()
	m.stop()

	iterationHandlers.Remove(m.slicer)

	m.Unlock()
	return
}

type history struct {
	currentOffset uint32
	storage       []*AggregativeValue
}

type histories struct {
	sync.Mutex

	ByPeriod []*history
}

func rotateHistory(h *history) {
	h.currentOffset++
	if h.currentOffset >= uint32(len(h.storage)) {
		h.currentOffset = 0
	}
}

func (m *metricCommonAggregative) calculateValue(h *history) (r *AggregativeValue) {
	depth := len(h.storage)
	offset := h.currentOffset
	if h.storage[offset] == nil {
		return
	}

	r = NewAggregativeValue()
	r.AggregativeStatistics = m.newAggregativeStatistics()

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

	r.NormalizeData()
	return
}

func (r *AggregativeValue) MergeData(e *AggregativeValue) {
	if (e.Min < r.Min || (r.Count == 0 && e.Count != 0)) && e.Min != 0 { // TODO: should work correctly without "e.Min != 0" but it doesn't: min value is always zero
		r.Min = e.Min
	}
	if e.Max > r.Max || (r.Count == 0 && e.Count != 0) {
		r.Max = e.Max
	}

	count := e.Count
	r.Count += count
	r.Avg.SetFast(r.Avg.GetFast() + e.Avg.GetFast()*float64(count))
	r.AggregativeStatistics.MergeStatistics(e.AggregativeStatistics)
}

func (r *AggregativeValue) NormalizeData() {
	count := r.Count
	if count == 0 {
		return
	}

	r.Avg.SetFast(r.Avg.GetFast() / float64(count))
}

func (m *metricCommonAggregative) considerFilledValue(filledValue *AggregativeValue) {
	m.histories.Lock()
	defer m.histories.Unlock()

	tick := atomic.AddUint64(&m.tick, 1)

	updateLastHistoryRecord := func(h *history, newValue *AggregativeValue) {
		if h.storage[h.currentOffset] != nil {
			h.storage[h.currentOffset].Release()
		}
		h.storage[h.currentOffset] = newValue
	}

	atomic.StorePointer((*unsafe.Pointer)((unsafe.Pointer)(&m.data.ByPeriod[0])), (unsafe.Pointer)(filledValue))
	rotateHistory(m.histories.ByPeriod[0])
	updateLastHistoryRecord(m.histories.ByPeriod[0], filledValue)

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

func (m *metricCommonAggregative) newAggregativeStatistics() AggregativeStatistics {
	return m.parent.(interface{ NewAggregativeStatistics() AggregativeStatistics }).NewAggregativeStatistics()
}

func (m *metricCommonAggregative) DoSlice() {
	nextValue := NewAggregativeValue()
	nextValue.AggregativeStatistics = m.newAggregativeStatistics()
	filledValue := (*AggregativeValue)(atomic.SwapPointer((*unsafe.Pointer)((unsafe.Pointer)(&m.data.Current)), (unsafe.Pointer)(nextValue)))
	m.considerFilledValue(filledValue)
}

func (m *metricCommonAggregative) GetFloat64() float64 {
	return m.data.Last.GetAvg()
}
