package metrics

import "github.com/sirupsen/logrus"

func AnyTagsToLogrusFields(tags AnyTags) logrus.Fields {
	fields := logrus.Fields{}
	tags.Each(func(k string, v interface{}) bool {
		fields[k] = v
		return true
	})
	return fields
}
