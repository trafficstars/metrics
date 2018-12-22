package metrics

type AnyTags interface {
	Get(key string) interface{}
	Set(key string, value interface{})
	Each(func(k string, v interface{}) bool)
	ToFastTags() *FastTags
}
