package metrics

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testTags = Tags{
		"tag0":           0,
		"tag1":           1,
		"success":        true,
		"hello":          "world",
		"service":        "rotator",
		"server":         "idk",
		"worker_id":      -1,
		"defaultTagBool": true,
	}
)

func initDefaultTags() {
	defaultTags = *Tags{
		"defaultTag0":       0,
		"defaultTagString":  "string",
		"defaultTagBool":    false,
		"defaultOneMoreTag": nil,
	}.ToFastTags()
}

func BenchmarkList(b *testing.B) {
	initDefaultTags()
	tags := Tags{
		"tag0":       0,
		"tagString":  "string",
		"tagBool":    false,
		"oneMoreTag": nil,
	}
	for i := 0; i < 10000; i++ {
		tags["value"] = i
		CreateOrGetWorkerGauge(`test_metric`, tags)
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
			buf := generateStorageKey("", "test", nil)
			if buf != nil {
				buf.Unlock()
			}
		}
	})
}

func BenchmarkGet(b *testing.B) {
	initDefaultTags()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Get("", "test", nil)
		}
	})
}

func BenchmarkRegistry(b *testing.B) {
	initDefaultTags()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			CreateOrGetWorkerGauge("", nil)
		}
	})
}

func BenchmarkRegistryReal(b *testing.B) {
	initDefaultTags()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			CreateOrGetWorkerGauge("test_key", testTags)
		}
	})
}

func BenchmarkRegistryReal_withHiddenTag(b *testing.B) {
	SetHiddenTags([]string{`success`})
	initDefaultTags()
	testTagsFast := testTags.ToFastTags()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			CreateOrGetWorkerGauge("test_key", testTagsFast)
		}
	})
	SetHiddenTags(nil)
}

func BenchmarkRegistryReal_FastTags(b *testing.B) {
	initDefaultTags()
	testTagsFast := testTags.ToFastTags()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			CreateOrGetWorkerGauge("test_key", testTagsFast)
		}
	})
}

func BenchmarkTagsString(b *testing.B) {
	initDefaultTags()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := generateStorageKey("", "testKey", testTags)
			buf.Unlock()
		}
	})
}

func BenchmarkTagsFastString(b *testing.B) {
	initDefaultTags()
	testTagsFast := testTags.ToFastTags()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := generateStorageKey("", "testKey", testTagsFast)
			buf.Unlock()
		}
	})
}

func TestGC(t *testing.T) {
	var memstats, cleanedMemstats runtime.MemStats
	goroutinesCount := runtime.NumGoroutine()
	runtime.GC()
	runtime.ReadMemStats(&memstats)
	metric := CreateOrGetWorkerGauge(`test_metric`, nil)
	newGoroutinesCount := runtime.NumGoroutine()
	metric.Stop()
	GC()
	runtime.GC()
	runtime.ReadMemStats(&cleanedMemstats)
	cleanedGoroutinesCount := runtime.NumGoroutine()
	assert.Equal(t, goroutinesCount+1, newGoroutinesCount)
	assert.Equal(t, cleanedGoroutinesCount, goroutinesCount)
	//assert.Equal(t, memstats.HeapInuse, cleanedMemstats.HeapInuse)
}

func TestRegistry(t *testing.T) {
	defaultTags = *Tags{
		"datacenter": "EU",
		"hostcode":   "999",
		"hostname":   "e0df6242fcbf",
		"service":    "rotator",
	}.ToFastTags()

	tags := Tags{
		"code":      400,
		"format_id": "unknown",
		"network":   "unknown",
	}

	tags0 := tags.Copy()
	tags0["key"] = "dsp.bid"
	CreateOrGetWorkerGauge(`requests`, tags0)

	tags1 := tags.Copy()
	tags1["key"] = "dsp.bid.tjnative"
	CreateOrGetWorkerGauge(`requests`, tags1)

	assert.Equal(t, "dsp.bid", Get(MetricTypeGauge, `requests`, tags0).tags["key"])
	assert.Equal(t, "dsp.bid.tjnative", Get(MetricTypeGauge, `requests`, tags1).tags["key"])
}

func TestTagsString(t *testing.T) {
	initDefaultTags()
	{
		buf := generateStorageKey("", "testKey", testTags)
		assert.Equal(t, `testKey,defaultOneMoreTag=null,defaultTag0=0,defaultTagBool=false,defaultTagString=string,hello=world,server=idk,service=rotator,success=true,tag0=0,tag1=1,worker_id=-1`, buf.result.String())
		buf.Unlock()
	}

	{
		SetHiddenTags([]string{`someHiddenTag`})
		assert.Equal(t, `someHiddenTag`, GetHiddenTags()[0])

		buf := generateStorageKey("", "testKey", Tags{"someHiddenTag": true})
		assert.Equal(t, `testKey,defaultOneMoreTag=null,defaultTag0=0,defaultTagBool=false,defaultTagString=string,someHiddenTag=hidden`, buf.result.String())
		buf.Unlock()
	}
}
