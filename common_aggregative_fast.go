package metrics

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"unsafe"
)

const (
	// If this value is too large then it will be required too many events
	// per second to calculate percentile correctly (Per50, Per99 etc).
	// If this value is too small then the percentile will be calculated
	// not accurate.
	iterationsRequiredPerSecond = 20
)

type history struct {
	currentOffset uint32
	storage       []*AggregativeValue
}

type histories struct {
	sync.Mutex

	ByPeriod []*history
}

type metricCommonAggregativeFast struct {
	metricCommonAggregative

	histories histories
}

func (m *metricCommonAggregativeFast) init(parent Metric, key string, tags AnyTags) {
	m.doSlicer = m
	m.metricCommonAggregative.init(parent, key, tags)
	m.histories.ByPeriod = make([]*history, 0, len(m.aggregationPeriods))
	for idx, period := range m.aggregationPeriods {
		hist := &history{}
		if idx+1 < len(m.aggregationPeriods) {
			nextPeriod := m.aggregationPeriods[idx+1]
			if nextPeriod.Interval%period.Interval != 0 {
				panic(fmt.Errorf("nextPeriod.Interval (%v) %% period.Interval (%v) != 0 (%v)", nextPeriod.Interval, period.Interval, nextPeriod.Interval%period.Interval))
			}
			hist.storage = make([]*AggregativeValue, nextPeriod.Interval/period.Interval)
		}
		m.histories.ByPeriod = append(m.histories.ByPeriod, hist)
	}
	m.data.Current.AggregativeStatistics = newAggregativeStatisticsFast()
	m.data.Last.AggregativeStatistics = newAggregativeStatisticsFast()
	m.data.Total.AggregativeStatistics = newAggregativeStatisticsFast()
}

// this is so-so correct only for big amount of events (> iterationsRequiredPerSecond)
func guessPercentile(curValue, newValue float64, count uint64, perc float64) float64 {
	inertness := float64(count) / iterationsRequiredPerSecond

	requireGreater := rand.Float64() > perc

	if newValue > curValue {
		if requireGreater {
			return curValue
		}
	} else {
		if !requireGreater {
			return curValue
		}
	}

	if requireGreater {
		inertness *= float64(perc)
	} else {
		inertness *= float64(1 - perc)
	}
	return (curValue*inertness + newValue) / (inertness + 1)
}

func (m *metricCommonAggregativeFast) considerValue(v float64) {
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
		stat := data.AggregativeStatistics.(*AggregativeStatisticsFast)
		if count == 0 {
			stat.Per1.SetFast(v)
			stat.Per10.SetFast(v)
			stat.Per50.SetFast(v)
			stat.Per90.SetFast(v)
			stat.Per99.SetFast(v)
		} else {
			stat.Per1.SetFast(guessPercentile(stat.Per1.GetFast(), v, count, 0.01))
			stat.Per10.SetFast(guessPercentile(stat.Per10.GetFast(), v, count, 0.1))
			stat.Per50.SetFast(guessPercentile(stat.Per50.GetFast(), v, count, 0.5))
			stat.Per90.SetFast(guessPercentile(stat.Per90.GetFast(), v, count, 0.9))
			stat.Per99.SetFast(guessPercentile(stat.Per99.GetFast(), v, count, 0.99))
		}

		data.Count++
		data.Unlock()
	}


	appendData((*AggregativeValue)(atomic.LoadPointer((*unsafe.Pointer)((unsafe.Pointer)(&m.data.Current)))))
	appendData((*AggregativeValue)(atomic.LoadPointer((*unsafe.Pointer)((unsafe.Pointer)(&m.data.Total)))))
	(*AggregativeValue)(atomic.LoadPointer((*unsafe.Pointer)((unsafe.Pointer)(&m.data.Last)))).set(v)

}

func rotateHistory(h *history) {
	h.currentOffset++
	if h.currentOffset >= uint32(len(h.storage)) {
		h.currentOffset = 0
	}
}

type AggregativeStatisticsFast struct {
	Per1  AtomicFloat64Ptr
	Per10 AtomicFloat64Ptr
	Per50 AtomicFloat64Ptr
	Per90 AtomicFloat64Ptr
	Per99 AtomicFloat64Ptr
}

func (s *AggregativeStatisticsFast) GetPercentile(percentile float64) *float64 {
	switch percentile {
	case 0.01:
		return s.Per1.Pointer
	case 0.1:
		return s.Per10.Pointer
	case 0.5:
		return s.Per50.Pointer
	case 0.9:
		return s.Per90.Pointer
	case 0.99:
		return s.Per99.Pointer
	}
	return nil
}

