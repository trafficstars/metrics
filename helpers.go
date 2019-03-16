package metrics

func TagsToMap(tags AnyTags, fieldMaps... map[string]interface{}) map[string]interface{} {
	fields := map[string]interface{}{}
	for _, fieldMap := range fieldMaps {
		for k, v := range fieldMap {
			fields[k] = v
		}
	}
	tags.Each(func(k string, v interface{}) bool {
		fields[k] = v
		return true
	})
	return fields
}
