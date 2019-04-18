package metrics

type registryItem struct {
	name        string
	tags        *FastTags
	description string
	storageKey  []byte

	parent Metric
}

func (item *registryItem) init(parent Metric, name string) {
	item.parent = parent
	item.name = name
}

func (item *registryItem) considerHiddenTags() {
	considerHiddenTags(item.tags)
}
func (item *registryItem) generateStorageKey() *keyGeneratorReusables {
	return generateStorageKey(item.parent.GetType(), item.name, item.tags)
}

func (item *registryItem) GetMetric() Metric {
	return item.parent
}
func (item *registryItem) SetGCEnabled(enable bool) {
	item.parent.SetGCEnabled(enable)
}
func (item *registryItem) GetName() string {
	return item.name
}
func (item *registryItem) GetTags() Tags {
	return item.tags.ToMap()
}
func (item *registryItem) GetKey() []byte {
	if item == nil {
		return nil
	}
	return item.storageKey
}
func (item *registryItem) GetTag(key string) interface{} {
	if item.tags == nil {
		return nil
	}
	return item.tags.Get(key)
}
