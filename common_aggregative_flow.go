package metrics

import (
	"math"
)

const (
	// If this value is too large then it will be required too many events
	// per second to calculate percentile correctly (Per50, Per99 etc).
	// If this value is too small then the percentile will be calculated
	// not accurate.
	iterationsRequiredPerSecond = 20
)

type commonAggregativeFlow struct {
	commonAggregative
}

func (m *commonAggregativeFlow) init(r *Registry, parent Metric, key string, tags AnyTags) {
	m.commonAggregative.init(r, parent, key, tags)
}

// NewAggregativeStatistics returns a "Flow" (see "Flow" in README.md) implementation of AggregativeStatistics.
func (m *commonAggregativeFlow) NewAggregativeStatistics() AggregativeStatistics {
	stats := newAggregativeStatisticsFlow()
	stats.percentiles = m.registry.defaultPercentiles
	return stats
}

// guessPercentileValue is a so-so correct way of correcting the percentile value for big amount of events
// (> iterationsRequiredPerSecond)
//
// See "Flow" in README.md
func guessPercentileValue(curValue, newValue float64, count uint64, perc float64) float64 {
	// The more events we received the more precise value we should get,
	// so will should change the value slower if the count is higher
	inertness := float64(count) / iterationsRequiredPerSecond

	// See "How the calculation of percentile values works" in README.md
	requireGreater := float64(randIntn(math.MaxUint32))/float64(math.MaxUint32) > perc

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
		inertness *= perc
	} else {
		inertness *= 1 - perc
	}
	return (curValue*inertness + newValue) / (inertness + 1)
}

type aggregativeStatisticsFlow struct {
	tickID uint64

	locker Spinlock

	percentiles      []float64
	percentileValues [maxPercentileValues]float64
}

// GetPercentile returns a percentile value for a given percentile (see https://en.wikipedia.org/wiki/Percentile).
//
// It returns nil if the percentile is not from the list: 0.01, 0.1, 0.5, 0.9, 0.99.
func (s *aggregativeStatisticsFlow) GetPercentile(percentile float64) *float64 {
	if s == nil {
		return nil
	}

	for idx, p := range s.percentiles {
		if p == percentile {
			return &s.percentileValues[idx]
		}
	}

	return nil
}

// GetPercentiles returns percentile values for a given slice of percentiles.
//
// Returned values are ordered accordingly to the input slice. An element of the returned
// slice is "nil" if the according percentile is not from the list:  0.01, 0.1, 0.5, 0.9, 0.99.
//
// There's no performance profit to prefer either of GetPercentile/GetPercentiles for any case (because it's a "Flow"
// method of percentile calculate), so just use what is more convenient.
func (s *aggregativeStatisticsFlow) GetPercentiles(percentiles []float64) []*float64 {
	if len(percentiles) == 0 {
		return nil
	}
	r := make([]*float64, 0, len(percentiles))
	for _, percentile := range percentiles {
		r = append(r, s.GetPercentile(percentile))
	}
	return r
}

// Note! considerValue should be called only for locked items
func (s *aggregativeStatisticsFlow) considerValue(v float64) {
	s.tickID++

	if s.tickID == 1 {
		for idx := range s.percentiles {
			s.percentileValues[idx] = v
		}
		return
	}

	for idx, p := range s.percentiles {
		s.percentileValues[idx] = guessPercentileValue(s.percentileValues[idx], v, s.tickID, p)
	}
}

// GetDefaultPercentiles returns all percentiles.
func (s *aggregativeStatisticsFlow) GetDefaultPercentiles() ([]float64, []float64) {
	s.locker.Lock()
	defer s.locker.Unlock()

	r := make([]float64, len(s.percentiles))
	copy(r, s.percentileValues[:])
	return s.percentiles, r
}

// ConsiderValue is an analog of Prometheus' observe (see "Aggregative metrics" in README.md)
func (s *aggregativeStatisticsFlow) ConsiderValue(v float64) {
	//s.locker.Lock()
	s.considerValue(v)
	//s.locker.Unlock()
}

// Set resets the statistics and sets only one event with the value passed as the argument,
// so all aggregative values (avg, min, max, ...) will be equal to the value
func (s *aggregativeStatisticsFlow) Set(value float64) {
	s.locker.Lock()
	for idx := range s.percentiles {
		s.percentileValues[idx] = value
	}
	s.locker.Unlock()
}

// MergeStatistics adds statistics of the argument to the own one.
//
// See "Attention!" of "How the calculation of percentile values works" in README.md.
func (s *aggregativeStatisticsFlow) MergeStatistics(oldSI AggregativeStatistics) {
	if oldSI == nil {
		return
	}
	oldS := oldSI.(*aggregativeStatisticsFlow)

	if s.tickID+oldS.tickID == 0 {
		return
	}

	for idx := range s.percentiles {
		s.percentileValues[idx] = (s.percentileValues[idx]*float64(s.tickID) + oldS.percentileValues[idx]*float64(oldS.tickID)) / float64(s.tickID+oldS.tickID)
	}

	s.tickID += oldS.tickID
}