func (s *AggregativeStatisticsFast) Set(value float64) {
	s.Per1.Set(value)
	s.Per10.Set(value)
	s.Per50.Set(value)
	s.Per90.Set(value)
	s.Per99.Set(value)
}

func (s *AggregativeStatisticsFast) Release() {
	s.Set(0)
	aggregativeStatisticsFastPool.Put(s)
}

func newAggregativeStatisticsFast() *AggregativeStatisticsFast {
	s := aggregativeStatisticsFastPool.Get().(*AggregativeStatisticsFast)
	return s
}

func (w *metricCommonAggregativeFast) calculateValue(h *history) (r *AggregativeValue) {
	depth := len(h.storage)
	offset := h.currentOffset
	if h.storage[offset] == nil {
		return
	}

	r = NewAggregativeValue()
	s := newAggregativeStatisticsFast()
	r.AggregativeStatistics = s

	for depth > 0 {
		e := h.storage[offset]
		if e == nil {
			break
		}
		oldS := e.AggregativeStatistics.(*AggregativeStatisticsFast)
		depth--
		offset--
		if offset == ^uint32(0) { // analog of "offset == -1", but for unsigned integer
			offset = uint32(len(h.storage) - 1)
		}

		if (e.Min < r.Min || (r.Count == 0 && e.Count != 0)) && e.Min != 0 { // TODO: should work correctly without "e.Min != 0" but it doesn't: min value is always zero
			r.Min = e.Min
		}
		if e.Max > r.Max || (r.Count == 0 && e.Count != 0) {
			r.Max = e.Max
		}

		count := e.Count

		r.Count += count

		s.Per1.SetFast(s.Per1.GetFast() + oldS.Per1.GetFast()*float64(count))
		s.Per10.SetFast(s.Per10.GetFast() + oldS.Per10.GetFast()*float64(count))
		s.Per50.SetFast(s.Per50.GetFast() + oldS.Per50.GetFast()*float64(count))
		r.Avg.SetFast(r.Avg.GetFast() + e.Avg.GetFast()*float64(count))
		s.Per90.SetFast(s.Per90.GetFast() + oldS.Per90.GetFast()*float64(count))
		s.Per99.SetFast(s.Per99.GetFast() + oldS.Per99.GetFast()*float64(count))
	}

	count := r.Count
	if count == 0 {
		return
	}

	r.Avg.SetFast(r.Avg.GetFast() / float64(count))

	// it seems to be incorrent, but I don't see other fast way to calculate it, yet
	s.Per1.SetFast(s.Per1.GetFast() / float64(count))
	s.Per10.SetFast(s.Per10.GetFast() / float64(count))
	s.Per50.SetFast(s.Per50.GetFast() / float64(count))
	s.Per90.SetFast(s.Per90.GetFast() / float64(count))
	s.Per99.SetFast(s.Per99.GetFast() / float64(count))

	return
}

func (m *metricCommonAggregativeFast) considerFilledValue(filledValue *AggregativeValue) {
	m.histories.Lock()
	defer m.histories.Unlock()

	tick := atomic.AddUint64(&m.tick, 1)

	updateLastHistoryRecord := func(h *history, newValue *AggregativeValue) {
		h.storage[h.currentOffset] = newValue
	}

	(*AggregativeValue)(atomic.SwapPointer((*unsafe.Pointer)((unsafe.Pointer)(&m.data.ByPeriod[0])), (unsafe.Pointer)(filledValue))).Release()
	rotateHistory(m.histories.ByPeriod[0])
	updateLastHistoryRecord(m.histories.ByPeriod[0], filledValue)

	if len(m.aggregationPeriods) > 1 {
		for lIdx, aggregationPeriod := range m.aggregationPeriods[1:] {
			idx := lIdx + 1
			newValue := m.calculateValue(m.histories.ByPeriod[idx-1])
			(*AggregativeValue)(atomic.SwapPointer((*unsafe.Pointer)((unsafe.Pointer)(&m.data.ByPeriod[idx])), (unsafe.Pointer)(newValue))).Release()
			if tick%aggregationPeriod.Interval == 0 {
				rotateHistory(m.histories.ByPeriod[idx])
			}
			if idx+1 < len(m.histories.ByPeriod) {
				updateLastHistoryRecord(m.histories.ByPeriod[idx], newValue)
			}
		}
	}
}

func (m *metricCommonAggregativeFast) DoSlice() {
	nextValue := NewAggregativeValue()
	nextValue.AggregativeStatistics = newAggregativeStatisticsFast()
	filledValue := (*AggregativeValue)(atomic.SwapPointer((*unsafe.Pointer)((unsafe.Pointer)(&m.data.Current)), (unsafe.Pointer)(nextValue)))
	m.considerFilledValue(filledValue)
}
