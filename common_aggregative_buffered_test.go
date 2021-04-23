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

type commonAggregativeFlowTest struct {
	commonAggregativeFlow
}

func (m *commonAggregativeFlowTest) Release() {
	return
}
func (m *commonAggregativeFlowTest) GetType() Type {
	return TypeTimingFlow
}

func BenchmarkConsiderValueFlow(b *testing.B) {
	Reset()
	m := commonAggregativeFlowTest{}
	m.init(&m, `test`, nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.considerValue(1000000)
	}
}

func BenchmarkDoSliceFlow(b *testing.B) {
	Reset()
	m := commonAggregativeFlowTest{}
	m.init(&m, `test`, nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.DoSlice()
	}
}

var (
	testMFlow = &commonAggregativeFlowTest{}
)

func BenchmarkGetPercentilesFlow(b *testing.B) {
	Reset()
	m := testMFlow
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.considerValue(float64(i))
		m.GetValuePointers().Total.AggregativeStatistics.GetPercentiles([]float64{0.5, 0.9, 0.99})
	}
}

type commonAggregativeBufferedTest struct {
	commonAggregativeBuffered
}

func (m *commonAggregativeBufferedTest) Release() {
	return
}
func (m *commonAggregativeBufferedTest) GetType() Type {
	return TypeTimingBuffered
}

func BenchmarkConsiderValueBuffered(b *testing.B) {
	Reset()
	m := commonAggregativeBufferedTest{}
	m.init(&m, `test`, nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.considerValue(1000000)
	}
}

func BenchmarkDoSliceBuffered(b *testing.B) {
	Reset()
	m := commonAggregativeBufferedTest{}
	m.init(&m, `test`, nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.DoSlice()
	}
}

var (
	testMBuffered = &commonAggregativeBufferedTest{}
)

func BenchmarkGetPercentilesBuffered(b *testing.B) {
	Reset()
	m := testMBuffered
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.considerValue(float64(i))
		m.GetValuePointers().Total.AggregativeStatistics.GetPercentiles([]float64{0.5, 0.9, 0.99})
	}
}

type commonAggregativeSimpleTest struct {
	commonAggregativeSimple
}

func (m *commonAggregativeSimpleTest) Release() {
	return
}
func (m *commonAggregativeSimpleTest) GetType() Type {
	return TypeTimingSimple
}

func BenchmarkConsiderValueSimple(b *testing.B) {
	Reset()
	m := commonAggregativeSimpleTest{}
	m.init(&m, `test`, nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.considerValue(1000000)
	}
}

var (
	testMSimple = &commonAggregativeSimpleTest{}
)

func init() {
	{
		m := testMBuffered
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

	{
		m := testMSimple
		m.init(m, `test`, nil)
		for i := uint(0); i < bufferSize; i++ {
			m.considerValue(float64(i))
		}
	}
}
