package metrics

type metricRegistryItem struct {
	name        string
	tags        *FastTags
	description string
	storageKey  []byte

	parent Metric
}

func (item *metricRegistryItem) init(parent Metric, name string) {
	item.parent = parent
	item.name = name
}

func (item *metricRegistryItem) considerHiddenTags() {
	considerHiddenTags(item.tags)
}
func (item *metricRegistryItem) generateStorageKey() *keyGeneratorReusables {
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
	return item.tags.ToMap()
}
func (item *metricRegistryItem) GetKey() []byte {
	if item == nil {
		return nil
	}
	return item.storageKey
}
func (item *metricRegistryItem) GetTag(key string) interface{} {
	if item.tags == nil {
		return nil
	}
	return item.tags.Get(key)
}
