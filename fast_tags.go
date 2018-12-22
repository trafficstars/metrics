package metrics

import (
	"sort"
)

type FastTag struct {
	Key   string
	Value []byte
}

func TagValueToBytes(value Tag) []byte {
	switch v := value.(type) {
	case []byte:
		return v
	default:
		return []byte(TagValueToString(value))
	}
}

func (tag *FastTag) Set(value interface{}) {
	tag.Value = TagValueToBytes(value)
}

type FastTags []FastTag

func (tags FastTags) Len() int {
	return len(tags)
}

func (tags FastTags) Less(i, j int) bool {
	return tags[i].Key < tags[j].Key
}

func (tags FastTags) Swap(i, j int) {
	tags[i], tags[j] = tags[j], tags[i]
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
	return tags[idx].Value
}

func (tags FastTags) Set(key string, value interface{}) {
	idx := tags.findFast(key)
	if idx != -1 {
		tags[idx].Set(value)
	}

	tags = append(tags, FastTag{})
	tags[len(tags)-1].Set(value)
}

func (tags FastTags) Each(fn func(k string, v interface{}) bool) {
	for _, tag := range tags {
		if !fn(tag.Key, tag.Value) {
			break
		}
	}
}

func (tags *FastTags) ToFastTags() *FastTags {
	return tags
}
