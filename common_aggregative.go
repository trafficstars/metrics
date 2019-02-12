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
		return strconv.FormatUint(period.Interval, 10) + `d`
	}
	if period.Interval%3600 == 0 {
		return strconv.FormatUint(period.Interval, 10) + `h`
	}
	if period.Interval%60 == 0 {
		return strconv.FormatUint(period.Interval, 10) + `m`
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

// Release is an opposite to NewAggregativeValue and it saves the variable to a pool to a prevent memory allocation in future.
// It's not necessary to call this method when you finished to work with an AggregativeValue, but recommended to (for better performance).
func (v *AggregativeValue) Release() {
	if v == nil {
		return
	}
	if v.AggregativeStatistics != nil {
		v.AggregativeStatistics.Release()
	}
	aggregativeValuePool.Put(v)
}

func (aggrV *AggregativeValue) set(v float64) {
	atomic.StoreUint64(&aggrV.Count, 1)
	aggrV.Min.Set(v)
	aggrV.Avg.Set(v)
	aggrV.Max.Set(v)
	aggrV.AggregativeStatistics.Set(v)
}

type AggregativeValues struct {
	Last     *AggregativeValue
	Current  *AggregativeValue
	ByPeriod []*AggregativeValue
	Total    *AggregativeValue
}

type doSlicer interface {
	DoSlice()
}

// slicer returns an object that will call method DoSlice() of metricCommonAggregative if method Iterate() was called.
type metricCommonAggregativeSlicer struct {
	metric *metricCommonAggregative
}

func (slicer *metricCommonAggregativeSlicer) Iterate() {
	slicer.metric.doSlicer.DoSlice()
}
func (slicer *metricCommonAggregativeSlicer) GetInterval() time.Duration {
	return slicerInterval
}
func (slicer *metricCommonAggregativeSlicer) IsRunning() bool {
	return slicer.metric.IsRunning()
}

type metricCommonAggregative struct {
	metricCommon

	aggregationPeriods []AggregationPeriod
	data               AggregativeValues
	currentSliceData   *AggregativeValue
	tick               uint64
	doSlicer           doSlicer
	slicer             iterator
}

func (metric *metricCommonAggregative) init(parent Metric, key string, tags AnyTags) {
	metric.slicer = &metricCommonAggregativeSlicer{
		metric: metric,
	}
	metric.aggregationPeriods = GetAggregationPeriods()
	metric.data.Last = NewAggregativeValue()
	metric.data.Current = NewAggregativeValue()
	metric.data.Total = NewAggregativeValue()
	metric.data.ByPeriod = make([]*AggregativeValue, 0, len(metric.aggregationPeriods)+1)
	metric.data.ByPeriod = append(metric.data.ByPeriod, NewAggregativeValue()) // no aggregation, yet
	for range metric.aggregationPeriods {
		metric.data.ByPeriod = append(metric.data.ByPeriod, NewAggregativeValue()) // aggregated ones
	}
	metric.metricCommon.init(parent, key, tags, func() bool { return atomic.LoadUint64(&metric.data.ByPeriod[0].Count) == 0 })
}

func (m *metricCommonAggregative) considerValue(v float64) {
	if m == nil {
		return
	}

	appendData := func(data *AggregativeValue) {
		data.Lock()
		count := data.Count
		if count == 0 || v < data.Min.GetFast() {
			data.Min.SetFast(v)
		}

		if count == 0 || v > data.Max.GetFast() {
			data.Max.SetFast(v)
		}

		data.Avg.SetFast((data.Avg.GetFast()*float64(count) + v) / (float64(count) + 1))
		data.AggregativeStatistics.ConsiderValue(v)
		data.Count++
		data.Unlock()
	}

	appendData((*AggregativeValue)(atomic.LoadPointer((*unsafe.Pointer)((unsafe.Pointer)(&m.data.Current)))))
	appendData((*AggregativeValue)(atomic.LoadPointer((*unsafe.Pointer)((unsafe.Pointer)(&m.data.Total)))))
	(*AggregativeValue)(atomic.LoadPointer((*unsafe.Pointer)((unsafe.Pointer)(&m.data.Last)))).set(v)

}

func (w *metricCommonAggregative) GetValuePointers() *AggregativeValues {
	if w == nil {
		return &AggregativeValues{}
	}
	return &w.data
}

func (metric *metricCommonAggregative) MarshalJSON() ([]byte, error) {
	var jsonValues []string

	considerValue := func(label string, data *AggregativeValue) {
		if data.Count == 0 {
			return
		}
		percentiles := data.AggregativeStatistics.GetPercentiles([]float64{0.01, 0.1, 0.5, 0.9, 0.99})
		jsonValues = append(jsonValues, fmt.Sprintf(`"%v":{"count":%d,"min":%g,"per1":%g,"per10":%g,"per50":%g,"avg":%g,"per90":%g,"per99":%g,"max":%g}`,
			label,
			atomic.LoadUint64(&data.Count),
			data.Min.Get(),
			*percentiles[0],
			*percentiles[1],
			*percentiles[2],
			data.Avg.Get(),
			*percentiles[3],
			*percentiles[4],
			data.Max.Get(),
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
	m.Lock()
	m.stop()

	iterationHandlers.Remove(m.slicer)

	m.Unlock()
	return
}
