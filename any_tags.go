package metrics

type AnyTags interface {
	Get(key string) interface{}
	Set(key string, value interface{}) AnyTags
	Each(func(k string, v interface{}) bool)
	ToFastTags() *FastTags
	ToMap(fieldMaps ...map[string]interface{}) map[string]interface{}
	Release()
}
