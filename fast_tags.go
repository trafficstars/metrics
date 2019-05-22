package metrics

import (
	"sort"
	"sync"
)

var (
	disableFastTags = false
)

// FastTag is an element of FastTags (see "FastTags")
type FastTag struct {
	Key         string
	StringValue string

	// The main value is the StringValue. "intValue" exists only for optimizations
	intValue      int64
	intValueIsSet bool

	// This check was temporary added only for debugging (to locate and fix one bug in this module)
	// It negatively affects the performance and should be removed in future (like in >= 2020 year)
	// A marker if the tag is already in-use and cannot be returned from a pool with Get()
	isInUse bool
	// EndOf the debugging check
}

var (
	fastTagPool = sync.Pool{
		New: func() interface{} {
			return &FastTag{}
		},
	}
)

func SetDisableFastTags(newDisableFastTags bool) {
	disableFastTags = newDisableFastTags
}

func newFastTag() *FastTag {
	tag := fastTagPool.Get().(*FastTag)

	// This check was temporary added only for debugging (to locate and fix one bug in this module)
	// It negatively affects the performance and should be removed in future (like in >= 2020 year)
	if tag.isInUse {
		panic(`A attempt to acquire a busy FastTag`)
	}
	tag.isInUse = true
	// EndOf the debugging check

	return tag
}

// Release puts the FastTag back into the pool. The pool is use for memory reuse (to do not GC and reallocate
// memory).
//
// This method is supposed to be used to internal needs, only.
func (tag *FastTag) Release() {

	// This check was temporary added only for debugging (to locate and fix one bug in this module)
	// It negatively affects the performance and should be removed in future (like in >= 2020 year)
	if !tag.isInUse {
		panic(`An attempt to release a (already) released FastTag`)
	}
	tag.isInUse = false
	// EndOf the debugging check

	tag.intValueIsSet = false
	fastTagPool.Put(tag)
}

/*func TagValueToBytes(value Tag) []byte {
	switch v := value.(type) {
	case []byte:
		return v
	default:
		return []byte(TagValueToString(value))
	}
}*/

// GetValue returns the value of the tag. It returns it as an int64 if the value could be represented as an integer, or
// as a string if it cannot be represented as an integer.
func (tag *FastTag) GetValue() interface{} {
	if tag.intValueIsSet {
		return tag.intValue
	}

	return tag.StringValue
}

// Set sets the key and the value.
//
// The value will be stored as a string and, if possible, as an int64.
func (tag *FastTag) Set(key string, value interface{}) {
	tag.Key = key
	tag.StringValue = TagValueToString(value)
	if intV, ok := toInt64(value); ok {
		tag.intValue = intV
		tag.intValueIsSet = true
	}
}

type FastTags struct {
	Slice []*FastTag

	isInUse bool
}

var (
	fastTagsPool = sync.Pool{
		New: func() interface{} {
			return &FastTags{}
		},
	}
)

func newFastTags() *FastTags {
	tags := fastTagsPool.Get().(*FastTags)
	if tags.isInUse {
		panic(`An attempt to acquire a busy FastTags`)
	}
	tags.isInUse = true
	return tags
}

// NewFastTags returns an implementation of AnyTags with a full memory reuse support.
//
// This implementation is supposed to be used if it's required to reduce a pressure on GC (see "GCCPUFraction",
// https://golang.org/pkg/runtime/#MemStats).
//
// It could be required if there's a metric that is retrieved very often and it's required to reduce CPU utilization.
//
// See "Tags" in README.md
func NewFastTags() AnyTags {
	if disableFastTags {
		return NewTags()
	}

	return newFastTags()
}

// Release clears the tags and puts the them back into the pool. It's required for memory reusing.
//
// See "Tags" in README.md
func (tags *FastTags) Release() {
	if !memoryReuse {
		return
	}
	if tags == nil {
		return
	}
	if !tags.isInUse {
		panic(`An attempt to release a released FastTags`)
	}
	tags.isInUse = false
	for _, tag := range tags.Slice {
		tag.Release()
	}
	tags.Slice = tags.Slice[:0]
	fastTagsPool.Put(tags)
}

// Len returns the amount/count of tags
func (tags *FastTags) Len() int {
	if tags == nil {
		return 0
	}
	return len(tags.Slice)
}

// Less returns if the Key of the tag by index "i" is less (strings comparison) than the Key of the tag by index "j".
func (tags *FastTags) Less(i, j int) bool {
	return tags.Slice[i].Key < tags.Slice[j].Key
}

