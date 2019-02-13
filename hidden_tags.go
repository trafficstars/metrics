package metrics

import (
	"sort"
)

type ExceptValues []interface{}
type HiddenTag struct {
	Key          string
	ExceptValues ExceptValues
}
type HiddenTags []HiddenTag

type hiddenTagExceptValue struct {
	Int    int64
	String string
}
type hiddenTagExceptValues []hiddenTagExceptValue

func (v *hiddenTagExceptValue) LessThan(i int64, s string) bool {
	if v.Int < i {
		return true
	}
	if v.Int > i {
		return false
	}
	return v.String < s
}

func (s hiddenTagExceptValues) Len() int {
	return len(s)
}

func (s hiddenTagExceptValues) Less(i, j int) bool {
	return s[i].LessThan(s[j].Int, s[j].String)
}

func (s hiddenTagExceptValues) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s hiddenTagExceptValues) Sort() {
	sort.Sort(s)
}

func (slice hiddenTagExceptValues) Search(i int64, s string) int {
	l := len(slice)
	foundIdx := sort.Search(l, func(idx int) bool {
		return !slice[idx].LessThan(i, s)
	})

	if foundIdx < 0 || foundIdx >= l {
		return -1
	}

	found := &slice[foundIdx]
	if found.Int != i || found.String != s {
		return -1
	}

	return foundIdx
}

type hiddenTagInternal struct {
	Key          string
	exceptValues hiddenTagExceptValues

	exceptIntsCount    int
	exceptStringsCount int
}
type hiddenTagsInternal []hiddenTagInternal

func toInt64(in interface{}) (int64, bool) {
	switch v := in.(type) {
	case uint8:
		return int64(v), true
	case int8:
		return int64(v), true
	case uint16:
		return int64(v), true
	case int16:
		return int64(v), true
	case uint32:
		return int64(v), true
	case int32:
		return int64(v), true
	case uint64:
		return int64(v), true
	case int64:
		return v, true
	case uint:
		return int64(v), true
	case int:
		return int64(v), true
	}

	return 0, false
}

func (rawHiddenTag *HiddenTag) toInternal() *hiddenTagInternal {
	r := &hiddenTagInternal{
		Key: rawHiddenTag.Key,
	}
	for _, rawExceptValue := range rawHiddenTag.ExceptValues {
		value := hiddenTagExceptValue{}
		if intV, ok := toInt64(rawExceptValue); ok {
			value.Int = intV
			r.exceptIntsCount++
		} else {
			value.String = TagValueToString(rawExceptValue)
			r.exceptStringsCount++
		}

		r.exceptValues = append(r.exceptValues, value)
	}

	r.exceptValues.Sort()
	return r
}

func (tag *hiddenTagInternal) HasExceptValues() bool {
	return tag.exceptIntsCount > 0 || tag.exceptStringsCount > 0
}

func (tag *hiddenTagInternal) HasExceptInts() bool {
	return tag.exceptIntsCount > 0
}

func (tag *hiddenTagInternal) HasExceptStrings() bool {
	return tag.exceptStringsCount > 0
}

func (tag *hiddenTagInternal) SearchExceptValue(i int64, s string) *hiddenTagExceptValue {
	idx := tag.exceptValues.Search(i, s)
	if idx < 0 {
		return nil
	}
	return &tag.exceptValues[idx]
}

func (tag *hiddenTagInternal) LessThan(cmp *hiddenTagInternal) bool {
	return tag.Key < cmp.Key
}

func (s hiddenTagsInternal) Len() int {
	return len(s)
}

func (s hiddenTagsInternal) Less(i, j int) bool {
	return s[i].LessThan(&s[j])
}

func (s hiddenTagsInternal) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s hiddenTagsInternal) Sort() {
	sort.Sort(s)
}

func (s hiddenTagsInternal) Search(tagKey string) int {
	l := len(s)
	idx := sort.Search(l, func(i int) bool {
		return s[i].Key >= tagKey
	})

	if idx < 0 || idx >= l {
		return -1
	}

	if s[idx].Key != tagKey {
		return -1
	}

	return idx
}

func (hiddenTags hiddenTagsInternal) IsHiddenTag(tagKey string, tagValue interface{}) bool {
	idx := hiddenTags.Search(tagKey)
	if idx < 0 {
		return false
	}

	hiddenTag := &hiddenTags[idx]

	if !hiddenTag.HasExceptValues() {
		return true
	}

	var i int64
	var s string
	if intV, ok := toInt64(tagValue); ok {
		if !hiddenTag.HasExceptInts() {
			return true
		}
		i = intV
	} else {
		if !hiddenTag.HasExceptStrings() {
			return true
		}
		s = TagValueToString(tagValue)
	}

	return hiddenTag.SearchExceptValue(i, s) == nil
}

// TODO: deduplicate with (hiddenTagsInternal).IsHiddenTag()
func (hiddenTags hiddenTagsInternal) isHiddenTagByIntAndString(tagKey string, intValue int64, stringValue string) bool {
	idx := hiddenTags.Search(tagKey)
	if idx < 0 {
		return false
	}

	hiddenTag := &hiddenTags[idx]

	if !hiddenTag.HasExceptValues() {
		return true
	}

	if len(stringValue) == 0 {
		if !hiddenTag.HasExceptInts() {
			return true
		}
	} else {
		if !hiddenTag.HasExceptStrings() {
			return true
		}
	}

	return hiddenTag.SearchExceptValue(intValue, stringValue) == nil
}
