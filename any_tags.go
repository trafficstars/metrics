package metrics

// AnyTags is an abstraction over "Tags" and "*FastTags"
type AnyTags interface {
	// Get value of tag by key
	Get(key string) interface{}

	// Set tag by key and value (if tag by the key does not exist then add it otherwise overwrite the value)
	Set(key string, value interface{}) AnyTags

	// Each iterates over all tags and passes tag key/value to the function (the argument)
	// The function may return false if it's required to finish the loop prematurely
	Each(func(key string, value interface{}) bool)

	// ToFastTags returns the tags as "*FastTags"
	ToFastTags() *FastTags

	// ToMap gets the tags as a map, overwrites values by according keys using overwriteMaps and returns
	// the result
	ToMap(overwriteMaps ...map[string]interface{}) map[string]interface{}

	// Release puts tags the structure/slice back to the pool to be reused in future
	Release()

	// WriteAsString writes tags in StatsD format through the WriteStringer passed as the argument
	WriteAsString(interface{ WriteString(string) (int, error) })

	// String returns the tags as a string in the StatsD format
	String() string

	// Len returns the amount/count of tags
	Len() int
}
