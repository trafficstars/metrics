package metrics

import (
	"testing"
)

func BenchmarkSortNative(b *testing.B) {
	initial := newAggregativeBuffer()
	initial.filledSize = 1000
	for idx, _ := range initial.data {
		initial.data[idx] = float64((282589933 % (idx + 1000)) * 1000 / (idx + 1000))
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s := newAggregativeBuffer()
			copy(s.data[:], initial.data[:])
			s.sortNative()
			s.Release()
		}
	})
}

func BenchmarkSortBuiltin(b *testing.B) {
	initial := newAggregativeBuffer()
	initial.filledSize = 1000
	for idx, _ := range initial.data {
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

type metricCommonAggregativeFastTest struct {
	metricCommonAggregativeFast
}
func (m *metricCommonAggregativeFastTest) Release() {
	return
}
func (m *metricCommonAggregativeFastTest) GetType() Type {
	return TypeGaugeFloat64
}

func BenchmarkConsiderValueFast(b *testing.B) {
	m := metricCommonAggregativeFastTest{}
	m.init(&m, `test`, nil)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.considerValue(1000000)
		}
	})
}

type metricCommonAggregativeShortBufTest struct {
	metricCommonAggregativeShortBuf
}
func (m *metricCommonAggregativeShortBufTest) Release() {
	return
}
func (m *metricCommonAggregativeShortBufTest) GetType() Type {
	return TypeGaugeFloat64
}

func BenchmarkConsiderValueShortBuf(b *testing.B) {
	m := metricCommonAggregativeShortBufTest{}
	m.init(&m, `test`, nil)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.considerValue(1000000)
		}
	})
}
