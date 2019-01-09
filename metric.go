package metrics

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Metric struct {
	worker      Worker
	name        string
	tags        Tags
	description string
	storageKey  []byte
}

func (metric *Metric) considerHiddenTags() {
	considerHiddenTags(metric.tags)
}
func (metric *Metric) generateStorageKey() *preallocatedStringerBuffer {
	return generateStorageKey(metric.worker.GetType(), metric.name, metric.tags)
}

func (metric *Metric) GetWorker() Worker {
	return metric.worker
}
func (metric *Metric) IsRunning() bool {
	return metric.worker.IsRunning()
}
func (metric *Metric) SetGCEnabled(enable bool) {
	metric.worker.SetGCEnabled(enable)
}
func (metric *Metric) Stop() {
	metric.worker.Stop()
}
func (metric *Metric) GetName() string {
	return metric.name
}
func (metric *Metric) GetTags() Tags {
	return metric.tags.Copy()
}

func (metric *Metric) MarshalJSON() ([]byte, error) {
	switch metric.worker.GetType() {
	case MetricTypeTiming:
		return metric.marshalJSONTiming()
	}
	return metric.marshalJSONDefault()

}

func (metric *Metric) marshalJSONTiming() ([]byte, error) {

	worker := metric.worker.(WorkerTiming)
	values := worker.GetValuePointers()
	var jsonValues []string

	considerValue := func(label string) func(data *TimingValue) {
		return func(data *TimingValue) {
			if data.Count == 0 {
				return
			}
			jsonValues = append(jsonValues, fmt.Sprintf(`"%v":{"count":%d,"min":%d,"mid":%d,"avg":%d,"per99":%d,"max":%d}`,
				label,
				data.Count,
				data.Min,
				data.Mid,
				data.Avg,
				data.Per99,
				data.Max,
			))
		}
	}

	values.Last.LockDo(considerValue(`last`))
	values.S1.LockDo(considerValue(`1_second`))
	values.S5.LockDo(considerValue(`5_seconds`))
	values.M1.LockDo(considerValue(`1_minute`))
	values.M5.LockDo(considerValue(`5_minutes`))
	values.H1.LockDo(considerValue(`1_hour`))
	values.H6.LockDo(considerValue(`6_hours`))
	values.D1.LockDo(considerValue(`1_day`))
	values.Total.LockDo(considerValue(`total`))

	nameJSON, _ := json.Marshal(metric.name)
	descriptionJSON, _ := json.Marshal(metric.description)
	tagsJSON, _ := json.Marshal(string(metric.storageKey[:strings.IndexByte(string(metric.storageKey), '@')]))
	typeJSON, _ := json.Marshal(string(metric.worker.GetType()))

	valueJSON := `{` + strings.Join(jsonValues, `,`) + `}`

	metricJSON := fmt.Sprintf(`{"name":%s,"tags":%s,"value":%s,"description":%s,"type":%s}`,
		string(nameJSON),
		tagsJSON,
		valueJSON,
		string(descriptionJSON),
		string(typeJSON),
	)
	return []byte(metricJSON), nil
}

func (metric *Metric) marshalJSONDefault() ([]byte, error) {
	nameJSON, _ := json.Marshal(metric.name)
	descriptionJSON, _ := json.Marshal(metric.description)
	tagsJSON, _ := json.Marshal(string(metric.storageKey[:strings.IndexByte(string(metric.storageKey), '@')]))
	typeJSON, _ := json.Marshal(string(metric.worker.GetType()))
	value := metric.worker.Get()

	metricJSON := fmt.Sprintf(`{"name":%s,"tags":%s,"value":%d,"description":%s,"type":%s}`,
		string(nameJSON),
		string(tagsJSON),
		value,
		string(descriptionJSON),
		string(typeJSON),
	)
	return []byte(metricJSON), nil
}