// Swap just swaps tags by indexes "i" and "j"
func (tags *FastTags) Swap(i, j int) {
	tags.Slice[i], tags.Slice[j] = tags.Slice[j], tags.Slice[i]
}

// Sort sorts tags by keys (using Swap, Less and Len)
func (tags *FastTags) Sort() {
	// We use our-own implementation of sorts without interfaces which  doesn't require a memory allocation
	if len(tags.Slice) < 16 {
		// On a small slice "Bubble" is really not that bad (k*O(n*n) with a small "k").
		tags.sortBubble()
	} else {
		// TODO: May be it's not required to reimplement QuickSort if some of magic comments
		// (https://github.com/xaionaro-go/hackery) may be used to safely prevent memory allocation.
		tags.sortQuick()
	}
}

// findStupid finds the tag with key "key" using a full scan
//
// It returns the index of the found tag. If the tag wasn't found then -1 will be returned.
func (tags *FastTags) findStupid(key string) int {
	for idx, tag := range tags.Slice {
		if tag.Key == key {
			return idx
		}
	}
	return -1
}

// findFast finds the tag with key "key" using a binary search.
//
// It returns the index of the found tag. If the tag wasn't found then -1 will be returned.
//
// Tags should be sorted before use this method.
func (tags *FastTags) findFast(key string) int {
	l := len(tags.Slice)
	idx := sort.Search(l, func(i int) bool {
		return tags.Slice[i].Key >= key
	})

	if idx < 0 || idx >= l {
		return -1
	}

	if tags.Slice[idx].Key != key {
		return -1
	}

	return idx
}

// IsSet returns true if there's a tag with key "key", otherwise -- false.
func (tags *FastTags) IsSet(key string) bool {
	return tags.findStupid(key) != -1
}

// Get returns the value of the tag with key "key".
//
// If there's no such tag then nil will be returned.
func (tags *FastTags) Get(key string) interface{} {
	idx := tags.findStupid(key)
	if idx == -1 {
		return nil
	}

	return tags.Slice[idx].GetValue()
}

// Set sets the value of the tag with key "key" to "value". If there's no such tag then creates it and sets the value.
func (tags *FastTags) Set(key string, value interface{}) AnyTags {
	idx := tags.findStupid(key)
	if idx != -1 {
		tags.Slice[idx].Set(key, value)
		return tags
	}

	newTag := newFastTag()
	newTag.Set(key, value)
	tags.Slice = append(tags.Slice, newTag)
	return tags
}

// Each is a function to call function "fn" for each tag. A key and a value of a tag will be passed as "k" and "v"
// arguments, accordingly.
func (tags *FastTags) Each(fn func(k string, v interface{}) bool) {
	if tags == nil {
		return
	}
	for _, tag := range tags.Slice {
		if !fn(tag.Key, tag.GetValue()) {
			break
		}
	}
}

// ToFastTags does nothing and returns the same tags.
//
// This method is required to implement interface "AnyTags".
func (tags *FastTags) ToFastTags() *FastTags {
	return tags
}

// ToMap returns tags as an map of tag keys to tag values ("map[string]interface{}").
//
// Any maps passed as an argument will overwrite values of the resulting map.
func (tags *FastTags) ToMap(fieldMaps ...map[string]interface{}) map[string]interface{} {
	if tags == nil {
		return nil
	}
	fields := map[string]interface{}{}
	if tags != nil {
		for _, tag := range tags.Slice {
			fields[tag.Key] = tag.GetValue()
		}
	}
	for _, fieldMap := range fieldMaps {
		for k, v := range fieldMap {
			fields[k] = v
		}
	}
	return fields
}

// String returns tags as a string compatible with StatsD format of tags.
func (tags *FastTags) String() string {
	buf := newBytesBuffer()
	tags.WriteAsString(buf)
	result := buf.String()
	buf.Release()
	return result
}

// WriteAsString writes tags in StatsD format through the WriteStringer (passed as the argument)
func (tags *FastTags) WriteAsString(writeStringer interface{ WriteString(string) (int, error) }) {
	if tags == nil {
		return
	}

	tags.Sort()
	tagsCount := 0
	for _, tag := range tags.Slice {
		if defaultTags.IsSet(tag.Key) {
			continue
		}
		if tagsCount != 0 {
			writeStringer.WriteString(`,`)
		}
		writeStringer.WriteString(tag.Key)
		writeStringer.WriteString(`=`)
		writeStringer.WriteString(tag.StringValue)
		tagsCount++
	}
}
