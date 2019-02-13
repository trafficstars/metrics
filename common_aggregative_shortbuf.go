package metrics

import (
	"math/rand"
	"sort"
)

const (
	// Buffer size. The more this buffer the more CPU is utilized (on metric `GetPercentiles` which is used by `List()`) and more precise values are.
	bufferSize = 1000
)

type aggregativeBufferItems [bufferSize]float64

func (s *aggregativeBuffer) sortBuiltin() {
	sort.Slice(s.data[:s.filledSize], func(i, j int) bool { return s.data[i] < s.data[j] })
}

func (s *aggregativeBuffer) sort() {
	if s.isSorted {
		return
	}
	s.sortBuiltin()
	s.isSorted = true
}

func (s *aggregativeBuffer) Sort() {
	s.locker.Lock()
	s.sort()
	s.locker.Unlock()
}

type aggregativeBuffer struct {
	locker     Spinlock
	filledSize uint32
	data       aggregativeBufferItems
	isSorted   bool
}

type metricCommonAggregativeShortBuf struct {
	metricCommonAggregative
}

func (m *metricCommonAggregativeShortBuf) init(parent Metric, key string, tags AnyTags) {
	m.metricCommonAggregative.init(parent, key, tags)
	m.data.Current.AggregativeStatistics = newAggregativeStatisticsShortBuf()
	m.data.Last.AggregativeStatistics = newAggregativeStatisticsShortBuf()
	m.data.Total.AggregativeStatistics = newAggregativeStatisticsShortBuf()
}

func (m *metricCommonAggregativeShortBuf) NewAggregativeStatistics() AggregativeStatistics {
	return newAggregativeStatisticsShortBuf()
}

type AggregativeStatisticsShortBuf struct {
	aggregativeBuffer

	tickID uint64
}

func (s *AggregativeStatisticsShortBuf) getPercentile(percentile float64) *float64 {
	if s.filledSize == 0 {
		return &[]float64{0}[0]
	}
	s.sort()
	percentileIdx := int(float64(s.filledSize) * percentile)
	return &[]float64{s.data[percentileIdx]}[0]
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

func (s *AggregativeStatisticsShortBuf) considerValue(v float64) {
	s.tickID++
	if s.filledSize < bufferSize {
		s.data[s.filledSize] = v
		s.filledSize++

		return
	}

	// The more history we have the more rarely we should update items
	// That's why here's rand.Intn(s.tickID) instead of rand.Intn(bufferSize)
	randIdx := rand.Intn(int(s.tickID))
	if randIdx >= bufferSize {
		return
	}

	s.data[randIdx] = v

	s.isSorted = false
}

func (s *AggregativeStatisticsShortBuf) ConsiderValue(v float64) {
	s.locker.Lock()
	s.considerValue(v)
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

func (s *AggregativeStatisticsShortBuf) MergeStatistics(oldSI AggregativeStatistics) {
	if oldSI == nil {
		return
	}
	oldS := oldSI.(*AggregativeStatisticsShortBuf)

	if s.filledSize+oldS.filledSize <= bufferSize {
		copy(s.data[s.filledSize:], oldS.data[:oldS.filledSize])
		s.filledSize += oldS.filledSize
		s.tickID += oldS.tickID
		// nothing overlaps, done
		return
	}

	origFilledSize := s.filledSize
	if s.filledSize < bufferSize {
		delta := bufferSize - s.filledSize
		copy(s.data[s.filledSize:], oldS.data[oldS.filledSize-delta:])
		s.filledSize = bufferSize
		oldS.filledSize -= delta
	}

	indexes := rand.Perm(int(origFilledSize))

	ratio := float64(oldS.tickID) / float64(s.tickID+oldS.tickID)
	for idx, value := range oldS.data[:oldS.filledSize] {
		if ratio > rand.Float64() {
			s.data[indexes[idx]] = value
		}
	}

	s.tickID += oldS.tickID
}
