package metrics

import (
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

const (
	maxConcurrency = 1024
	prebakeMax     = 65536
)

var hiddenTagValue = []byte("hidden")

type Tag interface{}
type Tags map[string]Tag

var (
	tagsPool = sync.Pool{
		New: func() interface{} {
			return Tags{}
		},
	}
)

func NewTags() Tags {
	return tagsPool.Get().(Tags)
}

func (tags Tags) Release() {
	for k, _ := range tags {
		delete(tags, k)
	}
	tagsPool.Put(tags)
}

/*

var (
	trueBytes        = []byte("true")
	falseBytes       = []byte("false")
	nullBytes        = []byte("null")
	unknownTypeBytes = []byte("<unknown_type>")
)

func TagValueToBytes(vI Tag) []byte {
	switch v := vI.(type) {
	case int:
		return CastStringToBytes(strconv.FormatInt(int64(v), 10))
	case uint64:
		return CastStringToBytes(strconv.FormatUint(v, 10))
	case int64:
		return CastStringToBytes(strconv.FormatInt(v, 10))
	case string:
		return CastStringToBytes(strings.Replace(v, ",", "_", -1))
	case bool:
		switch v {
		case true:
			return trueBytes
		case false:
			return falseBytes
		}
	case []byte:
		return v
	case nil:
		return nullBytes
	case interface{ String() string }:
		return CastStringToBytes(strings.Replace(v.String(), ",", "_", -1))
	}

	return unknownTypeBytes
}*/

var prebackedString [prebakeMax * 2]string

func init() {
	for i := -prebakeMax; i < prebakeMax; i++ {
		prebackedString[i+prebakeMax] = strconv.FormatInt(int64(i), 10)
	}
}

func getPrebakedString(v int32) string {
	if v >= prebakeMax || -v <= -prebakeMax {
		return ""
	}
	return prebackedString[v+prebakeMax]
}

func TagValueToString(vI Tag) string {
	switch v := vI.(type) {
	case int:
		r := getPrebakedString(int32(v))
		if len(r) != 0 {
			return r
		}
		return strconv.FormatInt(int64(v), 10)
	case uint64:
		r := getPrebakedString(int32(v))
		if len(r) != 0 {
			return r
		}
		return strconv.FormatUint(v, 10)
	case int64:
		r := getPrebakedString(int32(v))
		if len(r) != 0 {
			return r
		}
		return strconv.FormatInt(v, 10)
	case string:
		return strings.Replace(v, ",", "_", -1)
	case bool:
		switch v {
		case true:
			return "true"
		case false:
			return "false"
		}
	case []byte:
		return string(v)
	case nil:
		return "null"
	case interface{ String() string }:
		return strings.Replace(v.String(), ",", "_", -1)
	}

	return "<unknown_type>"
}

func (tags Tags) ForLogrus(merge logrus.Fields) logrus.Fields {
	fields := logrus.Fields{}
	for k, v := range tags {
		fields[k] = v
	}
	for k, v := range merge {
		fields[k] = v
	}
	return fields
}

func (tags Tags) Copy() Tags {
	cp := Tags{}
	for k, v := range tags {
		cp[k] = v
	}
	return cp
}

func (tags Tags) Keys() (result []string) {
	result = make([]string, 0, len(tags))
	for k, _ := range tags {
		result = append(result, k)
	}
	return
}

func (tags Tags) Get(key string) interface{} {
	return tags[key]
}
func (tags Tags) Set(key string, value interface{}) AnyTags {
	tags[key] = value
	return tags
}
func (tags Tags) Each(fn func(k string, v interface{}) bool) {
	for k, v := range tags {
		if !fn(k, v) {
			break
		}
	}
}

func (tags Tags) ToFastTags() *FastTags {
	keys := tags.Keys()
	sort.Strings(keys)
	r := make(FastTags, 0, len(keys))

	for _, k := range keys {
		intV, _ := toInt64(tags[k])
		r = append(r, FastTag{
			Key:   k,
			StringValue: TagValueToBytes(tags[k]),
			intValue: intV,
		})
	}
	return &r
}
