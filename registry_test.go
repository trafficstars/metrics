package metrics

import (
	"fmt"
	"runtime"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testTags = Tags{
		`tag0`:           0,
		`tag1`:           1,
		`success`:        true,
		`hello`:          `world`,
		`service`:        `rotator`,
		`server`:         `idk`,
		`worker_id`:      -1,
		`defaultTagBool`: true,
	}
)

func initDefaultTags() {
	defaultTags = *Tags{
		`defaultTag0`:       0,
		`defaultTagString`:  `string`,
		`defaultTagBool`:    false,
		`defaultOneMoreTag`: nil,
	}.ToFastTags()
}

func BenchmarkList(b *testing.B) {
	initDefaultTags()
	tags := Tags{
		`tag0`:       0,
		`tagString`:  `string`,
		`tagBool`:    false,
		`oneMoreTag`: nil,
	}
	for i := 0; i < 10000; i++ {
		tags[`value`] = i
		GaugeInt64(`test_metric`, tags)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		List()
	}
}

func BenchmarkGenerateStorageKey(b *testing.B) {
	initDefaultTags()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			generateStorageKey(TypeCount, `test`, nil)
		}
	})
}

func BenchmarkGet(b *testing.B) {
	initDefaultTags()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Get(TypeCount, `test`, nil)
		}
	})
}

func BenchmarkRegistry(b *testing.B) {
	initDefaultTags()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			GaugeInt64(``, nil)
		}
	})
}

func BenchmarkRegistryReal(b *testing.B) {
	initDefaultTags()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			GaugeInt64(`test_key`, testTags)
		}
	})
}
func BenchmarkAddToRegistryReal(b *testing.B) {
	var i uint64
	testTags[`i`] = &i
	initDefaultTags()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			atomic.StoreUint64(&i, 1)
			GaugeInt64(`test_key`, testTags)
		}
	})
}

func BenchmarkRegistryRealReal_lazy(b *testing.B) {
	SetHiddenTags(HiddenTags{HiddenTag{`success`, nil}, HiddenTag{`campaign_id`, ExceptValues{1}}})
	initDefaultTags()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			testTags := Tags{
				`string`:         `value`,
				`success`:        true,
				`worker_id`:      -1,
				`defaultTagBool`: true,
			}
			GaugeInt64(`test_key`, testTags)
		}
	})
	SetHiddenTags(nil)
}

func BenchmarkRegistryRealReal_normal(b *testing.B) {
	SetHiddenTags(HiddenTags{HiddenTag{`success`, nil}, HiddenTag{`campaign_id`, ExceptValues{1}}})
	initDefaultTags()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			testTags := NewTags()
			testTags[`string`] = `value`
			testTags[`success`] = true
			testTags[`worker_id`] = -1
			testTags[`defaultTagBool`] = true
			GaugeInt64(`test_key`, testTags)
			testTags.Release()
		}
	})
	SetHiddenTags(nil)
}

func BenchmarkRegistryRealReal_FastTags_withHiddenTag(b *testing.B) {
	SetHiddenTags(HiddenTags{HiddenTag{`success`, nil}, HiddenTag{`campaign_id`, ExceptValues{1}}})
	initDefaultTags()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			testTags := NewFastTags().
				Set(`string`, `value`).
				Set(`success`, true).
				Set(`worker_id`, -1).
				Set(`defaultTagBool`, true)
			GaugeInt64(`test_key`, testTags)
			testTags.Release()
		}
	})
	SetHiddenTags(nil)
}

func BenchmarkRegistryRealReal_FastTags(b *testing.B) {
	initDefaultTags()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			testTags := NewFastTags().
				Set(`string`, `value`).
				Set(`success`, true).
				Set(`worker_id`, -1).
				Set(`defaultTagBool`, true)
			GaugeInt64(`test_key`, testTags)
			testTags.Release()
		}
	})
}

func BenchmarkFastTag_Set(b *testing.B) {
	tag := newFastTag()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tag.Set(`a`, -1)
	}
	tag.Release()
}

