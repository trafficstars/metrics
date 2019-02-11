package metrics

type metricRegistryItem struct {
	name        string
	tags        Tags
	description string
	storageKey  []byte

	parent Metric
}

func (item *metricRegistryItem) init(parent Metric) {
	item.parent = parent
}

func (item *metricRegistryItem) considerHiddenTags() {
	considerHiddenTags(item.tags)
}
func (item *metricRegistryItem) generateStorageKey() *preallocatedStringerBuffer {
	return generateStorageKey(item.parent.GetType(), item.name, item.tags)
}

func (item *metricRegistryItem) GetMetric() Metric {
	return item.parent
}
func (item *metricRegistryItem) SetGCEnabled(enable bool) {
	item.parent.SetGCEnabled(enable)
}
func (item *metricRegistryItem) GetName() string {
	return item.name
}
func (item *metricRegistryItem) GetTags() Tags {
	return item.tags.Copy()
}
func (item *metricRegistryItem) GetKey() []byte {
	if item == nil {
		return []byte{}
	}
	return item.storageKey
}
func (item *metricRegistryItem) GetTag(key string) interface{} {
	return item.tags[key]
}
