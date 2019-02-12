package metrics

import (
	"math/rand"
)

const (
	// If this value is too large then it will be required too many events
	// per second to calculate percentile correctly (Per50, Per99 etc).
	// If this value is too small then the percentile will be calculated
	// not accurate.
	iterationsRequiredPerSecond = 20
)

type metricCommonAggregativeFlow struct {
	metricCommonAggregative
}

func (m *metricCommonAggregativeFlow) init(parent Metric, key string, tags AnyTags) {
	m.metricCommonAggregative.init(parent, key, tags)
	m.data.Current.AggregativeStatistics = newAggregativeStatisticsFlow()
	m.data.Last.AggregativeStatistics = newAggregativeStatisticsFlow()
	m.data.Total.AggregativeStatistics = newAggregativeStatisticsFlow()
}

func (m *metricCommonAggregativeFlow) NewAggregativeStatistics() AggregativeStatistics {
	return newAggregativeStatisticsFlow()
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

type AggregativeStatisticsFlow struct {
	tickID uint64

	Per1  AtomicFloat64Ptr
	Per10 AtomicFloat64Ptr
	Per50 AtomicFloat64Ptr
	Per90 AtomicFloat64Ptr
	Per99 AtomicFloat64Ptr
}

func (s *AggregativeStatisticsFlow) GetPercentile(percentile float64) *float64 {
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

func (s *AggregativeStatisticsFlow) GetPercentiles(percentiles []float64) []*float64 {
	r := make([]*float64, 0, len(percentiles))
	for _, percentile := range percentiles {
		r = append(r, s.GetPercentile(percentile))
	}
	return r
}

// ConsiderValue should be called only for locked items
func (s *AggregativeStatisticsFlow) ConsiderValue(v float64) {
	s.tickID++

	if s.tickID == 1 {
		s.Per1.SetFast(v)
		s.Per10.SetFast(v)
		s.Per50.SetFast(v)
		s.Per90.SetFast(v)
		s.Per99.SetFast(v)
		return
	}

	s.Per1.SetFast(guessPercentile(s.Per1.GetFast(), v, s.tickID, 0.01))
	s.Per10.SetFast(guessPercentile(s.Per10.GetFast(), v, s.tickID, 0.1))
	s.Per50.SetFast(guessPercentile(s.Per50.GetFast(), v, s.tickID, 0.5))
	s.Per90.SetFast(guessPercentile(s.Per90.GetFast(), v, s.tickID, 0.9))
	s.Per99.SetFast(guessPercentile(s.Per99.GetFast(), v, s.tickID, 0.99))
}

func (s *AggregativeStatisticsFlow) Set(value float64) {
	s.Per1.Set(value)
	s.Per10.Set(value)
	s.Per50.Set(value)
	s.Per90.Set(value)
	s.Per99.Set(value)
}

func (s *AggregativeStatisticsFlow) MergeStatistics(oldSI AggregativeStatistics) {
	if oldSI == nil {
		return
	}
	oldS := oldSI.(*AggregativeStatisticsFlow)

	s.Per1.SetFast((s.Per1.GetFast()*float64(s.tickID) + oldS.Per1.GetFast()*float64(oldS.tickID)) / float64(s.tickID+oldS.tickID))
	s.Per10.SetFast((s.Per10.GetFast()*float64(s.tickID) + oldS.Per10.GetFast()*float64(oldS.tickID)) / float64(s.tickID+oldS.tickID))
	s.Per50.SetFast((s.Per50.GetFast()*float64(s.tickID) + oldS.Per50.GetFast()*float64(oldS.tickID)) / float64(s.tickID+oldS.tickID))
	s.Per90.SetFast((s.Per90.GetFast()*float64(s.tickID) + oldS.Per90.GetFast()*float64(oldS.tickID)) / float64(s.tickID+oldS.tickID))
	s.Per99.SetFast((s.Per99.GetFast()*float64(s.tickID) + oldS.Per99.GetFast()*float64(oldS.tickID)) / float64(s.tickID+oldS.tickID))

	s.tickID += oldS.tickID
}
