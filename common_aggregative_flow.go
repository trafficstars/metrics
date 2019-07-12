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

func (m *commonAggregativeFlow) init(parent Metric, key string, tags AnyTags) {
	m.commonAggregative.init(parent, key, tags)
}

// NewAggregativeStatistics returns a "Flow" (see "Flow" in README.md) implementation of AggregativeStatistics.
func (m *commonAggregativeFlow) NewAggregativeStatistics() AggregativeStatistics {
	return newAggregativeStatisticsFlow()
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
		inertness *= float64(perc)
	} else {
		inertness *= float64(1 - perc)
	}
	return (curValue*inertness + newValue) / (inertness + 1)
}

type aggregativeStatisticsFlow struct {
	tickID uint64

	locker Spinlock

	Per1  float64
	Per10 float64
	Per50 float64
	Per90 float64
	Per99 float64
}

// GetPercentile returns a percentile value for a given percentile (see https://en.wikipedia.org/wiki/Percentile).
//
// It returns nil if the percentile is not from the list: 0.01, 0.1, 0.5, 0.9, 0.99.
func (s *aggregativeStatisticsFlow) GetPercentile(percentile float64) *float64 {
	if s == nil {
		return nil
	}
	switch percentile {
	case 0.01:
		return &s.Per1
	case 0.1:
		return &s.Per10
	case 0.5:
		return &s.Per50
	case 0.9:
		return &s.Per90
	case 0.99:
		return &s.Per99
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
		s.Per1 = v
		s.Per10 = v
		s.Per50 = v
		s.Per90 = v
		s.Per99 = v
		return
	}

	s.Per1 = guessPercentileValue(s.Per1, v, s.tickID, 0.01)
	s.Per10 = guessPercentileValue(s.Per10, v, s.tickID, 0.1)
	s.Per50 = guessPercentileValue(s.Per50, v, s.tickID, 0.5)
	s.Per90 = guessPercentileValue(s.Per90, v, s.tickID, 0.9)
	s.Per99 = guessPercentileValue(s.Per99, v, s.tickID, 0.99)
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
	s.Per1 = value
	s.Per10 = value
	s.Per50 = value
	s.Per90 = value
	s.Per99 = value
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

	s.Per1 = (s.Per1*float64(s.tickID) + oldS.Per1*float64(oldS.tickID)) / float64(s.tickID+oldS.tickID)
	s.Per10 = (s.Per10*float64(s.tickID) + oldS.Per10*float64(oldS.tickID)) / float64(s.tickID+oldS.tickID)
	s.Per50 = (s.Per50*float64(s.tickID) + oldS.Per50*float64(oldS.tickID)) / float64(s.tickID+oldS.tickID)
	s.Per90 = (s.Per90*float64(s.tickID) + oldS.Per90*float64(oldS.tickID)) / float64(s.tickID+oldS.tickID)
	s.Per99 = (s.Per99*float64(s.tickID) + oldS.Per99*float64(oldS.tickID)) / float64(s.tickID+oldS.tickID)

	s.tickID += oldS.tickID
}
