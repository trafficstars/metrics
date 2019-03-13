package metrics

import "github.com/sirupsen/logrus"

func AnyTagsToLogrusFields(tags AnyTags) logrus.Fields {
	fields := logrus.Fields{}
	for _, tag := range *tags.ToFastTags() {
		fields[tag.Key] = tag.GetValue()
	}
	return fields
}
