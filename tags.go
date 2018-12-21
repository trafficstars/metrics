package metrics

import (
	"sort"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	maxConcurrency = 1024
)

type Tag interface{}
type Tags map[string]Tag

func TagValueToString(vI Tag) string {
	switch v := vI.(type) {
	case int:
		return strconv.FormatInt(int64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case int64:
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
			Value: []byte(TagValueToString(tags[k])),
		})
	}
	return r
}
