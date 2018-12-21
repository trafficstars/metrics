package metrics

import (
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unsafe"

	"github.com/sirupsen/logrus"
)

const (
	maxConcurrency = 1024
)

type Tag interface{}
type Tags map[string]Tag

func CastStringToBytes(str string) []byte {
	hdr := *(*reflect.StringHeader)(unsafe.Pointer(&str))
	return *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: hdr.Data,
		Len:  hdr.Len,
		Cap:  hdr.Len,
	}))
}

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

func (tags Tags) Set(key string, value interface{}) {
	tags[key] = value
}
func (tags Tags) Each(fn func(k string, v interface{}) bool) {
	for k, v := range tags {
		if !fn(k, v) {
			break
		}
	}
}

func (tags Tags) ToFastTags() FastTags {
	keys := tags.Keys()
	sort.Strings(keys)
	r := make(FastTags, 0, len(keys))

	for _, k := range keys {
		r = append(r, FastTag{
			Key:   k,
			Value: TagValueToBytes(tags[k]),
		})
	}
	return r
}
