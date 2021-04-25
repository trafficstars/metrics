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
	testRegistry := New()
	defer testRegistry.Reset()
	m := &commonAggregativeFlowTest{}
	m.init(testRegistry, m, `test`, nil)
	for i := uint(0); i < bufferSize; i++ {
		m.considerValue(float64(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.considerValue(1000000)
	}
}

func BenchmarkDoSliceFlow(b *testing.B) {
	testRegistry := New()
	defer testRegistry.Reset()
	m := &commonAggregativeFlowTest{}
	m.init(testRegistry, m, `test`, nil)
	for i := uint(0); i < bufferSize; i++ {
		m.considerValue(float64(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.DoSlice()
	}
}

func BenchmarkGetPercentilesFlow(b *testing.B) {
	testRegistry := New()
	defer testRegistry.Reset()
	m := &commonAggregativeFlowTest{}
	m.init(testRegistry, m, `test`, nil)
	for i := uint(0); i < bufferSize; i++ {
		m.considerValue(float64(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.considerValue(float64(i))
		m.GetValuePointers().Total().AggregativeStatistics.GetPercentiles([]float64{0.1, 0.5, 0.9})
	}
}

func BenchmarkGetDefaultPercentilesFlow(b *testing.B) {
	testRegistry := New()
	defer testRegistry.Reset()
	m := &commonAggregativeFlowTest{}
	m.init(testRegistry, m, `test`, nil)
	for i := uint(0); i < bufferSize; i++ {
		m.considerValue(float64(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.considerValue(float64(i))
		m.GetValuePointers().Total().AggregativeStatistics.GetDefaultPercentiles()
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
	m := commonAggregativeBufferedTest{}
	testRegistry := New()
	defer testRegistry.Reset()
	m.init(testRegistry, &m, `test`, nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.considerValue(1000000)
	}
}

func BenchmarkDoSliceBuffered(b *testing.B) {
	m := commonAggregativeBufferedTest{}
	testRegistry := New()
	defer testRegistry.Reset()
	m.init(testRegistry, &m, `test`, nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.DoSlice()
	}
}

func BenchmarkGetPercentilesBuffered(b *testing.B) {
	testRegistry := New()
	defer testRegistry.Reset()
	m := &commonAggregativeBufferedTest{}
	m.init(testRegistry, m, `test`, nil)
	for i := uint(0); i < bufferSize; i++ {
		m.considerValue(float64(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.considerValue(float64(i))
		m.GetValuePointers().Total().AggregativeStatistics.GetPercentiles([]float64{0.01, 0.1, 0.5, 0.9, 0.99})
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
	defer Reset()
	testRegistry := New()
	m := commonAggregativeSimpleTest{}
	m.init(testRegistry, &m, `test`, nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.considerValue(1000000)
	}
}
