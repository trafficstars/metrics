package metrics

import (
	"strconv"
	"testing"
)

func benchmarkBubble(b *testing.B, sliceLength int) {
	tags := newFastTags()
	for i := 0; i < sliceLength; i++ {
		tags.Set(strconv.FormatInt(int64(i), 10), i)
	}
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		tags.shuffle()
		b.StartTimer()
		tags.sortBubble()
	}

}

func BenchmarkSortBubble4(b *testing.B) {
	benchmarkBubble(b, 4)
}

func BenchmarkSortBubble8(b *testing.B) {
	benchmarkBubble(b, 8)
}

func BenchmarkSortBubble16(b *testing.B) {
	benchmarkBubble(b, 16)
}

func BenchmarkSortBubble32(b *testing.B) {
	benchmarkBubble(b, 32)
}

func benchmarkSortQuick(b *testing.B, sliceLength int) {
	tags := newFastTags()
	for i := 0; i < sliceLength; i++ {
		tags.Set(strconv.FormatInt(int64(i), 10), i)
	}
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		tags.shuffle()
		b.StartTimer()
		tags.sortQuick()
	}

}

func BenchmarkSortQuick4(b *testing.B) {
	benchmarkSortQuick(b, 4)
}

func BenchmarkSortQuick8(b *testing.B) {
	benchmarkSortQuick(b, 8)
}

func BenchmarkSortQuick16(b *testing.B) {
	benchmarkSortQuick(b, 16)
}

func BenchmarkSortQuick32(b *testing.B) {
	benchmarkSortQuick(b, 32)
}

func benchmarkSortBuiltin(b *testing.B, sliceLength int) {
	tags := newFastTags()
	for i := 0; i < sliceLength; i++ {
		tags.Set(strconv.FormatInt(int64(i), 10), i)
	}
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		tags.shuffle()
		b.StartTimer()
		tags.sortBuiltin()
	}

}

func BenchmarkSortBuiltin4(b *testing.B) {
	benchmarkSortBuiltin(b, 4)
}

func BenchmarkSortBuiltin8(b *testing.B) {
	benchmarkSortBuiltin(b, 8)
}

func BenchmarkSortBuiltin16(b *testing.B) {
	benchmarkSortBuiltin(b, 16)
}

func BenchmarkSortBuiltin32(b *testing.B) {
	benchmarkSortBuiltin(b, 32)
}
