package metrics

import (
	"testing"
)

func BenchmarkSortBuiltin(b *testing.B) {
	initial := newAggregativeBuffer()
	initial.filledSize = 1000
	for idx := range initial.data {
		initial.data[idx] = float64((282589933 % (idx + 1000)) * 1000 / (idx + 1000))
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s := newAggregativeBuffer()
			copy(s.data[:], initial.data[:])
			s.sortBuiltin()
			s.Release()
		}
	})
}

type metricCommonAggregativeFlowTest struct {
	metricCommonAggregativeFlow
}

func (m *metricCommonAggregativeFlowTest) Release() {
	return
}
func (m *metricCommonAggregativeFlowTest) GetType() Type {
	return TypeTimingFlow
}

func BenchmarkConsiderValueFlow(b *testing.B) {
	Reset()
	m := metricCommonAggregativeFlowTest{}
	m.init(&m, `test`, nil)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.considerValue(1000000)
		}
	})
}

func BenchmarkDoSliceFlow(b *testing.B) {
	Reset()
	m := metricCommonAggregativeFlowTest{}
	m.init(&m, `test`, nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.DoSlice()
	}
}

var (
	testMFlow = &metricCommonAggregativeFlowTest{}
)

func BenchmarkGetPercentilesFlow(b *testing.B) {
	Reset()
	m := testMFlow
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.considerValue(float64(i))
		m.GetValuePointers().Total.AggregativeStatistics.GetPercentiles([]float64{0.01, 0.1, 0.5, 0.9, 0.99})
	}
}

type metricCommonAggregativeShortBufTest struct {
	metricCommonAggregativeShortBuf
}

func (m *metricCommonAggregativeShortBufTest) Release() {
	return
}
func (m *metricCommonAggregativeShortBufTest) GetType() Type {
	return TypeTimingBuffered
}

func BenchmarkConsiderValueShortBuf(b *testing.B) {
	Reset()
	m := metricCommonAggregativeShortBufTest{}
	m.init(&m, `test`, nil)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.considerValue(1000000)
		}
	})
}

func BenchmarkDoSliceShortBuf(b *testing.B) {
	Reset()
	m := metricCommonAggregativeShortBufTest{}
	m.init(&m, `test`, nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.DoSlice()
	}
}

var (
	testMShortBuf = &metricCommonAggregativeShortBufTest{}
)

func BenchmarkGetPercentilesShortBuf(b *testing.B) {
	Reset()
	m := testMShortBuf
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.considerValue(float64(i))
		m.GetValuePointers().Total.AggregativeStatistics.GetPercentiles([]float64{0.01, 0.1, 0.5, 0.9, 0.99})
	}
}

func init() {
	{
		m := testMShortBuf
		m.init(m, `test`, nil)
		for i := uint(0); i < bufferSize; i++ {
			m.considerValue(float64(i))
		}
	}

	{
		m := testMFlow
		m.init(m, `test`, nil)
		for i := uint(0); i < bufferSize; i++ {
			m.considerValue(float64(i))
		}
	}
}