func BenchmarkGenKeyTags(b *testing.B) {
	initDefaultTags()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			generateStorageKey(TypeGaugeInt64, `testKey`, testTags).Release()
		}
	})
}

func BenchmarkGenKeyFastTags(b *testing.B) {
	initDefaultTags()
	testTagsFast := testTags.ToFastTags()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			generateStorageKey(TypeGaugeInt64, `testKey`, testTagsFast).Release()
		}
	})
}

func TestGet(t *testing.T) {
	SetDefaultTags(Tags{
		"service":    "pixel",
		"datacenter": "DC_NAME",
		"hostname":   "hostname",
		"hostcode":   "hostcode",
	})

	tags := NewFastTags().Set("key", "TestTag").
		Set("result", "unknown").
		Set("format_id", "TestTag").
		Set("is_priv", true)

	Count(`TestGet`, tags)
	tags.Release()

	GC()

	tags = NewFastTags().Set("key", "TestTag").
		Set("result", "unknown").
		Set("format_id", "TestTag").
		Set("is_priv", true)

	m := registry.Get(TypeCount, `TestGet`, tags)
	if m == nil {
		considerHiddenTags(tags)
		fmt.Println("Key:", generateStorageKey(TypeCount, `TestGet`, tags).buf.String())
		for _, key := range registry.storage.Keys() {
			metric, _ := registry.storage.GetByBytes(key.([]byte))
			if metric == nil {
				continue
			}
			fmt.Println("The list:", string(key.([]byte)), metric)
		}
	}
	tags.Release()
	assert.NotNil(t, m)
}

func TestGC(t *testing.T) {
	var memstats, cleanedMemstats runtime.MemStats
	goroutinesCount := runtime.NumGoroutine()
	runtime.GC()
	runtime.ReadMemStats(&memstats)
	metric := GaugeInt64(`test_metric`, nil)
	//newGoroutinesCount := runtime.NumGoroutine()
	metric.Stop()
	GC()
	runtime.GC()
	runtime.ReadMemStats(&cleanedMemstats)
	cleanedGoroutinesCount := runtime.NumGoroutine()
	//assert.Equal(t, goroutinesCount+1, newGoroutinesCount)
	assert.Equal(t, cleanedGoroutinesCount, goroutinesCount)
	//assert.Equal(t, memstats.HeapInuse, cleanedMemstats.HeapInuse)
}

func TestRegistry(t *testing.T) {
	{
		metric0 := GaugeInt64(``, nil)
		metric1 := GaugeInt64(``, nil)
		assert.Equal(t, metric0, metric1)
	}

	{
		metric0 := GaugeInt64(`non_empty_key`, nil)
		metric1 := GaugeInt64(`non_empty_key`, nil)
		assert.Equal(t, metric0, metric1)
	}

	{
		metric0 := GaugeInt64(`non_empty_key`, Tags{"some_key": "some_value"})
		metric1 := GaugeInt64(`non_empty_key`, NewFastTags().Set("some_key", "some_value"))
		assert.Equal(t, metric0, metric1)
	}

	defaultTags = *Tags{
		`datacenter`: `EU`,
		`hostcode`:   `999`,
		`hostname`:   `e0df6242fcbf`,
		`service`:    `rotator`,
	}.ToFastTags()

	tags := Tags{
		`code`:      400,
		`format_id`: `unknown`,
		`network`:   `unknown`,
	}

	tags0 := tags.Copy()
	tags0[`key`] = `dsp.bid`
	GaugeInt64(`requests`, tags0)

	tags1 := tags.Copy()
	tags1[`key`] = `dsp.bid.tjnative`
	GaugeInt64(`requests`, tags1)

	metric0 := Get(TypeGaugeInt64, `requests`, tags0)
	if !assert.Equal(t, `dsp.bid`, metric0.GetTag(`key`)) {
		t.Errorf("tags: %v", metric0.GetTags())
	}
	metric1 := Get(TypeGaugeInt64, `requests`, tags1)
	if !assert.Equal(t, `dsp.bid.tjnative`, metric1.GetTag(`key`)) {
		t.Errorf("tags: %v", metric1.GetTags())
	}
}
