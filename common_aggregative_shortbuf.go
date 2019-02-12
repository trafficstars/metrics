package metrics

import (
	"math/rand"
	"sort"
)

const (
	// Buffer size. The more this buffer the more CPU is utilized and more precise values are.
	bufferSize = 1000
)

type aggregativeBufferItems [bufferSize]float64

func (s aggregativeBufferItems) swap(i, j uint32) {
	s[i], s[j] = s[j], s[i]
}

func (s aggregativeBufferItems) qsort_partition(p uint32, q uint32, pivotIdx uint32) uint32 {
	pivot := s[pivotIdx]
	s.swap(pivotIdx, q)
	i := p
	for j := p; j < q; j++ {
		if s[j] <= pivot {
			s.swap(i, j)
			i++
		}
	}
	s.swap(q, i)
	return i
}

func (s aggregativeBufferItems) qsort(start uint32, end uint32) {
	if start >= end {
		return
	}

	pivot := (end + start) / 2
	r := s.qsort_partition(start, end, pivot)
	if r > start {
		s.qsort(start, r-1)
	}
	s.qsort(r+1, end)
}

func (s *aggregativeBuffer) sortNative() {
	s.data.qsort(0, s.filledSize)
}

func (s *aggregativeBuffer) sortBuiltin() {
	sort.Slice(s.data[:], func(i, j int) bool{ return s.data[i] < s.data[j] })
}

func (s *aggregativeBuffer) sort() {
	if s.isSorted {
		s.locker.Unlock()
		return
	}
	s.sortNative()
	s.isSorted = true
}

func (s *aggregativeBuffer) Sort() {
	s.locker.Lock()
	s.sort()
	s.locker.Unlock()
}

type aggregativeBuffer struct {
	locker Spinlock
	filledSize    uint32
	data          aggregativeBufferItems
	isSorted      bool
}

type metricCommonAggregativeShortBuf struct {
	metricCommonAggregative
}

func (m *metricCommonAggregativeShortBuf) init(parent Metric, key string, tags AnyTags) {
	m.doSlicer = m
	m.metricCommonAggregative.init(parent, key, tags)
	m.data.Current.AggregativeStatistics = newAggregativeStatisticsShortBuf()
	m.data.Last.AggregativeStatistics = newAggregativeStatisticsShortBuf()
	m.data.Total.AggregativeStatistics = newAggregativeStatisticsShortBuf()
}

type AggregativeStatisticsShortBuf struct {
	aggregativeBuffer

	tickID uint64
}

func (s *AggregativeStatisticsShortBuf) getPercentile(percentile float64) *float64 {
	if s.filledSize == 0 {
		s.locker.Unlock()
		return nil
	}
	s.sort()
	percentileIdx := int(float64(s.filledSize) * percentile)
	return (*float64)(&s.data[percentileIdx])
}

func (s *AggregativeStatisticsShortBuf) GetPercentile(percentile float64) *float64 {
	s.locker.Lock()
	r := s.getPercentile(percentile)
	s.locker.Unlock()
	return r
}

func (s *AggregativeStatisticsShortBuf) GetPercentiles(percentiles []float64) []*float64 {
	r := make([]*float64, 0, len(percentiles))
	s.locker.Lock()
	for _, percentile := range percentiles {
		r = append(r, s.getPercentile(percentile))
	}
	s.locker.Unlock()
	return r
}

func (s *AggregativeStatisticsShortBuf) ConsiderValue(v float64) {
	s.tickID++
	if s.filledSize < bufferSize {
		s.data[s.filledSize] = v
		s.filledSize++

		s.locker.Unlock()
		return
	}

	// The more history we have the more rarely we should update items
	// That's why here's rand.Intn(s.tickID) instead of rand.Intn(bufferSize)
	randIdx := rand.Intn(int(s.tickID))
	if randIdx >= bufferSize {
		s.locker.Unlock()
		return
	}

	s.data[randIdx] = v

	s.locker.Unlock()
}

/*func (s *AggregativeStatisticsShortBuf) setItem(idx int, value float64, tickID uint64) {
	newItem := newAggregativeBufferItem()
	newItem.value = value
	newItem.tickID = tickID
	(*aggregativeBufferItem)(atomic.SwapPointer((*unsafe.Pointer)((unsafe.Pointer)(&s.data[idx]), newItem)).Release()
}*/

func (s *AggregativeStatisticsShortBuf) Set(value float64) {
	s.locker.Lock()
	s.data[0] = value
	s.filledSize = 1
	s.locker.Unlock()
}

func (m *metricCommonAggregativeShortBuf) DoSlice() {
}
