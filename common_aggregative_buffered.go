package metrics

import (
	"math"
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

func (m *commonAggregativeBuffered) init(r *Registry, parent Metric, key string, tags AnyTags) {
	m.commonAggregative.init(r, parent, key, tags)
}

// NewAggregativeStatistics returns a "Buffered" (see "Buffered" in README.md) implementation of AggregativeStatistics.
func (m *commonAggregativeBuffered) NewAggregativeStatistics() AggregativeStatistics {
	return newAggregativeStatisticsBuffered(m.registry.defaultPercentiles)
}

type aggregativeStatisticsBuffered struct {
	aggregativeBuffer

	defaultPercentiles []float64
	tickID             uint64
}

func (s *aggregativeStatisticsBuffered) getPercentile(percentile float64) float64 {
	if s.filledSize == 0 {
		return 0
	}
	percentileIdx := int(float64(s.filledSize) * percentile)
	return s.data[percentileIdx]
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
	return &r
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
		r = append(r, &[]float64{s.getPercentile(percentile)}[0])
	}
	s.locker.Unlock()
	return r
}

// GetDefaultPercentiles returns default percentiles and its values.
func (s *aggregativeStatisticsBuffered) GetDefaultPercentiles() ([]float64, []float64) {
	s.locker.Lock()
	defer s.locker.Unlock()

	r := make([]float64, len(s.defaultPercentiles))
	for idx, p := range s.defaultPercentiles {
		r[idx] = s.getPercentile(p)
	}

	return s.defaultPercentiles, r
}

var (
	randIntnPosition uint32
)

func init() {
	// get an adequate starting seed
	for i := 0; i < 100; i++ {
		randIntn(math.MaxUint32)
	}
}

//go:norace
func randIntn(n uint32) uint32 {
	// We don't require atomicity here because corrupted number is good enough for us, too
	randIntnPosition = 3948558707 * (randIntnPosition + 1948560947)
	if n == math.MaxUint32 {
		return randIntnPosition
	}
	return randIntnPosition % n
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
	// That's why here's randIntn(s.tickID) instead of randIntn(bufferSize)
	randIdx := randIntn(uint32(s.tickID))
	if randIdx >= uint32(bufferSize) {
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
	//s.locker.Lock()
	s.considerValue(v)
	//s.locker.Unlock()
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
