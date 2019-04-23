package metrics

import (
	"math/rand"
	"sort"
)

const (
	// The default value of the buffer size. The more this buffer the more CPU is utilized (on metric `GetPercentiles`
	// which is used by `List()`), the more RAM is utilized and more precise values are.
	defaultBufferSize = 1000
)

var (
	// See "Buffered" in README.md
	bufferSize = uint(defaultBufferSize)
)

// SetAggregativeBufferSize sets the size of the buffer to be used to store value samples
// The more this values is the more precise is the percentile value, but more RAM & CPU is consumed.
// (see "Buffered" in README.md)
func SetAggregativeBufferSize(newBufferSize uint) {
	bufferSize = newBufferSize
}

type aggregativeBufferItems []float64

// sortBuiltin uses golang's builtin sort function to sort the slice
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

// Sort just sorts the values in the buffer in the ascending order
//
// It's used to get a percentile value.
func (s *aggregativeBuffer) Sort() {
	s.locker.Lock()
	s.sort()
	s.locker.Unlock()
}

// aggregativeBuffer is a collection of values to be used for percentile calculations (see "Buffered" in README.md)
type aggregativeBuffer struct {
	locker     Spinlock
	filledSize uint32
	data       aggregativeBufferItems
	isSorted   bool
}

type commonAggregativeBuffered struct {
	commonAggregative
}

func (m *commonAggregativeBuffered) init(parent Metric, key string, tags AnyTags) {
	m.commonAggregative.init(parent, key, tags)
}

// NewAggregativeStatistics returns a "Buffered" (see "Buffered" in README.md) implementation of AggregativeStatistics.
func (m *commonAggregativeBuffered) NewAggregativeStatistics() AggregativeStatistics {
	return newAggregativeStatisticsBuffered()
}

type aggregativeStatisticsBuffered struct {
	aggregativeBuffer

	tickID uint64
}

func (s *aggregativeStatisticsBuffered) getPercentile(percentile float64) *float64 {
	if s.filledSize == 0 {
		return &[]float64{0}[0]
	}
	percentileIdx := int(float64(s.filledSize) * percentile)
	return &[]float64{s.data[percentileIdx]}[0]
}

// GetPercentile returns a percentile value for a given percentile (see https://en.wikipedia.org/wiki/Percentile).
//
// There will never be returned "nil" (because it's a "Buffered" aggregative statistics).
//
// If you need multiple percentile values then it would be better to use method `GetPercentiles`, it
// works faster (for multiple values).
func (s *aggregativeStatisticsBuffered) GetPercentile(percentile float64) *float64 {
	s.locker.Lock()
	s.sort()
	r := s.getPercentile(percentile)
	s.locker.Unlock()
	return r
}

// GetPercentiles returns percentile values for a given slice of percentiles.
//
// Returned values are ordered accordingly to the input slice. An element of the returned
// slice is never "nil" (because it's a "Buffered" aggregative statistics).
func (s *aggregativeStatisticsBuffered) GetPercentiles(percentiles []float64) []*float64 {
	r := make([]*float64, 0, len(percentiles))
	s.locker.Lock()
	s.sort()
	for _, percentile := range percentiles {
		r = append(r, s.getPercentile(percentile))
	}
	s.locker.Unlock()
	return r
}

func (s *aggregativeStatisticsBuffered) considerValue(v float64) {
	s.tickID++
	if s.filledSize < uint32(bufferSize) {
		s.isSorted = false
		s.data[s.filledSize] = v
		// We don't want to use atomic write because it's a much more expensive operation.
		// So we just set "isSorted = false" twice: before assigning the value and after
		s.isSorted = false
		s.filledSize++

		return
	}

	// The more history we have the more rarely we should update items
	// That's why here's rand.Intn(s.tickID) instead of rand.Intn(bufferSize)
	randIdx := rand.Intn(int(s.tickID))
	if randIdx >= int(bufferSize) {
		return
	}

	s.isSorted = false
	s.data[randIdx] = v
	// We don't want to use atomic write because it's a much more expensive operation.
	// So we just set "isSorted = false" twice: before assigning the value and after
	s.isSorted = false
}

// ConsiderValue is an analog of Prometheus' observe (see "Aggregative metrics" in README.md)
func (s *aggregativeStatisticsBuffered) ConsiderValue(v float64) {
	s.locker.Lock()
	s.considerValue(v)
	s.locker.Unlock()
}

/*func (s *AggregativeStatisticsBuffered) setItem(idx int, value float64, tickID uint64) {
	newItem := newAggregativeBufferItem()
	newItem.value = value
	newItem.tickID = tickID
	(*aggregativeBufferItem)(atomic.SwapPointer((*unsafe.Pointer)((unsafe.Pointer)(&s.data[idx]), newItem)).Release()
}*/

// Set resets the statistics and sets only one event with the value passed as the argument,
// so all aggregative values (avg, min, max, ...) will be equal to the value
func (s *aggregativeStatisticsBuffered) Set(value float64) {
	s.locker.Lock()
	s.data[0] = value
	s.filledSize = 1
	s.locker.Unlock()
}

// MergeStatistics adds statistics of the argument to the own one (see "Buffer handling" in README.md)
func (s *aggregativeStatisticsBuffered) MergeStatistics(oldSI AggregativeStatistics) {
	if oldSI == nil {
		return
	}
	oldS := oldSI.(*aggregativeStatisticsBuffered)

	if s.filledSize+oldS.filledSize <= uint32(bufferSize) {
		copy(s.data[s.filledSize:], oldS.data[:oldS.filledSize])
		s.filledSize += oldS.filledSize
		s.tickID += oldS.tickID
		// nothing overlaps, done
		return
	}

	origFilledSize := s.filledSize
	if s.filledSize < uint32(bufferSize) {
		delta := uint32(bufferSize) - s.filledSize
		copy(s.data[s.filledSize:], oldS.data[oldS.filledSize-delta:])
		s.filledSize = uint32(bufferSize)
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
