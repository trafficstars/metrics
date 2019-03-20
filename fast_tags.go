package metrics

import (
	"sort"
	"sync"
)

type FastTag struct {
	Key         string
	StringValue string

	// The main value is the StringValue. "intValue" exists only for optimizations
	intValue      int64
	intValueIsSet bool
}

var (
	fastTagPool = sync.Pool{
		New: func() interface{} {
			return &FastTag{}
		},
	}
)

func newFastTag() *FastTag {
	return fastTagPool.Get().(*FastTag)
}

func (tag *FastTag) Release() {
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

func (tag *FastTag) GetValue() interface{} {
	if tag.intValueIsSet {
		return tag.intValue
	}

	return tag.StringValue
}

func (tag *FastTag) Set(key string, value interface{}) {
	tag.Key = key
	tag.StringValue = TagValueToString(value)
	if intV, ok := toInt64(value); ok {
		tag.intValue = intV
		tag.intValueIsSet = true
	}
}

type FastTags []*FastTag

var (
	fastTagsPool = sync.Pool{
		New: func() interface{} {
			return &FastTags{}
		},
	}
)

/*func NewFastTags() *FastTags {
	return fastTagsPool.Get().(*FastTags)
}*/

func NewFastTags() Tags {
	return NewTags()
}

func (tags *FastTags) Release() {
	for _, tag := range *tags {
		tag.Release()
	}
	*tags = (*tags)[:0]
	fastTagsPool.Put(tags)
}

func (tags FastTags) Len() int {
	return len(tags)
}

func (tags FastTags) Less(i, j int) bool {
	return tags[i].Key < tags[j].Key
}

func (tags FastTags) Swap(i, j int) {
	tags[i], tags[j] = tags[j], tags[i]
}

func (tags FastTags) Sort() {
	if len(tags) < 16 {
		tags.sortBubble()
	} else {
		tags.sortQuick()
	}
}

func (tags FastTags) findStupid(key string) int {
	for idx, tag := range tags {
		if tag.Key == key {
			return idx
		}
	}
	return -1
}

func (tags FastTags) findFast(key string) int {
	l := len(tags)
	idx := sort.Search(l, func(i int) bool {
		return tags[i].Key >= key
	})

	if idx < 0 || idx >= l {
		return -1
	}

	if tags[idx].Key != key {
		return -1
	}

	return idx
}

func (tags FastTags) IsSet(key string) bool {
	return tags.findFast(key) != -1
}

func (tags FastTags) Get(key string) interface{} {
	idx := tags.findFast(key)
	if idx == -1 {
		return nil
	}

	return tags[idx].GetValue()
}

func (tags *FastTags) Set(key string, value interface{}) AnyTags {
	idx := tags.findFast(key)
	if idx != -1 {
		(*tags)[idx].Set(key, value)
		return tags
	}

	(*tags) = append((*tags), newFastTag())
	(*tags)[len(*tags)-1].Set(key, value)
	return tags
}

func (tags FastTags) Each(fn func(k string, v interface{}) bool) {
	for _, tag := range tags {
		if !fn(tag.Key, tag.GetValue()) {
			break
		}
	}
}

func (tags *FastTags) ToFastTags() *FastTags {
	return tags
}

func (tags FastTags) ToMap(fieldMaps ...map[string]interface{}) map[string]interface{} {
	fields := map[string]interface{}{}
	if tags != nil {
		for _, tag := range tags {
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
